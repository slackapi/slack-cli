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

package platform

import (
	"fmt"

	"github.com/slackapi/slack-cli/cmd/help"
	"github.com/slackapi/slack-cli/cmd/triggers"
	internalapp "github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/pkg/platform"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
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

	runFlags.orgGrantWorkspaceID, err = prompts.ValidateGetOrgWorkspaceGrant(ctx, clients, &selection, runFlags.orgGrantWorkspaceID, true /* top prompt option should be 'all workspaces' */)
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

	// Run dev app locally
	if _, err := runFunc(ctx, clients, runArgs); err != nil {
		return err
	}

	return nil
}
