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
	tests := map[string]struct {
		mockManifestInfo string
		mockManifestErr  error
		expectedErr      error
		expectedManifest types.SlackYaml
	}{
		"errors if no get-manifest hook exists": {
			expectedErr: slackerror.New(slackerror.ErrSDKHookNotFound),
		},
		"returns an existing manifest without errors": {
			mockManifestInfo: `{"display_information":{"name":"my-example-app"}}`,
			expectedManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "my-example-app",
					},
				},
			},
		},
		"errors if the hook execution errors": {
			mockManifestInfo: `{}`,
			mockManifestErr:  slackerror.New(slackerror.ErrNoFile),
			expectedErr:      slackerror.New(slackerror.ErrInvalidManifest),
		},
		"parses a manifest with random leading characters": {
			mockManifestInfo: `...{"display_information":{"name":"my-showcased-app"}}`,
			expectedManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "my-showcased-app",
					},
				},
			},
		},
		"errors if a manifest is not present in output": {
			mockManifestInfo: `...unknown`,
			expectedErr:      slackerror.New(slackerror.ErrInvalidManifest),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			mockManifestEnv := map[string]string{"EXAMPLE": "12"}
			mockSDKConfig := hooks.NewSDKConfigMock()
			mockHookExecutor := &hooks.MockHookExecutor{}
			if tc.mockManifestInfo != "" {
				mockSDKConfig.Hooks.GetManifest = hooks.HookScript{
					Name:    "GetManifest",
					Command: "cat manifest.json",
				}
				mockHookExecutor.On("Execute", mock.Anything, mock.Anything).
					Return(tc.mockManifestInfo, tc.mockManifestErr)
			} else {
				mockSDKConfig.Hooks.GetManifest = hooks.HookScript{Name: "GetManifest"}
			}
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			configMock := config.NewConfig(fsMock, osMock)
			configMock.DomainAuthTokens = "api.slack.com"
			configMock.ManifestEnv = mockManifestEnv
			manifestClient := NewManifestClient(&api.APIMock{}, configMock)

			actualManifest, err := manifestClient.GetManifestLocal(ctx, mockSDKConfig, mockHookExecutor)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t,
					tc.expectedErr.(*slackerror.Error).Code, err.(*slackerror.Error).Code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedManifest, actualManifest)
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
			manifestClient := NewManifestClient(apic, configMock)

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
