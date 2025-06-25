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
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// NewInitCommand returns a Cobra command to initialize projects with Slack CLI support
func NewInitCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [flags]",
		Short: "Initialize a project to work with the Slack CLI",
		Long: strings.Join([]string{
			"Initializes a project to support the Slack CLI.",
			"",
			"Adds a .slack directory with the following files:",
			"- " + filepath.Join("project-name", ".slack"),
			"- " + filepath.Join("project-name", ".slack", ".gitignore"),
			"- " + filepath.Join("project-name", ".slack", "config.json"),
			"- " + filepath.Join("project-name", ".slack", "hooks.json"),
			"",
			"Adds the Slack CLI hooks dependency to your project:",
			"- Deno:    Unsupported",
			"- Node.js: Updates package.json",
			"- Python:  Updates requirements.txt",
			"",
			"Installs your project dependencies when supported:",
			"- Deno:    Supported",
			"- Node.js: Supported",
			"- Python:  Unsupported",
			"",
			"Adds an existing app to your project (optional):",
			"- Prompts to add an existing app from app settings",
			"- Runs the command " + style.Commandf("app link", false),
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Initialize a project",
				Command: "init",
			},
		}),
		Args: cobra.MaximumNArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return projectInitCommandPreRunE(clients, cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectInitCommandRunE(clients, cmd, args)
		},
	}

	// TODO - add --environment flag to be used by link

	return cmd
}

// projectInitCommandPreRunE determines if the command is supported for a project
// and configures flags
func projectInitCommandPreRunE(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	clients.Config.SetFlags(cmd)
	return nil
}

// projectInitCommandRunE sets an app environment variable to given values
func projectInitCommandRunE(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	clients.IO.PrintTrace(ctx, slacktrace.ProjectInitStarted)

	// Display header section
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "nut_and_bolt",
		Text:  "Project Initialization",
		Secondary: []string{
			"Initializing the project to work with the Slack CLI",
		},
	}))

	// Get the current directory to use as the base for the project
	projectDirPath, err := clients.Os.Getwd()
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
	}

	// Install the project dependencies, such as .slack/ and runtime packages
	// Existing projects initialized always default to config.ManifestSourceLocal.
	// The link command will switch it to config.ManifestSourceRemote
	_ = create.InstallProjectDependencies(ctx, clients, projectDirPath, config.ManifestSourceLocal)

	// Add an existing app to the project
	err = app.LinkExistingApp(ctx, clients, &types.App{}, true)
	if err != nil {
		// Display the error but continue to init
		clients.IO.PrintError(ctx, err.Error())
	}

	printNextStepSection(ctx, clients, projectDirPath)

	clients.IO.PrintTrace(ctx, slacktrace.ProjectInitSuccess)

	return nil
}

// printNextStepSection outputs a section with next steps to get started with development
func printNextStepSection(ctx context.Context, clients *shared.ClientFactory, projectDirPath string) {
	var secondaryOutput []string

	// Output about the README.md
	if _, err := clients.Fs.Stat(filepath.Join(projectDirPath, "README.md")); !clients.Os.IsNotExist(err) {
		secondaryOutput = append(secondaryOutput, "Learn more about the project in the "+style.Highlight("README.md"))
	}

	// Output about general usage
	secondaryOutput = append(secondaryOutput,
		"Add more existing apps to your project with "+style.Commandf("app link", true),
		"Start developing and see changes in real-time with "+style.Commandf("run", true),
		"When you're ready to deploy for production with "+style.Commandf("deploy", true),
	)

	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji:     "clipboard",
		Text:      "Next steps to begin development",
		Secondary: secondaryOutput,
	}))
}
