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
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockLinkAppID1 = "A001"
var mockLinkAppID2 = "A002"

var mockLinkSlackAuth1 = types.SlackAuth{
	Token:        "xoxp-example1",
	TeamDomain:   "team1",
	TeamID:       "T001",
	EnterpriseID: "E001",
	UserID:       "U001",
}

var mockLinkSlackAuth2 = types.SlackAuth{
	Token:      "xoxp-example2",
	TeamDomain: "team2",
	TeamID:     "T002",
	UserID:     "U002",
}

func Test_Apps_Link(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"saves information about the provided deployed app": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth2,
					mockLinkSlackAuth1,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: mockLinkSlackAuth1.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID1, nil)
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
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth1.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth1.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			CmdArgs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedApp := types.App{
					AppID:        mockLinkAppID1,
					TeamDomain:   mockLinkSlackAuth1.TeamDomain,
					TeamID:       mockLinkSlackAuth1.TeamID,
					EnterpriseID: mockLinkSlackAuth1.EnterpriseID,
				}
				actualApp, err := cm.AppClient.GetDeployed(
					context.Background(),
					mockLinkSlackAuth1.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
			},
		},
		"saves information about the provided local app": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth2,
					mockLinkSlackAuth1,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Index:  1,
					Option: mockLinkSlackAuth2.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID2, nil)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "local",
				}, nil)
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth2.Token,
					[]string{mockLinkAppID2},
					mockLinkSlackAuth2.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			CmdArgs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedApp := types.App{
					AppID:      mockLinkAppID2,
					TeamDomain: mockLinkSlackAuth2.TeamDomain,
					TeamID:     mockLinkSlackAuth2.TeamID,
					IsDev:      true,
					UserID:     mockLinkSlackAuth2.UserID,
				}
				actualApp, err := cm.AppClient.GetLocal(
					context.Background(),
					mockLinkSlackAuth2.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
			},
		},
		"avoids overwriting an app saved in json without confirmation": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth1,
					mockLinkSlackAuth2,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				existingApp := types.App{
					AppID:        mockLinkAppID1,
					TeamDomain:   mockLinkSlackAuth1.TeamDomain,
					TeamID:       mockLinkSlackAuth1.TeamID,
					EnterpriseID: mockLinkSlackAuth1.EnterpriseID,
				}
				err := cm.AppClient.SaveDeployed(ctx, existingApp)
				require.NoError(t, err)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Index:  0,
					Option: mockLinkSlackAuth1.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID2, nil)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "deployed",
				}, nil)
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth1.Token,
					[]string{mockLinkAppID2},
					mockLinkSlackAuth1.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			CmdArgs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedApp := types.App{
					AppID:        mockLinkAppID1,
					TeamDomain:   mockLinkSlackAuth1.TeamDomain,
					TeamID:       mockLinkSlackAuth1.TeamID,
					EnterpriseID: mockLinkSlackAuth1.EnterpriseID,
				}
				actualApp, err := cm.AppClient.GetDeployed(
					context.Background(),
					mockLinkSlackAuth1.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
			},
			ExpectedError: slackerror.New(slackerror.ErrAppFound).
				WithMessage("A saved app was found and cannot be overwritten").
				WithRemediation("Remove the app from this project or try again with --force"),
		},
		"avoids overwriting a matching app id for the team without confirmation": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth1,
					mockLinkSlackAuth2,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				existingApp := types.App{
					AppID:        mockLinkAppID1,
					TeamDomain:   mockLinkSlackAuth1.TeamDomain,
					TeamID:       mockLinkSlackAuth1.TeamID,
					EnterpriseID: mockLinkSlackAuth1.EnterpriseID,
				}
				err := cm.AppClient.SaveDeployed(ctx, existingApp)
				require.NoError(t, err)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Index:  0,
					Option: mockLinkSlackAuth1.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID1, nil)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "local",
				}, nil)
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth1.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth1.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			CmdArgs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedApp := types.App{
					AppID:        mockLinkAppID1,
					TeamDomain:   mockLinkSlackAuth1.TeamDomain,
					TeamID:       mockLinkSlackAuth1.TeamID,
					EnterpriseID: mockLinkSlackAuth1.EnterpriseID,
				}
				actualApp, err := cm.AppClient.GetDeployed(
					context.Background(),
					mockLinkSlackAuth1.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
				unsavedApp, err := cm.AppClient.GetLocal(
					context.Background(),
					mockLinkSlackAuth1.TeamID,
				)
				require.NoError(t, err)
				assert.True(t, unsavedApp.IsNew())
			},
			ExpectedError: slackerror.New(slackerror.ErrAppFound).
				WithMessage("A saved app was found and cannot be overwritten").
				WithRemediation("Remove the app from this project or try again with --force"),
		},
		"completes overwriting an app saved in json with confirmation": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth1,
					mockLinkSlackAuth2,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				existingApp := types.App{
					AppID:      mockLinkAppID2,
					TeamDomain: mockLinkSlackAuth2.TeamDomain,
					TeamID:     mockLinkSlackAuth2.TeamID,
					IsDev:      true,
					UserID:     mockLinkSlackAuth2.UserID,
				}
				err := cm.AppClient.SaveLocal(ctx, existingApp)
				require.NoError(t, err)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Index:  1,
					Option: mockLinkSlackAuth2.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID1, nil)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "local",
				}, nil)
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth2.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth2.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			CmdArgs: []string{"--force"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedApp := types.App{
					AppID:      mockLinkAppID1,
					TeamDomain: mockLinkSlackAuth2.TeamDomain,
					TeamID:     mockLinkSlackAuth2.TeamID,
					IsDev:      true,
					UserID:     mockLinkSlackAuth2.UserID,
				}
				actualApp, err := cm.AppClient.GetLocal(
					context.Background(),
					mockLinkSlackAuth2.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
			},
		},
		"refuses to write an app with app id not existing upstream": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth1,
					mockLinkSlackAuth2,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Index:  0,
					Option: mockLinkSlackAuth1.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID1, nil)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "Deployed",
				}, nil)
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth1.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth1.TeamID,
				).Return(
					api.GetAppStatusResult{},
					slackerror.New(slackerror.ErrAppNotFound),
				)
			},
			CmdArgs:       []string{},
			ExpectedError: slackerror.New(slackerror.ErrAppNotFound),
		},
		"accept manifest source prompt and saves information about the provided deployed app": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth2,
					mockLinkSlackAuth1,
				}, nil)
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				// Set manifest source to project to trigger confirmation prompt
				if err := cm.Config.ProjectConfig.SetManifestSource(t.Context(), config.MANIFEST_SOURCE_LOCAL); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to set the manifest source in the memory-based file system: %s", err))
				}
				// Accept manifest source confirmation prompt
				cm.IO.On("ConfirmPrompt",
					mock.Anything,
					LinkAppManifestSourceConfirmPromptText,
					mock.Anything,
				).Return(true, nil)
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: mockLinkSlackAuth1.TeamDomain,
				}, nil)
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID1, nil)
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
				cm.ApiInterface.On(
					"GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth1.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth1.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			CmdArgs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedApp := types.App{
					AppID:        mockLinkAppID1,
					TeamDomain:   mockLinkSlackAuth1.TeamDomain,
					TeamID:       mockLinkSlackAuth1.TeamID,
					EnterpriseID: mockLinkSlackAuth1.EnterpriseID,
				}
				actualApp, err := cm.AppClient.GetDeployed(
					context.Background(),
					mockLinkSlackAuth1.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
				// Assert manifest confirmation prompt accepted
				cm.IO.AssertCalled(t, "ConfirmPrompt",
					mock.Anything,
					LinkAppManifestSourceConfirmPromptText,
					mock.Anything,
				)
			},
		},
		"decline manifest source prompt should not link app": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				setupAppLinkCommandMocks(t, cm, cf)
				// Set manifest source to project to trigger confirmation prompt
				if err := cm.Config.ProjectConfig.SetManifestSource(t.Context(), config.MANIFEST_SOURCE_LOCAL); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to set the manifest source in the memory-based file system: %s", err))
				}
				// Decline manifest source confirmation prompt
				cm.IO.On("ConfirmPrompt",
					mock.Anything,
					LinkAppManifestSourceConfirmPromptText,
					mock.Anything,
				).Return(false, nil)
			},
			CmdArgs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert manifest confirmation prompt accepted
				cm.IO.AssertCalled(t, "ConfirmPrompt",
					mock.Anything,
					LinkAppManifestSourceConfirmPromptText,
					mock.Anything,
				)

				// Assert no apps saved
				apps, _, err := cm.AppClient.GetDeployedAll(ctx)
				require.NoError(t, err)
				require.Len(t, apps, 0)

				apps, err = cm.AppClient.GetLocalAll(ctx)
				require.NoError(t, err)
				require.Len(t, apps, 0)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		clients.SDKConfig.WorkingDirectory = "."
		return NewLinkCommand(clients)
	})
}

