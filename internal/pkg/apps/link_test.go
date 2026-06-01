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

package apps

import (
	"encoding/json"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_FetchRemoteManifest(t *testing.T) {
	tests := map[string]struct {
		manifest  types.SlackYaml
		mockErr   error
		expectErr bool
	}{
		"returns manifest on success": {
			manifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{Name: "Test App"},
				},
			},
		},
		"preserves original error on failure": {
			mockErr:   assert.AnError,
			expectErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			clients := shared.NewClientFactory(cm.MockClientFactory())

			manifestMock := clients.AppClient().Manifest.(*app.ManifestMockObject)
			manifestMock.On("GetManifestRemote", mock.Anything, "token", "A123").
				Return(tc.manifest, tc.mockErr)

			result, err := FetchRemoteManifest(t.Context(), clients, "token", "A123")
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), assert.AnError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, "Test App", result.DisplayInformation.Name)
			}
		})
	}
}

func Test_WriteManifestToProject(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("test-project", 0755)

	manifest := types.SlackYaml{
		AppManifest: types.AppManifest{
			DisplayInformation: types.DisplayInformation{
				Name: "My App",
			},
			Settings: &types.AppSettings{
				FunctionRuntime: types.Remote,
			},
		},
	}

	err := WriteManifestToProject(fs, "test-project", manifest)
	require.NoError(t, err)

	data, err := afero.ReadFile(fs, "test-project/manifest.json")
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	displayInfo := result["display_information"].(map[string]any)
	assert.Equal(t, "My App", displayInfo["name"])
}

func Test_SaveAppToProject(t *testing.T) {
	tests := map[string]struct {
		app            types.App
		existingDeploy types.App
		existingLocal  types.App
		forceFlag      bool
		expectSave     string
		expectErr      bool
	}{
		"saves local dev app when no existing apps": {
			app:            types.App{AppID: "A1", TeamID: "T1", IsDev: true},
			existingDeploy: types.NewApp(),
			existingLocal:  types.NewApp(),
			expectSave:     "local",
		},
		"saves deployed app when no existing apps": {
			app:            types.App{AppID: "A1", TeamID: "T1", IsDev: false},
			existingDeploy: types.NewApp(),
			existingLocal:  types.NewApp(),
			expectSave:     "deployed",
		},
		"returns error when local app exists and same app in deploy": {
			app:            types.App{AppID: "A1", TeamID: "T1", IsDev: true},
			existingDeploy: types.App{AppID: "A1", TeamID: "T1"},
			existingLocal:  types.App{AppID: "A1", TeamID: "T1"},
			expectErr:      true,
		},
		"saves local with force when conflict exists": {
			app:            types.App{AppID: "A1", TeamID: "T1", IsDev: true},
			existingDeploy: types.App{AppID: "A1", TeamID: "T1"},
			existingLocal:  types.App{AppID: "A1", TeamID: "T1"},
			forceFlag:      true,
			expectSave:     "local",
		},
		"saves deployed with force when conflict exists": {
			app:            types.App{AppID: "A1", TeamID: "T1", IsDev: false},
			existingDeploy: types.App{AppID: "A1", TeamID: "T1"},
			existingLocal:  types.App{AppID: "A1", TeamID: "T1"},
			forceFlag:      true,
			expectSave:     "deployed",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			clients := shared.NewClientFactory(cm.MockClientFactory())
			clients.Config.ForceFlag = tc.forceFlag

			appClientMock := &app.AppClientMock{}
			appClientMock.On("GetDeployed", mock.Anything, tc.app.TeamID).Return(tc.existingDeploy, nil)
			appClientMock.On("GetLocal", mock.Anything, tc.app.TeamID).Return(tc.existingLocal, nil)
			appClientMock.On("SaveLocal", mock.Anything, mock.Anything).Return(nil)
			appClientMock.On("SaveDeployed", mock.Anything, mock.Anything).Return(nil)
			clients.AppClient().AppClientInterface = appClientMock

			err := SaveAppToProject(t.Context(), clients, tc.app)
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot be overwritten")
			} else {
				require.NoError(t, err)
				switch tc.expectSave {
				case "local":
					appClientMock.AssertCalled(t, "SaveLocal", mock.Anything, mock.Anything)
				case "deployed":
					appClientMock.AssertCalled(t, "SaveDeployed", mock.Anything, mock.Anything)
				}
			}
		})
	}
}

func Test_ResolveAuthForApp(t *testing.T) {
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

			auth, err := ResolveAuthForApp(t.Context(), clients, tc.appID)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectTeamID, auth.TeamID)
			}
		})
	}
}

func Test_LinkAppToProject(t *testing.T) {
	tests := map[string]struct {
		runtime     types.FunctionRuntime
		environment string
		expectDev   bool
	}{
		"infers local for remote runtime": {
			runtime:   types.Remote,
			expectDev: true,
		},
		"infers local for local runtime": {
			runtime:   types.LocallyRun,
			expectDev: true,
		},
		"infers deployed for slack-hosted runtime": {
			runtime:   types.SlackHosted,
			expectDev: false,
		},
		"environment flag deployed overrides manifest": {
			runtime:     types.Remote,
			environment: "deployed",
			expectDev:   false,
		},
		"environment flag local overrides manifest": {
			runtime:     types.SlackHosted,
			environment: "local",
			expectDev:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			clients := shared.NewClientFactory(cm.MockClientFactory())

			appClientMock := &app.AppClientMock{}
			appClientMock.On("GetDeployed", mock.Anything, "T123").Return(types.NewApp(), nil)
			appClientMock.On("GetLocal", mock.Anything, "T123").Return(types.NewApp(), nil)
			if tc.expectDev {
				appClientMock.On("SaveLocal", mock.Anything, mock.Anything).Return(nil)
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

			err := LinkAppToProject(t.Context(), clients, auth, "A999", manifest, tc.environment)
			require.NoError(t, err)

			if tc.expectDev {
				appClientMock.AssertCalled(t, "SaveLocal", mock.Anything, mock.Anything)
			} else {
				appClientMock.AssertCalled(t, "SaveDeployed", mock.Anything, mock.Anything)
			}
		})
	}
}

func Test_LinkAppToProject_invalidEnvironment(t *testing.T) {
	cm := shared.NewClientsMock()
	cm.AddDefaultMocks()
	clients := shared.NewClientFactory(cm.MockClientFactory())

	auth := types.SlackAuth{Token: "token", TeamID: "T1"}
	manifest := types.SlackYaml{}

	err := LinkAppToProject(t.Context(), clients, auth, "A1", manifest, "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local")
}
