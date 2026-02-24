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

var searchFlag string

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
				Meaning: "Search Slack developer docs",
				Command: "docs --search 'Block Kit'",
			},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd, args)
		},
	}

	cmd.Flags().StringVar(&searchFlag, "search", "", "search query for Slack docs")

	return cmd
}

// runDocsCommand opens Slack developer docs in the browser
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var docsURL string
	var sectionText string

	if searchFlag != "" {
		// Build search URL
		searchQuery := url.QueryEscape(searchFlag)
		docsURL = fmt.Sprintf("https://docs.slack.dev/search/?q=%s", searchQuery)
		sectionText = fmt.Sprintf("Searching Slack developer docs: \"%s\"", searchFlag)
	} else {
		// Default docs homepage
		docsURL = "https://docs.slack.dev"
		sectionText = "Slack developer docs"
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  sectionText,
		Secondary: []string{
			docsURL,
		},
	}))

	clients.Browser().OpenURL(docsURL)

	// Add trace for analytics
	if searchFlag != "" {
		clients.IO.PrintTrace(ctx, slacktrace.DocsSearchSuccess, searchFlag)
	} else {
		clients.IO.PrintTrace(ctx, slacktrace.DocsSuccess)
	}

	return nil
}
