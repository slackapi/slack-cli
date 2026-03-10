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
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/search"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var searchMode bool
var outputFormat string
var searchLimit int
var searchOffset int

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
				Meaning: "Search and get JSON results",
				Command: "docs --search \"Block Kit\" --output=json",
			},
			{
				Meaning: "Search with custom limit",
				Command: "docs --search \"Block Kit\" --output=json --limit=50",
			},
			{
				Meaning: "Search with pagination",
				Command: "docs --search \"Block Kit\" --output=json --limit=20 --offset=20",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd, args)
		},
	}

	cmd.Flags().BoolVar(&searchMode, "search", false, "search Slack docs with optional query")
	cmd.Flags().StringVar(&outputFormat, "output", "browser", "output format: browser, json")
	cmd.Flags().IntVar(&searchLimit, "limit", 20, "maximum number of results to return")
	cmd.Flags().IntVar(&searchOffset, "offset", 0, "number of results to skip (for pagination)")

	return cmd
}

// DocsOutput represents the structured output for --json mode
type DocsOutput struct {
	URL   string `json:"url"`
	Query string `json:"query,omitempty"`
	Type  string `json:"type"` // "homepage", "search", or "search_with_query"
}

// ProgrammaticSearchOutput represents the output from local docs search
type ProgrammaticSearchOutput = search.SearchResponse

// findDocsRepo tries to locate the docs repository
func findDocsRepo() string {
	return search.FindDocsRepo()
}

// runProgrammaticSearch executes the local search
func runProgrammaticSearch(query string, docsPath string) (*ProgrammaticSearchOutput, error) {
	contentDir := filepath.Join(docsPath, "content")
	return search.SearchDocs(query, "", searchLimit, searchOffset, contentDir)
}

// runDocsCommand opens Slack developer docs in the browser or performs programmatic search
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var docsURL string
	var sectionText string
	var query string
	var docType string

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
			query = strings.Join(args, " ")

			// Check output format
			if outputFormat == "json" {
				return runProgrammaticSearchCommand(clients, ctx, query)
			}

			// Default browser search
			encodedQuery := url.QueryEscape(query)
			docsURL = fmt.Sprintf("https://docs.slack.dev/search/?q=%s", encodedQuery)
			sectionText = "Docs Search"
			docType = "search_with_query"
		} else {
			// --search (no argument) - open search page
			docsURL = "https://docs.slack.dev/search/"
			sectionText = "Docs Search"
			docType = "search"
		}
	} else {
		// No search flag: default homepage
		docsURL = "https://docs.slack.dev"
		sectionText = "Docs Open"
		docType = "homepage"
	}

	// Handle JSON output mode (for browser-based results only)
	if outputFormat == "json" && !cmd.Flags().Changed("search") {
		output := DocsOutput{
			URL:   docsURL,
			Query: query,
			Type:  docType,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return slackerror.New(slackerror.ErrDocsJSONEncodeFailed)
		}

		fmt.Println(string(jsonBytes))

		// Still print trace for analytics
		if cmd.Flags().Changed("search") {
			traceValue := query
			clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, traceValue)
		} else {
			clients.IO.PrintTrace(ctx, slacktrace.DocsSuccess)
		}

		return nil
	}

	// Standard browser-opening mode
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

// runProgrammaticSearchCommand handles local documentation search
func runProgrammaticSearchCommand(clients *shared.ClientFactory, ctx context.Context, query string) error {
	// Find the docs repository
	docsPath := findDocsRepo()
	if docsPath == "" {
		clients.IO.PrintError(ctx, "❌ Docs repository not found")
		clients.IO.PrintInfo(ctx, false, "💡 Make sure the docs repository is cloned alongside slack-cli")
		clients.IO.PrintInfo(ctx, false, "   Expected structure:")
		clients.IO.PrintInfo(ctx, false, "   ├── slack-cli/")
		clients.IO.PrintInfo(ctx, false, "   └── docs/")
		return fmt.Errorf("docs repository not found")
	}

	// Run the search
	results, err := runProgrammaticSearch(query, docsPath)
	if err != nil {
		clients.IO.PrintError(ctx, "❌ Search failed: %v", err)
		return err
	}

	// Always output JSON for programmatic search
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}
