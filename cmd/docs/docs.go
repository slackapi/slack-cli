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
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

const docsURL = "https://docs.slack.dev"

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
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsCommand(clients, cmd)
		},
		// Disable automatic suggestions for unknown commands
		DisableSuggestions: true,
	}

	// Add the search subcommand
	cmd.AddCommand(NewSearchCommand(clients))

	// Catch removed --search flag
	cmd.Flags().BoolP("search", "", false, "DEPRECATED: use 'docs search' subcommand instead")

	return cmd
}

// runDocsCommand opens Slack developer docs in the browser
func runDocsCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()

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
