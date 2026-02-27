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
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
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
				Meaning: "Search Slack developer docs for Block Kit",
				Command: "docs --search \"Block Kit\"",
			},
			{
				Meaning: "Open Slack docs search page",
				Command: "docs --search",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd, args)
		},
	}

	cmd.Flags().BoolVar(&searchMode, "search", false, "open Slack docs search page or search with query")

	return cmd
}

// runDocsCommand opens Slack developer docs in the browser
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var docsURL string
	var sectionText string

	// Validate: if there are arguments, --search flag must be used
	if len(args) > 0 && !cmd.Flags().Changed("search") {
		query := strings.Join(args, " ")
		return slackerror.New(slackerror.ErrDocsSearchFlagRequired).WithRemediation(
			"Use --search flag: %s",
			style.Commandf(fmt.Sprintf("docs --search \"%s\"", query), false),
		)
	}

	if cmd.Flags().Changed("search") {
		if len(args) > 0 {
			// --search "query" (space-separated) - join all args as the query
			query := strings.Join(args, " ")
			encodedQuery := url.QueryEscape(query)
			docsURL = fmt.Sprintf("https://docs.slack.dev/search/?q=%s", encodedQuery)
			sectionText = "Docs Search"
		} else {
			// --search (no argument) - open search page
			docsURL = "https://docs.slack.dev/search/"
			sectionText = "Docs Search"
		}
	} else {
		// No search flag: default homepage
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

	if cmd.Flags().Changed("search") {
		traceValue := ""
		if len(args) > 0 {
			traceValue = strings.Join(args, " ")
		}
		clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, traceValue)
	} else {
		clients.IO.PrintTrace(ctx, slacktrace.DocsSuccess)
	}

	return nil
}
