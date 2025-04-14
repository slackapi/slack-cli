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

package app

import (
	"fmt"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// NewCommand returns a new Cobra command for apps
func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app",
		Aliases: []string{"workspace", "app", "apps", "team", "teams", "workspaces"},
		Short:   "Install, uninstall, and list teams with the app installed",
		Long:    "Install, uninstall, and list teams with the app installed",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "install", Meaning: "Install a production app to a team"},
			{Command: "link", Meaning: "Link an existing app to the project"},
			{Command: "list", Meaning: "List all teams with the app installed"},
			{Command: "uninstall", Meaning: "Uninstall an app from a team"},
			{Command: "delete", Meaning: "Delete an app and app info from a team"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if cmd.CalledAs() == "workspace" {
				clients.IO.PrintInfo(ctx, false, fmt.Sprintf(
					"\n%s It looks like you used %s. This command will be deprecated in an upcoming release.\n    You can now use %s instead of %s.\n ",
					style.Emoji("bulb"),
					style.Commandf("workspace", true),
					style.Commandf("app", true),
					style.Commandf("workspace", true),
				))
			}
			return nil
		},
	}

	// Add child commands
	cmd.AddCommand(NewAddCommand(clients))
	cmd.AddCommand(NewDeleteCommand(clients))
	cmd.AddCommand(NewLinkCommand(clients))
	cmd.AddCommand(NewListCommand(clients))
	cmd.AddCommand(NewUninstallCommand(clients))

	return cmd
}
