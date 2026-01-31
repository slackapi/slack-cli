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
	"fmt"

	"github.com/slackapi/slack-cli/cmd/help"
	"github.com/slackapi/slack-cli/cmd/triggers"
	internalapp "github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/platform"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type runCmdFlags struct {
	activityLevel       string
	noActivity          bool
	cleanup             bool
	hideTriggers        bool
	orgGrantWorkspaceID string
}

var runFlags runCmdFlags

// Create handle to the function for testing
// TODO - Stopgap until we learn the correct way to structure our code for testing.
var runFunc = platform.Run
var runRunCommandFunc = RunRunCommand
var runAppSelectPromptFunc = prompts.AppSelectPrompt

func NewRunCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"dev", "start-dev"}, // Aliases a few proposed alternative names
		Short:   "Start a local server to develop and run the app locally",
		Long:    `Start a local server to develop and run the app locally while watching for file changes`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "platform run", Meaning: "Start a local development server"},
			{Command: "platform run --activity-level debug", Meaning: "Run a local development server with debug activity"},
			{Command: "platform run --cleanup", Meaning: "Run a local development server with cleanup"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRunCommandFunc(clients, cmd, args)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&runFlags.activityLevel, "activity-level", platform.ActivityMinLevelDefault, "activity level to display")
	cmd.Flags().BoolVar(&runFlags.noActivity, "no-activity", false, "hide Slack Platform log activity")
	cmd.Flags().BoolVar(&runFlags.cleanup, "cleanup", false, "uninstall the local app after exiting")
	cmd.Flags().StringVar(&runFlags.orgGrantWorkspaceID, cmdutil.OrgGrantWorkspaceFlag, "", cmdutil.OrgGrantWorkspaceDescription())
	cmd.Flags().BoolVar(&runFlags.hideTriggers, "hide-triggers", false, "do not list triggers and skip trigger creation prompts")

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		style.ToggleStyles(clients.IO.IsTTY() && !clients.Config.NoColor)

		cmd.Flag("activity-level").DefValue = ""
		cmd.Flag("activity-level").Usage = fmt.Sprintf(
			"activity level to display (default \"%s\")\n  %s",
			platform.ActivityMinLevelDefault,
			style.Secondary("(trace, debug, info, warn, error, fatal)"),
		)
		cmd.Flag(cmdutil.OrgGrantWorkspaceFlag).Usage = cmdutil.OrgGrantWorkspaceDescription()

		help.PrintHelpTemplate(cmd, style.TemplateData{})
	})

	return cmd
}

// RunRunCommand executes the local run command
func RunRunCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	if cmd == nil {
		return slackerror.New("command is nil")
	}
	ctx := cmd.Context()

	// Get the workspace from the flag or prompt
	selection, err := runAppSelectPromptFunc(ctx, clients, prompts.ShowLocalOnly, prompts.ShowAllApps)
	if err != nil {
		switch slackerror.ToSlackError(err).Code {
		case slackerror.ErrDeployedAppNotSupported:
			if !clients.Config.SkipLocalFs() {
				return err
			} else {
				selection.App.IsDev = true
			}
		case slackerror.ErrMissingOptions:
			return slackerror.New(slackerror.ErrAppNotFound).
				WithMessage("No apps are available for selection").
				WithRemediation(
					"Create a new app on app settings: %s\nThen add the app to this project with %s",
					style.LinkText("https://api.slack.com/apps"),
					style.Commandf("app link", false),
				).
				WithDetails(slackerror.ErrorDetails{
					slackerror.ErrorDetail{
						Code:    slackerror.ErrProjectConfigManifestSource,
						Message: "App manifests for this project are sourced from app settings",
					},
				})
		default:
			return err
		}
	}

	runFlags.orgGrantWorkspaceID, err = prompts.ValidateGetOrgWorkspaceGrant(ctx, clients, &selection, runFlags.orgGrantWorkspaceID, false /* top prompt option should be 'all workspaces' */)
	if err != nil {
		return err
	}

	clients.Config.ManifestEnv = internalapp.SetManifestEnvTeamVars(clients.Config.ManifestEnv, selection.Auth.TeamDomain, selection.App.IsDev)

	runArgs := platform.RunArgs{
		Activity:            !runFlags.noActivity,
		ActivityLevel:       runFlags.activityLevel,
		App:                 selection.App,
		Auth:                selection.Auth,
		Cleanup:             runFlags.cleanup,
		ShowTriggers:        triggers.ShowTriggers(clients, runFlags.hideTriggers),
		OrgGrantWorkspaceID: runFlags.orgGrantWorkspaceID,
	}

	log := newRunLogger(clients, cmd)

	// Run dev app locally
	if _, _, err := runFunc(ctx, clients, log, runArgs); err != nil {
		return err
	}

	return nil
}

// newRunLogger creates a logger instance to receive event notifications
func newRunLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	ctx := cmd.Context()
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_update_app_install":
				cmd.Println(style.Secondary(fmt.Sprintf(
					`Updating local app install for "%s"`,
					event.DataToString("teamName"),
				)))
			case "on_cloud_run_connection_connected":
				clients.IO.PrintTrace(ctx, slacktrace.PlatformRunReady)
				cmd.Println(style.Secondary("Connected, awaiting events"))
			case "on_cloud_run_connection_message":
				message := event.DataToString("cloud_run_connection_message")
				clients.IO.PrintDebug(ctx, "received: %s", message)
			case "on_cloud_run_connection_command_error":
				message := event.DataToString("cloud_run_connection_command_error")
				clients.IO.PrintError(ctx, "Error: %s", message)
			case "on_cloud_run_watch_error":
				message := event.DataToString("cloud_run_watch_error")
				clients.IO.PrintError(ctx, "Error: %s", message)
			case "on_cloud_run_watch_manifest_change":
				path := event.DataToString("cloud_run_watch_manifest_change")
				cmd.Println(style.Secondary(fmt.Sprintf("Manifest change detected: %s, reinstalling app...", path)))
			case "on_cloud_run_watch_manifest_change_reinstalled":
				cmd.Println(style.Secondary("App successfully reinstalled"))
			case "on_cloud_run_watch_manifest_change_skipped_remote":
				path := event.DataToString("cloud_run_watch_manifest_change_skipped")
				cmd.Println(style.Secondary(fmt.Sprintf("Manifest change detected: %s, skipped reinstalling app because manifest.source=remote", path)))
			case "on_cloud_run_watch_app_change":
				path := event.DataToString("cloud_run_watch_app_change")
				cmd.Println(style.Secondary(fmt.Sprintf("App change detected: %s, restarting server...", path)))
			case "on_cleanup_app_install_done":
				cmd.Println(style.Secondary(fmt.Sprintf(
					`Cleaned up local app install for "%s".`,
					event.DataToString("teamName"),
				)))
			case "on_cleanup_app_install_failed":
				cmd.Println(style.Secondary(fmt.Sprintf(
					`Cleaning up local app install for "%s" failed.`,
					event.DataToString("teamName"),
				)))
				message := event.DataToString("on_cleanup_app_install_error")
				clients.IO.PrintWarning(ctx, "Local app cleanup failed: %s", message)
			case "on_abort_cleanup_app_install":
				cmd.Println(style.Secondary("Aborting, local app might not be cleaned up."))
			default:
				// Ignore the event
			}
		},
	)
}
