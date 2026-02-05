// Copyright 2022-2026 Salesforce, Inc.
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

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(AuthWithToken, mock.Anything, tc.tokenFlag).
				Return(tc.tokenAuth, tc.tokenErr)
			clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(tc.appStatus, tc.statusErr)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range tc.saveLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			selection, err := getTokenApp(ctx, clients, tc.tokenFlag, tc.appFlag)

			if tc.tokenErr != nil && assert.Error(t, err) {
				require.Equal(t, tc.tokenErr, err)
			} else if tc.statusErr != nil && assert.Error(t, err) {
				require.Equal(t, tc.statusErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.tokenAuth, selection.Auth)
				expectedApp := tc.appInfo
				expectedApp.UserID = tc.tokenAuth.UserID
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
		tokenFlag         string
		tokenAuth         types.SlackAuth
		appFlag           string
		appStatus         api.GetAppStatusResult
		appStatusErr      error
		selectEnvironment AppEnvironmentType
		selectStatus      AppInstallStatus
		expectedApp       SelectedApp
		expectedErr       error
	}{
		"error if an error occurred while collecting app info": {
			tokenFlag:    team1Token,
			tokenAuth:    fakeAuthsByTeamDomain[team1TeamDomain],
			appFlag:      localTeam1UninstalledApp.AppID,
			appStatus:    api.GetAppStatusResult{},
			appStatusErr: slackerror.New(slackerror.ErrAppNotFound),
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
			appStatusErr: nil,
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
					Hosted:    false,
				}},
			},
			appStatusErr: nil,
			selectStatus: ShowInstalledAppsOnly,
			expectedApp: SelectedApp{
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				App:  deployedTeam1InstalledApp,
			},
			expectedErr: nil,
		},
		"returns app details without respect to the app environment": {
			tokenFlag: team2Token,
			tokenAuth: fakeAuthsByTeamDomain[team2TeamDomain],
			appFlag:   deployedTeam2UninstalledAppID,
			appStatus: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{{
					AppID:     deployedTeam2UninstalledAppID,
					Installed: deployedTeam2AppIsInstalled,
					Hosted:    true,
				}},
			},
			appStatusErr:      nil,
			selectEnvironment: ShowLocalOnly,
			selectStatus:      ShowInstalledAndUninstalledApps,
			expectedApp: SelectedApp{
				Auth: fakeAuthsByTeamDomain[team2TeamDomain],
				App:  deployedTeam2UninstalledApp,
			},
			expectedErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(AuthWithToken, mock.Anything, tc.tokenFlag).
				Return(tc.tokenAuth, nil)
			clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(tc.appStatus, tc.appStatusErr)
			clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
				TeamName: &tc.tokenAuth.TeamDomain,
				TeamID:   &tc.tokenAuth.TeamID,
			}, nil)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			clients.Config.TokenFlag = tc.tokenFlag
			clients.Config.AppFlag = tc.appFlag

			selection, err := AppSelectPrompt(ctx, clients, ShowAllEnvironments, tc.selectStatus)

			if tc.appStatusErr != nil && assert.Error(t, err) {
				require.Equal(t, tc.appStatusErr, err)
			} else if tc.expectedErr != nil && assert.Error(t, err) {
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedApp.Auth, selection.Auth)
				expectedApp := tc.expectedApp.App
				expectedApp.UserID = tc.expectedApp.Auth.UserID
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
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				tc.mockTeam1StatusAppIDs,
				mock.Anything,
			).Return(
				tc.mockTeam1Status,
				tc.mockTeam1StatusError,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				tc.mockTeam2StatusAppIDs,
				mock.Anything,
			).Return(
				tc.mockTeam2Status,
				tc.mockTeam2StatusError,
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
				tc.mockAuths,
				nil,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				team2TeamID,
			).Return(
				tc.mockTeam2SavedAuth,
				tc.mockTeam2SavedAuthError,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				team1TeamID,
			).Return(
				tc.mockTeam1SavedAuth,
				tc.mockTeam1SavedAuthError,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				enterprise1TeamID,
			).Return(
				tc.mockEnterprise1SavedAuth,
				tc.mockEnterprise1SavedAuthError,
			)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range tc.mockAppsSavedDeployed {
				err := clients.AppClient().SaveDeployed(ctx, app)
				require.NoError(t, err)
			}
			for _, app := range tc.mockAppsSavedLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			apps, err := getApps(ctx, clients)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedApps, apps)
		})
	}
}

