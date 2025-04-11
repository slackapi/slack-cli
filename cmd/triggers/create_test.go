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
	"errors"
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	installedProdApp    = prompts.SelectedApp{Auth: types.SlackAuth{}, App: types.App{AppID: fakeAppID}}
	installedProdOrgApp = prompts.SelectedApp{Auth: types.SlackAuth{}, App: types.App{AppID: fakeAppID, TeamID: fakeAppEnterpriseID, EnterpriseID: fakeAppEnterpriseID}}
	newProdApp          = prompts.SelectedApp{Auth: types.SlackAuth{}, App: types.App{}}
	newDevApp           = prompts.SelectedApp{Auth: types.SlackAuth{}, App: types.App{IsDev: true}}
)

type AppCmdMock struct {
	mock.Mock
}

func (m *AppCmdMock) RunAddCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string, selectedApp *prompts.SelectedApp) (context.Context, types.InstallState, types.App, error) {
	m.Called()
	return ctx, "", types.App{}, nil
}

func TestTriggersCreateCommand(t *testing.T) {
	var appSelectTeardown func()

	testutil.TableTestCommand(t, testutil.CommandTests{
		"only pass --workflow": {
			CmdArgs:         []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedOutputs: []string{"Trigger successfully created!", "My Trigger", "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
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
				expectedTriggerRequest := api.TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Shortcut:      &api.Shortcut{},
					Name:          fakeTriggerName,
					Description:   "Runs the '#/workflows/my_workflow' workflow",
					Workflow:      "#/workflows/my_workflow",
					WorkflowAppId: fakeAppID,
				}
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"pass all shortcut parameters": {
			CmdArgs:         []string{"--workflow", "#/workflows/my_workflow", "--title", "unit tests", "--description", "are the best"},
			ExpectedOutputs: []string{"Trigger successfully created!", "unit tests", "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, "unit tests", fakeAppID, "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Name:          "unit tests",
					Description:   "are the best",
					Shortcut:      &api.Shortcut{},
					Workflow:      "#/workflows/my_workflow",
					WorkflowAppId: fakeAppID,
				}
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"pass --interactivity, default name": {
			CmdArgs:         []string{"--workflow", "#/workflows/my_workflow", "--interactivity", "--title", "unit tests", "--description", "are the best"},
			ExpectedOutputs: []string{"Trigger successfully created!", "unit tests", "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, "unit tests", fakeAppID, "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Name:          "unit tests",
					Description:   "are the best",
					Shortcut:      &api.Shortcut{},
					Workflow:      "#/workflows/my_workflow",
					WorkflowAppId: fakeAppID,
					Inputs: api.Inputs{
						"interactivity": &api.Input{
							Value: "{{data.interactivity}}",
						},
					},
				}
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"pass --interactivity, custom name": {
			CmdArgs:         []string{"--workflow", "#/workflows/my_workflow", "--interactivity", "--interactivity-name", "custom-interactivity", "--title", "unit tests", "--description", "are the best"},
			ExpectedOutputs: []string{"Trigger successfully created!", "unit tests", "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, "unit tests", fakeAppID, "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				expectedTriggerRequest := api.TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Name:          "unit tests",
					Description:   "are the best",
					Shortcut:      &api.Shortcut{},
					Workflow:      "#/workflows/my_workflow",
					WorkflowAppId: fakeAppID,
					Inputs: api.Inputs{
						"custom-interactivity": &api.Input{
							Value: "{{data.interactivity}}",
						},
					},
				}
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"api call fails": {
			CmdArgs:              []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"invalid_auth"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, errors.New("invalid_auth"))
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
			CmdArgs:         []string{"--trigger-def", "trigger_def.json"},
			ExpectedOutputs: []string{"Trigger successfully created!", "schedule"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				fakeTrigger := createFakeTrigger(fakeTriggerID, "name", fakeAppID, "scheduled")
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(fakeTrigger, nil)
				// no collaborators on app
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
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
				expectedTriggerRequest := api.TriggerRequest{
					Type:          types.TriggerTypeScheduled,
					Name:          "name",
					Description:   "desc",
					Workflow:      "#/workflows/my_workflow",
					WorkflowAppId: fakeAppID,
					Schedule:      types.ToRawJson(`{"start_time":"2020-03-15","frequency":{"type":"daily"}}`),
				}
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, expectedTriggerRequest)
			},
		},
		"--trigger-def, file missing": {
			CmdArgs:              []string{"--trigger-def", "foo.json"},
			ExpectedErrorStrings: []string{"File not found"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				clientsMock.HookExecutor.On("Execute", mock.Anything).Return("", nil)
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"--trigger-def, not json": {
			CmdArgs:         []string{"--trigger-def", "triggers/shortcut.ts"},
			ExpectedOutputs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{{}}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				clientsMock.HookExecutor.On("Execute", mock.Anything).Return(`{}`, nil)
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
				var content = `export default {}`
				err = afero.WriteFile(clients.Fs, "triggers/shortcut.ts", []byte(content), 0600)
				require.NoError(t, err, "Cant write apps.json")
				clients.SDKConfig.Hooks.GetTrigger.Command = "echo {}"
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.HookExecutor.AssertCalled(t, "Execute", mock.Anything)
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"--trigger-def, not json, `get-trigger` hook missing": {
			CmdArgs:              []string{"--trigger-def", "triggers/shortcut.ts"},
			ExpectedErrorStrings: []string{"sdk_hook_get_trigger_not_found"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
				var content = `export default {}`
				err = afero.WriteFile(clients.Fs, "triggers/shortcut.ts", []byte(content), 0600)
				require.NoError(t, err, "Cant write apps.json")
				clients.SDKConfig.Hooks.GetTrigger.Command = ""
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertNotCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCreateCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		return cmd
	})
}

func TestTriggersCreateCommand_MissingParameters(t *testing.T) {
	var appSelectTeardown func()
	var promptForInteractivityTeardown func()

	triggerRequestMissingInputs := api.TriggerRequest{
		Type:          types.TriggerTypeShortcut,
		Shortcut:      &api.Shortcut{},
		Name:          fakeTriggerName,
		Description:   "Runs the '#/workflows/my_workflow' workflow",
		Workflow:      "#/workflows/my_workflow",
		WorkflowAppId: fakeAppID,
	}

	triggerRequestWithInteractivityInputs := api.TriggerRequest{
		Type:          types.TriggerTypeShortcut,
		Shortcut:      &api.Shortcut{},
		Name:          fakeTriggerName,
		Description:   "Runs the '#/workflows/my_workflow' workflow",
		Workflow:      "#/workflows/my_workflow",
		WorkflowAppId: fakeAppID,
		Inputs: api.Inputs{
			"my-interactivity": &api.Input{
				Value: "{{data.interactivity}}",
			},
		},
	}

	testutil.TableTestCommand(t, testutil.CommandTests{
		"initial api call fails, missing interactivity, succeeds on retry": {
			CmdArgs:         []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedOutputs: []string{"Trigger successfully created!"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				promptForInteractivityTeardown = setupMockCreatePromptForInteractivity()
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				extendedErr := &api.TriggerCreateOrUpdateError{
					Err: errors.New("invalid_trigger_inputs"),
					MissingParameterDetail: api.MissingParameterDetail{
						Name: "my-interactivity",
						Type: "slack#/types/interactivity",
					},
				}
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestMissingInputs).Return(types.DeployedTrigger{}, extendedErr)

				fakeTrigger := createFakeTrigger(fakeTriggerID, "unit tests", fakeAppID, "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs).Return(fakeTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestMissingInputs)
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs)
			},
			Teardown: func() {
				appSelectTeardown()
				promptForInteractivityTeardown()
			},
		},
		"initial api call fails, missing interactivity, fails on retry": {
			CmdArgs:              []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"internal_error"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				promptForInteractivityTeardown = setupMockCreatePromptForInteractivity()
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				extendedErr := &api.TriggerCreateOrUpdateError{
					Err: errors.New("invalid_trigger_inputs"),
					MissingParameterDetail: api.MissingParameterDetail{
						Name: "my-interactivity",
						Type: "slack#/types/interactivity",
					},
				}
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestMissingInputs).Return(types.DeployedTrigger{}, extendedErr).Once()

				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs).Return(types.DeployedTrigger{}, errors.New("internal_error")).Once()
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestMissingInputs)
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, triggerRequestWithInteractivityInputs)
			},
			Teardown: func() {
				appSelectTeardown()
				promptForInteractivityTeardown()
			},
		},
		"initial api call fails, missing a different type": {
			CmdArgs:              []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"invalid_trigger_inputs"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockCreateAppSelection(installedProdApp)
				promptForInteractivityTeardown = setupMockCreatePromptForInteractivity()
				// TODO: always a) mock out calls and b) call AddDefaultMocks before making any clients.* calls
				extendedErr := &api.TriggerCreateOrUpdateError{
					Err: errors.New("invalid_trigger_inputs"),
					MissingParameterDetail: api.MissingParameterDetail{
						Name: "my-num",
						Type: "number",
					},
				}
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, extendedErr)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything)
			},
			Teardown: func() {
				appSelectTeardown()
				promptForInteractivityTeardown()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCreateCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		return cmd
	})
}

