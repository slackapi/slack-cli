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

package platform

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/platform"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Setup a mock for the package
type RunCmdMock struct {
	mock.Mock
}

func (m *RunCmdMock) RunRunCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	m.Called()
	return nil
}

type RunPkgMock struct {
	mock.Mock
}

func (m *RunPkgMock) Run(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, runArgs platform.RunArgs) (*logger.LogEvent, types.InstallState, error) {
	args := m.Called(ctx, clients, log, runArgs)
	return log.SuccessEvent(), types.InstallSuccess, args.Error(0)
}

func TestRunCommand_Flags(t *testing.T) {
	tests := map[string]struct {
		cmdArgs         []string
		appFlag         string
		tokenFlag       string
		selectedAppAuth prompts.SelectedApp
		selectedAppErr  error
		expectedRunArgs platform.RunArgs
		expectedErr     error
	}{
		"New app with org auth": {
			cmdArgs: []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123", "--hide-triggers"},
			selectedAppAuth: prompts.SelectedApp{
				App:  types.NewApp(),
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			expectedRunArgs: platform.RunArgs{
				Activity:            true,
				ActivityLevel:       "info",
				Auth:                types.SlackAuth{IsEnterpriseInstall: true},
				App:                 types.NewApp(),
				Cleanup:             false,
				ShowTriggers:        false,
				OrgGrantWorkspaceID: "T123", // Flag passed through
			},
			expectedErr: nil,
		},
		"Uninstalled org app": {
			cmdArgs: []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			selectedAppAuth: prompts.SelectedApp{
				App:  types.App{InstallStatus: types.AppStatusUninstalled, TeamID: "E123", EnterpriseID: "E123"},
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			expectedRunArgs: platform.RunArgs{
				Activity:            true,
				ActivityLevel:       "info",
				Auth:                types.SlackAuth{IsEnterpriseInstall: true},
				App:                 types.App{InstallStatus: types.AppStatusUninstalled, TeamID: "E123", EnterpriseID: "E123"},
				Cleanup:             false,
				ShowTriggers:        true,
				OrgGrantWorkspaceID: "T123", // Flag passed through
			},
			expectedErr: nil,
		},
		"Installed org app; can't set different workspace grant": {
			cmdArgs: []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			selectedAppAuth: prompts.SelectedApp{
				App: types.App{
					AppID:            "A123",
					InstallStatus:    types.AppStatusInstalled,
					TeamID:           "E123",
					EnterpriseGrants: []types.EnterpriseGrant{{WorkspaceID: "T0"}}},
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			expectedErr: slackerror.New(slackerror.ErrOrgGrantExists).
				WithMessage("A different org workspace grant already exists for installed app 'A123'\n   Workspace Grant: T0"),
		},
		"Installed org app; can pass same workspace grant": {
			cmdArgs: []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			selectedAppAuth: prompts.SelectedApp{
				App: types.App{
					InstallStatus:    types.AppStatusInstalled,
					TeamID:           "E123",
					EnterpriseGrants: []types.EnterpriseGrant{{WorkspaceID: "T123"}}},
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			expectedRunArgs: platform.RunArgs{
				Activity:      true,
				ActivityLevel: "info",
				Auth:          types.SlackAuth{IsEnterpriseInstall: true},
				App: types.App{
					InstallStatus:    types.AppStatusInstalled,
					TeamID:           "E123",
					EnterpriseGrants: []types.EnterpriseGrant{{WorkspaceID: "T123"}}},
				Cleanup:             false,
				ShowTriggers:        true,
				OrgGrantWorkspaceID: "T123", // Flag passed through
			},
		},
		"Standalone workspace app": {
			cmdArgs: []string{"--" + cmdutil.OrgGrantWorkspaceFlag, "T123"},
			selectedAppAuth: prompts.SelectedApp{
				App: types.App{
					IsDev:         true,
					InstallStatus: types.AppStatusUninstalled,
					TeamID:        "T123",
				},
				Auth: types.SlackAuth{IsEnterpriseInstall: true},
			},
			expectedRunArgs: platform.RunArgs{
				Activity:      true,
				ActivityLevel: "info",
				Auth:          types.SlackAuth{IsEnterpriseInstall: true},
				App: types.App{
					IsDev:         true,
					InstallStatus: types.AppStatusUninstalled,
					TeamID:        "T123",
				},
				Cleanup:             false,
				ShowTriggers:        true,
				OrgGrantWorkspaceID: "", // Flag not passed through
			},
			expectedErr: nil,
		},
		"Set the selected app to development when selecting with flags": {
			appFlag:   "A123",
			tokenFlag: "T123",
			selectedAppAuth: prompts.SelectedApp{
				App:  types.App{AppID: "A123"},
				Auth: types.SlackAuth{TeamID: "T123"},
			},
			selectedAppErr: slackerror.New(slackerror.ErrDeployedAppNotSupported),
			expectedRunArgs: platform.RunArgs{
				Activity:      true,
				ActivityLevel: "info",
				Auth:          types.SlackAuth{TeamID: "T123"},
				App: types.App{
					AppID: "A123",
					IsDev: true,
				},
				Cleanup:      false,
				ShowTriggers: true,
			},
		},
		"Error if interrupted during app selection": {
			selectedAppErr: slackerror.New(slackerror.ErrProcessInterrupted),
			expectedRunArgs: platform.RunArgs{
				Activity:      true,
				ActivityLevel: "info",
				Cleanup:       false,
				ShowTriggers:  true,
			},
			expectedErr: slackerror.New(slackerror.ErrProcessInterrupted),
		},
		"Error if no apps are available when using a remote manifest source": {
			selectedAppErr: slackerror.New(slackerror.ErrMissingOptions),
			expectedErr: slackerror.New(slackerror.ErrAppNotFound).
				WithMessage("No apps are available for selection").
				WithRemediation(
					"Create a new app on app settings: %s\nThen add the app to this project with %s",
					style.LinkText("https://api.slack.com/apps"),
					style.Commandf("app link", false),
				).
				WithDetails(slackerror.ErrorDetails{
					slackerror.ErrorDetail{
						Code:    slackerror.ErrProjectConfigManifestSource,
						Message: "App manifests for this project are sourced from app settings",
					},
				}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("IsTTY").Return(true)
			clientsMock.IO.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
				clients.Config.AppFlag = tt.appFlag
				clients.Config.TokenFlag = tt.tokenFlag
			})

			appSelectMock := prompts.NewAppSelectMock()
			appSelectMock.On("TeamAppSelectPrompt").Return(tt.selectedAppAuth, tt.selectedAppErr)
			runTeamAppSelectPromptFunc = appSelectMock.TeamAppSelectPrompt

			runPkgMock := new(RunPkgMock)
			runFunc = runPkgMock.Run
			runPkgMock.On("Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			cmd := NewRunCommand(clients)
			testutil.MockCmdIO(clients.IO, cmd)
			cmd.SetArgs(tt.cmdArgs)

			// Execute
			err := cmd.ExecuteContext(ctx)

			// Check args passed into the run function
			if tt.expectedErr == nil {
				assert.NoError(t, err)
				runPkgMock.AssertCalled(t, "Run", mock.Anything, mock.Anything, mock.Anything,
					tt.expectedRunArgs,
				)
			} else {
				assert.Equal(t, tt.expectedErr, slackerror.ToSlackError(err))
			}
		})
	}
}

func TestRunCommand_Help(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewRunCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--help"})

	runPkgMock := new(RunPkgMock)
	runFunc = runPkgMock.Run
	runPkgMock.On("Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := cmd.ExecuteContext(ctx)
	assert.NoError(t, err)
	runPkgMock.AssertNotCalled(t, "Run")

	assert.Contains(t, clientsMock.GetStdoutOutput(), "activity level to display (default \"info\")")
}
