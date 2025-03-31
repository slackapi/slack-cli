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
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	tests := map[string]struct {
		App           types.App
		Collaborators []types.SlackUser
	}{
		"lists no collaborators if none exist": {
			App: types.App{
				AppID: "A001",
			},
			Collaborators: []types.SlackUser{},
		},
		"lists the collaborator if one exists": {
			App: types.App{
				AppID: "A002",
			},
			Collaborators: []types.SlackUser{
				{
					ID:             "USLACKBOT",
					UserName:       "slackbot",
					Email:          "bots@slack.com",
					PermissionType: types.OWNER,
				},
			},
		},
		"lists all collaborators if many exist": {
			App: types.App{
				AppID: "A002",
			},
			Collaborators: []types.SlackUser{
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
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			appSelectMock := prompts.NewAppSelectMock()
			teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
			appSelectMock.On("TeamAppSelectPrompt").Return(prompts.SelectedApp{App: tt.App, Auth: types.SlackAuth{}}, nil)
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.Collaborators, nil)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
			})

			err := NewListCommand(clients).Execute()
			require.NoError(t, err)
			clientsMock.ApiInterface.AssertCalled(t, "ListCollaborators", mock.Anything, mock.Anything, tt.App.AppID)
			clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorListSuccess, mock.Anything)
			clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorListCount, []string{
				fmt.Sprintf("%d", len(tt.Collaborators)),
			})
			for _, collaborator := range tt.Collaborators {
				clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorListCollaborator, []string{
					collaborator.ID,
					collaborator.UserName,
					collaborator.Email,
					string(collaborator.PermissionType),
				})
			}
		})
	}
}
