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
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox <subcommand> [flags] --experiment=sandboxes",
		Short: "Create and manage your sandboxes",
		Long: `Create, list, or delete Slack developer sandboxes without leaving your terminal.
Use the --team flag to select the authentication to use for these commands.

Prefer a UI? Head over to https://api.slack.com/developer-program/sandboxes

New to the Developer Program? Sign up at https://api.slack.com/developer-program/join`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCreateCommand(clients))
	cmd.AddCommand(NewListCommand(clients))
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
