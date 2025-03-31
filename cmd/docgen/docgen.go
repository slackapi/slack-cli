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

package docgen

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// defaultDocsDirPath is the default value for the <path> argument
const defaultDocsDirPath = "docs"

//go:embed errors.tmpl
var errorsMarkdownTemplate string

// NewCommand creates a new Cobra command instance
func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := cobra.Command{
		Hidden: true,
		Use:    "docgen <path>",
		Short:  "Generate documentation for each command",
		Long: strings.Join([]string{
			"Generate documentation for each command as markdown files",
			"",
			"The generated files are output to the <path> directory (default: \"docs/\")",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Generate command documentation in the docs/ directory",
				Command: "docgen",
			},
			{
				Meaning: "Save generated documentation to the commands/ directory",
				Command: "docgen commands/",
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocGenCommandFunc(clients, cmd, args)
		},
	}
	return &cmd
}

// runDocGenCommandFunc generates and save command reference files
func runDocGenCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get the top-most command (`slack`) to generate the documentation tree and
	// fallback on current command
	var rootCmd = cmd.Parent()
	if rootCmd == nil {
		rootCmd = cmd
	}

	// Print a section without ending in a blank newline
	clients.IO.PrintInfo(ctx, false, "\n%s%s",
		style.Emoji("books"), "Generate Documentation")
	clients.IO.PrintInfo(ctx, false, style.Indent("%s"), style.Secondary(
		"Writing references to commands and errors in a markdown format",
	))

	// Save output relative to the current path
	var docsDirPath = defaultDocsDirPath
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		docsDirPath = args[0]
	}
	workingDirPath, err := clients.Os.Getwd()
	if err != nil {
		workingDirPath = "."
	}
	docsDirPath = filepath.Join(workingDirPath, docsDirPath)
	if err := clients.Fs.MkdirAll(docsDirPath, 0755); err != nil {
		return slackerror.New("MkdirAll failed").WithRootCause(err)
	}

	// Generate command reference
	commandsDirPath := filepath.Join(docsDirPath, "commands")
	if err := clients.Fs.MkdirAll(commandsDirPath, 0755); err != nil {
		return slackerror.New("MkdirAll failed").WithRootCause(err)
	}
	rootCmd.DisableAutoGenTag = true
	err = clients.Cobra.GenMarkdownTree(rootCmd, commandsDirPath)
	if err != nil {
		return slackerror.New("Cobra.GenMarkdownTree failed").WithRootCause(err)
	}

	// Generate errors reference
	file, err := clients.Fs.Create(filepath.Join(docsDirPath, "errors.md"))
	if err != nil {
		return err
	}
	defer file.Close()
	tmpl, err := template.New("errors").Parse(errorsMarkdownTemplate)
	if err != nil {
		return err
	}
	err = tmpl.Execute(file, slackerror.ErrorCodeMap)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, style.Indent("%s\n"), style.Secondary(
		fmt.Sprintf("References saved to: %s", docsDirPath),
	))
	return nil
}
