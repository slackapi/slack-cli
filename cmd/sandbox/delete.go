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
	"fmt"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type deleteFlags struct {
	sandboxID string
	force     bool
}

var deleteCmdFlags deleteFlags

func NewDeleteCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [flags]",
		Short: "Delete a developer sandbox",
		Long:  `Permanently delete a sandbox and all of its data`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox delete --sandbox-id E0123456", Meaning: "Delete a sandbox identified by its team ID"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&deleteCmdFlags.sandboxID, "sandbox-id", "", "Sandbox team ID to delete")
	cmd.Flags().BoolVar(&deleteCmdFlags.force, "force", false, "Skip confirmation prompt")

	if err := cmd.MarkFlagRequired("sandbox-id"); err != nil {
		panic(err)
	}

	return cmd
}

func runDeleteCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	auth, err := getSandboxAuth(ctx, clients)
	if err != nil {
		return err
	}

	skipConfirm := deleteCmdFlags.force
	if !skipConfirm {
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "warning",
			Text:  style.Bold(" Danger zone"),
			Secondary: []string{
				fmt.Sprintf("Sandbox (%s) and all of its data will be permanently deleted", deleteCmdFlags.sandboxID),
				"This cannot be undone",
			},
		}))

		proceed, err := clients.IO.ConfirmPrompt(ctx, "Are you sure you want to delete the sandbox?", false)
		if err != nil {
			if slackerror.Is(err, slackerror.ErrProcessInterrupted) {
				clients.IO.SetExitCode(iostreams.ExitCancel)
			}
			return err
		}
		if !proceed {
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "thumbs_up",
				Text:  "Deletion cancelled",
			}))
			return nil
		}
	}

	if err := clients.API().DeleteSandbox(ctx, auth.Token, deleteCmdFlags.sandboxID); err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "white_check_mark",
		Text:  "Sandbox Deleted",
		Secondary: []string{
			"Sandbox " + deleteCmdFlags.sandboxID + " has been permanently deleted",
		},
	}))

	err = printSandboxes(cmd, clients, auth.Token, auth)
	if err != nil {
		return err
	}

	return nil
}
