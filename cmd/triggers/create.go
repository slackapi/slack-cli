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

package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/internal/api"
	internalapp "github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type createCmdFlags struct {
	workflow            string
	title               string
	description         string
	triggerDef          string
	interactivity       bool
	interactivityName   string
	orgGrantWorkspaceID string
}

var createFlags createCmdFlags

var createAppSelectPromptFunc = prompts.AppSelectPrompt
var workspaceInstallAppFunc = app.RunAddCommand
var createPromptShouldRetryWithInteractivityFunc = promptShouldRetryCreateWithInteractivity

const dataInteractivityPayload = "{{data.interactivity}}"

// NewCommand creates a new Cobra command instance
func NewCreateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := cobra.Command{
		Use:   "create [flags]",
		Short: "Create a trigger for a workflow",
		Long:  "Create a trigger to start a workflow",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger create", Meaning: "Create a trigger by selecting an app and trigger definition"},
			{Command: "trigger create --trigger-def \"triggers/shortcut_trigger.ts\"", Meaning: "Create a trigger from a definition file"},
			{Command: "trigger create --workflow \"#/workflows/my_workflow\"", Meaning: "Create a trigger for a workflow"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateCommand(clients, cmd)
		},
	}

	cmd.Flags().StringVar(&createFlags.workflow, "workflow", "", "a reference to the workflow to execute\n  formatted as:\n  \"#/workflows/<workflow_callback_id>\"")
	cmd.Flags().StringVar(&createFlags.title, "title", "My Trigger", "the title of this trigger\n ")
	cmd.Flags().StringVar(&createFlags.description, "description", "", "the description of this trigger")
	cmd.Flags().StringVar(&createFlags.triggerDef, "trigger-def", "", "path to a JSON file containing the trigger\n  definition. Overrides other flags setting\n  trigger properties.")
	cmd.Flags().BoolVar(&createFlags.interactivity, "interactivity", false, "when used with --workflow, adds a\n  \"slack#/types/interactivity\" parameter\n  to the trigger with the name specified\n  by --interactivity-name")
	cmd.Flags().StringVar(&createFlags.interactivityName, "interactivity-name", "interactivity", "when used with --interactivity, specifies\n  the name of the interactivity parameter\n  to use")
	cmd.Flags().StringVar(&createFlags.orgGrantWorkspaceID, cmdutil.OrgGrantWorkspaceFlag, "", cmdutil.OrgGrantWorkspaceDescription())
	return &cmd
}

func runCreateCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	var ctx = cmd.Context()
	var span, _ = opentracing.StartSpanFromContext(ctx, "cmd.triggers.create")
	defer span.Finish()

	// Get the app selection and accompanying auth from the flag or prompt
	selection, err := createAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndNewApps)
	if err != nil {
		return err
	}
	token := selection.Auth.Token
	ctx = config.SetContextToken(ctx, token)
	app := selection.App

	clients.Config.ManifestEnv = internalapp.SetManifestEnvTeamVars(clients.Config.ManifestEnv, selection.App.TeamDomain, selection.App.IsDev)

	if selection.App.IsNew() || selection.App.AppID == "" {
		_ctx, installState, _app, err := workspaceInstallAppFunc(ctx, clients, &selection, createFlags.orgGrantWorkspaceID)
		if err != nil {
			return err
		}
		if installState == types.InstallRequestPending || installState == types.InstallRequestCancelled || installState == types.InstallRequestNotSent {
			return nil
		}
		ctx = _ctx
		app = _app
	}

	err = validateCreateCmdFlags(ctx, clients, &createFlags)
	if err != nil {
		return err
	}

	var triggerArg api.TriggerRequest
	if createFlags.triggerDef != "" {
		triggerArg, err = triggerRequestFromDef(ctx, clients, createFlags, app.IsDev)
		if err != nil {
			return err
		}
	} else {
		triggerArg = triggerRequestFromFlags(createFlags, app.IsDev)
	}

	// Fix the app ID selected from the menu. In the --trigger-def case, this lets you use the same
	// def file for dev and prod.
	triggerArg.WorkflowAppID = app.AppID

	createdTrigger, err := clients.API().WorkflowsTriggersCreate(ctx, token, triggerArg)
	if extendedErr, ok := err.(*api.TriggerCreateOrUpdateError); ok {
		// If the user used --workflow and the creation failed because we were missing the interactivity
		// context, lets prompt and optionally add it
		if createFlags.workflow != "" && extendedErr.MissingParameterDetail.Type == "slack#/types/interactivity" {
			triggerArg.Inputs = api.Inputs{
				extendedErr.MissingParameterDetail.Name: &api.Input{Value: dataInteractivityPayload},
			}
			var retryTriggerCreate bool
			retryTriggerCreate, err = createPromptShouldRetryWithInteractivityFunc(cmd, clients.IO, triggerArg)
			if err != nil {
				return err
			}
			if retryTriggerCreate {
				createdTrigger, err = clients.API().WorkflowsTriggersCreate(ctx, token, triggerArg)
			}
		}
	}

	if err != nil {
		// If the error is workflow_not_found, show it to the user and prompt them to re-install
		if strings.Contains(err.Error(), "workflow_not_found") {
			clients.IO.PrintError(ctx, "Error: %s", err.Error())
			err := ListWorkflows(ctx, clients, selection.App, selection.Auth)
			if err != nil {
				return err
			}
			trigger, ok, err := promptShouldInstallAndRetry(ctx, clients, cmd, selection, token, triggerArg)
			if err != nil {
				return err
			}
			if ok {
				createdTrigger = trigger
			}
		} else {
			return err
		}
	}

	if createdTrigger.ID == "" {
		return nil
	}

	cmd.Printf("\n%s", style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  "Trigger successfully created!",
	}))
	trigs, err := sprintTrigger(ctx, createdTrigger, clients, true, app)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", strings.Join(trigs, "\n"))
	cmd.Println()

	clients.IO.PrintTrace(ctx, slacktrace.TriggersCreateSuccess)
	if createdTrigger.Type == "shortcut" {
		clients.IO.PrintTrace(ctx, slacktrace.TriggersCreateURL, createdTrigger.ShortcutURL)
	}
	if createdTrigger.Type == "webhook" {
		clients.IO.PrintTrace(ctx, slacktrace.TriggersCreateURL, createdTrigger.Webhook)
	}
	return nil
}

// promptShouldInstallAndRetry will prompt to re-install the app to apply any local code changes before attempting to create the trigger
func promptShouldInstallAndRetry(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, selectedApp prompts.SelectedApp, token string, triggerArg api.TriggerRequest) (types.DeployedTrigger, bool, error) {
	shouldRetry, err := clients.IO.ConfirmPrompt(ctx, "Re-install app to apply local file changes and try again?", true)
	if err != nil {
		return types.DeployedTrigger{}, false, err
	}

	if shouldRetry {
		_, installState, _, err := workspaceInstallAppFunc(ctx, clients, &selectedApp, "")
		if err != nil {
			return types.DeployedTrigger{}, false, slackerror.Wrap(err, slackerror.ErrInstallationFailed)
		}
		if installState == types.InstallRequestPending || installState == types.InstallRequestCancelled || installState == types.InstallRequestNotSent {
			return types.DeployedTrigger{}, false, nil
		}

		trigger, err := clients.API().WorkflowsTriggersCreate(ctx, token, triggerArg)
		if err != nil {
			return types.DeployedTrigger{}, false, err
		}

		return trigger, true, nil
	}

	return types.DeployedTrigger{}, false, nil
}

// ListWorkflows displays a list of valid workflow identifiers
func ListWorkflows(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) error {
	slackYaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return err
	}

	if len(slackYaml.Workflows) > 0 {
		workflows := ""
		for callbackID := range slackYaml.Workflows {
			workflows = workflows + fmt.Sprintf("- #/workflows/%s\n", callbackID)
		}
		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji: "bulb",
			Text:  `The "workflow" property in the trigger definition file should be one of the following:`,
			Secondary: []string{
				workflows,
				"Is your workflow missing? Make sure it has been added to the manifest in your app's source code",
			},
		}))
	} else {
		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji: "warning",
			Text:  "Your app has no workflows",
			Secondary: []string{
				"Triggers execute workflows\nYou need to have a workflow before you can create a trigger\nGet started by defining a workflow in your app's source code",
			},
		}))
	}
	return nil
}

