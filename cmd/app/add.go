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

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
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
var runAddCommandFunc = RunAddCommand
var appInstallProdAppFunc = apps.Add
var appInstallDevAppFunc = apps.InstallLocalApp
var appSelectPromptFunc = prompts.AppSelectPrompt

// Flags

type addCmdFlags struct {
	orgGrantWorkspaceID string
	environmentFlag     string
}

var addFlags addCmdFlags

// NewAddCommand returns a new Cobra command
func NewAddCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install [flags]",
		Aliases: []string{"add"},
		Short:   "Install the app to a team",
		Long:    "Install the app to a team",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app install", Meaning: "Install a production app to a team"},
			{Command: "app install --team T0123456 --environment deployed", Meaning: "Install a production app to a specific team"},
			{Command: "app install --team T0123456 --environment local", Meaning: "Install a local dev app to a specific team"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunAddCommand(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			_, _, appInstance, err := runAddCommandFunc(ctx, clients, nil, addFlags.orgGrantWorkspaceID)
			if err != nil {
				return err
			}
			return printAddSuccess(clients, cmd, appInstance)
		},
	}

	cmd.Flags().StringVar(&addFlags.orgGrantWorkspaceID, cmdutil.OrgGrantWorkspaceFlag, "", cmdutil.OrgGrantWorkspaceDescription())
	cmd.Flags().StringVarP(&addFlags.environmentFlag, "environment", "E", "", "environment of app (local, deployed)")

	return cmd
}

// preRunAddCommand confirms an app is available for installation
func preRunAddCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return err
	}
	if !clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		return nil
	}
	clients.Config.SetFlags(cmd)
	return nil
}

// RunAddCommand executes the workspace install command, prints output, and returns any errors.
func RunAddCommand(ctx context.Context, clients *shared.ClientFactory, selection *prompts.SelectedApp, orgGrantWorkspaceID string) (context.Context, types.InstallState, types.App, error) {
	if selection == nil {
		// TODO: Move to the promptIsProduction when the prompt is refactored and tested.
		// Validate that the --app flag is not an app ID when the --environment flag is set.
		if types.IsAppID(clients.Config.AppFlag) && addFlags.environmentFlag != "" {
			return ctx, "", types.App{}, slackerror.New(slackerror.ErrMismatchedFlags).WithRemediation("When '--app <app_id>' is set, please do not set the flag --environment.")
		}

		// TODO: Move to the promptIsProduction when the prompt is refactored and tested.
		// Validate that the --environment flag matches the --app flag, when the value is `--app local` or `--app deployed`.
		if types.IsAppFlagEnvironment(clients.Config.AppFlag) {
			if addFlags.environmentFlag != "" && addFlags.environmentFlag != clients.Config.AppFlag {
				return ctx, "", types.App{}, slackerror.New(slackerror.ErrMismatchedFlags).WithRemediation("When '--app local' or '--app deployed' is set, please set the flag --environment to match the --app flag.")
			}

			if addFlags.environmentFlag == "" {
				err := clients.Config.Flags.Lookup("environment").Value.Set(clients.Config.AppFlag)
				if err != nil {
					return ctx, "", types.App{}, err
				}
				clients.Config.Flags.Lookup("environment").Changed = true
			}
		}

		// Default to `--environment deployed` when there is no `--environment` flag and `--team <id>` is set.
		// Skip when `--app <id>` flag is set, because the environment is looked up in the app selector prompt.
		// TODO(semver:major): This is backwards compatibility for when `install` only supported deployed environments.
		if !types.IsAppID(clients.Config.AppFlag) && (addFlags.environmentFlag == "" && clients.Config.TeamFlag != "") {
			err := clients.Config.Flags.Lookup("environment").Value.Set("deployed")
			if err != nil {
				return ctx, "", types.App{}, err
			}
			clients.Config.Flags.Lookup("environment").Changed = true

			clients.IO.PrintInfo(ctx, false, "\n"+style.Sectionf(style.TextSection{
				Emoji: "warning",
				Text:  "Warning: Default App Environment",
				Secondary: []string{
					"App environment is set to deployed when only the --team flag is provided.",
					"The next major version will change this behavior.",
					"When the --team flag is provided, the --environment flag will be required.",
					"Add the '--environment deployed' to avoid breaking changes.",
				},
			}))
		}

		// When the app flag is an app ID, the app select prompt can resolve the app.
		// Otherwise, prompt for the app environment and app.
		if types.IsAppID(clients.Config.AppFlag) {
			selected, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowAllApps)
			if err != nil {
				return ctx, "", types.App{}, err
			}
			selection = &selected
		} else {
			// Prompt for deployed or local app environment.
			isProductionApp, err := promptIsProduction(ctx, clients)
			if err != nil {
				return ctx, "", types.App{}, err
			}

			// Set the app environment type based on the prompt.
			var appEnvironmentType prompts.AppEnvironmentType
			if isProductionApp {
				appEnvironmentType = prompts.ShowHostedOnly
			} else {
				appEnvironmentType = prompts.ShowLocalOnly
			}

			selected, err := appSelectPromptFunc(ctx, clients, appEnvironmentType, prompts.ShowAllApps)
			if err != nil {
				return ctx, "", types.App{}, err
			}
			selection = &selected

			if !isProductionApp {
				selection.App.IsDev = true
			}
		}
	}

	if selection.Auth.TeamDomain == "" {
		return ctx, "", types.App{}, slackerror.New(slackerror.ErrCredentialsNotFound)
	}

	var err error
	orgGrantWorkspaceID, err = prompts.ValidateGetOrgWorkspaceGrant(ctx, clients, selection, orgGrantWorkspaceID, true /* top prompt option should be 'all workspaces' */)
	if err != nil {
		return ctx, "", types.App{}, err
	}

	clients.Config.ManifestEnv = app.SetManifestEnvTeamVars(clients.Config.ManifestEnv, selection.App.TeamDomain, selection.App.IsDev)

	// Set up event logger
	log := newAddLogger(clients, selection.Auth.TeamDomain)

	// Install dev app or prod app to a workspace
	installedApp, installState, err := appInstall(ctx, clients, log, selection, orgGrantWorkspaceID)
	if err != nil {
		return ctx, installState, types.App{}, err // pass the installState because some callers may use it to handle the error
	}

	// Update the context with the token
	ctx = config.SetContextToken(ctx, selection.Auth.Token)

	return ctx, installState, installedApp, nil
}

