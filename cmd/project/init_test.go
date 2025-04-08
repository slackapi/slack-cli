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

package project

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockLinkAppID1 = "A001"

var mockLinkSlackAuth1 = types.SlackAuth{
	Token:        "xoxp-example1",
	TeamDomain:   "team1",
	TeamID:       "T001",
	EnterpriseID: "E001",
	UserID:       "U001",
}

var mockLinkSlackAuth2 = types.SlackAuth{
	Token:      "xoxp-example2",
	TeamDomain: "team2",
	TeamID:     "T002",
	UserID:     "U002",
}

func Test_Project_InitCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"requires bolt experiment": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Do not set experiment flag
				setupProjectInitCommandMocks(t, cm, cf, false)
			},
			ExpectedErrorStrings: []string{"Command requires the Bolt Framework experiment"},
		},
		"init a project and do not link an existing app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupProjectInitCommandMocks(t, cm, cf, true)
				// Do not link an existing app
				cm.IO.On("ConfirmPrompt", mock.Anything, app.LinkAppConfirmPromptText, mock.Anything).Return(false, nil)
			},
			ExpectedStdoutOutputs: []string{
				"Project Initialization",          // Assert section header
				"App Link",                        // Assert section header
				"Next steps to begin development", // Assert section header
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert installing project dependencies
				output := cm.GetCombinedOutput()
				require.Contains(t, output, "Installed project dependencies")
				require.Contains(t, output, "Added "+filepath.Join("project-name", ".slack"))
				require.Contains(t, output, "Added "+filepath.Join("project-name", ".slack", ".gitignore"))
				require.Contains(t, output, "Added "+filepath.Join("project-name", ".slack", "hooks.json"))
				require.Contains(t, output, "Updated config.json manifest source to local")
				// Assert prompt to add existing apps was called
				cm.IO.AssertCalled(
					t,
					"ConfirmPrompt",
					mock.Anything,
					app.LinkAppConfirmPromptText,
					mock.Anything,
				)
				// Assert Trace
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.ProjectInitSuccess,
					mock.Anything,
				)
				cm.IO.AssertNotCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.AppLinkStart,
					mock.Anything,
				)
			},
		},
		"init a project and link an existing app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Mocks auths to match against team and app
				cm.AuthInterface.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth2,
					mockLinkSlackAuth1,
				}, nil)
				// Default setup
				setupProjectInitCommandMocks(t, cm, cf, true)
				// Do not link an existing app
				cm.IO.On("ConfirmPrompt", mock.Anything, app.LinkAppConfirmPromptText, mock.Anything).Return(true, nil)
				// Mock prompt to link an existing app
				cm.IO.On("ConfirmPrompt", mock.Anything, app.LinkAppManifestSourceConfirmPromptText, mock.Anything).Return(true, nil)
				// Mock prompt for team
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Index:  1,
					Option: mockLinkSlackAuth2.TeamDomain,
				}, nil)
				// Mock prompt for App ID
				cm.IO.On("InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				).Return(mockLinkAppID1, nil)
				// Mock prompt for environment
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "local",
				}, nil)
				// Mock status of app for footer
				cm.ApiInterface.On("GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth2.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth2.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			ExpectedStdoutOutputs: []string{
				"Project Initialization",          // Assert section header
				"App Link",                        // Assert section header
				"Next steps to begin development", // Assert section header
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert prompt to add existing apps was called
				cm.IO.AssertCalled(t, "ConfirmPrompt",
					mock.Anything,
					app.LinkAppConfirmPromptText,
					mock.Anything,
				)
				// Assert prompt for team
				cm.IO.AssertCalled(t, "SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				// Assert prompt for app
				cm.IO.AssertCalled(t, "InputPrompt",
					mock.Anything,
					"Enter the existing app ID",
					mock.Anything,
				)
				// Assert prompt for environment
				cm.IO.AssertCalled(t, "SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				)
				// Assert trace
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.ProjectInitSuccess,
					mock.Anything,
				)
				// Assert app written to file
				expectedApp := types.App{
					AppID:      mockLinkAppID1,
					TeamDomain: mockLinkSlackAuth2.TeamDomain,
					TeamID:     mockLinkSlackAuth2.TeamID,
					IsDev:      true,
					UserID:     mockLinkSlackAuth2.UserID,
				}
				actualApp, err := cm.AppClient.GetLocal(
					context.Background(),
					mockLinkSlackAuth2.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewInitCommand(cf)
		return cmd
	})
}

// setupProjectInitCommandMocks prepares common mocks for these tests
func setupProjectInitCommandMocks(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory, boltExperimentEnabled bool) {
	// Mocks
	projectDirPath := "/path/to/project-name"
	cm.Os.On("Getwd").Return(projectDirPath, nil)
	cm.AddDefaultMocks()

	// Set experiment flag
	if boltExperimentEnabled {
		cm.Config.ExperimentsFlag = append(cm.Config.ExperimentsFlag, "bolt")
		cm.Config.LoadExperiments(context.Background(), cm.IO.PrintDebug)
	}

	// Create project directory
	if err := cm.Fs.MkdirAll(filepath.Dir(projectDirPath), 0755); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", projectDirPath))
	}
}
