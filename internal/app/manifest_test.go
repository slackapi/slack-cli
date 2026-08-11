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

package app

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_AppManifest_SetManifestEnvTeamVars(t *testing.T) {
	tests := map[string]struct {
		teamDomain string
		isDev      bool
		manifest   map[string]string
		expected   map[string]string
	}{
		"workspace and prod environment is set": {
			teamDomain: "bigspeck",
			isDev:      false,
			manifest:   nil,
			expected: map[string]string{
				"SLACK_WORKSPACE": "bigspeck",
				"SLACK_ENV":       "deployed",
			},
		},
		"workspace and local environment is set": {
			teamDomain: "sandbox",
			isDev:      true,
			manifest:   map[string]string{"SLACK_APP_ID": "A1234"},
			expected: map[string]string{
				"SLACK_APP_ID":    "A1234",
				"SLACK_WORKSPACE": "sandbox",
				"SLACK_ENV":       "local",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			teamManifest := SetManifestEnvTeamVars(tc.manifest, tc.teamDomain, tc.isDev)
			require.Equal(t, len(tc.expected), len(teamManifest))
			for key, val := range tc.expected {
				require.Equal(t, val, teamManifest[key])
			}
		})
	}
}

func Test_AppManifest_GetManifestLocal(t *testing.T) {
	t.Run("uses hook when get-manifest is available", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		configMock := config.NewConfig(fsMock, osMock)
		configMock.DomainAuthTokens = "api.slack.com"
		mockSDKConfig := hooks.NewSDKConfigMock()
		mockSDKConfig.WorkingDirectory = "/project"
		mockSDKConfig.Hooks.GetManifest = hooks.HookScript{Name: "GetManifest", Command: "echo manifest"}

		_ = fsMock.MkdirAll("/project", 0755)
		_ = afero.WriteFile(fsMock, "/project/manifest.json", []byte(`{"display_information":{"name":"file-app"}}`), 0644)

		mockHookExecutor := &hooks.MockHookExecutor{}
		mockHookExecutor.On("Execute", mock.Anything, mock.Anything).
			Return(`{"display_information":{"name":"hook-app"}}`, nil)
		manifestClient := NewManifestClient(&api.APIMock{}, configMock, fsMock)

		result, err := manifestClient.GetManifestLocal(ctx, mockSDKConfig, mockHookExecutor)
		require.NoError(t, err)
		assert.Equal(t, "hook-app", result.DisplayInformation.Name)
		mockHookExecutor.AssertCalled(t, "Execute", mock.Anything, mock.Anything)
	})

	t.Run("falls back to manifest.json when no hook exists", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		configMock := config.NewConfig(fsMock, osMock)
		mockSDKConfig := hooks.NewSDKConfigMock()
		mockSDKConfig.WorkingDirectory = "/project"
		mockSDKConfig.Hooks.GetManifest = hooks.HookScript{Name: "GetManifest"}

		_ = fsMock.MkdirAll("/project", 0755)
		_ = afero.WriteFile(fsMock, "/project/manifest.json", []byte(`{"display_information":{"name":"file-app"}}`), 0644)

		mockHookExecutor := &hooks.MockHookExecutor{}
		manifestClient := NewManifestClient(&api.APIMock{}, configMock, fsMock)

		result, err := manifestClient.GetManifestLocal(ctx, mockSDKConfig, mockHookExecutor)
		require.NoError(t, err)
		assert.Equal(t, "file-app", result.DisplayInformation.Name)
		mockHookExecutor.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	})

	t.Run("errors if no hook and no manifest.json", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		configMock := config.NewConfig(fsMock, osMock)
		mockSDKConfig := hooks.NewSDKConfigMock()
		mockSDKConfig.WorkingDirectory = "/project"
		mockSDKConfig.Hooks.GetManifest = hooks.HookScript{Name: "GetManifest"}

		mockHookExecutor := &hooks.MockHookExecutor{}
		manifestClient := NewManifestClient(&api.APIMock{}, configMock, fsMock)

		_, err := manifestClient.GetManifestLocal(ctx, mockSDKConfig, mockHookExecutor)
		require.Error(t, err)
		assert.Equal(t, slackerror.ErrInvalidManifest, err.(*slackerror.Error).Code)
	})

	t.Run("errors if manifest.json contains invalid JSON", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		configMock := config.NewConfig(fsMock, osMock)
		mockSDKConfig := hooks.NewSDKConfigMock()
		mockSDKConfig.WorkingDirectory = "/project"
		mockSDKConfig.Hooks.GetManifest = hooks.HookScript{Name: "GetManifest"}

		_ = fsMock.MkdirAll("/project", 0755)
		_ = afero.WriteFile(fsMock, "/project/manifest.json", []byte(`not json`), 0644)

		mockHookExecutor := &hooks.MockHookExecutor{}
		manifestClient := NewManifestClient(&api.APIMock{}, configMock, fsMock)

		_, err := manifestClient.GetManifestLocal(ctx, mockSDKConfig, mockHookExecutor)
		require.Error(t, err)
		assert.Equal(t, slackerror.ErrInvalidManifest, err.(*slackerror.Error).Code)
	})

	hookTests := map[string]struct {
		hookOutput   string
		hookErr      error
		expectedName string
		expectedErr  string
	}{
		"returns manifest from hook output": {
			hookOutput:   `{"display_information":{"name":"hook-app"}}`,
			expectedName: "hook-app",
		},
		"parses hook output with leading characters": {
			hookOutput:   `...{"display_information":{"name":"hook-app"}}`,
			expectedName: "hook-app",
		},
		"errors if hook execution errors": {
			hookOutput:  `{}`,
			hookErr:     slackerror.New(slackerror.ErrNoFile),
			expectedErr: slackerror.ErrInvalidManifest,
		},
		"errors if hook output has no JSON": {
			hookOutput:  `...unknown`,
			expectedErr: slackerror.ErrInvalidManifest,
		},
	}
	for name, tc := range hookTests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			configMock := config.NewConfig(fsMock, osMock)
			configMock.DomainAuthTokens = "api.slack.com"
			mockSDKConfig := hooks.NewSDKConfigMock()
			mockSDKConfig.Hooks.GetManifest = hooks.HookScript{Name: "GetManifest", Command: "generate-manifest"}

			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).
				Return(tc.hookOutput, tc.hookErr)

			manifestClient := NewManifestClient(&api.APIMock{}, configMock, fsMock)

			result, err := manifestClient.GetManifestLocal(ctx, mockSDKConfig, mockHookExecutor)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.(*slackerror.Error).Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedName, result.DisplayInformation.Name)
			}
		})
	}
}

