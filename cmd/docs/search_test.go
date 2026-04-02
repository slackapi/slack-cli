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
	"github.com/stretchr/testify/mock"
)

func Test_Docs_SearchCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"returns text results": {
			CmdArgs: []string{"search", "Block Kit"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "Block Kit", 20).Return(&api.DocsSearchResponse{
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
		"returns JSON results": {
			CmdArgs: []string{"search", "Block Kit", "--output=json"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "Block Kit", 20).Return(&api.DocsSearchResponse{
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
		"returns empty results": {
			CmdArgs: []string{"search", "nonexistent"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "nonexistent", 20).Return(&api.DocsSearchResponse{
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
				cm.API.On("DocsSearch", mock.Anything, "test", 20).Return(nil, slackerror.New(slackerror.ErrHTTPRequestFailed))
			},
			ExpectedErrorStrings: []string{slackerror.ErrHTTPRequestFailed},
		},
		"passes custom limit": {
			CmdArgs: []string{"search", "test", "--limit=5"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "test", 5).Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        5,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "test", 5)
			},
		},
		"joins multiple arguments into query": {
			CmdArgs: []string{"search", "Block", "Kit", "Element"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsSearch", mock.Anything, "Block Kit Element", 20).Return(&api.DocsSearchResponse{
					TotalResults: 0,
					Results:      []api.DocsSearchItem{},
					Limit:        20,
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "DocsSearch", mock.Anything, "Block Kit Element", 20)
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
