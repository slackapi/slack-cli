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

package project

import (
	"context"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
)

// templateSelectionResult holds the user's selections from the dynamic template form.
type templateSelectionResult struct {
	CategoryID   string // e.g. "slack-cli#getting-started" or viewMoreSamples
	TemplateRepo string // e.g. "slack-samples/bolt-js-starter-template"
}

// charmPromptTemplateSelectionFunc is a package-level function variable for test overriding.
var charmPromptTemplateSelectionFunc = charmPromptTemplateSelection

// buildTemplateSelectionForm constructs a single-screen huh form where the category
// and template selects are in the same group. Changing the category dynamically
// updates the template options via OptionsFunc.
func buildTemplateSelectionForm(clients *shared.ClientFactory, category *string, template *string) *huh.Form {
	categoryOptions := getSelectionOptionsForCategory(clients)
	var catOpts []huh.Option[string]
	for _, opt := range categoryOptions {
		catOpts = append(catOpts, huh.NewOption(opt.Title, opt.Repository))
	}

	categorySelect := huh.NewSelect[string]().
		Title("Select an app:").
		Options(catOpts...).
		Value(category)

	templateSelect := huh.NewSelect[string]().
		Title("Select a language:").
		OptionsFunc(func() []huh.Option[string] {
			if *category == viewMoreSamples {
				return []huh.Option[string]{
					huh.NewOption("Browse sample gallery...", viewMoreSamples),
				}
			}

			options := getSelectionOptions(clients, *category)
			var opts []huh.Option[string]
			for _, opt := range options {
				opts = append(opts, huh.NewOption(opt.Title, opt.Repository))
			}
			return opts
		}, category).
		Value(template)

	return huh.NewForm(
		huh.NewGroup(categorySelect, templateSelect),
	).WithTheme(style.ThemeSlack())
}

// charmPromptTemplateSelection runs the dynamic template selection form and returns the result.
func charmPromptTemplateSelection(ctx context.Context, clients *shared.ClientFactory) (templateSelectionResult, error) {
	// Print trace with category options
	categoryOptions := getSelectionOptionsForCategory(clients)
	categoryTitles := make([]string, len(categoryOptions))
	for i, opt := range categoryOptions {
		categoryTitles[i] = opt.Title
	}
	clients.IO.PrintTrace(ctx, slacktrace.CreateCategoryOptions, strings.Join(categoryTitles, ", "))

	var category string
	var template string
	err := buildTemplateSelectionForm(clients, &category, &template).Run()
	if err != nil {
		return templateSelectionResult{}, slackerror.ToSlackError(err)
	}

	// Print trace with template options
	templateOptions := getSelectionOptions(clients, category)
	templateTitles := make([]string, len(templateOptions))
	for i, opt := range templateOptions {
		templateTitles[i] = opt.Title
	}
	if len(templateTitles) > 0 {
		clients.IO.PrintTrace(ctx, slacktrace.CreateTemplateOptions, strings.Join(templateTitles, ", "))
	}

	return templateSelectionResult{
		CategoryID:   category,
		TemplateRepo: template,
	}, nil
}
