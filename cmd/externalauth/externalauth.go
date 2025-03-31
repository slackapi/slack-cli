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

package externalauth

import (
	"strings"

	"github.com/slackapi/slack-cli/internal/pkg/externalauth"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var providerFlag string
var providerSelectionFunc = externalauth.ProviderSelectPrompt
var tokenSelectionFunc = externalauth.TokenSelectPrompt
var appSelectPromptFunc = prompts.AppSelectPrompt

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-auth <subcommand>",
		Short: "Adjust settings of external authentication providers",
		Long: strings.Join([]string{
			"Adjust external authorization and authentication providers of a workflow app.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
			"",
			`Explore providers: {{LinkText "https://api.slack.com/automation/external-auth"}}`,
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Initiate OAuth2 flow for a selected provider",
				Command: "external-auth add",
			},
			{
				Meaning: "Set client secret for an app and provider",
				Command: "external-auth add-secret",
			},
			{
				Meaning: "Remove authorization for a specific provider",
				Command: "external-auth remove",
			},
			{
				Meaning: "Select authorization for a specific provider in a workflow",
				Command: "external-auth select-auth",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add child commands
	cmd.AddCommand(NewAddCommand(clients))
	cmd.AddCommand(NewRemoveCommand(clients))
	cmd.AddCommand(NewAddClientSecretCommand(clients))
	cmd.AddCommand(NewSelectAuthCommand(clients))

	return cmd
}
