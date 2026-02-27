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
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Open",
				"https://docs.slack.dev",
			},
		},
		"fails when positional argument provided without search flag": {
			CmdArgs:              []string{"Block Kit"},
			ExpectedErrorStrings: []string{"Invalid docs command. Did you mean to search?"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// No browser calls should be made when command fails
				cm.Browser.AssertNotCalled(t, "OpenURL")
			},
		},
		"fails when multiple positional arguments provided without search flag": {
			CmdArgs:              []string{"webhook", "send", "message"},
			ExpectedErrorStrings: []string{"Invalid docs command. Did you mean to search?"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// No browser calls should be made when command fails
				cm.Browser.AssertNotCalled(t, "OpenURL")
			},
		},
		"opens docs with search query using space syntax": {
			CmdArgs: []string{"--search", "messaging"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=messaging"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=messaging",
			},
		},
		"handles search with multiple arguments": {
			CmdArgs: []string{"--search", "Block", "Kit", "Element"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=Block+Kit+Element"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=Block+Kit+Element",
			},
		},
		"handles search query with multiple words": {
			CmdArgs: []string{"--search", "socket mode"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=socket+mode"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=socket+mode",
			},
		},
		"handles special characters in search query": {
			CmdArgs: []string{"--search", "messages & webhooks"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=messages+%26+webhooks"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=messages+%26+webhooks",
			},
		},
		"handles search query with quotes": {
			CmdArgs: []string{"--search", "webhook \"send message\""},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=webhook+%22send+message%22"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=webhook+%22send+message%22",
			},
		},
		"handles search flag without argument": {
			CmdArgs: []string{"--search"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DocsSearchSuccess, mock.Anything)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
}
