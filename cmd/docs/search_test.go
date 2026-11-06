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

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// clearAgentEnvVars unsets the environment variables that identify a coding
// agent, so tests behave the same when run from one
func clearAgentEnvVars(t *testing.T) {
	t.Helper()
	for _, env := range []string{"CLAUDECODE", "CLAUDE_CODE_ENTRYPOINT", "CODEX_CI", "GEMINI_CLI", "CLINE_ACTIVE", "CURSOR_AGENT", "OPENCODE", "AGENT"} {
		t.Setenv(env, "")
	}
}

func Test_Docs_makeMarkdownURL(t *testing.T) {
	tests := map[string]struct {
		url      string
		expected string
	}{
		"appends .md to a page path": {
			url:      "https://docs.slack.dev/block-kit",
			expected: "https://docs.slack.dev/block-kit.md",
		},
		"drops a trailing slash before appending .md": {
			url:      "https://docs.slack.dev/tools/bolt-js/concepts/socket-mode/",
			expected: "https://docs.slack.dev/tools/bolt-js/concepts/socket-mode.md",
		},
		"keeps an anchor": {
			url:      "https://docs.slack.dev/tools/developer-sandboxes#basic",
			expected: "https://docs.slack.dev/tools/developer-sandboxes.md#basic",
		},
		"keeps an anchor after a trailing slash": {
			url:      "https://docs.slack.dev/tools/bolt-js/concepts/socket-mode/#custom-socketmode-receiver",
			expected: "https://docs.slack.dev/tools/bolt-js/concepts/socket-mode.md#custom-socketmode-receiver",
		},
		"keeps generated .html reference pages": {
			url:      "https://docs.slack.dev/tools/python-slack-sdk/reference/web/base_client.html",
			expected: "https://docs.slack.dev/tools/python-slack-sdk/reference/web/base_client.html",
		},
		"keeps a URL that is already markdown": {
			url:      "https://docs.slack.dev/quickstart.md",
			expected: "https://docs.slack.dev/quickstart.md",
		},
		"keeps the homepage": {
			url:      "https://docs.slack.dev/",
			expected: "https://docs.slack.dev/",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, makeMarkdownURL(tc.url))
		})
	}
}

