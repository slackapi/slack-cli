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
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/apps"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var uninstallAppSelectPromptFunc = prompts.AppSelectPrompt

// NewUninstallCommand returns a new Cobra command
func NewUninstallCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uninstall [flags]",
		Aliases: []string{"uninstal"},
		Short:   "Uninstall the app from a team",
		Long:    "Uninstall the app from a team without deleting the app or its data",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app uninstall", Meaning: "Uninstall an app from a team"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			env, err := RunUninstallCommand(ctx, clients, cmd, args)
			if err != nil {
				return err
			}
			return printUninstallSuccess(ctx, clients, cmd, env)
		},
	}

	return cmd
}

// RunUninstallCommand executes the workspace uninstall command, prints output, and returns any errors.
func RunUninstallCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string) (types.App, error) {
	if cmd == nil {
		return types.App{}, slackerror.New("command is nil")
	}

	// Get the workspace from the flag or prompt
	selection, err := uninstallAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
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

	teamDomain := selection.Auth.TeamDomain

	if !clients.Config.ForceFlag {
		proceed, err := confirmUninstall(ctx, clients.IO, cmd, selection)
		if err != nil {
			return types.App{}, err
		}
		if !proceed {
			cmd.Printf("\n%s", style.Sectionf(style.TextSection{
				Emoji: "thumbs_up",
				Text:  "Your app will not be uninstalled",
			}))
			return types.App{}, nil
		}
	}

	// Set up event logger and execute the command
	log := newUninstallLogger(clients, cmd, teamDomain)
	log.Data["appID"] = selection.App.AppID
	env, err := apps.Uninstall(ctx, clients, log, teamDomain, selection.App, selection.Auth)

	return env, err
}

// newUninstallLogger creates a logger instance to receive event notifications
func newUninstallLogger(clients *shared.ClientFactory, cmd *cobra.Command, envName string) *logger.Logger {
	ctx := cmd.Context()
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			appID := event.DataToString("appID")
			teamName := event.DataToString("teamName")
			switch event.Name {
			case "on_apps_uninstall_app_success":
				printUninstallApp(ctx, clients, appID, teamName)
			default:
				// Ignore the event
			}
		},
	)
}

func confirmUninstall(ctx context.Context, IO iostreams.IOStreamer, cmd *cobra.Command, selection prompts.SelectedApp) (bool, error) {
	cmd.Printf("\n%s\n", style.Sectionf(style.TextSection{
		Emoji: "warning",
		Text:  style.Bold("Warning"),
		Secondary: []string{
			fmt.Sprintf("App (%s) will be uninstalled from %s (%s)", selection.App.AppID, selection.App.TeamDomain, selection.App.TeamID),
			"All triggers, workflows, and functions will be deleted",
			"Datastore records will be persisted",
		},
	}))

	return IO.ConfirmPrompt(ctx, "Are you sure you want to uninstall?", false)
}

// printUninstallApp displays info about removing app from API
func printUninstallApp(ctx context.Context, clients *shared.ClientFactory, appID string, teamName string) {
	_, _ = clients.IO.WriteOut().Write([]byte(fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "house",
		Text:  "App Uninstall",
		Secondary: []string{
			fmt.Sprintf(`Uninstalled the app "%s" from "%s"`, appID, teamName),
		},
	}))))
}

// printUninstallSuccess will suggest next steps and provide context
func printUninstallSuccess(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, app types.App) error {
	if app.AppID != "" {
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "broom",
			Text:  fmt.Sprintf("Run %s to fully remove your app", style.Commandf("delete", false)),
		}))
	} else {
		return runListCommand(cmd, clients)
	}
	return nil
}
