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
	"context"
	"fmt"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flags
var samplesTemplateURLFlag string
var samplesGitBranchFlag string
var samplesListFlag bool
var samplesLanguageFlag string

func NewSamplesCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "samples [name]",
		Aliases: []string{"sample"},
		Short:   "List available sample apps",
		Long:    "List and create an app from the available samples",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Select a sample app to create",
				Command: "samples my-project",
			},
		}),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return runSamplesCommand(clients, cmd, args)
		},
	}

	cmd.Flags().StringVarP(&samplesTemplateURLFlag, "template", "t", "", "template URL for your app")
	cmd.Flags().StringVarP(&samplesGitBranchFlag, "branch", "b", "", "name of git branch to checkout")
	cmd.Flags().StringVar(&samplesLanguageFlag, "language", "", "runtime for the app framework\n  ex: \"deno\", \"node\", \"python\"")
	cmd.Flags().BoolVar(&samplesListFlag, "list", false, "print recommended samples")

	return cmd
}

// runSamplesCommand prompts for a sample then clones with the create command
func runSamplesCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	sampler := api.NewHTTPClient(api.HTTPClientOptions{
		TotalTimeOut: 60 * time.Second,
	})
	samples, err := create.GetSampleRepos(sampler)
	if err != nil {
		return err
	}
	if samplesListFlag || !clients.IO.IsTTY() {
		err := listSampleSelection(ctx, clients, samples)
		if err != nil {
			return err
		}
		return nil
	}
	selectedSample, err := promptSampleSelection(ctx, clients, samples)
	if err != nil {
		return err
	}

	// Instantiate the `create` command to call it using programmatically set flags
	createCmd := NewCreateCommand(clients)

	// Prepare template and branch flags with selected or provided repo values
	if err := createCmd.Flag("template").Value.Set(selectedSample); err != nil {
		return err
	}
	createCmd.Flag("template").Changed = true
	if err := createCmd.Flag("branch").Value.Set(samplesGitBranchFlag); err != nil {
		return err
	}
	createCmd.Flag("branch").Changed = cmd.Flag("branch").Changed

	// If preferred directory name is passed in as an argument to the `create`
	// command first, honor that preference and use it to create the project
	createCmd.SetArgs(args)

	// Execute the `create` command with the set flag
	return createCmd.ExecuteContext(ctx)
}

// listSampleSelection outputs available samples matching a language flag filter
func listSampleSelection(ctx context.Context, clients *shared.ClientFactory, sampleRepos []create.GithubRepo) error {
	filteredRepos := filterRepos(sampleRepos, samplesLanguageFlag)
	sortedRepos := sortRepos(filteredRepos)
	message := ""
	if samplesLanguageFlag != "" {
		message = fmt.Sprintf("Listing %d \"%s\" project samples", len(sortedRepos), samplesLanguageFlag)
	} else {
		message = fmt.Sprintf("Listing %d project samples", len(sortedRepos))
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "house_buildings",
		Text:  "Samples",
		Secondary: []string{
			message,
		},
	}))
	emojis := []string{
		"microscope",
		"test_tube",
		"petri_dish",
		"dna",
	}
	for ii, sample := range sortedRepos {
		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji: emojis[ii%len(emojis)],
			Text:  sample.Name,
			Secondary: []string{
				sample.Description,
				fmt.Sprintf("https://github.com/%s", sample.FullName),
			},
		}))
	}
	return nil
}
