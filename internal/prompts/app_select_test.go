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
	"context"
	"errors"
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
// getTeamApps tests
//

func TestGetTeamApps(t *testing.T) {
	tests := map[string]struct {
		tokenFlag                   string
		deployedApps                []types.App
		localApps                   []types.App
		mockAuthWithTokenResponse   types.SlackAuth
		mockAuthWithTokenError      error
		mockAuthWithTeamIDResponse  types.SlackAuth
		mockAuthWithTeamIDError     error
		mockGetAppStatusResponse    api.GetAppStatusResult
		mockGetAppStatusError       error
		mockValidateSessionResponse api.AuthSession
		mockValidateSessionError    error
		expectedAuth                types.SlackAuth
		expectedTeamID              string
		expectedError               error
	}{
		"error when the token authentication returns an error": {
			tokenFlag:              team1Token,
			mockAuthWithTokenError: slackerror.New(slackerror.ErrTokenExpired),
			expectedAuth:           types.SlackAuth{},
			expectedError:          slackerror.New(slackerror.ErrTokenExpired),
		},
		"retrieve apps and authentication associated with a token": {
			tokenFlag:                 team1Token,
			deployedApps:              []types.App{deployedTeam1InstalledApp},
			localApps:                 []types.App{localTeam1UninstalledApp},
			mockAuthWithTokenResponse: fakeAuthsByTeamDomain[team1TeamDomain],
			mockAuthWithTeamIDError:   slackerror.New(slackerror.ErrCredentialsNotFound),
			mockGetAppStatusResponse: api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					{
						AppID:     deployedTeam1InstalledAppID,
						Installed: true,
					},
					{
						AppID:     localTeam1UninstalledAppID,
						Installed: false,
					},
				},
			},
			expectedTeamID: team1TeamID,
			expectedAuth:   fakeAuthsByTeamDomain[team1TeamDomain],
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(
				AuthWithToken,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockAuthWithTokenResponse,
				tt.mockAuthWithTokenError,
			)
			clientsMock.Auth.On(
				AuthWithTeamID,
				mock.Anything,
				tt.mockAuthWithTokenResponse.TeamID,
			).Return(
				tt.mockAuthWithTeamIDResponse,
				tt.mockAuthWithTeamIDError,
			)
			clientsMock.API.On(
				GetAppStatus,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockGetAppStatusResponse,
				tt.mockGetAppStatusError,
			)
			clientsMock.API.On(
				"ValidateSession",
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockValidateSessionResponse,
				tt.mockValidateSessionError,
			)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			for _, app := range tt.deployedApps {
				err := clients.AppClient().SaveDeployed(ctx, app)
				require.NoError(t, err)
			}
			for _, app := range tt.localApps {
				err := clients.AppClient().SaveLocal(ctx, app)
				require.NoError(t, err)
			}
			clients.Config.TokenFlag = tt.tokenFlag

			actual, err := getTeamApps(ctx, clients)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(
					t,
					slackerror.ToSlackError(err).Code,
					slackerror.ToSlackError(tt.expectedError).Code,
				)
			} else {
				require.NoError(t, err)
				if tt.tokenFlag != "" {
					tt.expectedAuth.Token = tt.tokenFlag // expect the same token
				}
				assert.Equal(
					t,
					tt.tokenFlag,
					actual[tt.expectedTeamID].Auth.Token,
				)
			}
		})
	}
}

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
// filterAuthsByToken tests
//

func TestFilterAuthsByToken_NoLogin(t *testing.T) {
	test := struct {
		title        string
		TokenFlag    string
		expectedAuth types.SlackAuth
	}{
		title:        "return the correct mocked api response",
		TokenFlag:    team1Token,
		expectedAuth: fakeAuthsByTeamDomain[team1TeamDomain],
	}
	test.expectedAuth.Token = test.TokenFlag // expect the same token

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(AuthWithToken, mock.Anything, test.TokenFlag).
		Return(test.expectedAuth, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, errors.New(slackerror.ErrCredentialsNotFound))
	clientsMock.Auth.On(SetAuth, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.Config.TokenFlag = test.TokenFlag

	workspaceApps, err := getTeamApps(ctx, clients)
	assert.NoError(t, err)

	auth, err := filterAuthsByToken(ctx, clients, workspaceApps)
	auth.LastUpdated = time.Time{} // ignore time for this test

	assert.NoError(t, err)
	assert.Equal(t, test.expectedAuth, auth)
}

func Test_FilterAuthsByToken_Flags(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())

	mockAuthTeam1 := fakeAuthsByTeamDomain[team1TeamDomain]
	mockAuthTeam1.Token = team1Token
	mockAuthTeam2 := fakeAuthsByTeamDomain[team2TeamDomain]
	mockAuthTeam2.Token = team2Token

	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(AuthWithToken, mock.Anything, team1Token).
		Return(mockAuthTeam1, nil)
	clientsMock.Auth.On(AuthWithToken, mock.Anything, team2Token).
		Return(mockAuthTeam2, nil)

	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(SetAuth, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, errors.New(slackerror.ErrCredentialsNotFound))

	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:     team2TeamID,
		TeamDomain: team2TeamDomain,
		AppID:      "A1EXAMPLE03",
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:     team1TeamID,
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE01",
	})
	require.NoError(t, err)

	tests := map[string]struct {
		AppFlag      string
		TokenFlag    string
		TeamFlag     string
		expectedAuth types.SlackAuth
		expectedErr  slackerror.Error
	}{
		"standalone token to select team": {
			AppFlag:      "",
			TokenFlag:    team1Token,
			TeamFlag:     "",
			expectedAuth: mockAuthTeam1,
			expectedErr:  slackerror.Error{},
		},
		"token with matching workspace flag": {
			AppFlag:      "",
			TokenFlag:    team1Token,
			TeamFlag:     team1TeamDomain,
			expectedAuth: mockAuthTeam1,
			expectedErr:  slackerror.Error{},
		},
		"token with mismatched workspace flag": {
			AppFlag:      "",
			TokenFlag:    team1Token,
			TeamFlag:     team2TeamDomain,
			expectedAuth: types.SlackAuth{},
			expectedErr:  *slackerror.New(slackerror.ErrInvalidToken),
		},
		"standalone token for another team": {
			AppFlag:      "",
			TokenFlag:    team2Token,
			TeamFlag:     "",
			expectedAuth: mockAuthTeam2,
			expectedErr:  slackerror.Error{},
		},
		"token with matching app flag": {
			AppFlag:      "A1EXAMPLE03",
			TokenFlag:    team2Token,
			TeamFlag:     "",
			expectedAuth: mockAuthTeam2,
			expectedErr:  slackerror.Error{},
		},
		"token with matching app and workspace flag": {
			AppFlag:      "A1EXAMPLE03",
			TokenFlag:    team2Token,
			TeamFlag:     team2TeamID,
			expectedAuth: mockAuthTeam2,
			expectedErr:  slackerror.Error{},
		},
		"token and workspace with app flag for an app that doesn't exist": {
			AppFlag:      "A1EXAMPLE04",
			TokenFlag:    team2Token,
			TeamFlag:     team2TeamID,
			expectedAuth: types.SlackAuth{},
			expectedErr:  *slackerror.New(slackerror.ErrAppNotFound),
		},
		"token and workspace with mismatched app flag for an app that does exist": {
			AppFlag:      "A1EXAMPLE03", // this app exists just not for team1
			TokenFlag:    team1Token,
			TeamFlag:     team1TeamID,
			expectedAuth: mockAuthTeam1,
			expectedErr:  *slackerror.New(slackerror.ErrInvalidToken),
		},
		"token with mismatched app flag": {
			AppFlag:      "A1EXAMPLE01",
			TokenFlag:    team2Token,
			TeamFlag:     "",
			expectedAuth: types.SlackAuth{},
			expectedErr:  *slackerror.New(slackerror.ErrInvalidToken),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			clients.Config.AppFlag = test.AppFlag
			clients.Config.TokenFlag = test.TokenFlag
			clients.Config.TeamFlag = test.TeamFlag

			workspaceApps, err := getTeamApps(ctx, clients)
			require.NoError(t, err)

			auth, err := filterAuthsByToken(ctx, clients, workspaceApps)
			auth.LastUpdated = time.Time{} // ignore time for this test

			if err != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Code, slackerror.ToSlackError(err).Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedAuth, auth)
			}
		})
	}
}

//
// AppSelectPrompt tests
//

