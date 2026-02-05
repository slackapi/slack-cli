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

package docgen

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
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
	err = genMarkdownTree(rootCmd, clients.Fs, commandsDirPath)
	if err != nil {
		return slackerror.New(slackerror.ErrDocumentationGenerationFailed).WithRootCause(err)
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

// genMarkdownTree creates markdown reference of commands for the docs site.
//
// Reference: https://github.com/spf13/cobra/blob/3f3b81882534a51628f3286e93c6842d9b2e29ea/doc/md_docs.go#L119-L158
func genMarkdownTree(cmd *cobra.Command, fs afero.Fs, dir string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genMarkdownTree(c, fs, dir); err != nil {
			return err
		}
	}
	basename := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
	filename := filepath.Join(dir, basename)
	f, err := fs.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := genMarkdownCommand(cmd, f); err != nil {
		return err
	}
	return nil
}

// genMarkdownCommand creates custom markdown output for a command.
//
// Reference: https://github.com/spf13/cobra/blob/3f3b81882534a51628f3286e93c6842d9b2e29ea/doc/md_docs.go#L56-L117
func genMarkdownCommand(cmd *cobra.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	fmt.Fprintf(buf, "# `%s`\n\n", name)
	fmt.Fprintf(buf, "%s\n\n", cmd.Short)
	if len(cmd.Long) > 0 {
		fmt.Fprintf(buf, "## Description\n\n")
		description, err := render(cmd.Long)
		if err != nil {
			return err
		}
		fmt.Fprintf(buf, "%s\n\n", description)
	}
	if cmd.Runnable() {
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.UseLine())
	}
	if err := genMarkdownCommandFlags(buf, cmd); err != nil {
		return err
	}
	if len(cmd.Example) > 0 {
		fmt.Fprintf(buf, "## Examples\n\n")
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.Example)
	}
	if hasSeeAlso(cmd) {
		fmt.Fprintf(buf, "## See also\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := strings.ReplaceAll(pname, " ", "_")
			fmt.Fprintf(buf, "* [%s](%s)\t - %s\n", pname, link, parent.Short)
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}
		children := cmd.Commands()
		slices.SortFunc(children, func(a *cobra.Command, b *cobra.Command) int {
			if a.Name() < b.Name() {
				return -1
			} else {
				return 1
			}
		})
		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := strings.ReplaceAll(cname, " ", "_")
			fmt.Fprintf(buf, "* [%s](%s)\t - %s\n", cname, link, child.Short)
		}
		fmt.Fprintf(buf, "\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

// genMarkdownCommandFlags outputs flag information.
//
// Reference: https://github.com/spf13/cobra/blob/3f3b81882534a51628f3286e93c6842d9b2e29ea/doc/md_docs.go#L32-L49
func genMarkdownCommandFlags(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		fmt.Fprintf(buf, "## Flags\n\n```\n")
		flags.PrintDefaults()
		fmt.Fprintf(buf, "```\n\n")
	}
	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		fmt.Fprintf(buf, "## Global flags\n\n```\n")
		parentFlags.PrintDefaults()
		fmt.Fprintf(buf, "```\n\n")
	}
	return nil
}

// hasSeeAlso checks for adjancet commands to include in reference.
//
// Reference: https://github.com/spf13/cobra/blob/3f3b81882534a51628f3286e93c6842d9b2e29ea/doc/util.go#L23-L37
func hasSeeAlso(cmd *cobra.Command) bool {
	if cmd.HasParent() {
		return true
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		return true
	}
	return false
}

// render formats the templating from a command into markdown.
func render(input string) (string, error) {
	tmpl, err := template.New("md").Funcs(template.FuncMap{
		"Emoji": func(s string) string {
			return ""
		},
		"LinkText": func(s string) string {
			return fmt.Sprintf("[%s](%s)", s, s)
		},
		"ToBold": func(s string) string {
			return fmt.Sprintf("**%s**", s)
		},
	}).Parse(input)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}