func TestPrompt_AppSelectPrompt(t *testing.T) {
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
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedSelection: SelectedApp{
				App:  types.NewApp(),
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns existing application for app environment flag and team id flag if app saved": {
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
		"returns installed application for app environment flag and team id flag if app saved": {
			mockAuths: fakeAuthsByTeamDomainSlice,
			mockAppsLocal: []types.App{
				localTeam2InstalledApp,
			},
			mockFlagApp:                "local",
			mockFlagTeam:               team2TeamID,
			appPromptConfigEnvironment: ShowAllEnvironments,
			appPromptConfigStatus:      ShowInstalledAppsOnly,
			expectedSelection: SelectedApp{
				App: types.App{
					AppID:         localTeam2InstalledAppID,
					TeamDomain:    team2TeamDomain,
					TeamID:        team2TeamID,
					UserID:        team2UserID,
					IsDev:         true,
					InstallStatus: types.AppStatusInstalled,
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
			appPromptConfigStatus:   ShowInstalledAndUninstalledApps,
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
			appPromptConfigStatus:      ShowAllApps,
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
			appPromptConfigStatus:      ShowAllApps,
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
		"errors if app id flag is not valid": {
			mockFlagApp:   "123",
			expectedError: slackerror.New(slackerror.ErrInvalidAppFlag),
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
			appPromptConfigStatus:      ShowAllApps,
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
			expectedError:              slackerror.New(slackerror.ErrTeamNotFound),
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
			appPromptConfigEnvironment: ShowLocalOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedError:              slackerror.New(slackerror.ErrDeployedAppNotSupported),
		},
		"errors if local app environment flag and team id flag for hosted app prompt": {
			mockFlagApp:                "local",
			mockFlagTeam:               team1TeamID,
			appPromptConfigEnvironment: ShowHostedOnly,
			appPromptConfigStatus:      ShowInstalledAndNewApps,
			expectedError:              slackerror.New(slackerror.ErrLocalAppNotSupported),
		},
		"errors if team id flag does not have authorization": {
			mockFlagTeam:               team1TeamID,
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
			appPromptConfigStatus:      ShowAllApps,
			appPromptConfigEnvironment: ShowLocalOnly,
			expectedSelection: SelectedApp{
				App:  localTeam1UninstalledApp,
				Auth: fakeAuthsByTeamDomain[team1TeamDomain],
			},
		},
		"returns selection for token flag if one app saved": {
			mockAppsLocal: []types.App{
				localTeam1UninstalledApp,
				localTeam2InstalledApp,
			},
			mockAuthWithToken:        fakeAuthsByTeamDomain[team2TeamDomain],
			mockAuthWithTeamIDError:  slackerror.New(slackerror.ErrCredentialsNotFound),
			mockAuthWithTeamIDTeamID: mock.Anything,
			mockFlagToken:            fakeAuthsByTeamDomain[team2TeamDomain].Token,
			expectedSelection: SelectedApp{
				App:  localTeam2InstalledApp,
				Auth: fakeAuthsByTeamDomain[team2TeamDomain],
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
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(
				Auths,
				mock.Anything,
			).Return(
				tc.mockAuths,
				nil,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				tc.mockAuthWithTeamIDTeamID,
			).Return(
				types.SlackAuth{},
				tc.mockAuthWithTeamIDError,
			)
			clientsMock.Auth.On(
				AuthWithToken,
				mock.Anything,
				tc.mockFlagToken,
			).Return(
				tc.mockAuthWithToken,
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
					Flag:   tc.teamPromptResponseFlag,
					Prompt: tc.teamPromptResponsePrompt,
					Option: tc.teamPromptResponseOption,
					Index:  tc.teamPromptResponseIndex,
				},
				nil,
			)
			clientsMock.IO.On(
				SelectPrompt,
				mock.Anything,
				"Select an app",
				tc.appPromptConfigOptions,
				iostreams.MatchPromptConfig(
					iostreams.SelectPromptConfig{
						Required: true,
					},
				),
			).Return(
				iostreams.SelectPromptResponse{
					Flag:   tc.appPromptResponseFlag,
					Prompt: tc.appPromptResponsePrompt,
					Option: tc.appPromptResponseOption,
					Index:  tc.appPromptResponseIndex,
				},
				nil,
			)
			clientsMock.AddDefaultMocks()
			clientsMock.Config.AppFlag = tc.mockFlagApp
			clientsMock.Config.TeamFlag = tc.mockFlagTeam
			clientsMock.Config.TokenFlag = tc.mockFlagToken
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range tc.mockAppsDeployed {
				err := clients.AppClient().SaveDeployed(ctx, app)
				require.NoError(t, err)
			}
			for _, app := range tc.mockAppsLocal {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			selectedApp, err := AppSelectPrompt(ctx, clients, tc.appPromptConfigEnvironment, tc.appPromptConfigStatus)
			require.Equal(t, tc.expectedError, err)
			require.Equal(t, tc.expectedSelection, selectedApp)
			require.Contains(t, clientsMock.GetStdoutOutput(), tc.expectedStdout)
			require.Contains(t, clientsMock.GetStderrOutput(), tc.expectedStderr)
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
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
			})
			if tc.mockPrompt != nil {
				tc.mockPrompt(clientsMock)
			}
			returnedGrant, err := ValidateGetOrgWorkspaceGrant(ctx, clients, tc.app, tc.inputGrant, tc.firstPromptOptionAll)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedGrant, returnedGrant)
		})
	}
}

