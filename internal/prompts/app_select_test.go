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

package prompts

import (
	"fmt"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Testing Mock Constants

// Mock method names
const (
	AuthWithTeamID = "AuthWithTeamID"
	AuthWithToken  = "AuthWithToken"
	Auths          = "Auths"
	GetAppStatus   = "GetAppStatus"
	SelectPrompt   = "SelectPrompt"
	SetAuth        = "SetAuth"
)

// Mock teams
const (
	// Team1
	team1TeamDomain = "team1"
	team1TeamID     = "T1"
	team1UserID     = "U1"
	team1Token      = "xoxe.xoxp-1-token"

	// Team2
	team2TeamDomain = "team2"
	team2TeamID     = "T2"
	team2UserID     = "U2"
	team2Token      = "xoxe.xoxp-2-token"

	// Enterprise1
	enterprise1TeamDomain = "enterprise1"
	enterprise1TeamID     = "E1"
	enterprise1UserID     = "U3"
	enterprise1Token      = "xoxe.xoxp-3-token"
)

// Mock team convenience structures

// IMPORTANT: Auths keyed by TeamDomain is legacy
var fakeAuthsByTeamDomain = map[string]types.SlackAuth{
	team1TeamDomain: {
		TeamDomain: team1TeamDomain,
		TeamID:     team1TeamID,
		UserID:     team1UserID,
		Token:      team1Token,
	},
	team2TeamDomain: {
		TeamDomain: team2TeamDomain,
		TeamID:     team2TeamID,
		UserID:     team2UserID,
		Token:      team2Token,
	},
}

var fakeAuthsByTeamDomainSlice = getValues(fakeAuthsByTeamDomain)

func getValues(auths map[string]types.SlackAuth) []types.SlackAuth {
	var slice = []types.SlackAuth{}
	for _, fakeAuth := range auths {
		slice = append(slice, fakeAuth)
	}
	return slice
}

// Mock app IDs and install status
const (
	deployedTeam1InstalledAppID   = "A1"
	deployedTeam1AppIsInstalled   = true
	deployedTeam2UninstalledAppID = "A2"
	deployedTeam2AppIsInstalled   = false
	localTeam1UninstalledAppID    = "A3"
	localTeam1AppIsInstalled      = false
	localTeam2InstalledAppID      = "A4"
	localTeam2AppIsInstalled      = true
)

// Mock app structs
var (
	deployedTeam1InstalledApp = types.App{
		AppID:         deployedTeam1InstalledAppID,
		TeamDomain:    team1TeamDomain,
		TeamID:        team1TeamID,
		InstallStatus: types.AppStatusInstalled,
	}
	deployedTeam2UninstalledApp = types.App{
		AppID:         deployedTeam2UninstalledAppID,
		TeamDomain:    team2TeamDomain,
		TeamID:        team2TeamID,
		InstallStatus: types.AppStatusUninstalled,
	}
	localTeam1UninstalledApp = types.App{
		AppID:         localTeam1UninstalledAppID,
		TeamDomain:    team1TeamDomain,
		TeamID:        team1TeamID,
		IsDev:         true,
		UserID:        team1UserID,
		InstallStatus: types.AppStatusUninstalled,
	}
	localTeam2InstalledApp = types.App{
		AppID:         localTeam2InstalledAppID,
		TeamDomain:    team2TeamDomain,
		TeamID:        team2TeamID,
		IsDev:         true,
		UserID:        team2UserID,
		InstallStatus: types.AppStatusInstalled,
	}
)

// Mock GetAppStatus responses
var (
	deployedTeam1InstalledAppStatus = api.AppStatusResultAppInfo{
		AppID:     deployedTeam1InstalledAppID,
		Installed: true,
		Hosted:    true,
	}
	deployedTeam2UninstalledAppStatus = api.AppStatusResultAppInfo{
		AppID:     deployedTeam2UninstalledAppID,
		Installed: false,
		Hosted:    true,
	}
	localTeam1UninstallAppStatus = api.AppStatusResultAppInfo{
		AppID:     localTeam1UninstalledAppID,
		Installed: false,
		Hosted:    false,
	}
	localTeam2InstalledAppStatus = api.AppStatusResultAppInfo{
		AppID:     localTeam2InstalledAppID,
		Installed: true,
		Hosted:    false,
	}
)

// App environment select labels
const (
	Deployed = "Deployed"
	Local    = "Local"
)

//
// getTokenApp tests
//

func TestGetTokenApp(t *testing.T) {
	tests := map[string]struct {
		tokenFlag string
		tokenAuth types.SlackAuth
		tokenErr  error
		saveLocal []types.App
		appFlag   string
		appStatus api.GetAppStatusResult
		appInfo   types.App
		statusErr error
	}{
		"error on an unknown token": {
			tokenFlag: "xoxp-unknown",
			tokenAuth: types.SlackAuth{},
			tokenErr:  slackerror.New(slackerror.ErrInvalidAuth),
			appFlag:   deployedTeam1InstalledAppID,
			appStatus: api.GetAppStatusResult{},
			appInfo:   types.App{},
			statusErr: nil,
		},
		"error on an unknown app": {
			tokenFlag: team1Token,
			tokenAuth: fakeAuthsByTeamDomain[team1TeamDomain],
			tokenErr:  nil,
			appFlag:   "A01001101",
			appStatus: api.GetAppStatusResult{},
			appInfo:   types.App{},
			statusErr: slackerror.New(slackerror.ErrHTTPRequestFailed),
		},
		"error if no app status is returned": {
			tokenFlag: team1Token,
			tokenAuth: fakeAuthsByTeamDomain[team1TeamDomain],
			tokenErr:  nil,
			appFlag:   deployedTeam1InstalledAppID,
			appStatus: api.GetAppStatusResult{},
			appInfo:   types.App{},
			statusErr: slackerror.New(slackerror.ErrAppNotFound),
		},
		"return an uninstalled hosted app": {
			tokenFlag: team2Token,
			tokenAuth: fakeAuthsByTeamDomain[team2TeamDomain],
			tokenErr:  nil,
			appFlag:   deployedTeam2UninstalledAppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam2UninstalledAppID,
					Installed: deployedTeam2AppIsInstalled,
					Hosted:    true,
				}},
			},
			appInfo:   deployedTeam2UninstalledApp,
			statusErr: nil,
		},
		"return an uninstalled but saved local app": {
			tokenAuth: fakeAuthsByTeamDomain[team1TeamDomain],
			tokenFlag: team1Token,
			tokenErr:  nil,
			saveLocal: []types.App{localTeam1UninstalledApp},
			appFlag:   localTeam1UninstalledAppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     localTeam1UninstalledAppID,
					Installed: localTeam1AppIsInstalled,
					Hosted:    false,
				}},
			},
			appInfo:   localTeam1UninstalledApp,
			statusErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(AuthWithToken, mock.Anything, test.tokenFlag).
				Return(test.tokenAuth, test.tokenErr)
			clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(test.appStatus, test.statusErr)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range test.saveLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			selection, err := getTokenApp(ctx, clients, test.tokenFlag, test.appFlag)

			if test.tokenErr != nil && assert.Error(t, err) {
				require.Equal(t, test.tokenErr, err)
			} else if test.statusErr != nil && assert.Error(t, err) {
				require.Equal(t, test.statusErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.tokenAuth, selection.Auth)
				expectedApp := test.appInfo
				expectedApp.UserID = test.tokenAuth.UserID
				assert.Equal(t, expectedApp, selection.App)
			}
		})
	}
}

