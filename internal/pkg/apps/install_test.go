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

package apps

import (
	"bytes"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInstall(t *testing.T) {
	mockEnterpriseID := "E001"
	mockTeamID := "T001"
	mockTeamDomain := "sandbox"
	mockToken := "xoxe.xoxp-example"
	mockTrue := true
	mockUserID := "U001"

	tests := map[string]struct {
		mockApp                 types.App
		mockAPICreate           api.CreateAppResult
		mockAPICreateError      error
		mockAPIInstall          api.DeveloperAppInstallResult
		mockAPIInstallState     types.InstallState
		mockAPIInstallError     error
		mockAPIUpdate           api.UpdateAppResult
		mockAPIUpdateError      error
		mockAuth                types.SlackAuth
		mockAuthSession         api.AuthSession
		mockBoltExperiment      bool
		mockConfirmPrompt       bool
		mockIsTTY               bool
		mockManifest            types.SlackYaml
		mockManifestHashInitial cache.Hash
		mockManifestHashUpdated cache.Hash
		mockManifestSource      config.ManifestSource
		mockOrgGrantWorkspaceID string
		expectedApp             types.App
		expectedCreate          bool
		expectedError           error
		expectedInstallState    types.InstallState
		expectedManifest        types.AppManifest
		expectedUpdate          bool
	}{
		"create a hosted app manifest with expected rosi values": {
			mockApp: types.App{},
			mockAPICreate: api.CreateAppResult{
				AppID: "A001",
			},
			mockAPIUpdateError: slackerror.New(slackerror.ErrAppAdd),
			mockAuth: types.SlackAuth{
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				Token:        mockToken,
				UserID:       mockUserID,
			},
			mockAuthSession: api.AuthSession{
				EnterpriseID: &mockEnterpriseID,
				TeamID:       &mockTeamID,
				TeamName:     &mockTeamDomain,
				UserID:       &mockUserID,
			},
			mockBoltExperiment: true,
			mockManifestSource: config.ManifestSourceLocal,
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 2,
					},
					Settings: &types.AppSettings{
						FunctionRuntime: types.SlackHosted,
					},
				},
			},
			expectedApp: types.App{
				AppID:        "A001",
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
			},
			expectedManifest: types.AppManifest{
				Metadata: &types.ManifestMetadata{
					MajorVersion: 2,
				},
				Settings: &types.AppSettings{
					FunctionRuntime: types.SlackHosted,
					EventSubscriptions: &types.ManifestEventSubscriptions{
						RequestURL: "https://slack.com",
					},
					Interactivity: &types.ManifestInteractivity{
						IsEnabled:             true,
						RequestURL:            "https://slack.com",
						MessageMenuOptionsURL: "https://slack.com",
					},
				},
			},
			expectedCreate: true,
		},
		"updates a hosted app manifest with expected rosi values": {
			mockApp: types.App{
				AppID:      "A001",
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
			},
			mockAPICreateError: slackerror.New(slackerror.ErrAppCreate),
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A001",
			},
			mockAuth: types.SlackAuth{
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				Token:        mockToken,
				UserID:       mockUserID,
			},
			mockAuthSession: api.AuthSession{
				EnterpriseID: &mockEnterpriseID,
				TeamID:       &mockTeamID,
				TeamName:     &mockTeamDomain,
				UserID:       &mockUserID,
			},
			mockBoltExperiment: true,
			mockManifestSource: config.ManifestSourceLocal,
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 2,
					},
					Settings: &types.AppSettings{
						FunctionRuntime: types.SlackHosted,
					},
				},
			},
			expectedApp: types.App{
				AppID:        "A001",
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
			},
			expectedManifest: types.AppManifest{
				Metadata: &types.ManifestMetadata{
					MajorVersion: 2,
				},
				Settings: &types.AppSettings{
					FunctionRuntime: types.SlackHosted,
					EventSubscriptions: &types.ManifestEventSubscriptions{
						RequestURL: "https://slack.com",
					},
					Interactivity: &types.ManifestInteractivity{
						IsEnabled:             true,
						RequestURL:            "https://slack.com",
						MessageMenuOptionsURL: "https://slack.com",
					},
				},
			},
			expectedUpdate: true,
		},
		"avoid changing the manifest if a remote function runtime is specified": {
			mockApp: types.App{
				AppID:  "A002",
				TeamID: mockTeamID,
			},
			mockAPICreateError: slackerror.New(slackerror.ErrAppCreate),
			mockAPIInstall: api.DeveloperAppInstallResult{
				AppID: "A002",
			},
			mockAPIInstallState: types.InstallSuccess,
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A002",
			},
			mockAuth: types.SlackAuth{
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
				Token:      mockToken,
				UserID:     mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockBoltExperiment: true,
			mockConfirmPrompt:  true,
			mockIsTTY:          true,
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 1,
					},
					DisplayInformation: types.DisplayInformation{
						Name: "example-2",
					},
					Settings: &types.AppSettings{
						FunctionRuntime: types.Remote,
						EventSubscriptions: &types.ManifestEventSubscriptions{
							RequestURL: "https://example.com",
						},
					},
				},
			},
			mockManifestHashInitial: cache.Hash("123"),
			mockManifestHashUpdated: cache.Hash("789"),
			expectedApp: types.App{
				AppID:  "A002",
				TeamID: mockTeamID,
			},
			expectedInstallState: types.InstallSuccess,
			expectedManifest: types.AppManifest{
				Metadata: &types.ManifestMetadata{
					MajorVersion: 1,
				},
				DisplayInformation: types.DisplayInformation{
					Name: "example-2",
				},
				Settings: &types.AppSettings{
					FunctionRuntime: types.Remote,
					EventSubscriptions: &types.ManifestEventSubscriptions{
						RequestURL: "https://example.com",
					},
				},
			},
			expectedUpdate: true,
		},
		"avoid changing the manifest if no function runtime is specified": {
			mockApp: types.App{
				AppID:  "A003",
				TeamID: mockTeamID,
			},
			mockAPIInstall: api.DeveloperAppInstallResult{
				AppID: "A003",
			},
			mockAPIInstallState: types.InstallSuccess,
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A003",
			},
			mockAuth: types.SlackAuth{
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
				Token:      mockToken,
				UserID:     mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "example-3",
					},
					Settings: &types.AppSettings{
						SocketModeEnabled: &mockTrue,
					},
				},
			},
			mockManifestHashInitial: cache.Hash("123"),
			mockManifestHashUpdated: cache.Hash("789"),
			expectedApp: types.App{
				AppID:  "A003",
				TeamID: mockTeamID,
			},
			expectedInstallState: types.InstallSuccess,
			expectedManifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "example-3",
				},
				Settings: &types.AppSettings{
					SocketModeEnabled: &mockTrue,
				},
			},
			expectedUpdate: true,
		},
		"avoids updating or installing an app with a remote manifest": {
			mockApp: types.App{
				AppID:  "A004",
				TeamID: mockTeamID,
			},
			mockAuth: types.SlackAuth{
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
				Token:      mockToken,
				UserID:     mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockAPICreateError:  slackerror.New(slackerror.ErrAppCreate),
			mockAPIUpdateError:  slackerror.New(slackerror.ErrAppAdd),
			mockAPIInstallError: slackerror.New(slackerror.ErrAppInstall),
			mockBoltExperiment:  true,
			mockManifestSource:  config.ManifestSourceRemote,
			expectedApp: types.App{
				AppID:  "A004",
				TeamID: mockTeamID,
			},
			expectedCreate:       false,
			expectedInstallState: "",
			expectedUpdate:       false,
		},
		"errors if the remote manifest has an unexpected cache": {
			mockApp: types.App{
				AppID:  "A005",
				TeamID: mockTeamID,
			},
			mockAuth: types.SlackAuth{
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
				Token:      mockToken,
				UserID:     mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockAPICreateError:      slackerror.New(slackerror.ErrAppCreate),
			mockAPIUpdateError:      slackerror.New(slackerror.ErrAppAdd),
			mockAPIInstallError:     slackerror.New(slackerror.ErrAppInstall),
			mockBoltExperiment:      true,
			mockManifestHashInitial: "pt1",
			mockManifestHashUpdated: "pt2",
			mockManifestSource:      config.ManifestSourceLocal,
			expectedCreate:          false,
			expectedError:           slackerror.New(slackerror.ErrAppManifestUpdate),
			expectedInstallState:    "",
			expectedUpdate:          false,
		},
		"errors if the manifest cache is unset without confirmation": {
			mockApp: types.App{
				AppID:        "A005",
				TeamID:       mockTeamID,
				EnterpriseID: mockEnterpriseID,
			},
			mockAPICreateError: slackerror.New(slackerror.ErrAppCreate),
			mockAPIInstall: api.DeveloperAppInstallResult{
				AppID: "A005",
			},
			mockAPIInstallState: types.InstallSuccess,
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A005",
			},
			mockAuth: types.SlackAuth{
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				Token:        mockToken,
				UserID:       mockUserID,
			},
			mockAuthSession: api.AuthSession{
				EnterpriseID: &mockEnterpriseID,
				TeamID:       &mockTeamID,
				TeamName:     &mockTeamDomain,
				UserID:       &mockUserID,
			},
			mockBoltExperiment: true,
			mockConfirmPrompt:  false,
			mockIsTTY:          true,
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 1,
					},
					DisplayInformation: types.DisplayInformation{
						Name: "example-5",
					},
				},
			},
			mockManifestHashInitial: cache.Hash(""),
			mockManifestHashUpdated: cache.Hash("abc"),
			expectedError:           slackerror.New(slackerror.ErrAppManifestUpdate),
			expectedUpdate:          false,
		},
		"continues if the remote manifest cache matches the saved": {
			mockApp: types.App{
				AppID:  "A006",
				TeamID: mockTeamID,
			},
			mockAPICreateError: slackerror.New(slackerror.ErrAppCreate),
			mockAPIInstall: api.DeveloperAppInstallResult{
				AppID: "A006",
			},
			mockAPIInstallState: types.InstallSuccess,
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A006",
			},
			mockAuth: types.SlackAuth{
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
				Token:      mockToken,
				UserID:     mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockBoltExperiment: true,
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 1,
					},
					DisplayInformation: types.DisplayInformation{
						Name: "example-6",
					},
				},
			},
			mockManifestHashInitial: cache.Hash("abc"),
			mockManifestHashUpdated: cache.Hash("abc"),
			expectedApp: types.App{
				AppID:  "A006",
				TeamID: mockTeamID,
			},
			expectedInstallState: types.InstallSuccess,
			expectedManifest: types.AppManifest{
				Metadata: &types.ManifestMetadata{
					MajorVersion: 1,
				},
				DisplayInformation: types.DisplayInformation{
					Name: "example-6",
				},
			},
			expectedUpdate: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("IsTTY").Return(tt.mockIsTTY)
			clientsMock.AddDefaultMocks()
			clientsMock.API.On(
				"CreateApp",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAPICreate,
				tt.mockAPICreateError,
			)
			clientsMock.API.On(
				"DeveloperAppInstall",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAPIInstall,
				tt.mockAPIInstallState,
				tt.mockAPIInstallError,
			)
			clientsMock.API.On(
				"ExportAppManifest",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				api.ExportAppResult{},
				nil,
			)
			clientsMock.API.On(
				"ValidateAppManifest",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				tt.mockApp.AppID,
			).Return(
				api.ValidateAppManifestResult{},
				nil,
			)
			clientsMock.API.On(
				"UpdateApp",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAPIUpdate,
				tt.mockAPIUpdateError,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAuthSession,
				nil,
			)
			if tt.mockIsTTY {
				clientsMock.IO.On(
					"ConfirmPrompt",
					mock.Anything,
					"Update app settings with changes to the local manifest?",
					false,
				).Return(
					tt.mockConfirmPrompt,
					nil,
				)
			}
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockManifest, nil)
			clientsMock.AppClient.Manifest = manifestMock
			mockProjectConfig := config.NewProjectConfigMock()
			if tt.mockBoltExperiment {
				clientsMock.Config.ExperimentsFlag = append(clientsMock.Config.ExperimentsFlag, string(experiment.BoltFrameworks))
				clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)
				mockProjectConfig.On("GetManifestSource", mock.Anything).Return(tt.mockManifestSource, nil)
			}
			mockProjectCache := cache.NewCacheMock()
			mockProjectCache.On(
				"GetManifestHash",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestHashInitial,
				nil,
			)
			mockProjectCache.On(
				"NewManifestHash",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestHashUpdated,
				nil,
			)
			mockProjectCache.On(
				"SetManifestHash",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil)
			mockProjectConfig.On("Cache").Return(mockProjectCache)
			clientsMock.Config.ProjectConfig = mockProjectConfig

			log := logger.New(func(event *logger.LogEvent) {})
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			app, state, err := Install(
				ctx,
				clients,
				log,
				tt.mockAuth,
				false,
				tt.mockApp,
				tt.mockOrgGrantWorkspaceID,
			)

			if tt.expectedError != nil {
				assert.Equal(
					t,
					slackerror.ToSlackError(tt.expectedError).Code,
					slackerror.ToSlackError(err).Code,
				)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedInstallState, state)
			assert.Equal(t, tt.expectedApp, app)
			if tt.expectedUpdate {
				clientsMock.API.AssertCalled(
					t,
					"UpdateApp",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				clientsMock.API.AssertNotCalled(t, "CreateApp")
			} else if tt.expectedCreate {
				clientsMock.API.AssertCalled(
					t,
					"CreateApp",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				clientsMock.API.AssertNotCalled(t, "UpdateApp")
			}
			for _, call := range clientsMock.API.Calls {
				args := call.Arguments
				switch call.Method {
				case "CreateApp":
					assert.Equal(t, tt.mockAuth.Token, args.Get(1))
					assert.Equal(t, tt.expectedManifest, args.Get(2))
				case "UpdateApp":
					assert.Equal(t, tt.mockAuth.Token, args.Get(1))
					assert.Equal(t, tt.mockApp.AppID, args.Get(2))
					assert.Equal(t, tt.expectedManifest, args.Get(3))
				}
			}
		})
	}
}

