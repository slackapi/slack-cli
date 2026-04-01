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

func Test_Env_ListCommandPreRun(t *testing.T) {
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
			cmd := NewEnvListCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_Env_ListCommand(t *testing.T) {
	mockAppSelect := func() {
		appSelectMock := prompts.NewAppSelectMock()
		appSelectPromptFunc = appSelectMock.AppSelectPrompt
		appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, nil)
	}

	// setupNonHostedManifest configures manifest mocks to return a non-hosted
	// runtime so the command skips app selection and reads from the .env file.
	setupNonHostedManifest := func(cm *shared.ClientsMock, cf *shared.ClientFactory) {
		manifestMock := &app.ManifestMockObject{}
		manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(
			types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.Remote,
					},
				},
			},
			nil,
		)
		cm.AppClient.Manifest = manifestMock
		cf.SDKConfig.WorkingDirectory = "/slack/path/to/project"
	}

	testutil.TableTestCommand(t, testutil.CommandTests{
		"lists variables from the .env file": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupNonHostedManifest(cm, cf)
				err := afero.WriteFile(cf.Fs, ".env", []byte("SECRET_KEY=abc123\nAPI_TOKEN=xyz789\n"), 0644)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvListCount,
					[]string{
						"2",
					},
				)
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvListVariables,
					[]string{
						"API_TOKEN",
						"SECRET_KEY",
					},
				)
			},
		},
		"lists no variables when the .env file does not exist": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupNonHostedManifest(cm, cf)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvListCount,
					[]string{
						"0",
					},
				)
			},
		},
		"lists no variables when the .env file is empty": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupNonHostedManifest(cm, cf)
				err := afero.WriteFile(cf.Fs, ".env", []byte(""), 0644)
				assert.NoError(t, err)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvListCount,
					[]string{
						"0",
					},
				)
			},
		},
		"lists hosted variables using the API": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				mockAppSelect()
				cm.API.On(
					"ListVariables",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					[]string{
						"EXAMPLE_VARIABLE_004",
						"EXAMPLE_VARIABLE_001",
						"EXAMPLE_VARIABLE_002",
					},
					nil,
				)
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
				cf.SDKConfig.WorkingDirectory = "/slack/path/to/project"
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
					slacktrace.EnvListCount,
					[]string{
						"3",
					},
				)
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.EnvListVariables,
					[]string{
						"EXAMPLE_VARIABLE_001",
						"EXAMPLE_VARIABLE_002",
						"EXAMPLE_VARIABLE_004",
					},
				)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewEnvListCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}
