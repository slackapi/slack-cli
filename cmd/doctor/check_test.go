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
	"fmt"
	"runtime"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/deputil"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDoctorCheckOS(t *testing.T) {
	t.Run("returns the operating system version", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clientsMock.AddDefaultMocks()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		expected := Section{
			Label: "Operating System",
			Value: "the computer conductor",
			Subsections: []Section{
				{
					Label: "Version",
					Value: fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH),
				},
			},
		}

		section := checkOS(ctx, clients)
		assert.Equal(t, expected, section)
	})
}

func TestDoctorCheckCLIVersion(t *testing.T) {
	t.Run("returns the current version of this tool", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clientsMock.AddDefaultMocks()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		expected := Section{
			Label: "CLI",
			Value: "this tool for building Slack apps",
			Subsections: []Section{
				{
					Label:       "Version",
					Value:       version.Version,
					Subsections: []Section{},
					Errors:      []slackerror.Error{},
				},
			},
			Errors: []slackerror.Error{},
		}

		section, err := checkCLIVersion(ctx, clients)
		assert.NoError(t, err)
		assert.Equal(t, expected, section)
	})
}

func TestDoctorCheckProjectConfig(t *testing.T) {
	const projectSectionLabel = "Project ID"
	const manifestSourceSectionLabel = "Manifest source"

	tests := map[string]struct {
		projectConfig                 config.ProjectConfig
		expectedProjectSection        *Section
		expectedManifestSourceSection *Section
		expectedErrors                []slackerror.Error
	}{
		"returns a valid project ID and valid manifest source": {
			projectConfig: config.ProjectConfig{
				ProjectID: "project-abcdef",
				Manifest: &config.ManifestConfig{
					Source: config.ManifestSourceLocal.String(),
				},
			},
			expectedProjectSection: &Section{
				Label: projectSectionLabel,
				Value: "project-abcdef",
			},
			expectedManifestSourceSection: &Section{
				Label: manifestSourceSectionLabel,
				Value: "local",
			},
			expectedErrors: nil,
		},
		"returns a valid project ID and missing manifest source": {
			projectConfig: config.ProjectConfig{
				ProjectID: "project-abcdef",
				Manifest:  nil,
			},
			expectedProjectSection: &Section{
				Label: projectSectionLabel,
				Value: "project-abcdef",
			},
			expectedManifestSourceSection: nil,
			expectedErrors: []slackerror.Error{
				*slackerror.New(slackerror.ErrProjectConfigManifestSource),
			},
		},
		"returns a valid project ID and empty manifest source": {
			projectConfig: config.ProjectConfig{
				ProjectID: "project-abcdef",
				Manifest: &config.ManifestConfig{
					Source: "",
				},
			},
			expectedProjectSection: &Section{
				Label: projectSectionLabel,
				Value: "project-abcdef",
			},
			expectedManifestSourceSection: nil,
			expectedErrors: []slackerror.Error{
				*slackerror.New(slackerror.ErrProjectConfigManifestSource),
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			pcm := &config.ProjectConfigMock{}
			pcm.On("ReadProjectConfigFile", mock.Anything).Return(tt.projectConfig, nil)
			clientsMock.Config.ProjectConfig = pcm
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			expected := Section{
				Label:       "Configurations",
				Value:       "your project's CLI settings",
				Subsections: []Section{},
				Errors:      tt.expectedErrors,
			}

			if tt.expectedManifestSourceSection != nil {
				expected.Subsections = append(expected.Subsections, *tt.expectedManifestSourceSection)
			}

			if tt.expectedProjectSection != nil {
				expected.Subsections = append(expected.Subsections, *tt.expectedProjectSection)
			}

			section := checkProjectConfig(ctx, clients)
			assert.Equal(t, expected, section)
		})
	}
}

func TestDoctorCheckProjectDeps(t *testing.T) {
	tests := map[string]struct {
		mockHookSetup        func(cm *shared.ClientsMock) *shared.ClientFactory
		expectedSubsections  []Section
		expectedErrorSection []slackerror.Error
	}{
		"errors without a check-update hook": {
			expectedErrorSection: []slackerror.Error{
				*slackerror.New(slackerror.ErrSDKHookNotFound).
					WithMessage("The `check-update` hook was not found").
					WithRemediation("Debug responses from the Slack hooks file (%s)", config.GetProjectHooksJSONFilePath()),
			},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				return shared.NewClientFactory(cm.MockClientFactory())
			},
		},
		"recommends any known updates to dependencies": {
			expectedSubsections: []Section{
				{
					Label: "deno_slack_sdk",
					Value: "1.0.0 → 2.2.2 (update available)",
				},
				{
					Label: "deno_slack_api",
					Value: "4.0.0",
				},
				{
					Label: "deno_slack_hooks",
					Value: "2.6.0 → 2.7.0 (update available)",
				},
			},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				mockSDKUpdate := `{"name": "deno_slack_sdk", "current": "1.0.0", "latest": "2.2.2", "breaking": true, "update": true}`
				mockAPIUpdate := `{"name": "deno_slack_api", "current": "4.0.0", "latest": "4.0.0", "breaking": false, "update": false}`
				mockHookUpdate := `{"name": "deno_slack_hooks", "current": "2.6.0", "latest": "2.7.0", "breaking": false, "update": true}`
				mockUpdate := fmt.Sprintf(`{"name": "the Slack SDK", "releases": [%s, %s, %s]}`, mockSDKUpdate, mockAPIUpdate, mockHookUpdate)
				mockUpdateScript := hooks.HookScript{Command: "echo updates"}
				cm.HookExecutor.On("Execute", mock.Anything, hooks.HookExecOpts{Hook: mockUpdateScript}).
					Return(mockUpdate, nil)
				return shared.NewClientFactory(cm.MockClientFactory(), func(clients *shared.ClientFactory) {
					clients.SDKConfig.WorkingDirectory = "."
					clients.SDKConfig.Hooks.CheckUpdate = mockUpdateScript
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
			expected := Section{
				Label:       "Dependencies",
				Value:       "requisites for development",
				Subsections: tt.expectedSubsections,
				Errors:      tt.expectedErrorSection,
			}

			section := checkProjectDeps(ctx, clients)
			assert.Equal(t, expected, section)
		})
	}
}

