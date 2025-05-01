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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/cmd/feedback"
	"github.com/slackapi/slack-cli/cmd/triggers"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/platform"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Create handle to Deploy function for testing
// TODO - Stopgap until we learn the correct way to structure our code for testing.
var deployFunc = platform.Deploy
var teamAppSelectPromptFunc = prompts.TeamAppSelectPrompt

// TODO - Same as above, but probably even worse
var runAddCommandFunc = app.RunAddCommand

type deployCmdFlags struct {
	hideTriggers        bool
	orgGrantWorkspaceID string
}

var deployFlags deployCmdFlags

var packageSpinner *style.Spinner
var deploySpinner *style.Spinner

func NewDeployCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [flags]",
		Short: "Deploy the app to the Slack Platform",
		Long:  `Deploy the app to the Slack Platform`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "platform deploy", Meaning: "Select the workspace to deploy to"},
			{Command: "platform deploy --team T0123456", Meaning: "Deploy to a specific team"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var event *logger.LogEvent
			ctx := cmd.Context()

			packageSpinner = style.NewSpinner(cmd.OutOrStdout())
			deploySpinner = style.NewSpinner(cmd.OutOrStdout())
			defer func() {
				packageSpinner.Stop()
				deploySpinner.Stop()
			}()

			selection, err := teamAppSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowAllApps)
			if err != nil {
				return err
			}
			err = hasValidDeploymentMethod(ctx, clients, selection.App, selection.Auth)
			if err != nil {
				return err
			}

			ctx, installState, app, err := runAddCommandFunc(ctx, clients, &selection, deployFlags.orgGrantWorkspaceID)
			if err != nil {
				return err
			}
			if installState == types.InstallRequestPending || installState == types.InstallRequestCancelled || installState == types.InstallRequestNotSent {
				return nil
			}
			switch {
			case clients.SDKConfig.Hooks.Deploy.IsAvailable():
				event, err = deployHook(ctx, clients)
				if err != nil {
					return err
				}
			default:
				log := newDeployLogger(cmd)
				showTriggers := triggers.ShowTriggers(clients, deployFlags.hideTriggers)
				event, err = deployFunc(ctx, clients, showTriggers, log, app)
				if err != nil {
					return err
				}
			}

			err = printDeployHostingCompletion(clients, cmd, event)
			if err != nil {
				return err
			}
			err = feedback.ShowSurveyMessages(ctx, clients)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&deployFlags.hideTriggers, "hide-triggers", false, "do not list triggers and skip trigger creation prompts")
	cmd.Flags().StringVar(&deployFlags.orgGrantWorkspaceID, cmdutil.OrgGrantWorkspaceFlag, "", cmdutil.OrgGrantWorkspaceDescription())

	return cmd
}

// newDeployLogger creates a logger instance to receive event notifications
func newDeployLogger(cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_app_package":
				printDeployPackage(cmd, event)
			case "on_app_package_completion":
				printDeployPackageCompletion(cmd, event)
			case "on_app_deploy_hosting":
				printDeployHosting(cmd, event)
			default:
				// Ignore the event
			}
		},
	)
}

// hasValidDeploymentMethod errors if an app has no known ways to deploy
func hasValidDeploymentMethod(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) error {
	if clients.SDKConfig.Hooks.Deploy.IsAvailable() {
		return nil
	}
	manifest := types.SlackYaml{}
	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return err
	}
	switch {
	case manifestSource.Equals(config.ManifestSourceLocal):
		manifest, err = clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
		if err != nil {
			return err
		}
	case manifestSource.Equals(config.ManifestSourceRemote):
		manifest, err = clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
		if err != nil {
			return err
		}
	}
	if manifest.FunctionRuntime() == types.SlackHosted {
		return nil
	}
	return errorMissingDeployHook(clients)
}

// deployHook executes the provided program and streams IO for the process
func deployHook(ctx context.Context, clients *shared.ClientFactory) (*logger.LogEvent, error) {
	var log = logger.LogEvent{
		// FIXME: Include app information
		Data: logger.LogData{"authSession": "{}"},
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "mailbox_with_mail",
		Text:      "App Deploy",
		Secondary: []string{"Running the command provided to the deploy hook"},
	}))
	_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
		Emoji: "robot",
		Text:  clients.SDKConfig.Hooks.Deploy.Command,
	})))
	var hookExecOpts = hooks.HookExecOpts{
		Hook:   clients.SDKConfig.Hooks.Deploy,
		Stdin:  clients.IO.ReadIn(),
		Stdout: clients.IO.WriteIndent(clients.IO.WriteSecondary(clients.IO.WriteOut())),
		Stderr: clients.IO.WriteIndent(clients.IO.WriteSecondary(clients.IO.WriteErr())),
	}
	// The "default" protocol is used because a certain response is not expected of
	// scripts provided to the "deploy" hook.
	//
	// The "message boundaries" protocol appends information to scripts using flags
	// which might cause some commands to error.
	//
	// The hook executor attached to the provided clients might use either protocol
	// so we instantiate the default here.
	shell := hooks.HookExecutorDefaultProtocol{
		IO: clients.IO,
	}
	if _, err := shell.Execute(ctx, hookExecOpts); err != nil {
		return &log, err
	}
	// Follow successful hook executions with a newline to match section formatting
	// but break immediately after an error!
	_, _ = clients.IO.WriteOut().Write([]byte("\n"))
	return &log, nil
}