//
// AppSelectPrompt tests
//

func TestPrompt_AppSelectPrompt_TokenAppFlag(t *testing.T) {
	tests := map[string]struct {
		tokenFlag    string
		tokenAuth    types.SlackAuth
		appFlag      string
		appStatus    api.GetAppStatusResult
		statusErr    error
		selectStatus AppInstallStatus
		expectedApp  SelectedApp
		expectedErr  error
	}{
		"error if an error occurred while collecting app info": {
			tokenFlag:    team1Token,
			tokenAuth:    fakeAuthsByTeamDomain[team1TeamDomain],
			appFlag:      localTeam1UninstalledApp.AppID,
			appStatus:    api.GetAppStatusResult{},
			statusErr:    slackerror.New(slackerror.ErrAppNotFound),
			selectStatus: ShowAllApps,
			expectedApp:  SelectedApp{},
			expectedErr:  slackerror.New(slackerror.ErrAppNotFound),
		},
		"error if an uninstalled app is used for an installed only prompt": {
			tokenFlag: team2Token,
			tokenAuth: fakeAuthsByTeamDomain[team2TeamDomain],
			appFlag:   deployedTeam2UninstalledApp.AppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam2UninstalledApp.AppID,
					Installed: deployedTeam2AppIsInstalled,
					Hosted:    true,
				}},
			},
			statusErr:    nil,
			selectStatus: ShowInstalledAppsOnly,
			expectedApp:  SelectedApp{},
			expectedErr:  slackerror.New(slackerror.ErrInstallationRequired),
		},
		"returns known information about the requested app": {
			tokenFlag: team1Token,
			tokenAuth: fakeAuthsByTeamDomain[team1TeamDomain],
			appFlag:   deployedTeam1InstalledAppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam1InstalledAppID,
					Installed: deployedTeam1AppIsInstalled,
					Hosted:    true,
				}},
			},
			statusErr:    nil,
			selectStatus: ShowInstalledAppsOnly,
			expectedApp: SelectedApp{
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				App:  deployedTeam1InstalledApp,
			},
			expectedErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(AuthWithToken, mock.Anything, test.tokenFlag).
				Return(test.tokenAuth, nil)
			clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(test.appStatus, test.statusErr)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			clients.Config.TokenFlag = test.tokenFlag
			clients.Config.AppFlag = test.appFlag

			selection, err := AppSelectPrompt(ctx, clients, test.selectStatus)

			if test.statusErr != nil && assert.Error(t, err) {
				require.Equal(t, test.statusErr, err)
			} else if test.expectedErr != nil && assert.Error(t, err) {
				require.Equal(t, test.expectedErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedApp.Auth, selection.Auth)
				expectedApp := test.expectedApp.App
				expectedApp.UserID = test.expectedApp.Auth.UserID
				assert.Equal(t, expectedApp, selection.App)
			}
		})
	}
}

