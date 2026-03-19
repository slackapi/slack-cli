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
	"time"

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
				cm.API.On("CreateSandbox", mock.Anything, testToken, "test-box", "test-box", "mypass", "", "", 0, "", int64(0), false).
					Return("T123", "https://test-box.slack.com", nil)
				cm.API.On("UsersInfo", mock.Anything, mock.Anything, mock.Anything).Return(&types.UserInfo{Profile: types.UserProfile{}}, nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T123", "https://test-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Auth.AssertCalled(t, "AuthWithToken", mock.Anything, "xoxb-test-token")
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "test-box", "test-box", "mypass", "", "", 0, "", int64(0), false)
			},
		},
		"create with json-box": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "json-box",
				"--domain", "json-box",
				"--password", "secret",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "json-box", "json-box", "secret", "", "", 0, "", int64(0), false).
					Return("T456", "https://json-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T456", "https://json-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "json-box", "json-box", "secret", "", "", 0, "", int64(0), false)
			},
		},
		"create with partner": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "partner-box",
				"--domain", "partner-box",
				"--password", "pass",
				"--partner",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "partner-box", "partner-box", "pass", "", "", 0, "", int64(0), true).
					Return("T999", "https://partner-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T999", "https://partner-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "partner-box", "partner-box", "pass", "", "", 0, "", int64(0), true)
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
				cm.API.On("CreateSandbox", mock.Anything, testToken, "My Test Box", "my-test-box", "pass", "", "", 0, "", int64(0), false).
					Return("T789", "https://my-test-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "My Test Box", "my-test-box", "pass", "", "", 0, "", int64(0), false)
			},
		},
		"create with partner": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "partner-box",
				"--domain", "partner-box",
				"--password", "pass",
				"--partner",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "partner-box", "partner-box", "pass", "", "", 0, "", int64(0), true).
					Return("T999", "https://partner-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T999", "https://partner-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "partner-box", "partner-box", "pass", "", "", 0, "", int64(0), true)
			},
		},
		"create with a relative time-to-live value": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tmp-box",
				"--domain", "tmp-box",
				"--password", "pass",
				"--archive-ttl", "1d",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "tmp-box", "tmp-box", "pass", "", "", 0, "", mock.MatchedBy(func(v int64) bool { return v > 0 }), false).
					Return("T111", "https://tmp-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "tmp-box", "tmp-box", "pass", "", "", 0, "", mock.MatchedBy(func(v int64) bool { return v > 0 }), false)
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
				cm.API.On("CreateSandbox", mock.Anything, testToken, "err-box", "err-box", "pass", "", "", 0, "", int64(0), false).
					Return("", "", errors.New("api_error"))

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"api_error"},
		},
		"create with template default": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tpl-box",
				"--domain", "tpl-box",
				"--password", "pass",
				"--template", "default",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "tpl-box", "tpl-box", "pass", "", "", 1, "", int64(0), false).
					Return("T333", "https://tpl-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T333", "https://tpl-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "tpl-box", "tpl-box", "pass", "", "", 1, "", int64(0), false)
			},
		},
		"create with partner flag": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "partner-box",
				"--domain", "partner-box",
				"--password", "pass",
				"--partner",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "partner-box", "partner-box", "pass", "", "", 0, "", int64(0), true).
					Return("T555", "https://partner-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"T555", "https://partner-box.slack.com", "Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "partner-box", "partner-box", "pass", "", "", 0, "", int64(0), true)
			},
		},
		"create with template empty": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tmpl-box",
				"--domain", "tmpl-box",
				"--password", "pass",
				"--template", "empty",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "tmpl-box", "tmpl-box", "pass", "", "", 0, "", int64(0), false).
					Return("T444", "https://tmpl-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "tmpl-box", "tmpl-box", "pass", "", "", 0, "", int64(0), false)
			},
		},
		"create with invalid template fails": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tmpl-box",
				"--domain", "tmpl-box",
				"--password", "pass",
				"--template", "invalid",
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
			ExpectedErrorStrings: []string{"Invalid template", "default, empty"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"create with archive-date": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "date-box",
				"--domain", "date-box",
				"--password", "pass",
				"--archive-date", "2025-12-31",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "date-box", "date-box", "pass", "", "", 0, "", int64(1767139200), false).
					Return("T222", "https://date-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "date-box", "date-box", "pass", "", "", 0, "", int64(1767139200), false)
			},
		},
		"create with both archive and archive-date fails": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tmp-box",
				"--domain", "tmp-box",
				"--password", "pass",
				"--archive-ttl", "1d",
				"--archive-date", "2025-12-31",
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
			ExpectedErrorStrings: []string{"Cannot use both --archive-ttl and --archive-date"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"invalid archive value": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tmp-box",
				"--domain", "tmp-box",
				"--password", "pass",
				"--archive-ttl", "invalid",
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
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
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
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func Test_getEpochFromTTL(t *testing.T) {
	tests := []struct {
		name    string
		ttl     string
		wantErr bool
	}{
		{"1d", "1d", false},
		{"7d", "7d", false},
		{"1w", "1w", false},
		{"2w", "2w", false},
		{"1mo", "1mo", false},
		{"6mo", "6mo", false},
		{"hours rejected", "12h", true},
		{"7mo exceeds max", "7mo", true},
		{"invalid", "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEpochFromTTL(tt.ttl)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Greater(t, got, int64(0), "archive date should be in the future")
		})
	}
}

func Test_getEpochFromDate(t *testing.T) {
	tests := []struct {
		name       string
		dateStr    string
		want       int64
		wantErr    bool
		errContain string
	}{
		{"valid", "2025-12-31", 1767139200, false, ""}, // 2025-12-31 00:00:00 UTC
		{"invalid format", "12-31-2025", 0, true, ""},
		{"invalid date", "not-a-date", 0, true, ""},
		{"date in past", "2020-01-01", 0, true, "at least 1 day"},
		{"date is today", "", 0, true, "at least 1 day"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dateStr := tt.dateStr
			if tt.name == "date is today" {
				dateStr = time.Now().UTC().Format("2006-01-02")
			}
			got, err := getEpochFromDate(dateStr)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContain != "" {
					assert.ErrorContains(t, err, tt.errContain)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getTemplateID(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"empty string", "", 0, false},
		{"default", "default", 1, false},
		{"empty", "empty", 0, false},
		{"default case insensitive", "Default", 1, false},
		{"default case insensitive uppercase", "DEFAULT", 1, false},
		{"empty case insensitive", "Empty", 0, false},
		{"empty case insensitive uppercase", "EMPTY", 0, false},
		{"invalid", "invalid", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTemplateID(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_domainFromName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"simple", "test-box", "test-box", false},
		{"spaces", "My Test Box", "my-test-box", false},
		{"uppercase", "MyBox", "mybox", false},
		{"mixed", "Hello_World 123", "hello-world-123", false},
		{"hyphens", "a--b", "a-b", false},
		{"leading trailing", "-test-", "test", false},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domainFromName(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
