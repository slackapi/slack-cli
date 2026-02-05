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
	"strings"

	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flags
var variableNameFlag string
var variableValueFlag string

// appSelectPromptFunc is a handle to the AppSelectPrompt that can be mocked in tests
var appSelectPromptFunc = prompts.AppSelectPrompt

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "env <subcommand>",
		Aliases: []string{"var", "vars", "variable", "variables"},
		Short:   "Add, remove, or list environment variables",
		Long: strings.Join([]string{
			"Add, remove, or list environment variables for apps deployed to Slack managed",
			"infrastructure.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
			"",
			`Explore more: {{LinkText "https://docs.slack.dev/tools/slack-cli/guides/using-environment-variables-with-the-slack-cli"}}`,
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Add an environment variable",
				Command: "env add MAGIC_PASSWORD abracadbra",
			},
			{
				Meaning: "List all environment variables",
				Command: "env list",
			},
			{
				Meaning: "Remove an environment variable",
				Command: "env remove MAGIC_PASSWORD",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add child commands
	cmd.AddCommand(NewEnvAddCommand(clients))
	cmd.AddCommand(NewEnvListCommand(clients))
	cmd.AddCommand(NewEnvRemoveCommand(clients))

	return cmd
}