func TestPrompt_AppSelectPrompt_SelectedAuthExpired_UserReAuthenticates(t *testing.T) {
	// Setup
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	// Auth is present but invalid
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	mockReauthentication(clientsMock)
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)

	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("app"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: "deployed",
		Index:  1,
	}, nil)

	clientsMock.IO.On(SelectPrompt, mock.Anything, "Select a team", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	err := clients.AppClient().SaveDeployed(ctx, deployedTeam1InstalledApp)
	require.NoError(t, err)

	// Execute test

	selection, err := AppSelectPrompt(ctx, clients, ShowAllApps)
	require.NoError(t, err)
	selection.Auth.LastUpdated = time.Time{} // ignore time for this test
	require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
	clientsMock.API.AssertCalled(t, "ExchangeAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPrompt_AppSelectPrompt_AuthsNoApps(t *testing.T) {

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clientsMock.AddDefaultMocks()

	// Execute test
	selectedApp, err := AppSelectPrompt(ctx, clients, AppInstallStatus(ShowInstalledAppsOnly))
	require.Equal(t, selectedApp, SelectedApp{})
	require.Error(t, err, slackerror.New(slackerror.ErrInstallationRequired))
}

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

func TestPrompt_AppSelectPrompt_AuthsWithDeployedAppInstalled_ShowAllApps(t *testing.T) {

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamDomain: team1TeamDomain,
		TeamID:     team1TeamID,
		AppID:      "A1EXAMPLE01",
	})
	require.NoError(t, err)

	clientsMock.IO.On(SelectPrompt, mock.Anything, SelectTeamPrompt, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("app"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: "deployed",
		Index:  1,
	}, nil)

	// Execute test
	selectedApp, err := AppSelectPrompt(ctx, clients, ShowAllApps)
	require.NoError(t, err)

	app, err := clients.AppClient().GetDeployed(ctx, team1TeamID)
	require.NoError(t, err)
	app.InstallStatus = types.AppStatusInstalled
	require.Equal(t, SelectedApp{App: app, Auth: fakeAuthsByTeamDomain[team1TeamDomain]}, selectedApp)
}

func TestPrompt_AppSelectPrompt_AuthsWithDeployedAppInstalled_ShowInstalledAppsOnly(t *testing.T) {

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	// Installed app
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:     team1TeamID,
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE01",
	})
	require.NoError(t, err)

	clientsMock.IO.On(SelectPrompt, mock.Anything, SelectTeamPrompt, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("app"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: "deployed",
		Index:  0,
	}, nil)

	// Execute test
	selectedApp, err := AppSelectPrompt(ctx, clients, ShowInstalledAppsOnly)
	require.NoError(t, err)

	app, err := clients.AppClient().GetDeployed(ctx, team1TeamID)
	app.InstallStatus = types.AppStatusInstalled
	require.NoError(t, err)
	require.Equal(t, selectedApp, SelectedApp{App: app, Auth: fakeAuthsByTeamDomain[team1TeamDomain]})
}

func TestPrompt_AppSelectPrompt_AuthsWithDeployedAppInstalled_InstalledAppOnly_Flags(t *testing.T) {

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Installed app
	deployedApp := types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	}

	err := clients.AppClient().SaveDeployed(ctx, deployedApp)
	require.NoError(t, err)

	// Execute tests
	tests := []struct {
		title     string
		appFlag   string
		teamFlag  string
		err       *slackerror.Error
		selection SelectedApp
	}{
		{
			"standalone app ID",
			"A1EXAMPLE01",
			"",
			nil,
			SelectedApp{App: deployedApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"standalone app environment",
			"deployed",
			"",
			slackerror.New(slackerror.ErrCredentialsNotFound),
			SelectedApp{},
		}, {
			"standalone team ID",
			"",
			team1TeamDomain,
			slackerror.New(slackerror.ErrInvalidAppFlag),
			SelectedApp{},
		}, {
			"app ID with matching team",
			"A1EXAMPLE01",
			team1TeamDomain,
			nil,
			SelectedApp{App: deployedApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"app ID with mismatched team",
			"A1EXAMPLE01",
			team2TeamDomain,
			slackerror.New(slackerror.ErrAppAuthTeamMismatch),
			SelectedApp{},
		}, {
			"unknown app ID",
			"A1EXAMPLE23",
			"",
			slackerror.New(slackerror.ErrAppNotFound),
			SelectedApp{},
		}, {
			"flags for deployed environment on a team",
			"deploy",
			team1TeamDomain,
			nil,
			SelectedApp{App: deployedApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"flags for local environment on a team",
			"local",
			team1TeamDomain,
			slackerror.New(slackerror.ErrInstallationRequired),
			SelectedApp{},
		}, {
			"invalid app ID flag",
			"brokenflag",
			team1TeamDomain,
			slackerror.New(slackerror.ErrInvalidAppFlag),
			SelectedApp{},
		},
	}

	for _, test := range tests {
		clientsMock.Config.AppFlag = test.appFlag
		clientsMock.Config.TeamFlag = test.teamFlag
		clientsMock.IO.On(SelectPrompt, mock.Anything, SelectTeamPrompt, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Flag:   true,
			Option: test.teamFlag,
		}, nil)
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("app"),
		})).Return(iostreams.SelectPromptResponse{
			Flag:   true,
			Option: test.appFlag,
		}, nil)

		selectedApp, err := AppSelectPrompt(ctx, clients, ShowInstalledAppsOnly)
		if test.err == nil {
			require.NoError(t, err, test.title)

			_, err := clients.AppClient().GetDeployed(ctx, test.selection.App.TeamID)
			require.NoError(t, err, test.title)
			require.Equal(t, selectedApp, test.selection, test.title)
		} else if assert.Error(t, err, test.title) {
			assert.Equal(t, test.err.Code, err.(*slackerror.Error).Code, test.title)
			require.Equal(t, selectedApp, SelectedApp{}, test.title)
		}
	}
}

func TestPrompt_AppSelectPrompt_AuthsWithBothEnvsInstalled_InstalledAppOnly_Flags(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())

	mockAuthTeam1 := fakeAuthsByTeamDomain[team1TeamDomain]
	mockAuthTeam1.Token = team1Token
	mockAuthTeam2 := fakeAuthsByTeamDomain[team2TeamDomain]
	mockAuthTeam2.Token = team2Token

	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{
				{AppID: "A1EXAMPLE01", Installed: true},
				{AppID: "A1EXAMPLE02", Installed: true},
			},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mockAuthTeam1.TeamID).Return(mockAuthTeam1, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mockAuthTeam2.TeamID).Return(mockAuthTeam2, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)

	clientsMock.Auth.On(AuthWithToken, mock.Anything, team1Token).
		Return(mockAuthTeam1, nil)
	clientsMock.Auth.On(AuthWithToken, mock.Anything, team2Token).
		Return(mockAuthTeam2, nil)

	clientsMock.AddDefaultMocks()
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("app"),
	})).Return(iostreams.SelectPromptResponse{
		Flag:   true,
		Option: "A1EXAMPLE02",
	}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Installed app team 1 deployed
	deployedApp := types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	}

	err := clients.AppClient().SaveDeployed(ctx, deployedApp)
	require.NoError(t, err)

	// Installed app team 1 local
	localApp := types.App{
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE02",
		TeamID:        team1TeamID,
		UserID:        team1UserID,
		IsDev:         true,
		InstallStatus: types.AppStatusInstalled,
	}

	err = clients.AppClient().SaveLocal(ctx, localApp)
	require.NoError(t, err)

	// Execute tests
	tests := []struct {
		title             string
		appFlag           string
		teamFlag          string
		tokenFlag         string
		err               *slackerror.Error
		expectedSelection SelectedApp
	}{
		{
			"standalone app ID for deployed app",
			"A1EXAMPLE01",
			"",
			"",
			nil,
			SelectedApp{App: deployedApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"standalone token, select the deployed app",
			"deploy",
			"",
			team1Token,
			nil,
			SelectedApp{App: deployedApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"app environment and team domain for deployed app",
			"deployed",
			team1TeamDomain,
			"",
			nil,
			SelectedApp{App: deployedApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"standalone app ID for local app",
			"A1EXAMPLE02",
			"",
			"",
			nil,
			SelectedApp{App: localApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"app environment and team domain for local app",
			"local",
			team1TeamDomain,
			"",
			nil,
			SelectedApp{App: localApp, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"mismatched app ID for team domain",
			"A1EXAMPLE01",
			team2TeamDomain,
			"",
			slackerror.New(slackerror.ErrAppAuthTeamMismatch),
			SelectedApp{},
		}, {
			"app environment not installed for team domain",
			"local",
			team2TeamDomain,
			"",
			slackerror.New(slackerror.ErrInstallationRequired),
			SelectedApp{},
		}, {
			"unknown team domain",
			"local",
			"team3",
			"",
			slackerror.New(slackerror.ErrTeamNotFound),
			SelectedApp{},
		},
	}

	for _, test := range tests {
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag
		clients.Config.TokenFlag = test.tokenFlag

		actualSelected, err := AppSelectPrompt(ctx, clients, ShowInstalledAppsOnly)
		actualSelected.Auth.LastUpdated = time.Time{} // ignore time for this test

		if test.err == nil {
			require.NoError(t, err)

			// App should exist
			_, err := clients.AppClient().GetDeployed(ctx, test.expectedSelection.App.TeamID)
			require.NoError(t, err)

			if test.tokenFlag != "" {
				assert.Equal(t, test.tokenFlag, actualSelected.Auth.Token, test.title, "should use provided token")
				test.expectedSelection.Auth.Token = test.tokenFlag
			}

			require.Equal(t, test.expectedSelection, actualSelected, test.title)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.err.Code, err.(*slackerror.Error).Code)
			require.Equal(t, SelectedApp{}, actualSelected)
		}
	}
}

func TestPrompt_AppSelectPrompt_AuthsWithBothEnvsInstalled_MultiWorkspaceAllApps_Flags(t *testing.T) {

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{
				{AppID: "A1EXAMPLE01", Installed: true},
				{AppID: "A1EXAMPLE02", Installed: true},
				{AppID: "A1EXAMPLE03", Installed: true},
			}}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Installed app

	// Deployed App Team 1
	deployedApp1 := types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	}

	err := clients.AppClient().SaveDeployed(ctx, deployedApp1)
	require.NoError(t, err)

	// Local App Team 1
	localApp1 := types.App{
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE02",
		TeamID:        team1TeamID,
		UserID:        team1UserID,
		IsDev:         true,
		InstallStatus: types.AppStatusInstalled,
	}

	err = clients.AppClient().SaveLocal(ctx, localApp1)
	require.NoError(t, err)

	// Deployed App Team 2
	deployedApp2 := types.App{
		TeamID:        team2TeamID,
		TeamDomain:    team2TeamDomain,
		AppID:         "A1EXAMPLE03",
		InstallStatus: types.AppStatusInstalled,
	}

	err = clients.AppClient().SaveDeployed(ctx, deployedApp2)
	require.NoError(t, err)

	// This should be a new app since we have not saved any pre-existing
	// local app
	localApp2, err := clients.AppClient().GetLocal(ctx, team2TeamID)
	require.NoError(t, err)
	require.True(t, localApp2.IsNew())

	// Execute tests
	tests := []struct {
		appFlag           string
		teamFlag          string
		err               *slackerror.Error
		expectedSelection SelectedApp
	}{
		{
			"A1EXAMPLE01",
			"",
			nil,
			SelectedApp{App: deployedApp1, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"A1EXAMPLE02",
			"",
			nil,
			SelectedApp{App: localApp1, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"A1EXAMPLE03",
			"",
			nil,
			SelectedApp{App: deployedApp2, Auth: fakeAuthsByTeamDomain[team2TeamDomain]},
		}, {
			"deployed",
			team1TeamDomain,
			nil,
			SelectedApp{App: deployedApp1, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"deploy",
			team2TeamDomain,
			nil,
			SelectedApp{App: deployedApp2, Auth: fakeAuthsByTeamDomain[team2TeamDomain]},
		}, {
			"local",
			team1TeamDomain,
			nil,
			SelectedApp{App: localApp1, Auth: fakeAuthsByTeamDomain[team1TeamDomain]},
		}, {
			"local",
			team2TeamDomain,
			nil,
			SelectedApp{App: localApp2, Auth: fakeAuthsByTeamDomain[team2TeamDomain]},
		},
	}

	for _, test := range tests {
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag
		selectedApp, err := AppSelectPrompt(ctx, clients, ShowAllApps)

		if test.err == nil {
			require.NoError(t, err)

			_, err := clients.AppClient().GetDeployed(ctx, test.expectedSelection.App.TeamID)

			require.NoError(t, err)
			require.Equal(t, test.expectedSelection, selectedApp)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.err.Code, err.(*slackerror.Error).Code)
			require.Equal(t, selectedApp, SelectedApp{})
		}
	}
}

func TestPrompt_AppSelectPrompt_AuthsWithHostedInstalled_AllApps_CreateNew(t *testing.T) {

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{
				{AppID: "A1EXAMPLE01", Installed: true},
				{AppID: "A1EXAMPLE02", Installed: true},
			}}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Installed apps
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE02",
		TeamID:        team1TeamID,
		IsDev:         true,
		UserID:        team1UserID,
		InstallStatus: types.AppStatusInstalled,
	})
	require.NoError(t, err)

	// Create a new local app in team2
	clientsMock.IO.On(SelectPrompt, mock.Anything, SelectTeamPrompt, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: "Install to a new workspace",
		Index:  1,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: "deployed",
		Index:  0,
	}, nil)

	// Execute test
	selectedApp, err := AppSelectPrompt(ctx, clients, ShowAllApps)
	require.NoError(t, err)

	app, err := clients.AppClient().GetLocal(ctx, team2TeamID)
	require.NoError(t, err)
	expected := SelectedApp{App: app, Auth: fakeAuthsByTeamDomain[team2TeamDomain]}
	require.Equal(t, expected, selectedApp)
}

