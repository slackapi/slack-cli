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

package project

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	internalApp "github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
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
		"init a project and do not link an existing app": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Prompts are possible but none are expected
				cm.IO.On("IsTTY").Return(true)
				setupProjectInitCommandMocks(t, ctx, cm, cf)
			},
			ExpectedStdoutOutputs: []string{
				"Project Initialization",                 // Assert section header
				"Next Steps",                             // Assert section header
				"Add existing apps to your project with", // Assert link hint
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert installing project dependencies
				output := cm.GetCombinedOutput()
				require.Contains(t, output, "Project Dependencies")
				require.Contains(t, output, "Added "+filepath.Join("project-name", ".slack"))
				require.Contains(t, output, "Added "+filepath.Join("project-name", ".slack", ".gitignore"))
				require.Contains(t, output, "Added "+filepath.Join("project-name", ".slack", "hooks.json"))
				require.Contains(t, output, `Updated app manifest source to "project" (local)`)
				// Assert no prompts to add existing apps
				require.NotContains(t, output, "App Link")
				cm.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, mock.Anything, mock.Anything)
				cm.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				// Assert Trace
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.ProjectInitSuccess,
					mock.Anything,
				)
			},
		},
		"init a project and link an existing app from flags": {
			CmdArgs: []string{"--app", mockLinkAppID1, "--team", mockLinkSlackAuth2.TeamID, "--environment", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Mocks auths to match against team and app
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth2,
					mockLinkSlackAuth1,
				}, nil)
				// Default setup
				setupProjectInitCommandMocks(t, ctx, cm, cf)
				// Mock team prompt answered with the team flag
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Select the existing app team",
					mock.Anything,
					mock.MatchedBy(func(cfg iostreams.SelectPromptConfig) bool {
						return cfg.Flag != nil && cfg.Flag.Name == "team" && cfg.Flag.Value.String() == mockLinkSlackAuth2.TeamID
					}),
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: mockLinkSlackAuth2.TeamID,
				}, nil)
				// Mock environment prompt answered with the environment flag
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.MatchedBy(func(cfg iostreams.SelectPromptConfig) bool {
						return cfg.Flag != nil && cfg.Flag.Name == "environment" && cfg.Flag.Value.String() == "local"
					}),
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "local",
				}, nil)
				// Mock status of app for footer
				cm.API.On("GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth2.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth2.TeamID,
				).Return(api.GetAppStatusResult{}, nil)
			},
			ExpectedStdoutOutputs: []string{
				"Project Initialization", // Assert section header
				"App Link",               // Assert section header
				"Next Steps",             // Assert section header
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert no prompts for input
				cm.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, mock.Anything, mock.Anything)
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
					ctx,
					mockLinkSlackAuth2.TeamID,
				)
				require.NoError(t, err)
				assert.Equal(t, expectedApp, actualApp)
			},
		},
		"init a project and continue after failing to link an existing app": {
			CmdArgs: []string{"--app", mockLinkAppID1, "--environment", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Mocks a single auth that is selected without a prompt
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					mockLinkSlackAuth1,
				}, nil)
				// Default setup
				setupProjectInitCommandMocks(t, ctx, cm, cf)
				// Mock environment prompt answered with the environment flag
				cm.IO.On("SelectPrompt",
					mock.Anything,
					"Choose the app environment",
					mock.Anything,
					mock.Anything,
				).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "local",
				}, nil)
				// Mock an app that cannot be found
				cm.API.On("GetAppStatus",
					mock.Anything,
					mockLinkSlackAuth1.Token,
					[]string{mockLinkAppID1},
					mockLinkSlackAuth1.TeamID,
				).Return(api.GetAppStatusResult{}, slackerror.New(slackerror.ErrAppNotFound))
			},
			ExpectedOutputs: []string{
				slackerror.ErrAppNotFound, // Assert error is displayed
			},
			ExpectedStdoutOutputs: []string{
				"App Link",   // Assert section header
				"Next Steps", // Assert section header
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert init still succeeds
				cm.IO.AssertCalled(
					t,
					"PrintTrace",
					mock.Anything,
					slacktrace.ProjectInitSuccess,
					mock.Anything,
				)
				// Assert no app written to file
				actualApp, err := cm.AppClient.GetLocal(
					ctx,
					mockLinkSlackAuth1.TeamID,
				)
				require.NoError(t, err)
				assert.True(t, actualApp.IsNew())
			},
		},
		"environment flag without app flag returns error": {
			CmdArgs: []string{"--environment", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupProjectInitCommandMocks(t, ctx, cm, cf)
			},
			ExpectedErrorStrings: []string{"The --environment flag requires the --app flag when used with init"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Assert the project is not initialized
				cm.Os.AssertNotCalled(t, "Getwd")
				require.NotContains(t, cm.GetCombinedOutput(), "Project Initialization")
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewInitCommand(cf)
		return cmd
	})
}

// setupProjectInitCommandMocks prepares common mocks for these tests
func setupProjectInitCommandMocks(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	// Mocks
	projectDirPath := "/path/to/project-name"
	cm.Os.On("Getwd").Return(projectDirPath, nil)
	cm.AddDefaultMocks()

	// Create project directory
	if err := cm.Fs.MkdirAll(filepath.Dir(projectDirPath), 0755); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", projectDirPath))
	}

	// Mock manifest
	manifestMock := &internalApp.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{}, nil)
	cf.AppClient().Manifest = manifestMock
}
