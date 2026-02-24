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

package docs

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func Test_Docs_DocsCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"opens docs homepage without search": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// No special setup needed for basic functionality
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Slack developer docs",
				"https://docs.slack.dev",
			},
		},
		"opens docs with basic search query": {
			CmdArgs: []string{"--search", "Block Kit"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=Block+Kit"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Searching Slack developer docs: \"Block Kit\"",
				"https://docs.slack.dev/search/?q=Block+Kit",
			},
		},
		"handles search query with multiple words": {
			CmdArgs: []string{"--search", "socket mode authentication"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=socket+mode+authentication"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Searching Slack developer docs: \"socket mode authentication\"",
				"https://docs.slack.dev/search/?q=socket+mode+authentication",
			},
		},
		"handles special characters in search query": {
			CmdArgs: []string{"--search", "API & webhooks"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=API+%26+webhooks"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Searching Slack developer docs: \"API & webhooks\"",
				"https://docs.slack.dev/search/?q=API+%26+webhooks",
			},
		},
		"handles search query with quotes": {
			CmdArgs: []string{"--search", "function \"hello world\""},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=function+%22hello+world%22"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Searching Slack developer docs: \"function \"hello world\"\"",
				"https://docs.slack.dev/search/?q=function+%22hello+world%22",
			},
		},
		"handles empty search query as homepage": {
			CmdArgs: []string{"--search", ""},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Slack developer docs",
				"https://docs.slack.dev",
			},
		},
		"handles the exact user request example": {
			CmdArgs: []string{"--search", "something example"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=something+example"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Searching Slack developer docs: \"something example\"",
				"https://docs.slack.dev/search/?q=something+example",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
}
