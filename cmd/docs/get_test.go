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
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Docs_GetCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"prints the markdown of a page path": {
			CmdArgs: []string{"get", "/tools/bolt-js/concepts/socket-mode"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsGetPage", mock.Anything, "/tools/bolt-js/concepts/socket-mode.md").Return("# Using Socket Mode\n", nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				assert.Equal(t, "# Using Socket Mode\n", cm.GetStdoutOutput())
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsGetSuccess, mock.Anything)
			},
		},
		"prints the markdown of a search result link": {
			CmdArgs: []string{"get", "https://docs.slack.dev/tools/bolt-js/concepts/socket-mode.md#custom-socketmode-receiver"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsGetPage", mock.Anything, "/tools/bolt-js/concepts/socket-mode.md").Return("# Using Socket Mode\n", nil)
			},
			ExpectedStdoutOutputs: []string{"# Using Socket Mode"},
		},
		"accepts a docs URL with an uppercase host": {
			CmdArgs: []string{"get", "https://DOCS.SLACK.DEV/block-kit"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsGetPage", mock.Anything, "/block-kit.md").Return("# Block Kit\n", nil)
			},
			ExpectedStdoutOutputs: []string{"# Block Kit"},
		},
		"returns the page as JSON": {
			CmdArgs: []string{"get", "docs.slack.dev/block-kit/", "--output=json"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsGetPage", mock.Anything, "/block-kit.md").Return("# Block Kit\n", nil)
			},
			ExpectedStdoutOutputs: []string{
				`"url": "https://docs.slack.dev/block-kit/"`,
				`"markdown_url": "https://docs.slack.dev/block-kit.md"`,
				`"content": "# Block Kit\n"`,
			},
		},
		"opens the page in the browser": {
			CmdArgs: []string{"get", "/block-kit", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", "https://docs.slack.dev/block-kit")
				cm.API.AssertNotCalled(t, "DocsGetPage", mock.Anything, mock.Anything)
			},
		},
		"prints the homepage": {
			CmdArgs: []string{"get", "https://docs.slack.dev"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsGetPage", mock.Anything, "/index.md").Return("# Build an app with a coding agent\n", nil)
			},
			ExpectedStdoutOutputs: []string{"# Build an app with a coding agent"},
		},
		"errors when no page is given": {
			CmdArgs:              []string{"get"},
			ExpectedErrorStrings: []string{"A docs page path or URL is required", "docs search"},
		},
		"errors for a page with no markdown version": {
			CmdArgs:              []string{"get", "/tools/python-slack-sdk/reference/web/base_client.html"},
			ExpectedErrorStrings: []string{"has no markdown version", "--output=browser"},
		},
		"errors for a page on another site": {
			CmdArgs:              []string{"get", "https://example.com/block-kit"},
			ExpectedErrorStrings: []string{"is not on docs.slack.dev"},
		},
		"errors when the page cannot be fetched": {
			CmdArgs: []string{"get", "/missing"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("DocsGetPage", mock.Anything, "/missing.md").Return("", errors.New("unexpected status code 404"))
			},
			ExpectedErrorStrings: []string{"unexpected status code 404", "docs search"},
		},
		"errors for an invalid output format": {
			CmdArgs: []string{"get", "/block-kit", "--output=yaml"},
			ExpectedError: slackerror.New(slackerror.ErrInvalidFlag).
				WithMessage("Invalid output format: %s", "yaml").
				WithRemediation("Use one of: text, json, browser"),
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
}
