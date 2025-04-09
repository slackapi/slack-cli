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

package manifest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestManifestCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"calls the manifest info command with a remote --app flag": {
			CmdArgs: []string{"--app", "A0001"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(
					prompts.SelectedApp{
						App:  types.App{AppID: "A001"},
						Auth: types.SlackAuth{Token: "xapp"},
					}, nil)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						DisplayInformation: types.DisplayInformation{
							Name: "app001",
						},
					},
				}, nil)
				cf.AppClient().Manifest = manifestMock
				cf.SDKConfig = hooks.NewSDKConfigMock()
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				mockManifest := types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "app001",
					},
				}
				manifest, err := json.MarshalIndent(mockManifest, "", "  ")
				require.NoError(t, err)
				assert.Equal(t, string(manifest)+"\n", cm.GetStdoutOutput())
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		return NewCommand(clients)
	})
}