func TestPrompt_AppSelectPrompt_GetApps(t *testing.T) {
	tests := map[string]struct {
		mockAppsSavedDeployed         []types.App
		mockAppsSavedDeployedError    error
		mockAppsSavedLocal            []types.App
		mockAppsSavedLocalError       error
		mockAuths                     []types.SlackAuth
		mockEnterprise1SavedAuth      types.SlackAuth
		mockEnterprise1SavedAuthError error
		mockEnterprise2SavedAuth      types.SlackAuth
		mockEnterprise2SavedAuthError error
		mockTeam1SavedAuth            types.SlackAuth
		mockTeam1SavedAuthError       error
		mockTeam1SavedDeployed        types.App
		mockTeam1SavedDeployedError   error
		mockTeam1SavedLocal           types.App
		mockTeam1SavedLocalError      error
		mockTeam1Status               api.GetAppStatusResult
		mockTeam1StatusAppIDs         []string
		mockTeam1StatusError          error
		mockTeam2SavedAuth            types.SlackAuth
		mockTeam2SavedAuthError       error
		mockTeam2SavedDeployed        types.App
		mockTeam2SavedDeployedError   error
		mockTeam2SavedLocal           types.App
		mockTeam2SavedLocalError      error
		mockTeam2Status               api.GetAppStatusResult
		mockTeam2StatusAppIDs         []string
		mockTeam2StatusError          error
		expectedApps                  map[string]SelectedApp
		expectedError                 error
	}{
		"returns deployed and local apps with matching auths": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsSavedDeployed: []types.App{
				deployedTeam1InstalledApp,
				deployedTeam2UninstalledApp,
			},
			mockAppsSavedLocal: []types.App{
				localTeam1UninstalledApp,
				localTeam2InstalledApp,
			},
			mockTeam1StatusAppIDs: []string{
				deployedTeam1InstalledAppID,
				localTeam1UninstalledAppID,
			},
			mockTeam1Status: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					deployedTeam1InstalledAppStatus,
					localTeam1UninstallAppStatus,
				},
			},
			mockTeam2StatusAppIDs: []string{
				deployedTeam2UninstalledAppID,
				localTeam2InstalledAppID,
			},
			mockTeam2Status: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					deployedTeam2UninstalledAppStatus,
					localTeam2InstalledAppStatus,
				},
			},
			expectedApps: map[string]SelectedApp{
				deployedTeam1InstalledAppID: {
					App:  deployedTeam1InstalledApp,
					Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				},
				localTeam1UninstalledAppID: {
					App:  localTeam1UninstalledApp,
					Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				},
				deployedTeam2UninstalledAppID: {
					App:  deployedTeam2UninstalledApp,
					Auth: fakeAuthsByTeamDomain[team2TeamDomain],
				},
				localTeam2InstalledAppID: {
					App:  localTeam2InstalledApp,
					Auth: fakeAuthsByTeamDomain[team2TeamDomain],
				},
			},
		},
		"returns enterprise workspace apps with matching auths": {
			mockAuths: []types.SlackAuth{
				{
					Token:        enterprise1Token,
					TeamDomain:   enterprise1TeamDomain,
					TeamID:       enterprise1TeamID,
					EnterpriseID: enterprise1TeamID,
				},
			},
			mockAppsSavedDeployed: []types.App{
				{
					AppID:        deployedTeam1InstalledAppID,
					EnterpriseID: enterprise1TeamID,
					TeamID:       team1TeamID,
				},
			},
			mockAppsSavedLocal: []types.App{
				{
					AppID:        localTeam2InstalledAppID,
					EnterpriseID: enterprise1TeamID,
					TeamID:       team2TeamID,
				},
			},
			mockTeam1SavedAuthError: slackerror.New(slackerror.ErrCredentialsNotFound),
			mockTeam1StatusAppIDs: []string{
				deployedTeam1InstalledAppID,
			},
			mockTeam1Status: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					deployedTeam1InstalledAppStatus,
				},
			},
			mockTeam2SavedAuthError: slackerror.New(slackerror.ErrCredentialsNotFound),
			mockTeam2StatusAppIDs: []string{
				localTeam2InstalledAppID,
			},
			mockTeam2Status: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					localTeam2InstalledAppStatus,
				},
			},
			mockEnterprise1SavedAuth: types.SlackAuth{
				Token: enterprise1Token,
			},
			expectedApps: map[string]SelectedApp{
				deployedTeam1InstalledAppID: {
					App: types.App{
						AppID:         deployedTeam1InstalledAppID,
						EnterpriseID:  enterprise1TeamID,
						TeamID:        team1TeamID,
						InstallStatus: types.AppStatusInstalled,
					},
					Auth: types.SlackAuth{
						Token: enterprise1Token,
					},
				},
				localTeam2InstalledAppID: {
					App: types.App{
						AppID:         localTeam2InstalledAppID,
						EnterpriseID:  enterprise1TeamID,
						TeamID:        team2TeamID,
						InstallStatus: types.AppStatusInstalled,
						IsDev:         true,
					},
					Auth: types.SlackAuth{
						Token: enterprise1Token,
					},
				},
			},
		},
		"returns unknown installation statuses for apps without auths": {
			mockAppsSavedDeployed: []types.App{
				deployedTeam1InstalledApp,
			},
			mockTeam1StatusAppIDs: []string{
				deployedTeam1InstalledAppID,
			},
			mockTeam1StatusError: fmt.Errorf("404"),
			expectedApps: map[string]SelectedApp{
				deployedTeam1InstalledAppID: {
					App: types.App{
						AppID:         deployedTeam1InstalledAppID,
						TeamDomain:    team1TeamDomain,
						TeamID:        team1TeamID,
						InstallStatus: types.AppInstallationStatusUnknown,
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				tt.mockTeam1StatusAppIDs,
				mock.Anything,
			).Return(
				tt.mockTeam1Status,
				tt.mockTeam1StatusError,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				tt.mockTeam2StatusAppIDs,
				mock.Anything,
			).Return(
				tt.mockTeam2Status,
				tt.mockTeam2StatusError,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				mock.Anything,
			).Return(
				api.AuthSession{},
				nil,
			)
			clientsMock.Auth.On(
				Auths,
				mock.Anything,
			).Return(
				tt.mockAuths,
				nil,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				team2TeamID,
			).Return(
				tt.mockTeam2SavedAuth,
				tt.mockTeam2SavedAuthError,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				team1TeamID,
			).Return(
				tt.mockTeam1SavedAuth,
				tt.mockTeam1SavedAuthError,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				enterprise1TeamID,
			).Return(
				tt.mockEnterprise1SavedAuth,
				tt.mockEnterprise1SavedAuthError,
			)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range tt.mockAppsSavedDeployed {
				err := clients.AppClient().SaveDeployed(ctx, app)
				require.NoError(t, err)
			}
			for _, app := range tt.mockAppsSavedLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			apps, err := getApps(ctx, clients)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedApps, apps)
		})
	}
}

