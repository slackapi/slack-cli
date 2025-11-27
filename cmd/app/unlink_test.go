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

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestAppsUnlinkCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"happy path; unlink the deployed app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonUnlinkMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				unlinkAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeDeployedApp.TeamDomain},
					App:  fakeDeployedApp,
				}, nil)
				// Mock unlink confirmation prompt
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to unlink this app?", mock.Anything).Return(true, nil)
				// Mock AppClient calls
				appClientMock := &app.AppClientMock{}
				appClientMock.On("Remove", mock.Anything, mock.Anything).Return(fakeDeployedApp, nil)
				appClientMock.On("CleanUp").Return()
				cf.AppClient().AppClientInterface = appClientMock
			},
			ExpectedStdoutOutputs: []string{
				fmt.Sprintf("Removed app %s from project", fakeDeployedApp.AppID),
				fmt.Sprintf("Team: %s", fakeDeployedApp.TeamDomain),
			},
		},
		"happy path; unlink the local app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonUnlinkMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				unlinkAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeLocalApp.TeamDomain},
					App:  fakeLocalApp,
				}, nil)
				// Mock unlink confirmation prompt
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to unlink this app?", mock.Anything).Return(true, nil)
				// Mock AppClient calls
				appClientMock := &app.AppClientMock{}
				appClientMock.On("Remove", mock.Anything, mock.Anything).Return(fakeLocalApp, nil)
				appClientMock.On("CleanUp").Return()
				cf.AppClient().AppClientInterface = appClientMock
			},
			ExpectedStdoutOutputs: []string{
				fmt.Sprintf("Removed app %s from project", fakeLocalApp.AppID),
				fmt.Sprintf("Team: %s", fakeLocalApp.TeamDomain),
			},
		},
		"sad path; unlinking the deployed app fails": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonUnlinkMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				unlinkAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeDeployedApp.TeamDomain},
					App:  fakeDeployedApp,
				}, nil)
				// Mock unlink confirmation prompt
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to unlink this app?", mock.Anything).Return(true, nil)
				// Mock AppClient calls - return error
				appClientMock := &app.AppClientMock{}
				appClientMock.On("Remove", mock.Anything, mock.Anything).Return(types.App{}, fmt.Errorf("failed to remove app from project"))
				cf.AppClient().AppClientInterface = appClientMock
			},
			ExpectedError: fmt.Errorf("failed to remove app from project"),
		},
		"user cancels unlink": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonUnlinkMocks(t, cf, cm)
				// Mock App Selection
				appSelectMock := prompts.NewAppSelectMock()
				unlinkAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{
					Auth: types.SlackAuth{TeamDomain: fakeDeployedApp.TeamDomain},
					App:  fakeDeployedApp,
				}, nil)
				// Mock unlink confirmation prompt - user says no
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to unlink this app?", mock.Anything).Return(false, nil)
			},
			ExpectedStdoutOutputs: []string{
				"Your app will not be unlinked",
			},
		},
		"errors if app selection fails": {
			CmdArgs:       []string{},
			ExpectedError: fmt.Errorf("failed to select app"),
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				prepareCommonUnlinkMocks(t, cf, cm)
				appSelectMock := prompts.NewAppSelectMock()
				unlinkAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps).Return(prompts.SelectedApp{}, fmt.Errorf("failed to select app"))
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewUnlinkCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func prepareCommonUnlinkMocks(t *testing.T, cf *shared.ClientFactory, cm *shared.ClientsMock) {
	cm.AddDefaultMocks()
}
