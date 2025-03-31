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

package collaborators

import (
	"context"
	"fmt"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// teamAppSelectPromptFunc is a handle to the TeamAppSelectPrompt that can be mocked in tests
var teamAppSelectPromptFunc = prompts.TeamAppSelectPrompt

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collaborator <subcommand> [flags]",
		Aliases: []string{"collaborators", "owners"},
		Short:   "Manage app collaborators",
		Long:    "Manage app collaborators",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "collaborator add bots@slack.com", Meaning: "Add a collaborator from email"},
			{Command: "collaborator list", Meaning: "List all of the collaborators"},
			{Command: "collaborator remove USLACKBOT", Meaning: "Remove a collaborator by user ID"},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	// Add child commands
	cmd.AddCommand(NewAddCommand(clients))
	cmd.AddCommand(NewListCommand(clients))
	cmd.AddCommand(NewRemoveCommand(clients))
	cmd.AddCommand(NewUpdateCommand(clients))

	return cmd
}

// printSuccess displays a string for user feedback when the operation has completed successfully
func printSuccess(ctx context.Context, io iostreams.IOStreamer, user types.SlackUser, actionStr string) {
	successText := fmt.Sprintf(
		"\n%s%s successfully %s as %s on this app",
		style.Emoji("sparkles"),
		user.ShorthandF(),
		actionStr,
		user.PermissionType.AppCollaboratorPermissionF(),
	)

	io.PrintInfo(ctx, false, "%s\n", successText)
}
