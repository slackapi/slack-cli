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
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdotenv"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewEnvSetCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set [name] [value] [flags]",
		Aliases: []string{"add"},
		Short:   "Set an environment variable for the project",
		Long: strings.Join([]string{
			"Set an environment variable for the project.",
			"",
			"If a name or value is not provided, you will be prompted to provide these.",
			"",
			"Commands that run in the context of a project source environment variables from",
			`the ".env" file. This includes the "run" command.`,
			"",
			`The "deploy" command gathers environment variables from the ".env" file as well`,
			"unless the app is using ROSI features.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Prompt for an environment variable",
				Command: "env set",
			},
			{
				Meaning: "Set an environment variable",
				Command: "env set MAGIC_PASSWORD abracadbra",
			},
			{
				Meaning: "Prompt for an environment variable value",
				Command: "env set SECRET_PASSWORD",
			},
		}),
		Args: cobra.MaximumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvSetCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvSetCommandFunc(clients, cmd, args)
		},
	}

	cmd.Flags().StringVar(&variableValueFlag, "value", "", "set the environment variable value")

	return cmd
}

// preRunEnvSetCommandFunc determines if the command is run in a valid project
// and configures flags
func preRunEnvSetCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	return cmdutil.IsValidProjectDirectory(clients)
}

// runEnvSetCommandFunc sets an app environment variable to given values
func runEnvSetCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
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

	// Get the variable name from the args or prompt
	variableName := ""
	if len(args) < 1 {
		name, err := clients.IO.InputPrompt(ctx, "Variable name", iostreams.InputPromptConfig{
			Required: false,
		})
		if err != nil {
			return err
		}
		variableName = name
	} else {
		variableName = args[0]
	}

	// Get the variable value from the args or prompt
	variableValue := ""
	if len(args) < 2 {
		response, err := clients.IO.PasswordPrompt(ctx, "Variable value", iostreams.PasswordPromptConfig{
			Flag: clients.Config.Flags.Lookup("value"),
		})
		if err != nil {
			return err
		}
		variableValue = response.Value
	} else {
		variableValue = args[1]
	}

	// Add the environment variable using either the Slack API method or the
	// project ".env" file depending on the app hosting.
	var details []string
	if hosted && !selection.App.IsDev {
		err := clients.API().AddVariable(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
			variableName,
			variableValue,
		)
		if err != nil {
			return err
		}
		details = append(details, fmt.Sprintf("Successfully added \"%s\" as an app environment variable", variableName))
	} else {
		exists, err := afero.Exists(clients.Fs, ".env")
		if err != nil {
			return err
		}
		err = slackdotenv.Set(clients.Fs, variableName, variableValue)
		if err != nil {
			return err
		}
		if !exists {
			details = append(details, "Created a project .env file that shouldn't be added to version control")
		}
		details = append(details, fmt.Sprintf("Successfully added \"%s\" as a project environment variable", variableName))
	}

	clients.IO.PrintTrace(ctx, slacktrace.EnvSetSuccess)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "evergreen_tree",
		Text:      "Environment Set",
		Secondary: details,
	}))
	return nil
}
