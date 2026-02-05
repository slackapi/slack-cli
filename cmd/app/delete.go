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
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/apps"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Handle to client's function used for testing
var runDeleteCommandFunc = RunDeleteCommand

var deleteAppSelectPromptFunc = prompts.AppSelectPrompt

// NewDeleteCommand returns a new Cobra command
func NewDeleteCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [flags]",
		Aliases: []string{"del"},
		Short:   "Delete the app",
		Long:    "Uninstall the app from the team and permanently delete the app and all of its data",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app delete", Meaning: "Delete an app and app info from a team"},
			{Command: "app delete --team T0123456 --app local", Meaning: "Delete a specific app from a team"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			env, err := runDeleteCommandFunc(ctx, clients, cmd, args)
			if err != nil {
				return err
			}
			return printDeleteSuccess(ctx, clients, cmd, env)
		},
	}

	return cmd
}

// RunDeleteCommand executes the workspace delete command, prints output, and returns any errors.
func RunDeleteCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string) (types.App, error) {
	if cmd == nil {
		return types.App{}, slackerror.New("command is nil")
	}

	// Get the app auth selection from the flag or prompt
	selection, err := deleteAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		if slackerror.ToSlackError(err).Code == slackerror.ErrInstallationRequired {
			return types.App{}, nil
		} else {
			return types.App{}, err
		}
	}
	if selection.Auth.TeamDomain == "" {
		return types.App{}, slackerror.New(slackerror.ErrCredentialsNotFound)
	}

	team := selection.Auth.TeamDomain

	if !clients.Config.ForceFlag {
		proceed, err := confirmDeletion(ctx, clients.IO, selection)
		if err != nil {
			return types.App{}, err
		}
		if !proceed {
			cmd.Printf("\n%s", style.Sectionf(style.TextSection{
				Emoji: "thumbs_up",
				Text:  "Your app will not be deleted",
			}))
			return types.App{}, nil
		}
	}

	// Set up event logger and execute the command
	log := newDeleteLogger(clients, cmd, team)
	log.Data["appID"] = selection.App.AppID
	env, err := apps.Delete(ctx, clients, log, team, selection.App, selection.Auth)

	return env, err
}

// newDeleteLogger creates a logger instance to receive event notifications
func newDeleteLogger(clients *shared.ClientFactory, cmd *cobra.Command, envName string) *logger.Logger {
	ctx := cmd.Context()
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			appID := event.DataToString("appID")
			teamName := event.DataToString("teamName")
			switch event.Name {
			case "on_apps_delete_app_success":
				printDeleteApp(ctx, clients, appID, teamName)
			default:
				// Ignore the event
			}
		},
	)
}

func confirmDeletion(ctx context.Context, IO iostreams.IOStreamer, app prompts.SelectedApp) (bool, error) {
	IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "warning",
		Text:  style.Bold("Danger zone"),
		Secondary: []string{
			fmt.Sprintf("App (%s) will be permanently deleted", app.App.AppID),
			"All triggers, workflows, and functions will be deleted",
			"All datastores for this app will be deleted",
			"Once you delete this app, there is no going back",
		},
	}))

	proceed, err := IO.ConfirmPrompt(ctx, "Are you sure you want to delete the app?", false)
	return proceed, err
}

// printDeleteApp displays info about removing app from API
func printDeleteApp(ctx context.Context, clients *shared.ClientFactory, appID string, teamName string) {
	_, _ = clients.IO.WriteOut().Write([]byte(fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "house",
		Text:  "App Uninstall",
		Secondary: []string{
			fmt.Sprintf(`Uninstalled the app "%s" from "%s"`, appID, teamName),
		},
	}))))
	_, _ = clients.IO.WriteOut().Write([]byte(fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "App Manifest",
		Secondary: []string{
			fmt.Sprintf(`Deleted the app manifest for "%s" from "%s"`, appID, teamName),
		},
	}))))
}

// printDeleteSuccess will print a list of the apps
func printDeleteSuccess(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, app types.App) error {
	// Print all apps
	return runListCommand(cmd, clients)
}
