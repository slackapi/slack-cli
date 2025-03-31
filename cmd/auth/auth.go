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

package auth

import (
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <subcommand> [flags]",
		Short: "Add and remove local team authorizations",
		Long:  `Add and remove local team authorizations`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "auth list", Meaning: "List all authorized accounts"},
			{Command: "auth login", Meaning: "Log in to a Slack account"},
			{Command: "auth logout", Meaning: "Log out of a team"},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	// Add child commands
	cmd.AddCommand(NewListCommand(clients))
	cmd.AddCommand(NewLoginCommand(clients))
	cmd.AddCommand(NewLogoutCommand(clients))
	cmd.AddCommand(NewRevokeCommand(clients))
	cmd.AddCommand(NewTokenCommand(clients))

	return cmd
}
