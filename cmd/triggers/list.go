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
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// TODO - Find best practice, such as using an Interface and Struct to create a client
var listAppSelectPromptFunc = prompts.AppSelectPrompt

type listCmdFlags struct {
	triggerLimit int
	triggerType  string
}

var listFlags listCmdFlags

func NewListCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List details of existing triggers",
		Long:  "List details of existing triggers",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger list", Meaning: "List details for all existing triggers"},
			{Command: "trigger list --team T0123456 --app local", Meaning: "List triggers for a specific app"},
		}),
		Aliases: []string{"all"},
		Args:    cobra.NoArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	cmd.Flags().IntVarP(&listFlags.triggerLimit, "limit", "L", 4, "Limit the number of triggers to show")
	cmd.Flags().StringVarP(&listFlags.triggerType, "type", "T", "all", "Only display triggers of the given type, can be one of 'all', 'shortcut', 'event', 'scheduled', 'webhook', and 'external'")

	return cmd
}

// runListCommand will execute the list command
func runListCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(cmd.Context(), "cmd.triggers.list")
	defer span.Finish()

	// Get the app selection and accompanying auth from the flag or prompt
	selection, err := listAppSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	token := selection.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selection.App

	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}

	if listFlags.triggerLimit == 0 {
		listFlags.triggerLimit = 4
	}

	args := api.TriggerListRequest{
		AppId: app.AppID,
		Limit: listFlags.triggerLimit,
		Type:  listFlags.triggerType,
	}
	deployedTriggers, cursor, err := clients.ApiInterface().WorkflowsTriggersList(ctx, token, args)
	if err != nil {
		return err
	}

	var triggers = []types.DeployedTrigger{}
	for _, t := range deployedTriggers {
		if t.Workflow.AppID == app.AppID {
			triggers = append(triggers, t)
		}
	}

	return outputTriggersList(ctx, triggers, cmd, clients, app, cursor, listFlags.triggerType)
}

func outputTriggersList(ctx context.Context, triggers []types.DeployedTrigger, cmd *cobra.Command, clients *shared.ClientFactory, app types.App, cursor string, triggerType string) error {
	var triggersList = []string{}

	if len(triggers) == 0 {
		noTriggersMessage := style.Indent(style.Secondary("There are no triggers installed for the app"))
		if triggerType != "" && triggerType != "all" {
			noTriggersMessage = fmt.Sprintf(style.Indent(style.Secondary("There are no %s triggers installed for the app")), triggerType)
		}

		triggersList = append(triggersList, noTriggersMessage)
	}

	trigs, err := sprintTriggers(ctx, triggers, clients, app)
	if err != nil {
		return err
	}
	triggersList = append(triggersList, trigs...)

	if cmd != nil {
		cmd.Printf("\n%s", style.Sectionf(style.TextSection{
			Emoji: "zap",
			Text:  "Listing triggers installed to the app...",
		}))
		cmd.Printf("%s\n", strings.Join(triggersList, "\n"))
		cmd.Println()
	} else {
		fmt.Printf("\n%s", style.Sectionf(style.TextSection{
			Emoji: "zap",
			Text:  "Listing triggers installed to the app...",
		}))
		fmt.Printf("%s\n", strings.Join(triggersList, "\n"))
		fmt.Println()
	}

	// It's safe to add check here instead of putting into cmd != nil block as generate only create 1 trigger
	if cursor != "" {
		return showMoreTriggers(ctx, cmd, clients, app, cursor)
	}
	return nil
}

func showMoreTriggers(ctx context.Context, cmd *cobra.Command, clients *shared.ClientFactory, app types.App, cursor string) error {
	var triggersList = []string{}

	token := config.GetContextToken(ctx)
	args := api.TriggerListRequest{
		AppId:  app.AppID,
		Limit:  listFlags.triggerLimit,
		Cursor: cursor,
		Type:   listFlags.triggerType,
	}

	proceed, err := clients.IO.ConfirmPrompt(ctx, "Show more triggers?", false)
	if err != nil {
		return err
	}

	for proceed && args.Cursor != "" {
		proceed = false
		deployedTriggers, nextCursor, err := clients.ApiInterface().WorkflowsTriggersList(ctx, token, args)
		if err != nil {
			return err
		}

		trigs, err := sprintTriggers(ctx, deployedTriggers, clients, app)
		if err != nil {
			return err
		}
		triggersList = append(triggersList, trigs...)
		cmd.Printf("%s\n", strings.Join(triggersList, "\n"))
		cmd.Println()
		if nextCursor != "" {
			proceed, err = clients.IO.ConfirmPrompt(ctx, "Show more triggers?", false)
			if err != nil {
				return err
			}
			args.Cursor = nextCursor
		}
	}
	return nil
}
