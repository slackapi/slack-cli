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

package shell

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/spf13/cobra"
)

// BuildCommandTree is set by cmd.Init() to build a fresh command tree for each
// REPL iteration. This avoids a circular import between cmd/shell and cmd.
var BuildCommandTree func(context.Context, *shared.ClientFactory) *cobra.Command

// readLineFunc is a package-level function variable for test overriding.
var readLineFunc = readLine

// NewCommand creates the shell command.
func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Start an interactive shell session",
		Long:  "Start an interactive shell where commands can be entered without the 'slack' prefix.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !clients.Config.WithExperimentOn(experiment.Charm) {
				return fmt.Errorf("the shell command requires the charm experiment: --experiment=charm")
			}
			if !clients.IO.IsTTY() {
				return fmt.Errorf("the shell command requires an interactive terminal (TTY)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunShell(cmd.Context(), clients)
		},
	}
}

// RunShell implements the interactive REPL loop.
func RunShell(ctx context.Context, clients *shared.ClientFactory) error {
	out := clients.IO.WriteOut()

	ver := version.Get()
	var history []string
	for {
		line, err := readLineFunc(clients, history, ver)
		if err != nil {
			break
		}

		// Clear version after first prompt so banner only shows once
		ver = ""

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" {
			renderGoodbye(out)
			return nil
		}

		// Add to history (skip consecutive duplicates)
		if len(history) == 0 || history[len(history)-1] != line {
			history = append(history, line)
		}

		if line == "shell" {
			fmt.Fprintln(out, "Already in shell mode")
			continue
		}

		if line == "help" {
			line = "--help"
		}

		if BuildCommandTree == nil {
			return fmt.Errorf("shell command tree builder is not initialized")
		}

		args := strings.Fields(line)
		root := BuildCommandTree(ctx, clients)
		root.SetOut(out)
		root.SetErr(clients.IO.WriteErr())
		root.SetIn(clients.IO.ReadIn())
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			fmt.Fprintf(clients.IO.WriteErr(), "%s\n", err.Error())
		}
	}

	renderGoodbye(out)
	return nil
}