func TestPrompt_AppSelectPrompt_ShowExpectedLabels(t *testing.T) {

	// Set up mocks

	setupClientsMock := func() *shared.ClientsMock {
		clientsMock := shared.NewClientsMock()
		auths := append(fakeAuthsByTeamDomainSlice, types.SlackAuth{
			TeamDomain: "team3",
			TeamID:     "T3",
			UserID:     "U3",
			Token:      "xoxe.xoxp-2-token",
		})
		clientsMock.Auth.On(Auths, mock.Anything).Return(auths, nil)
		clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
		clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
		clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					{AppID: deployedTeam1InstalledAppID, Installed: deployedTeam1AppIsInstalled},
					{AppID: localTeam1UninstalledAppID, Installed: localTeam1AppIsInstalled},
					{AppID: deployedTeam2UninstalledAppID, Installed: deployedTeam2AppIsInstalled},
					{AppID: localTeam2InstalledAppID, Installed: localTeam2AppIsInstalled},
				},
			}, nil)
		clientsMock.AddDefaultMocks()

		return clientsMock
	}

	saveApps := func(ctx context.Context, clients *shared.ClientFactory) {
		// Save deployed apps
		err := clients.AppClient().SaveDeployed(ctx, types.App{
			TeamDomain: team1TeamDomain,
			AppID:      deployedTeam1InstalledAppID,
			TeamID:     team1TeamID,
		})
		require.NoError(t, err)
		err = clients.AppClient().SaveDeployed(ctx, types.App{
			TeamDomain: team2TeamDomain,
			AppID:      deployedTeam2UninstalledAppID,
			TeamID:     team2TeamID,
		})
		require.NoError(t, err)

		// Save local apps
		err = clients.AppClient().SaveLocal(ctx, types.App{
			TeamDomain: team1TeamDomain,
			AppID:      localTeam1UninstalledAppID,
			TeamID:     team1TeamID,
			IsDev:      true,
			UserID:     team1UserID,
		})
		require.NoError(t, err)
		err = clients.AppClient().SaveLocal(ctx, types.App{
			TeamDomain: team2TeamDomain,
			AppID:      localTeam2InstalledAppID,
			TeamID:     team2TeamID,
			IsDev:      true,
			UserID:     team2UserID,
		})
		require.NoError(t, err)
	}

	// Execute tests

	var tests = []struct {
		name               string
		status             AppInstallStatus
		expectedTeamLabels []string
		selectedTeamIndex  int
		expectedAppLabels  []string
		selectedAppIndex   int
		expectedApp        func(ctx context.Context, clients *shared.ClientFactory) SelectedApp
	}{
		{
			name:   "All apps; select local installed app",
			status: ShowAllApps,
			expectedTeamLabels: []string{
				style.TeamSelectLabel(team1TeamDomain, team1TeamID),
				style.TeamSelectLabel(team2TeamDomain, team2TeamID),
				style.Secondary(appInstallPromptNew),
			},
			selectedTeamIndex: 1,
			expectedAppLabels: []string{
				style.AppSelectLabel(Local, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
				style.AppSelectLabel(Deployed, deployedTeam2UninstalledAppID, !deployedTeam2AppIsInstalled),
			},
			selectedAppIndex: 0,
			expectedApp: func(ctx context.Context, clients *shared.ClientFactory) SelectedApp {
				app, _ := clients.AppClient().GetLocal(ctx, team2TeamID)
				app.InstallStatus = types.AppStatusInstalled
				return SelectedApp{
					App:  app,
					Auth: fakeAuthsByTeamDomain[team2TeamDomain],
				}
			},
		},
		{
			name:   "Installed apps only; select deployed installed app",
			status: ShowInstalledAppsOnly,
			expectedTeamLabels: []string{
				style.TeamSelectLabel(team1TeamDomain, team1TeamID),
				style.TeamSelectLabel(team2TeamDomain, team2TeamID),
			},
			selectedTeamIndex: 0,
			expectedAppLabels: []string{
				style.AppSelectLabel(Deployed, deployedTeam1InstalledAppID, !deployedTeam1AppIsInstalled),
			},
			selectedAppIndex: 0,
			expectedApp: func(ctx context.Context, clients *shared.ClientFactory) SelectedApp {

				app, _ := clients.AppClient().GetDeployed(ctx, team1TeamID)
				app.InstallStatus = types.AppStatusInstalled
				return SelectedApp{
					App:  app,
					Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				}
			},
		},
		{
			name:   "Installed apps only; select local installed app",
			status: ShowInstalledAppsOnly,
			expectedTeamLabels: []string{
				style.TeamSelectLabel(team1TeamDomain, team1TeamID),
				style.TeamSelectLabel(team2TeamDomain, team2TeamID),
			},
			selectedTeamIndex: 1, // should select team2
			expectedAppLabels: []string{
				style.AppSelectLabel(Local, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
			},
			selectedAppIndex: 0, // should select the local app
			expectedApp: func(ctx context.Context, clients *shared.ClientFactory) SelectedApp {
				app, _ := clients.AppClient().GetLocal(ctx, team2TeamID)
				app.InstallStatus = types.AppStatusInstalled
				return SelectedApp{
					App:  app,
					Auth: fakeAuthsByTeamDomain[team2TeamDomain],
				}
			},
		},
		{
			name:   "Installed and uninstalled apps; select deployed uninstalled app",
			status: ShowInstalledAndUninstalledApps,
			expectedTeamLabels: []string{
				style.TeamSelectLabel(team1TeamDomain, team1TeamID),
				style.TeamSelectLabel(team2TeamDomain, team2TeamID),
			},
			selectedTeamIndex: 1,
			expectedAppLabels: []string{
				style.AppSelectLabel(Local, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
				style.AppSelectLabel(Deployed, deployedTeam2UninstalledAppID, !deployedTeam2AppIsInstalled),
			},
			selectedAppIndex: 1,
			expectedApp: func(ctx context.Context, clients *shared.ClientFactory) SelectedApp {

				app, _ := clients.AppClient().GetDeployed(ctx, team2TeamID)
				app.InstallStatus = types.AppStatusUninstalled
				return SelectedApp{
					App:  app,
					Auth: fakeAuthsByTeamDomain[team2TeamDomain],
				}
			},
		},
		{
			name:   "Installed and non-existent apps; select deployed installed app",
			status: ShowInstalledAndNewApps,
			expectedTeamLabels: []string{
				style.TeamSelectLabel(team1TeamDomain, team1TeamID),
				style.TeamSelectLabel(team2TeamDomain, team2TeamID),
				style.Secondary(appInstallPromptNew),
			},
			selectedTeamIndex: 0,
			expectedAppLabels: []string{
				style.AppSelectLabel(Deployed, deployedTeam1InstalledAppID, !deployedTeam1AppIsInstalled),
			},
			selectedAppIndex: 0,
			expectedApp: func(ctx context.Context, clients *shared.ClientFactory) SelectedApp {

				app, _ := clients.AppClient().GetDeployed(ctx, team1TeamID)
				app.InstallStatus = types.AppStatusInstalled
				return SelectedApp{
					App:  app,
					Auth: fakeAuthsByTeamDomain[team1TeamDomain],
				}
			},
		},
	}

	for _, test := range tests {
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := setupClientsMock()

		// On select a team, choose
		clientsMock.IO.On(SelectPrompt, mock.Anything, SelectTeamPrompt, test.expectedTeamLabels, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.expectedTeamLabels[test.selectedTeamIndex],
			Index:  test.selectedTeamIndex,
		}, nil)

		// On chosen deployed or local
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", test.expectedAppLabels, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("app"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.expectedAppLabels[test.selectedAppIndex],
			Index:  test.selectedAppIndex,
		}, nil)

		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		saveApps(ctx, clients)

		selectedApp, err := AppSelectPrompt(ctx, clients, test.status)
		require.NoError(t, err)

		expectedApp := test.expectedApp(ctx, clients)
		require.Equal(t, expectedApp, selectedApp)
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
		"selects a saved applications using prompts": {
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
		"creates new application if selected": {
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
		"creates new application for app environment flag and team id flag if not app saved": {
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
		"filters deployed apps for app environment flag before selection": {
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
		"filters local apps for app environment flag before selection": {
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
		"creates new application with token flag and team id flag if app not saved": {
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

func TestPrompt_TeamAppSelectPrompt_SelectedAuthExpired_UserReAuthenticates(t *testing.T) {
	// Setup
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	// Auth is present but invalid
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	mockReauthentication(clientsMock)
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)

	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	err := clients.AppClient().SaveDeployed(ctx, deployedTeam1InstalledApp)
	require.NoError(t, err)

	// Execute test

	selection, err := TeamAppSelectPrompt(ctx, clients, ShowHostedOnly, ShowAllApps)
	require.NoError(t, err)
	selection.Auth.LastUpdated = time.Time{} // ignore time for this test
	require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
	clientsMock.API.AssertCalled(t, "ExchangeAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPrompt_TeamAppSelectPrompt_NoAuths_UserReAuthenticates(t *testing.T) {
	// Setup
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	// No auths present
	clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{}, nil)
	mockReauthentication(clientsMock)
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)

	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	err := clients.AppClient().SaveDeployed(ctx, deployedTeam1InstalledApp)
	require.NoError(t, err)

	// Execute test

	selection, err := TeamAppSelectPrompt(ctx, clients, ShowHostedOnly, ShowAllApps)
	require.NoError(t, err)
	selection.Auth.LastUpdated = time.Time{} // ignore time for this test
	require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
}

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

func TestPrompt_TeamAppSelectPrompt_TeamNotFoundFor_TeamFlag(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = "unauthed-domain" // resets each loop
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)
		if assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrTeamNotFound, err.(*slackerror.Error).Code)
		}
		require.Equal(t, types.SlackAuth{}, selection.Auth)
	}
}

func TestPrompt_TeamAppSelectPrompt_NoApps(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(fakeAuthsByTeamDomain[team1TeamDomain], nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team2TeamID).Return(fakeAuthsByTeamDomain[team2TeamDomain], nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install the app to team1
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: appInstallPromptNew,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: appInstallPromptNew,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  0,
	}, nil)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors if installation is required, otherwise expect a successful choice
		if test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_NoInstalls_TeamFlagDomain(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team1TeamDomain
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors if installation is required, otherwise expect a successful choice
		if test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_NoInstalls_TeamFlagID(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)

	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team2TeamID
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors if installation is required, otherwise expect a successful choice
		if test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team2TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_NoInstalls_Flags(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Execute tests
	tests := []struct {
		env    AppEnvironmentType
		status AppInstallStatus

		appFlag      string
		teamFlag     string
		err          *slackerror.Error
		expectedAuth types.SlackAuth
	}{
		{
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"deploy",
			"",
			slackerror.New(slackerror.ErrTeamFlagRequired),
			types.SlackAuth{},
		}, {
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"local",
			"",
			slackerror.New(slackerror.ErrTeamFlagRequired),
			types.SlackAuth{},
		}, {
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"deploy",
			team1TeamDomain,
			slackerror.New(slackerror.ErrInstallationRequired),
			types.SlackAuth{},
		}, {
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"local",
			team2TeamDomain,
			slackerror.New(slackerror.ErrInstallationRequired),
			types.SlackAuth{},
		}, {
			ShowHostedOnly,
			ShowAllApps,
			"local",
			"",
			slackerror.New(slackerror.ErrLocalAppNotSupported),
			types.SlackAuth{},
		}, {
			ShowLocalOnly,
			ShowAllApps,
			"deployed",
			"",
			slackerror.New(slackerror.ErrDeployedAppNotSupported),
			types.SlackAuth{},
		}, {
			ShowHostedOnly,
			ShowAllApps,
			"A1234567890",
			"",
			slackerror.New(slackerror.ErrAppNotFound),
			types.SlackAuth{},
		}, {
			ShowLocalOnly,
			ShowAllApps,
			"A1234567890",
			"",
			slackerror.New(slackerror.ErrAppNotFound),
			types.SlackAuth{},
		}, {
			ShowLocalOnly,
			ShowAllApps,
			"local",
			team1TeamDomain,
			nil,
			fakeAuthsByTeamDomain[team1TeamDomain],
		}, {
			ShowHostedOnly,
			ShowAllApps,
			"deploy",
			team2TeamDomain,
			nil,
			fakeAuthsByTeamDomain[team2TeamDomain],
		},
	}

	for _, test := range tests {
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		if test.err == nil {
			require.NoError(t, err)

			_, err := clients.AppClient().GetDeployed(ctx, test.expectedAuth.TeamID)
			require.NoError(t, err)
			require.Equal(t, test.expectedAuth, selection.Auth)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.err.Code, err.(*slackerror.Error).Code)
			require.Equal(t, types.SlackAuth{}, selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_TokenFlag(t *testing.T) {
	appInstallStatus := []api.AppStatusResultAppInfo{
		{AppID: "A1EXAMPLE01", Installed: true},
		{AppID: "A1EXAMPLE02", Installed: true},
		{AppID: "A1EXAMPLE04", Installed: false},
	}
	installedHostedApp := types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	}
	installedLocalApp := types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE02",
		InstallStatus: types.AppStatusInstalled,
		UserID:        "U1",
		IsDev:         true,
	}
	uninstalledHostedApp := types.App{
		TeamID:        team2TeamID,
		TeamDomain:    team2TeamDomain,
		AppID:         "A1EXAMPLE03",
		InstallStatus: types.AppInstallationStatusUnknown,
	}
	uninstalledLocalApp := types.App{
		TeamID:        team2TeamID,
		TeamDomain:    team2TeamDomain,
		AppID:         "A1EXAMPLE04",
		InstallStatus: types.AppStatusUninstalled,
		UserID:        "U2",
		IsDev:         true,
	}

	var tests = map[string]struct {
		env        AppEnvironmentType
		status     AppInstallStatus
		teamDomain string
		token      string
		app        types.App
		err        *slackerror.Error
	}{
		"return the hosted app of a valid token": {
			ShowHostedOnly,
			ShowAllApps,
			team1TeamDomain,
			team1Token,
			installedHostedApp,
			nil,
		},
		"error when installation is required": {
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			team2TeamDomain,
			team2Token,
			uninstalledHostedApp,
			slackerror.New(slackerror.ErrInstallationRequired),
		},
		"return an uninstalled app if allowed": {
			ShowLocalOnly,
			ShowAllApps,
			team2TeamDomain,
			team2Token,
			uninstalledLocalApp,
			nil,
		},
		"return the local app of a valid token": {
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			team1TeamDomain,
			team1Token,
			installedLocalApp,
			nil,
		},
	}

	for name, test := range tests {
		ctx := slackcontext.MockContext(t.Context())

		mockAuth := fakeAuthsByTeamDomain[test.teamDomain]
		mockAuth.Token = test.token

		clientsMock := shared.NewClientsMock()

		clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			api.GetAppStatusResult{Apps: appInstallStatus}, nil)
		clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
		clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).
			Return(types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound))
		clientsMock.Auth.On(AuthWithToken, mock.Anything, test.token).
			Return(mockAuth, nil)
		clientsMock.AddDefaultMocks()

		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		var err error
		err = clients.AppClient().SaveDeployed(ctx, installedHostedApp)
		require.NoError(t, err)
		err = clients.AppClient().SaveLocal(ctx, installedLocalApp)
		require.NoError(t, err)
		err = clients.AppClient().SaveLocal(ctx, uninstalledLocalApp)
		require.NoError(t, err)

		clients.Config.TokenFlag = test.token
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		if test.err != nil && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, selection.Auth, fakeAuthsByTeamDomain[test.teamDomain], name)
			require.Equal(t, selection.App, test.app, name)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_HostedAppsOnly(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}, {AppID: "A124", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Select team2
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team2TeamDomain,
		Index:  1,
	}, nil)

	// Install the app to team2
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  1,
	}, nil)

	// Installed apps
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:        team2TeamID,
		TeamDomain:    team2TeamDomain,
		AppID:         "A124",
		InstallStatus: types.AppStatusInstalled,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors if installation is required, otherwise expect a successful choice
		if test.status == ShowInstalledAppsOnly && test.env == ShowLocalOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team2TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_HostedAppsOnly_TeamFlagDomain(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install apps
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team1TeamDomain
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors when installation is required for local apps, otherwise expect a successful choice
		if test.env == ShowLocalOnly && test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, selection.Auth, fakeAuthsByTeamDomain[team1TeamDomain])
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_HostedAppsOnly_TeamFlagID(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install apps
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:        team1TeamID,
		TeamDomain:    team1TeamDomain,
		AppID:         "A1EXAMPLE01",
		InstallStatus: types.AppStatusInstalled,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team1TeamID
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors when installation is required for local apps, otherwise expect a successful choice
		if test.env == ShowLocalOnly && test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_LocalAppsOnly(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A1EXAMPLE01", Installed: true}, {AppID: "A124", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(fakeAuthsByTeamDomain[team1TeamDomain], nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team2TeamID).Return(fakeAuthsByTeamDomain[team2TeamDomain], nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install the app to team1
	clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: team1TeamDomain,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  0,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  0,
	}, nil)

	// Installed apps
	err := clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain:    "dev",
		AppID:         "A1EXAMPLE01",
		TeamID:        team1TeamID,
		IsDev:         true,
		InstallStatus: types.AppStatusInstalled,
		UserID:        team1UserID,
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain:    "dev",
		AppID:         "A124",
		TeamID:        team2TeamID,
		IsDev:         true,
		InstallStatus: types.AppStatusInstalled,
		UserID:        team2UserID,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors if installation is required, otherwise expect a successful choice
		if test.status == ShowInstalledAppsOnly && test.env == ShowHostedOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_LocalAppsOnly_TeamFlagDomain(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A124", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(fakeAuthsByTeamDomain[team1TeamDomain], nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team2TeamID).Return(fakeAuthsByTeamDomain[team2TeamDomain], nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install apps
	err := clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain:    team2TeamDomain,
		AppID:         "A124",
		TeamID:        team2TeamID,
		IsDev:         true,
		InstallStatus: types.AppStatusInstalled,
		UserID:        team2UserID,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team2TeamDomain
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors when installation is required for hosted apps, otherwise expect a successful choice
		if test.env == ShowHostedOnly && test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team2TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_LocalAppsOnly_TeamFlagID(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: "A124", Installed: true}},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(fakeAuthsByTeamDomain[team1TeamDomain], nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team2TeamID).Return(fakeAuthsByTeamDomain[team2TeamDomain], nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install apps
	err := clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain:    team2TeamDomain,
		AppID:         "A124",
		TeamID:        team2TeamID,
		IsDev:         true,
		InstallStatus: types.AppStatusInstalled,
		UserID:        team2UserID,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team2TeamID
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		// Check for errors when installation is required for hosted apps, otherwise expect a successful choice
		if test.env == ShowHostedOnly && test.status == ShowInstalledAppsOnly && assert.Error(t, err) {
			assert.Equal(t, slackerror.ErrInstallationRequired, err.(*slackerror.Error).Code)
		} else {
			require.NoError(t, err)
			require.Equal(t, fakeAuthsByTeamDomain[team2TeamDomain], selection.Auth)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_AllApps(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{
				{AppID: "A1EXAMPLE01", Installed: true},
				{AppID: "A124", Installed: true},
				{AppID: "A1EXAMPLE01dev", Installed: true},
				{AppID: "A124dev", Installed: true},
			},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Select team2
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  1,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  1,
	}, nil)

	// Install apps
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:     team1TeamID,
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE01",
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveDeployed(ctx, types.App{
		TeamID:     team2TeamID,
		TeamDomain: team2TeamDomain,
		AppID:      "A124",
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE01dev",
		TeamID:     team1TeamID,
		IsDev:      true,
		UserID:     team1UserID,
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain: team2TeamDomain,
		AppID:      "A124dev",
		TeamID:     team2TeamID,
		IsDev:      true,
		UserID:     team2UserID,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		require.NoError(t, err)
		require.Equal(t, fakeAuthsByTeamDomain[team2TeamDomain], selection.Auth)
	}
}

func TestPrompt_TeamAppSelectPrompt_LegacyDevApps(t *testing.T) {
	// Test to ensure that legacy apps.dev.json entries which have
	// team_domain set as "dev" are overridden with the correct team_domain when the auth
	// context is known

	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{
				{AppID: "A1EXAMPLE01dev", Installed: true},
				{AppID: "A124dev", Installed: true},
			},
		}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Select team2
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  1,
	}, nil)
	clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: mock.Anything,
		Index:  1,
	}, nil)

	err := clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain: "dev", // mock legacy apps.dev.json which has teamDomain as 'dev'
		AppID:      "A1EXAMPLE01dev",
		TeamID:     team1TeamID,
		IsDev:      true,
		UserID:     team1UserID,
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain: "dev", // mock legacy apps.dev.json which has teamDomain as 'dev'
		AppID:      "A124dev",
		TeamID:     team2TeamID,
		IsDev:      true,
		UserID:     team2UserID,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		require.NoError(t, err)
		require.Equal(t, fakeAuthsByTeamDomain[team2TeamDomain], selection.Auth)

		team1App, err := clients.AppClient().GetLocal(ctx, team1TeamID)
		require.NoError(t, err)
		// app team domain should be overridden
		require.Equal(t, team1TeamDomain, team1App.TeamDomain)
		require.NotEqual(t, "dev", team1App.TeamDomain)

		team2App, err := clients.AppClient().GetLocal(ctx, team2TeamID)
		require.NoError(t, err)
		// app team domain should be overridden from "dev"
		// app team domain should be overridden
		require.Equal(t, team2TeamDomain, team2App.TeamDomain)
		require.NotEqual(t, "dev", team2App.TeamDomain)
	}
}

