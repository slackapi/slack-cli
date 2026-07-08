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

package manifest

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_SetManifestLocal(t *testing.T) {
	testManifest := types.SlackYaml{
		AppManifest: types.AppManifest{
			DisplayInformation: types.DisplayInformation{Name: "Test App"},
		},
	}

	tests := map[string]struct {
		setupMocks  func(*shared.ClientsMock, *shared.ClientFactory)
		expectError bool
		assertFile  bool
	}{
		"writes manifest and saves hash on success": {
			setupMocks: func(clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, "xoxp-token", "A001").
					Return(testManifest, nil)
				clients.AppClient().Manifest = manifestMock

				cc := cache.NewCacheMock()
				cc.On("NewManifestHash", mock.Anything, testManifest.AppManifest).
					Return(cache.Hash("abc123"), nil)
				cc.On("SetManifestHash", mock.Anything, "A001", cache.Hash("abc123")).
					Return(nil)

				proj := config.NewProjectConfigMock()
				proj.On("Cache").Return(cc)
				clients.Config.ProjectConfig = proj
			},
			expectError: false,
			assertFile:  true,
		},
		"returns error when GetManifestRemote fails": {
			setupMocks: func(clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
					Return(types.SlackYaml{}, slackerror.New("api error"))
				clients.AppClient().Manifest = manifestMock
			},
			expectError: true,
			assertFile:  false,
		},
		"returns error when file write fails": {
			setupMocks: func(clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
					Return(testManifest, nil)
				clients.AppClient().Manifest = manifestMock

				clients.Fs = afero.NewReadOnlyFs(&afero.MemMapFs{})
			},
			expectError: true,
			assertFile:  false,
		},
		"returns error when SetManifestHash fails": {
			setupMocks: func(clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
					Return(testManifest, nil)
				clients.AppClient().Manifest = manifestMock

				cc := cache.NewCacheMock()
				cc.On("NewManifestHash", mock.Anything, mock.Anything).
					Return(cache.Hash("abc123"), nil)
				cc.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
					Return(slackerror.New("cache write error"))

				proj := config.NewProjectConfigMock()
				proj.On("Cache").Return(cc)
				clients.Config.ProjectConfig = proj
			},
			expectError: true,
			assertFile:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(c *shared.ClientFactory) {
				c.SDKConfig = hooks.NewSDKConfigMock()
			})

			tc.setupMocks(clientsMock, clients)

			err := SetManifestLocal(ctx, clients, "xoxp-token", "A001", "/project")

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.assertFile {
				content, readErr := afero.ReadFile(clients.Fs, "/project/manifest.json")
				require.NoError(t, readErr)

				expected, _ := goutils.JSONMarshalUnescapedIndent(testManifest.AppManifest)
				assert.Equal(t, expected, string(content))
			}
		})
	}
}
