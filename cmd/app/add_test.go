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

package app

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

// Mock teams
const (
	// Team1
	team1TeamDomain = "team1"
	team1TeamID     = "T1"
	team1UserID     = "U1"
	team1Token      = "xoxe.xoxp-1-token"
)

var mockAuthTeam1 = types.SlackAuth{
	Token:      team1Token,
	TeamID:     team1TeamID,
	UserID:     team1UserID,
	TeamDomain: team1TeamDomain,
}

var mockAppTeam1 = types.App{
	AppID:      "A1",
	TeamID:     team1TeamID,
	TeamDomain: team1TeamDomain,
	IsDev:      false,
}

var mockOrgAuth = types.SlackAuth{
	Token:               "token",
	TeamID:              "E123",
	UserID:              "U123",
	TeamDomain:          "org",
	IsEnterpriseInstall: true,
}

var mockOrgApp = types.App{
	AppID:      "A1",
	TeamID:     "E123",
	TeamDomain: "org",
	IsDev:      false,
}

func TestAppAddCommandPreRun(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors if not run in a project directory": {
			ExpectedError: slackerror.New(slackerror.ErrInvalidAppDirectory),
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
			},
		},
		"proceeds if run in a project directory": {
			ExpectedError: nil,
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
			},
		},
		"proceeds if manifest.source is local with the bolt experiment": {
			ExpectedError: nil,
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = append(cm.Config.ExperimentsFlag, string(experiment.BoltFrameworks))
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		"proceeds if manifest.source is remote with the bolt experiment": {
			ExpectedError: nil,
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = append(cm.Config.ExperimentsFlag, string(experiment.BoltFrameworks))
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceRemote, nil)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewAddCommand(cf)
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func TestAppAddCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"adds a new local app": {
			CmdArgs:         []string{},
			ExpectedOutputs: []string{"Creating app manifest", "Installing"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "local")

				// Mock TeamSelector prompt to return "team1"
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(prompts.SelectedApp{Auth: mockAuthTeam1}, nil)

				// Mock valid session for team1
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockAuthTeam1.UserID,
					TeamID:   &mockAuthTeam1.TeamID,
					TeamName: &mockAuthTeam1.TeamDomain,
				}, nil)

				// Mock a clean ValidateAppManifest result
				cm.API.On("ValidateAppManifest", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{
						Warnings: slackerror.Warnings{},
					}, nil,
				)

				// Mock Host
				cm.API.On("Host").Return("")

				// Mock a successful CreateApp call and return our mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockAppTeam1.AppID,
					},
					nil,
				)

				// Mock a successful DeveloperAppInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockAppTeam1.AppID,
						APIAccessTokens: struct {
							Bot      string "json:\"bot,omitempty\""
							AppLevel string "json:\"app_level,omitempty\""
							User     string "json:\"user,omitempty\""
						}{},
					},
					types.InstallSuccess,
					nil,
				)

				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		"adds a new deployed app": {
			CmdArgs:         []string{},
			ExpectedOutputs: []string{"Creating app manifest", "Installing"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "deployed")

				// Mock TeamSelector prompt to return "team1"
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowAllApps).Return(prompts.SelectedApp{Auth: mockAuthTeam1}, nil)

				// Mock valid session for team1
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockAuthTeam1.UserID,
					TeamID:   &mockAuthTeam1.TeamID,
					TeamName: &mockAuthTeam1.TeamDomain,
				}, nil)

				// Mock a clean ValidateAppManifest result
				cm.API.On("ValidateAppManifest", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{
						Warnings: slackerror.Warnings{},
					}, nil,
				)

				// Mock Host
				cm.API.On("Host").Return("")

				// Mock a successful CreateApp call and return our mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockAppTeam1.AppID,
					},
					nil,
				)

				// Mock a successful DeveloperAppInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockAppTeam1.AppID,
						APIAccessTokens: struct {
							Bot      string "json:\"bot,omitempty\""
							AppLevel string "json:\"app_level,omitempty\""
							User     string "json:\"user,omitempty\""
						}{},
					},
					types.InstallSuccess,
					nil,
				)

				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		"updates an existing deployed app": {
			CmdArgs:         []string{},
			ExpectedOutputs: []string{"Updated app manifest"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "deployed")

				// Mock TeamSelector prompt to return "team1"
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowAllApps).Return(prompts.SelectedApp{App: mockAppTeam1, Auth: mockAuthTeam1}, nil)

				// Mock valid session for team1
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockAuthTeam1.UserID,
					TeamID:   &mockAuthTeam1.TeamID,
					TeamName: &mockAuthTeam1.TeamDomain,
				}, nil)

				// Mock a clean ValidateAppManifest result
				cm.API.On("ValidateAppManifest", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{
						Warnings: slackerror.Warnings{},
					}, nil,
				)

				// Mock Host
				cm.API.On("Host").Return("")

				// Mock to ensure that an existing deployed app is found
				appClientMock := &app.AppClientMock{}
				appClientMock.On("GetDeployed", mock.Anything, mock.Anything).Return(mockAppTeam1, nil)
				appClientMock.On("SaveDeployed", mock.Anything, mockAppTeam1).Return(nil)
				appClientMock.On("NewDeployed", mock.Anything, mockAppTeam1.TeamID).Return(types.App{}, slackerror.New(slackerror.ErrAppFound))

				cf.AppClient().AppClientInterface = appClientMock

				// Mock to ensure that the existing deployed app is updated successfully
				cm.API.On("UpdateApp", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.UpdateAppResult{
						AppID:             mockAppTeam1.AppID,
						Credentials:       api.Credentials{},
						OAuthAuthorizeURL: "",
					},
					nil,
				)

				// Mock a successful DeveloperAppInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockAppTeam1.AppID,
						APIAccessTokens: struct {
							Bot      string "json:\"bot,omitempty\""
							AppLevel string "json:\"app_level,omitempty\""
							User     string "json:\"user,omitempty\""
						}{},
					},
					types.InstallSuccess,
					nil,
				)

				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("b4b4"), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		"errors if authentication for the team is missing": {
			CmdArgs:       []string{},
			ExpectedError: slackerror.New(slackerror.ErrCredentialsNotFound),
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "deployed")
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowAllApps).Return(prompts.SelectedApp{App: mockAppTeam1}, nil)
			},
		},
		"adds a new deployed app to an org with a workspace grant": {
			CmdArgs:         []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			ExpectedOutputs: []string{"Creating app manifest", "Installing"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "deployed")
				// Select workspace
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowAllApps).Return(prompts.SelectedApp{App: types.NewApp(), Auth: mockOrgAuth}, nil)
				// Mock calls
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockOrgAuth.UserID,
					TeamID:   &mockOrgAuth.TeamID,
					TeamName: &mockOrgAuth.TeamDomain,
				}, nil)
				cm.API.On("ValidateAppManifest", mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{}, nil,
				)
				cm.API.On("Host").Return("")
				// Return mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockOrgApp.AppID,
					},
					nil,
				)
				// Mock call to apps.developerInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockOrgApp.AppID,
					},
					types.InstallSuccess,
					nil,
				)

				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DeveloperAppInstall", mock.Anything, mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything, mock.Anything, "T123", mock.Anything)
			},
		},
		"adds a new local app when --environment local": {
			CmdArgs:         []string{"--team", "T123", "--environment", "local"},
			ExpectedOutputs: []string{"Creating app manifest", "Installing"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "") // Do not set the environment flag

				// Mock SelectPrompt to receive "--environment local"
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "local",
				}, nil)

				// Mock TeamSelector prompt to return "team1"
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(prompts.SelectedApp{Auth: mockAuthTeam1}, nil)

				// Mock valid session for team1
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockAuthTeam1.UserID,
					TeamID:   &mockAuthTeam1.TeamID,
					TeamName: &mockAuthTeam1.TeamDomain,
				}, nil)

				// Mock a clean ValidateAppManifest result
				cm.API.On("ValidateAppManifest", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{
						Warnings: slackerror.Warnings{},
					}, nil,
				)

				// Mock Host
				cm.API.On("Host").Return("")

				// Mock a successful CreateApp call and return our mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockAppTeam1.AppID,
					},
					nil,
				)

				// Mock a successful DeveloperAppInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockAppTeam1.AppID,
						APIAccessTokens: struct {
							Bot      string "json:\"bot,omitempty\""
							AppLevel string "json:\"app_level,omitempty\""
							User     string "json:\"user,omitempty\""
						}{},
					},
					types.InstallSuccess,
					nil,
				)

				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		// TODO(semver:major): Remove this test when the defaulting to deployed is removed.
		"adds a new deployed app when team flag is provided and environment flag is not set": {
			CmdArgs:         []string{"--team", "T123"},
			ExpectedOutputs: []string{"Creating app manifest", "Installing"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "") // Do not set the environment flag

				// Mock SelectPrompt to receive "--environment deployed"
				// It would be better to remove this mock and rely on the SelectPrompt implementation, but we require other parts of IO to be mocked.
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "deployed",
				}, nil)

				// Mock TeamSelector prompt to return "team1"
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(prompts.SelectedApp{Auth: mockAuthTeam1}, nil)

				// Mock valid session for team1
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockAuthTeam1.UserID,
					TeamID:   &mockAuthTeam1.TeamID,
					TeamName: &mockAuthTeam1.TeamDomain,
				}, nil)

				// Mock a clean ValidateAppManifest result
				cm.API.On("ValidateAppManifest", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{
						Warnings: slackerror.Warnings{},
					}, nil,
				)

				// Mock Host
				cm.API.On("Host").Return("")

				// Mock a successful CreateApp call and return our mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockAppTeam1.AppID,
					},
					nil,
				)

				// Mock a successful DeveloperAppInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockAuthTeam1.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockAppTeam1.AppID,
						APIAccessTokens: struct {
							Bot      string "json:\"bot,omitempty\""
							AppLevel string "json:\"app_level,omitempty\""
							User     string "json:\"user,omitempty\""
						}{},
					},
					types.InstallSuccess,
					nil,
				)

				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		"When admin approval request is pending, outputs instructions": {
			CmdArgs:         []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			ExpectedOutputs: []string{"Creating app manifest", "Installing", "Your request to install the app is pending", "complete installation by re-running"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "deployed")
				// Select workspace
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowAllApps).Return(prompts.SelectedApp{App: types.NewApp(), Auth: mockOrgAuth}, nil)
				// Mock calls
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockOrgAuth.UserID,
					TeamID:   &mockOrgAuth.TeamID,
					TeamName: &mockOrgAuth.TeamDomain,
				}, nil)
				cm.API.On("ValidateAppManifest", mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{}, nil,
				)
				cm.API.On("Host").Return("")
				// Return mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockOrgApp.AppID,
					},
					nil,
				)
				// Mock call to apps.developerInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockOrgApp.AppID,
					},
					types.InstallRequestPending,
					nil,
				)
				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
		"When admin approval request is cancelled, outputs instructions": {
			CmdArgs:         []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			ExpectedOutputs: []string{"Creating app manifest", "Installing", "Your request to install the app has been cancelled"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareAddMocks(t, cf, cm, "deployed")
				// Select workspace
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowAllApps).Return(prompts.SelectedApp{App: types.NewApp(), Auth: mockOrgAuth}, nil)
				// Mock calls
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockOrgAuth.UserID,
					TeamID:   &mockOrgAuth.TeamID,
					TeamName: &mockOrgAuth.TeamDomain,
				}, nil)
				cm.API.On("ValidateAppManifest", mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything).Return(
					api.ValidateAppManifestResult{}, nil,
				)
				cm.API.On("Host").Return("")
				// Return mocked AppID
				cm.API.On("CreateApp", mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything).Return(
					api.CreateAppResult{
						AppID: mockOrgApp.AppID,
					},
					nil,
				)
				// Mock call to apps.developerInstall
				cm.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, mockOrgAuth.Token, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					api.DeveloperAppInstallResult{
						AppID: mockOrgApp.AppID,
					},
					types.InstallRequestCancelled,
					nil,
				)
				// Mock existing and updated cache
				cm.API.On(
					"ExportAppManifest",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExportAppResult{},
					nil,
				)
				mockProjectCache := cache.NewCacheMock()
				mockProjectCache.On("GetManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash(""), nil)
				mockProjectCache.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("xoxo"), nil)
				mockProjectCache.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("Cache").Return(mockProjectCache)
				cm.Config.ProjectConfig = mockProjectConfig
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewAddCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		cf.Config.SetFlags(cmd)
		return cmd
	})
}

func prepareAddMocks(t *testing.T, clients *shared.ClientFactory, clientsMock *shared.ClientsMock, appEnvironment string) {
	clientsMock.AddDefaultMocks()

	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).
		Return("api host")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).
		Return("logstash host")

	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			DisplayInformation: types.DisplayInformation{
				Name: team1TeamDomain,
			},
			Workflows: map[string]types.Workflow{"test_workflow": {Title: "test workflow", InputParameters: types.ToRawJSON(`{}`)}},
		},
	}, nil)
	clients.AppClient().Manifest = manifestMock

	// Mock list command
	listPkgMock := new(ListPkgMock)
	listFunc = listPkgMock.List
	listPkgMock.On("List").Return(nil)

	// Mock the prompt to select the app environment.
	if appEnvironment != "" {
		clientsMock.IO.On("SelectPrompt",
			mock.Anything,
			"Choose the app environment",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(iostreams.SelectPromptResponse{
			Flag:   true,
			Option: appEnvironment,
		}, nil)
	}
}