func Test_AppManifest_GetManifestRemote(t *testing.T) {
	tests := map[string]struct {
		mockAppID            string
		mockToken            string
		mockManifestResponse types.SlackYaml
		mockManifestError    error
		expectedManifest     types.SlackYaml
		expectedError        error
	}{
		"returns the manifest from a successful api response": {
			mockAppID: "A0123",
			mockToken: "xoxb-example",
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "slackbot",
					}},
			},
			expectedManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "slackbot",
					}},
			},
		},
		"errors if the api response returns an error": {
			mockAppID:         "A0123",
			mockToken:         "xoxb-broken",
			mockManifestError: slackerror.New(slackerror.ErrAppManifestAccess),
			expectedError:     slackerror.New(slackerror.ErrAppManifestAccess),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			configMock := config.NewConfig(fsMock, osMock)
			apic := &api.APIMock{}
			apic.On("ExportAppManifest", mock.Anything, mock.Anything, mock.Anything).
				Return(api.ExportAppResult{Manifest: tc.mockManifestResponse}, tc.mockManifestError)
			manifestClient := NewManifestClient(apic, configMock, fsMock)

			manifest, err := manifestClient.GetManifestRemote(ctx, tc.mockToken, tc.mockAppID)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedManifest, manifest)
				apic.AssertCalled(t, "ExportAppManifest", mock.Anything, tc.mockToken, tc.mockAppID)
			}
		})
	}
}