// Test_ValidateAuth tests edge cases of the reauthentication logic for certain
// errors
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
	for name, tc := range tests {
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
				tc.apiExchangeAuthTicketResultResponse,
				tc.apiExchangeAuthTicketResultError,
			)
			clientsMock.API.On(
				"GenerateAuthTicket",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tc.apiGenerateAuthTicketResultResponse,
				tc.apiGenerateAuthTicketResultError,
			)
			if tc.authProvided.APIHost != nil {
				clientsMock.API.On(
					"Host",
				).Return(
					*tc.authProvided.APIHost,
				)
			}
			clientsMock.API.On(
				"SetHost",
				mock.Anything,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				tc.authProvided.Token,
			).Return(
				tc.apiValidateSessionResponse,
				tc.apiValidateSessionError,
			)
			clientsMock.Auth.On(
				"FilterKnownAuthErrors",
				mock.Anything,
				tc.apiValidateSessionError,
			).Return(
				tc.authFilteredKnownAuthErrorsResponse,
				tc.authFilteredKnownAuthErrorsError,
			)
			clientsMock.Auth.On(
				"IsAPIHostSlackProd",
				mock.Anything,
			).Return(
				tc.authIsAPIHostSlackProdResponse,
			)
			clientsMock.Auth.On(
				"SetAuth",
				mock.Anything,
				mock.Anything,
			).Return(
				tc.authSetAuthResponse,
				"",
				tc.authSetAuthError,
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
				tc.ioIsTTYResponse,
			)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			err := validateAuth(ctx, clients, &tc.authProvided)

			tc.authProvided.LastUpdated = time.Time{} // ignore time for this tc
			assert.Equal(t, tc.expectedErr, err)
			if tc.authExpected.APIHost != nil {
				clientsMock.API.AssertCalled(t, "SetHost", *tc.authExpected.APIHost)
			}
			assert.Equal(t, tc.authExpected, tc.authProvided)
		})
	}
}
