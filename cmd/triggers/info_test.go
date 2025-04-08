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
	"testing"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTriggersInfoCommand(t *testing.T) {
	var appSelectTeardown func()

	testutil.TableTestCommand(t, testutil.CommandTests{
		"no params; use prompts to select a trigger": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectTeardown = setupMockInfoAppSelection(installedProdApp)
				mockRequestTrigger := createFakeTrigger(fakeTriggerID, "test trigger", "test app", "shortcut")
				// promptForTriggerID lists available triggers
				cm.ApiInterface.On("WorkflowsTriggersList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.DeployedTrigger{{Name: "test trigger", ID: fakeTriggerID, Type: "Shortcut", Workflow: types.TriggerWorkflow{AppID: fakeAppID}}}, "", nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Choose a trigger:", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 0, Prompt: true}, nil)
				cm.ApiInterface.On("WorkflowsTriggersInfo", mock.Anything, mock.Anything, mock.Anything).Return(mockRequestTrigger, nil)
				cm.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				cm.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).Return(types.EVERYONE, []string{}, nil).Once()
			},
			ExpectedOutputs: []string{
				"Trigger Info",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersInfo", mock.Anything, mock.Anything, fakeTriggerID)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"app not installed": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID},
			ExpectedErrorStrings: []string{cmdutil.DeployedAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockInfoAppSelection(newProdApp)
				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass --trigger-id for a workspace app, success": {
			CmdArgs: []string{"--trigger-id", fakeTriggerID},
			ExpectedOutputs: []string{
				"Trigger Info",
				fakeTriggerID,
				"everyone in the workspace",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockInfoAppSelection(installedProdApp)
				mockRequestTrigger := createFakeTrigger(fakeTriggerID, "test trigger", "test app", "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersInfo", mock.Anything, mock.Anything, mock.Anything).Return(mockRequestTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersInfo", mock.Anything, mock.Anything, fakeTriggerID)
			},
		},
		"pass --trigger-id for an org app, success": {
			CmdArgs: []string{"--trigger-id", fakeTriggerID},
			ExpectedOutputs: []string{
				"Trigger Info",
				fakeTriggerID,
				"everyone in all workspaces in this org granted to this app",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockInfoAppSelection(installedProdOrgApp)
				mockRequestTrigger := createFakeTrigger(fakeTriggerID, "test trigger", "test app", "shortcut")
				clientsMock.ApiInterface.On("WorkflowsTriggersInfo", mock.Anything, mock.Anything, mock.Anything).Return(mockRequestTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersInfo", mock.Anything, mock.Anything, fakeTriggerID)
			},
		},
		"pass --trigger-id, failure": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID},
			ExpectedErrorStrings: []string{"invalid_auth"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockInfoAppSelection(installedProdApp)
				clientsMock.ApiInterface.On("WorkflowsTriggersInfo", mock.Anything, mock.Anything, mock.Anything).Return(types.DeployedTrigger{}, errors.New("invalid_auth"))
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersInfo", mock.Anything, mock.Anything, fakeTriggerID)
			},
		},
		"event trigger displays hints and warnings": {
			CmdArgs: []string{"--trigger-id", fakeTriggerID},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockInfoAppSelection(installedProdApp)
				mockRequestTrigger := createFakeTrigger(fakeTriggerID, "test trigger", "test app", "event")
				clientsMock.ApiInterface.On("WorkflowsTriggersInfo", mock.Anything, mock.Anything, mock.Anything).Return(mockRequestTrigger, nil)
				clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.ApiInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedOutputs: []string{
				"Trigger Info",
				fakeTriggerID,
				"Hint:\n",
				"Warning:\n",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.ApiInterface.AssertCalled(t, "WorkflowsTriggersInfo", mock.Anything, mock.Anything, fakeTriggerID)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewInfoCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func TestTriggersInfoCommand_AppSelection(t *testing.T) {
	var appSelectTeardown func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"selection error": {
			CmdArgs:              []string{"--trigger-id", "Ft01435GGLBD"},
			ExpectedErrorStrings: []string{"Error"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockInfoAppSelection(newDevApp)
				appSelectMock := prompts.NewAppSelectMock()
				var originalPromptFunc = createAppSelectPromptFunc
				createAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(prompts.SelectedApp{}, errors.New("error"))
				appSelectTeardown = func() {
					createAppSelectPromptFunc = originalPromptFunc
				}
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"select an non-installed local app": {
			CmdArgs:              []string{"--trigger-id", "Ft01435GGLBD"},
			ExpectedErrorStrings: []string{cmdutil.LocalAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")

				appSelectTeardown = setupMockInfoAppSelection(newDevApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"select an non-installed prod app": {
			CmdArgs:              []string{"--trigger-id", "Ft01435GGLBD"},
			ExpectedErrorStrings: []string{cmdutil.DeployedAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")

				appSelectTeardown = setupMockInfoAppSelection(newProdApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewInfoCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func setupMockInfoAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = infoAppSelectPromptFunc
	infoAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedApp, nil)
	return func() {
		infoAppSelectPromptFunc = originalPromptFunc
	}
}