func TestPrompt_TeamAppSelectPrompt_ShowExpectedLabels(t *testing.T) {

	// Set up mocks

	setupClientsMock := func() *shared.ClientsMock {
		clientsMock := shared.NewClientsMock()
		clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
			api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					{AppID: deployedTeam1InstalledAppID, Installed: deployedTeam1AppIsInstalled},
					{AppID: localTeam1UninstalledAppID, Installed: localTeam1AppIsInstalled},
					{AppID: deployedTeam2UninstalledAppID, Installed: deployedTeam2AppIsInstalled},
					{AppID: localTeam2InstalledAppID, Installed: localTeam2AppIsInstalled},
				},
			}, nil)
		auths := append(fakeAuthsByTeamDomainSlice, types.SlackAuth{
			TeamDomain: "team3",
			TeamID:     "T3",
			UserID:     "U3",
			Token:      "xoxe.xoxp-2-token",
		})
		clientsMock.Auth.On(Auths, mock.Anything).Return(auths, nil)
		clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
		clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
		clientsMock.AddDefaultMocks()

		return clientsMock
	}

	saveApps := func(ctx context.Context, clients *shared.ClientFactory) {
		// Save deployed apps
		err := clients.AppClient().SaveDeployed(ctx, types.App{
			TeamDomain: team1TeamDomain,
			AppID:      deployedTeam1InstalledAppID,
			TeamID:     team1TeamID,
		})
		require.NoError(t, err)
		err = clients.AppClient().SaveDeployed(ctx, types.App{
			TeamDomain: team2TeamDomain,
			AppID:      deployedTeam2UninstalledAppID,
			TeamID:     team2TeamID,
		})
		require.NoError(t, err)

		// Save local apps
		err = clients.AppClient().SaveLocal(ctx, types.App{
			TeamDomain: team1TeamDomain,
			AppID:      localTeam1UninstalledAppID,
			TeamID:     team1TeamID,
			IsDev:      true,
			UserID:     team1UserID,
		})
		require.NoError(t, err)
		err = clients.AppClient().SaveLocal(ctx, types.App{
			TeamDomain: team2TeamDomain,
			AppID:      localTeam2InstalledAppID,
			TeamID:     team2TeamID,
			IsDev:      true,
			UserID:     team2UserID,
		})
		require.NoError(t, err)
	}

	// Execute tests

	var tests = []struct {
		env                   AppEnvironmentType
		status                AppInstallStatus
		promptText            string
		expectedTeamLabels    []string
		selectedTeamIndex     int
		expectedTeamSelection string
	}{
		{
			ShowHostedOnly,
			ShowAllApps,
			"Choose a deployed environment",
			[]string{
				style.TeamAppSelectLabel(team1TeamDomain, team1TeamID, deployedTeam1InstalledAppID, !deployedTeam1AppIsInstalled),
				style.TeamAppSelectLabel(team2TeamDomain, team2TeamID, deployedTeam2UninstalledAppID, !deployedTeam2AppIsInstalled),
				style.Secondary(appInstallPromptNew),
			},
			1,
			team2TeamDomain,
		},
		{
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"Choose a deployed environment",
			[]string{
				style.TeamAppSelectLabel(team1TeamDomain, team1TeamID, deployedTeam1InstalledAppID, !deployedTeam1AppIsInstalled),
			},
			0,
			team1TeamDomain,
		},
		{
			ShowHostedOnly,
			ShowInstalledAndNewApps,
			"Choose a deployed environment",
			[]string{
				style.TeamAppSelectLabel(team1TeamDomain, team1TeamID, deployedTeam1InstalledAppID, !deployedTeam1AppIsInstalled),
				style.Secondary(appInstallPromptNew),
			},
			0,
			team1TeamDomain,
		},
		{
			ShowHostedOnly,
			ShowInstalledAndUninstalledApps,
			"Choose a deployed environment",
			[]string{
				style.TeamAppSelectLabel(team1TeamDomain, team1TeamID, deployedTeam1InstalledAppID, !deployedTeam1AppIsInstalled),
				style.TeamAppSelectLabel(team2TeamDomain, team2TeamID, deployedTeam2UninstalledAppID, !deployedTeam2AppIsInstalled),
			},
			1,
			team2TeamDomain,
		},
		{
			ShowLocalOnly,
			ShowAllApps,
			"Choose a local environment",
			[]string{
				style.TeamAppSelectLabel(team1TeamDomain, team1TeamID, localTeam1UninstalledAppID, !localTeam1AppIsInstalled),
				style.TeamAppSelectLabel(team2TeamDomain, team2TeamID, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
				style.Secondary(appInstallPromptNew),
			},
			1,
			team2TeamDomain,
		},
		{
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"Choose a local environment",
			[]string{
				style.TeamAppSelectLabel(team2TeamDomain, team2TeamID, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
			},
			0,
			team2TeamDomain,
		},
		{
			ShowLocalOnly,
			ShowInstalledAndNewApps,
			"Choose a local environment",
			[]string{
				style.TeamAppSelectLabel(team2TeamDomain, team2TeamID, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
				style.Secondary(appInstallPromptNew),
			},
			0,
			team2TeamDomain,
		},
		{
			ShowLocalOnly,
			ShowInstalledAndUninstalledApps,
			"Choose a local environment",
			[]string{
				style.TeamAppSelectLabel(team1TeamDomain, team1TeamID, localTeam1UninstalledAppID, !localTeam1AppIsInstalled),
				style.TeamAppSelectLabel(team2TeamDomain, team2TeamID, localTeam2InstalledAppID, !localTeam2AppIsInstalled),
			},
			0,
			team1TeamDomain,
		},
	}

	for _, test := range tests {
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := setupClientsMock()
		clientsMock.IO.On(SelectPrompt, mock.Anything, test.promptText, test.expectedTeamLabels, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: mock.Anything,
			Index:  test.selectedTeamIndex,
		}, nil)
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		saveApps(ctx, clients)

		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		require.NoError(t, err)
		require.Equal(t, fakeAuthsByTeamDomain[test.expectedTeamSelection], selection.Auth)
	}
}

func TestPrompt_TeamAppSelectPrompt_AllApps_TeamFlagID(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install apps
	err := clients.AppClient().SaveDeployed(ctx, types.App{
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE01",
		TeamID:     team1TeamID,
	})
	require.NoError(t, err)

	err = clients.AppClient().SaveLocal(ctx, types.App{
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE02",
		TeamID:     team1TeamID,
		IsDev:      true,
		UserID:     team1UserID,
	})
	require.NoError(t, err)

	// Perform tests
	var tests = []struct {
		env    AppEnvironmentType
		status AppInstallStatus
	}{
		{ShowHostedOnly, ShowAllApps},
		{ShowHostedOnly, ShowInstalledAppsOnly},
		{ShowLocalOnly, ShowAllApps},
		{ShowLocalOnly, ShowInstalledAppsOnly},
	}

	for _, test := range tests {
		clients.Config.TeamFlag = team1TeamID
		selection, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		require.NoError(t, err)
		require.Equal(t, fakeAuthsByTeamDomain[team1TeamDomain], selection.Auth)
	}
}

func TestPrompt_TeamAppSelectPrompt_AllApps_Flags(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)
	clientsMock.Auth.On(Auths, mock.Anything).Return(fakeAuthsByTeamDomainSlice, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Install apps
	appTeam1Hosted := types.App{
		TeamDomain: team1TeamDomain,
		TeamID:     team1TeamID,
		AppID:      "A1EXAMPLE01",
	}
	err := clients.AppClient().SaveDeployed(ctx, appTeam1Hosted)
	require.NoError(t, err)

	appTeam1Local := types.App{
		TeamDomain: team1TeamDomain,
		AppID:      "A1EXAMPLE02",
		TeamID:     team1TeamID,
		IsDev:      true,
		UserID:     team1UserID,
	}
	err = clients.AppClient().SaveLocal(ctx, appTeam1Local)
	require.NoError(t, err)

	appTeam2Hosted := types.App{
		TeamDomain: team2TeamDomain,
		TeamID:     team2TeamID,
		AppID:      "A1EXAMPLE03",
	}
	err = clients.AppClient().SaveDeployed(ctx, appTeam2Hosted)
	require.NoError(t, err)

	// Execute tests
	tests := []struct {
		desc   string
		env    AppEnvironmentType
		status AppInstallStatus

		appFlag  string
		teamFlag string
		err      *slackerror.Error
		expected SelectedApp
	}{
		{
			"valid flags: --app deploy --team team1",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"deploy",
			team1TeamDomain,
			nil,
			SelectedApp{Auth: fakeAuthsByTeamDomain[team1TeamDomain], App: appTeam1Hosted},
		}, {
			"valid flags: --app <deploy_id>",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE03",
			"",
			nil,
			SelectedApp{Auth: fakeAuthsByTeamDomain[team2TeamDomain], App: appTeam2Hosted},
		}, {
			"valid flags: --app local --team team1",
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"local",
			team1TeamDomain,
			nil,
			SelectedApp{Auth: fakeAuthsByTeamDomain[team1TeamDomain], App: appTeam1Local},
		}, {
			"valid flags: --app <local_id>",
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE02",
			"",
			nil,
			SelectedApp{Auth: fakeAuthsByTeamDomain[team1TeamDomain], App: appTeam1Local},
		}, {
			"valid flags: --app <local_id> --team team1",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE01",
			team1TeamDomain,
			nil,
			SelectedApp{Auth: fakeAuthsByTeamDomain[team1TeamDomain], App: appTeam1Hosted},
		}, {
			"invalid flags: --app <team1_local_app_id> --team team2",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE01",
			team2TeamDomain,
			slackerror.New(slackerror.ErrAppNotFound),
			SelectedApp{},
		}, {
			"invalid flags: --app <unknown_app_id>",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE04",
			"",
			slackerror.New(slackerror.ErrAppNotFound),
			SelectedApp{},
		}, {
			"invalid flags: --app <team1_local_app_id> --team team2",
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE02",
			team2TeamDomain,
			slackerror.New(slackerror.ErrAppNotFound),
			SelectedApp{},
		}, {
			"invalid flags for local only: --app <deploy_app_id>",
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE01",
			"",
			slackerror.New(slackerror.ErrDeployedAppNotSupported),
			SelectedApp{},
		}, {
			"invalid flags for deploy only: --app <local_app_id>",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE02",
			"",
			slackerror.New(slackerror.ErrLocalAppNotSupported),
			SelectedApp{},
		}, {
			"invalid flags for local only: --app <team1_deploy_app_id> --team team1",
			ShowLocalOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE01",
			team1TeamDomain,
			slackerror.New(slackerror.ErrDeployedAppNotSupported),
			SelectedApp{},
		}, {
			"invalid flags for deploy only: --app <team1_local_app_id> --team team1",
			ShowHostedOnly,
			ShowInstalledAppsOnly,
			"A1EXAMPLE02",
			team1TeamDomain,
			slackerror.New(slackerror.ErrLocalAppNotSupported),
			SelectedApp{},
		},
	}

	for _, test := range tests {
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag
		actualSelected, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		if test.err == nil {
			require.NoError(t, err)

			_, err := clients.AppClient().GetDeployed(ctx, test.expected.Auth.TeamID)
			require.NoError(t, err)
			require.Equal(t, test.expected, actualSelected)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.err.Code, err.(*slackerror.Error).Code, test.desc)
			require.Equal(t, test.expected, actualSelected)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_AppSelectPrompt_EnterpriseWorkspaceApps_HasWorkspaceAuth(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)

	// Auths
	// Enterprise (org) Auth
	// Mock the enterprise that Team1 belongs to
	const enterprise1TeamDomain = "enterprise1"
	const enterprise1TeamID = "E1"
	const enterprise1UserID = "UE1"
	const enterprise1Token = "xoxp-xoxe-12345"

	authEnterprise1 := types.SlackAuth{
		TeamID:              enterprise1TeamID,
		TeamDomain:          enterprise1TeamDomain,
		Token:               enterprise1Token,
		IsEnterpriseInstall: true,
		UserID:              enterprise1UserID,
		EnterpriseID:        enterprise1TeamID,
	}

	// Set up team 1 to belong to enterprise 1
	authTeam1 := fakeAuthsByTeamDomain[team1TeamDomain]
	authTeam1.EnterpriseID = authEnterprise1.TeamID

	// Return one auth
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(authTeam1, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, enterprise1TeamID).Return(authEnterprise1, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)

	//
	// This test uses a single auth - Mock the underlying auth
	//
	clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{
		authTeam1,
	}, nil)

	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Save apps
	// Save a hosted and local enterprise workspace-level app
	appTeam1Hosted := types.App{
		TeamDomain:   team1TeamDomain,
		TeamID:       team1TeamID,
		AppID:        "A1EXAMPLE01",
		EnterpriseID: enterprise1TeamID,
	}

	appTeam1Local := types.App{
		TeamDomain:   team1TeamDomain,
		AppID:        "A1EXAMPLE02",
		TeamID:       team1TeamID,
		IsDev:        true,
		UserID:       team1UserID,
		EnterpriseID: enterprise1TeamID,
	}

	err := clients.AppClient().SaveDeployed(ctx, appTeam1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appTeam1Local)
	require.NoError(t, err)

	// Save a hosted and local enterprise / org level app
	appEnterprise1Hosted := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE03",
		EnterpriseID: enterprise1TeamID,
	}

	appEnterprise1Local := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE04",
		IsDev:        true,
		EnterpriseID: enterprise1TeamID,
	}

	err = clients.AppClient().SaveDeployed(ctx, appEnterprise1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appEnterprise1Local)
	require.NoError(t, err)

	// We won't mock standard workspace apps and auths in this test, as that
	// Should be well covered by other tests in this file.

	// Test cases
	// Execute tests
	tests := []struct {
		desc string

		expected      SelectedApp
		expectedError *slackerror.Error

		env    AppEnvironmentType
		status AppInstallStatus

		appFlag  string
		teamFlag string

		authState []types.SlackAuth
		authError error

		selectTeamDomain  string
		selectEnvironment string
	}{
		{
			// Test description
			"When there is an existing workspace auth, workspace app returned with existing auth",

			// Expected result
			SelectedApp{
				App:  appTeam1Hosted,
				Auth: authTeam1,
			},
			nil,

			// Env + app install status
			ShowHostedOnly,
			ShowInstalledAppsOnly,

			// flags
			"",
			"",

			// Auth state
			[]types.SlackAuth{
				authTeam1,
			},
			nil,

			// Selector preferences team 1
			team1TeamDomain,
			"Deployed",
		},
	}

	for _, test := range tests {

		// Set app flags
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag

		// Return the auth state depending on test specs
		clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{
			authTeam1,
		}, nil)

		clientsMock.AddDefaultMocks()

		// Finally we mock the select prompt based on test specs
		clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectTeamDomain,
			Index:  0,
		}, nil)
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectEnvironment,
			Index:  0,
		}, nil)
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectEnvironment,
			Index:  0,
		}, nil)
		// App selector mock
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("app"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "deployed",
			Index:  1,
		}, nil)

		actualSelected1, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)
		actualSelected2, err2 := AppSelectPrompt(ctx, clients, test.status)

		if test.expectedError == nil {
			require.NoError(t, err)
			require.NoError(t, err2)

			// There's a valid auth for this auth record
			_, err := clients.AppClient().GetDeployed(ctx, test.expected.Auth.TeamID)
			require.NoError(t, err)

			require.Equal(t, test.expected, actualSelected1)
			require.Equal(t, test.expected, actualSelected2)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.expectedError.Code, err.(*slackerror.Error).Code, test.desc)

			require.Equal(t, test.expected, actualSelected1)
			require.Equal(t, test.expected, actualSelected2)
		}
	}
}

