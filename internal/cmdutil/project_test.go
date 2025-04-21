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

package cmdutil

import (
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsSlackHostedProject(t *testing.T) {
	tests := map[string]struct {
		mockManifestResponse types.SlackYaml
		mockManifestError    error
		mockManifestSource   config.ManifestSource
		expectedError        error
	}{
		"continues if the project has a slack hosted function runtime": {
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.SLACK_HOSTED,
					},
				},
			},
			mockManifestError:  nil,
			mockManifestSource: config.MANIFEST_SOURCE_LOCAL,
			expectedError:      nil,
		},
		"errors if the project does not have a slack function runtime": {
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.REMOTE,
					},
				},
			},
			mockManifestError:  nil,
			mockManifestSource: config.MANIFEST_SOURCE_LOCAL,
			expectedError:      slackerror.New(slackerror.ErrAppNotHosted),
		},
		"errors if the project manifest cannot be gathered from hook": {
			mockManifestResponse: types.SlackYaml{},
			mockManifestError:    slackerror.New(slackerror.ErrSDKHookInvocationFailed),
			mockManifestSource:   config.MANIFEST_SOURCE_LOCAL,
			expectedError:        slackerror.New(slackerror.ErrSDKHookInvocationFailed),
		},
		"errors if the manifest source is configured to the remote": {
			mockManifestSource: config.MANIFEST_SOURCE_REMOTE,
			expectedError: slackerror.New(slackerror.ErrAppNotHosted).
				WithDetails(slackerror.ErrorDetails{
					{
						Code:        slackerror.ErrInvalidManifestSource,
						Message:     fmt.Sprintf("Slack hosted projects use \"%s\" manifest source", config.MANIFEST_SOURCE_LOCAL),
						Remediation: fmt.Sprintf("This value can be changed in configuration: \"%s\"", config.GetProjectConfigJSONFilePath("")),
					},
				}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On(
				"GetManifestLocal",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestResponse,
				tt.mockManifestError,
			)
			clientsMock.AppClient.Manifest = manifestMock
			projectConfigMock := config.NewProjectConfigMock()
			projectConfigMock.On("GetManifestSource", mock.Anything).Return(tt.mockManifestSource, nil)
			clientsMock.Config.ProjectConfig = projectConfigMock
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			err := IsSlackHostedProject(ctx, clients)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestIsValidProjectDirectory(t *testing.T) {
	tests := map[string]struct {
		mockWorkingDirectory string
		expectedError        error
	}{
		"continues if the project has the sdk working directory set": {
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        nil,
		},
		"errors if the process does not have a working directory": {
			mockWorkingDirectory: "",
			expectedError:        slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = tt.mockWorkingDirectory
			})
			err := IsValidProjectDirectory(clients)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