func TestPrompt_AppSelectPrompt_FlatAppSelectPrompt(t *testing.T) {
	tests := map[string]struct {
		mockAuths                  []types.SlackAuth
		mockAuthWithTeamIDError    error
		mockAuthWithTeamIDTeamID   string
		mockAuthWithToken          types.SlackAuth
		mockAppsDeployed           []types.App
		mockAppsLocal              []types.App
		mockFlagApp                string
		mockFlagTeam               string
		mockFlagToken              string
		mockManifestSource         config.ManifestSource
		appPromptConfigEnvironment AppEnvironmentType
		appPromptConfigOptions     []string
		appPromptConfigStatus      AppInstallStatus
		appPromptResponseFlag      bool
		appPromptResponsePrompt    bool
		appPromptResponseOption    string
		appPromptResponseIndex     int
		teamPromptResponseFlag     bool
		teamPromptResponsePrompt   bool
		teamPromptResponseOption   string
		teamPromptResponseIndex    int
		expectedError              error
		expectedSelection          SelectedApp
		expectedStdout             string
		expectedStderr             string
	}{
		"returns a saved applications using prompts": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
				{
					AppID:      deployedTeam2UninstalledAppID,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
				},
			},
			mockAppsLocal: []types.App{
				{
					AppID:      localTeam1UninstalledAppID,
					IsDev:      true,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
					UserID:     team1UserID,
				},
				{
					AppID:      localTeam2InstalledAppID,
					IsDev:      true,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
					UserID:     team2UserID,
				},
			},
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowAllEnvironments,
			appPromptConfigOptions: []string{
				"A1 team1 T1",
				"A3 team1 T1",
				"A2 team2 T2",
				"A4 team2 T2",
			},
			appPromptConfigStatus:   ShowInstalledAndUninstalledApps,
			appPromptResponsePrompt: true,
			appPromptResponseOption: "ExampleApplication123",
			appPromptResponseIndex:  0,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         deployedTeam1InstalledAppID,
					TeamDomain:    team1TeamDomain,
					TeamID:        team1TeamID,
					InstallStatus: types.AppStatusInstalled,
				},
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns new application if selected": {
			mockAuths:                  fakeAuthsByTeamDomainSlice,
			mockAppsDeployed:           []types.App{},
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigOptions: []string{
				"Create a new app",
			},
			appPromptConfigStatus:    ShowInstalledAndNewApps,
			appPromptResponsePrompt:  true,
			appPromptResponseOption:  "Create a new app",
			teamPromptResponseFlag:   true,
			teamPromptResponseOption: team1TeamID,
			expectedSelection: SelectedApp{
				App:  types.NewApp(),
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
			expectedStdout: "Installed apps will belong to the team if you leave the workspace",
		},
		"errors if installation required and no apps saved": {
			mockAuths:                  fakeAuthsByTeamDomainSlice,
			mockAppsDeployed:           []types.App{},
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowInstalledAppsOnly,
			expectedError:              slackerror.New(slackerror.ErrInstallationRequired),
			expectedSelection:          SelectedApp{},
		},
		"errors if creating app while app is saved using selections": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
			},
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigOptions: []string{
				"A1 team1 T1",
				"Create a new app",
			},
			appPromptConfigStatus:    ShowInstalledAndNewApps,
			appPromptResponsePrompt:  true,
			appPromptResponseIndex:   1,
			appPromptResponseOption:  "Create a new app",
			teamPromptResponseFlag:   true,
			teamPromptResponseOption: team1TeamDomain,
			mockManifestSource:       config.ManifestSourceLocal,
			expectedError: slackerror.New(slackerror.ErrAppExists).
				WithDetails(slackerror.ErrorDetails{{
					Message: `The app "A1" already exists for team "team1" (T1)`,
				}}).
				WithRemediation("To learn more run: %s", style.Commandf("app list", false)),
		},
		"returns new application for app environment flag and team id flag if not app saved": {
			mockAuths:                  fakeAuthsByTeamDomainSlice,
			mockFlagApp:                "deployed",
			mockFlagTeam:               team1TeamID,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedSelection: SelectedApp{
				App:  types.NewApp(),
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"selects existing application for app environment flag and team id flag if app saved": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam2UninstalledAppID,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
				},
			},
			mockFlagApp:                "deployed",
			mockFlagTeam:               team2TeamID,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowAllApps,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         deployedTeam2UninstalledAppID,
					TeamDomain:    team2TeamDomain,
					TeamID:        team2TeamID,
					InstallStatus: types.AppStatusUninstalled,
				},
				Auth: fakeAuthsByTeamDomain[team2TeamDomain],
			},
		},
		"returns filtered deployed apps for app environment flag before selection": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
				{
					AppID:      deployedTeam2UninstalledAppID,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
				},
			},
			mockAppsLocal: []types.App{
				{
					AppID:      localTeam1UninstalledAppID,
					IsDev:      true,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
					UserID:     team1UserID,
				},
				{
					AppID:      localTeam2InstalledAppID,
					IsDev:      true,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
					UserID:     team2UserID,
				},
			},
			mockFlagApp:                "deploy",
			appPromptConfigEnvironment: ShowAllEnvironments,
			appPromptConfigOptions: []string{
				"A1 team1 T1",
				"A2 team2 T2",
			},
			appPromptResponsePrompt: true,
			appPromptResponseIndex:  1,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         deployedTeam2UninstalledAppID,
					TeamDomain:    team2TeamDomain,
					TeamID:        team2TeamID,
					InstallStatus: types.AppStatusUninstalled,
				},
				Auth: fakeAuthsByTeamDomain[team2TeamDomain],
			},
		},
		"returns filtered local apps for app environment flag before selection": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
				{
					AppID:      deployedTeam2UninstalledAppID,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
				},
			},
			mockAppsLocal: []types.App{
				{
					AppID:      localTeam1UninstalledAppID,
					IsDev:      true,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
					UserID:     team1UserID,
				},
				{
					AppID:      localTeam2InstalledAppID,
					IsDev:      true,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
					UserID:     team2UserID,
				},
			},
			mockFlagApp:                "local",
			appPromptConfigEnvironment: ShowAllEnvironments,
			appPromptConfigOptions: []string{
				"A3 team1 T1",
				"A4 team2 T2",
			},
			appPromptConfigStatus:   ShowInstalledAndUninstalledApps,
			appPromptResponsePrompt: true,
			appPromptResponseIndex:  1,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         localTeam2InstalledAppID,
					IsDev:         true,
					TeamDomain:    team2TeamDomain,
					TeamID:        team2TeamID,
					UserID:        team2UserID,
					InstallStatus: types.AppStatusInstalled,
				},
				Auth: fakeAuthsByTeamDomain[team2TeamDomain],
			},
		},
		"returns selection for app id flag": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsLocal: []types.App{
				{
					AppID:      localTeam1UninstalledAppID,
					IsDev:      true,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
					UserID:     team1UserID,
				},
			},
			mockFlagApp:                localTeam1UninstalledAppID,
			appPromptConfigEnvironment: ShowAllEnvironments,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         localTeam1UninstalledAppID,
					IsDev:         true,
					TeamDomain:    team1TeamDomain,
					TeamID:        team1TeamID,
					UserID:        team1UserID,
					InstallStatus: types.AppStatusUninstalled,
				},
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns selection for app id flag and team id flag": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
			},
			mockFlagApp:                deployedTeam1InstalledAppID,
			mockFlagTeam:               team1TeamID,
			appPromptConfigEnvironment: ShowHostedOnly,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         deployedTeam1InstalledAppID,
					TeamDomain:    team1TeamDomain,
					TeamID:        team1TeamID,
					InstallStatus: types.AppStatusInstalled,
				},
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns selection for app id flag and team domain flag": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
			},
			mockFlagApp:                deployedTeam1InstalledAppID,
			mockFlagTeam:               team1TeamDomain,
			appPromptConfigEnvironment: ShowAllEnvironments,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         deployedTeam1InstalledAppID,
					TeamDomain:    team1TeamDomain,
					TeamID:        team1TeamID,
					InstallStatus: types.AppStatusInstalled,
				},
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"errors if app id flag has a team id flag that does not match": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsDeployed: []types.App{
				{
					AppID:      deployedTeam1InstalledAppID,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
				},
			},
			mockFlagApp:                deployedTeam1InstalledAppID,
			mockFlagTeam:               team2TeamID,
			appPromptConfigEnvironment: ShowHostedOnly,
			expectedError:              slackerror.New(slackerror.ErrAppNotFound),
		},
		"returns selection for app environment flag and team id flag": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsLocal: []types.App{
				{
					AppID:      localTeam1UninstalledAppID,
					IsDev:      true,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
					UserID:     team1UserID,
				},
				{
					AppID:      localTeam2InstalledAppID,
					IsDev:      true,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
					UserID:     team2UserID,
				},
			},
			mockFlagApp:                "local",
			mockFlagTeam:               team1TeamID,
			appPromptConfigEnvironment: ShowAllEnvironments,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         localTeam1UninstalledAppID,
					IsDev:         true,
					TeamDomain:    team1TeamDomain,
					TeamID:        team1TeamID,
					UserID:        team1UserID,
					InstallStatus: types.AppStatusUninstalled,
				},
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"errors if app flag does not match team flag": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsLocal: []types.App{
				{
					AppID:      localTeam1UninstalledAppID,
					IsDev:      true,
					TeamDomain: team1TeamDomain,
					TeamID:     team1TeamID,
					UserID:     team1UserID,
				},
			},
			mockFlagApp:                "local",
			mockFlagTeam:               "TNOTFOUND",
			appPromptConfigEnvironment: ShowAllEnvironments,
			appPromptConfigStatus:      ShowAllApps,
			expectedError:              slackerror.New(slackerror.ErrAppNotFound),
		},
		"errors if deployed app environment flag for local app prompt": {
			mockFlagApp:                "deploy",
			appPromptConfigEnvironment: ShowLocalOnly,
			expectedError:              slackerror.New(slackerror.ErrDeployedAppNotSupported),
		},
		"errors if local app environment flag for deployed app prompt": {
			mockFlagApp:                "local",
			appPromptConfigEnvironment: ShowHostedOnly,
			expectedError:              slackerror.New(slackerror.ErrLocalAppNotSupported),
		},
		"errors if deployed app environment flag and team id flag for local app prompt": {
			mockFlagApp:                "deployed",
			mockFlagTeam:               team1TeamID,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowLocalOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedError:              slackerror.New(slackerror.ErrDeployedAppNotSupported),
		},
		"errors if local app environment flag and team id flag for hosted app prompt": {
			mockFlagApp:                "local",
			mockFlagTeam:               team1TeamID,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedError:              slackerror.New(slackerror.ErrLocalAppNotSupported),
		},
		"errors if team id flag does not have authorization": {
			mockFlagTeam:               team1TeamID,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedError:              slackerror.New(slackerror.ErrTeamNotFound),
		},
		"returns selection for token flag and app id flag of an unsaved local app": {
			mockAuthWithToken:     fakeAuthsByTeamDomain[team2TeamDomain],
			mockFlagApp:           localTeam2InstalledAppID,
			mockFlagToken:         fakeAuthsByTeamDomain[team2TeamDomain].Token,
			appPromptConfigStatus: ShowInstalledAppsOnly,
			// ShowLocalOnly checks that unsaved apps can be selected with flags
			appPromptConfigEnvironment: ShowLocalOnly,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:      localTeam2InstalledAppID,
					TeamDomain: team2TeamDomain,
					TeamID:     team2TeamID,
					// IsDev is not known if the apps.dev.json file doesn't exist
					IsDev:         false,
					UserID:        team2UserID,
					InstallStatus: types.AppStatusInstalled,
				},
				Auth: fakeAuthsByTeamDomain[team2TeamDomain],
			},
			expectedError: slackerror.New(slackerror.ErrDeployedAppNotSupported),
		},
		"returns selection for token flag and app id flag of an unsaved hosted app": {
			mockAuthWithToken:          fakeAuthsByTeamDomain[team1TeamDomain],
			mockFlagApp:                deployedTeam1InstalledAppID,
			mockFlagToken:              fakeAuthsByTeamDomain[team1TeamDomain].Token,
			appPromptConfigStatus:      ShowAllApps,
			appPromptConfigEnvironment: ShowAllEnvironments,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         deployedTeam1InstalledAppID,
					TeamDomain:    team1TeamDomain,
					TeamID:        team1TeamID,
					UserID:        team1UserID,
					InstallStatus: types.AppStatusInstalled,
				},
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns new application with token flag and team id flag if app not saved": {
			mockAuthWithToken:          fakeAuthsByTeamDomain[team1TeamDomain],
			mockAuthWithTeamIDError:    slackerror.New(slackerror.ErrCredentialsNotFound),
			mockAuthWithTeamIDTeamID:   team1TeamID,
			mockFlagTeam:               team1TeamID,
			mockFlagToken:              fakeAuthsByTeamDomain[team1TeamDomain].Token,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			appPromptConfigEnvironment: ShowHostedOnly,
			expectedSelection: SelectedApp{
				App:  types.NewApp(),
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns selection for token flag and team id flag if app saved": {
			mockAppsLocal: []types.App{
				localTeam1UninstalledApp,
			},
			mockAuthWithToken:          fakeAuthsByTeamDomain[team1TeamDomain],
			mockAuthWithTeamIDError:    slackerror.New(slackerror.ErrCredentialsNotFound),
			mockAuthWithTeamIDTeamID:   team1TeamID,
			mockFlagTeam:               team1TeamID,
			mockFlagToken:              fakeAuthsByTeamDomain[team1TeamDomain].Token,
			mockManifestSource:         config.ManifestSourceLocal,
			appPromptConfigStatus:      ShowAllApps,
			appPromptConfigEnvironment: ShowLocalOnly,
			expectedSelection: SelectedApp{
				App:  localTeam1UninstalledApp,
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"errors if token flag and team id flag do not match": {
			mockFlagTeam:             team1TeamID,
			mockFlagToken:            fakeAuthsByTeamDomain[team2TeamDomain].Token,
			mockAuthWithToken:        fakeAuthsByTeamDomain[team2TeamDomain],
			mockAuthWithTeamIDError:  slackerror.New(slackerror.ErrCredentialsNotFound),
			mockAuthWithTeamIDTeamID: team2TeamID,
			expectedError:            slackerror.New(slackerror.ErrTeamNotFound),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(
				Auths,
				mock.Anything,
			).Return(
				tt.mockAuths,
				nil,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				tt.mockAuthWithTeamIDTeamID,
			).Return(
				types.SlackAuth{},
				tt.mockAuthWithTeamIDError,
			)
			clientsMock.Auth.On(
				AuthWithToken,
				mock.Anything,
				tt.mockFlagToken,
			).Return(
				tt.mockAuthWithToken,
				nil,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				mock.Anything,
			).Return(
				api.AuthSession{},
				nil,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				[]string{
					deployedTeam1InstalledAppID,
					localTeam1UninstalledAppID,
				},
				mock.Anything,
			).Return(
				api.GetAppStatusResult{
					Apps: []api.AppStatusResultAppInfo{
						{
							AppID:     deployedTeam1InstalledAppID,
							Installed: deployedTeam1AppIsInstalled,
						},
						{
							AppID:     localTeam1UninstalledAppID,
							Installed: localTeam1AppIsInstalled,
						},
					},
				},
				nil,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				[]string{
					deployedTeam1InstalledAppID,
				},
				mock.Anything,
			).Return(
				api.GetAppStatusResult{
					Apps: []api.AppStatusResultAppInfo{
						{
							AppID:     deployedTeam1InstalledAppID,
							Installed: deployedTeam1AppIsInstalled,
						},
					},
				},
				nil,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				[]string{
					localTeam1UninstalledAppID,
				},
				mock.Anything,
			).Return(
				api.GetAppStatusResult{
					Apps: []api.AppStatusResultAppInfo{
						{
							AppID:     localTeam1UninstalledAppID,
							Installed: localTeam1AppIsInstalled,
						},
					},
				},
				nil,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				[]string{
					deployedTeam2UninstalledAppID,
					localTeam2InstalledAppID,
				},
				mock.Anything,
			).Return(
				api.GetAppStatusResult{
					Apps: []api.AppStatusResultAppInfo{
						{
							AppID:     deployedTeam2UninstalledAppID,
							Installed: deployedTeam2AppIsInstalled,
						},
						{
							AppID:     localTeam2InstalledAppID,
							Installed: localTeam2AppIsInstalled,
						},
					},
				},
				nil,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				[]string{
					deployedTeam2UninstalledAppID,
				},
				mock.Anything,
			).Return(
				api.GetAppStatusResult{
					Apps: []api.AppStatusResultAppInfo{
						{
							AppID:     deployedTeam2UninstalledAppID,
							Installed: deployedTeam2AppIsInstalled,
						},
					},
				},
				nil,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				[]string{
					localTeam2InstalledAppID,
				},
				mock.Anything,
			).Return(
				api.GetAppStatusResult{
					Apps: []api.AppStatusResultAppInfo{
						{
							AppID:     localTeam2InstalledAppID,
							Installed: localTeam2AppIsInstalled,
						},
					},
				},
				nil,
			)
			clientsMock.IO.On(
				SelectPrompt,
				mock.Anything,
				"Choose a team",
				mock.Anything,
				iostreams.MatchPromptConfig(
					iostreams.SelectPromptConfig{
						Flag:     clientsMock.Config.Flags.Lookup("team"),
						Required: true,
					},
				),
			).Return(
				iostreams.SelectPromptResponse{
					Flag:   tt.teamPromptResponseFlag,
					Prompt: tt.teamPromptResponsePrompt,
					Option: tt.teamPromptResponseOption,
					Index:  tt.teamPromptResponseIndex,
				},
				nil,
			)
			clientsMock.IO.On(
				SelectPrompt,
				mock.Anything,
				"Select an app",
				tt.appPromptConfigOptions,
				iostreams.MatchPromptConfig(
					iostreams.SelectPromptConfig{
						Required: true,
					},
				),
			).Return(
				iostreams.SelectPromptResponse{
					Flag:   tt.appPromptResponseFlag,
					Prompt: tt.appPromptResponsePrompt,
					Option: tt.appPromptResponseOption,
					Index:  tt.appPromptResponseIndex,
				},
				nil,
			)
			clientsMock.AddDefaultMocks()
			projectConfigMock := config.NewProjectConfigMock()
			projectConfigMock.On(
				"GetManifestSource",
				mock.Anything,
			).Return(
				tt.mockManifestSource,
				nil,
			)
			clientsMock.Config.AppFlag = tt.mockFlagApp
			clientsMock.Config.ProjectConfig = projectConfigMock
			clientsMock.Config.TeamFlag = tt.mockFlagTeam
			clientsMock.Config.TokenFlag = tt.mockFlagToken
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range tt.mockAppsDeployed {
				err := clients.AppClient().SaveDeployed(ctx, app)
				require.NoError(t, err)
			}
			for _, app := range tt.mockAppsLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			selectedApp, err := flatAppSelectPrompt(ctx, clients, tt.appPromptConfigEnvironment, tt.appPromptConfigStatus)
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedSelection, selectedApp)
			require.Contains(t, clientsMock.GetStdoutOutput(), tt.expectedStdout)
			require.Contains(t, clientsMock.GetStderrOutput(), tt.expectedStderr)
		})
	}
}