func Test_Docs_SearchCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"returns text results": {
			CmdArgs: []string{"search", "Block Kit"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				clearAgentEnvVars(t)
				cm.API.On("DocsSearch", mock.Anything, "Block Kit", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 2,
					Limit:        20,
					Results: []api.DocsSearchItem{
						{Title: "Block Kit", URL: "/block-kit"},
						{Title: "Block Kit Elements", URL: "/block-kit/elements"},
					},
				}, nil)
			},
			ExpectedStdoutOutputs: []string{
				"Block Kit",
				"https://docs.slack.dev/block-kit",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
		},
		"returns text results without markdown links for people": {
			CmdArgs: []string{"search", "Block Kit"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				clearAgentEnvVars(t)
				cm.API.On("DocsSearch", mock.Anything, "Block Kit", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 1,
					Limit:        20,
					Results: []api.DocsSearchItem{
						{Title: "Block Kit", URL: "/block-kit/"},
					},
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				assert.Contains(t, cm.GetStdoutOutput(), "https://docs.slack.dev/block-kit/")
				assert.NotContains(t, cm.GetStdoutOutput(), ".md")
			},
		},
		"returns text results with markdown links for coding agents": {
			CmdArgs: []string{"search", "Block Kit"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				clearAgentEnvVars(t)
				t.Setenv("CLAUDECODE", "1")
				cm.API.On("DocsSearch", mock.Anything, "Block Kit", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 2,
					Limit:        20,
					Results: []api.DocsSearchItem{
						{Title: "Block Kit", URL: "/block-kit/"},
						{Title: "Base client", URL: "/tools/python-slack-sdk/reference/web/base_client.html"},
					},
				}, nil)
			},
			ExpectedStdoutOutputs: []string{
				"https://docs.slack.dev/block-kit.md",
				"https://docs.slack.dev/tools/python-slack-sdk/reference/web/base_client.html",
			},
		},
		"returns JSON results": {
			CmdArgs: []string{"search", "Block Kit", "--output=json"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "Block Kit", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 2,
					Limit:        20,
					Results: []api.DocsSearchItem{
						{Title: "Block Kit", URL: "/block-kit"},
						{Title: "Block Kit Elements", URL: "/block-kit/elements"},
					},
				}, nil)
			},
			ExpectedStdoutOutputs: []string{
				`{
  "total_results": 2,
  "results": [
    {
      "url": "https://docs.slack.dev/block-kit",
      "title": "Block Kit"
    },
    {
      "url": "https://docs.slack.dev/block-kit/elements",
      "title": "Block Kit Elements"
    }
  ],
  "limit": 20
}
`,
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
		},
		"returns JSON results with absolute URLs": {
			CmdArgs: []string{"search", "test", "--output=json"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "test", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 1,
					Limit:        20,
					Results: []api.DocsSearchItem{
						{Title: "Test", URL: "https://docs.slack.dev/test"},
					},
				}, nil)
			},
			ExpectedStdoutOutputs: []string{
				`"url": "https://docs.slack.dev/test"`,
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
		},
		"returns empty results": {
			CmdArgs: []string{"search", "nonexistent"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "nonexistent", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        20,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
		},
		"returns error on API failure": {
			CmdArgs: []string{"search", "test"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "test", 20, "").Return(nil, slackerror.New(slackerror.ErrHTTPRequestFailed))
			},
			ExpectedErrorStrings: []string{slackerror.ErrHTTPRequestFailed},
		},
		"returns error on API failure for JSON output": {
			CmdArgs: []string{"search", "test", "--output=json"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "test", 20, "").Return(nil, slackerror.New(slackerror.ErrHTTPRequestFailed))
			},
			ExpectedErrorStrings: []string{slackerror.ErrHTTPRequestFailed},
		},
		"passes custom limit": {
			CmdArgs: []string{"search", "test", "--limit=5"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "test", 5, "").Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        5,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "test", 5, "")
			},
		},
		"joins multiple arguments into query": {
			CmdArgs: []string{"search", "Block", "Kit", "Element"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "Block Kit Element", 20, "").Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        20,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "Block Kit Element", 20, "")
			},
		},
		"passes category to API for text output": {
			CmdArgs: []string{"search", "chat.postMessage", "--category=reference"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "chat.postMessage", 20, "reference").Return(&api.DocsSearchResponse{
					TotalResults: 1,
					Limit:        20,
					Results: []api.DocsSearchItem{
						{Title: "chat.postMessage", URL: "/reference/methods/chat.postMessage"},
					},
				}, nil)
			},
			ExpectedStdoutOutputs: []string{
				`Displaying first 1 of 1 results for "chat.postMessage" in category "reference"`,
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "chat.postMessage", 20, "reference")
			},
		},
		"includes category in zero results message": {
			CmdArgs: []string{"search", "nonexistent", "--category=reference"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "nonexistent", 20, "reference").Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        20,
				}, nil)
			},
			ExpectedStdoutOutputs: []string{
				`Found zero results for "nonexistent" in category "reference"`,
			},
		},
		"passes category to API for json output": {
			CmdArgs: []string{"search", "events", "--category=reference", "--output=json"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "events", 20, "reference").Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        20,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "events", 20, "reference")
			},
		},
		"passes unknown category through to API": {
			CmdArgs: []string{"search", "test", "--category=bogus"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "test", 20, "bogus").Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        20,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "test", 20, "bogus")
			},
		},
		"opens browser with category filter": {
			CmdArgs: []string{"search", "webhooks", "--category=reference", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", "https://docs.slack.dev/search/?filter=reference&q=webhooks")
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"https://docs.slack.dev/search/?filter=reference&q=webhooks",
			},
		},
		"rejects invalid output format": {
			CmdArgs: []string{"search", "test", "--output=invalid"},
			ExpectedErrorStrings: []string{
				"Invalid output format",
				"Use one of: text, json, browser",
			},
		},
		"opens browser with search query": {
			CmdArgs: []string{"search", "messaging", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", "https://docs.slack.dev/search/?q=messaging")
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=messaging",
			},
		},
		"opens browser with special characters": {
			CmdArgs: []string{"search", "messages & webhooks", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", "https://docs.slack.dev/search/?q=messages+%26+webhooks")
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=messages+%26+webhooks",
			},
		},
		"opens browser with multiple arguments": {
			CmdArgs: []string{"search", "Block", "Kit", "Element", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", "https://docs.slack.dev/search/?q=Block+Kit+Element")
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=Block+Kit+Element",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
}
