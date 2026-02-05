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
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTriggersUpdateCommand(t *testing.T) {
	var appSelectTeardown func()

	testutil.TableTestCommand(t, testutil.CommandTests{
		"with prompts": {
			CmdArgs:         []string{},
			ExpectedOutputs: []string{"Trigger successfully updated!", fakeTriggerName, "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				clientsMock.AddDefaultMocks()
				// Prompt for Trigger ID
				clientsMock.API.On("WorkflowsTriggersList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.DeployedTrigger{{Name: fakeTriggerName, ID: fakeTriggerID, Type: "Shortcut", Workflow: types.TriggerWorkflow{AppID: fakeAppID}}}, "", nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger:", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 0, Prompt: true}, nil)
				// Prompt for trigger definition file
				clientsMock.Os.On("Glob", mock.Anything).Return([]string{"/Users/user.name/app/triggers/trigger.ts"}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 0, Prompt: true}, nil)
				// Execute update
				fakeTrigger := createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).Return(types.PermissionEveryone, []string{}, nil).Once()
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).Return(types.PermissionEveryone, []string{}, nil).Once()
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerUpdateRequest{
					TriggerID: fakeTriggerID,
					TriggerRequest: api.TriggerRequest{
						Type:          types.TriggerTypeShortcut,
						Shortcut:      &api.Shortcut{},
						Name:          fakeTriggerName,
						WorkflowAppID: fakeAppID,
					},
				}
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"hosted app not installed": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{cmdutil.DeployedAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")

				appSelectTeardown = setupMockUpdateAppSelection(newProdApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"local app not installed": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{cmdutil.LocalAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")

				appSelectTeardown = setupMockUpdateAppSelection(newDevApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"only pass --workflow and --trigger-id": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedOutputs: []string{"Trigger successfully updated!", fakeTriggerName, "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerUpdateRequest{
					TriggerID: fakeTriggerID,
					TriggerRequest: api.TriggerRequest{
						Type:          types.TriggerTypeShortcut,
						Shortcut:      &api.Shortcut{},
						Name:          fakeTriggerName,
						Description:   "Runs the '#/workflows/my_workflow' workflow",
						Workflow:      "#/workflows/my_workflow",
						WorkflowAppID: fakeAppID,
					},
				}
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"only pass --workflow and --trigger-id, with interactivity": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow", "--interactivity"},
			ExpectedOutputs: []string{"Trigger successfully updated!", fakeTriggerName, "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerUpdateRequest{
					TriggerID: fakeTriggerID,
					TriggerRequest: api.TriggerRequest{
						Type:          types.TriggerTypeShortcut,
						Shortcut:      &api.Shortcut{},
						Name:          fakeTriggerName,
						Description:   "Runs the '#/workflows/my_workflow' workflow",
						Workflow:      "#/workflows/my_workflow",
						WorkflowAppID: fakeAppID,
						Inputs: api.Inputs{
							"interactivity": &api.Input{
								Value: "{{data.interactivity}}",
							},
						},
					},
				}
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"only pass --workflow and --trigger-id, with interactivity and custom name": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow", "--interactivity", "--interactivity-name", "custom-interactivity"},
			ExpectedOutputs: []string{"Trigger successfully updated!", fakeTriggerName, "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerUpdateRequest{
					TriggerID: fakeTriggerID,
					TriggerRequest: api.TriggerRequest{
						Type:          types.TriggerTypeShortcut,
						Shortcut:      &api.Shortcut{},
						Name:          fakeTriggerName,
						Description:   "Runs the '#/workflows/my_workflow' workflow",
						Workflow:      "#/workflows/my_workflow",
						WorkflowAppID: fakeAppID,
						Inputs: api.Inputs{
							"custom-interactivity": &api.Input{
								Value: "{{data.interactivity}}",
							},
						},
					},
				}
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"pass all shortcut parameters": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow", "--title", "unit tests", "--description", "are the best"},
			ExpectedOutputs: []string{"Trigger successfully updated!", "unit tests", "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, "unit tests", fakeAppID, "shortcut")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerUpdateRequest{
					TriggerID: fakeTriggerID,
					TriggerRequest: api.TriggerRequest{
						Type:          types.TriggerTypeShortcut,
						Name:          "unit tests",
						Description:   "are the best",
						Shortcut:      &api.Shortcut{},
						Workflow:      "#/workflows/my_workflow",
						WorkflowAppID: fakeAppID,
					},
				}
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"api call fails": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"invalid_auth"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, errors.New("invalid_auth"))
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass --trigger-def, scheduled trigger": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--trigger-def", "trigger_def.json"},
			ExpectedOutputs: []string{"Trigger successfully updated!", "schedule"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, "name", fakeAppID, "scheduled")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
				jsonPayload := `{
							"type":"scheduled",
							"name":"name",
							"description":"desc",
							"workflow":"#/workflows/my_workflow",
							"schedule":{"start_time":"2020-03-15","frequency":{"type":"daily"}}
						}
						`
				err = afero.WriteFile(clients.Fs, "trigger_def.json", []byte(jsonPayload), 0600)
				require.NoError(t, err, "Cant write trigger_def.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerUpdateRequest{
					TriggerID: fakeTriggerID,
					TriggerRequest: api.TriggerRequest{
						Type:          types.TriggerTypeScheduled,
						Name:          "name",
						Description:   "desc",
						Workflow:      "#/workflows/my_workflow",
						WorkflowAppID: fakeAppID,
						Schedule:      types.ToRawJSON(`{"start_time":"2020-03-15","frequency":{"type":"daily"}}`),
					},
				}
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"--trigger-def, file missing": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--trigger-def", "foo.json"},
			ExpectedErrorStrings: []string{"File not found"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"--trigger-def, not json": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--trigger-def", "triggers/shortcut.ts"},
			ExpectedErrorStrings: []string{"unexpected end of JSON"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				clientsMock.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return(`{`, nil)
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
				var content = `export default {}`
				err = afero.WriteFile(clients.Fs, "triggers/shortcut.ts", []byte(content), 0600)
				require.NoError(t, err, "Cant write shortcut.ts")
				// TODO: are we shelling out in these tests?
				clients.SDKConfig.Hooks.GetTrigger.Command = "echo {}'"
			},
			Teardown: func() {
				appSelectTeardown()
			},

			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.HookExecutor.AssertCalled(t, "Execute", mock.Anything, mock.Anything)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewUpdateCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		return cmd
	})
}

func TestTriggersUpdateCommand_MissingParameters(t *testing.T) {
	var appSelectTeardown func()
	var promptForInteractivityTeardown func()

	triggerRequestMissingInputs := api.TriggerUpdateRequest{
		TriggerID: fakeTriggerID,
		TriggerRequest: api.TriggerRequest{
			Type:          types.TriggerTypeShortcut,
			Shortcut:      &api.Shortcut{},
			Name:          fakeTriggerName,
			Description:   "Runs the '#/workflows/my_workflow' workflow",
			Workflow:      "#/workflows/my_workflow",
			WorkflowAppID: fakeAppID,
		},
	}

	triggerRequestWithInteractivityInputs := api.TriggerUpdateRequest{
		TriggerID: fakeTriggerID,
		TriggerRequest: api.TriggerRequest{
			Type:          types.TriggerTypeShortcut,
			Shortcut:      &api.Shortcut{},
			Name:          fakeTriggerName,
			Description:   "Runs the '#/workflows/my_workflow' workflow",
			Workflow:      "#/workflows/my_workflow",
			WorkflowAppID: fakeAppID,
			Inputs: api.Inputs{
				"my-interactivity": &api.Input{
					Value: "{{data.interactivity}}",
				},
			},
		},
	}

	testutil.TableTestCommand(t, testutil.CommandTests{
		"initial api call fails, missing interactivity, succeeds on retry": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedOutputs: []string{"Trigger successfully updated!"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				promptForInteractivityTeardown = setupMockUpdatePromptForInteractivity()
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				extendedErr := &api.TriggerCreateOrUpdateError{
					Err: errors.New("invalid_trigger_inputs"),
					MissingParameterDetail: api.MissingParameterDetail{
						Name: "my-interactivity",
						Type: "slack#/types/interactivity",
					},
				}
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestMissingInputs).Return(types.DeployedTrigger{}, extendedErr)

				fakeTrigger := createFakeTrigger(fakeTriggerID, "unit tests", fakeAppID, "shortcut")
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs).Return(fakeTrigger, nil)
				clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.PermissionEveryone, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
				promptForInteractivityTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestMissingInputs)
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs)
			},
		},
		"initial api call fails, missing interactivity, fails on retry": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"internal_error"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				promptForInteractivityTeardown = setupMockUpdatePromptForInteractivity()
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				extendedErr := &api.TriggerCreateOrUpdateError{
					Err: errors.New("invalid_trigger_inputs"),
					MissingParameterDetail: api.MissingParameterDetail{
						Name: "my-interactivity",
						Type: "slack#/types/interactivity",
					},
				}
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestMissingInputs).Return(types.DeployedTrigger{}, extendedErr)

				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs).Return(types.DeployedTrigger{}, errors.New("internal_error"))
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
				promptForInteractivityTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestMissingInputs)
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs)
			},
		},
		"initial api call fails, missing a different type": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID, "--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"invalid_trigger_inputs"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockUpdateAppSelection(installedProdApp)
				promptForInteractivityTeardown = setupMockUpdatePromptForInteractivity()
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				extendedErr := &api.TriggerCreateOrUpdateError{
					Err: errors.New("invalid_trigger_inputs"),
					MissingParameterDetail: api.MissingParameterDetail{
						Name: "my-num",
						Type: "number",
					},
				}
				clientsMock.API.On("WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, extendedErr)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
				promptForInteractivityTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersUpdate", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewUpdateCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		return cmd
	})
}

func setupMockUpdateAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = updateAppSelectPromptFunc
	updateAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(selectedApp, nil)
	return func() {
		updateAppSelectPromptFunc = originalPromptFunc
	}
}

func setupMockUpdatePromptForInteractivity() func() {
	var originalPromptFunc = updatePromptShouldRetryWithInteractivityFunc
	updatePromptShouldRetryWithInteractivityFunc =
		func(c *cobra.Command, IO iostreams.IOStreamer, t api.TriggerRequest) (bool, error) {
			return true, nil
		}
	return func() {
		updatePromptShouldRetryWithInteractivityFunc = originalPromptFunc
	}
}
