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

package doctor

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/deputil"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDoctorCommand(t *testing.T) {
	expectedCLIVersion := version.Get()
	expectedCredentials := types.SlackAuth{
		TeamDomain: "team123",
		TeamID:     "T123",
		UserID:     "U123",
	}
	expectedGitVersion, err := deputil.GetGitVersion()
	require.NoError(t, err)
	expectedOSVersion := fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
	expectedManifestSource := "local"
	expectedProjectID := "project-abcdef"
	expectedSystemID := "system-123456"
	expectedUpdateTime := "0001-01-01 00:00:00 Z"

	t.Run("creates a complete report", func(t *testing.T) {
		clientsMock := shared.NewClientsMock()
		clientsMock.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{expectedCredentials}, nil)
		clientsMock.AuthInterface.On("ResolveApiHost", mock.Anything, mock.Anything, mock.Anything).Return("api.slack.com")
		clientsMock.ApiInterface.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
		clientsMock.AddDefaultMocks()
		pcm := &config.ProjectConfigMock{}
		pcm.On("ReadProjectConfigFile", mock.Anything).Return(config.ProjectConfig{
			ProjectID: expectedProjectID,
			Manifest: &config.ManifestConfig{
				Source: config.MANIFEST_SOURCE_LOCAL.String(),
			},
		}, nil)
		clientsMock.Config.ProjectConfig = pcm
		scm := &config.SystemConfigMock{}
		scm.On("GetSurveyConfig", mock.Anything, mock.Anything).Return(config.SurveyConfig{}, nil)
		scm.On("SetSurveyConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		scm.On("UserConfig", mock.Anything).Return(&config.SystemConfig{
			SystemID: expectedSystemID,
		}, nil)
		clientsMock.Config.SystemConfig = scm
		mockDoctorScript := hooks.HookScript{Command: "echo checkup"}
		mockDoctorHook := `{"versions": [{"name": "node", "current": "20.11.1"}]}`
		clientsMock.HookExecutor.On("Execute", hooks.HookExecOpts{Hook: mockDoctorScript}).
			Return(mockDoctorHook, nil)
		mockUpdateScript := hooks.HookScript{Command: "echo update"}
		mockUpdateHook := `{"name": "the Slack SDK", "releases": [{"name": "@slack/bolt", "current": "1.0.0", "latest": "2.2.2", "breaking": true, "update": true}]}`
		clientsMock.HookExecutor.On("Execute", hooks.HookExecOpts{Hook: mockUpdateScript}).
			Return(mockUpdateHook, nil)
		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig.WorkingDirectory = "."
			clients.SDKConfig.Hooks.CheckUpdate = mockUpdateScript
			clients.SDKConfig.Hooks.Doctor = mockDoctorScript
		})

		cmd := NewDoctorCommand(clients)
		testutil.MockCmdIO(clients.IO, cmd)
		err := cmd.Execute()
		require.NoError(t, err)

		report, err := performChecks(cmd.Context(), clients)
		require.NoError(t, err)

		expectedValues := DoctorReport{
			Sections: []Section{
				{
					Label: "SYSTEM",
					Subsections: []Section{
						{
							Label: "Operating System",
							Value: "the computer conductor",
							Subsections: []Section{
								{
									Label: "Version",
									Value: expectedOSVersion,
								},
							},
						},
						{
							Label: "Git",
							Value: "a version control system",
							Subsections: []Section{
								{
									Label:       "Version",
									Value:       expectedGitVersion,
									Subsections: []Section{},
									Errors:      []slackerror.Error{},
								},
							},
							Errors: []slackerror.Error{},
						},
					},
				},
				{
					Label: "SLACK",
					Subsections: []Section{
						{
							Label: "CLI",
							Value: "this tool for building Slack apps",
							Subsections: []Section{
								{
									Label:       "Version",
									Value:       expectedCLIVersion,
									Subsections: []Section{},
									Errors:      []slackerror.Error{},
								},
							},
							Errors: []slackerror.Error{},
						},
						{
							Label: "Configurations",
							Value: "any adjustments to settings",
							Subsections: []Section{
								{
									Label:       "System ID",
									Value:       expectedSystemID,
									Subsections: []Section{},
									Errors:      []slackerror.Error{},
								},
								{
									Label:       "Last updated",
									Value:       expectedUpdateTime,
									Subsections: []Section{},
									Errors:      []slackerror.Error{},
								},
								{
									Label:       "Experiments",
									Value:       "None",
									Subsections: []Section{},
									Errors:      []slackerror.Error{},
								},
							},
							Errors: []slackerror.Error{},
						},
						{
							Label: "Credentials",
							Value: "your Slack authentication",
							Subsections: []Section{
								{
									Label: "",
									Value: "",
									Subsections: []Section{
										{
											Label:       "Team domain",
											Value:       expectedCredentials.TeamDomain,
											Subsections: []Section{},
											Errors:      []slackerror.Error{},
										},
										{
											Label:       "Team ID",
											Value:       expectedCredentials.TeamID,
											Subsections: []Section{},
											Errors:      []slackerror.Error{},
										},
										{
											Label:       "User ID",
											Value:       expectedCredentials.UserID,
											Subsections: []Section{},
											Errors:      []slackerror.Error{},
										},
										{
											Label:       "Last updated",
											Value:       expectedUpdateTime,
											Subsections: []Section{},
											Errors:      []slackerror.Error{},
										},
										{
											Label:       "Authorization level",
											Value:       "Workspace",
											Subsections: []Section{},
											Errors:      []slackerror.Error{},
										},
										{
											Label:       "Token status",
											Value:       "Valid",
											Subsections: []Section{},
											Errors:      []slackerror.Error{},
										},
									},
									Errors: []slackerror.Error{},
								},
							},
							Errors: []slackerror.Error{},
						},
					},
				},
				{
					Label: "PROJECT",
					Subsections: []Section{
						{
							Label: "Configurations",
							Value: "your project's CLI settings",
							Subsections: []Section{
								{
									Label: "Manifest source",
									Value: expectedManifestSource,
								},
								{
									Label: "Project ID",
									Value: expectedProjectID,
								},
							},
						},
						{
							Label: "Runtime",
							Value: "foundations for the application",
							Subsections: []Section{
								{
									Label:       "node",
									Value:       "20.11.1",
									Subsections: []Section{},
									Errors:      []slackerror.Error{},
								},
							},
							Errors: []slackerror.Error{},
						},
						{
							Label: "Dependencies",
							Value: "requisites for development",
							Subsections: []Section{
								{
									Label: "@slack/bolt",
									Value: "1.0.0 → 2.2.2 (update available)",
								},
							},
						},
					},
				},
			},
		}

		expectedStrings := []string{
			"SYSTEM",
			"Operating System (the computer conductor)",
			fmt.Sprintf("Version: %s", expectedOSVersion),
			"Git (a version control system)",
			fmt.Sprintf("Version: %s", expectedGitVersion),
			"SLACK",
			"CLI (this tool for building Slack apps)",
			fmt.Sprintf("Version: %s", expectedCLIVersion),
			"Configurations (any adjustments to settings)",
			fmt.Sprintf("System ID: %s", expectedSystemID),
			fmt.Sprintf("Last updated: %s", expectedUpdateTime),
			"Experiments: None",
			"Credentials (your Slack authentication)",
			fmt.Sprintf("Team domain: %s", expectedCredentials.TeamDomain),
			fmt.Sprintf("Team ID: %s", expectedCredentials.TeamID),
			fmt.Sprintf("User ID: %s", expectedCredentials.UserID),
			fmt.Sprintf("Last updated: %s", expectedUpdateTime),
			"Authorization level: Workspace",
			"Token status: Valid",
			"PROJECT",
			"Configurations (your project's CLI settings)",
			fmt.Sprintf("Manifest source: %s", expectedManifestSource),
			fmt.Sprintf("Project ID: %s", expectedProjectID),
			"Runtime (foundations for the application)",
			"node: 20.11.1",
			"Dependencies (requisites for development)",
			"@slack/bolt: 1.0.0 → 2.2.2 (update available)",
			"Errors: 0",
		}

		assert.Equal(t, expectedValues, report)
		for _, str := range expectedStrings {
			assert.Contains(t, clientsMock.GetStdoutOutput(), str)
		}
	})

	t.Run("errors on broken template", func(t *testing.T) {
		clientsMock := shared.NewClientsMock()
		clientsMock.AddDefaultMocks()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		cmd := NewDoctorCommand(clients)
		testutil.MockCmdIO(clients.IO, cmd)

		embedDocTmplHolder := embedDocTmpl
		embedDocTmpl = bytes.NewBufferString("{{ BrokenTemplate }").Bytes()
		defer func() {
			embedDocTmpl = embedDocTmplHolder
		}()

		err := cmd.Execute()
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "function \"BrokenTemplate\" not defined")
		}
	})
}