func TestDoctorCheckCLIConfig(t *testing.T) {
	tests := map[string]struct {
		systemID string
	}{
		"returns any adjustments to settings": {
			systemID: "system-123456",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			scm := &config.SystemConfigMock{}
			scm.On("UserConfig", mock.Anything).Return(&config.SystemConfig{
				SystemID: tt.systemID,
			}, nil)
			clientsMock.Config.SystemConfig = scm
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			expected := Section{
				Label: "Configurations",
				Value: "any adjustments to settings",
				Subsections: []Section{
					{
						Label:       "System ID",
						Value:       tt.systemID,
						Subsections: []Section{},
						Errors:      []slackerror.Error{},
					},
					{
						Label:       "Last updated",
						Value:       "0001-01-01 00:00:00 Z",
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
			}

			section, err := checkCLIConfig(ctx, clients)
			assert.NoError(t, err)
			assert.Equal(t, expected, section)
		})
	}
}

func TestDoctorCheckCLICreds(t *testing.T) {
	tests := map[string]struct {
		auths types.SlackAuth
	}{
		"errors without available authorizations": {},
	}

	for name := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			expected := Section{
				Label:       "Credentials",
				Value:       "your Slack authentication",
				Subsections: []Section{},
				Errors:      []slackerror.Error{*slackerror.New(slackerror.ErrNotAuthed)},
			}

			section, err := checkCLICreds(ctx, clients)
			assert.NoError(t, err)
			assert.Equal(t, expected, section)
		})
	}
}

func TestDoctorCheckProjectTooling(t *testing.T) {
	tests := map[string]struct {
		mockHookSetup        func(cm *shared.ClientsMock) *shared.ClientFactory
		expectedSubsections  []Section
		expectedErrorSection []slackerror.Error
	}{
		"errors without a doctor hook": {
			expectedSubsections: []Section{},
			expectedErrorSection: []slackerror.Error{
				*slackerror.New(slackerror.ErrSDKHookNotFound).
					WithMessage("The `doctor` hook was not found").
					WithRemediation("Debug responses from the Slack hooks file (%s)", config.GetProjectHooksJSONFilePath()),
			},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				return shared.NewClientFactory(cm.MockClientFactory())
			},
		},
		"returns the application runtime version": {
			expectedSubsections: []Section{
				{
					Label:       "deno",
					Value:       "1.0.0",
					Subsections: []Section{},
					Errors:      []slackerror.Error{},
				},
				{
					Label:       "typescript",
					Value:       "5.4.3",
					Subsections: []Section{},
					Errors:      []slackerror.Error{},
				},
			},
			expectedErrorSection: []slackerror.Error{},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				mockDoctorHook := `{"versions": [{"name": "deno", "current": "1.0.0"}, {"name": "typescript", "current": "5.4.3"}]}`
				mockDoctorScript := hooks.HookScript{Command: "echo checkup"}
				cm.HookExecutor.On("Execute", mock.Anything, hooks.HookExecOpts{Hook: mockDoctorScript}).
					Return(mockDoctorHook, nil)
				return shared.NewClientFactory(cm.MockClientFactory(), func(clients *shared.ClientFactory) {
					clients.SDKConfig.WorkingDirectory = "."
					clients.SDKConfig.Hooks.Doctor = mockDoctorScript
				})
			},
		},
		"errors and displays with the message and error provided by hook": {
			expectedSubsections: []Section{
				{
					Label: "deno",
					Value: "1.0.0",
					Subsections: []Section{
						{
							Label: "Note: Secure runtimes make safer code",
						},
					},
					Errors: []slackerror.Error{
						*slackerror.New(slackerror.ErrRuntimeNotSupported).
							WithMessage("Something isn't right with this installation"),
					},
				},
			},
			expectedErrorSection: []slackerror.Error{},
			mockHookSetup: func(cm *shared.ClientsMock) *shared.ClientFactory {
				mockDoctorHook := `{"versions": [{"name": "deno", "current": "1.0.0", "message": "Secure runtimes make safer code", "error": {"message": "Something isn't right with this installation"}}]}`
				mockDoctorScript := hooks.HookScript{Command: "echo checkup"}
				cm.HookExecutor.On("Execute", mock.Anything, hooks.HookExecOpts{Hook: mockDoctorScript}).
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
			expected := Section{
				Label:       "Runtime",
				Value:       "foundations for the application",
				Subsections: tt.expectedSubsections,
				Errors:      tt.expectedErrorSection,
			}

			section := checkProjectTooling(ctx, clients)
			assert.Equal(t, expected, section)
		})
	}
}

func TestDoctorCheckGit(t *testing.T) {
	t.Run("returns the version of git", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		gitVersion, err := deputil.GetGitVersion()
		require.NoError(t, err)
		expected := Section{
			Label: "Git",
			Value: "a version control system",
			Subsections: []Section{
				{
					Label:       "Version",
					Value:       gitVersion,
					Subsections: []Section{},
					Errors:      []slackerror.Error{},
				},
			},
			Errors: []slackerror.Error{},
		}

		section, err := CheckGit(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, section)
	})
}
