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

package env

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewEnvRemoveCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name> [flags]",
		Short: "Remove an environment variable from the app",
		Long: strings.Join([]string{
			"Remove an environment variable from an app deployed to Slack managed",
			"infrastructure.",
			"",
			"If no variable name is provided, you will be prompted to select one.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Select an environment variable to remove",
				Command: "env remove",
			},
			{
				Meaning: "Remove an environment variable",
				Command: "env remove MAGIC_PASSWORD",
			},
		}),
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvRemoveCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvRemoveCommandFunc(clients, cmd, args)
		},
	}

	cmd.Flags().StringVar(&variableNameFlag, "name", "", "choose the environment variable name")

	return cmd
}

// preRunEnvRemoveCommandFunc determines if the command is supported for a project
// and configures flags
func preRunEnvRemoveCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return err
	}
	if clients.Config.ForceFlag {
		return nil
	}
	return cmdutil.IsSlackHostedProject(ctx, clients)
}

// runEnvRemoveCommandFunc removes an environment variable from an app
func runEnvRemoveCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	var ctx = cmd.Context()

	// Get the workspace from the flag or prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Get the variable name from the flags, args, or select from the environment
	var variableName string
	if len(args) > 0 {
		variableName = args[0]
	} else {
		variables, err := clients.API().ListVariables(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
		)
		if err != nil {
			return err
		}
		if len(variables) <= 0 {
			clients.IO.PrintTrace(ctx, slacktrace.EnvRemoveSuccess)
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "evergreen_tree",
				Text:  "App Environment",
				Secondary: []string{
					"The app has no environment variables to remove",
				},
			}))
			return nil
		}
		selection, err := clients.IO.SelectPrompt(
			ctx,
			"Select a variable to remove",
			variables,
			iostreams.SelectPromptConfig{
				Flag:     clients.Config.Flags.Lookup("name"),
				Required: true,
			},
		)
		if err != nil {
			return err
		} else {
			variableName = selection.Option
		}
	}

	err = clients.API().RemoveVariable(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		variableName,
	)
	if err != nil {
		return err
	}

	clients.IO.PrintTrace(ctx, slacktrace.EnvRemoveSuccess)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "evergreen_tree",
		Text:  "App Environment",
		Secondary: []string{
			fmt.Sprintf(
				"Successfully removed \"%s\" from the app's environment variables",
				variableName,
			),
		},
	}))

	return nil
}
