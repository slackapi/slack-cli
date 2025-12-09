// Copyright 2022-2025 Salesforce, Inc.
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
	"fmt"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func getSelectionOptions(clients *shared.ClientFactory, categoryID string) []promptObject {
	if strings.TrimSpace(categoryID) == "" {
		categoryID = "slack-cli#getting-started"
	}

	// App categories and templates
	templatePromptObjects := map[string]([]promptObject){
		"slack-cli#getting-started": []promptObject{
			{
				Title:      fmt.Sprintf("Bolt for JavaScript %s", style.Secondary("Node.js")),
				Repository: "slack-samples/bolt-js-starter-template",
			},
			{
				Title:      fmt.Sprintf("Bolt for Python %s", style.Secondary("Python")),
				Repository: "slack-samples/bolt-python-starter-template",
			},
		},
		"slack-cli#automation-apps": {
			{
				Title:      fmt.Sprintf("Bolt for JavaScript %s", style.Secondary("Node.js")),
				Repository: "slack-samples/bolt-js-custom-function-template",
			},
			{
				Title:      fmt.Sprintf("Bolt for Python %s", style.Secondary("Python")),
				Repository: "slack-samples/bolt-python-custom-function-template",
			},
			{
				Title:      fmt.Sprintf("Deno Slack SDK %s", style.Secondary("Deno")),
				Repository: "slack-samples/deno-starter-template",
			},
		},
		"slack-cli#ai-apps": {
			{
				Title:      fmt.Sprintf("Bolt for JavaScript %s", style.Secondary("Node.js")),
				Repository: "slack-samples/bolt-js-assistant-template",
			},
			{
				Title:      fmt.Sprintf("Bolt for Python %s", style.Secondary("Python")),
				Repository: "slack-samples/bolt-python-assistant-template",
			},
		},
	}

	return templatePromptObjects[categoryID]
}

func getSelectionOptionsForCategory(clients *shared.ClientFactory) []promptObject {
	return []promptObject{
		{
			Title:      fmt.Sprintf("Starter App %s", style.Secondary("Getting started Slack app")),
			Repository: "slack-cli#getting-started",
		},
		{
			Title:      fmt.Sprintf("Agentic AI App %s %s", style.Secondary("Slack agents and assistants"), style.Emoji("sparkles")),
			Repository: "slack-cli#ai-apps",
		},
		{
			Title:      fmt.Sprintf("Automation App %s", style.Secondary("Custom steps and workflows")),
			Repository: "slack-cli#automation-apps",
		},
		{
			Title:      "View more samples",
			Repository: viewMoreSamples,
		},
	}
}

// promptTemplateSelection prompts the user to select a project template
func promptTemplateSelection(cmd *cobra.Command, clients *shared.ClientFactory) (create.Template, error) {
	ctx := cmd.Context()
	var categoryID string
	var selectedTemplate string

	// Prompt for the category
	promptForCategory := "Select an app:"
	optionsForCategory := getSelectionOptionsForCategory(clients)
	titlesForCategory := make([]string, len(optionsForCategory))
	for i, m := range optionsForCategory {
		titlesForCategory[i] = m.Title
	}
	templateForCategory := getSelectionTemplate(clients)

	// Print a trace with info about the category title options provided by CLI
	clients.IO.PrintTrace(ctx, slacktrace.CreateCategoryOptions, strings.Join(titlesForCategory, ", "))

	// Prompt to choose a category
	selection, err := clients.IO.SelectPrompt(ctx, promptForCategory, titlesForCategory, iostreams.SelectPromptConfig{
		Description: func(value string, index int) string {
			return optionsForCategory[index].Description
		},
		Flag:     clients.Config.Flags.Lookup("template"),
		Required: true,
		Template: templateForCategory,
	})
	if err != nil {
		return create.Template{}, slackerror.ToSlackError(err)
	} else if selection.Flag {
		selectedTemplate = selection.Option
	} else if selection.Prompt {
		categoryID = optionsForCategory[selection.Index].Repository
	}

	// Set template to view more samples, so the sample prompt is triggered
	if categoryID == viewMoreSamples {
		selectedTemplate = viewMoreSamples
	}

	// Prompt for the template
	if selectedTemplate == "" {
		prompt := "Select a language:"
		options := getSelectionOptions(clients, categoryID)
		titles := make([]string, len(options))
		for i, m := range options {
			titles[i] = m.Title
		}
		template := getSelectionTemplate(clients)

		// Print a trace with info about the template title options provided by CLI
		clients.IO.PrintTrace(ctx, slacktrace.CreateTemplateOptions, strings.Join(titles, ", "))

		// Prompt to choose a template
		selection, err := clients.IO.SelectPrompt(ctx, prompt, titles, iostreams.SelectPromptConfig{
			Description: func(value string, index int) string {
				return options[index].Description
			},
			Flag:     clients.Config.Flags.Lookup("template"),
			Required: true,
			Template: template,
		})
		if err != nil {
			return create.Template{}, slackerror.ToSlackError(err)
		} else if selection.Flag {
			selectedTemplate = selection.Option
		} else if selection.Prompt {
			selectedTemplate = options[selection.Index].Repository
		}
	}

	// Ensure user is okay to proceed if template source is from a non-trusted source
	switch selectedTemplate {
	case viewMoreSamples:
		sampler := api.NewHTTPClient(api.HTTPClientOptions{
			TotalTimeOut: 60 * time.Second,
		})
		samples, err := create.GetSampleRepos(sampler)
		if err != nil {
			return create.Template{}, err
		}
		selectedSample, err := promptSampleSelection(ctx, clients, samples)
		if err != nil {
			return create.Template{}, err
		}
		return create.ResolveTemplateURL(selectedSample)
	default:
		template, err := create.ResolveTemplateURL(selectedTemplate)
		if err != nil {
			return create.Template{}, err
		}
		confirm, err := confirmExternalTemplateSelection(cmd, clients, template)
		if err != nil {
			return create.Template{}, slackerror.ToSlackError(err)
		} else if !confirm {
			return create.Template{}, slackerror.New(slackerror.ErrUntrustedSource)
		}
		return template, nil
	}
}

