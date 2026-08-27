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

package app

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_NewStatusCommand(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	cmd := NewStatusCommand(clients)
	assert.Equal(t, "status", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}

func Test_runStatusCommand(t *testing.T) {
	tests := map[string]struct {
		appID          string
		teamID         string
		teamDomain     string
		installed      bool
		expectedStatus AppInstallRequestStatus
	}{
		"installed app shows installed status": {
			appID:          "A0123456789",
			teamID:         "T0001",
			teamDomain:     "test-team",
			installed:      true,
			expectedStatus: InstallRequestStatusInstalled,
		},
		"uninstalled app shows unknown status": {
			appID:          "A0123456789",
			teamID:         "T0001",
			teamDomain:     "test-team",
			installed:      false,
			expectedStatus: InstallRequestStatusUnknown,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
			})

			cmd := NewStatusCommand(clients)
			testutil.MockCmdIO(clients.IO, cmd)

			statusAppSelectPromptFunc = func(ctx context.Context, clients *shared.ClientFactory, environment prompts.AppEnvironmentType, status prompts.AppInstallStatus, opts ...prompts.AppSelectOption) (prompts.SelectedApp, error) {
				return prompts.SelectedApp{
					Auth: types.SlackAuth{
						Token:      "xoxp-test-token",
						TeamID:     tc.teamID,
						TeamDomain: tc.teamDomain,
					},
					App: types.App{
						AppID:  tc.appID,
						TeamID: tc.teamID,
					},
				}, nil
			}

			clientsMock.API.On("GetAppStatus", ctx, "xoxp-test-token", []string{tc.appID}, tc.teamID).Return(api.GetAppStatusResult{
				Apps: []api.AppStatusResultAppInfo{
					{
						AppID:     tc.appID,
						Installed: tc.installed,
					},
				},
			}, nil)

			err := cmd.ExecuteContext(ctx)
			assert.NoError(t, err)
		})
	}
}

func Test_formatStatusLabel(t *testing.T) {
	tests := map[string]struct {
		status   AppInstallRequestStatus
		contains string
	}{
		"installed contains status text": {
			status:   InstallRequestStatusInstalled,
			contains: "Installed",
		},
		"approved contains status text": {
			status:   InstallRequestStatusApproved,
			contains: "Approved",
		},
		"pending contains status text": {
			status:   InstallRequestStatusPending,
			contains: "Pending",
		},
		"denied contains status text": {
			status:   InstallRequestStatusDenied,
			contains: "Denied",
		},
		"unknown contains status text": {
			status:   InstallRequestStatusUnknown,
			contains: "Unknown",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := formatStatusLabel(tc.status)
			assert.Contains(t, result, tc.contains)
		})
	}
}
