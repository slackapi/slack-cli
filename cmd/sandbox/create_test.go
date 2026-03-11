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

func TestCreateCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"create success": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "test-box",
				"--domain", "test-box",
				"--password", "mypass",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "test-box", "test-box", "mypass", "", "", "", "", int64(0)).
					Return("T123", "https://test-box.slack.com", nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T123", "https://test-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Auth.AssertCalled(t, "AuthWithToken", mock.Anything, "xoxb-test-token")
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "test-box", "test-box", "mypass", "", "", "", "", int64(0))
			},
		},
		"create with JSON output": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "json-box",
				"--domain", "json-box",
				"--password", "secret",
				"--output", "json",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "json-box", "json-box", "secret", "", "", "", "", int64(0)).
					Return("T456", "https://json-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{`"team_id": "T456"`, `"url": "https://json-box.slack.com"`},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "json-box", "json-box", "secret", "", "", "", "", int64(0))
			},
		},
		"create with derived domain": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "My Test Box",
				"--domain", "my-test-box",
				"--password", "pass",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "My Test Box", "my-test-box", "pass", "", "", "", "", int64(0)).
					Return("T789", "https://my-test-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "My Test Box", "my-test-box", "pass", "", "", "", "", int64(0))
			},
		},
		"create with TTL": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "ttl-box",
				"--domain", "ttl-box",
				"--password", "pass",
				"--ttl", "24h",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "ttl-box", "ttl-box", "pass", "", "", "", "", mock.MatchedBy(func(v int64) bool { return v > 0 })).
					Return("T111", "https://ttl-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "ttl-box", "ttl-box", "pass", "", "", "", "", mock.MatchedBy(func(v int64) bool { return v > 0 }))
			},
		},
		"create API error": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "err-box",
				"--domain", "err-box",
				"--password", "pass",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "err-box", "err-box", "pass", "", "", "", "", int64(0)).
					Return("", "", errors.New("api_error"))

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"api_error"},
		},
		"invalid TTL": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "ttl-box",
				"--domain", "ttl-box",
				"--password", "pass",
				"--ttl", "invalid",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"Invalid TTL"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"experiment required": {
			CmdArgs: []string{
				"--name", "test-box",
				"--domain", "test-box",
				"--password", "pass",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
			},
			ExpectedErrorStrings: []string{"sandbox"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func Test_ttlToArchiveDate(t *testing.T) {
	tests := []struct {
		name    string
		ttl     string
		wantErr bool
	}{
		{"empty", "", false},
		{"24h", "24h", false},
		{"1d", "1d", false},
		{"7d", "7d", false},
		{"invalid", "invalid", true},
		{"exceeds max", "200d", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ttlToArchiveDate(tt.ttl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.ttl == "" {
				assert.Equal(t, int64(0), got)
			} else {
				assert.Greater(t, got, int64(0), "archive date should be in the future")
			}
		})
	}
}

func Test_slugFromsandboxName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "test-box", "test-box"},
		{"spaces", "My Test Box", "my-test-box"},
		{"uppercase", "MyBox", "mybox"},
		{"mixed", "Hello_World 123", "hello-world-123"},
		{"hyphens", "a--b", "a-b"},
		{"leading trailing", "-test-", "test"},
		{"empty", "", "sandbox"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugFromsandboxName(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