func TestInstallLocalApp(t *testing.T) {
	mockEnterpriseID := "E001"
	mockTeamID := "T001"
	mockTeamDomain := "sandbox"
	mockToken := "xoxe.xoxp-example"
	mockTrue := true
	mockUserID := "U001"

	tests := map[string]struct {
		isExperimental          bool
		mockApp                 types.App
		mockAPICreate           api.CreateAppResult
		mockAPICreateError      error
		mockAPIInstall          api.DeveloperAppInstallResult
		mockAPIInstallState     types.InstallState
		mockAPIInstallError     error
		mockAPIUpdate           api.UpdateAppResult
		mockAPIUpdateError      error
		mockAuth                types.SlackAuth
		mockAuthSession         api.AuthSession
		mockBoltExperiment      bool
		mockConfirmPrompt       bool
		mockIsTTY               bool
		mockManifest            types.SlackYaml
		mockManifestHashInitial cache.Hash
		mockManifestHashUpdated cache.Hash
		mockManifestSource      config.ManifestSource
		mockOrgGrantWorkspaceID string
		expectedApp             types.App
		expectedCreate          bool
		expectedInstallState    types.InstallState
		expectedManifest        types.AppManifest
		expectedUpdate          bool
	}{
		"create a new run on slack app with a local function runtime using expected rosi defaults": {
			isExperimental: false,
			mockApp:        types.App{},
			mockAPICreate: api.CreateAppResult{
				AppID: "A001",
			},
			mockAPIUpdateError: slackerror.New(slackerror.ErrAppAdd),
			mockAuth: types.SlackAuth{
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				Token:        mockToken,
				UserID:       mockUserID,
			},
			mockAuthSession: api.AuthSession{
				EnterpriseID: &mockEnterpriseID,
				TeamID:       &mockTeamID,
				TeamName:     &mockTeamDomain,
				UserID:       &mockUserID,
			},
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 2,
					},
					DisplayInformation: types.DisplayInformation{
						Name: "example-1",
					},
					Settings: &types.AppSettings{
						FunctionRuntime: types.SlackHosted,
					},
				},
			},
			expectedApp: types.App{
				AppID:        "A001",
				EnterpriseID: mockEnterpriseID,
				IsDev:        true,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				UserID:       mockUserID,
			},
			expectedManifest: types.AppManifest{
				Metadata: &types.ManifestMetadata{
					MajorVersion: 2,
				},
				DisplayInformation: types.DisplayInformation{
					Name: "example-1 (local)",
				},
				Settings: &types.AppSettings{
					FunctionRuntime:   types.LocallyRun,
					SocketModeEnabled: &mockTrue,
					Interactivity: &types.ManifestInteractivity{
						IsEnabled: true,
					},
					EventSubscriptions: &types.ManifestEventSubscriptions{},
				},
			},
			expectedCreate: true,
		},
		"update an existing local bolt app with a remote function runtime without manifest changes": {
			isExperimental: true,
			mockApp: types.App{
				AppID:  "A002",
				TeamID: mockTeamID,
				UserID: mockUserID,
			},
			mockAPICreateError: slackerror.New(slackerror.ErrAppCreate),
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A002",
			},
			mockAuth: types.SlackAuth{
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				Token:        mockToken,
				UserID:       mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Metadata: &types.ManifestMetadata{
						MajorVersion: 1,
					},
					DisplayInformation: types.DisplayInformation{
						Name: "example-2",
					},
					Features: &types.AppFeatures{
						BotUser: types.BotUser{
							DisplayName: "example-2",
						},
					},
					Settings: &types.AppSettings{
						FunctionRuntime: types.Remote,
						EventSubscriptions: &types.ManifestEventSubscriptions{
							RequestURL: "https://example.com",
						},
					},
				},
			},
			mockManifestHashInitial: cache.Hash("123"),
			mockManifestHashUpdated: cache.Hash("789"),
			expectedApp: types.App{
				AppID:  "A002",
				IsDev:  true,
				TeamID: mockTeamID,
				UserID: mockUserID,
			},
			expectedManifest: types.AppManifest{
				Metadata: &types.ManifestMetadata{
					MajorVersion: 1,
				},
				DisplayInformation: types.DisplayInformation{
					Name: "example-2 (local)",
				},
				Features: &types.AppFeatures{
					BotUser: types.BotUser{
						DisplayName: "example-2 (local)",
					},
				},
				Settings: &types.AppSettings{
					FunctionRuntime: types.Remote,
					EventSubscriptions: &types.ManifestEventSubscriptions{
						RequestURL: "https://example.com",
					},
				},
			},
			expectedUpdate: true,
		},
		"update an existing local bolt app without a function runtime without manifest changes": {
			isExperimental: true,
			mockApp: types.App{
				AppID:  "A003",
				TeamID: mockTeamID,
				UserID: mockUserID,
			},
			mockAPICreateError: slackerror.New(slackerror.ErrAppCreate),
			mockAPIUpdate: api.UpdateAppResult{
				AppID: "A003",
			},
			mockAuth: types.SlackAuth{
				EnterpriseID: mockEnterpriseID,
				TeamID:       mockTeamID,
				TeamDomain:   mockTeamDomain,
				Token:        mockToken,
				UserID:       mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockBoltExperiment: true,
			mockConfirmPrompt:  true,
			mockIsTTY:          true,
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "example-3",
					},
					Features: &types.AppFeatures{
						BotUser: types.BotUser{
							DisplayName: "example-3",
						},
					},
					Settings: &types.AppSettings{
						SocketModeEnabled: &mockTrue,
					},
				},
			},
			mockManifestHashInitial: cache.Hash("abc"),
			mockManifestHashUpdated: cache.Hash("def"),
			expectedApp: types.App{
				AppID:  "A003",
				IsDev:  true,
				TeamID: mockTeamID,
				UserID: mockUserID,
			},
			expectedManifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "example-3 (local)",
				},
				Features: &types.AppFeatures{
					BotUser: types.BotUser{
						DisplayName: "example-3 (local)",
					},
				},
				Settings: &types.AppSettings{
					SocketModeEnabled: &mockTrue,
				},
			},
			expectedUpdate: true,
		},
		"avoids updating or installing an app with a remote manifest": {
			mockApp: types.App{
				AppID:  "A004",
				IsDev:  true,
				TeamID: mockTeamID,
				UserID: mockUserID,
			},
			mockAuth: types.SlackAuth{
				TeamID:     mockTeamID,
				TeamDomain: mockTeamDomain,
				Token:      mockToken,
				UserID:     mockUserID,
			},
			mockAuthSession: api.AuthSession{
				TeamID:   &mockTeamID,
				TeamName: &mockTeamDomain,
				UserID:   &mockUserID,
			},
			mockAPICreateError:  slackerror.New(slackerror.ErrAppCreate),
			mockAPIUpdateError:  slackerror.New(slackerror.ErrAppAdd),
			mockAPIInstallError: slackerror.New(slackerror.ErrAppInstall),
			mockBoltExperiment:  true,
			mockManifestSource:  config.ManifestSourceRemote,
			expectedApp: types.App{
				AppID:  "A004",
				IsDev:  true,
				TeamID: mockTeamID,
				UserID: mockUserID,
			},
			expectedCreate:       false,
			expectedInstallState: "",
			expectedUpdate:       false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("IsTTY").Return(tt.mockIsTTY)
			clientsMock.AddDefaultMocks()
			clientsMock.API.On(
				"CreateApp",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAPICreate,
				tt.mockAPICreateError,
			)
			clientsMock.API.On(
				"DeveloperAppInstall",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAPIInstall,
				tt.mockAPIInstallState,
				tt.mockAPIInstallError,
			)
			clientsMock.API.On(
				"ExportAppManifest",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				api.ExportAppResult{Manifest: tt.mockManifest},
				nil,
			)
			clientsMock.API.On(
				"ValidateAppManifest",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				tt.mockApp.AppID,
			).Return(
				api.ValidateAppManifestResult{},
				nil,
			)
			clientsMock.API.On(
				"UpdateApp",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAPIUpdate,
				tt.mockAPIUpdateError,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAuthSession,
				nil,
			)
			if tt.mockIsTTY {
				clientsMock.IO.On(
					"ConfirmPrompt",
					mock.Anything,
					"Update app settings with changes to the local manifest?",
					false,
				).Return(
					tt.mockConfirmPrompt,
					nil,
				)
			}
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockManifest, nil)
			clientsMock.AppClient.Manifest = manifestMock
			mockProjectConfig := config.NewProjectConfigMock()
			if tt.mockBoltExperiment {
				clientsMock.Config.ExperimentsFlag = append(clientsMock.Config.ExperimentsFlag, string(experiment.BoltFrameworks))
				clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)
				mockProjectConfig.On("GetManifestSource", mock.Anything).Return(tt.mockManifestSource, nil)
			}
			mockProjectCache := cache.NewCacheMock()
			mockProjectCache.On(
				"GetManifestHash",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestHashInitial,
				nil,
			)
			mockProjectCache.On(
				"NewManifestHash",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestHashUpdated,
				nil,
			)
			mockProjectCache.On(
				"SetManifestHash",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(nil)
			mockProjectConfig.On("Cache").Return(mockProjectCache)
			clientsMock.Config.ProjectConfig = mockProjectConfig

			log := logger.New(func(event *logger.LogEvent) {})
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			app, _, state, err := InstallLocalApp(
				ctx,
				clients,
				tt.mockOrgGrantWorkspaceID,
				log,
				tt.mockAuth,
				tt.mockApp,
			)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedInstallState, state)
			assert.Equal(t, tt.expectedApp, app)
			if tt.expectedUpdate {
				clientsMock.API.AssertCalled(
					t,
					"UpdateApp",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				clientsMock.API.AssertNotCalled(t, "CreateApp")
			} else if tt.expectedCreate {
				clientsMock.API.AssertCalled(
					t,
					"CreateApp",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				clientsMock.API.AssertNotCalled(t, "UpdateApp")
			}
			for _, call := range clientsMock.API.Calls {
				args := call.Arguments
				switch call.Method {
				case "CreateApp":
					assert.Equal(t, tt.mockAuth.Token, args.Get(1))
					assert.Equal(t, tt.expectedManifest, args.Get(2))
				case "UpdateApp":
					assert.Equal(t, tt.mockAuth.Token, args.Get(1))
					assert.Equal(t, tt.mockApp.AppID, args.Get(2))
					assert.Equal(t, tt.expectedManifest, args.Get(3))
				}
			}
		})
	}
}

