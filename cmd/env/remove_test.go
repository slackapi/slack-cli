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

package env

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Env_RemoveCommandPreRun(t *testing.T) {
	tests := map[string]struct {
		mockFlagForce        bool
		mockManifestResponse types.SlackYaml
		mockManifestError    error
		mockManifestSource   config.ManifestSource
		mockWorkingDirectory string
		expectedError        error
	}{
		"continues if the application is hosted on slack": {
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.SlackHosted,
					},
				},
			},
			mockManifestError:    nil,
			mockManifestSource:   config.ManifestSourceLocal,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        nil,
		},
		"errors if the application is not hosted on slack": {
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.Remote,
					},
				},
			},
			mockManifestError:    nil,
			mockManifestSource:   config.ManifestSourceLocal,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrAppNotHosted),
		},
		"continues if the force flag is used in a project": {
			mockFlagForce:        true,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        nil,
		},
		"errors if the project manifest cannot be retrieved": {
			mockManifestResponse: types.SlackYaml{},
			mockManifestError:    slackerror.New(slackerror.ErrSDKHookInvocationFailed),
			mockManifestSource:   config.ManifestSourceLocal,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrSDKHookInvocationFailed),
		},
		"errors if the command is not run in a project": {
			mockManifestResponse: types.SlackYaml{},
			mockManifestError:    slackerror.New(slackerror.ErrSDKHookNotFound),
			mockWorkingDirectory: "",
			expectedError:        slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
		"errors if the manifest source is set to remote": {
			mockManifestSource:   config.ManifestSourceRemote,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrAppNotHosted),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On(
				"GetManifestLocal",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestResponse,
				tt.mockManifestError,
			)
			clientsMock.AppClient.Manifest = manifestMock
			projectConfigMock := config.NewProjectConfigMock()
			projectConfigMock.On(
				"GetManifestSource",
				mock.Anything,
			).Return(
				tt.mockManifestSource,
				nil,
			)
			clientsMock.Config.ProjectConfig = projectConfigMock
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.Config.ForceFlag = tt.mockFlagForce
				cf.SDKConfig.WorkingDirectory = tt.mockWorkingDirectory
			})
			cmd := NewEnvRemoveCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tt.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tt.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_Env_RemoveCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"remove a variable using arguments": {
			CmdArgs: []string{"ENV_NAME"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On(
					"ListVariables",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]string{"example"},
					nil,
				)
				cm.API.On(
					"RemoveVariable",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
				)
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"RemoveVariable",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					"ENV_NAME",
				)
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvRemoveSuccess,
					mock.Anything,
				)
			},
		},
		"remove a variable using prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On(
					"ListVariables",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]string{"example"},
					nil,
				)
				cm.API.On(
					"RemoveVariable",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
				)
				cm.IO.On(
					"SelectPrompt",
					mock.Anything,
					"Select a variable to remove",
					mock.Anything,
					iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
						Flag: cm.Config.Flags.Lookup("name"),
					}),
				).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "SELECTED_ENV_VAR",
						Index:  0,
					},
					nil,
				)
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"ListVariables",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				cm.API.AssertCalled(
					t,
					"RemoveVariable",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					"SELECTED_ENV_VAR",
				)
			},
		},
		"exit without errors when prompting zero environment variables": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On(
					"ListVariables",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]string{},
					nil,
				)
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowHostedOnly, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"ListVariables",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvRemoveSuccess,
					mock.Anything,
				)
			},
			ExpectedStdoutOutputs: []string{
				"The app has no environment variables to remove",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewEnvRemoveCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}