//
// TeamAppSelectPrompt tests
//

func TestPrompt_TeamAppSelectPrompt_TokenAppFlag(t *testing.T) {
	tests := map[string]struct {
		tokenFlag    string
		tokenAuth    types.SlackAuth
		appFlag      string
		appStatus    api.GetAppStatusResult
		statusErr    error
		saveLocal    []types.App
		selectEnv    AppEnvironmentType
		selectStatus AppInstallStatus
		expectedApp  SelectedApp
		expectedErr  error
	}{
		"error if an error occurred while collecting app info": {
			tokenFlag:    team1Token,
			tokenAuth:    fakeAuthsByTeamDomain[team1TeamDomain],
			appFlag:      localTeam1UninstalledApp.AppID,
			appStatus:    api.GetAppStatusResult{},
			statusErr:    slackerror.New(slackerror.ErrAppNotFound),
			selectEnv:    ShowHostedOnly,
			selectStatus: ShowAllApps,
			expectedApp:  SelectedApp{},
			expectedErr:  slackerror.New(slackerror.ErrAppNotFound),
		},
		"continue if a saved local app is used for a deployed only prompt": {
			tokenFlag: team1Token,
			tokenAuth: fakeAuthsByTeamDomain[team1TeamDomain],
			appFlag:   localTeam1UninstalledApp.AppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     localTeam1UninstalledAppID,
					Installed: localTeam1AppIsInstalled,
					Hosted:    false,
				}},
			},
			statusErr:    nil,
			saveLocal:    []types.App{localTeam1UninstalledApp},
			selectEnv:    ShowHostedOnly,
			selectStatus: ShowAllApps,
			expectedApp: SelectedApp{
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				App:  localTeam1UninstalledApp,
			},
			expectedErr: slackerror.New(slackerror.ErrLocalAppNotSupported),
		},
		"error if a deployed app is used for a local only prompt": {
			tokenFlag: team2Token,
			tokenAuth: fakeAuthsByTeamDomain[team2TeamDomain],
			appFlag:   deployedTeam2UninstalledApp.AppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam2UninstalledApp.AppID,
					Installed: deployedTeam2AppIsInstalled,
					Hosted:    true,
				}},
			},
			statusErr:    nil,
			selectEnv:    ShowLocalOnly,
			selectStatus: ShowAllApps,
			expectedApp:  SelectedApp{},
			expectedErr:  slackerror.New(slackerror.ErrDeployedAppNotSupported),
		},
		"error if an uninstalled app is used for an installed only prompt": {
			tokenFlag: team2Token,
			tokenAuth: fakeAuthsByTeamDomain[team2TeamDomain],
			appFlag:   deployedTeam2UninstalledApp.AppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam2UninstalledApp.AppID,
					Installed: deployedTeam2AppIsInstalled,
					Hosted:    true,
				}},
			},
			statusErr:    nil,
			selectEnv:    ShowHostedOnly,
			selectStatus: ShowInstalledAppsOnly,
			expectedApp:  SelectedApp{},
			expectedErr:  slackerror.New(slackerror.ErrInstallationRequired),
		},
		"returns known information about the request app": {
			tokenFlag: team1Token,
			tokenAuth: fakeAuthsByTeamDomain[team1TeamDomain],
			appFlag:   deployedTeam1InstalledAppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam1InstalledAppID,
					Installed: deployedTeam1AppIsInstalled,
					Hosted:    true,
				}},
			},
			statusErr:    nil,
			selectEnv:    ShowHostedOnly,
			selectStatus: ShowInstalledAppsOnly,
			expectedApp: SelectedApp{
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				App:  deployedTeam1InstalledApp,
			},
			expectedErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(AuthWithToken, mock.Anything, test.tokenFlag).
				Return(test.tokenAuth, nil)
			clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(test.appStatus, test.statusErr)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range test.saveLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			clients.Config.TokenFlag = test.tokenFlag
			clients.Config.AppFlag = test.appFlag

			selection, err := TeamAppSelectPrompt(ctx, clients, test.selectEnv, test.selectStatus)

			if test.statusErr != nil && assert.Error(t, err) {
				require.Equal(t, test.statusErr, err)
			} else if test.expectedErr != nil && assert.Error(t, err) {
				require.Equal(t, test.expectedErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedApp.Auth, selection.Auth)
				expectedApp := test.expectedApp.App
				expectedApp.UserID = test.expectedApp.Auth.UserID
				assert.Equal(t, expectedApp, selection.App)
			}
		})
	}
}