// confirmExternalTemplateSelection prompts the user to confirm that they want to create an app from
// an external template and saves their preference if they choose to ignore future warnings
func confirmExternalTemplateSelection(cmd *cobra.Command, clients *shared.ClientFactory, template create.Template) (bool, error) {
	ctx := cmd.Context()

	trustSources, err := clients.Config.SystemConfig.GetTrustUnknownSources(ctx)
	if err != nil {
		return false, err
	}
	if trustSources || template.IsTrusted() {
		return true, nil
	}

	clients.IO.PrintWarning(ctx, style.Sectionf(style.TextSection{
		Text: style.Bold("You are trying to use code published by an unknown author"),
		Secondary: []string{
			"We strongly advise reviewing the source code and dependencies of external",
			"projects before use",
		},
	}))

	selection, err := clients.IO.SelectPrompt(ctx, "Proceed?", []string{"Yes", "Yes, don't ask again", "No"}, iostreams.SelectPromptConfig{
		Required: true,
		Flag:     clients.Config.Flags.Lookup("force"),
	})
	if err != nil {
		return false, err
	} else if selection.Option == "No" {
		return false, nil
	} else if selection.Option == "Yes, don't ask again" {
		err = clients.Config.SystemConfig.SetTrustUnknownSources(ctx, true)
		if err != nil {
			return true, slackerror.Wrap(err, "failed to set trust_unknown_sources property to config")
		}
	}
	return true, nil
}

// getSelectionTemplate returns a custom formatted template used for selecting a
// project template during creation
func getSelectionTemplate(clients *shared.ClientFactory) string {
	samplesURL := style.LinkText("https://docs.slack.dev/samples")
	return fmt.Sprintf(`
{{- define "option"}}
{{- if eq .SelectedIndex .CurrentIndex }}{{color .Config.Icons.SelectFocus.Format }}{{ .Config.Icons.SelectFocus.Text }} {{else}}{{color "default+hb"}}  {{end}}
{{- .CurrentOpt.Value}}{{color "reset"}}{{ if ne ($.GetDescription .CurrentOpt) "" }}{{"\n  "}}{{color "250"}}{{ $.GetDescription .CurrentOpt }}{{"\n"}}{{end}}
{{- color "reset"}}
{{end}}
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }}{{color "reset"}}
{{- if .ShowAnswer}}{{color "39+b"}} {{.Answer}}{{color "reset"}}
{{- else}}
  {{- " "}}{{- color "39+b"}}[Use arrows to move]{{color "reset"}}
  {{- "\n"}}
  {{- range $ix, $option := .PageEntries}}
    {{- template "option" $.IterateOption $ix $option}}
  {{- end}}
  {{- "Guided tutorials can be found at %s"}}{{color "reset"}}
{{end}}
`,
		samplesURL)
}
