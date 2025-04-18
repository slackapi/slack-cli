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

package platform

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Setup a mock for the Deploy package
type DeployPkgMock struct {
	mock.Mock
}

func (m *DeployPkgMock) Deploy(ctx context.Context, clients *shared.ClientFactory, showPrompts bool, log *logger.Logger, app types.App) (*logger.LogEvent, error) {
	args := m.Called(ctx, clients, showPrompts, log, app)
	return args.Get(0).(*logger.LogEvent), args.Error(1)
}

// Setup a mock for Install package
type AppCmdMock struct {
	mock.Mock
}

func (m *AppCmdMock) RunAddCommand(ctx context.Context, clients *shared.ClientFactory, selectedApp *prompts.SelectedApp, orgGrant string) (context.Context, types.InstallState, types.App, error) {
	m.Called()
	return ctx, "", types.App{}, nil
}

// TODO: improve this test, it only tests the mock that we install ourselves on the function doing all the deploy work is called. Add actual tests here.
func TestDeployCommand(t *testing.T) {
	// Create mocks
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		projectConfigMock := config.NewProjectConfigMock()
		projectConfigMock.AddDefaultMocks()
		clients.Config.ProjectConfig = projectConfigMock
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewDeployCommand(clients)
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
	testutil.MockCmdIO(clients.IO, cmd)

	deployPkgMock := new(DeployPkgMock)
	deployFunc = deployPkgMock.Deploy
	deployPkgMock.On("Deploy", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&logger.LogEvent{
		Data: logger.LogData{"authSession": "{}"},
	}, nil)

	appSelectMock := prompts.NewAppSelectMock()
	appSelectMock.On("TeamAppSelectPrompt").Return(prompts.SelectedApp{}, nil)
	teamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt

	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			Settings: &types.AppSettings{
				FunctionRuntime: types.SLACK_HOSTED,
			},
		},
	}, nil)
	clients.AppClient().Manifest = manifestMock

	appCmdMock := new(AppCmdMock)
	runAddCommandFunc = appCmdMock.RunAddCommand
	appCmdMock.On("RunAddCommand").Return()

	err := cmd.Execute()
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error", err.Error())
	}

	deployPkgMock.AssertCalled(t, "Deploy", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestDeployCommand_HasValidDeploymentMethod(t *testing.T) {
	tests := map[string]struct {
		manifest       types.SlackYaml
		manifestError  error
		manifestSource config.ManifestSource
		deployScript   string
		expectedError  error
	}{
		"fails when no manifest exists": {
			manifestError:  slackerror.New(slackerror.ErrInvalidManifest),
			manifestSource: config.MANIFEST_SOURCE_LOCAL,
			expectedError:  slackerror.New(slackerror.ErrInvalidManifest),
		},
		"succeeds with a slack hosted function runtime": {
			manifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.SLACK_HOSTED,
					},
				},
			},
			manifestSource: config.MANIFEST_SOURCE_LOCAL,
		},
		"succeeds if a deploy hook script is available to project manifest sources": {
			manifestSource: config.MANIFEST_SOURCE_LOCAL,
			deployScript:   "echo go!",
		},
		"continues if a deploy hook script is available to remote manifest sources": {
			manifestSource: config.MANIFEST_SOURCE_REMOTE,
			deployScript:   "sleep 4",
		},
		"fails if no deploy hook is provided": {
			manifestSource: config.MANIFEST_SOURCE_LOCAL,
			expectedError:  slackerror.New(slackerror.ErrSDKHookNotFound),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(tt.manifest, tt.manifestError)
				clients.AppClient().Manifest = manifestMock
				projectConfigMock := config.NewProjectConfigMock()
				projectConfigMock.On("GetManifestSource", mock.Anything).
					Return(tt.manifestSource, nil)
				clients.Config.ProjectConfig = projectConfigMock
				clients.SDKConfig = hooks.NewSDKConfigMock()
				clients.SDKConfig.Hooks.Deploy.Command = tt.deployScript
			})
			err := hasValidDeploymentMethod(ctx, clients, types.App{}, types.SlackAuth{})
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(
					t,
					slackerror.ToSlackError(tt.expectedError).Code,
					slackerror.ToSlackError(err).Code,
				)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeployCommand_DeployHook(t *testing.T) {
	tests := map[string]struct {
		command        string
		expectedStderr []string
		expectedStdout string
		expectedError  error
	}{
		"fails to execute an unknown script path": {
			command:       "./deployer.sh",
			expectedError: slackerror.New(slackerror.ErrSDKHookInvocationFailed),
			expectedStderr: []string{
				"./deployer.sh: No such file or directory",
				"./deployer.sh: not found",
				"no such file or directory: ./deployer.sh",
			},
		},
		"echos stdout output to standard out": {
			command:        "echo example_output_goes_here",
			expectedStdout: "example_output_goes_here",
		},
		"echos stderr output to standard err": {
			command:        ">&2 echo 'uhoh'",
			expectedStderr: []string{"uhoh"},
		},
		"works well with shell built ins": {
			command:        "exit 2",
			expectedStderr: []string{"exit status 2"},
			expectedError:  slackerror.New(slackerror.ErrSDKHookInvocationFailed),
		},
		"recognizes and runs compound commands": {
			command:        "echo 'DENIED!!!!' && exit 1",
			expectedStdout: "DENIED!!!!",
			expectedStderr: []string{"exit status 1"},
			expectedError:  slackerror.New(slackerror.ErrSDKHookInvocationFailed),
		},
		"performs truthful conditionals correctly": {
			command: "true || exit 1",
		},
		"performs falseful conditionals correctly": {
			command:        "false || echo 'not facts'",
			expectedStdout: "not facts",
		},
	}
	containsSubstring := func(buffer string, sub []string) bool {
		if len(sub) == 0 {
			return true
		}
		contains := false
		for _, str := range sub {
			if strings.Contains(buffer, str) {
				contains = true
			}
		}
		return contains
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			sdkConfigMock := hooks.NewSDKConfigMock()
			sdkConfigMock.Config.SupportedProtocols = []hooks.Protocol{hooks.HOOK_PROTOCOL_DEFAULT}
			sdkConfigMock.Hooks.Deploy = hooks.HookScript{Name: "Deploy", Command: tt.command}

			stdoutLogger := log.Logger{}
			stdoutBuffer := bytes.Buffer{}
			stdoutLogger.SetOutput(&stdoutBuffer)
			clientsMock.IO.Stdout = &stdoutLogger
			stderrLogger := log.Logger{}
			stderrBuffer := bytes.Buffer{}
			stderrLogger.SetOutput(&stderrBuffer)
			clientsMock.IO.Stderr = &stderrLogger
			clientsMock.IO.AddDefaultMocks()
			clientsMock.IO.On("WriteIndent", iostreams.WriteSecondarier{Writer: clientsMock.IO.Stdout.Writer()}).
				Return(iostreams.WriteIndenter{Writer: clientsMock.IO.Stdout.Writer()})
			clientsMock.IO.On("WriteIndent", iostreams.WriteSecondarier{Writer: clientsMock.IO.Stderr.Writer()}).
				Return(iostreams.WriteIndenter{Writer: clientsMock.IO.Stderr.Writer()})

			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = sdkConfigMock
				clients.HookExecutor = hooks.GetHookExecutor(clientsMock.IO, sdkConfigMock)
			})
			cmd := NewDeployCommand(clients)
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
			testutil.MockCmdIO(clients.IO, cmd)

			err := cmd.Execute()
			assert.Contains(t, stdoutBuffer.String(), tt.command)
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, err.(*slackerror.Error).Code, tt.expectedError.(*slackerror.Error).Code)
				assert.Contains(t, stdoutBuffer.String(), tt.expectedStdout)
				assert.True(t, containsSubstring(stderrBuffer.String(), tt.expectedStderr))
			} else {
				assert.NoError(t, err)
				assert.Contains(t, stdoutBuffer.String(), tt.expectedStdout)
				assert.True(t, containsSubstring(stderrBuffer.String(), tt.expectedStderr))
			}
		})
	}
}

