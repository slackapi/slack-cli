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

package env

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Env_RemoveCommandPreRun(t *testing.T) {
	tests := map[string]struct {
		mockWorkingDirectory string
		expectedError        error
	}{
		"continues if the command is run in a project": {
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        nil,
		},
		"errors if the command is not run in a project": {
			mockWorkingDirectory: "",
			expectedError:        slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = tc.mockWorkingDirectory
			})
			cmd := NewEnvUnsetCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_Env_RemoveCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"exit without errors when .env has no variables to remove": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, nil)
				cm.AppClient.Manifest = manifestMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "RemoveVariable")
			},
			ExpectedStdoutOutputs: []string{
				"The project has no environment variables to remove",
			},
		},
		"exit without errors when hosted app has zero variables": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveHostedMocks(ctx, cm, cf)
				cm.API.ExpectedCalls = nil
				cm.API.On("ListVariables", mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				cm.API.On("RemoveVariable", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "RemoveVariable")
			},
			ExpectedStdoutOutputs: []string{
				"The app has no environment variables to remove",
			},
		},
		"remove a hosted variable using arguments": {
			CmdArgs: []string{"ENV_NAME"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveHostedMocks(ctx, cm, cf)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"RemoveVariable",
					mock.Anything,
					mock.Anything,
					mockApp.AppID,
					"ENV_NAME",
				)
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvUnsetSuccess,
					mock.Anything,
				)
			},
		},
		"remove a hosted variable using prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveHostedMocks(ctx, cm, cf)
				cm.IO.On(
					"SelectPrompt",
					mock.Anything,
					"Select a variable to unset",
					mock.Anything,
					iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
						Flag:     cm.Config.Flags.Lookup("name"),
						Required: true,
					}),
				).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "SELECTED_ENV_VAR",
						Index:  0,
					},
					nil,
				)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"RemoveVariable",
					mock.Anything,
					mock.Anything,
					mockApp.AppID,
					"SELECTED_ENV_VAR",
				)
			},
		},
		"remove a variable from the .env file for remote runtime": {
			CmdArgs: []string{"SECRET"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
					types.SlackYaml{AppManifest: types.AppManifest{Settings: &types.AppSettings{FunctionRuntime: types.Remote}}},
					nil,
				)
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env", []byte("SECRET=abc\nOTHER=keep\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "RemoveVariable")
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvUnsetSuccess,
					mock.Anything,
				)
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "OTHER=keep\n", string(content))
			},
			ExpectedStdoutOutputs: []string{
				"Successfully unset \"SECRET\" as a project environment variable",
			},
		},
		"remove a variable from the .env file using prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
					types.SlackYaml{AppManifest: types.AppManifest{Settings: &types.AppSettings{FunctionRuntime: types.Remote}}},
					nil,
				)
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env", []byte("ALPHA=one\nBETA=two\n"), 0600)
				assert.NoError(t, err)
				cm.IO.On(
					"SelectPrompt",
					mock.Anything,
					"Select a variable to unset",
					mock.Anything,
					iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
						Flag:     cm.Config.Flags.Lookup("name"),
						Required: true,
					}),
				).Return(
					iostreams.SelectPromptResponse{
						Prompt: true,
						Option: "ALPHA",
						Index:  0,
					},
					nil,
				)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "RemoveVariable")
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "BETA=two\n", string(content))
			},
		},
		"remove a variable from the .env file when manifest fetch errors": {
			CmdArgs: []string{"SECRET"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvRemoveDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
					types.SlackYaml{},
					slackerror.New(slackerror.ErrSDKHookNotFound),
				)
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env", []byte("SECRET=abc\nOTHER=keep\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "RemoveVariable")
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "OTHER=keep\n", string(content))
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewEnvUnsetCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

// setupEnvRemoveHostedMocks prepares common mocks for hosted app tests
func setupEnvRemoveHostedMocks(ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cf.SDKConfig = hooks.NewSDKConfigMock()
	cm.AddDefaultMocks()
	_ = cf.AppClient().SaveDeployed(ctx, mockApp)

	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{Auth: mockAuth, App: mockApp}, nil)

	cm.Config.Flags.String("name", "", "mock name flag")

	cm.API.On("ListVariables", mock.Anything, mock.Anything, mock.Anything).Return([]string{"example"}, nil)
	cm.API.On("RemoveVariable", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
		types.SlackYaml{
			AppManifest: types.AppManifest{
				Settings: &types.AppSettings{
					FunctionRuntime: types.SlackHosted,
				},
			},
		},
		nil,
	)
	cm.AppClient.Manifest = manifestMock
	projectConfigMock := config.NewProjectConfigMock()
	projectConfigMock.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
	cm.Config.ProjectConfig = projectConfigMock
	cf.SDKConfig.WorkingDirectory = "/slack/path/to/project"
}

// setupEnvRemoveDotenvMocks prepares common mocks for non-hosted (dotenv) app tests.
// Callers must set their own manifest mock on cm.AppClient.Manifest.
func setupEnvRemoveDotenvMocks(_ context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cf.SDKConfig = hooks.NewSDKConfigMock()
	cm.AddDefaultMocks()

	cm.Config.Flags.String("name", "", "mock name flag")
}