func TestPrompt_TeamAppSelectPrompt_AppSelectPrompt_EnterpriseWorkspaceApps_MissingWorkspaceAuth_MissingOrgAuth_UserReAuthenticates(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)

	mockReauthentication(clientsMock)

	// Auths
	// Enterprise (org) Auth
	// Mock the enterprise that Team1 belongs to
	const enterprise1TeamDomain = "enterprise1"
	const enterprise1TeamID = "E1"
	const enterprise1UserID = "UE1"
	const enterprise1Token = "xoxp-xoxe-12345"

	authEnterprise1 := types.SlackAuth{
		TeamID:              enterprise1TeamID,
		TeamDomain:          enterprise1TeamDomain,
		Token:               enterprise1Token,
		IsEnterpriseInstall: true,
		UserID:              enterprise1UserID,
		EnterpriseID:        enterprise1TeamID,
	}

	// Set up team 1 to belong to enterprise 1
	authTeam1 := fakeAuthsByTeamDomain[team1TeamDomain]
	authTeam1.EnterpriseID = authEnterprise1.TeamID

	// Return one auth
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(authTeam1, nil)
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, enterprise1TeamID).Return(authEnterprise1, nil)

	// This test uses zero auths
	clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{}, nil)

	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Save apps
	// Save a hosted and local enterprise workspace-level app
	appTeam1Hosted := types.App{
		TeamDomain:   team1TeamDomain,
		TeamID:       team1TeamID,
		AppID:        "A1EXAMPLE01",
		EnterpriseID: enterprise1TeamID,
	}

	appTeam1Local := types.App{
		TeamDomain:   team1TeamDomain,
		AppID:        "A1EXAMPLE02",
		TeamID:       team1TeamID,
		IsDev:        true,
		UserID:       team1UserID,
		EnterpriseID: enterprise1TeamID,
	}

	err := clients.AppClient().SaveDeployed(ctx, appTeam1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appTeam1Local)
	require.NoError(t, err)

	// Save a hosted and local enterprise / org level app
	appEnterprise1Hosted := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE03",
		EnterpriseID: enterprise1TeamID,
	}

	appEnterprise1Local := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE04",
		IsDev:        true,
		EnterpriseID: enterprise1TeamID,
	}

	err = clients.AppClient().SaveDeployed(ctx, appEnterprise1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appEnterprise1Local)
	require.NoError(t, err)

	// We won't mock standard workspace apps and auths in this test, as that
	// Should be well covered by other tests in this file.

	// Test cases
	// Execute tests
	tests := []struct {
		desc string

		expected      SelectedApp
		expectedError *slackerror.Error

		env    AppEnvironmentType
		status AppInstallStatus

		appFlag  string
		teamFlag string

		authState []types.SlackAuth
		authError error

		selectTeamDomain  string
		selectEnvironment string
	}{
		{
			// Test description
			"When there is no existing workspace auth, returns empty SelectedApp with error",

			// Expected results
			SelectedApp{Auth: fakeAuthsByTeamDomain[team1TeamDomain], App: appTeam1Hosted},
			nil,

			// Env + app install status
			ShowHostedOnly,
			ShowInstalledAppsOnly,

			// flags
			"",
			"",

			// Auth state
			[]types.SlackAuth{},
			nil,

			// Selector preferences team 1
			team1TeamDomain,
			"Deployed",
		},
	}

	for _, test := range tests {

		// Set app flags
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag

		// Finally we mock the select prompt based on test specs
		clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectTeamDomain,
			Index:  0,
		}, nil)
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectEnvironment,
			Index:  0,
		}, nil)
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a local environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectEnvironment,
			Index:  0,
		}, nil)

		// App selector mock
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("app"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "deployed",
			Index:  1,
		}, nil)

		actualSelected1, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)
		actualSelected2, err2 := AppSelectPrompt(ctx, clients, test.status)

		actualSelected1.Auth.LastUpdated = time.Time{} // ignore time for this test
		actualSelected2.Auth.LastUpdated = time.Time{} // ignore time for this test

		require.NoError(t, err)
		require.NoError(t, err2)

		// There's a valid auth for this auth record
		_, err = clients.AppClient().GetDeployed(ctx, test.expected.Auth.TeamID)
		require.NoError(t, err)

		require.Equal(t, test.expected, actualSelected1)
		require.Equal(t, test.expected, actualSelected2)
	}
}

