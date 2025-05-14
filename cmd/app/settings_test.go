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
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func Test_App_SettingsCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"requires a valid project directory": {
			ExpectedError: slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
		"errors for rosi applications": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
				projectConfigMock := config.NewProjectConfigMock()
				projectConfigMock.On(
					"GetManifestSource",
					mock.Anything,
				).Return(
					config.ManifestSourceLocal,
					nil,
				)
				cm.Config.ProjectConfig = projectConfigMock
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On(
					"GetManifestLocal",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					types.SlackYaml{
						AppManifest: types.AppManifest{
							Settings: &types.AppSettings{FunctionRuntime: types.SlackHosted},
						},
					},
					nil,
				)
				cm.AppClient.Manifest = manifestMock
			},
			ExpectedError: slackerror.New(slackerror.ErrAppHosted),
		},
		"requires an existing application": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
				projectConfigMock := config.NewProjectConfigMock()
				projectConfigMock.On(
					"GetManifestSource",
					mock.Anything,
				).Return(
					config.ManifestSourceRemote,
					nil,
				)
				cm.Config.ProjectConfig = projectConfigMock
				appSelectMock := prompts.NewAppSelectMock()
				appSelectMock.On(
					"AppSelectPrompt",
				).Return(
					prompts.SelectedApp{},
					slackerror.New(slackerror.ErrInstallationRequired),
				)
				settingsAppSelectPromptFunc = appSelectMock.AppSelectPrompt
			},
			ExpectedError: slackerror.New(slackerror.ErrInstallationRequired),
		},
		"opens the url to app settings of an app in production": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cf.SDKConfig.WorkingDirectory = "."
				projectConfigMock := config.NewProjectConfigMock()
				projectConfigMock.On(
					"GetManifestSource",
					mock.Anything,
				).Return(
					config.ManifestSourceRemote,
					nil,
				)
				cm.Config.ProjectConfig = projectConfigMock
				appSelectMock := prompts.NewAppSelectMock()
				appSelectMock.On(
					"AppSelectPrompt",
				).Return(
					prompts.SelectedApp{App: types.App{AppID: "A0123456789"}},
					nil,
				)
				settingsAppSelectPromptFunc = appSelectMock.AppSelectPrompt
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://api.slack.com/apps/A0123456789"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.AppSettingsStart, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.AppSettingsSuccess, []string{expectedURL})
			},
		},
		"opens the url to app settings of an app in development": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				host := "https://dev1234.slack.com"
				cf.SDKConfig.WorkingDirectory = "."
				projectConfigMock := config.NewProjectConfigMock()
				projectConfigMock.On(
					"GetManifestSource",
					mock.Anything,
				).Return(
					config.ManifestSourceRemote,
					nil,
				)
				cm.Config.ProjectConfig = projectConfigMock
				appSelectMock := prompts.NewAppSelectMock()
				appSelectMock.On(
					"AppSelectPrompt",
				).Return(
					prompts.SelectedApp{
						App:  types.App{AppID: "A0123456789"},
						Auth: types.SlackAuth{APIHost: &host},
					},
					nil,
				)
				settingsAppSelectPromptFunc = appSelectMock.AppSelectPrompt
				cm.API.On("Host").Return(host) // SetHost is implemented in AppSelectPrompt
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://api.dev1234.slack.com/apps/A0123456789"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.AppSettingsStart, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.AppSettingsSuccess, []string{expectedURL})
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewSettingsCommand(cf)
	})
}