func TestTriggersCreateCommand_AppSelection(t *testing.T) {
	var appSelectTeardown func()
	var appCommandMock *app.AppMock
	var workspaceInstallAppTeardown func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"selection error": {
			CmdArgs:              []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{"selection error"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockCreateAppSelection(newDevApp)
				appSelectMock := prompts.NewAppSelectMock()
				var originalPromptFunc = createAppSelectPromptFunc
				createAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(prompts.SelectedApp{}, errors.New("selection error"))
				appSelectTeardown = func() {
					createAppSelectPromptFunc = originalPromptFunc
				}
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"select a non-installed local app": {
			CmdArgs: []string{"--workflow", "#/workflows/my_workflow"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				// Define app selector mock to choose local app
				appSelectTeardown = setupMockCreateAppSelection(newDevApp)
				// Define app install mock
				appCommandMock = app.NewAppCommandMock()
				var originalWorkspaceInstallAppFunc = workspaceInstallAppFunc
				workspaceInstallAppFunc = appCommandMock.RunAddCommand
				workspaceInstallAppTeardown = func() {
					workspaceInstallAppFunc = originalWorkspaceInstallAppFunc
				}
				appCommandMock.On("RunAddCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(ctx, types.SUCCESS, newDevApp.App, nil)
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).
					Return(types.DeployedTrigger{}, nil)
				// Define default mocks
				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
				workspaceInstallAppTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				appCommandMock.AssertCalled(t, "RunAddCommand", mock.Anything, mock.Anything, &newDevApp, mock.Anything)
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"select a non-installed prod app": {
			CmdArgs: []string{"--workflow", "#/workflows/my_workflow"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				// Define app selector mock to choose production app
				appSelectTeardown = setupMockCreateAppSelection(newProdApp)
				// Define workspace install mock
				appCommandMock = app.NewAppCommandMock()
				var originalWorkspaceInstallAppFunc = workspaceInstallAppFunc
				workspaceInstallAppFunc = appCommandMock.RunAddCommand
				workspaceInstallAppTeardown = func() {
					workspaceInstallAppFunc = originalWorkspaceInstallAppFunc
				}
				appCommandMock.On("RunAddCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(ctx, types.SUCCESS, newDevApp.App, nil)
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).
					Return(types.DeployedTrigger{}, nil)
				// Define default mocks
				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
				workspaceInstallAppTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				appCommandMock.AssertCalled(t, "RunAddCommand", mock.Anything, mock.Anything, &newProdApp, mock.Anything)
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCreateCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		return cmd
	})
}

func TestTriggersCreateCommand_promptShouldInstallAndRetry(t *testing.T) {

	testcases := []struct {
		name         string
		prepareMocks func(*testing.T, context.Context, *shared.ClientsMock, *app.AppMock)
		check        func(*testing.T, types.DeployedTrigger, bool, error)
	}{
		{
			name: "Accept prompt to reinstall and create the trigger successfully",
			prepareMocks: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, appCmdMock *app.AppMock) {
				appCmdMock.On("RunAddCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ctx, types.SUCCESS, types.App{}, nil)
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).
					Return(types.DeployedTrigger{Name: "trigger name", ID: "Ft123", Type: "shortcut"}, nil)
				clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Re-install app to apply local file changes and try again?", mock.Anything).Return(true, nil)
			},
			check: func(t *testing.T, trigger types.DeployedTrigger, ok bool, err error) {
				assert.Equal(t, ok, true)
				assert.Equal(t, trigger.ID, "Ft123")
				assert.Nil(t, err)
			},
		},
		{
			name: "Decline prompt to reinstall and do nothing",
			prepareMocks: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, appCmdMock *app.AppMock) {
				clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Re-install app to apply local file changes and try again?", mock.Anything).Return(false, nil)
			},
			check: func(t *testing.T, trigger types.DeployedTrigger, ok bool, err error) {
				assert.Equal(t, ok, false)
				assert.Nil(t, err)
			},
		},
		{
			name: "Accept prompt to reinstall but fail to create the trigger",
			prepareMocks: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, appCmdMock *app.AppMock) {
				appCmdMock.On("RunAddCommand", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ctx, types.SUCCESS, types.App{}, nil)
				clientsMock.ApiInterface.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).
					Return(types.DeployedTrigger{}, errors.New("something_else_went_wrong"))
				clientsMock.Os.On("UserHomeDir").Return("", nil) // Called by clients.IO.PrintError
				clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Re-install app to apply local file changes and try again?", mock.Anything).Return(true, nil)
			},
			check: func(t *testing.T, trigger types.DeployedTrigger, ok bool, err error) {
				assert.Equal(t, ok, false)
				assert.ErrorContains(t, err, "something_else_went_wrong")
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			fmt.Println("Running TriggersCreate test: ", testcase.name)

			// Prepare mocks
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cmd := NewCreateCommand(clients)
			triggerRequest := api.TriggerRequest{}

			appCmdMock := new(app.AppMock)
			workspaceInstallAppFunc = appCmdMock.RunAddCommand

			testcase.prepareMocks(t, ctx, clientsMock, appCmdMock)

			// Execute test

			var trigger types.DeployedTrigger
			var ok bool
			trigger, ok, err := promptShouldInstallAndRetry(ctx, clients, cmd, installedProdApp, "token", triggerRequest)
			testcase.check(t, trigger, ok, err)
		})
	}
}

