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

package triggers

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type deleteCmdFlags struct {
	triggerID string
}

var deleteFlags deleteCmdFlags

// TODO(mcodik) figure out a way to mock this more nicely
var deleteAppSelectPromptFunc = prompts.AppSelectPrompt

// NewDeleteCommand creates a new Cobra command instance
func NewDeleteCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete --trigger-id <id>",
		Short: "Delete an existing trigger",
		Long:  `Delete an existing trigger`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger delete --trigger-id Ft01234ABCD", Meaning: "Delete a specific trigger in a selected workspace"},
			{Command: "trigger delete --trigger-id Ft01234ABCD --app A0123456", Meaning: "Delete a specific trigger for an app"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(clients, cmd)
		},
	}

	cmd.Flags().StringVar(&deleteFlags.triggerID, "trigger-id", "", "the ID of the trigger")

	return &cmd
}

func runDeleteCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(ctx, "cmd.triggers.delete")
	defer span.Finish()

	// Get the app from the flag or prompt
	selection, err := deleteAppSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	token := selection.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selection.App

	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}

	if deleteFlags.triggerID == "" {
		deleteFlags.triggerID, err = promptForTriggerID(ctx, cmd, clients, app, token, defaultLabels)
		if err != nil {
			if slackerror.ToSlackError(err).Code == slackerror.ErrNoTriggers {
				printNoTriggersMessage(ctx, clients.IO)
				return nil
			}
			return err
		}
	}

	err = clients.API().WorkflowsTriggersDelete(ctx, token, deleteFlags.triggerID)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "wastebasket",
		Text:  fmt.Sprintf("Trigger '%s' deleted", deleteFlags.triggerID),
	}))
	return nil
}
