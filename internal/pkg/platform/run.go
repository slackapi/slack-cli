// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/cmd/feedback"
	"github.com/slackapi/slack-cli/cmd/triggers"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/apps"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
)

// RunArgs are the arguments passed into the Run function
type RunArgs struct {
	Activity            bool
	ActivityLevel       string
	App                 types.App
	Auth                types.SlackAuth
	Cleanup             bool
	ShowTriggers        bool
	OrgGrantWorkspaceID string
}

// Run locally runs your app.
func Run(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, runArgs RunArgs) (*logger.LogEvent, types.InstallState, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.create")
	defer span.Finish()

	// Save token for use later on
	ctx = config.SetContextToken(ctx, runArgs.Auth.Token)

	// Validate auth session
	authSession, err := clients.API().ValidateSession(ctx, runArgs.Auth.Token)
	if err != nil {
		err = slackerror.Wrap(err, "No auth session found")
		return nil, "", slackerror.Wrap(err, slackerror.ErrLocalAppRun)
	}
	log.Data["teamName"] = *authSession.TeamName

	if authSession.UserID != nil {
		ctx = config.SetContextUserID(ctx, *authSession.UserID)
		clients.EventTracker.SetAuthUserID(*authSession.UserID)
	}

	// Load local dev environment
	clients.Config.TeamFlag = *authSession.TeamID

	// Load the project's CLI config
	var cliConfig = clients.SDKConfig

	// A Start script hook must be provided in SDKCLIConfig in order
	// For CLI Run command execute successfully
	if !clients.SDKConfig.Hooks.Start.IsAvailable() {
		var err = slackerror.New(slackerror.ErrSDKHookNotFound).WithMessage("The `start` script was not found")
		return nil, "", err
	}

	// Update local install
	installedApp, localInstallResult, installState, err := apps.InstallLocalApp(ctx, clients, runArgs.OrgGrantWorkspaceID, log, runArgs.Auth, runArgs.App)
	if err != nil {
		return nil, "", slackerror.Wrap(err, slackerror.ErrLocalAppRun)
	}

	if installState == types.InstallRequestPending || installState == types.InstallRequestCancelled || installState == types.InstallRequestNotSent {
		return log.SuccessEvent(), types.InstallSuccess, nil
	}

	if runArgs.ShowTriggers {
		// Generate an optional trigger when none exist
		_, err = triggers.TriggerGenerate(ctx, clients, installedApp)
		if err != nil {
			if strings.Contains(err.Error(), "workflow_not_found") {
				listErr := triggers.ListWorkflows(ctx, clients, runArgs.App, runArgs.Auth)
				if listErr != nil {
					return nil, "", slackerror.Wrap(listErr, "Error listing workflows").WithRootCause(err)
				}
			}
			return nil, "", err
		}
	}

	// Gather environment variables from an environment file
	variables, err := clients.Config.GetDotEnvFileVariables()
	if err != nil {
		return nil, "", slackerror.Wrap(err, slackerror.ErrLocalAppRun).
			WithMessage("Failed to read the local .env file")
	}

	// Set SLACK_API_URL to the resolved host value found in the environment
	if value, ok := variables["SLACK_API_URL"]; ok {
		_ = clients.Os.Setenv("SLACK_API_URL", value)
	} else {
		variables["SLACK_API_URL"] = fmt.Sprintf("%s/api/", clients.Config.APIHostResolved)
	}

	var localHostedContext = LocalHostedContext{
		BotAccessToken: localInstallResult.APIAccessTokens.Bot,
		AppID:          installedApp.AppID,
		TeamID:         *authSession.TeamID,
		Variables:      variables,
	}

	var server = LocalServer{
		clients:            clients,
		log:                log,
		token:              localInstallResult.APIAccessTokens.AppLevel,
		localHostedContext: localHostedContext,
		cliConfig:          cliConfig,
		Connection:         nil,
	}

	// Once the "run" command completes, delete the app if the --cleanup flag is
	// provided and gracefully shutdown websocket connections.
	//
	// Signal to CLI that the cleanup in this goroutine needs to complete before
	// exiting the process.
	clients.CleanupWaitGroup.Add(1)
	go func(cleanup bool) {
		// Wait until process interrupt via ctx.Done() / canceled context
		<-ctx.Done()
		clients.IO.PrintDebug(ctx, "Interrupt signal received in Run command, cleaning up and shutting down")
		if cleanup {
			deleteAppOnTerminate(ctx, clients, runArgs.Auth, installedApp, log)
		}
		feedback.ShowFeedbackMessageOnTerminate(ctx, clients)
		// Notify Slack backend we are closing connection; this should trigger an echoing close message from Slack
		// in Listen() below (as per WS spec), which we can detect and gracefully return from.
		// In turn, should trigger the final cleanup routine below (see deferred function below), which closes the socket connection.
		// In case this is an SDK-managed run, the next line is a no-op.
		sendWebSocketCloseControlMessage(ctx, clients, server.Connection)
		clients.IO.PrintTrace(ctx, slacktrace.PlatformRunStop)
		clients.CleanupWaitGroup.Done()
	}(runArgs.Cleanup)

	// Coordinate three goroutines for long running processes and exit on error:
	//  1. Watch for changes to the app manifest
	//  2. Watch for activity logs from the Slack API
	//  3. Listen to events over a managed socket connection or wait for the SDK
	//    delegated "start" command to exit
	//
	// An error channel is shared between the goroutines so that the context can
	// be canceled, then cleanup performed, with the erroring error returned.
	errChan := make(chan error)
	clients.IO.PrintTrace(ctx, slacktrace.PlatformRunStart)

	// Start watching for Slack Platform log activity
	if runArgs.Activity {
		go func() {
			errChan <- server.WatchActivityLogs(ctx, runArgs.ActivityLevel)
		}()
	}

	// Start watching for manifest changes
	// TODO - reinstalled apps via FS watcher do nothing with new tokens returned - may lead to permission issues / missing events?
	go func() {
		errChan <- server.Watch(ctx, runArgs.Auth, installedApp)
	}()

	// Check to see whether the SDK managed connection flag is enabled
	// If so Delegate the connection to the SDK otherwise Start connection
	if cliConfig.Config.SDKManagedConnection {
		clients.IO.PrintDebug(ctx, "Delegating connection to SDK managed script hook")
		// Delegate connection to hook; this should be a blocking call, as the delegate should be a server, too.
		go func() {
			errChan <- server.StartDelegate(ctx)
		}()
	} else {
		// Listen for messages in a goroutine, and provide an error channel for raising errors and a done channel for signifying clean exit
		go func() {
			errChan <- server.Start(ctx)
		}()
	}
	if err := <-errChan; err != nil {
		switch slackerror.ToSlackError(err).Code {
		case slackerror.ErrLocalAppRunCleanExit:
			return log.SuccessEvent(), types.InstallSuccess, nil
		case slackerror.ErrSDKHookInvocationFailed:
			return nil, "", err
		}
		return nil, "", slackerror.Wrap(err, slackerror.ErrLocalAppRun)
	}
	return log.SuccessEvent(), types.InstallSuccess, nil
}

func deleteAppOnTerminate(ctx context.Context, clients *shared.ClientFactory, auth types.SlackAuth, app types.App, log *logger.Logger) {
	clients.IO.PrintDebug(ctx, "Removing the local version of this app from the workspace")
	_, err := apps.Delete(ctx, clients, log, app.TeamDomain, app, auth)
	if err != nil {
		log.Data["on_cleanup_app_install_error"] = err.Error()
		log.Info("on_cleanup_app_install_failed")
	}
	log.Info("on_cleanup_app_install_done")
}
