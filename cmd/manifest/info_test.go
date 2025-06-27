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
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInfoCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors when the source is project and app id is set": {
			CmdArgs: []string{"--source", "local", "--app", "A0001"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig = hooks.NewSDKConfigMock()
			},
			ExpectedError: slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage("The \"--source\" flag must be \"remote\" when using \"--app\""),
		},
		"errors when the source is an unexpected value": {
			CmdArgs: []string{"--source", "paper"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig = hooks.NewSDKConfigMock()
				cm.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return("", nil)
			},
			ExpectedError: slackerror.New(slackerror.ErrInvalidFlag).
				WithMessage("The \"--source\" flag must be \"local\" or \"remote\""),
		},
		"gathers the --source local from the get-manifest hook": {
			CmdArgs: []string{"--source", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						DisplayInformation: types.DisplayInformation{
							Name: "app001",
						},
					},
				}, nil)
				cf.AppClient().Manifest = manifestMock
				cf.SDKConfig = hooks.NewSDKConfigMock()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
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
		"gathers the --source remote from the apps.manifest.export method": {
			CmdArgs: []string{"--source", "remote"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(
					prompts.SelectedApp{
						App:  types.App{AppID: "A001"},
						Auth: types.SlackAuth{Token: "xapp"}}, nil)
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						DisplayInformation: types.DisplayInformation{
							Name: "app002",
						},
					},
				}, nil)
				cf.AppClient().Manifest = manifestMock
				cf.SDKConfig = hooks.NewSDKConfigMock()
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				mockManifest := types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "app002",
					},
				}
				manifest, err := json.MarshalIndent(mockManifest, "", "  ")
				require.NoError(t, err)
				assert.Equal(t, string(manifest)+"\n", cm.GetStdoutOutput())
			},
		},
		"gathers manifest.source local from project configurations": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(
					prompts.SelectedApp{
						App:  types.App{AppID: "A004"},
						Auth: types.SlackAuth{Token: "xapp"}}, nil)
				cm.IO.AddDefaultMocks()
				cm.Os.AddDefaultMocks()
				cf.SDKConfig.WorkingDirectory = "."
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						DisplayInformation: types.DisplayInformation{
							Name: "app002",
						},
					},
				}, nil)
				cf.AppClient().Manifest = manifestMock
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
				cm.Config.ProjectConfig = mockProjectConfig
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				mockManifest := types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "app002",
					},
				}
				manifest, err := json.MarshalIndent(mockManifest, "", "  ")
				require.NoError(t, err)
				assert.Equal(t, string(manifest)+"\n", cm.GetStdoutOutput())
			},
		},
		"gathers manifest.source remote from project configurations": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
				cm.IO.AddDefaultMocks()
				cm.Os.AddDefaultMocks()
				mockProjectConfig := config.NewProjectConfigMock()
				mockProjectConfig.On("GetManifestSource", mock.Anything).
					Return(config.ManifestSource(config.ManifestSourceRemote), nil)
				cm.Config.ProjectConfig = mockProjectConfig
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						DisplayInformation: types.DisplayInformation{
							Name: "app005",
						},
					},
				}, nil)
				cf.AppClient().Manifest = manifestMock
				appSelectMock := prompts.NewAppSelectMock()
				appSelectPromptFunc = appSelectMock.AppSelectPrompt
				appSelectMock.On("AppSelectPrompt").Return(prompts.SelectedApp{
					App:  types.App{AppID: "A005"},
					Auth: types.SlackAuth{Token: "xapp-example-005"},
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				mockManifest := types.AppManifest{
					DisplayInformation: types.DisplayInformation{
						Name: "app005",
					},
				}
				manifest, err := json.MarshalIndent(mockManifest, "", "  ")
				require.NoError(t, err)
				assert.Equal(t, string(manifest)+"\n", cm.GetStdoutOutput())
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		return NewInfoCommand(clients)
	})
}