func TestDeployCommand_PrintHostingCompletion(t *testing.T) {
	tests := map[string]struct {
		event    logger.LogData
		expected []string
	}{
		"information from a workspace deploy is printed": {
			event: logger.LogData{
				"appName":     "DeployerApp",
				"appId":       "A123",
				"deployTime":  "12.34",
				"authSession": `{"user": "slackbot", "user_id": "USLACKBOT", "team": "speck", "team_id": "T001"}`,
			},
			expected: []string{
				"DeployerApp deployed in 12.34",
				"Dashboard:  https://slacker.com/apps/A123",
				"App Owner:  slackbot (USLACKBOT)",
				"Workspace:  speck (T001)",
			},
		},
		"information from an enterprise deploy is printed": {
			event: logger.LogData{
				"appName":     "Spackulen",
				"appId":       "A999",
				"deployTime":  "8.05",
				"authSession": `{"user": "stub", "user_id": "U111", "team": "spack", "team_id": "E002", "is_enterprise_install": true, "enterprise_id": "E002"}`,
			},
			expected: []string{
				"Spackulen deployed in 8.05",
				"Dashboard   :  https://slacker.com/apps/A999",
				"App Owner   :  stub (U111)",
				"Organization:  spack (E002)",
			},
		},
		"a message is still displayed with missing info": {
			event: logger.LogData{
				"authSession": "{}",
			},
			expected: []string{
				"Successfully deployed the app!",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.ApiInterface.On("Host").Return("https://slacker.com")
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cmd := NewDeployCommand(clients)
			log := &logger.LogEvent{Data: tt.event}
			err := printDeployHostingCompletion(clients, cmd, log)
			assert.NoError(t, err)
			clientsMock.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.PlatformDeploySuccess, mock.Anything)
			spinnerText, _ := deploySpinner.Status()
			for _, line := range tt.expected {
				assert.Contains(t, spinnerText, line)
			}
		})
	}
}
