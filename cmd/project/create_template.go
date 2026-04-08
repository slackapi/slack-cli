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

// getSelectionOptions returns the app template options for a given category.
func getSelectionOptions(categoryID string) []promptObject {
	templatePromptObjects := map[string]([]promptObject){
		"slack-cli#getting-started": {
			{
				Title:      fmt.Sprintf("Bolt for JavaScript %s", style.Secondary("Node.js")),
				Repository: "slack-samples/bolt-js-starter-template",
			},
			{
				Title:      fmt.Sprintf("Bolt for Python %s", style.Secondary("Python")),
				Repository: "slack-samples/bolt-python-starter-template",
			},
		},
		"slack-cli#ai-apps": {
			{
				Title:      fmt.Sprintf("Support Agent %s", style.Secondary("Resolve IT support cases")),
				Repository: "slack-cli#ai-apps/support-agent",
			},
			{
				Title:      fmt.Sprintf("Starter Agent %s", style.Secondary("Start from scratch")),
				Repository: "slack-cli#ai-apps/starter-agent",
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
	}
	return templatePromptObjects[categoryID]
}

// getFrameworkOptions returns the framework choices for a given template.
func getFrameworkOptions(template string) []promptObject {
	frameworkPromptObjects := map[string][]promptObject{
		"slack-cli#ai-apps/support-agent": {
			{
				Title:      fmt.Sprintf("Claude Agent SDK %s", style.Secondary("Bolt for Python")),
				Repository: "slack-samples/bolt-python-support-agent",
				Subdir:     "claude-agent-sdk",
			},
			{
				Title:      fmt.Sprintf("OpenAI Agents SDK %s", style.Secondary("Bolt for Python")),
				Repository: "slack-samples/bolt-python-support-agent",
				Subdir:     "openai-agents-sdk",
			},
			{
				Title:      fmt.Sprintf("Pydantic AI %s", style.Secondary("Bolt for Python")),
				Repository: "slack-samples/bolt-python-support-agent",
				Subdir:     "pydantic-ai",
			},
		},
		"slack-cli#ai-apps/starter-agent": {
			{
				Title:      fmt.Sprintf("Bolt for JavaScript %s", style.Secondary("Node.js")),
				Repository: "slack-samples/bolt-js-starter-agent",
			},
			{
				Title:      fmt.Sprintf("Bolt for Python %s", style.Secondary("Python")),
				Repository: "slack-samples/bolt-python-starter-agent",
			},
		},
	}
	return frameworkPromptObjects[template]
}

// getSelectionOptionsForCategory returns the top-level category options for
// the create command template selection.
func getSelectionOptionsForCategory(clients *shared.ClientFactory) []promptObject {
	return []promptObject{
		{
			Title:      fmt.Sprintf("Starter app %s", style.Secondary("Getting started Slack app")),
			Repository: "slack-cli#getting-started",
		},
		{
			Title:      fmt.Sprintf("AI Agent app %s", style.Secondary("Slack agents and assistants")),
			Repository: "slack-cli#ai-apps",
		},
		{
			Title:      fmt.Sprintf("Automation app %s", style.Secondary("Custom steps and workflows")),
			Repository: "slack-cli#automation-apps",
		},
		{
			Title:      "View more samples",
			Repository: viewMoreSamples,
		},
	}
}

// promptTemplateSelection prompts the user to select a project template
func promptTemplateSelection(cmd *cobra.Command, clients *shared.ClientFactory, categoryShortcut string) (create.Template, error) {
	ctx := cmd.Context()
	var categoryID string

	// Check if a category shortcut was provided
	if categoryShortcut != "" {
		switch categoryShortcut {
		case "agent":
			categoryID = "slack-cli#ai-apps"
		default:
			return create.Template{}, slackerror.New(slackerror.ErrInvalidArgs).
				WithMessage("The %s category was not found", categoryShortcut)
		}
	} else {
		// Prompt for the category
		promptForCategory := "Select a category:"
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
			Flag:     clients.Config.Flags.Lookup("template"),
			Required: true,
			Template: templateForCategory,
		})
		if err != nil {
			return create.Template{}, slackerror.ToSlackError(err)
		} else if selection.Flag {
			template, err := create.ResolveTemplateURL(selection.Option)
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
		} else if selection.Prompt {
			categoryID = optionsForCategory[selection.Index].Repository
		}

		if categoryID == viewMoreSamples {
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
		}
	}

	// Prompt for the example template
	prompt := "Select a framework:"
	if categoryID == "slack-cli#ai-apps" {
		prompt = "Select a template:"
	}
	options := getSelectionOptions(categoryID)
	titles := make([]string, len(options))
	for i, m := range options {
		titles[i] = m.Title
	}
	clients.IO.PrintTrace(ctx, slacktrace.CreateTemplateOptions, strings.Join(titles, ", "))

	selection, err := clients.IO.SelectPrompt(ctx, prompt, titles, iostreams.SelectPromptConfig{
		Description: func(value string, index int) string {
			return options[index].Description
		},
		Required: true,
		Template: getSelectionTemplate(clients),
	})
	if err != nil {
		return create.Template{}, err
	} else if selection.Flag {
		return create.Template{}, slackerror.New(slackerror.ErrPrompt)
	} else if selection.Prompt && !strings.HasPrefix(options[selection.Index].Repository, "slack-cli#") {
		return create.ResolveTemplateURL(options[selection.Index].Repository)
	}
	template := options[selection.Index].Repository

	// Prompt for the example framework
	examples := getFrameworkOptions(template)
	choices := make([]string, len(examples))
	for i, opt := range examples {
		choices[i] = opt.Title
	}
	choice, err := clients.IO.SelectPrompt(ctx, "Select a framework:", choices, iostreams.SelectPromptConfig{
		Description: func(value string, index int) string {
			return examples[index].Description
		},
		Required: true,
		Template: getSelectionTemplate(clients),
	})
	if err != nil {
		return create.Template{}, err
	} else if choice.Flag {
		return create.Template{}, slackerror.New(slackerror.ErrPrompt)
	}
	example := examples[choice.Index]
	resolved, err := create.ResolveTemplateURL(example.Repository)
	if err != nil {
		return create.Template{}, err
	}
	if example.Subdir != "" {
		resolved.SetSubdir(example.Subdir)
	}
	return resolved, nil
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

	clients.IO.PrintWarning(ctx, "%s", style.Sectionf(style.TextSection{
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

// listTemplates prints available templates for the create command
func listTemplates(ctx context.Context, clients *shared.ClientFactory, categoryShortcut string) error {
	type categoryInfo struct {
		id   string
		name string
	}

	var categories []categoryInfo
	if categoryShortcut == "agent" {
		categories = []categoryInfo{
			{id: "slack-cli#ai-apps/support-agent", name: "Support agent"},
			{id: "slack-cli#ai-apps/starter-agent", name: "Starter agent"},
		}
	} else {
		categories = []categoryInfo{
			{id: "slack-cli#getting-started", name: "Getting started"},
			{id: "slack-cli#ai-apps/support-agent", name: "Support agent"},
			{id: "slack-cli#ai-apps/starter-agent", name: "Starter agent"},
			{id: "slack-cli#automation-apps", name: "Automation apps"},
		}
	}

	for _, category := range categories {
		var secondary []string
		if frameworks := getFrameworkOptions(category.id); len(frameworks) > 0 {
			for _, tmpl := range frameworks {
				repo := tmpl.Repository
				if tmpl.Subdir != "" {
					repo = fmt.Sprintf("%s --subdir %s", repo, tmpl.Subdir)
				}
				secondary = append(secondary, repo)
			}
		} else {
			for _, tmpl := range getSelectionOptions(category.id) {
				secondary = append(secondary, tmpl.Repository)
			}
		}
		clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
			Emoji:     "house_buildings",
			Text:      style.Bold(category.name),
			Secondary: secondary,
		}))
	}

	return nil
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
