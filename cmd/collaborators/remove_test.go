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
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestRemoveCommand(t *testing.T) {
	mockSelection := prompts.SelectedApp{
		App: types.App{
			AppID: "A001",
		},
		Auth: types.SlackAuth{
			UserID: "USLACKBOT",
		},
	}
	mockCollaborators := []types.SlackUser{
		{
			ID:             "USLACKBOT",
			PermissionType: types.OWNER,
		},
		{
			Email:          "reader@slack.com",
			PermissionType: types.READER,
		},
	}
	testutil.TableTestCommand(t, testutil.CommandTests{
		"always attempts to remove the collaborator provided via argument": {
			CmdArgs: []string{"USLACKBOT"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectMock := prompts.NewAppSelectMock()
				teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
				appSelectMock.On("TeamAppSelectPrompt").
					Return(mockSelection, nil)
				cm.API.On("RemoveCollaborator", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				collaborator := types.SlackUser{
					ID: "USLACKBOT",
				}
				cm.API.AssertCalled(t, "RemoveCollaborator", mock.Anything, mock.Anything, "A001", collaborator)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorRemoveSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorRemoveCollaborator, []string{"USLACKBOT"})
			},
		},
		"still attempts to remove the collaborator provided via prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectMock := prompts.NewAppSelectMock()
				teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
				appSelectMock.On("TeamAppSelectPrompt").
					Return(mockSelection, nil)
				cm.API.On("RemoveCollaborator", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				cm.IO.On("IsTTY").Return(true)
				cm.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).
					Return(mockCollaborators, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Remove a collaborator", mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 1}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "RemoveCollaborator", mock.Anything, mock.Anything, "A001", mockCollaborators[1])
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorRemoveSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.CollaboratorRemoveCollaborator, []string{"reader@slack.com"})
			},
		},
		"avoids removing the user performing the command without confirmation": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectMock := prompts.NewAppSelectMock()
				teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt
				appSelectMock.On("TeamAppSelectPrompt").
					Return(mockSelection, nil)
				cm.API.On("RemoveCollaborator", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				cm.IO.On("IsTTY").Return(true)
				cm.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).
					Return(mockCollaborators, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Remove a collaborator", mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 0}, nil)
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to remove yourself?", mock.Anything, mock.Anything).
					Return(false, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "RemoveCollaborator", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			ExpectedError: slackerror.New(slackerror.ErrProcessInterrupted),
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewRemoveCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
		return cmd
	})
}
