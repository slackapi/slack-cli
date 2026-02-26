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
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type deleteFlags struct {
	sandboxType string
	force       bool
	yes         bool
	token       string
}

var deleteCmdFlags deleteFlags

func NewDeleteCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <sandbox-team-id> [flags]",
		Short: "Delete a Developer Sandbox",
		Long: `Permanently delete a Developer Sandbox.

Requires confirmation unless --force or --yes is used.`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox delete T0123456", Meaning: "Delete a sandbox by team ID"},
			{Command: "sandbox delete T0123456 --force", Meaning: "Delete without confirmation"},
		}),
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(cmd, clients, args[0])
		},
	}

	cmd.Flags().StringVar(&deleteCmdFlags.sandboxType, "sandbox-type", "", "Type of sandbox: regular or auto (default: regular)")
	cmd.Flags().BoolVar(&deleteCmdFlags.force, "force", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&deleteCmdFlags.yes, "yes", false, "Auto-confirm deletion")
	cmd.Flags().StringVar(&deleteCmdFlags.token, "token", "", "Service account token for CI/CD authentication")

	return cmd
}

func runDeleteCommand(cmd *cobra.Command, clients *shared.ClientFactory, sandboxTeamID string) error {
	ctx := cmd.Context()

	token, err := getSandboxToken(ctx, clients, deleteCmdFlags.token)
	if err != nil {
		return err
	}

	skipConfirm := deleteCmdFlags.force || deleteCmdFlags.yes
	if !skipConfirm {
		proceed, err := clients.IO.ConfirmPrompt(ctx, "Are you sure you want to permanently delete this sandbox? This cannot be undone.", false)
		if err != nil {
			if slackerror.Is(err, slackerror.ErrProcessInterrupted) {
				clients.IO.SetExitCode(iostreams.ExitCancel)
			}
			return err
		}
		if !proceed {
			clients.IO.PrintInfo(ctx, false, "\n%s\n", style.Sectionf(style.TextSection{
				Emoji: "thumbs_up",
				Text:  "Deletion cancelled",
			}))
			return nil
		}
	}

	if err := clients.API().DeleteSandbox(ctx, token, sandboxTeamID, deleteCmdFlags.sandboxType); err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s\n", style.Sectionf(style.TextSection{
		Emoji: "white_check_mark",
		Text:  "Sandbox deleted",
		Secondary: []string{
			"Sandbox " + sandboxTeamID + " has been permanently deleted.",
		},
	}))

	return nil
}
