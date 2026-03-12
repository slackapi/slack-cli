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

package sandbox

import (
	"context"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox <subcommand> [flags] --experiment=sandboxes",
		Short: "Manage developer sandboxes",
		Long: `Manage Slack developer sandboxes without leaving your terminal.
Use the --team flag to select the authentication to use for these commands.

Prefer a UI? Head over to
{{LinkText "https://api.slack.com/developer-program/sandboxes"}}

New to the Developer Program? Sign up at
{{LinkText "https://api.slack.com/developer-program/join"}}`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{}),
		Aliases: []string{"sandboxes"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewListCommand(clients))
	cmd.AddCommand(NewCreateCommand(clients))
	cmd.AddCommand(NewDeleteCommand(clients))

	return cmd
}

func requireSandboxExperiment(clients *shared.ClientFactory) error {
	if !clients.Config.WithExperimentOn(experiment.Sandboxes) {
		return slackerror.New(slackerror.ErrMissingExperiment).
			WithMessage("%sThe sandbox management commands are under construction", style.Emoji("construction")).
			WithRemediation("To try them out, just add the --experiment=sandboxes flag to your command!")
	}
	return nil
}

// getSandboxAuth returns the auth to be used for sandbox management.
// Uses the global --token or --team flag if present, otherwise prompts the user to select a team.
func getSandboxAuth(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, error) {
	// Check for the global --token flag
	if clients.Config.TokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return nil, err
		}
		return &auth, nil
	}

	// Prompt the user to select a team to use for authentication
	clients.IO.PrintInfo(ctx, false, "%s", style.Secondary("Choose a Slack team where your email address matches your Slack developer account"))
	auth, err := prompts.PromptTeamSlackAuth(ctx, clients, "Select a team for authentication")
	if err != nil {
		return nil, err
	}

	return auth, nil
}
