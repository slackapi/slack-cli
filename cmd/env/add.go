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

func NewEnvAddCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <value> [flags]",
		Short: "Add an environment variable to the app",
		Long: strings.Join([]string{
			"Add an environment variable to an app deployed to Slack managed infrastructure.",
			"",
			"If a name or value is not provided, you will be prompted to provide these.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Prompt for an environment variable",
				Command: "env add",
			},
			{
				Meaning: "Add an environment variable",
				Command: "env add MAGIC_PASSWORD abracadbra",
			},
			{
				Meaning: "Prompt for an environment variable value",
				Command: "env add SECRET_PASSWORD",
			},
		}),
		Args: cobra.MaximumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvAddCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvAddCommandFunc(clients, cmd, args)
		},
	}

	cmd.Flags().StringVar(&variableValueFlag, "value", "", "set the environment variable value")

	return cmd
}

// preRunEnvAddCommandFunc determines if the command is supported for a project
// and configures flags
func preRunEnvAddCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

// runEnvAddCommandFunc sets an app environment variable to given values
func runEnvAddCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get the workspace from the flag or prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Get the variable name from the args or prompt
	var variableName string
	if len(args) < 1 {
		variableName, err = clients.IO.InputPrompt(ctx, "Variable name", iostreams.InputPromptConfig{
			Required: false,
		})
		if err != nil {
			return err
		}
	} else {
		variableName = args[0]

		// Display the variable name before getting the variable value
		if len(args) < 2 && !clients.Config.Flags.Lookup("value").Changed {
			mimickedInput := iostreams.MimicInputPrompt("Variable name", variableName)
			clients.IO.PrintInfo(ctx, false, mimickedInput)
		}
	}

	// Get the variable value from the args or prompt
	var variableValue string
	if len(args) < 2 {
		response, err := clients.IO.PasswordPrompt(ctx, "Variable value", iostreams.PasswordPromptConfig{
			Flag: clients.Config.Flags.Lookup("value"),
		})
		if err != nil {
			return err
		} else {
			variableValue = response.Value
		}
	} else {
		variableValue = args[1]
	}

	err = clients.API().AddVariable(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		variableName,
		variableValue,
	)
	if err != nil {
		return err
	}

	clients.IO.PrintTrace(ctx, slacktrace.EnvAddSuccess)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "evergreen_tree",
		Text:  "App Environment",
		Secondary: []string{
			fmt.Sprintf(
				"Successfully added \"%s\" as an environment variable",
				variableName,
			),
		},
	}))
	return nil
}