// newAddLogger creates a logger instance to receive event notifications
func newAddLogger(clients *shared.ClientFactory, envName string) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			teamName := event.DataToString("teamName")
			appName := event.DataToString("appName")
			switch event.Name {
			case "app_install_manifest":
				// Ignore this event and format manifest outputs in create/update events
			case "app_install_manifest_create":
				_, _ = clients.IO.WriteOut().Write([]byte(style.Sectionf(style.TextSection{
					Emoji: "books",
					Text:  "App Manifest",
					Secondary: []string{
						fmt.Sprintf(`Creating app manifest for "%s" in "%s"`, appName, teamName),
					},
				})))
			case "app_install_manifest_update":
				_, _ = clients.IO.WriteOut().Write([]byte("\n" + style.Sectionf(style.TextSection{
					Emoji: "books",
					Text:  "App Manifest",
					Secondary: []string{
						fmt.Sprintf(`Updated app manifest for "%s" in "%s"`, appName, teamName),
					},
				})))
			case "app_install_start":
				_, _ = clients.IO.WriteOut().Write([]byte("\n" + style.Sectionf(style.TextSection{
					Emoji: "house",
					Text:  "App Install",
					Secondary: []string{
						fmt.Sprintf(`Installing "%s" app to "%s"`, appName, teamName),
					},
				})))
			case "app_install_icon_success":
				iconPath := event.DataToString("iconPath")
				_, _ = clients.IO.WriteOut().Write([]byte(
					style.SectionSecondaryf("Updated app icon: %s", iconPath),
				))
			case "app_install_icon_error":
				iconError := event.DataToString("iconError")
				_, _ = clients.IO.WriteOut().Write([]byte(
					style.SectionSecondaryf("Error updating app icon: %s", iconError),
				))
			case "app_install_complete":
				_, _ = clients.IO.WriteOut().Write([]byte(
					style.SectionSecondaryf("Finished in %s", event.DataToString("installTime")),
				))
			default:
				// Ignore the event
			}
		},
	)
}

// printAddSuccess will print a list of the environments
func printAddSuccess(clients *shared.ClientFactory, cmd *cobra.Command, appInstance types.App) error {
	return runListCommand(cmd, clients)
}

// appInstall will install an app to a team. It supports both local and deployed app types.
func appInstall(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, selection *prompts.SelectedApp, orgGrantWorkspaceID string) (types.App, types.InstallState, error) {
	if selection != nil && selection.App.IsDev {
		// Install local dev app to a team
		installedApp, _, installState, err := appInstallDevAppFunc(ctx, clients, "", log, selection.Auth, selection.App)
		return installedApp, installState, err
	} else {
		installState, installedApp, err := appInstallProdAppFunc(ctx, clients, log, selection.Auth, selection.App, orgGrantWorkspaceID)
		return installedApp, installState, err
	}
}
