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

package sandbox

import (
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"empty list": {
			CmdArgs: []string{"--experiment=sandboxes", "--token", "xoxb-test-token"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return([]types.Sandbox{}, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"No sandboxes found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Auth.AssertCalled(t, "AuthWithToken", mock.Anything, "xoxb-test-token")
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "")
			},
		},
		"with active sandboxes": {
			CmdArgs: []string{"--experiment=sandboxes", "--token", "xoxb-test-token"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				sandboxes := []types.Sandbox{
					{
						TeamID:       "T123",
						Name:         "my-sandbox",
						Domain:       "my-sandbox",
						Status:       "active",
						DateCreated:  1700000000,
						DateArchived: 0,
					},
				}
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return(sandboxes, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"my-sandbox", "T123", "https://my-sandbox.slack.com", "Status: ACTIVE"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "")
				assert.NotContains(t, cm.GetStdoutOutput(), "Type:")
			},
		},
		"with archived sandbox": {
			CmdArgs: []string{"--experiment=sandboxes", "--token", "xoxb-test-token"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				sandboxes := []types.Sandbox{
					{
						TeamID:       "T456",
						Name:         "old-sandbox",
						Domain:       "old-sandbox",
						Status:       "archived",
						DateCreated:  1700000000,
						DateArchived: 1710000000,
					},
				}
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return(sandboxes, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"old-sandbox", "T456", "Status: ARCHIVED"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "")
			},
		},
		"with partner sandbox shows type for all sandboxes": {
			CmdArgs: []string{"--experiment=sandboxes", "--token", "xoxb-test-token"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				sandboxes := []types.Sandbox{
					{
						TeamID:      "T123",
						Name:        "regular-sandbox",
						Domain:      "regular-sandbox",
						Status:      "active",
						DateCreated: 1700000000,
					},
					{
						TeamID:      "T789",
						Name:        "partner-sandbox",
						Domain:      "partner-sandbox",
						Status:      "active",
						DateCreated: 1700000000,
						IsPartner:   true,
					},
				}
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return(sandboxes, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"regular-sandbox", "Type: Regular", "partner-sandbox", "Type: Partner"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "")
			},
		},
		"with status": {
			CmdArgs: []string{"--experiment=sandboxes", "--token", "xoxb-test-token", "--status", "active"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("ListSandboxes", mock.Anything, testToken, "active").Return([]types.Sandbox{}, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "active")
			},
		},
		"list error": {
			CmdArgs: []string{"--experiment=sandboxes", "--token", "xoxb-test-token"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").
					Return([]types.Sandbox(nil), errors.New("api_error"))

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"api_error"},
		},
		"experiment required": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				// Do NOT enable sandboxes experiment
			},
			ExpectedErrorStrings: []string{"sandbox"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "ListSandboxes", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewListCommand(cf)
	})
}