func TestPrompt_TeamAppSelectPrompt_EnterpriseWorkspaceApps_MissingWorkspaceAuth_HasOrgAuth(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)

	// Auths
	// Enterprise (org) Auth
	// Mock the enterprise that Team1 belongs to
	const enterprise1TeamDomain = "enterprise1"
	const enterprise1TeamID = "E1"
	const enterprise1UserID = "UE1"
	const enterprise1Token = "xoxp-xoxe-12345"

	authEnterprise1 := types.SlackAuth{
		TeamID:              enterprise1TeamID,
		TeamDomain:          enterprise1TeamDomain,
		Token:               enterprise1Token,
		IsEnterpriseInstall: true,
		UserID:              enterprise1UserID,
		EnterpriseID:        enterprise1TeamID,
	}

	// Set up team 1 to belong to enterprise 1
	authTeam1 := fakeAuthsByTeamDomain[team1TeamDomain]
	authTeam1.EnterpriseID = authEnterprise1.TeamID

	// For this test we want to make sure that no auth is found for team1, and a credentials not found
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound))
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, enterprise1TeamID).Return(authEnterprise1, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)

	// This test uses a single auth - the enterprise auth
	clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{
		authEnterprise1,
	}, nil)

	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Save apps
	// Save a hosted and local enterprise workspace-level app
	appTeam1Hosted := types.App{
		TeamDomain:   team1TeamDomain,
		TeamID:       team1TeamID,
		AppID:        "A1EXAMPLE01",
		EnterpriseID: enterprise1TeamID,
	}

	appTeam1Local := types.App{
		TeamDomain:   team1TeamDomain,
		AppID:        "A1EXAMPLE02",
		TeamID:       team1TeamID,
		IsDev:        true,
		UserID:       team1UserID,
		EnterpriseID: enterprise1TeamID,
	}

	err := clients.AppClient().SaveDeployed(ctx, appTeam1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appTeam1Local)
	require.NoError(t, err)

	// Save a hosted and local enterprise / org level app
	appEnterprise1Hosted := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE03",
		EnterpriseID: enterprise1TeamID,
	}

	appEnterprise1Local := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE04",
		IsDev:        true,
		EnterpriseID: enterprise1TeamID,
	}

	err = clients.AppClient().SaveDeployed(ctx, appEnterprise1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appEnterprise1Local)
	require.NoError(t, err)

	// Test cases
	// Execute tests
	tests := []struct {
		desc string

		expected      SelectedApp
		expectedError *slackerror.Error

		env    AppEnvironmentType
		status AppInstallStatus

		appFlag  string
		teamFlag string

		authState []types.SlackAuth
		authError error

		selectTeamDomain  string
		selectEnvironment string
	}{
		{
			// Test description
			"Has no existing workspace auth, but has enterprise/org auth, return App and enterprise Auth",

			// Expected results
			SelectedApp{
				App:  appTeam1Hosted,
				Auth: authEnterprise1,
			},
			nil,

			// Env + app install status
			ShowHostedOnly,
			ShowInstalledAppsOnly,

			// flags
			"",
			"",

			// Auth state
			[]types.SlackAuth{
				authEnterprise1,
			},
			nil,

			// Selector preferences team 1
			team1TeamDomain,
			"Deployed",
		},
	}

	for _, test := range tests {

		// Set app flags
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag

		// Finally we mock the select prompt based on test specs

		// Workspace selector mocks
		clientsMock.IO.On(SelectPrompt, mock.Anything, appInstallPromptNew, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectTeamDomain,
			Index:  1, // should select team1
		}, nil)
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose a deployed environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectEnvironment,
			Index:  1, // should select team1
		}, nil)

		actualSelected1, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)

		if test.expectedError == nil {
			require.NoError(t, err)

			// There's a valid auth for this auth record
			_, err := clients.AppClient().GetDeployed(ctx, test.expected.Auth.TeamID)
			require.NoError(t, err)

			require.Equal(t, test.expected, actualSelected1)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.expectedError.Code, err.(*slackerror.Error).Code, test.desc)

			require.Equal(t, test.expected, actualSelected1)
		}
	}
}