func TestSortAlphaNumeric_Sorted(t *testing.T) {
	items := []string{"alphabetical", "bordering"}
	labels := []string{"_alphabetical_ T001", "_bordering_ T1"}
	teamIDs := []string{"T001", "T1"}
	err := SortAlphaNumeric(items, labels, teamIDs)
	require.NoError(t, err)

	require.Equal(t, items[0], "alphabetical")
	require.Equal(t, labels[0], "_alphabetical_ T001")
	require.Equal(t, teamIDs[0], "T001")

	require.Equal(t, items[1], "bordering")
	require.Equal(t, labels[1], "_bordering_ T1")
	require.Equal(t, teamIDs[1], "T1")
}

func TestSortAlphaNumeric_Unsorted(t *testing.T) {
	items := []string{"bordering", "alphabetical"}
	labels := []string{"_bordering_ T1", "_alphabetical_ T001"}
	teamIDs := []string{"T1", "T001"}
	err := SortAlphaNumeric(items, labels, teamIDs)
	require.NoError(t, err)

	require.Equal(t, "alphabetical", items[0])
	require.Equal(t, "_alphabetical_ T001", labels[0])
	require.Equal(t, "T001", teamIDs[0])

	require.Equal(t, "bordering", items[1])
	require.Equal(t, "_bordering_ T1", labels[1])
	require.Equal(t, "T1", teamIDs[1])
}

