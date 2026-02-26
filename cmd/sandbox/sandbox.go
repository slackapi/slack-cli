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
		Use:   "sandbox <subcommand> [flags]",
		Short: "Create and manage your sandboxes",
		Long: `Create and manage your Slack developer sandboxes.

Developer Sandboxes are Enterprise Grid environments for building and testing Slack apps.
Use these commands to provision, list, and delete sandboxes without leaving your terminal.

Gated behind the manage-sandboxes experiment: --experiment=manage-sandboxes`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox create --name my-test-app", Meaning: "Create a new sandbox workspace"},
			{Command: "sandbox list", Meaning: "List all your sandboxes"},
			{Command: "sandbox delete T0123456", Meaning: "Delete a sandbox by team ID"},
		}),
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
	if !clients.Config.WithExperimentOn(experiment.ManageSandboxes) {
		return slackerror.New(slackerror.ErrMissingExperiment).
			WithMessage("%s The sandbox management commands are under construction!", style.Emoji("beach_with_umbrella")).
			WithRemediation("Feel free to try them out; just add the --experiment=manage-sandboxes flag to your command!")
	}
	return nil
}
