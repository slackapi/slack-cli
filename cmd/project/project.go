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

package project

import (
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// NewCommand returns a Cobra command for the project parent command
func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project <subcommand>",
		Aliases: []string{},
		Short:   "Create, manage, and doctor a project",
		Long: strings.Join([]string{
			`Create, manage, and doctor a project and its configuration files.`,
			``,
			`Get started by creating a new project using the {{ToBold "create"}} command.`,
			``,
			`Initialize an existing project with CLI support using the {{ToBold "init"}} command.`,
			``,
			`Check your project health and diagnose problems with the {{ToBold "doctor"}} command.`,
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Creates a new Slack project from an optional template",
				Command: "project create",
			},
			{
				Meaning: "Initialize an existing project to work with the Slack CLI",
				Command: "project init",
			},
			{
				Meaning: "Creates a new Slack project from the sample gallery",
				Command: "project samples",
			},
		}),
		Args: cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add child commands
	cmd.AddCommand(NewCreateCommand(clients))
	cmd.AddCommand(NewInitCommand(clients))
	cmd.AddCommand(NewSamplesCommand(clients))

	return cmd
}