func TestTriggersCreateCommand_promptShouldDisplayTriggerDefinitionFiles(t *testing.T) {

	testcases := []struct {
		name         string
		prepareMocks func(*shared.ClientsMock)
		check        func(t *testing.T, cf createCmdFlags, err error)
	}{
		{
			name: "Sets the triggerDef flag when user selects a trigger definition file from the prompt",
			prepareMocks: func(clientsMock *shared.ClientsMock) {
				clientsMock.Os.On("Getwd").Return("", nil)
				clientsMock.Os.On("Glob", mock.Anything).Return([]string{"triggers/trigger.ts"}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("trigger-def"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "triggers/trigger.ts",
					Index:  0,
				}, nil)
			},
			check: func(t *testing.T, cf createCmdFlags, err error) {
				assert.Equal(t, cf.triggerDef, "triggers/trigger.ts")
			},
		},
		{
			name: "Handles error",
			prepareMocks: func(clientsMock *shared.ClientsMock) {
				clientsMock.Os.On("Getwd").Return("", fmt.Errorf("something went wrong"))
			},
			check: func(t *testing.T, cf createCmdFlags, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {

			// Prepare mocks
			clientsMock := shared.NewClientsMock()
			testcase.prepareMocks(clientsMock)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Execute test

			var err error
			createFlags := createCmdFlags{}
			err = maybeSetTriggerDefFlag(clients, &createFlags)
			testcase.check(t, createFlags, err)
		})
	}
}

func setupMockCreateAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = createAppSelectPromptFunc
	createAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedApp, nil)
	return func() {
		createAppSelectPromptFunc = originalPromptFunc
	}
}

