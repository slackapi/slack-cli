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

package project

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand_AppFlag(t *testing.T) {
	var createClientMock *CreateClientMock

	testutil.TableTestCommand(t, testutil.CommandTests{
		"app flag without template flag returns error": {
			CmdArgs: []string{"my-app", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --app flag requires the --template flag when used with create"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with template fetches manifest and links as local by default": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-test-token", TeamID: "T123", TeamDomain: "test-team", UserID: "U123"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "xoxp-test-token", []string{"A0123456789"}, "T123").
					Return(api.GetAppStatusResult{}, nil)

				manifestMock := cf.AppClient().Manifest.(*app.ManifestMockObject)
				manifestMock.On("GetManifestRemote", mock.Anything, "xoxp-test-token", "A0123456789").
					Return(types.SlackYaml{
						AppManifest: types.AppManifest{
							Settings: &types.AppSettings{
								FunctionRuntime: types.Remote,
							},
						},
					}, nil)

				appClientMock := &app.AppClientMock{}
				appClientMock.On("SaveLocal", mock.Anything).Return(nil)
				cf.AppClient().AppClientInterface = appClientMock

				// Create a real temp directory that os.Chdir can navigate to
				tmpDir := t.TempDir()
				projectDir := filepath.Join(tmpDir, "my-app")
				require.NoError(t, os.MkdirAll(projectDir, 0755))

				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(projectDir, nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with slack-hosted runtime links as deployed": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-test-token", TeamID: "T123", TeamDomain: "test-team", UserID: "U123"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "xoxp-test-token", []string{"A0123456789"}, "T123").
					Return(api.GetAppStatusResult{}, nil)

				manifestMock := cf.AppClient().Manifest.(*app.ManifestMockObject)
				manifestMock.On("GetManifestRemote", mock.Anything, "xoxp-test-token", "A0123456789").
					Return(types.SlackYaml{
						AppManifest: types.AppManifest{
							Settings: &types.AppSettings{
								FunctionRuntime: types.SlackHosted,
							},
						},
					}, nil)

				appClientMock := &app.AppClientMock{}
				appClientMock.On("SaveDeployed", mock.Anything, mock.MatchedBy(func(a types.App) bool {
					return a.AppID == "A0123456789" && a.TeamID == "T123" && !a.IsDev
				})).Return(nil)
				cf.AppClient().AppClientInterface = appClientMock

				// Create a real temp directory that os.Chdir can navigate to
				tmpDir := t.TempDir()
				projectDir := filepath.Join(tmpDir, "my-app")
				require.NoError(t, os.MkdirAll(projectDir, 0755))

				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(projectDir, nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with local runtime links as dev app": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-test-token", TeamID: "T123", TeamDomain: "test-team", UserID: "U123"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "xoxp-test-token", []string{"A0123456789"}, "T123").
					Return(api.GetAppStatusResult{}, nil)

				manifestMock := cf.AppClient().Manifest.(*app.ManifestMockObject)
				manifestMock.On("GetManifestRemote", mock.Anything, "xoxp-test-token", "A0123456789").
					Return(types.SlackYaml{
						AppManifest: types.AppManifest{
							Settings: &types.AppSettings{
								FunctionRuntime: types.LocallyRun,
							},
						},
					}, nil)

				appClientMock := &app.AppClientMock{}
				appClientMock.On("SaveLocal", mock.Anything).Return(nil)
				cf.AppClient().AppClientInterface = appClientMock

				// Create a real temp directory that os.Chdir can navigate to
				tmpDir := t.TempDir()
				projectDir := filepath.Join(tmpDir, "my-app")
				require.NoError(t, os.MkdirAll(projectDir, 0755))

				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(projectDir, nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"name flag overrides manifest display name": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789", "--name", "Custom Name"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-test-token", TeamID: "T123", TeamDomain: "test-team", UserID: "U123"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "xoxp-test-token", []string{"A0123456789"}, "T123").
					Return(api.GetAppStatusResult{}, nil)

				manifestMock := cf.AppClient().Manifest.(*app.ManifestMockObject)
				manifestMock.On("GetManifestRemote", mock.Anything, "xoxp-test-token", "A0123456789").
					Return(types.SlackYaml{
						AppManifest: types.AppManifest{
							DisplayInformation: types.DisplayInformation{
								Name: "Original Remote Name",
							},
							Settings: &types.AppSettings{
								FunctionRuntime: types.Remote,
							},
						},
					}, nil)

				appClientMock := &app.AppClientMock{}
				appClientMock.On("SaveLocal", mock.Anything).Return(nil)
				cf.AppClient().AppClientInterface = appClientMock

				tmpDir := t.TempDir()
				projectDir := filepath.Join(tmpDir, "my-app")
				require.NoError(t, os.MkdirAll(projectDir, 0755))

				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(projectDir, nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
				call := createClientMock.Calls[0]
				projectDir := call.ReturnArguments[0].(string)
				data, err := afero.ReadFile(cm.Fs, filepath.Join(projectDir, "manifest.json"))
				require.NoError(t, err)
				var result map[string]any
				require.NoError(t, json.Unmarshal(data, &result))
				displayInfo := result["display_information"].(map[string]any)
				assert.Equal(t, "Custom Name", displayInfo["name"])
			},
		},
		"app flag with no authenticated workspace returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)

				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"No workspaces connected"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with inaccessible app returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-test-token", TeamID: "T123", TeamDomain: "test-team", UserID: "U123"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "xoxp-test-token", []string{"A0123456789"}, "T123").
					Return(api.GetAppStatusResult{}, assert.AnError)

				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"No authenticated workspace has access to app A0123456789"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with manifest export failure returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-test-token", TeamID: "T123", TeamDomain: "test-team", UserID: "U123"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "xoxp-test-token", []string{"A0123456789"}, "T123").
					Return(api.GetAppStatusResult{}, nil)

				manifestMock := cf.AppClient().Manifest.(*app.ManifestMockObject)
				manifestMock.On("GetManifestRemote", mock.Anything, "xoxp-test-token", "A0123456789").
					Return(types.SlackYaml{}, assert.AnError)

				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"Failed to fetch manifest for app A0123456789"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func Test_resolveAuthForApp(t *testing.T) {
	tests := map[string]struct {
		setupAuth    func(*shared.ClientsMock)
		appID        string
		expectErr    bool
		expectTeamID string
	}{
		"returns first auth that has access": {
			setupAuth: func(cm *shared.ClientsMock) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "token-a", TeamID: "T001", TeamDomain: "team-a"},
					{Token: "token-b", TeamID: "T002", TeamDomain: "team-b"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "token-a", []string{"A111"}, "T001").
					Return(api.GetAppStatusResult{}, assert.AnError)
				cm.API.On("GetAppStatus", mock.Anything, "token-b", []string{"A111"}, "T002").
					Return(api.GetAppStatusResult{}, nil)
			},
			appID:        "A111",
			expectTeamID: "T002",
		},
		"returns error when no auths": {
			setupAuth: func(cm *shared.ClientsMock) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)
			},
			appID:     "A111",
			expectErr: true,
		},
		"returns error when no auth has access": {
			setupAuth: func(cm *shared.ClientsMock) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "token-a", TeamID: "T001", TeamDomain: "team-a"},
				}, nil)
				cm.API.On("GetAppStatus", mock.Anything, "token-a", []string{"A111"}, "T001").
					Return(api.GetAppStatusResult{}, assert.AnError)
			},
			appID:     "A111",
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cm := shared.NewClientsMock()
			tc.setupAuth(cm)
			cm.AddDefaultMocks()
			clients := shared.NewClientFactory(cm.MockClientFactory())

			auth, err := resolveAuthForApp(t.Context(), clients, tc.appID)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectTeamID, auth.TeamID)
			}
		})
	}
}