func TestPrompt_AppSelectPrompt_EnterpriseWorkspaceApps_MissingWorkspaceAuth_HasOrgAuth(t *testing.T) {
	// Set up mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.API.On(GetAppStatus, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.GetAppStatusResult{}, nil)

	// Auths
	// Enterprise (org) Auth
	// Mock the enterprise that Team1 belongs to
	const enterprise1TeamDomain = "enterprise1"
	const enterprise1TeamID = "E1"
	const enterprise1UserID = "UE1"
	const enterprise1Token = "xoxp-xoxe-12345"

	authEnterprise1 := types.SlackAuth{
		TeamID:              enterprise1TeamID,
		TeamDomain:          enterprise1TeamDomain,
		Token:               enterprise1Token,
		IsEnterpriseInstall: true,
		UserID:              enterprise1UserID,
		EnterpriseID:        enterprise1TeamID,
	}

	// Set up team 1 to belong to enterprise 1
	authTeam1 := fakeAuthsByTeamDomain[team1TeamDomain]
	authTeam1.EnterpriseID = authEnterprise1.TeamID

	// For this test we want to make sure that no auth is found for team1, and a credentials not found
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, team1TeamID).Return(types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound))
	clientsMock.Auth.On(AuthWithTeamID, mock.Anything, enterprise1TeamID).Return(authEnterprise1, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)

	// This test uses a single auth - the enterprise auth
	clientsMock.Auth.On(Auths, mock.Anything).Return([]types.SlackAuth{
		authEnterprise1,
	}, nil)

	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Save apps
	// Save a hosted and local enterprise workspace-level app
	appTeam1Hosted := types.App{
		TeamDomain:   team1TeamDomain,
		TeamID:       team1TeamID,
		AppID:        "A1EXAMPLE01",
		EnterpriseID: enterprise1TeamID,
	}

	appTeam1Local := types.App{
		TeamDomain:   team1TeamDomain,
		AppID:        "A1EXAMPLE02",
		TeamID:       team1TeamID,
		IsDev:        true,
		UserID:       team1UserID,
		EnterpriseID: enterprise1TeamID,
	}

	err := clients.AppClient().SaveDeployed(ctx, appTeam1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appTeam1Local)
	require.NoError(t, err)

	// Save a hosted and local enterprise / org level app
	appEnterprise1Hosted := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE03",
		EnterpriseID: enterprise1TeamID,
	}

	appEnterprise1Local := types.App{
		TeamDomain:   enterprise1TeamDomain,
		TeamID:       enterprise1TeamID,
		AppID:        "A1EXAMPLE04",
		IsDev:        true,
		EnterpriseID: enterprise1TeamID,
	}

	err = clients.AppClient().SaveDeployed(ctx, appEnterprise1Hosted)
	require.NoError(t, err)
	err = clients.AppClient().SaveLocal(ctx, appEnterprise1Local)
	require.NoError(t, err)

	// Test cases
	// Execute tests
	tests := []struct {
		desc string

		expected      SelectedApp
		expectedError *slackerror.Error

		env    AppEnvironmentType
		status AppInstallStatus

		appFlag  string
		teamFlag string

		authState []types.SlackAuth
		authError error

		selectTeamDomain  string
		selectEnvironment string
	}{
		{
			// Test description
			"Has no existing workspace auth, but has enterprise/org auth, return App and enterprise Auth",

			// Expected results
			SelectedApp{
				App:  appTeam1Hosted,
				Auth: authEnterprise1,
			},
			nil,

			// Env + app install status
			ShowHostedOnly,
			ShowInstalledAppsOnly,

			// flags
			"",
			"",

			// Auth state
			[]types.SlackAuth{
				authEnterprise1,
			},
			nil,

			// Selector preferences team 1
			team1TeamDomain,
			"Deployed",
		},
	}

	for _, test := range tests {

		// Set app flags
		clients.Config.AppFlag = test.appFlag
		clients.Config.TeamFlag = test.teamFlag

		// Finally we mock the select prompt based on test specs
		// App selector mocks
		clientsMock.IO.On(SelectPrompt, mock.Anything, "Choose an app environment", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("app"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "deployed",
			Index:  1, // should select team1
		}, nil)

		clientsMock.IO.On(SelectPrompt, mock.Anything, SelectTeamPrompt, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("team"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: test.selectTeamDomain,
			Index:  1,
		}, nil)

		// actualSelected1, err := TeamAppSelectPrompt(ctx, clients, test.env, test.status)
		actualSelected2, err := AppSelectPrompt(ctx, clients, test.status)

		if test.expectedError == nil {
			require.NoError(t, err)

			// There's a valid auth for this auth record
			_, err := clients.AppClient().GetDeployed(ctx, test.expected.Auth.TeamID)
			require.NoError(t, err)

			require.Equal(t, test.expected, actualSelected2)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.expectedError.Code, err.(*slackerror.Error).Code, test.desc)
			require.Equal(t, test.expected, actualSelected2)
		}
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

