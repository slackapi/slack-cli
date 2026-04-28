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

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestDeleteCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"delete success": {
			CmdArgs: []string{
				"--token", "xoxb-test-token",
				"--sandbox-id", "T123",
				"--force",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("DeleteSandbox", mock.Anything, testToken, "T123").Return(nil)
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return([]types.Sandbox{}, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
			},
			ExpectedStdoutOutputs: []string{"Sandbox Deleted", "T123", "No sandboxes found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Auth.AssertCalled(t, "AuthWithToken", mock.Anything, "xoxb-test-token")
				cm.API.AssertCalled(t, "DeleteSandbox", mock.Anything, "xoxb-test-token", "T123")
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "")
			},
		},
		"delete with remaining sandboxes": {
			CmdArgs: []string{
				"--token", "xoxb-test-token",
				"--sandbox-id", "T123",
				"--force",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("DeleteSandbox", mock.Anything, testToken, "T123").Return(nil)
				sandboxes := []types.Sandbox{
					{
						TeamID:       "T456",
						Name:         "other-sandbox",
						Domain:       "other-sandbox",
						Status:       "active",
						DateCreated:  1700000000,
						DateArchived: 0,
					},
				}
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return(sandboxes, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
			},
			ExpectedStdoutOutputs: []string{"Sandbox Deleted", "T123", "other-sandbox", "T456"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DeleteSandbox", mock.Anything, "xoxb-test-token", "T123")
				cm.API.AssertCalled(t, "ListSandboxes", mock.Anything, "xoxb-test-token", "")
			},
		},
		"deletion cancelled": {
			CmdArgs: []string{
				"--token", "xoxb-test-token",
				"--sandbox-id", "T123",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to delete the sandbox?", false).Return(false, nil)

				cm.AddDefaultMocks()
			},
			ExpectedStdoutOutputs: []string{"Deletion cancelled"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "ConfirmPrompt", mock.Anything, "Are you sure you want to delete the sandbox?", false)
				cm.API.AssertNotCalled(t, "DeleteSandbox", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"delete confirmation proceeds": {
			CmdArgs: []string{
				"--token", "xoxb-test-token",
				"--sandbox-id", "E0123456",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.IO.On("ConfirmPrompt", mock.Anything, "Are you sure you want to delete the sandbox?", false).Return(true, nil)
				cm.API.On("DeleteSandbox", mock.Anything, testToken, "E0123456").Return(nil)
				cm.API.On("ListSandboxes", mock.Anything, testToken, "").Return([]types.Sandbox{}, nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
			},
			ExpectedStdoutOutputs: []string{"Sandbox Deleted", "E0123456"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "ConfirmPrompt", mock.Anything, "Are you sure you want to delete the sandbox?", false)
				cm.API.AssertCalled(t, "DeleteSandbox", mock.Anything, "xoxb-test-token", "E0123456")
			},
		},
		"delete API error": {
			CmdArgs: []string{
				"--token", "xoxb-test-token",
				"--sandbox-id", "T123",
				"--force",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("DeleteSandbox", mock.Anything, testToken, "T123").Return(errors.New("api_error"))

				cm.AddDefaultMocks()
			},
			ExpectedErrorStrings: []string{"api_error"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DeleteSandbox", mock.Anything, "xoxb-test-token", "T123")
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewDeleteCommand(cf)
	})
}
