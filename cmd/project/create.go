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
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flags
var createTemplateURLFlag string
var createGitBranchFlag string

// Handle to client's create function used for testing
// TODO - Find best practice, such as using an Interface and Struct to create a client
var CreateFunc = create.Create

var appCreateSpinner *style.Spinner

const copyTemplate = "Copying"
const cloneTemplate = "Cloning"

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
		Use:        "create [name] [flags]",
		Short:      "Create a new Slack project",
		Long:       `Create a new Slack project on your local machine from an optional template`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "create my-project", Meaning: "Create a new project from a template"},
			{Command: "create my-project -t slack-samples/deno-hello-world", Meaning: "Start a new project from a specific template"},
		}),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return runCreateCommand(clients, cmd, args)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&createTemplateURLFlag, "template", "t", "", "template URL for your app")
	cmd.Flags().StringVarP(&createGitBranchFlag, "branch", "b", "", "name of git branch to checkout")

	return cmd
}

func runCreateCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {

	// Set up event logger
	log := newCreateLogger(clients, cmd)

	// Get optional app name passed as an arg
	appNameArg := ""
	if len(args) > 0 && args[0] != "samples" && args[0] != "create" {
		appNameArg = args[0]
	}

	// Collect the template URL or select a starting template
	template, err := promptTemplateSelection(cmd, clients)
	if err != nil {
		return err
	}

	// Set up spinners
	appCreateSpinner = style.NewSpinner(cmd.OutOrStdout())

	createArgs := create.CreateArgs{
		AppName:   appNameArg,
		Template:  template,
		GitBranch: createGitBranchFlag,
	}
	clients.EventTracker.SetAppTemplate(template.GetTemplatePath())

	ctx := cmd.Context()
	appDirPath, err := CreateFunc(ctx, clients, log, createArgs)
	if err != nil {
		printAppCreateError(clients, cmd, err)
		return err
	}
	printCreateSuccess(ctx, clients, appDirPath)
	return nil
}

/*
App creation is setting up local project directory
Events: on_app_create_completion
*/

// newCreateLogger creates a logger instance to receive event notifications
func newCreateLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_app_create_template_default":
				printAppCreateDefaultemplate(cmd, event)
			case "on_app_create_template_custom":
				printAppCreateCustomTemplate(cmd, event)
			case "on_app_create_completion":
				printProjectCreateCompletion(clients, cmd, event)
			default:
				// Ignore the event
			}
		},
	)
}

/*
App creation (not Create command) is cloning the template and creating the project directory
Events: on_app_create_template_custom, on_app_create_completion
*/

func printAppCreateDefaultemplate(cmd *cobra.Command, event *logger.LogEvent) {
	startAppCreateSpinner(copyTemplate)
}

// Print template URL if using custom app template
func printAppCreateCustomTemplate(cmd *cobra.Command, event *logger.LogEvent) {
	var verb string
	templatePath := event.DataToString("templatePath")
	isGit := event.DataToBool("isGit")
	gitBranch := event.DataToString("gitBranch")

	if isGit {
		verb = cloneTemplate
	} else {
		verb = copyTemplate
	}
	templateText := fmt.Sprintf(
		"%s template from %s",
		verb,
		templatePath,
	)

	if gitBranch != "" {
		templateText = fmt.Sprintf("%s (branch: %s)", templateText, gitBranch)
	}

	cmd.Print(style.Secondary(templateText), "\n\n")

	startAppCreateSpinner(verb)
}

func startAppCreateSpinner(verb string) {
	appCreateSpinner.Update(verb+" app template", "").Start()
}

// Display message and added files at completion of app creation
func printProjectCreateCompletion(clients *shared.ClientFactory, cmd *cobra.Command, event *logger.LogEvent) {
	createCompletionText := style.Sectionf(style.TextSection{
		Emoji: "gear",
		Text:  "Created project directory",
	})
	appCreateSpinner.Update(createCompletionText, "").Stop()
}

// printCreateSuccess outputs an informative message after creating a new app
func printCreateSuccess(ctx context.Context, clients *shared.ClientFactory, appPath string) {
	// Check if this is a Deno project to conditionally enable some features
	var isDenoProject = false
	if clients.Runtime != nil {
		isDenoProject = strings.Contains(strings.ToLower(clients.Runtime.Name()), "deno")
	}

	// Display the original next steps section when the Bolt Experiment is OFF
	// or when the Bolt Experiment is ON and a Deno SDK project is created
	if !clients.Config.WithExperimentOn(experiment.BoltFrameworks) || isDenoProject {
		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji: "compass",
			Text:  "Explore the documentation to learn more",
			Secondary: []string{
				"Read the README.md or peruse the docs over at " + style.Highlight("api.slack.com/automation"),
				"Find available commands and usage info with " + style.Commandf("help", false),
			},
		}))

		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
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

		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji:     "clipboard",
			Text:      "Next steps to begin development",
			Secondary: secondaryOutput,
		}))
	}
	clients.IO.PrintTrace(ctx, slacktrace.CreateSuccess)
}

// printAppCreateError stops the creation spinners and displays the returned error message
func printAppCreateError(clients *shared.ClientFactory, cmd *cobra.Command, err error) {
	switch {
	case appCreateSpinner.Active():
		errorText := fmt.Sprintf("Error creating project directory: %s", err)
		appCreateSpinner.Update(errorText, "warning").Stop()
	default:
	}
	clients.IO.PrintTrace(cmd.Context(), slacktrace.CreateError)
}
