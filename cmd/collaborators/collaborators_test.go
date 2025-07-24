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
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCollaboratorsCommand(t *testing.T) {
	tests := map[string]struct {
		app             types.App
		collaborators   []types.SlackUser
		expectedOutputs []string
	}{
		"lists no collaborators if none exist": {
			app: types.App{
				AppID: "A001",
			},
			collaborators: []types.SlackUser{},
			expectedOutputs: []string{
				" 0 collaborators", // Include space to not match "10 collaborators"
			},
		},
		"lists the collaborator if one exists": {
			app: types.App{
				AppID: "A002",
			},
			collaborators: []types.SlackUser{
				{
					ID:             "USLACKBOT",
					UserName:       "slackbot",
					Email:          "bots@slack.com",
					PermissionType: types.OWNER,
				},
			},
			expectedOutputs: []string{
				"1 collaborator",
				// User info: slackbot
				"USLACKBOT",
				"slackbot",
				"bots@slack.com",
				string(types.OWNER),
			},
		},
		"lists all collaborators if many exist": {
			app: types.App{
				AppID: "A002",
			},
			collaborators: []types.SlackUser{
				{
					ID:             "USLACKBOT",
					UserName:       "slackbot",
					Email:          "bots@slack.com",
					PermissionType: types.OWNER,
				},
				{
					ID:             "U00READER",
					UserName:       "bookworm",
					Email:          "reader@slack.com",
					PermissionType: types.READER,
				},
			},
			expectedOutputs: []string{
				"2 collaborators",
				// User info: slackbot
				"USLACKBOT",
				"slackbot",
				"bots@slack.com",
				string(types.OWNER),
				// User info: bookworm
				"U00READER",
				"bookworm",
				"reader@slack.com",
				string(types.READER),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			appSelectMock := prompts.NewAppSelectMock()
			appSelectPromptFunc = appSelectMock.AppSelectPrompt
			appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{App: tt.app, Auth: types.SlackAuth{}}, nil)
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.collaborators, nil)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
			})

			err := NewCommand(clients).ExecuteContext(ctx)
			require.NoError(t, err)
			clientsMock.API.AssertCalled(t, "ListCollaborators", mock.Anything, mock.Anything, tt.app.AppID)
			clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorListSuccess, mock.Anything)
			clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorListCount, []string{
				fmt.Sprintf("%d", len(tt.collaborators)),
			})
			for _, collaborator := range tt.collaborators {
				clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorListCollaborator, []string{
					collaborator.ID,
					collaborator.UserName,
					collaborator.Email,
					string(collaborator.PermissionType),
				})
			}
			output := clientsMock.GetCombinedOutput()
			for _, expectedOutput := range tt.expectedOutputs {
				require.Contains(t, output, expectedOutput)
			}
		})
	}
}