func TestDoctorHook(t *testing.T) {
	tests := map[string]struct {
		mockHookSetup func(cm *shared.ClientsMock) *shared.ClientFactory
		expectedHook  DoctorHookJSON
		expectedError *slackerror.Error
	}{
		"errors without a doctor hook": {
			expectedError: slackerror.New(slackerror.ErrSDKHookNotFound).
				WithMessage("The `doctor` hook was not found").
				WithRemediation("Debug responses from the Slack hooks file (%s)", config.GetProjectHooksJSONFilePath()),
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				return shared.NewClientFactory(cm.MockClientFactory())
			},
		},
		"returns the application runtime version": {
			expectedHook: DoctorHookJSON{
				Versions: []struct {
					Name    string `json:"name"`
					Current string `json:"current"`
					Message string `json:"message"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}{
					{
						Name:    "deno",
						Current: "1.0.0",
					},
					{
						Name:    "typescript",
						Current: "5.4.3",
					},
				},
			},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				mockDoctorHook := `{"versions": [{"name": "deno", "current": "1.0.0"}, {"name": "typescript", "current": "5.4.3"}]}`
				mockDoctorScript := hooks.HookScript{Command: "echo checkup"}
				cm.HookExecutor.On("Execute", hooks.HookExecOpts{Hook: mockDoctorScript}).
					Return(mockDoctorHook, nil)
				return shared.NewClientFactory(cm.MockClientFactory(), func(clients *shared.ClientFactory) {
					clients.SDKConfig.WorkingDirectory = "."
					clients.SDKConfig.Hooks.Doctor = mockDoctorScript
				})
			},
		},
		"errors and displays with the message and error provided by hook": {
			expectedHook: DoctorHookJSON{
				Versions: []struct {
					Name    string `json:"name"`
					Current string `json:"current"`
					Message string `json:"message"`
					Error   struct {
						Message string `json:"message"`
					} `json:"error"`
				}{
					{
						Name:    "deno",
						Current: "1.0.0",
						Message: "Secure runtimes make safer code",
						Error: struct {
							Message string `json:"message"`
						}{
							Message: "Something isn't right with this installation",
						},
					},
				},
			},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				mockDoctorHook := `{"versions": [{"name": "deno", "current": "1.0.0", "message": "Secure runtimes make safer code", "error": {"message": "Something isn't right with this installation"}}]}`
				mockDoctorScript := hooks.HookScript{Command: "echo checkup"}
				cm.HookExecutor.On("Execute", hooks.HookExecOpts{Hook: mockDoctorScript}).
					Return(mockDoctorHook, nil)
				return shared.NewClientFactory(cm.MockClientFactory(), func(clients *shared.ClientFactory) {
					clients.SDKConfig.WorkingDirectory = "."
					clients.SDKConfig.Hooks.Doctor = mockDoctorScript
				})
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := tt.mockHookSetup(clientsMock)
			response, err := doctorHook(ctx, clients)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHook, response)
			}
		})
	}
}
