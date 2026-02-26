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
	"fmt"
	"net/url"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var searchMode bool

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Open Slack developer docs",
		Long:  "Open the Slack developer docs in your browser, with optional search functionality",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Open Slack developer docs homepage",
				Command: "docs",
			},
			{
				Meaning: "Open Slack docs search page",
				Command: "docs --search",
			},
			{
				Meaning: "Search Slack docs",
				Command: "docs --search \"Block Kit\"",
			},
			{
				Meaning: "Search Slack docs without search flag",
				Command: "docs \"Block Kit\"",
			},
		}),
		Args: cobra.MaximumNArgs(1), // Allow 0-1 arguments for search query
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd, args)
		},
	}

	cmd.Flags().BoolVar(&searchMode, "search", false, "open Slack docs search page")

	return cmd
}

// runDocsCommand opens Slack developer docs in the browser
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var docsURL string
	var sectionText string

	if len(args) > 0 {
		// Search query provided as positional argument: slack docs "query"
		searchQuery := url.QueryEscape(args[0])
		docsURL = fmt.Sprintf("https://docs.slack.dev/search/?q=%s", searchQuery)
		sectionText = "Docs Search"
	} else if searchMode {
		// Search flag provided without query: slack docs --search
		docsURL = "https://docs.slack.dev/search/"
		sectionText = "Docs Search"
	} else {
		// Default homepage: slack docs
		docsURL = "https://docs.slack.dev"
		sectionText = "Docs Open"
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  sectionText,
		Secondary: []string{
			docsURL,
		},
	}))

	clients.Browser().OpenURL(docsURL)

	if len(args) > 0 || searchMode {
		traceValue := ""
		if len(args) > 0 {
			traceValue = args[0]
		}
		clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, traceValue)
	} else {
		clients.IO.PrintTrace(ctx, slacktrace.DocsSuccess)
	}

	return nil
}
