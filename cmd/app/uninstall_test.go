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
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var fakeAppID = "A1234"
var fakeAppTeamID = "T1234"
var fakeAppUserID = "U1234"

var fakeApp = types.App{
	TeamDomain: "test",
	AppID:      fakeAppID,
	TeamID:     fakeAppTeamID,
	UserID:     fakeAppUserID,
}

var selectedProdApp = prompts.SelectedApp{Auth: types.SlackAuth{TeamDomain: "team1234"}, App: types.App{AppID: fakeAppID, TeamID: fakeAppTeamID}}

func TestAppsUninstall(t *testing.T) {

	testutil.TableTestCommand(t, testutil.CommandTests{
		"Successfully uninstall": {
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				prepareCommonUninstallMocks(clients, clientsMock)
				clientsMock.ApiInterface.On("UninstallApp", mock.Anything, mock.Anything, fakeAppID, fakeAppTeamID).
					Return(nil).Once()
			},
			ExpectedStdoutOutputs: []string{
				fmt.Sprintf(`Uninstalled the app "%s" from "%s"`, fakeAppID, fakeApp.TeamDomain),
			},
		},
		"Successfully uninstall with a get-manifest hook error": {
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				prepareCommonUninstallMocks(clients, clientsMock)
				clientsMock.ApiInterface.On("UninstallApp", mock.Anything, mock.Anything, fakeAppID, fakeAppTeamID).
					Return(nil).Once()
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything).
					Return(types.SlackYaml{}, slackerror.New(slackerror.ErrSDKHookNotFound))
				clients.AppClient().Manifest = manifestMock
			},
			ExpectedStdoutOutputs: []string{
				fmt.Sprintf(`Uninstalled the app "%s" from "%s"`, fakeAppID, fakeApp.TeamDomain),
			},
		},
		"Fail to uninstall due to API error": {
			ExpectedError: slackerror.New("something went wrong"),
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				prepareCommonUninstallMocks(clients, clientsMock)
				clientsMock.ApiInterface.On("UninstallApp", mock.Anything, mock.Anything, fakeAppID, fakeAppTeamID).
					Return(slackerror.New("something went wrong")).Once()
			},
		},
		"errors if authentication for the team is missing": {
			CmdArgs:       []string{},
			ExpectedError: slackerror.New(slackerror.ErrCredentialsNotFound),
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonUninstallMocks(cf, cm)
				appSelectMock := prompts.NewAppSelectMock()
				uninstallAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(prompts.SelectedApp{App: fakeApp}, nil)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewUninstallCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func prepareCommonUninstallMocks(clients *shared.ClientFactory, clientsMock *shared.ClientsMock) *shared.ClientFactory {

	// Mock App Selection
	appSelectMock := prompts.NewAppSelectMock()
	uninstallAppSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt").Return(selectedProdApp, nil)

	// Mock API calls
	clientsMock.AuthInterface.On("ResolveApiHost", mock.Anything, mock.Anything, mock.Anything).
		Return("api host")
	clientsMock.AuthInterface.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).
		Return("logstash host")

	clientsMock.ApiInterface.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
		TeamName: &fakeApp.TeamDomain,
		TeamID:   &fakeApp.TeamID,
	}, nil)
	clientsMock.AddDefaultMocks()

	// Mock prompts
	clientsMock.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to uninstall?", mock.Anything).Return(true, nil)

	// Mock AppClient calls
	appClientMock := &app.AppClientMock{}
	appClientMock.On("GetDeployed", mock.Anything, mock.Anything).Return(fakeApp, nil)
	appClientMock.On("SaveDeployed", mock.Anything, mock.Anything).Return(nil)

	clients.AppClient().AppClientInterface = appClientMock

	err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
	if err != nil {
		panic("error setting up test; cant write apps.json")
	}

	return clients
}