func TestValidateManifestForInstall(t *testing.T) {
	tests := map[string]struct {
		app      types.App
		manifest types.AppManifest
		result   api.ValidateAppManifestResult
		err      error
		setup    func(cm *shared.ClientsMock)
		check    func(cm *shared.ClientsMock)
	}{
		"no errors or warnings for a nil response": {
			app:      types.App{AppID: "A123"},
			manifest: types.AppManifest{},
			result: api.ValidateAppManifestResult{
				Warnings: nil,
			},
			err: nil,
			setup: func(cm *shared.ClientsMock) {
				cm.AddDefaultMocks()
			},
			check: func(cm *shared.ClientsMock) {
				assert.NotContains(t, cm.GetCombinedOutput(), additionalManifestInfoNotice)
			},
		},
		"no errors or warnings for a valid manifest": {
			app:      types.App{AppID: "A123"},
			manifest: types.AppManifest{},
			result: api.ValidateAppManifestResult{
				Warnings: slackerror.Warnings{},
			},
			err: nil,
			setup: func(cm *shared.ClientsMock) {
				cm.AddDefaultMocks()
			},
			check: func(cm *shared.ClientsMock) {
				assert.NotContains(t, cm.GetCombinedOutput(), additionalManifestInfoNotice)
			},
		},
		"include manifest warnings when present": {
			app:      types.App{AppID: "A123"},
			manifest: types.AppManifest{},
			result: api.ValidateAppManifestResult{
				Warnings: slackerror.Warnings{
					slackerror.Warning{
						Code:    "invalid_manifest_field",
						Message: "Something isn't right with the manifest",
					},
				}},
			err: nil,
			setup: func(cm *shared.ClientsMock) {
				cm.AddDefaultMocks()
				cm.Config = &config.Config{ForceFlag: false} // force flag is not enabled
			},
			check: func(cm *shared.ClientsMock) {
				assert.Contains(t, cm.GetCombinedOutput(), additionalManifestInfoNotice)
			},
		},
		"don't include manifest warnings the --force flag is set": {
			app:      types.App{AppID: "A123"},
			manifest: types.AppManifest{},
			result: api.ValidateAppManifestResult{
				Warnings: slackerror.Warnings{
					slackerror.Warning{
						Code:    "breaking_change",
						Message: "You're going to break existing workflows",
					},
				}},
			err: nil,
			setup: func(cm *shared.ClientsMock) {
				cm.AddDefaultMocks()
				cm.Config = &config.Config{ForceFlag: true}
			},
			check: func(cm *shared.ClientsMock) {
				assert.NotContains(t, cm.GetCombinedOutput(), additionalManifestInfoNotice)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			tt.setup(clientsMock)
			clientsMock.API.On("ValidateAppManifest", mock.Anything, mock.Anything, mock.Anything, tt.app.AppID).
				Return(tt.result, tt.err)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			err := validateManifestForInstall(ctx, clients, tt.app, tt.manifest)
			assert.NoError(t, err)

			tt.check(clientsMock)
		})
	}
}