func TestSortAlphaNumeric_Unbalanced(t *testing.T) {
	items := []string{"bordering", "alphabetical"}
	labels := []string{"_bordering_ T1"} // oops
	teamIDs := []string{"T1"}            // also oops
	err := SortAlphaNumeric(items, labels, teamIDs)
	expected := slackerror.New(slackerror.ErrTeamList)

	require.Equal(t, err, expected)
}

func Test_ValidateGetOrgWorkspaceGrant(t *testing.T) {
	orgGrant := "T123"

	tests := map[string]struct {
		app                  *SelectedApp
		mockPrompt           func(clientsMock *shared.ClientsMock)
		inputGrant           string
		expectedGrant        string
		expectedErr          error
		firstPromptOptionAll bool
	}{
		"Workspace grant can be used for new org apps": {
			app: &SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			inputGrant:    orgGrant,
			expectedGrant: orgGrant,
		},
		"Workspace grant can be used for uninstalled org apps": {
			app: &SelectedApp{
				App:  types.App{InstallStatus: types.AppStatusUninstalled, TeamID: "E123"},
				Auth: types.SlackAuth{},
			},
			inputGrant:    orgGrant,
			expectedGrant: orgGrant,
		},
		"Workspace grant should not be used for non-org apps": {
			app: &SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: false},
			},
			inputGrant:    orgGrant,
			expectedGrant: "",
		},
		"Workspace grant 'all' not overwritten": {
			app: &SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			inputGrant:    "all",
			expectedGrant: "all",
		},
		"Workspace grant can be used with an uninstalled org app if the grant is equal to the app's current grant": {
			app: &SelectedApp{
				App: types.App{
					InstallStatus: types.AppStatusUninstalled,
					TeamID:        "E123",
					EnterpriseGrants: []types.EnterpriseGrant{
						{WorkspaceID: "T1", WorkspaceDomain: "workspace1"}}},
				Auth: types.SlackAuth{},
			},
			inputGrant:    "T1",
			expectedGrant: "T1",
		},
		"Workspace grant cannot be used with installed org app if different from app's current grants": {
			app: &SelectedApp{
				App: types.App{
					AppID:         "A123",
					InstallStatus: types.AppStatusInstalled,
					TeamID:        "E123",
					EnterpriseGrants: []types.EnterpriseGrant{
						{WorkspaceID: "T1", WorkspaceDomain: "workspace1"}}},
				Auth: types.SlackAuth{},
			},
			inputGrant:    orgGrant,
			expectedGrant: "",
			expectedErr: slackerror.New(slackerror.ErrOrgGrantExists).
				WithMessage("A different org workspace grant already exists for installed app 'A123'\n   Workspace Grant: T1"),
		},
		"Prompt user; 'all workspaces' is last option": {
			app: &SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			mockPrompt: func(clientsMock *shared.ClientsMock) {
				clientsMock.API.On("AuthTeamsList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.TeamInfo{
						{ID: "T1", Name: "team1"},
						{ID: "T2", Name: "team2"},
						{ID: "T3", Name: "team3"},
					},
					"",
					nil)
				clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a workspace to grant access:", mock.Anything, mock.Anything).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "team2 T2",
						Index:  1,
					}, nil)
			},
			inputGrant:           "",
			firstPromptOptionAll: false,
			expectedGrant:        "T2",
		},
		"Prompt user; 'all workspaces' is first option": {
			app: &SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			mockPrompt: func(clientsMock *shared.ClientsMock) {
				clientsMock.API.On("AuthTeamsList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.TeamInfo{
						{ID: "T1", Name: "team1"},
						{ID: "T2", Name: "team2"},
						{ID: "T3", Name: "team3"},
					},
					"",
					nil)
				clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a workspace to grant access:", mock.Anything, mock.Anything).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "team2 T2",
						Index:  2,
					}, nil)
			},
			inputGrant:           "",
			firstPromptOptionAll: true,
			expectedGrant:        "T2",
		},
		"Prompt user; select 'all workspaces'": {
			app: &SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			mockPrompt: func(clientsMock *shared.ClientsMock) {
				clientsMock.API.On("AuthTeamsList", mock.Anything, mock.Anything, mock.Anything).Return(
					[]types.TeamInfo{
						{ID: "T1", Name: "team1"},
						{ID: "T2", Name: "team2"},
						{ID: "T3", Name: "team3"},
					},
					"",
					nil)
				clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a workspace to grant access:", mock.Anything, mock.Anything).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "All of them",
						Index:  0,
					}, nil)
			},
			inputGrant:           "",
			firstPromptOptionAll: true,
			expectedGrant:        "all",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
			})
			if tt.mockPrompt != nil {
				tt.mockPrompt(clientsMock)
			}
			returnedGrant, err := ValidateGetOrgWorkspaceGrant(ctx, clients, tt.app, tt.inputGrant, tt.firstPromptOptionAll)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedGrant, returnedGrant)
		})
	}
}

