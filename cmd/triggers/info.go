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
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type infoCmdFlags struct {
	triggerID string
}

var infoFlags infoCmdFlags

var infoAppSelectPromptFunc = prompts.AppSelectPrompt

func NewInfoCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info --trigger-id <id>",
		Short: "Get details for a specific trigger",
		Long:  "Get details for a specific trigger",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger info --trigger-id Ft01234ABCD", Meaning: "Get details for a specific trigger in a selected workspace"},
			{Command: "trigger info --trigger-id Ft01234ABCD --app A0123456", Meaning: "Get details for a specific trigger"},
		}),
		Aliases: []string{"information", "show"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfoCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&infoFlags.triggerID, "trigger-id", "", "the ID of the trigger")

	return cmd
}

func runInfoCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(ctx, "cmd.triggers.info")
	defer span.Finish()

	// Get the app from the flag or prompt
	selection, err := infoAppSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	token := selection.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selection.App

	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}

	if infoFlags.triggerID == "" {
		infoFlags.triggerID, err = promptForTriggerID(ctx, cmd, clients, app, token, defaultLabels)
		if err != nil {
			if slackerror.ToSlackError(err).Code == slackerror.ErrNoTriggers {
				printNoTriggersMessage(ctx, clients.IO)
				return nil
			}
			return err
		}
	}

	requestedTrigger, err := clients.API().WorkflowsTriggersInfo(ctx, token, infoFlags.triggerID)
	if err != nil {
		return err
	}

	cmd.Printf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  "Trigger Info",
	}))
	trigs, err := sprintTrigger(ctx, requestedTrigger, clients, true, app)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", strings.Join(trigs, "\n"))
	cmd.Println()
	return nil
}
