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

package app

import (
	"context"
	"fmt"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Handle to function used for testing
var unlinkAppSelectPromptFunc = prompts.AppSelectPrompt

// NewUnlinkCommand returns a new Cobra command for unlinking apps
func NewUnlinkCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove linked app from the project",
		Long:  "Unlink a previously linked app from the project",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Remove an existing app from the project",
				Command: "app unlink",
			},
			{
				Meaning: "Remove a specific app without using prompts",
				Command: "app unlink --team T0123456789 --app A0123456789 --environment deployed",
			},
		}),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AppUnlinkStart)

			app, err := UnlinkCommandRunE(ctx, clients, cmd, args)
			if err != nil {
				return err
			}
			return printUnlinkSuccess(ctx, clients, app)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AppUnlinkSuccess)
			return nil
		},
	}
	return cmd
}

// UnlinkCommandRunE executes the unlink command, prints output, and returns any errors.
func UnlinkCommandRunE(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string) (types.App, error) {
	// Get the app selection from the flag or prompt
	selection, err := unlinkAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return types.App{}, err
	}

	// Confirm with user unless --force flag is used
	if !clients.Config.ForceFlag {
		proceed, err := confirmUnlink(ctx, clients.IO, selection)
		if err != nil {
			return types.App{}, err
		}
		if !proceed {
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "thumbs_up",
				Text:  "Your app will not be unlinked",
			}))
			return types.App{}, nil
		}
	}

	// Remove the app from the project
	app, err := clients.AppClient().Remove(ctx, selection.App)
	if err != nil {
		return types.App{}, err
	}

	// Clean up empty files
	clients.AppClient().CleanUp()

	return app, nil
}

// confirmUnlink prompts the user to confirm unlinking the app
func confirmUnlink(ctx context.Context, IO iostreams.IOStreamer, selection prompts.SelectedApp) (bool, error) {
	IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "warning",
		Text:  "Confirm Unlink",
		Secondary: []string{
			fmt.Sprintf("App (%s) will be removed from this project", selection.App.AppID),
			fmt.Sprintf("Team: %s", selection.Auth.TeamDomain),
			"The app will not be deleted from Slack",
			"You can re-link it later with 'slack app link'",
		},
	}))

	proceed, err := IO.ConfirmPrompt(ctx, "Are you sure you want to unlink this app?", false)
	return proceed, err
}

// printUnlinkSuccess displays success message after unlinking
func printUnlinkSuccess(ctx context.Context, clients *shared.ClientFactory, app types.App) error {
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "white_check_mark",
		Text:  "App Unlinked",
		Secondary: []string{
			fmt.Sprintf("Removed app %s from project", app.AppID),
			fmt.Sprintf("Team: %s", app.TeamDomain),
		},
	}))
	return nil
}