// Test_ValidateAuth tests edge cases of the reauthentication logic for certain
// errors
//
// Successful reauthentication in other functions might use default mocks below
func Test_ValidateAuth(t *testing.T) {
	apiHostDev := "https://dev.slack.com"
	tests := map[string]struct {
		authProvided                        types.SlackAuth
		authExpected                        types.SlackAuth
		expectedErr                         error
		apiExchangeAuthTicketResultResponse api.ExchangeAuthTicketResult
		apiExchangeAuthTicketResultError    error
		authFilteredKnownAuthErrorsResponse bool
		authFilteredKnownAuthErrorsError    error
		apiGenerateAuthTicketResultResponse api.GenerateAuthTicketResult
		apiGenerateAuthTicketResultError    error
		apiValidateSessionResponse          api.AuthSession
		apiValidateSessionError             error
		authIsAPIHostSlackProdResponse      bool
		authSetAuthResponse                 types.SlackAuth
		authSetAuthError                    error
		ioIsTTYResponse                     bool
	}{
		"returns a valid authentication without changes": {
			authProvided: types.SlackAuth{
				Token: "xoxb-original",
			},
			authExpected: types.SlackAuth{
				Token: "xoxb-original",
			},
		},
		"revalidates an expired authentication on a dev instance": {
			authProvided: types.SlackAuth{
				APIHost: &apiHostDev,
				Token:   "xoxb-development",
			},
			authExpected: types.SlackAuth{
				APIHost:    &apiHostDev,
				TeamDomain: team1TeamDomain,
				TeamID:     team1TeamID,
				Token:      fakeAuthsByTeamDomain[team1TeamDomain].Token,
				UserID:     "U1",
			},
			apiExchangeAuthTicketResultResponse: api.ExchangeAuthTicketResult{
				TeamDomain: team1TeamDomain,
				TeamID:     team1TeamID,
				Token:      fakeAuthsByTeamDomain[team1TeamDomain].Token,
				UserID:     "U1",
			},
			apiValidateSessionError:             slackerror.New(slackerror.ErrInvalidAuth),
			authFilteredKnownAuthErrorsResponse: true,
			authFilteredKnownAuthErrorsError:    nil,
			authIsAPIHostSlackProdResponse:      false,
			ioIsTTYResponse:                     true,
		},
		"returns unexpected errors from validate session": {
			authProvided: types.SlackAuth{
				Token: "xoxb-testing",
			},
			authExpected: types.SlackAuth{
				Token: "xoxb-testing",
			},
			expectedErr:                         slackerror.New(slackerror.ErrHTTPRequestFailed),
			apiValidateSessionError:             slackerror.New(slackerror.ErrHTTPRequestFailed),
			authFilteredKnownAuthErrorsResponse: false,
			authFilteredKnownAuthErrorsError:    slackerror.New(slackerror.ErrHTTPRequestFailed),
		},
		"errors without revalidation if the terminal is not interactive": {
			authProvided: types.SlackAuth{
				Token: "xoxb-abcdefghijkl",
			},
			authExpected: types.SlackAuth{
				Token: "xoxb-abcdefghijkl",
			},
			expectedErr:                         slackerror.New(slackerror.ErrAlreadyLoggedOut),
			apiValidateSessionError:             slackerror.New(slackerror.ErrAlreadyLoggedOut),
			authFilteredKnownAuthErrorsResponse: true,
			authFilteredKnownAuthErrorsError:    nil,
			ioIsTTYResponse:                     false,
		},
		"errors if the ticket generation returns errors for a reason": {
			authProvided: types.SlackAuth{
				Token: "xoxb-expired",
			},
			authExpected: types.SlackAuth{
				Token: "xoxb-expired",
			},
			expectedErr:                         slackerror.New(slackerror.ErrInvalidChallenge),
			apiGenerateAuthTicketResultError:    slackerror.New(slackerror.ErrInvalidChallenge),
			apiValidateSessionError:             slackerror.New(slackerror.ErrTokenExpired),
			authFilteredKnownAuthErrorsResponse: true,
			authFilteredKnownAuthErrorsError:    nil,
			ioIsTTYResponse:                     true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.API.On(
				"ExchangeAuthTicket",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.apiExchangeAuthTicketResultResponse,
				tt.apiExchangeAuthTicketResultError,
			)
			clientsMock.API.On(
				"GenerateAuthTicket",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.apiGenerateAuthTicketResultResponse,
				tt.apiGenerateAuthTicketResultError,
			)
			if tt.authProvided.APIHost != nil {
				clientsMock.API.On(
					"Host",
				).Return(
					*tt.authProvided.APIHost,
				)
			}
			clientsMock.API.On(
				"SetHost",
				mock.Anything,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				tt.authProvided.Token,
			).Return(
				tt.apiValidateSessionResponse,
				tt.apiValidateSessionError,
			)
			clientsMock.Auth.On(
				"FilterKnownAuthErrors",
				mock.Anything,
				tt.apiValidateSessionError,
			).Return(
				tt.authFilteredKnownAuthErrorsResponse,
				tt.authFilteredKnownAuthErrorsError,
			)
			clientsMock.Auth.On(
				"IsAPIHostSlackProd",
				mock.Anything,
			).Return(
				tt.authIsAPIHostSlackProdResponse,
			)
			clientsMock.Auth.On(
				"SetAuth",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.authSetAuthResponse,
				"",
				tt.authSetAuthError,
			)
			clientsMock.Auth.On(
				"SetSelectedAuth",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return()
			clientsMock.IO.On(
				"InputPrompt",
				mock.Anything,
				"Enter challenge code",
				iostreams.InputPromptConfig{Required: true},
			).Return(
				"challengeCode",
				nil,
			)
			clientsMock.IO.On(
				"IsTTY",
			).Return(
				tt.ioIsTTYResponse,
			)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			err := validateAuth(ctx, clients, &tt.authProvided)

			tt.authProvided.LastUpdated = time.Time{} // ignore time for this test
			assert.Equal(t, tt.expectedErr, err)
			if tt.authExpected.APIHost != nil {
				clientsMock.API.AssertCalled(t, "SetHost", *tt.authExpected.APIHost)
			}
			assert.Equal(t, tt.authExpected, tt.authProvided)
		})
	}
}
