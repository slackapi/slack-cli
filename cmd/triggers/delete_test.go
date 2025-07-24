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

func TestTriggersDeleteCommand(t *testing.T) {
	var appSelectTeardown func()

	testutil.TableTestCommand(t, testutil.CommandTests{
		"no params; use prompts to successfully delete": {
			CmdArgs:         []string{},
			ExpectedOutputs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockDeleteAppSelection(installedProdApp)
				clientsMock.AddDefaultMocks()
				// promptForTriggerID lists available triggers
				clientsMock.API.On("WorkflowsTriggersList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.DeployedTrigger{{Name: "Trigger 1", ID: fakeTriggerID, Type: "Shortcut", Workflow: types.TriggerWorkflow{AppID: fakeAppID}}}, "", nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger:", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{}, nil)
				clientsMock.API.On("WorkflowsTriggersDelete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"app not installed": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID},
			ExpectedErrorStrings: []string{cmdutil.DeployedAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockDeleteAppSelection(newProdApp)
				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass --trigger-id, success": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID},
			ExpectedOutputs: []string{"Trigger '" + fakeTriggerID + "' deleted"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockDeleteAppSelection(installedProdApp)
				clientsMock.API.On("WorkflowsTriggersDelete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersDelete", mock.Anything, mock.Anything, fakeTriggerID)
			},
		},
		"pass --trigger-id, failure": {
			CmdArgs:              []string{"--trigger-id", fakeTriggerID},
			ExpectedErrorStrings: []string{"invalid_auth"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockDeleteAppSelection(installedProdApp)
				clientsMock.API.On("WorkflowsTriggersDelete", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid_auth"))
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "WorkflowsTriggersDelete", mock.Anything, mock.Anything, fakeTriggerID)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewDeleteCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func TestTriggersDeleteCommand_AppSelection(t *testing.T) {
	var appSelectTeardown func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"selection error": {
			CmdArgs:              []string{"--trigger-id", "Ft01435GGLBD"},
			ExpectedErrorStrings: []string{"selection error"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockDeleteAppSelection(installedProdApp)
				appSelectMock := prompts.NewAppSelectMock()
				deleteAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, errors.New("selection error"))
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"select a non-installed local app": {
			CmdArgs:              []string{"--trigger-id", "Ft01435GGLBD"},
			ExpectedErrorStrings: []string{cmdutil.LocalAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")

				appSelectTeardown = setupMockDeleteAppSelection(newDevApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"select a non-installed prod app": {
			CmdArgs:              []string{"--trigger-id", "Ft01435GGLBD"},
			ExpectedErrorStrings: []string{cmdutil.DeployedAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")

				appSelectTeardown = setupMockDeleteAppSelection(newProdApp)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewDeleteCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func setupMockDeleteAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = deleteAppSelectPromptFunc
	deleteAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(selectedApp, nil)
	return func() {
		deleteAppSelectPromptFunc = originalPromptFunc
	}
}
