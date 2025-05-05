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
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock team information
const (
	team1TeamDomain = "team1"
	team1TeamID     = "T1"
	team1UserID     = "U1"
	team1Token      = "xoxe.xoxp-1-token"

	team2TeamDomain = "team2"
	team2TeamID     = "T2"
	team2UserID     = "U2"
	team2Token      = "xoxe.xoxp-2-token"

	enterprise1TeamDomain = "enterprise1"
	enterprise1TeamID     = "E1"
	enterprise1UserID     = "UE1"
	enterprise1Token      = "xoxp-xoxe-12345"
)

// Mock auths
var authTeam1 = types.SlackAuth{
	TeamDomain:   team1TeamDomain,
	TeamID:       team1TeamID,
	UserID:       team1UserID,
	Token:        team1Token,
	EnterpriseID: enterprise1TeamID,
}
var authTeam2 = types.SlackAuth{
	TeamDomain: team2TeamDomain,
	TeamID:     team2TeamID,
	UserID:     team2UserID,
	Token:      team2Token,
}
var authEnterprise1 = types.SlackAuth{
	TeamID:              enterprise1TeamID,
	TeamDomain:          enterprise1TeamDomain,
	Token:               enterprise1Token,
	IsEnterpriseInstall: true,
	UserID:              enterprise1UserID,
	EnterpriseID:        enterprise1TeamID,
}

// Mock app information
const (
	team1AppID = "A1"
	team2AppID = "A2"
)

// Mock apps
var team1DeployedApp = types.App{
	AppID:        team1AppID,
	TeamID:       team1TeamID,
	TeamDomain:   team1TeamDomain,
	EnterpriseID: enterprise1TeamID,
	IsDev:        false,
}
var team2LocalApp = types.App{
	AppID:      team2AppID,
	TeamID:     team2TeamID,
	TeamDomain: team2TeamDomain,
	IsDev:      true,
	UserID:     team2UserID,
}

func TestAppsList_FetchInstallStates_NoAuthsShouldReturnUnknownState(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	apps, err := FetchAppInstallStates(ctx, clients, []types.App{team1DeployedApp, team2LocalApp})
	require.NoError(t, err)
	clientsMock.API.AssertNotCalled(t, "GetAppStatus")

	require.Contains(t, apps, team1DeployedApp)
	require.Contains(t, apps, team2LocalApp)
	for _, app := range apps {
		require.Equal(t, app.InstallStatus, types.AppInstallationStatusUnknown)
	}
}

func TestAppsList_FetchInstallStates_HasEnterpriseApp_HasEnterpriseAuth(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	clientsMock.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
		authEnterprise1,
	}, nil)
	clientsMock.AddDefaultMocks()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Return installed true
	clientsMock.API.On("GetAppStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: team1AppID, Installed: true}},
		}, nil)

	require.True(t, team1DeployedApp.IsEnterpriseWorkspaceApp())

	// Should successfully fetchAppInstallStates when the auth is enterprise and the app is enterprise workspace app
	appsWithStatus, _ := FetchAppInstallStates(ctx, clients, []types.App{team1DeployedApp})
	require.NotEmpty(t, appsWithStatus)
	require.Equal(t, team1DeployedApp.AppID, appsWithStatus[0].AppID)
}

func TestAppsList_FetchInstallStates_TokenFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	clientsMock.AuthInterface.On("Auths", mock.Anything).
		Return([]types.SlackAuth{authTeam1}, nil)
	clientsMock.AuthInterface.On("AuthWithToken", mock.Anything, team2Token).
		Return(authTeam2, nil)

	clientsMock.API.On("GetAppStatus", mock.Anything, team1Token, []string{team1DeployedApp.AppID}, team1TeamID).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: team1DeployedApp.AppID, Installed: true}},
		}, nil)
	clientsMock.API.On("GetAppStatus", mock.Anything, team2Token, []string{team2LocalApp.AppID}, team2TeamID).Return(
		api.GetAppStatusResult{
			Apps: []api.AppStatusResultAppInfo{{AppID: team2LocalApp.AppID, Installed: false}},
		}, nil)

	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clients.Config.TokenFlag = team2Token

	apps, err := FetchAppInstallStates(ctx, clients, []types.App{team1DeployedApp, team2LocalApp})
	require.NoError(t, err)
	require.Len(t, apps, 2)
	clientsMock.API.AssertNumberOfCalls(t, "GetAppStatus", 2)

	for _, app := range apps {
		switch app.AppID {
		case team1AppID:
			require.Equal(t, app.InstallStatus, types.AppStatusInstalled)
		case team2AppID:
			require.Equal(t, app.InstallStatus, types.AppStatusUninstalled)
		}
	}
}

func TestAppsList_FetchInstallStates_InvalidTokenFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AuthInterface.On("Auths", mock.Anything).
		Return([]types.SlackAuth{}, nil)
	clientsMock.AuthInterface.On("AuthWithToken", mock.Anything, mock.Anything).
		Return(types.SlackAuth{}, slackerror.New(slackerror.ErrHTTPRequestFailed))
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	clients.Config.TokenFlag = "xoxp-invalid"

	apps, err := FetchAppInstallStates(ctx, clients, []types.App{team1DeployedApp, team2LocalApp})
	if assert.Error(t, err) {
		assert.Equal(t, slackerror.New(slackerror.ErrHTTPRequestFailed), err)
	}
	assert.Equal(t, []types.App{}, apps)
}
