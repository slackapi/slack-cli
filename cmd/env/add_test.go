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
		"add a variable to the .env file for non-hosted app": {
			CmdArgs: []string{"ENV_NAME", "ENV_VALUE"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvAddSuccess,
					mock.Anything,
				)
			},
		},
		"add a variable preserving existing variables and comments in .env": {
			CmdArgs: []string{"NEW_VAR", "new_value"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
				err := afero.WriteFile(cf.Fs, ".env", []byte("# Config\nEXISTING=value\n"), 0644)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
			},
		},
		"overwrite an existing variable in .env file": {
			CmdArgs: []string{"MY_VAR", "new_value"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
				err := afero.WriteFile(cf.Fs, ".env", []byte("MY_VAR=old_value\n"), 0644)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
			},
		},
		"create .env file when it does not exist": {
			CmdArgs: []string{"FIRST_VAR", "first_value"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupEnvAddDotenvMocks(ctx, cm, cf)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "AddVariable")
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvAddSuccess,
					mock.Anything,
				)
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

// setupEnvAddDotenvMocks prepares common mocks for non-hosted (dotenv) app tests
func setupEnvAddDotenvMocks(_ context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cf.SDKConfig = hooks.NewSDKConfigMock()
	cm.AddDefaultMocks()

	mockDevApp := types.App{
		TeamID:     "T1",
		TeamDomain: "team1",
		AppID:      "A0123456789",
		IsDev:      true,
	}
	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{Auth: mockAuth, App: mockDevApp}, nil)

	cm.Config.Flags.String("value", "", "mock value flag")
}

func Test_setDotEnv(t *testing.T) {
	tests := map[string]struct {
		existingEnv   string
		writeExisting bool
		name          string
		value         string
		expectedFile  string
	}{
		"creates .env file when it does not exist": {
			name:         "FOO",
			value:        "bar",
			expectedFile: "FOO=\"bar\"\n",
		},
		"adds a variable to an empty .env file": {
			existingEnv:   "",
			writeExisting: true,
			name:          "FOO",
			value:         "bar",
			expectedFile:  "FOO=\"bar\"\n",
		},
		"adds a variable preserving existing variables": {
			existingEnv:   "EXISTING=value\n",
			writeExisting: true,
			name:          "NEW_VAR",
			value:         "new_value",
			expectedFile:  "EXISTING=value\nNEW_VAR=\"new_value\"\n",
		},
		"adds a variable preserving comments and blank lines": {
			existingEnv:   "# Database config\nDB_HOST=localhost\n\n# API keys\nAPI_KEY=secret\n",
			writeExisting: true,
			name:          "NEW_VAR",
			value:         "new_value",
			expectedFile:  "# Database config\nDB_HOST=localhost\n\n# API keys\nAPI_KEY=secret\nNEW_VAR=\"new_value\"\n",
		},
		"updates an existing unquoted variable in-place": {
			existingEnv:   "# Config\nFOO=old_value\nBAR=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "# Config\nFOO=\"new_value\"\nBAR=keep\n",
		},
		"updates an existing quoted variable in-place": {
			existingEnv:   "FOO=\"old_value\"\nBAR=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "FOO=\"new_value\"\nBAR=keep\n",
		},
		"updates a variable with export prefix": {
			existingEnv:   "export SECRET=old_secret\nOTHER=keep\n",
			writeExisting: true,
			name:          "SECRET",
			value:         "new_secret",
			expectedFile:  "export SECRET=\"new_secret\"\nOTHER=keep\n",
		},
		"escapes special characters in values": {
			name:         "SPECIAL",
			value:        "has \"quotes\" and $vars and \\ backslash",
			expectedFile: "SPECIAL=\"has \\\"quotes\\\" and \\$vars and \\\\ backslash\"\n",
		},
		"round-trips through LoadDotEnv": {
			name:         "ROUND_TRIP",
			value:        "hello world",
			expectedFile: "ROUND_TRIP=\"hello world\"\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tc.writeExisting {
				err := afero.WriteFile(fs, ".env", []byte(tc.existingEnv), 0644)
				assert.NoError(t, err)
			}
			err := setDotEnv(fs, tc.name, tc.value)
			assert.NoError(t, err)
			content, err := afero.ReadFile(fs, ".env")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFile, string(content))
		})
	}
}