func TestSetAppEnvironmentTokens(t *testing.T) {
	tests := map[string]struct {
		envAppToken      string
		envBotToken      string
		result           api.DeveloperAppInstallResult
		expectedAppToken string
		expectedBotToken string
		expectedOutput   string
	}{
		"defaults to resulting tokens": {
			result: api.DeveloperAppInstallResult{
				APIAccessTokens: struct {
					Bot      string `json:"bot,omitempty"`
					AppLevel string `json:"app_level,omitempty"`
					User     string `json:"user,omitempty"`
				}{
					Bot:      "xoxb-1234",
					AppLevel: "xapp-1-wf",
				},
			},
			expectedBotToken: "xoxb-1234",
			expectedAppToken: "xapp-1-wf",
		},
		"does not override existing app token": {
			envAppToken: "xapp-2-custom",
			result: api.DeveloperAppInstallResult{
				APIAccessTokens: struct {
					Bot      string `json:"bot,omitempty"`
					AppLevel string `json:"app_level,omitempty"`
					User     string `json:"user,omitempty"`
				}{
					Bot:      "xoxb-0000",
					AppLevel: "xapp-1-wf",
				},
			},
			expectedBotToken: "xoxb-0000",
			expectedAppToken: "xapp-2-custom",
			expectedOutput:   "The app token differs from the set SLACK_APP_TOKEN environment variable",
		},
		"does not override existing bot token": {
			envBotToken: "xoxb-unique",
			result: api.DeveloperAppInstallResult{
				APIAccessTokens: struct {
					Bot      string `json:"bot,omitempty"`
					AppLevel string `json:"app_level,omitempty"`
					User     string `json:"user,omitempty"`
				}{
					Bot:      "xoxb-beep",
					AppLevel: "xapp-4-fn",
				},
			},
			expectedBotToken: "xoxb-unique",
			expectedAppToken: "xapp-4-fn",
			expectedOutput:   "The bot token differs from the set SLACK_BOT_TOKEN environment variable",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.AddDefaultMocks()
			if tt.envAppToken != "" {
				clientsMock.Os.On("Getenv", "SLACK_APP_TOKEN").Return(tt.envAppToken)
			}
			if tt.envBotToken != "" {
				clientsMock.Os.On("Getenv", "SLACK_BOT_TOKEN").Unset()
				clientsMock.Os.On("Getenv", "SLACK_BOT_TOKEN").Return(tt.envBotToken)
			}
			clientsMock.Os.On("LookupEnv", "SLACK_APP_TOKEN").
				Return(tt.envAppToken, tt.envAppToken != "")
			clientsMock.Os.On("LookupEnv", "SLACK_BOT_TOKEN").
				Return(tt.envBotToken, tt.envBotToken != "")
			clientsMock.Os.On("Setenv", "SLACK_APP_TOKEN", mock.Anything).Return(nil)
			clientsMock.Os.On("Setenv", "SLACK_BOT_TOKEN", mock.Anything).Return(nil)
			output := &bytes.Buffer{}
			clientsMock.IO.Stdout.SetOutput(output)

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			err := setAppEnvironmentTokens(ctx, clients, tt.result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAppToken, clients.Os.Getenv("SLACK_APP_TOKEN"))
			assert.Equal(t, tt.expectedBotToken, clients.Os.Getenv("SLACK_BOT_TOKEN"))
			assert.Contains(t, output.String(), tt.expectedOutput)
		})
	}
}