func setupMockCreatePromptForInteractivity() func() {
	var originalPromptFunc = createPromptShouldRetryWithInteractivityFunc
	createPromptShouldRetryWithInteractivityFunc =
		func(c *cobra.Command, IO iostreams.IOStreamer, t api.TriggerRequest) (bool, error) {
			return true, nil
		}
	return func() {
		createPromptShouldRetryWithInteractivityFunc = originalPromptFunc
	}
}

func Test_triggerRequestFromFlags(t *testing.T) {
	flags := createCmdFlags{
		workflow: "#/workflows/my_workflow",
		title:    "Fake Trigger",
	}

	devReq := triggerRequestFromFlags(flags, true)
	assert.Equal(t, "Fake Trigger (local)", devReq.Name, "should have (local) suffix")

	prodReq := triggerRequestFromFlags(flags, false)
	assert.Equal(t, "Fake Trigger", prodReq.Name, "should NOT have (local) suffix")

	flagsEmptyTitle := createCmdFlags{
		workflow: "#/workflows/my_workflow",
		title:    "",
	}
	devReq2 := triggerRequestFromFlags(flagsEmptyTitle, true)
	assert.Equal(t, "", devReq2.Name, "should NOT have (local) suffix")

	prodReq2 := triggerRequestFromFlags(flagsEmptyTitle, false)
	assert.Equal(t, "", prodReq2.Name, "should NOT have (local) suffix")
}
