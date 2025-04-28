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

func TestTriggersAccessCommand(t *testing.T) {
	var appSelectTeardown func()
	var user1Profile = types.UserProfile{DisplayName: "USER1"}
	var user2Profile = types.UserProfile{DisplayName: "USER2"}

	testutil.TableTestCommand(t, testutil.CommandTests{
		"pass flags to set access to app collaborators": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--app-collaborators"},
			ExpectedOutputs: []string{fmt.Sprintf("Trigger '%s'", fakeTriggerID), "app collaborator"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// set access type to collaborators
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, types.APP_COLLABORATORS, "").
					Return([]string{"collaborator_ID"}, nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for updated access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to set access to everyone/public for a workspace app": {
			CmdArgs: []string{"--trigger-id", fakeTriggerID, "--everyone"},
			ExpectedOutputs: []string{fmt.Sprintf(
				"Trigger '%s'", fakeTriggerID),
				"everyone in the workspace",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// set access type to everyone
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, mock.Anything, types.EVERYONE, "").
					Return([]string{}, nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to set access to everyone/public for an org app": {
			CmdArgs: []string{"--trigger-id", fakeTriggerID, "--everyone"},
			ExpectedOutputs: []string{
				fmt.Sprintf("Trigger '%s'", fakeTriggerID),
				"everyone in all workspaces in this org granted to this app",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdOrgApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// set access type to everyone
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, mock.Anything, types.EVERYONE, "").
					Return([]string{}, nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to grant access to specific users (previous access: app collaborators)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--users", "user1, user2", "--grant", "--include-app-collaborators=false"},
			ExpectedOutputs: []string{"Users added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "these users", "USER1", "USER2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// set access type to named_entities
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, "USER1,USER2", types.NAMED_ENTITIES, "users").
					Return([]string{}, nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "USER2"}, nil).Once()
				// display user info for updated access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{RealName: "User One", Name: "USER1", Profile: user1Profile}, nil).Once()
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER2").
					Return(&types.UserInfo{RealName: "User Two", Name: "USER2", Profile: user2Profile}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to grant access to specific channels (previous access: app collaborators)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--channels", "channel1, channel2", "--grant", "--include-app-collaborators=false"},
			ExpectedOutputs: []string{"Channels added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "all members of these channels", "CHANNEL1", "CHANNEL2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// set access type to named_entities
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL1,CHANNEL2", types.NAMED_ENTITIES, "channels").
					Return([]string{}, nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"CHANNEL1", "CHANNEL2"}, nil).Once()
				// display channel info for updated access
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL1").
					Return(&types.ChannelInfo{ID: "CHANNEL1", Name: "Channel One"}, nil).Once()
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL2").
					Return(&types.ChannelInfo{ID: "CHANNEL2", Name: "Channel Two"}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to grant access to specific team (previous access: app collaborators)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workspaces", "team1", "--grant"},
			ExpectedOutputs: []string{"Workspace added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "all members of this workspace", "TEAM1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// confirm to add app collaborators
				clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Include app collaborators?", mock.Anything).Return(true, nil)
				// set access type to named_entities
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, "collaborator_ID", types.NAMED_ENTITIES, "users").
					Return([]string{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "TEAM1", "workspaces").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"collaborator_ID", "TEAM1"}, nil).Once()
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// display workspace info for updated access
				clientsMock.APIInterface.On("TeamsInfo", mock.Anything, mock.Anything, "TEAM1").
					Return(&types.TeamInfo{ID: "TEAM1", Name: "Team One"}, nil).Once()
				// no-op add-collaborators
				clientsMock.APIInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to grant access to specific teams with include-app-collaborators flag provided (previous access: app collaborators)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workspaces", "team1", "--grant", "--include-app-collaborators"},
			ExpectedOutputs: []string{"Workspace added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "all members of this workspace", "TEAM1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// set access type to named_entities
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, "collaborator_ID", types.NAMED_ENTITIES, "users").
					Return([]string{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "TEAM1", "workspaces").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"collaborator_ID", "TEAM1"}, nil).Once()
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// display workspace info for updated access
				clientsMock.APIInterface.On("TeamsInfo", mock.Anything, mock.Anything, "TEAM1").
					Return(&types.TeamInfo{ID: "TEAM1", Name: "Team One"}, nil).Once()
				// no-op add-collaborators
				clientsMock.APIInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to grant access to specific users and channels (previous access: app collaborators)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--users", "user1, user2", "--channels", "channel1, channel2", "--grant"},
			ExpectedOutputs: []string{"Users added", "Channels added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "these users", "all members of these channels", "USER1", "USER2", "CHANNEL1", "CHANNEL2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []string{"collaborator_ID"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "collaborator_ID").
					Return(&types.UserInfo{}, nil).Once()
				// confirm to add app collaborators
				clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Include app collaborators?", mock.Anything).Return(true)
				// set access type to named_entities
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, "collaborator_ID", types.NAMED_ENTITIES, "users").Maybe().
					Return([]string{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "USER1,USER2", types.NAMED_ENTITIES, "users").Maybe().
					Return([]string{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL1,CHANNEL2", types.NAMED_ENTITIES, "channels").Maybe().
					Return([]string{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL1,CHANNEL2", "channels").Maybe().
					Return(nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "USER1,USER2", "users").Maybe().
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "USER2", "CHANNEL1", "CHANNEL2"}, nil).Once()
				// display user info for updated access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{RealName: "User One", Name: "USER1", Profile: user1Profile}, nil).Once()
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER2").
					Return(&types.UserInfo{RealName: "User Two", Name: "USER2", Profile: user2Profile}, nil).Once()

				// display channel info for updated access
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL1").
					Return(&types.ChannelInfo{ID: "CHANNEL1", Name: "Channel One"}, nil).Once()
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL2").
					Return(&types.ChannelInfo{ID: "CHANNEL2", Name: "Channel Two"}, nil).Once()

				// no-op add-collaborators
				clientsMock.APIInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to grant access to specific users (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--users", "USER2", "--grant"},
			ExpectedOutputs: []string{"User added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "these users", "USER1", "USER2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{}, nil).Once()
				// add users to named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "USER2", "users").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "USER2"}, nil).Once()
				// display user info for updated access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{RealName: "User One", Name: "USER1", Profile: user1Profile}, nil).Once()
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER2").
					Return(&types.UserInfo{RealName: "User Two", Name: "USER2", Profile: user2Profile}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to grant access to specific channels (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--channels", "CHANNEL2", "--grant"},
			ExpectedOutputs: []string{"Channel added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "all members of these channels", "CHANNEL1", "CHANNEL2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"CHANNEL1"}, nil).Once()
				// display user info for current access
				// clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "USER1").
				// 	Return(&types.UserInfo{}, nil).Once()
				// add users to named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL2", "channels").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"CHANNEL1", "CHANNEL2"}, nil).Once()
				// display channel info for updated access
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL1").
					Return(&types.ChannelInfo{ID: "CHANNEL1", Name: "Channel One"}, nil).Once()
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL2").
					Return(&types.ChannelInfo{ID: "CHANNEL2", Name: "Channel Two"}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to grant access to specific users, channels and workspaces (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--users", "user1, user2", "--channels", "channel2", "--workspaces", "team1", "--grant"},
			ExpectedOutputs: []string{"Users added", "Channel added", fmt.Sprintf("Trigger '%s'", fakeTriggerID), "these users", "all members of these channels", "all members of this workspace", "USER1", "USER2", "CHANNEL1", "CHANNEL2", "TEAM1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"CHANNEL1"}, nil).Once()
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL2", "channels").
					Return(nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "USER1,USER2", "users").
					Return(nil)
				clientsMock.APIInterface.On("TriggerPermissionsAddEntities", mock.Anything, mock.Anything, fakeTriggerID, "TEAM1", "workspaces").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "USER2", "CHANNEL1", "CHANNEL2", "TEAM1"}, nil).Once()
				// display user info for updated access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{RealName: "User One", Name: "USER1", Profile: user1Profile}, nil).Once()
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER2").
					Return(&types.UserInfo{RealName: "User Two", Name: "USER2", Profile: user2Profile}, nil).Once()

				// display channel info for updated access
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL1").
					Return(&types.ChannelInfo{ID: "CHANNEL1", Name: "Channel One"}, nil).Once()
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL2").
					Return(&types.ChannelInfo{ID: "CHANNEL2", Name: "Channel Two"}, nil).Once()

				// display workspace info for updated access
				clientsMock.APIInterface.On("TeamsInfo", mock.Anything, mock.Anything, "TEAM1").
					Return(&types.TeamInfo{ID: "TEAM1", Name: "Team One"}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to revoke access from specific users (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--users", "USER2", "--revoke"},
			ExpectedOutputs: []string{fmt.Sprintf("Trigger '%s'", fakeTriggerID), "this user", "USER1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "USER2"}, nil).Once()
				// display user info for current access
				clientsMock.APIInterface.On("UserInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{}, nil).Once()
				// remove users from named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsRemoveEntities", mock.Anything, mock.Anything, fakeTriggerID, "USER2", "users").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1"}, nil).Once()
				// display user info for updated access
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{RealName: "User One", Name: "USER1", Profile: user1Profile}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to revoke access from specific channels (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--channels", "CHANNEL2", "--revoke"},
			ExpectedOutputs: []string{fmt.Sprintf("Trigger '%s'", fakeTriggerID), "all members of this channel", "CHANNEL1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"CHANNEL1", "CHANNEL2"}, nil).Once()
				// remove channel from named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsRemoveEntities", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL2", "channels").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"CHANNEL1"}, nil).Once()

				// display channel info for updated access
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL1").
					Return(&types.ChannelInfo{ID: "CHANNEL1", Name: "Channel One"}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to revoke access from specific workspace (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--workspaces", "TEAM2", "--revoke"},
			ExpectedOutputs: []string{fmt.Sprintf("Trigger '%s'", fakeTriggerID), "all members of this workspace", "TEAM1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"TEAM1", "TEAM2"}, nil).Once()
				// remove workspace from named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsRemoveEntities", mock.Anything, mock.Anything, fakeTriggerID, "TEAM2", "workspaces").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"TEAM1"}, nil).Once()

				// display team info for updated access
				clientsMock.APIInterface.On("TeamsInfo", mock.Anything, mock.Anything, "TEAM1").
					Return(&types.TeamInfo{ID: "TEAM1", Name: "Team One"}, nil).Once()

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"pass flags to revoke access from specific users and channels (previous access: named_entities)": {
			CmdArgs:         []string{"--trigger-id", fakeTriggerID, "--users", "USER2", "--channels", "CHANNEL2", "--revoke"},
			ExpectedOutputs: []string{fmt.Sprintf("Trigger '%s'", fakeTriggerID), "this user", "all members of this channel", "USER1", "CHANNEL1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// get current access type
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "USER2", "CHANNEL1", "CHANNEL2"}, nil).Once()
				// remove user from named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsRemoveEntities", mock.Anything, mock.Anything, fakeTriggerID, "USER2", "users").
					Return(nil)
				// remove channel from named_entities ACL
				clientsMock.APIInterface.On("TriggerPermissionsRemoveEntities", mock.Anything, mock.Anything, fakeTriggerID, "CHANNEL2", "channels").
					Return(nil)
				// get access type from backend to display
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []string{"USER1", "CHANNEL1"}, nil).Once()
				clientsMock.APIInterface.On("UsersInfo", mock.Anything, mock.Anything, "USER1").
					Return(&types.UserInfo{RealName: "User One", Name: "USER1", Profile: user1Profile}, nil).Once()
				clientsMock.AddDefaultMocks()
				// display channel info for updated access
				clientsMock.APIInterface.On("ChannelsInfo", mock.Anything, mock.Anything, "CHANNEL1").
					Return(&types.ChannelInfo{ID: "CHANNEL1", Name: "Channel One"}, nil).Once()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},

		"set trigger ID through prompt": {
			CmdArgs:         []string{"--everyone"},
			ExpectedOutputs: []string{fmt.Sprintf("Trigger '%s'", fakeTriggerID), "everyone"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				appSelectTeardown = setupMockAccessAppSelection(installedProdApp)
				// trigger ID prompt lists available triggers, including current access
				clientsMock.APIInterface.On("WorkflowsTriggersList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.DeployedTrigger{{Name: "Trigger 1", ID: fakeTriggerID, Type: "Shortcut", Workflow: types.TriggerWorkflow{AppID: fakeAppID}}}, "", nil)
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				// set and display access
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()
				clientsMock.APIInterface.On("TriggerPermissionsSet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, types.EVERYONE, "").
					Return([]string{}, nil)
				clientsMock.APIInterface.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.EVERYONE, []string{}, nil).Once()

				clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("trigger-id"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "(Ft012345, TRIGGER_INFO)",
					Index:  0,
				}, nil)

				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewAccessCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func TestTriggersAccessCommand_AppSelection(t *testing.T) {
	var appSelectTeardown func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"select an non-installed app": {
			CmdArgs:              []string{"--workflow", "#/workflows/my_workflow"},
			ExpectedErrorStrings: []string{cmdutil.DeployedAppNotInstalledMsg},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
				appSelectTeardown = setupMockAccessAppSelection(
					prompts.SelectedApp{Auth: types.SlackAuth{}, App: types.App{}},
				)
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewAccessCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func setupMockAccessAppSelection(selectedApp prompts.SelectedApp) func() {
	appSelectMock := prompts.NewAppSelectMock()
	var originalPromptFunc = accessAppSelectPromptFunc
	accessAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedApp, nil)
	return func() {
		accessAppSelectPromptFunc = originalPromptFunc
	}
}
