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

package manifest

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
)

// ManifestValidate validates the manifest from the project "get-manifest" hook
func ManifestValidate(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, app types.App, token string) (*logger.LogEvent, slackerror.Warnings, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pkg.ManifestValidate")
	defer span.Finish()

	if strings.TrimSpace(token) == "" {
		return nil, nil, slackerror.New(slackerror.ErrAuthToken)
	}

	_, err := clients.ApiInterface().ValidateSession(ctx, token)
	if err != nil {
		return nil, nil, slackerror.New(slackerror.ErrAuthToken).WithRootCause(err)
	}

	slackManifest, err := clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
	if err != nil {
		return nil, nil, slackerror.Wrap(err, slackerror.ErrAppManifestGenerate)
	}

	// validate the manifest
	validationResult, err := clients.ApiInterface().ValidateAppManifest(ctx, token, slackManifest.AppManifest, app.AppID)

	if retryValidate := HandleConnectorNotInstalled(ctx, clients, token, err); retryValidate {
		validationResult, err = clients.ApiInterface().ValidateAppManifest(ctx, token, slackManifest.AppManifest, app.AppID)
	}

	if err := HandleConnectorApprovalRequired(ctx, clients, token, err); err != nil {
		return nil, nil, err
	}

	if err != nil || len(validationResult.Warnings) > 0 {
		return nil, validationResult.Warnings, err
	}

	log.Data["isValid"] = true
	return log.SuccessEvent(), nil, nil
}

// HandleConnectorNotInstalled attempts install the certified app associated with any connector_not_installed error
// And returns true if it attempted to install any single app
func HandleConnectorNotInstalled(ctx context.Context, clients *shared.ClientFactory, token string, err error) (attemptInstall bool) {
	var slackError *slackerror.Error
	if !errors.As(err, &slackError) {
		return attemptInstall
	}

	for _, detail := range slackError.Details {
		// Handle error connector not installed
		if detail.Code == slackerror.ErrConnectorNotInstalled {
			attemptInstall = true
			clients.IO.PrintDebug(ctx, "Attempting to install connector app: %s", detail.RelatedComponent)
			_, err := clients.ApiInterface().CertifiedAppInstall(ctx, token, detail.RelatedComponent)
			if err != nil {
				clients.IO.PrintDebug(ctx, "Error installing connector app: %s", detail.RelatedComponent)
			}
		}
	}
	return attemptInstall
}

// HandleConnectorNotInstalled looks for Errors related to connectors requiring approval, ultimately
// it prompts the user to send the approval request.
// Note that this function will also clean up the related required approval errors in order to provide
// a clean output to the end user.
func HandleConnectorApprovalRequired(ctx context.Context, clients *shared.ClientFactory, token string, err error) error {
	var slackError *slackerror.Error
	if !errors.As(err, &slackError) {
		return err
	}

	approvalRequiredErrors := slackerror.ErrorDetails{}
	for i := len(slackError.Details) - 1; i >= 0; i-- { // Iterate in reverse to remove elements in place
		if slackError.Details[i].Code == slackerror.ErrConnectorApprovalRequired {
			approvalRequiredErrors = append(approvalRequiredErrors, slackError.Details[i])
			slackError.Details = slices.Delete(slackError.Details, i, i+1)
		}
	}

	if len(approvalRequiredErrors) > 0 {
		err := attemptConnectorAppsApprovalRequests(ctx, clients, token, approvalRequiredErrors)
		if err != nil {
			return err
		}
	}
	return nil
}

// attemptCertifiedAppsApprovalRequests prompts the user to send a request for approval, if the user authorizes the request, it
// will collect a single 'reason' and send a request for each app requiring admin approval.
func attemptConnectorAppsApprovalRequests(ctx context.Context, clients *shared.ClientFactory, token string, approvalRequiredErrorDetails slackerror.ErrorDetails) error {

	messages := []string{}
	for _, detail := range approvalRequiredErrorDetails {
		messages = append(messages, detail.Message)
	}
	var administratorApprovalNotice = fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji:     "bell",
		Text:      "Administrator approval is required to use connectors",
		Secondary: messages,
	}))
	clients.IO.PrintInfo(ctx, false, administratorApprovalNotice)

	sendApprovalRequests, err := clients.IO.ConfirmPrompt(ctx, "Request approval to install missing connectors?", true)
	if err != nil {
		return err
	}
	if !sendApprovalRequests {
		return nil
	}

	// Create an app approval request
	var reason string
	clients.IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestShouldSend)
	reason, err = clients.IO.InputPrompt(ctx, "Enter a reason for the installation:", iostreams.InputPromptConfig{
		Required: false,
	})
	if err != nil {
		return err
	}

	clients.IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestReasonSubmitted, reason)

	for _, errorDetail := range approvalRequiredErrorDetails {
		// Passing in an empty string for team_id here, meaning connectors will be requested at the org level
		_, err := clients.ApiInterface().RequestAppApproval(ctx, token, errorDetail.RelatedComponent, "", reason, "", []string{})
		if err != nil {
			clients.IO.PrintDebug(ctx, "Error requesting approval for %s", errorDetail.RelatedComponent)
			return err
		}
	}

	clients.IO.PrintTrace(ctx, slacktrace.AdminAppApprovalRequestPending)
	return nil
}