func Test_Apps_LinkAppHeaderSection(t *testing.T) {
	tests := map[string]struct {
		shouldConfirm     bool
		expectedOutputs   []string
		unexpectedOutputs []string
	}{
		"When shouldConfirm is false": {
			shouldConfirm: false,
			expectedOutputs: []string{
				"Add an existing app created on app settings",
				"Find your existing apps at: https://api.slack.com/apps",
			},
			unexpectedOutputs: []string{
				"Manually add apps later with",
			},
		},
		"When shouldConfirm is true": {
			shouldConfirm: true,
			expectedOutputs: []string{
				"Add an existing app created on app settings",
				"Find your existing apps at: https://api.slack.com/apps",
				"Manually add apps later with",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup parameters for test
			ctx := context.Background()

			// Create mocks
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()

			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Run the test
			LinkAppHeaderSection(ctx, clients, tt.shouldConfirm)

			// Assertions
			output := clientsMock.GetCombinedOutput()
			for _, expectedOutput := range tt.expectedOutputs {
				require.Contains(t, output, expectedOutput)
			}
			for _, unexpectedOutput := range tt.unexpectedOutputs {
				require.NotContains(t, output, unexpectedOutput)
			}
		})
	}
}

func setupAppLinkCommandMocks(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	ctx := t.Context()
	projectDirPath := slackdeps.MockWorkingDirectory
	cm.Os.On("Getwd").Return(projectDirPath, nil)

	// Setup a legit slack project, so that the config.json file can be read
	if _, err := config.CreateProjectConfigDir(ctx, cm.Fs, projectDirPath); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to create the project config directory in the memory-based file system: %s", err))
	}

	if _, err := config.CreateProjectHooksJSONFile(cm.Fs, projectDirPath, []byte("{}")); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to create the hooks file in the memory-based file system: %s", err))
	}

	if err := cm.Config.ProjectConfig.SetManifestSource(ctx, config.MANIFEST_SOURCE_REMOTE); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to set the manifest source in the memory-based file system: %s", err))
	}
}
