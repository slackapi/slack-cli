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

package env

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdotenv"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewEnvUnsetCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unset [name] [flags]",
		Aliases: []string{"remove"},
		Short:   "Unset an environment variable from the project",
		Long: strings.Join([]string{
			"Unset an environment variable from the project.",
			"",
			"If no variable name is provided, you will be prompted to select one.",
			"",
			"Commands that run in the context of a project source environment variables from",
			`the ".env" file. This includes the "run" command.`,
			"",
			`The "deploy" command gathers environment variables from the ".env" file as well`,
			"unless the app is using ROSI features.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Select an environment variable to unset",
				Command: "env unset",
			},
			{
				Meaning: "Unset an environment variable",
				Command: "env unset MAGIC_PASSWORD",
			},
		}),
		Args: cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvUnsetCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvUnsetCommandFunc(clients, cmd, args)
		},
	}

	cmd.Flags().StringVar(&variableNameFlag, "name", "", "choose the environment variable name")

	return cmd
}

// preRunEnvUnsetCommandFunc determines if the command is run in a valid project
// and configures flags
func preRunEnvUnsetCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	return cmdutil.IsValidProjectDirectory(clients)
}

// runEnvUnsetCommandFunc removes an environment variable from an app
func runEnvUnsetCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Hosted apps require selecting an app before gathering variable inputs.
	hosted := isHostedRuntime(ctx, clients)
	var selection prompts.SelectedApp
	if hosted {
		s, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
		if err != nil {
			return err
		}
		selection = s
	}

	// Get the variable name from args, or prompt from the appropriate source.
	var variableName string
	if len(args) > 0 {
		variableName = args[0]
	} else if hosted && !selection.App.IsDev {
		variables, err := clients.API().ListVariables(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
		)
		if err != nil {
			return err
		}
		if len(variables) <= 0 {
			clients.IO.PrintTrace(ctx, slacktrace.EnvUnsetSuccess)
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "evergreen_tree",
				Text:  "Environment Unset",
				Secondary: []string{
					"The app has no environment variables to remove",
				},
			}))
			return nil
		}
		selected, err := clients.IO.SelectPrompt(
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
		}
		variableName = selected.Option
	} else {
		dotEnv, err := slackdotenv.Read(clients.Fs)
		if err != nil {
			return err
		}
		if len(dotEnv) <= 0 {
			clients.IO.PrintTrace(ctx, slacktrace.EnvUnsetSuccess)
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "evergreen_tree",
				Text:  "Environment Unset",
				Secondary: []string{
					"The project has no environment variables to remove",
				},
			}))
			return nil
		}
		variables := make([]string, 0, len(dotEnv))
		for k := range dotEnv {
			variables = append(variables, k)
		}
		sort.Strings(variables)
		selected, err := clients.IO.SelectPrompt(
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
		}
		variableName = selected.Option
	}

	// Remove the environment variable using either the Slack API method or the
	// project ".env" file depending on the app hosting.
	if hosted && !selection.App.IsDev {
		err := clients.API().RemoveVariable(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
			variableName,
		)
		if err != nil {
			return err
		}
		clients.IO.PrintTrace(ctx, slacktrace.EnvUnsetSuccess)
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "evergreen_tree",
			Text:  "Environment Unset",
			Secondary: []string{
				fmt.Sprintf("Successfully removed \"%s\" as an app environment variable", variableName),
			},
		}))
	} else {
		err := slackdotenv.Unset(clients.Fs, variableName)
		if err != nil {
			return err
		}
		clients.IO.PrintTrace(ctx, slacktrace.EnvUnsetSuccess)
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "evergreen_tree",
			Text:  "Environment Unset",
			Secondary: []string{
				fmt.Sprintf("Successfully removed \"%s\" as a project environment variable", variableName),
			},
		}))
	}
	return nil
}
