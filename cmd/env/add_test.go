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

var mockAuth = types.SlackAuth{
	TeamID: "T1",
	Token:  "12345",
}

var mockApp = types.App{
	TeamID:     "T1",
	TeamDomain: "team1",
	AppID:      "A0123456789",
}

func Test_Env_AddCommandPreRun(t *testing.T) {
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
			cmd := NewEnvAddCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_Env_AddCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"add a variable using arguments": {
			CmdArgs: []string{"ENV_NAME", "ENV_VALUE"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddHostedMocks(ctx, cm, cf)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"AddVariable",
					mock.Anything,
					mock.Anything,
					mockApp.AppID,
					"ENV_NAME",
					"ENV_VALUE",
				)
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvAddSuccess,
					mock.Anything,
				)
			},
		},
		"provide a variable name by argument and value by prompt": {
			CmdArgs: []string{"ENV_NAME"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddHostedMocks(ctx, cm, cf)
				cm.IO.On(
					"PasswordPrompt",
					mock.Anything,
					"Variable value",
					iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
						Flag: cm.Config.Flags.Lookup("value"),
					}),
				).Return(
					iostreams.PasswordPromptResponse{
						Prompt: true,
						Value:  "secret_key_1234",
					},
					nil,
				)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"AddVariable",
					mock.Anything,
					mock.Anything,
					mockApp.AppID,
					"ENV_NAME",
					"secret_key_1234",
				)
			},
		},
		"provide a variable name by argument and value by flag": {
			CmdArgs: []string{"ENV_NAME", "--value", "example_value"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddHostedMocks(ctx, cm, cf)
				cm.IO.On(
					"PasswordPrompt",
					mock.Anything,
					"Variable value",
					iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
						Flag: cm.Config.Flags.Lookup("value"),
					}),
				).Return(
					iostreams.PasswordPromptResponse{
						Flag:  true,
						Value: "example_value",
					},
					nil,
				)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"AddVariable",
					mock.Anything,
					mock.Anything,
					mockApp.AppID,
					"ENV_NAME",
					"example_value",
				)
			},
		},
		"provide both variable name and value by prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddHostedMocks(ctx, cm, cf)
				cm.IO.On(
					"InputPrompt",
					mock.Anything,
					"Variable name",
					mock.Anything,
				).Return(
					"some_name",
					nil,
				)
				cm.IO.On(
					"PasswordPrompt",
					mock.Anything,
					"Variable value",
					iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
						Flag: cm.Config.Flags.Lookup("value"),
					}),
				).Return(
					iostreams.PasswordPromptResponse{
						Prompt: true,
						Value:  "example_value",
					},
					nil,
				)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(
					t,
					"AddVariable",
					mock.Anything,
					mock.Anything,
					mockApp.AppID,
					"some_name",
					"example_value",
				)
			},
		},
		"add a variable to the .env file for remote runtime": {
			CmdArgs: []string{"PORT", "3000"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
					types.SlackYaml{AppManifest: types.AppManifest{Settings: &types.AppSettings{FunctionRuntime: types.Remote}}},
					nil,
				)
				cm.AppClient.Manifest = manifestMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "PORT=3000\n", string(content))
			},
		},
		"add a variable to the .env file when no runtime is set": {
			CmdArgs: []string{"NEW_VAR", "new_value"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, nil)
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env", []byte("# Config\nEXISTING=value\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "# Config\nEXISTING=value\n"+`NEW_VAR="new_value"`+"\n", string(content))
			},
		},
		"add a variable to the .env file when manifest fetch errors": {
			CmdArgs: []string{"API_KEY", "sk-1234"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
					types.SlackYaml{},
					slackerror.New(slackerror.ErrSDKHookNotFound),
				)
				cm.AppClient.Manifest = manifestMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, `API_KEY="sk-1234"`+"\n", string(content))
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewEnvAddCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

// setupEnvAddHostedMocks prepares common mocks for hosted app tests
func setupEnvAddHostedMocks(ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cf.SDKConfig = hooks.NewSDKConfigMock()
	cm.AddDefaultMocks()
	_ = cf.AppClient().SaveDeployed(ctx, mockApp)

	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{Auth: mockAuth, App: mockApp}, nil)

	cm.Config.Flags.String("value", "", "mock value flag")

	cm.API.On("AddVariable", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

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

// setupEnvAddDotenvMocks prepares common mocks for non-hosted (dotenv) app tests.
// Callers must set their own manifest mock on cm.AppClient.Manifest.
func setupEnvAddDotenvMocks(_ context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cf.SDKConfig = hooks.NewSDKConfigMock()
	cm.AddDefaultMocks()

	cm.Config.Flags.String("value", "", "mock value flag")
}
