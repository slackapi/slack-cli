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

package openformresponse

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type exportCmdFlags struct {
	workflow string
	stepID   string
}

var exportFlags exportCmdFlags

var exportAppSelectPromptFunc = prompts.AppSelectPrompt

func NewExportCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := cobra.Command{
		Use:   "export --workflow <reference> [flags]",
		Short: "Export OpenForm responses to CSV",
		Long:  `Export user responses to an OpenForm Slack function as a CSV file.`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "open-form-response export --workflow \"#/workflows/my_workflow\"", Meaning: "Export form responses from a workflow"},
			{Command: "open-form-response export \\\n    --workflow \"#/workflows/my_workflow\" --step-id 0", Meaning: "Export form responses from a step in a workflow"},
		}),
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExportCommand(clients, cmd)
		},
	}

	cmd.PersistentFlags().StringVar(&exportFlags.workflow, "workflow", "", "a reference to the workflow containing the form\n  formatted as \"#/workflows/<workflow_callback_id>\"")
	cmd.PersistentFlags().StringVar(&exportFlags.stepID, "step-id", "", "the ID of an OpenForm step in this workflow")

	return &cmd
}

func runExportCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	var ctx = cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(ctx, "cmd.open-form-response.export")
	defer span.Finish()

	selection, err := exportAppSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	token := selection.Auth.Token
	appID := selection.App.AppID
	ctx = config.SetContextToken(ctx, token)

	if exportFlags.workflow == "" {
		return slackerror.New(slackerror.ErrMissingFlag).WithMessage("--workflow flag required")
	}

	var stepID = exportFlags.stepID
	if stepID == "" {
		stepID, err = pickStepFromPrompt(ctx, clients, token, exportFlags.workflow, appID)
		if err != nil {
			return err
		}
	}

	err = clients.APIInterface().StepsResponsesExport(ctx, token, exportFlags.workflow, appID, stepID)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "gear",
		Text:  "Export started",
		Secondary: []string{
			"Slackbot will DM you with a CSV file once it's ready",
		},
	}))
	return nil
}

func pickStepFromPrompt(ctx context.Context, clients *shared.ClientFactory, token string, workflow string, appID string) (string, error) {
	stepVersions, err := clients.APIInterface().StepsList(ctx, token, workflow, appID)
	if err != nil {
		return "", err
	}

	if len(stepVersions) == 0 {
		return "", slackerror.New("No OpenForm steps found in this workflow")
	}

	stepLabels := []string{}
	for _, stepVersion := range stepVersions {
		deleted := ""
		if stepVersion.IsDeleted {
			deleted = " (deleted)"
		}
		stepLabels = append(stepLabels, fmt.Sprintf("%s - %s%s", stepVersion.StepID, stepVersion.Title, deleted))
	}

	var selectedStep string
	selection, err := clients.IO.SelectPrompt(ctx, "Choose an OpenForm step from this workflow:", stepLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("step-id"),
		Required: true,
	})
	if err != nil {
		return "", err
	} else if selection.Flag {
		selectedStep = selection.Option
	} else if selection.Prompt {
		selectedStep = stepVersions[selection.Index].StepID
	}
	return selectedStep, nil
}
