// Copyright 2022-2026 Salesforce, Inc.
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

package app

import (
	"context"
	"fmt"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// statusAppSelectPromptFunc is a handle to the app select prompt used for testing
var statusAppSelectPromptFunc = prompts.AppSelectPrompt

// NewStatusCommand returns a new Cobra command for app status
func NewStatusCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check the install request status of an app",
		Long:  "Check the install request status of an app on a workspace where direct installation is not permitted",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app status", Meaning: "Check install request status for an app in the current project"},
			{Command: "app status --app A0123456789", Meaning: "Check install request status for a specific app"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusCommand(cmd, clients)
		},
	}

	return cmd
}

// runStatusCommand executes the app status command
func runStatusCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	selection, err := statusAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return err
	}

	if selection.App.AppID == "" {
		return slackerror.New(slackerror.ErrAppNotFound)
	}

	appStatus, err := fetchAppInstallRequestStatus(ctx, clients, selection)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "clipboard",
		Text:      "App Install Request Status",
		Secondary: formatStatusOutput(selection, appStatus),
	}))

	return nil
}

// AppInstallRequestStatus represents the approval status of an install request
type AppInstallRequestStatus string

const (
	InstallRequestStatusInstalled AppInstallRequestStatus = "Installed"
	InstallRequestStatusApproved  AppInstallRequestStatus = "Approved"
	InstallRequestStatusPending   AppInstallRequestStatus = "Pending"
	InstallRequestStatusDenied    AppInstallRequestStatus = "Denied"
	InstallRequestStatusUnknown   AppInstallRequestStatus = "Unknown"
)

// fetchAppInstallRequestStatus fetches the install request status for an app
// TODO: Update to use the new API endpoint when available
func fetchAppInstallRequestStatus(ctx context.Context, clients *shared.ClientFactory, selection prompts.SelectedApp) (AppInstallRequestStatus, error) {
	result, err := clients.API().GetAppStatus(ctx, selection.Auth.Token, []string{selection.App.AppID}, selection.App.TeamID)
	if err != nil {
		return InstallRequestStatusUnknown, slackerror.Wrap(err, slackerror.ErrAppNotFound)
	}

	for _, app := range result.Apps {
		if app.AppID == selection.App.AppID {
			if app.Installed {
				return InstallRequestStatusInstalled, nil
			}
			// TODO: Map approval_status field from API response when available
			// For now, return Unknown for uninstalled apps
			return InstallRequestStatusUnknown, nil
		}
	}

	return InstallRequestStatusUnknown, nil
}

// formatStatusOutput formats the status command output
func formatStatusOutput(selection prompts.SelectedApp, status AppInstallRequestStatus) []string {
	var output []string

	output = append(output, fmt.Sprintf(style.Bold("%s:"), selection.Auth.TeamDomain))
	output = append(output, fmt.Sprintf(style.Indent(style.Secondary("App ID:  %s")), selection.App.AppID))
	output = append(output, fmt.Sprintf(style.Indent(style.Secondary("Team ID: %s")), selection.App.TeamID))
	output = append(output, fmt.Sprintf(style.Indent(style.Secondary("Status:  %s")), formatStatusLabel(status)))

	return output
}

// formatStatusLabel returns a styled label for the given status
func formatStatusLabel(status AppInstallRequestStatus) string {
	switch status {
	case InstallRequestStatusInstalled:
		return style.Green(string(status))
	case InstallRequestStatusApproved:
		return style.Green(string(status))
	case InstallRequestStatusPending:
		return style.Warning(string(status))
	case InstallRequestStatusDenied:
		return style.Red(string(status))
	default:
		return string(InstallRequestStatusUnknown)
	}
}
