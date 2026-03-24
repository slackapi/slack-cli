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
	"net/http"
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var searchOutputFlag string
var searchLimitFlag int

// response from the Slack docs search API
type DocsSearchResponse struct {
	TotalResults int                `json:"total_results"`
	Results      []DocsSearchResult `json:"results"`
	Limit        int                `json:"limit"`
}

// single search result
type DocsSearchResult struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func NewSearchCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Slack developer docs (experimental)",
		Long:  "Search the Slack developer docs and return results in browser or JSON format",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Search docs and open results in browser",
				Command: "docs search \"Block Kit\"",
			},
			{
				Meaning: "Search docs and return JSON results",
				Command: "docs search \"webhooks\" --output=json",
			},
			{
				Meaning: "Search docs with limited JSON results",
				Command: "docs search \"api\" --output=json --limit=5",
			},
		}),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsSearchCommand(clients, cmd, args, http.DefaultClient)
		},
	}

	cmd.Flags().StringVar(&searchOutputFlag, "output", "json", "output format: browser, json")
	cmd.Flags().IntVar(&searchLimitFlag, "limit", 20, "maximum number of search results to return (only applies with --output=json)")

	return cmd
}

// handles the docs search subcommand
func runDocsSearchCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string, httpClient *http.Client) error {
	ctx := cmd.Context()

	query := strings.Join(args, " ")

	if searchOutputFlag == "json" {
		return fetchAndOutputSearchResults(ctx, clients, query, searchLimitFlag, httpClient)
	}

	// Browser output - open search results in browser
	encodedQuery := url.QueryEscape(query)
	docsURL := fmt.Sprintf("https://docs.slack.dev/search/?q=%s", encodedQuery)

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "Docs Search",
		Secondary: []string{
			docsURL,
		},
	}))

	clients.Browser().OpenURL(docsURL)
	clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, query)

	return nil
}

// fetches search results from the docs API and outputs as JSON
func fetchAndOutputSearchResults(ctx context.Context, clients *shared.ClientFactory, query string, limit int, httpClient *http.Client) error {
	// Build API URL with limit parameter
	apiURL := fmt.Sprintf("https://docs-slack-d-search-api-duu9zr.herokuapp.com/api/search?q=%s&limit=%d", url.QueryEscape(query), limit)

	// Make HTTP request
	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch search results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Parse JSON response
	var searchResponse DocsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return fmt.Errorf("failed to parse search results: %w", err)
	}

	// Output as JSON
	output, err := json.MarshalIndent(searchResponse, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON output: %w", err)
	}

	fmt.Println(string(output))

	// Trace the successful API call
	clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, query)

	return nil
}
