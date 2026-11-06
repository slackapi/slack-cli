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
	"net/url"
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type getConfig struct {
	output string
}

// docsPage is the JSON output of the get command
type docsPage struct {
	URL         string `json:"url"`
	MarkdownURL string `json:"markdown_url"`
	Content     string `json:"content"`
}

// resolvePageURL turns a docs page reference, either a site path or a full
// docs URL with or without ".md", into the page's absolute URL
func resolvePageURL(page string) (string, error) {
	docs, _ := url.Parse(docsURL)
	if strings.HasPrefix(page, docs.Host+"/") {
		page = docs.Scheme + "://" + page
	}
	u, err := url.Parse(page)
	if err != nil {
		return "", err
	}
	if u.Host != "" && !strings.EqualFold(u.Hostname(), docs.Hostname()) {
		return "", slackerror.New(slackerror.ErrInvalidArgs).
			WithMessage("The page %s is not on %s", page, docs.Host).
			WithRemediation("Find the page with %s", style.Commandf("docs search", false))
	}
	path := u.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimSuffix(path, ".md")
	page = docsURL + path
	if u.Fragment != "" {
		page += "#" + u.Fragment
	}
	return page, nil
}

func NewGetCommand(clients *shared.ClientFactory) *cobra.Command {
	cfg := &getConfig{}

	cmd := &cobra.Command{
		Use:   "get <page>",
		Short: "Print a Slack developer docs page",
		Long: strings.Join([]string{
			"Print a page of the Slack developer docs as markdown, return it as JSON, or open",
			"it in your browser.",
			"",
			"The page can be a path or a full URL on docs.slack.dev, like a link returned by",
			"the search subcommand.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Print the Socket Mode guide for Bolt for JavaScript",
				Command: "docs get /tools/bolt-js/concepts/socket-mode",
			},
			{
				Meaning: "Return a page and its links as JSON",
				Command: "docs get https://docs.slack.dev/reference/methods/chat.postMessage --output=json",
			},
			{
				Meaning: "Open a page in your browser",
				Command: "docs get /block-kit --output=browser",
			},
		}),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocsGetCommand(clients, cmd, args[0], cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.output, "output", "text", "output format: text, json, browser")

	return cmd
}

func runDocsGetCommand(clients *shared.ClientFactory, cmd *cobra.Command, page string, cfg *getConfig) error {
	ctx := cmd.Context()

	pageURL, err := resolvePageURL(page)
	if err != nil {
		return err
	}

	switch cfg.output {
	case "text", "json":
		markdownURL := makeMarkdownURL(pageURL)
		if markdownURL == pageURL {
			return slackerror.New(slackerror.ErrInvalidArgs).
				WithMessage("The page %s has no markdown version", pageURL).
				WithRemediation("Open it in your browser with %s", style.Highlight("--output=browser"))
		}
		u, _ := url.Parse(markdownURL)
		content, err := clients.API().DocsGetPage(ctx, u.Path)
		if err != nil {
			return slackerror.New(slackerror.ErrHTTPRequestFailed).
				WithMessage("The page %s could not be loaded", markdownURL).
				WithRootCause(err).
				WithRemediation("Check the page path, or find the page with %s", style.Commandf("docs search", false))
		}
		if cfg.output == "json" {
			encoder := json.NewEncoder(clients.IO.WriteOut())
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(docsPage{URL: pageURL, MarkdownURL: markdownURL, Content: content}); err != nil {
				return slackerror.New(slackerror.ErrUnableToParseJSON).WithRootCause(err)
			}
		} else {
			if _, err := clients.IO.WriteOut().Write([]byte(content)); err != nil {
				return err
			}
		}
		clients.IO.PrintTrace(ctx, slacktrace.DocsGetSuccess, pageURL)
		return nil
	case "browser":
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji:     "books",
			Text:      "Docs",
			Secondary: []string{pageURL},
		}))
		clients.Browser().OpenURL(pageURL)
		clients.IO.PrintTrace(ctx, slacktrace.DocsGetSuccess, pageURL)
		return nil
	default:
		return slackerror.New(slackerror.ErrInvalidFlag).WithMessage(
			"Invalid output format: %s", cfg.output,
		).WithRemediation(
			"Use one of: text, json, browser",
		)
	}
}