// mockReauthentication is a test helper function to cut down on code duplication
func mockReauthentication(clientsMock *shared.ClientsMock) {
	// Default mocks
	clientsMock.Os.AddDefaultMocks()
	clientsMock.API.AddDefaultMocks()
	// Enable interactivity
	clientsMock.IO.On("IsTTY").Return(true)
	clientsMock.IO.AddDefaultMocks()

	// Mock invalid auth response
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, fmt.Errorf(slackerror.ErrInvalidAuth))
	clientsMock.Auth.On("FilterKnownAuthErrors", mock.Anything, mock.Anything).Return(true, nil)
	// Mocks for reauthentication
	clientsMock.API.On("GenerateAuthTicket", mock.Anything, mock.Anything, mock.Anything).Return(api.GenerateAuthTicketResult{}, nil)
	clientsMock.IO.On("InputPrompt", mock.Anything, "Enter challenge code", iostreams.InputPromptConfig{
		Required: true,
	}).Return("challengeCode", nil)
	clientsMock.API.On("ExchangeAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.ExchangeAuthTicketResult{Token: fakeAuthsByTeamDomain[team1TeamDomain].Token, TeamDomain: team1TeamDomain,
		TeamID: team1TeamID, UserID: "U1"}, nil)
	clientsMock.Auth.On("IsAPIHostSlackProd", mock.Anything).Return(true)
	clientsMock.Auth.On("SetAuth", mock.Anything, mock.Anything).Return(types.SlackAuth{}, "", nil)
	clientsMock.Auth.On("SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
}
