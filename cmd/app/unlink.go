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
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
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
	var unlinkedApp types.App // capture app for PostRunE

	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove a linked app from the project",
		Long: strings.Join([]string{
			"Unlink removes an existing app from the project.",
			"",
			"This command removes a saved app ID from the files of a project without deleting",
			"the app from Slack.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Remove an existing app from the project",
				Command: "app unlink",
			},
			{
				Meaning: "Remove a specific app without using prompts",
				Command: "app unlink --app A0123456789",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			app, err := UnlinkCommandRunE(ctx, clients, cmd, args)
			if err != nil {
				return err
			}
			if app.AppID == "" { // user canceled
				return nil
			}
			unlinkedApp = app // stored for PostRunE
			return printUnlinkSuccess(ctx, clients, app)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AppUnlinkSuccess, unlinkedApp.AppID)
			return nil
		},
	}
	return cmd
}

// UnlinkCommandRunE executes the unlink command, prints output, and returns any errors.
func UnlinkCommandRunE(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string) (types.App, error) {
	clients.IO.PrintTrace(ctx, slacktrace.AppUnlinkStart)

	// Get the app selection from the flag or prompt
	selection, err := unlinkAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return types.App{}, err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "unlock",
		Text:  "App Unlink",
		Secondary: []string{
			fmt.Sprintf("App (%s) will be removed from this project", selection.App.AppID),
			"The app will not be deleted from Slack",
			fmt.Sprintf("You can re-link it later with %s", style.Commandf("app link", false)),
		},
	}))

	// Confirm with user unless --force flag is used
	if !clients.Config.ForceFlag {
		proceed, err := clients.IO.ConfirmPrompt(ctx, "Are you sure you want to unlink this app?", false)
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

	return app, nil
}

// printUnlinkSuccess displays success message after unlinking
func printUnlinkSuccess(ctx context.Context, clients *shared.ClientFactory, app types.App) error {
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "unlock",
		Text:  "App Unlink",
		Secondary: []string{
			fmt.Sprintf("Removed app %s from project", app.AppID),
			fmt.Sprintf("Team: %s", app.TeamDomain),
		},
	}))
	return nil
}
