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
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	internalapp "github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type updateCmdFlags struct {
	createCmdFlags
	triggerID string
}

var updateFlags updateCmdFlags

// TODO(mcodik) figure out a way to mock this more nicely
var updateAppSelectPromptFunc = prompts.AppSelectPrompt
var updatePromptShouldRetryWithInteractivityFunc = promptShouldRetryUpdateWithInteractivity

// NewUpdateCommand creates a new Cobra command instance
func NewUpdateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := cobra.Command{
		Use:   "update --trigger-id <id> [flags]",
		Short: "Updates an existing trigger",
		Long:  `Updates an existing trigger with the provided definition. Only supports full replacement, no partial update.`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger update --trigger-id Ft01234ABCD", Meaning: "Update a trigger definition with a selected file"},
			{Command: "trigger update --trigger-id Ft01234ABCD \\\n    --workflow \"#/workflows/my_workflow\" --title \"Updated trigger\"", Meaning: "Update a trigger with a workflow id and title"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateCommand(clients, cmd)
		},
	}

	cmd.Flags().StringVar(&updateFlags.triggerID, "trigger-id", "", "the ID of the trigger to update")
	cmd.Flags().StringVar(&updateFlags.workflow, "workflow", "", "a reference to the workflow to execute\n  formatted as:\n  \"#/workflows/<workflow_callback_id>\"")
	cmd.Flags().StringVar(&updateFlags.title, "title", "My Trigger", "the title of this trigger\n  ")
	cmd.Flags().StringVar(&updateFlags.description, "description", "", "the description of this trigger")
	cmd.Flags().StringVar(&updateFlags.triggerDef, "trigger-def", "", "path to a JSON file containing the trigger\n  definition. Overrides other flags setting\n  trigger properties.")
	cmd.Flags().BoolVar(&updateFlags.interactivity, "interactivity", false, "when used with --workflow, adds a\n  \"slack#/types/interactivity\" parameter\n  to the trigger with the name specified\n  by --interactivity-name")
	cmd.Flags().StringVar(&updateFlags.interactivityName, "interactivity-name", "interactivity", "when used with --interactivity, specifies\n  the name of the interactivity parameter\n  to use")

	return &cmd
}

func runUpdateCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(ctx, "cmd.triggers.update")
	defer span.Finish()

	// Get the app selection and accompanying auth from the flag or prompt
	selection, err := updateAppSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	token := selection.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selection.App

	clients.Config.ManifestEnv = internalapp.SetManifestEnvTeamVars(clients.Config.ManifestEnv, selection.App.TeamDomain, selection.App.IsDev)

	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}

	// Get trigger ID from flag or prompt
	if updateFlags.triggerID == "" {
		updateFlags.triggerID, err = promptForTriggerID(ctx, cmd, clients, app, token, defaultLabels)
		if err != nil {
			if slackerror.ToSlackError(err).Code == slackerror.ErrNoTriggers {
				printNoTriggersMessage(ctx, clients.IO)
				return nil
			}
			return err
		}
	}

	err = validateCreateCmdFlags(ctx, clients, &updateFlags.createCmdFlags)
	if err != nil {
		return err
	}

	cmd.Printf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  "Updating trigger definition...",
	}))

	var triggerArg api.TriggerRequest
	if updateFlags.triggerDef != "" {
		triggerArg, err = triggerRequestFromDef(ctx, clients, updateFlags.createCmdFlags, app.IsDev)
		if err != nil {
			return err
		}
	} else {
		triggerArg = triggerRequestFromFlags(updateFlags.createCmdFlags, app.IsDev)
	}

	// Fix the app ID selected from the menu. In the --trigger-def case, this lets you use the same
	// def file for dev and prod.
	triggerArg.WorkflowAppID = app.AppID

	updateRequest := api.TriggerUpdateRequest{
		TriggerID:      updateFlags.triggerID,
		TriggerRequest: triggerArg,
	}

	updatedTrigger, err := clients.APIInterface().WorkflowsTriggersUpdate(ctx, token, updateRequest)
	if extendedErr, ok := err.(*api.TriggerCreateOrUpdateError); ok {
		// If the user used --workflow and the creation failed because we were missing the interactivity
		// context, lets prompt and optionally add it
		if updateFlags.workflow != "" && extendedErr.MissingParameterDetail.Type == "slack#/types/interactivity" {
			updateRequest.TriggerRequest.Inputs = api.Inputs{
				extendedErr.MissingParameterDetail.Name: &api.Input{Value: dataInteractivityPayload},
			}
			shouldUpdate, innerErr := updatePromptShouldRetryWithInteractivityFunc(cmd, clients.IO, triggerArg)
			if innerErr != nil {
				return err
			}
			if shouldUpdate {
				// TODO: based on the unit tests, I _think_ this should be the behaviour.. but needs a review.
				// Assumption is: if trigger update fails due to missing interactivity, we prompt user to tweak their definition to include interactivity, recreate, and if successful, proceed.
				updatedTrigger, innerErr = clients.APIInterface().WorkflowsTriggersUpdate(ctx, token, updateRequest)
				if innerErr != nil {
					return innerErr
				} else {
					// TODO: Previously this logic did not exist, meaning, even if user was prompted and they opted to update trigger with interactivity, the error check ~7 lines below would fail. I think that was incorrect?
					err = nil
				}
			}
		}
	}

	if err != nil {
		return err
	}

	trigs, err := sprintTrigger(ctx, updatedTrigger, clients, true, app)
	if err != nil {
		return err
	}

	cmd.Printf("%s\n", style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  fmt.Sprintf("Trigger successfully updated!\n%s", strings.Join(trigs, "\n")),
	}))
	return nil
}

func promptShouldRetryUpdateWithInteractivity(cmd *cobra.Command, IO iostreams.IOStreamer, triggerArg api.TriggerRequest) (bool, error) {
	return promptShouldRetryWithInteractivity("Would you like to update the trigger with this definition?", cmd, IO, triggerArg)
}