func promptShouldRetryCreateWithInteractivity(cmd *cobra.Command, IO iostreams.IOStreamer, triggerArg api.TriggerRequest) (bool, error) {
	return promptShouldRetryWithInteractivity("Would you like to create this trigger?", cmd, IO, triggerArg)
}

func promptShouldRetryWithInteractivity(promptMsg string, cmd *cobra.Command, IO iostreams.IOStreamer, triggerArg api.TriggerRequest) (bool, error) {
	ctx := cmd.Context()

	pretty, err := json.MarshalIndent(triggerArg, "", "    ")
	if err != nil {
		return false, err
	}

	IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "memo",
		Text:  "Workflow requires interactivity",
		Secondary: []string{
			"This workflow requires a `slack#/types/interactivity` parameter. Here's a trigger that would work:\n",
			string(pretty),
		},
	}))

	return IO.ConfirmPrompt(ctx, promptMsg, true)
}

func triggerRequestFromFlags(flags createCmdFlags, isDev bool) api.TriggerRequest {
	req := api.TriggerRequest{
		Type:        types.TriggerTypeShortcut,
		Shortcut:    &api.Shortcut{},
		Name:        flags.title,
		Description: flags.description,
		Workflow:    flags.workflow,
	}
	if flags.interactivity {
		req.Inputs = make(api.Inputs)
		req.Inputs[flags.interactivityName] = &api.Input{Value: dataInteractivityPayload}
	}
	if isDev && req.Name != "" {
		req.Name = style.LocalRunDisplayName(req.Name)
	}
	return req
}

func triggerRequestViaHook(ctx context.Context, clients *shared.ClientFactory, path string, isDev bool) (api.TriggerRequest, error) {
	if !clients.SDKConfig.Hooks.GetTrigger.IsAvailable() {
		return api.TriggerRequest{}, slackerror.New(slackerror.ErrSDKHookNotFound).
			WithMessage("The `get-trigger` hook script in `%s` was not found", filepath.Join(".slack", "hooks.json")).
			WithRemediation("Try defining your trigger by specifying a json file instead.")
	}

	hookExecOpts := hooks.HookExecOpts{
		Hook: clients.SDKConfig.Hooks.GetTrigger,
		Args: map[string]string{"source": path},
		Env:  map[string]string{},
	}
	for name, val := range clients.Config.ManifestEnv {
		hookExecOpts.Env[name] = val
	}
	triggerDefAsStr, err := clients.HookExecutor.Execute(
		ctx,
		hookExecOpts,
	)
	if err != nil {
		return api.TriggerRequest{}, err
	}

	triggerDefAsStr = goutils.ExtractFirstJSONFromString(triggerDefAsStr)
	var triggerReq = api.TriggerRequest{}
	err = json.Unmarshal([]byte(triggerDefAsStr), &triggerReq)
	if isDev && triggerReq.Name != "" {
		triggerReq.Name = style.LocalRunDisplayName(triggerReq.Name)
	}
	return triggerReq, err
}

func triggerRequestFromJSONFile(clients *shared.ClientFactory, path string, isDev bool) (api.TriggerRequest, error) {
	defBytes, err := afero.ReadFile(clients.Fs, path)
	req := api.TriggerRequest{}
	if err != nil {
		return req, err
	}
	err = json.Unmarshal(defBytes, &req)
	if err != nil {
		return req, err
	}
	if isDev && req.Name != "" {
		req.Name = style.LocalRunDisplayName(req.Name)
	}
	return req, nil
}

func triggerRequestFromDef(ctx context.Context, clients *shared.ClientFactory, flags createCmdFlags, isDev bool) (api.TriggerRequest, error) {
	if strings.HasSuffix(flags.triggerDef, ".json") {
		return triggerRequestFromJSONFile(clients, flags.triggerDef, isDev)
	}

	return triggerRequestViaHook(ctx, clients, flags.triggerDef, isDev)
}
