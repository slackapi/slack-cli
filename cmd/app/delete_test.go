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

var fakeDeployedApp = types.App{
	TeamDomain: "test",
	AppID:      "A1",
	TeamID:     fakeAppTeamID,
	UserID:     fakeAppUserID,
	IsDev:      false,
}

var fakeLocalApp = types.App{
	TeamDomain: "test",
	AppID:      "A2",
	TeamID:     fakeAppTeamID,
	UserID:     fakeAppUserID,
	IsDev:      true,
}

func TestAppsDeleteCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"happy path; delete the deployed app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonDeleteMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				deleteAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeDeployedApp.TeamDomain},
					App:  fakeDeployedApp,
				}, nil)
				// Mock delete confirmation prompt
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to delete the app?", mock.Anything).Return(true, nil)
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					TeamName: &fakeDeployedApp.TeamDomain,
					TeamID:   &fakeDeployedApp.TeamID,
				}, nil)
				// Mock delete API call
				cm.API.On("DeleteApp", mock.Anything, mock.Anything, fakeDeployedApp.AppID).Return(nil)
				// Mock AppClient calls
				appClientMock := &app.AppClientMock{}
				appClientMock.On("GetDeployed", mock.Anything, mock.Anything).Return(fakeDeployedApp, nil)
				appClientMock.On("SaveDeployed", mock.Anything, mock.Anything).Return(nil)
				appClientMock.On("Remove", mock.Anything, mock.Anything).Return(fakeDeployedApp, nil)
				cf.AppClient().AppClientInterface = appClientMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DeleteApp", mock.Anything, mock.Anything, fakeDeployedApp.AppID)
			},
			ExpectedStdoutOutputs: []string{
				fmt.Sprintf(`Uninstalled the app "%s" from "%s"`, fakeDeployedApp.AppID, fakeDeployedApp.TeamDomain),
				fmt.Sprintf(`Deleted the app manifest for "%s" from "%s"`, fakeDeployedApp.AppID, fakeDeployedApp.TeamDomain),
			},
		},
		"happy path; delete the local app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonDeleteMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				deleteAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeLocalApp.TeamDomain},
					App:  fakeLocalApp,
				}, nil)
				// Mock delete confirmation prompt
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to delete the app?", mock.Anything).Return(true, nil)
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					TeamName: &fakeLocalApp.TeamDomain,
					TeamID:   &fakeLocalApp.TeamID,
				}, nil)
				// Mock delete API call
				cm.API.On("DeleteApp", mock.Anything, mock.Anything, fakeLocalApp.AppID).Return(nil)
				// Mock AppClient calls
				appClientMock := &app.AppClientMock{}
				appClientMock.On("GetLocal", mock.Anything, mock.Anything).Return(fakeLocalApp, nil)
				appClientMock.On("Remove", mock.Anything, mock.Anything).Return(fakeLocalApp, nil)
				cf.AppClient().AppClientInterface = appClientMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DeleteApp", mock.Anything, mock.Anything, fakeLocalApp.AppID)
			},
			ExpectedStdoutOutputs: []string{
				fmt.Sprintf(`Uninstalled the app "%s" from "%s"`, fakeLocalApp.AppID, fakeLocalApp.TeamDomain),
				fmt.Sprintf(`Deleted the app manifest for "%s" from "%s"`, fakeLocalApp.AppID, fakeLocalApp.TeamDomain),
			},
		},
		"sad path; deleting the deployed app fails": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonDeleteMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				deleteAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeDeployedApp.TeamDomain},
					App:  fakeDeployedApp,
				}, nil)
				// Mock delete confirmation prompt
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to delete the app?", mock.Anything).Return(true, nil)
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					TeamName: &fakeDeployedApp.TeamDomain,
					TeamID:   &fakeDeployedApp.TeamID,
				}, nil)
				// Mock delete API call
				cm.API.On("DeleteApp", mock.Anything, mock.Anything, fakeDeployedApp.AppID).Return(fmt.Errorf("something went terribly wrong"))
				// Mock AppClient calls
				appClientMock := &app.AppClientMock{}
				appClientMock.On("GetDeployed", mock.Anything, mock.Anything).Return(fakeDeployedApp, nil)
				cf.AppClient().AppClientInterface = appClientMock
			},
			ExpectedError: fmt.Errorf("something went terribly wrong"),
		},
		"errors if authentication for the team is missing": {
			CmdArgs:       []string{},
			ExpectedError: slackerror.New(slackerror.ErrCredentialsNotFound),
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonDeleteMocks(t, cf, cm)
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
				appSelectMock := prompts.NewAppSelectMock()
				deleteAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{App: fakeDeployedApp}, nil)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewDeleteCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func prepareCommonDeleteMocks(t *testing.T, cf *shared.ClientFactory, cm *shared.ClientsMock) {
	cm.AddDefaultMocks()

	cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).
		Return("api host")
	cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).
		Return("logstash host")

	// Mock list command
	listPkgMock := new(ListPkgMock)
	listFunc = listPkgMock.List
	listPkgMock.On("List").Return(nil)
}
