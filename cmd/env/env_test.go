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
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Env_Command(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"shows the help page without commands or arguments or flags": {
			ExpectedStdoutOutputs: []string{
				"Add an environment variable",
				"List all environment variables",
				"Remove an environment variable",
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCommand(clients)
		return cmd
	})
}

func Test_isHostedRuntime(t *testing.T) {
	tests := map[string]struct {
		mockManifest types.SlackYaml
		mockError    error
		expected     bool
	}{
		"returns true for slack hosted runtime": {
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.SlackHosted,
					},
				},
			},
			expected: true,
		},
		"returns true for local runtime": {
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.LocallyRun,
					},
				},
			},
			expected: true,
		},
		"returns false for remote runtime": {
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.Remote,
					},
				},
			},
			expected: false,
		},
		"returns false for empty runtime": {
			mockManifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{},
				},
			},
			expected: false,
		},
		"returns false when manifest fetch fails": {
			mockError: slackerror.New(slackerror.ErrSDKHookInvocationFailed),
			expected:  false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(tc.mockManifest, tc.mockError)
			clientsMock.AppClient.Manifest = manifestMock
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			assert.Equal(t, tc.expected, isHostedRuntime(ctx, clients))
		})
	}
}
