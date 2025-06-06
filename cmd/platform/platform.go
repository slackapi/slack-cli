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

package platform

import (
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "platform <subcommand> [flags]",
		Short: "Deploy and run apps on the Slack Platform",
		Long:  `Deploy and run apps on the Slack Platform`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "run", Meaning: "Run an app locally in a workspace"},
			{Command: "deploy --team T0123456", Meaning: "Deploy to a specific team"},
			{Command: "activity -t", Meaning: "Continuously poll for new activity logs"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add child commands
	cmd.AddCommand(NewActivityCommand(clients))
	cmd.AddCommand(NewDeployCommand(clients))
	cmd.AddCommand(NewRunCommand(clients))

	return cmd
}
