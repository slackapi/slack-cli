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

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"given user ID, update a collaborator to be a reader": {
			CmdArgs:         []string{"U123", "--permission-type", "reader"},
			ExpectedOutputs: []string{"U123 successfully updated as a reader collaborator on this app"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{App: types.App{AppID: "A123"}, Auth: types.SlackAuth{}}, nil)
				// Mock APi call
				clientsMock.API.On("UpdateCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{ID: "U123", PermissionType: types.READER}).Return(nil)
			},
		},
		"given email, update a collaborator to be an owner": {
			CmdArgs:         []string{"joe.smith@company.com", "--permission-type", "owner"},
			ExpectedOutputs: []string{"joe.smith@company.com successfully updated as an owner collaborator on this app"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{App: types.App{AppID: "A123"}, Auth: types.SlackAuth{}}, nil)
				// Mock API call
				clientsMock.API.On("UpdateCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{Email: "joe.smith@company.com", PermissionType: types.OWNER}).Return(nil)
			},
		},
		"prompts when permission type not specified": {
			CmdArgs:         []string{"joe.smith@company.com"},
			ExpectedOutputs: []string{"joe.smith@company.com successfully updated as a reader collaborator on this app"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				// Mock app selection
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{App: types.App{AppID: "A123"}, Auth: types.SlackAuth{}}, nil)
				// Mock permission selection prompt
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a permission type", mock.Anything, mock.Anything).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "reader",
						Index:  1,
					}, nil)
				// Mock API call
				clientsMock.API.On("UpdateCollaborator", mock.Anything, mock.Anything,
					"A123",
					types.SlackUser{Email: "joe.smith@company.com", PermissionType: types.READER}).Return(nil)
			},
		},
		"user ID must be provided": {
			CmdArgs:              []string{},
			ExpectedErrorStrings: []string{"accepts 1 arg(s), received 0"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewUpdateCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
		return cmd
	})
}
