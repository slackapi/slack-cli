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
	"fmt"
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var settingsAppSelectPromptFunc = prompts.AppSelectPrompt

func NewSettingsCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings [flags]",
		Short: "Open app settings for configurations",
		Long: strings.Join([]string{
			"Open app settings to configure an application in a web browser.",
			"",
			"Discovering new features and customizing an app manifest can be done from this",
			fmt.Sprintf("web interface for apps with a \"%s\" manifest source.", config.ManifestSourceRemote.String()),
			"",
			"This command does not support apps deployed to Run on Slack infrastructure.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Open app settings dashboard",
				Command: "app settings",
			},
			{
				Meaning: "Open app settings for a specific app",
				Command: "app settings --app A0123456789",
			},
		}),
		Args: cobra.MaximumNArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return appSettingsCommandPreRunE(clients, cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return appSettingsCommandRunE(clients, cmd, args)
		},
	}
	return cmd
}

// appSettingsCommandPreRunE determines if the command can be run in a project
// or if the command is run outside of a project
func appSettingsCommandPreRunE(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return nil
	}
	// Allow the force flag to ignore hosted apps and try to open app settings
	if clients.Config.ForceFlag {
		return nil
	}
	err = cmdutil.IsSlackHostedProject(ctx, clients)
	if err != nil {
		if slackerror.Is(err, slackerror.ErrAppNotHosted) {
			return nil
		} else {
			return err
		}
	}
	return slackerror.New(slackerror.ErrAppHosted).
		WithDetails(slackerror.ErrorDetails{
			{
				Message: "App settings is not supported with Run on Slack infrastructure",
			},
		})
}

// appSettingsCommandRunE opens app settings in a browser for the selected app
func appSettingsCommandRunE(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	clients.IO.PrintTrace(ctx, slacktrace.AppSettingsStart)

	app, err := settingsAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		// If no apps exist, open the list of all apps known to the developer
		if slackerror.Is(err, slackerror.ErrInstallationRequired) {
			// Clean up any empty .slack directory and files created during app selection
			clients.AppClient().CleanUp()

			host := clients.API().Host()
			parsed, err := url.Parse(host)
			if err != nil {
				return err
			}
			parsed.Host = "api." + parsed.Host
			settingsURL := fmt.Sprintf("%s/apps", parsed.String())

			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "house",
				Text:  "App Settings",
				Secondary: []string{
					settingsURL,
				},
			}))
			clients.Browser().OpenURL(settingsURL)

			clients.IO.PrintTrace(ctx, slacktrace.AppSettingsSuccess, settingsURL)
			return nil
		}
		return err
	}
	host := clients.API().Host()
	parsed, err := url.Parse(host)
	if err != nil {
		return err
	}
	parsed.Host = "api." + parsed.Host
	settingsURL := fmt.Sprintf("%s/apps/%s", parsed.String(), app.App.AppID)

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "house",
		Text:  "App Settings",
		Secondary: []string{
			settingsURL,
		},
	}))
	clients.Browser().OpenURL(settingsURL)

	clients.IO.PrintTrace(ctx, slacktrace.AppSettingsSuccess, settingsURL)
	return nil
}
