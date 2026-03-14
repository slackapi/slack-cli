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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/search"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var outputFormat string
var searchLimit int
var searchFilter string

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Open Slack developer docs",
		Long:  "Open the Slack developer docs in your browser, or search docs with subcommands",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Open Slack developer docs homepage",
				Command: "docs",
			},
			{
				Meaning: "Search and return JSON (default)",
				Command: "docs search \"Block Kit\"",
			},
			{
				Meaning: "Search and open in browser",
				Command: "docs search \"webhooks\" --output=browser",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd, args)
		},
	}

	// Add search subcommand
	cmd.AddCommand(newSearchCommand(clients))

	return cmd
}

// newSearchCommand creates the search subcommand
func newSearchCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search Slack developer documentation",
		Long:  "Search the Slack developer documentation and return results in JSON format (default) or open in browser. If no query provided, opens search page in browser.",
		Args:  cobra.MaximumNArgs(1),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Open docs search page in browser",
				Command: "docs search",
			},
			{
				Meaning: "Search for Block Kit (returns JSON by default)",
				Command: "docs search \"Block Kit\"",
			},
			{
				Meaning: "Search and open in browser",
				Command: "docs search \"Block Kit\" --output=browser",
			},
			{
				Meaning: "Search with custom limit",
				Command: "docs search \"webhooks\" --limit=50",
			},
			{
				Meaning: "Search with filter",
				Command: "docs search \"webhooks\" --filter=guides",
			},
			{
				Meaning: "Search Python documentation and open in browser",
				Command: "docs search \"bolt\" --filter=python --output=browser",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runSearchBrowserCommand(clients, cmd)
			}
			return runSearchCommand(clients, cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&outputFormat, "output", "json", "output format: json, browser")
	cmd.Flags().IntVar(&searchLimit, "limit", 20, "maximum number of results to return")
	cmd.Flags().StringVar(&searchFilter, "filter", "", "filter results by content type: guides, reference, changelog, python, javascript, java, slack_cli, slack_github_action, deno_slack_sdk")

	return cmd
}

// openSearchInBrowser opens the docs search page in browser
func openSearchInBrowser(clients *shared.ClientFactory, ctx context.Context, searchURL string) error {
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "Docs Search",
		Secondary: []string{
			searchURL,
		},
	}))

	clients.Browser().OpenURL(searchURL)
	clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, "")
	return nil
}

// runSearchBrowserCommand opens the docs search page in browser
func runSearchBrowserCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()
	searchURL := "https://docs.slack.dev/search"
	return openSearchInBrowser(clients, ctx, searchURL)
}

// runSearchCommand handles the search subcommand
func runSearchCommand(clients *shared.ClientFactory, cmd *cobra.Command, query string) error {
	ctx := cmd.Context()

	results, err := search.SearchDocs(query, searchFilter, searchLimit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Output results
	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Browser output - open search page with query
		searchURL := fmt.Sprintf("https://docs.slack.dev/search/?q=%s", url.QueryEscape(query))
		if searchFilter != "" {
			searchURL += fmt.Sprintf("&filter=%s", url.QueryEscape(searchFilter))
		}
		return openSearchInBrowser(clients, ctx, searchURL)
	}

	return nil
}

// runDocsCommand opens Slack developer docs in the browser
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// If any arguments provided, suggest using search subcommand
	if len(args) > 0 {
		query := strings.Join(args, " ")
		return slackerror.New(slackerror.ErrDocsSearchFlagRequired).WithRemediation(
			"Use search subcommand: %s",
			style.Commandf(fmt.Sprintf("docs search \"%s\"", query), false),
		)
	}

	// Open docs homepage
	docsURL := "https://docs.slack.dev"

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "Docs Open",
		Secondary: []string{
			docsURL,
		},
	}))

	clients.Browser().OpenURL(docsURL)
	clients.IO.PrintTrace(ctx, slacktrace.DocsSuccess)

	return nil
}
