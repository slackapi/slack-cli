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

package collaborators

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestAddCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"add an experimental reader collaborator from user id": {
			CmdArgs: []string{"U123", "--permission-type", "reader"},
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
				appSelectMock.On("TeamAppSelectPrompt").Return(prompts.SelectedApp{App: types.App{AppID: "A123"}, Auth: types.SlackAuth{}}, nil)
				// Set experiment flag
				cm.Config.ExperimentsFlag = append(cm.Config.ExperimentsFlag, "read-only-collaborators")
				cm.Config.LoadExperiments(context.Background(), cm.IO.PrintDebug)
				// Mock API call
				cm.ApiInterface.On("AddCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{ID: "U123", PermissionType: types.READER}).Return(nil)
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertCalled(t, "AddCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{ID: "U123", PermissionType: types.READER})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorAddSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorAddCollaborator, []string{"U123", "reader"})
			},
		},
		"add an owner collaborator from collaborator email": {
			CmdArgs: []string{"joe.smith@company.com", "--permission-type", "owner"},
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
				appSelectMock.On("TeamAppSelectPrompt").Return(prompts.SelectedApp{App: types.App{AppID: "A123"}, Auth: types.SlackAuth{}}, nil)
				// Set experiment flag
				cm.Config.ExperimentsFlag = append(cm.Config.ExperimentsFlag, "read-only-collaborators")
				cm.Config.LoadExperiments(context.Background(), cm.IO.PrintDebug)
				// Mock API call
				cm.ApiInterface.On("AddCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{Email: "joe.smith@company.com", PermissionType: types.OWNER}).Return(nil)
				addFlags.permissionType = "owner"
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertCalled(t, "AddCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{Email: "joe.smith@company.com", PermissionType: types.OWNER})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorAddSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorAddCollaborator, []string{"joe.smith@company.com", "owner"})
			},
		},
		"default to owner if permission type is not specified": {
			CmdArgs: []string{"joe.smith@company.com"},
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
				appSelectMock.On("TeamAppSelectPrompt").Return(prompts.SelectedApp{App: types.App{AppID: "A123"}, Auth: types.SlackAuth{}}, nil)
				// Mock API call
				cm.ApiInterface.On("AddCollaborator", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertCalled(t, "AddCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{Email: "joe.smith@company.com", PermissionType: types.OWNER})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorAddSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorAddCollaborator, []string{"joe.smith@company.com", "owner"})
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewAddCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		return cmd
	})
}
