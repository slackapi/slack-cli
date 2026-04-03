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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func buildDocsSearchURL(query string) string {
	encodedQuery := url.QueryEscape(query)
	return fmt.Sprintf("%s/search/?q=%s", docsURL, encodedQuery)
}

type searchConfig struct {
	output string
	limit  int
}

func makeAbsoluteURL(relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http") {
		return relativeURL
	}
	return docsURL + relativeURL
}

func NewSearchCommand(clients *shared.ClientFactory) *cobra.Command {
	cfg := &searchConfig{}

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search Slack developer docs",
		Long: strings.Join([]string{
			"Search the Slack developer docs and return results in text, JSON, or browser",
			"format.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Search docs and return text results",
				Command: "docs search \"Block Kit\"",
			},
			{
				Meaning: "Search docs and open results in browser",
				Command: "docs search \"webhooks\" --output=browser",
			},
			{
				Meaning: "Search docs with limited JSON results",
				Command: "docs search \"api\" --output=json --limit=5",
			},
		}),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsSearchCommand(clients, cmd, args, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.output, "output", "text", "output format: text, json, browser")
	cmd.Flags().IntVar(&cfg.limit, "limit", 20, "maximum number of text or json search results to return")

	return cmd
}

func runDocsSearchCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string, cfg *searchConfig) error {
	ctx := cmd.Context()

	query := strings.Join(args, " ")

	switch cfg.output {
	case "json":
		searchResponse, err := clients.API().DocsSearch(ctx, query, cfg.limit)
		if err != nil {
			return err
		}

		for i := range searchResponse.Results {
			searchResponse.Results[i].URL = makeAbsoluteURL(searchResponse.Results[i].URL)
		}

		encoder := json.NewEncoder(clients.IO.WriteOut())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(searchResponse); err != nil {
			return slackerror.New(slackerror.ErrUnableToParseJSON).WithRootCause(err)
		}

		clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, query)

		return nil
	case "text":
		searchResponse, err := clients.API().DocsSearch(ctx, query, cfg.limit)
		if err != nil {
			return err
		}

		if len(searchResponse.Results) == 0 {
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "books",
				Text:  "Docs Search",
				Secondary: []string{
					fmt.Sprintf("Found zero results for \"%s\"", query),
				},
			}))
			clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, query)
			return nil
		}

		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "books",
			Text:  "Docs Search",
			Secondary: []string{
				fmt.Sprintf("Displaying first %d of %d results for \"%s\"", len(searchResponse.Results), searchResponse.TotalResults, query),
			},
		}))

		for _, result := range searchResponse.Results {
			absoluteURL := makeAbsoluteURL(result.URL)
			clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
				Emoji:     "book",
				Text:      result.Title,
				Secondary: []string{absoluteURL},
			}))
		}

		clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, query)

		return nil
	case "browser":
		docsSearchURL := buildDocsSearchURL(query)

		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "books",
			Text:  "Docs Search",
			Secondary: []string{
				docsSearchURL,
			},
		}))

		clients.Browser().OpenURL(docsSearchURL)
		clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, query)

		return nil
	default:
		return slackerror.New(slackerror.ErrInvalidFlag).WithMessage(
			"Invalid output format: %s", cfg.output,
		).WithRemediation(
			"Use one of: text, json, browser",
		)
	}
}
