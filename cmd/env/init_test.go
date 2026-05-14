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
	"github.com/slackapi/slack-cli/internal/hooks"
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

func Test_Env_InitCommandPreRun(t *testing.T) {
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
			cmd := NewEnvInitCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_Env_InitCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"copies .env.sample to .env": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, slackerror.New(slackerror.ErrSDKHookNotFound))
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env.sample", []byte("SECRET=placeholder\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.EnvInitSuccess, mock.Anything)
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "SECRET=placeholder\n", string(content))
			},
		},
		"copies .env.example to .env": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, slackerror.New(slackerror.ErrSDKHookNotFound))
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env.example", []byte("TOKEN=example\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.EnvInitSuccess, mock.Anything)
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "TOKEN=example\n", string(content))
			},
		},
		"prints message when .env already exists": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, slackerror.New(slackerror.ErrSDKHookNotFound))
				cm.AppClient.Manifest = manifestMock
				err := afero.WriteFile(cf.Fs, ".env", []byte("EXISTING=value\n"), 0600)
				assert.NoError(t, err)
				err = afero.WriteFile(cf.Fs, ".env.sample", []byte("NEW=value\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.EnvInitSuccess, mock.Anything)
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "EXISTING=value\n", string(content))
			},
		},
		"prints message when no sample file exists": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, slackerror.New(slackerror.ErrSDKHookNotFound))
				cm.AppClient.Manifest = manifestMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.EnvInitSuccess, mock.Anything)
			},
		},
		"prints ROSI message for hosted non-dev app": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig = hooks.NewSDKConfigMock()
				cm.AddDefaultMocks()
				_ = cf.AppClient().SaveDeployed(ctx, mockApp)

				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{Auth: mockAuth, App: mockApp}, nil)

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
			},
			ExpectedStdoutOutputs: []string{
				"ROSI features",
				"env set",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.EnvInitSuccess, mock.Anything)
				_, err := afero.ReadFile(cm.Fs, ".env")
				assert.Error(t, err)
			},
		},
		"copies placeholder for hosted dev app": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig = hooks.NewSDKConfigMock()
				cm.AddDefaultMocks()

				devApp := types.App{
					TeamID:     "T1",
					TeamDomain: "team1",
					AppID:      "A0123456789",
					IsDev:      true,
				}
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{Auth: mockAuth, App: devApp}, nil)

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

				err := afero.WriteFile(cf.Fs, ".env.sample", []byte("SECRET=placeholder\n"), 0600)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.EnvInitSuccess, mock.Anything)
				content, err := afero.ReadFile(cm.Fs, ".env")
				assert.NoError(t, err)
				assert.Equal(t, "SECRET=placeholder\n", string(content))
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewEnvInitCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}
