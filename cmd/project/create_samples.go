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
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/pflag"
)

//go:embed samples.tmpl
var embedPromptSamplesTmpl string

// promptSampleSelection gathers upstream samples to select from
func promptSampleSelection(ctx context.Context, clients *shared.ClientFactory, sampleRepos []create.GithubRepo) (string, error) {
	filteredRepos := []create.GithubRepo{}
	selection, err := clients.IO.SelectPrompt(ctx, "Select a language:",
		[]string{
			fmt.Sprintf("Bolt for JavaScript %s", style.Secondary("Node.js")),
			fmt.Sprintf("Bolt for Python %s", style.Secondary("Python")),
			fmt.Sprintf("Deno Slack SDK %s", style.Secondary("Deno")),
		},
		iostreams.SelectPromptConfig{
			Flags: []*pflag.Flag{
				clients.Config.Flags.Lookup("language"),
				clients.Config.Flags.Lookup("template"), // Skip filtering with a template
			},
			Required: false,
		},
	)
	if err != nil {
		return "", err
	} else if selection.Prompt {
		switch selection.Index {
		case 0:
			filteredRepos = filterRepos(sampleRepos, "node")
		case 1:
			filteredRepos = filterRepos(sampleRepos, "python")
		case 2:
			filteredRepos = filterRepos(sampleRepos, "deno")
		}
	} else if selection.Flag {
		filteredRepos = filterRepos(sampleRepos, selection.Option)
	}

	sortedRepos := sortRepos(filteredRepos)
	selectOptions := createSelectOptions(sortedRepos)

	var selectedTemplate string
	selection, err = clients.IO.SelectPrompt(ctx, "Select a sample to build upon:", selectOptions, iostreams.SelectPromptConfig{
		Description: func(value string, index int) string {
			return sortedRepos[index].Description + "\n  https://github.com/" + sortedRepos[index].FullName
		},
		Flag:     clients.Config.Flags.Lookup("template"),
		PageSize: 4, // Supports standard terminal height (24 rows)
		Required: true,
		Template: embedPromptSamplesTmpl,
	})
	if err != nil {
		return "", err
	} else if selection.Flag {
		selectedTemplate = selection.Option
	} else if selection.Prompt {
		selectedTemplate = sortedRepos[selection.Index].FullName
	}
	return selectedTemplate, nil
}

// filterRepos returns a list of samples matching the provided project type
// according to the project naming conventions of @slack-samples.
//
// Ex: "node" matches both "bolt-js" and "bolt-ts" prefixed samples.
func filterRepos(sampleRepos []create.GithubRepo, projectType string) []create.GithubRepo {
	filteredRepos := make([]create.GithubRepo, 0)
	for _, s := range sampleRepos {
		search := strings.TrimSpace(strings.ToLower(projectType))
		switch search {
		case "java":
			if strings.HasPrefix(s.Name, "bolt-java") {
				filteredRepos = append(filteredRepos, s)
			}
		case "node":
			if strings.HasPrefix(s.Name, "bolt-js") || strings.HasPrefix(s.Name, "bolt-ts") {
				filteredRepos = append(filteredRepos, s)
			}
		case "python":
			if strings.HasPrefix(s.Name, "bolt-python") {
				filteredRepos = append(filteredRepos, s)
			}
		case "deno":
			fallthrough
		default:
			if strings.HasPrefix(s.Name, search) || search == "" {
				filteredRepos = append(filteredRepos, s)
			}
		}
	}
	return filteredRepos
}

// sortRepos sorts the provided repositories by the
// StargazersCount field in descending order
func sortRepos(sampleRepos []create.GithubRepo) []create.GithubRepo {
	sortedRepos := sampleRepos
	sort.Slice(sortedRepos, func(i, j int) bool {
		return sortedRepos[i].StargazersCount > sortedRepos[j].StargazersCount
	})
	return sortedRepos
}

// createSelectOptions takes in a list of repositories
// and returns an array of strings, each value being
// equal to the repository name (ie, deno-starter-template)
// and prepended with a number for a prompt visual aid
func createSelectOptions(filteredRepos []create.GithubRepo) []string {
	// Create a slice of repository names to use as
	// the primary item selection in the prompt
	selectOptions := make([]string, 0)
	for i, f := range filteredRepos {
		selectOption := fmt.Sprint(i+1, ". ", f.Name)
		selectOptions = append(selectOptions, selectOption)
	}
	return selectOptions
}