func Test_writeManifestToProject(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("test-project", 0755)

	manifest := types.SlackYaml{
		AppManifest: types.AppManifest{
			DisplayInformation: types.DisplayInformation{
				Name: "Test App",
			},
			Settings: &types.AppSettings{
				FunctionRuntime: types.Remote,
			},
		},
	}

	err := writeManifestToProject(fs, "test-project", manifest)
	require.NoError(t, err)

	data, err := afero.ReadFile(fs, "test-project/manifest.json")
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	displayInfo := result["display_information"].(map[string]any)
	assert.Equal(t, "Test App", displayInfo["name"])
}

func Test_linkAppToProject(t *testing.T) {
	tests := map[string]struct {
		runtime   types.FunctionRuntime
		expectDev bool
	}{
		"links as local for remote runtime": {
			runtime:   types.Remote,
			expectDev: true,
		},
		"links as local for local runtime": {
			runtime:   types.LocallyRun,
			expectDev: true,
		},
		"links as deployed for slack-hosted runtime": {
			runtime:   types.SlackHosted,
			expectDev: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			clients := shared.NewClientFactory(cm.MockClientFactory())

			appClientMock := &app.AppClientMock{}
			if tc.expectDev {
				appClientMock.On("SaveLocal", mock.Anything).Return(nil)
			} else {
				appClientMock.On("SaveDeployed", mock.Anything, mock.MatchedBy(func(a types.App) bool {
					return a.AppID == "A999" && a.TeamID == "T123" && !a.IsDev
				})).Return(nil)
			}
			clients.AppClient().AppClientInterface = appClientMock

			auth := types.SlackAuth{
				Token:      "xoxp-token",
				TeamID:     "T123",
				TeamDomain: "my-team",
				UserID:     "U456",
			}
			manifest := types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: tc.runtime,
					},
				},
			}

			err := linkAppToProject(t.Context(), clients, auth, "A999", manifest)
			require.NoError(t, err)

			if tc.expectDev {
				appClientMock.AssertCalled(t, "SaveLocal", mock.Anything)
			} else {
				appClientMock.AssertCalled(t, "SaveDeployed", mock.Anything, mock.Anything)
			}
		})
	}
}
