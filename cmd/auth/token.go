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

// NewTokenCommand prepares the process to perform authentication for a service
// token via the existing flow used by the login command
func NewTokenCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Collect a service token",
		Long:  "Log in to a Slack account in your team",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "auth token", Meaning: "Create a service token with prompts"},
			{Command: "auth token --no-prompt", Meaning: "Gather a service token without prompts, this returns a ticket"},
			{Command: "auth token --challenge 6d0a31c9 --ticket ISQWLiZT0tOMLO3YWNTJO0...", Meaning: "Complete authentication using a ticket and challenge code"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceTokenFlag = true
			_, err := RunLoginCommand(clients, cmd)
			return err
		},
	}

	// Support login in promptless fashion
	cmd.Flags().BoolVarP(&noPromptFlag, "no-prompt", "", false, "login without prompts using ticket and challenge code")
	cmd.Flags().StringVarP(&ticketArg, "ticket", "", "", "provide an auth ticket value")
	cmd.Flags().StringVarP(&challengeCodeArg, "challenge", "", "", "provide a challenge code for pre-authenticated login")

	return cmd
}
