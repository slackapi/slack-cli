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

const docsURL = "https://docs.slack.dev"

var searchMode bool

func buildDocsSearchURL(query string) string {
	encodedQuery := url.QueryEscape(query)
	return fmt.Sprintf("%s/search/?q=%s", docsURL, encodedQuery)
}

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Open Slack developer docs",
		Long:  "Open the Slack developer docs in your browser or search them using the search subcommand",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Open Slack developer docs homepage",
				Command: "docs",
			},
			{
				Meaning: "Search Slack developer docs for Block Kit",
				Command: "docs search \"Block Kit\"",
			},
			{
				Meaning: "Search docs and open results in browser",
				Command: "docs search \"Block Kit\" --output=browser",
			},
		}),
		Args: cobra.ArbitraryArgs, // Allow any arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd, args)
		},
		// Disable automatic suggestions for unknown commands
		DisableSuggestions: true,
	}

	cmd.Flags().BoolVar(&searchMode, "search", false, "open Slack docs search page or search with query")

	// Add the search subcommand
	cmd.AddCommand(NewSearchCommand(clients))

	return cmd
}

// runDocsCommand opens Slack developer docs in the browser
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var finalURL string
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
			finalURL = buildDocsSearchURL(query)
			sectionText = "Docs Search"
		} else {
			// --search (no argument) - open search page
			finalURL = fmt.Sprintf("%s/search/", docsURL)
			sectionText = "Docs Search"
		}
	} else {
		// No search flag: default homepage
		finalURL = docsURL
		sectionText = "Docs Open"
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  sectionText,
		Secondary: []string{
			finalURL,
		},
	}))

	clients.Browser().OpenURL(finalURL)

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