func printDeployPackage(cmd *cobra.Command, event *logger.LogEvent) {
	cmd.Println()
	packageSpinner.Update("Packaging app for deployment", "").Start()
}

func printDeployPackageCompletion(cmd *cobra.Command, event *logger.LogEvent) {
	packagedSize := event.DataToString("packagedSize")
	packagedTime := event.DataToString("packagedTime")

	deployPackageSuccessText := style.Sectionf(style.TextSection{
		Emoji:     "gift",
		Text:      "App packaged and ready to deploy",
		Secondary: []string{fmt.Sprintf("%s was packaged in %s", packagedSize, packagedTime)},
	})
	packageSpinner.Update(deployPackageSuccessText, "").Stop()
}

func printDeployHosting(cmd *cobra.Command, event *logger.LogEvent) {
	deployHostingText := "Deploying to Slack Platform" + style.Secondary(" (this will take a moment)")
	deploySpinner.Update(deployHostingText, "").Start()
}

func printDeployHostingCompletion(clients *shared.ClientFactory, cmd *cobra.Command, event *logger.LogEvent) error {
	var ctx = cmd.Context()
	var authSession api.AuthSession

	appName := event.DataToString("appName")
	deployTime := event.DataToString("deployTime")
	err := json.Unmarshal([]byte(event.DataToString("authSession")), &authSession)
	if err != nil {
		return err
	}

	parsedAppInfo := map[string]string{}

	host := clients.APIInterface().Host()
	if appID := event.DataToString("appID"); appID != "" && host != "" {
		parsedAppInfo["Dashboard"] = fmt.Sprintf("%s/apps/%s", host, appID)
	}

	if authSession.UserName != nil && authSession.UserID != nil {
		userInfo := fmt.Sprintf("%s (%s)", *authSession.UserName, *authSession.UserID)
		parsedAppInfo["App Owner"] = userInfo
	}

	if authSession.TeamName != nil && authSession.TeamID != nil {
		workspaceInfo := fmt.Sprintf("%s (%s)", *authSession.TeamName, *authSession.TeamID)
		if authSession.EnterpriseID != nil {
			parsedAppInfo["Organization"] = workspaceInfo
		} else {
			parsedAppInfo["Workspace"] = workspaceInfo
		}
	}

	var finalMessage string
	if appName != "" && deployTime != "" {
		finalMessage = fmt.Sprintf("%s deployed in %s", style.Bold(appName), deployTime)
	} else {
		finalMessage = "Successfully deployed the app!"
	}

	successfulDeployText := style.Sectionf(style.TextSection{
		Emoji:     "rocket",
		Text:      finalMessage,
		Secondary: []string{style.Mapf(parsedAppInfo)},
	})

	deploySpinner.Update(successfulDeployText, "").Stop()

	clients.IO.PrintTrace(ctx, slacktrace.PlatformDeploySuccess)

	navigateText := style.Sectionf(style.TextSection{
		Emoji: "cloud_with_lightning",
		Text:  "Visit Slack to try out your live app!",
		Secondary: []string{
			"When you make any changes, update your app by re-running " + style.Commandf("deploy", false),
			"Review the current activity logs using " + style.Commandf("activity --tail", false),
		},
	})

	clients.IO.PrintInfo(ctx, false, navigateText)
	return nil
}

// errorMissingDeployHook returns a descriptive error for a missing deploy hook
func errorMissingDeployHook(clients *shared.ClientFactory) error {
	if !clients.SDKConfig.Hooks.Deploy.IsAvailable() {
		return slackerror.New(slackerror.ErrSDKHookNotFound).
			WithMessage("Missing the `deploy` hook from the `%s` file", config.GetProjectHooksJSONFilePath()).
			WithRemediation("%s", strings.Join([]string{
				"Provide a command or script to run with the deploy command by adding a new hook.",
				"",
				fmt.Sprintf("Example `%s` `deploy` hook:", config.GetProjectHooksJSONFilePath()),
				"{",
				`  "hooks": {`,
				`    "deploy": "./deploy.sh"`,
				"  }",
				"}",
			}, "\n"))
	}
	return nil
}
