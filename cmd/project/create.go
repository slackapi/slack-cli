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
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flags
var createTemplateURLFlag string
var createGitBranchFlag string
var createAppNameFlag string
var createListFlag bool
var createSubdirFlag string

// Handle to client's create function used for testing
// TODO - Find best practice, such as using an Interface and Struct to create a client
var CreateFunc = create.Create

// promptObject describes the Github app template
type promptObject struct {
	Title       string // "Reverse string"
	Repository  string // "slack-samples/reverse-string"
	Description string // "A function that reverses a given string"
}

const viewMoreSamples = "slack-cli#view-more-samples"

func NewCreateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		SuggestFor: []string{"new"},
		Use:        "create [name | agent <name>] [flags]",
		Short:      "Create a new Slack project",
		Long: `Create a new Slack project on your local machine from an optional template.

The 'agent' argument is a shortcut to create an AI Agent app. If you want to
name your app 'agent' (not create an AI Agent), use the --name flag instead.`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "create my-project", Meaning: "Create a new project from a template"},
			{Command: "create agent my-agent-app", Meaning: "Create a new AI Agent app"},
			{Command: "create my-project -t slack-samples/deno-hello-world", Meaning: "Start a new project from a specific template"},
			{Command: "create --name my-project", Meaning: "Create a project named 'my-project'"},
			{Command: "create my-project -t org/monorepo --subdir apps/my-app", Meaning: "Create from a subdirectory of a template"},
		}),
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return runCreateCommand(clients, cmd, args)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&createTemplateURLFlag, "template", "t", "", "template URL for your app")
	cmd.Flags().StringVarP(&createGitBranchFlag, "branch", "b", "", "name of git branch to checkout")
	cmd.Flags().StringVarP(&createAppNameFlag, "name", "n", "", "name for your app (overrides the name argument)")
	cmd.Flags().BoolVar(&createListFlag, "list", false, "list available app templates")
	cmd.Flags().StringVar(&createSubdirFlag, "subdir", "", "subdirectory in the template to use as project")

	return cmd
}

func runCreateCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get optional app name passed as an arg and check for category shortcuts
	appNameArg := ""
	categoryShortcut := ""
	templateFlagProvided := cmd.Flags().Changed("template")
	nameFlagProvided := cmd.Flags().Changed("name")

	if len(args) > 0 {
		switch args[0] {
		case "samples", "create":
			// These are special commands, not app names
		case "agent":
			// Only treat as shortcut if --template flag is not provided
			if !templateFlagProvided {
				// Shortcut to AI apps category
				categoryShortcut = "agent"
				// Check if a second argument was provided as the app name
				if len(args) > 1 {
					appNameArg = args[1]
				}
			} else {
				// When --template is provided, "agent" is the app name
				appNameArg = args[0]
			}
		default:
			appNameArg = args[0]
		}
	}

	// --name flag overrides any positional app name argument
	// This allows users to name their app "agent" without triggering the AI Agent shortcut
	if nameFlagProvided {
		appNameArg = createAppNameFlag
	}

	// List templates and exit early if the --list flag is set
	if createListFlag {
		return listTemplates(ctx, clients, categoryShortcut)
	}

	// --subdir requires --template
	if cmd.Flags().Changed("subdir") && !templateFlagProvided {
		return slackerror.New(slackerror.ErrMismatchedFlags).
			WithMessage("The --subdir flag requires the --template flag")
	}

	// Collect the template URL or select a starting template
	template, err := promptTemplateSelection(cmd, clients, categoryShortcut)
	if err != nil {
		return err
	}

	// Prompt for app name if not provided via flag or argument
	if appNameArg == "" {
		if clients.IO.IsTTY() {
			defaultName := generateRandomAppName()
			name, err := clients.IO.InputPrompt(ctx, "Name your app:", iostreams.InputPromptConfig{
				Placeholder: defaultName,
			})
			if err != nil {
				return err
			}
			if name != "" {
				appNameArg = name
			} else {
				appNameArg = defaultName
			}
		} else {
			appNameArg = generateRandomAppName()
		}
	}

	createArgs := create.CreateArgs{
		AppName:   appNameArg,
		Template:  template,
		GitBranch: createGitBranchFlag,
		Subdir:    createSubdirFlag,
	}
	clients.EventTracker.SetAppTemplate(template.GetTemplatePath())

	appDirPath, err := CreateFunc(ctx, clients, createArgs)
	if err != nil {
		return err
	}

	printCreateSuccess(ctx, clients, appDirPath)
	return nil
}

// printCreateSuccess outputs an informative message after creating a new app
func printCreateSuccess(ctx context.Context, clients *shared.ClientFactory, appPath string) {
	// Check if this is a Deno project to conditionally enable some features
	var isDenoProject = false
	if clients.Runtime != nil {
		isDenoProject = strings.Contains(strings.ToLower(clients.Runtime.Name()), "deno")
	}

	// Include documentation and information about ROSI for deno apps
	if isDenoProject {
		clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
			Emoji: "compass",
			Text:  "Explore the documentation to learn more",
			Secondary: []string{
				"Read the README.md or peruse the docs over at " + style.Highlight("https://docs.slack.dev/tools/deno-slack-sdk"),
				"Find available commands and usage info with " + style.Commandf("help", false),
			},
		}))

		clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
			Emoji: "clipboard",
			Text:  "Follow the steps below to begin development",
			Secondary: []string{
				"Change into your project directory with " + style.CommandText(fmt.Sprintf("cd %s%s", appPath, string(filepath.Separator))),
				"Develop locally and see changes in real-time with " + style.Commandf("run", true),
				"When you're ready to deploy for production with " + style.Commandf("deploy", true),
			},
		}))
	} else {
		var secondaryOutput []string

		// Output about the README.md
		if _, err := clients.Fs.Stat(filepath.Join(appPath, "README.md")); !clients.Os.IsNotExist(err) {
			secondaryOutput = append(secondaryOutput, "Learn more about the project in the "+style.Highlight("README.md"))
		}

		// Output about general usage
		secondaryOutput = append(secondaryOutput,
			"Change into your project with "+style.CommandText(fmt.Sprintf("cd %s%s", appPath, string(filepath.Separator))),
			"Start developing and see changes in real-time with "+style.Commandf("run", true),
		)

		clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
			Emoji:     "clipboard",
			Text:      "Next steps to begin development",
			Secondary: secondaryOutput,
		}))
	}
	clients.IO.PrintTrace(ctx, slacktrace.CreateSuccess)
}

// generateRandomAppName will create a random app name based on two words and a number
func generateRandomAppName() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	var firstRandomNum = rand.Intn(len(create.Adjectives))
	var secondRandomNum = rand.Intn(len(create.Animals))
	var randomName = fmt.Sprintf("%s-%s-%d", create.Adjectives[firstRandomNum], create.Animals[secondRandomNum], rand.Intn(1000))
	return randomName
}
