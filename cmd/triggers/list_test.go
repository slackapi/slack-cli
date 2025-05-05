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
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestTriggersListCommand(t *testing.T) {
	var appSelectTeardown func()

	var triggerListRequestArgs api.TriggerListRequest
	testutil.TableTestCommand(t, testutil.CommandTests{
		"list no triggers": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockListAppSelection(installedProdApp)

				// Mock API responses
				triggerListRequestArgs = api.TriggerListRequest{
					AppID:  fakeAppID,
					Limit:  listFlags.triggerLimit,
					Cursor: "",
					Type:   listFlags.triggerType,
				}
				clientsMock.APIInterface.On("WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs).Return([]types.DeployedTrigger{}, "", nil)

				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedOutputs: []string{
				"Listing triggers installed to the app...",
				"There are no triggers installed for the app",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.APIInterface.AssertCalled(t, "WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs)
			},
		},

		"list one trigger for a workspace app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockListAppSelection(installedProdApp)

				// Mock API responses
				triggerListRequestArgs = api.TriggerListRequest{
					AppID:  fakeAppID,
					Limit:  listFlags.triggerLimit,
					Cursor: "",
					Type:   listFlags.triggerType,
				}
				clientsMock.APIInterface.On("WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs).Return(
					[]types.DeployedTrigger{
						createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut"),
					},
					"",
					nil,
				)
				clientsMock.APIInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).Return(types.PermissionEveryone, []string{}, nil)

				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedOutputs: []string{
				"Listing triggers installed to the app...",
				fmt.Sprintf("%s %s (%s)", fakeTriggerName, fakeTriggerID, "shortcut"),
				"everyone in the workspace",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.APIInterface.AssertCalled(t, "WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs)
			},
		},

		"list one trigger for an org app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockListAppSelection(installedProdOrgApp)

				// Mock API responses
				triggerListRequestArgs = api.TriggerListRequest{
					AppID:  fakeAppID,
					Limit:  listFlags.triggerLimit,
					Cursor: "",
					Type:   listFlags.triggerType,
				}
				clientsMock.APIInterface.On("WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs).Return(
					[]types.DeployedTrigger{
						createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut"),
					},
					"",
					nil,
				)
				clientsMock.APIInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).Return(types.PermissionEveryone, []string{}, nil)

				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedOutputs: []string{
				"Listing triggers installed to the app...",
				fmt.Sprintf("%s %s (%s)", fakeTriggerName, fakeTriggerID, "shortcut"),
				"everyone in all workspaces in this org granted to this app",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.APIInterface.AssertCalled(t, "WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs)
			},
		},

		"list multiple triggers": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockListAppSelection(installedProdApp)

				// Mock API responses
				triggerListRequestArgs = api.TriggerListRequest{
					AppID:  fakeAppID,
					Limit:  listFlags.triggerLimit,
					Cursor: "",
					Type:   listFlags.triggerType,
				}
				clientsMock.APIInterface.On("WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs).Return(
					[]types.DeployedTrigger{
						createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "shortcut"),
						createFakeTrigger(fakeTriggerID, fakeTriggerName, fakeAppID, "scheduled"),
					},
					"",
					nil,
				)
				clientsMock.APIInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).Return(types.PermissionEveryone, []string{}, nil)

				clientsMock.AddDefaultMocks()
			},
			Teardown: func() {
				appSelectTeardown()
			},
			ExpectedOutputs: []string{
				"Listing triggers installed to the app...",
				fmt.Sprintf("%s %s (%s)", fakeTriggerName, fakeTriggerID, "shortcut"),
				fmt.Sprintf("Trigger ID: %s (%s)", fakeTriggerID, "scheduled"),
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.APIInterface.AssertCalled(t, "WorkflowsTriggersList", mock.Anything, mock.Anything, triggerListRequestArgs)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewListCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func setupMockListAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = listAppSelectPromptFunc
	listAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedApp, nil)
	return func() {
		listAppSelectPromptFunc = originalPromptFunc
	}
}
