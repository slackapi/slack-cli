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
	"github.com/stretchr/testify/mock"
)

func TestCreateCommand(t *testing.T) {
	archiveDate := time.Now().UTC().AddDate(0, 6, 0).Truncate(24 * time.Hour)
	archiveDateStr := archiveDate.Format("2006-01-02")
	archiveEpoch := archiveDate.Unix()

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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "err-box", "err-box", "pass", "", "", 0, "", int64(0), false).
					Return("", "", errors.New("api_error"))

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"api_error"},
		},
		"create with 'default' template": {
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"Invalid template"},
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
				"--archive-date", archiveDateStr,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "date-box", "date-box", "pass", "", "", 0, "", archiveEpoch, false).
					Return("T222", "https://date-box.slack.com", nil)

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "date-box", "date-box", "pass", "", "", 0, "", archiveEpoch, false)
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")

				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"Cannot use both --archive-ttl and --archive-date"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"create with invalid archive-ttl value fails": {
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")

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

func setupCreateMocks(t *testing.T, ctx context.Context, cm *shared.ClientsMock, name, domain, password string, archiveEpoch interface{}, partner bool) {
	t.Helper()
	testToken := "xoxb-test-token"
	cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
	cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
	cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
	cm.API.On("CreateSandbox", mock.Anything, testToken, name, domain, password, "", "", 0, "", archiveEpoch, partner).
		Return("T222", "https://"+domain+".slack.com", nil)
	cm.AddDefaultMocks()
	cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
	cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
}

func setupCreateAuthOnly(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
	t.Helper()
	testToken := "xoxb-test-token"
	cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
	cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
	cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
	cm.AddDefaultMocks()
	cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
	cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
}

func Test_getEpochFromTTL(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"1d": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "ttl-box",
				"--domain", "ttl-box",
				"--password", "pass",
				"--archive-ttl", "1d",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateMocks(t, ctx, cm, "ttl-box", "ttl-box", "pass", mock.MatchedBy(func(x int64) bool { return x > 0 }), false)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "ttl-box", "ttl-box", "pass", "", "", 0, "", mock.MatchedBy(func(x int64) bool { return x > 0 }), false)
			},
		},
		"1w": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "ttl-box",
				"--domain", "ttl-box",
				"--password", "pass",
				"--archive-ttl", "1w",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateMocks(t, ctx, cm, "ttl-box", "ttl-box", "pass", mock.MatchedBy(func(x int64) bool { return x > 0 }), false)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "ttl-box", "ttl-box", "pass", "", "", 0, "", mock.MatchedBy(func(x int64) bool { return x > 0 }), false)
			},
		},
		"6mo": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "ttl-box",
				"--domain", "ttl-box",
				"--password", "pass",
				"--archive-ttl", "6mo",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateMocks(t, ctx, cm, "ttl-box", "ttl-box", "pass", mock.MatchedBy(func(x int64) bool { return x > 0 }), false)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "ttl-box", "ttl-box", "pass", "", "", 0, "", mock.MatchedBy(func(x int64) bool { return x > 0 }), false)
			},
		},
		"invalid": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "ttl-box",
				"--domain", "ttl-box",
				"--password", "pass",
				"--archive-ttl", "invalid",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateAuthOnly(t, ctx, cm)
			},
			ExpectedErrorStrings: []string{"Invalid TTL"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func Test_getEpochFromDate(t *testing.T) {
	validDate := time.Now().UTC().Add(7 * 24 * time.Hour).Truncate(24 * time.Hour)
	validDateStr := validDate.Format("2006-01-02")
	validEpoch := validDate.Unix()

	testutil.TableTestCommand(t, testutil.CommandTests{
		"valid": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "date-box",
				"--domain", "date-box",
				"--password", "pass",
				"--archive-date", validDateStr,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateMocks(t, ctx, cm, "date-box", "date-box", "pass", validEpoch, false)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "date-box", "date-box", "pass", "", "", 0, "", validEpoch, false)
			},
		},
		"invalid format": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "date-box",
				"--domain", "date-box",
				"--password", "pass",
				"--archive-date", "12-31-2025",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateAuthOnly(t, ctx, cm)
			},
			ExpectedErrorStrings: []string{"Invalid archive date"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"invalid date": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "date-box",
				"--domain", "date-box",
				"--password", "pass",
				"--archive-date", "not-a-date",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateAuthOnly(t, ctx, cm)
			},
			ExpectedErrorStrings: []string{"Invalid archive date"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"date in past": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "date-box",
				"--domain", "date-box",
				"--password", "pass",
				"--archive-date", "2020-01-01",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateAuthOnly(t, ctx, cm)
			},
			ExpectedErrorStrings: []string{"Archive date must be in the future"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func Test_getTemplateID(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"valid template name": {
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
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "tpl-box", "tpl-box", "pass", "", "", 1, "", int64(0), false).
					Return("T333", "https://tpl-box.slack.com", nil)
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "tpl-box", "tpl-box", "pass", "", "", 1, "", int64(0), false)
			},
		},
		"integer value also accepted": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tpl-box",
				"--domain", "tpl-box",
				"--password", "pass",
				"--template", "1",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				testToken := "xoxb-test-token"
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
				cm.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
				cm.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
				cm.API.On("CreateSandbox", mock.Anything, testToken, "tpl-box", "tpl-box", "pass", "", "", 1, "", int64(0), false).
					Return("T333", "https://tpl-box.slack.com", nil)
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "tpl-box", "tpl-box", "pass", "", "", 1, "", int64(0), false)
			},
		},
		"invalid template name fails": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "tpl-box",
				"--domain", "tpl-box",
				"--password", "pass",
				"--template", "invalid",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateAuthOnly(t, ctx, cm)
			},
			ExpectedErrorStrings: []string{"Invalid template"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func Test_domainFromName(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"handles invalid URL characters": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "-Hello_World 123-",
				"--password", "pass",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateMocks(t, ctx, cm, "-Hello_World 123-", "hello-world-123", "pass", int64(0), false)
			},
			ExpectedStdoutOutputs: []string{"Sandbox Created"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "CreateSandbox", mock.Anything, "xoxb-test-token", "-Hello_World 123-", "hello-world-123", "pass", "", "", 0, "", int64(0), false)
			},
		},
		"empty": {
			CmdArgs: []string{
				"--experiment=sandboxes",
				"--token", "xoxb-test-token",
				"--name", "",
				"--password", "pass",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				setupCreateAuthOnly(t, ctx, cm)
				cm.IO.On("InputPrompt", mock.Anything, "Enter a name for the sandbox", mock.Anything).Return("", nil)
			},
			ExpectedErrorStrings: []string{"Provide a valid domain name"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "CreateSandbox", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}
