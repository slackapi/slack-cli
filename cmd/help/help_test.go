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

package help

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestHelpFunc(t *testing.T) {
	tests := map[string]struct {
		exampleCommands []style.ExampleCommand
		experiments     []string
		expectedOutput  []string
	}{
		"basic command information is included": {
			expectedOutput: []string{
				"USAGE",
				"$ root [flags]",
				"ALIASES",
				"root, su",
				"SUBCOMMANDS",
				"demo        A short command",
				"FLAGS",
				"--help   mock help flag",
				"EXPERIMENTS",
				"None",
			},
		},
		"examples are included in output": {
			exampleCommands: []style.ExampleCommand{
				{Command: "root", Meaning: "Eat more dirt"},
			},
			expectedOutput: []string{
				"EXAMPLE",
				"$ help.test root  # Eat more dirt",
			},
		},
		"experiments are included in output": {
			experiments: []string{"placeholder", "unknown"},
			expectedOutput: []string{
				"EXPERIMENTS",
				"placeholder",
				"unknown (invalid)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Remove any enabled experiments during the test and restore afterward
			var _EnabledExperiments = experiment.EnabledExperiments
			experiment.EnabledExperiments = []experiment.Experiment{}
			defer func() {
				// Restore original EnabledExperiments
				experiment.EnabledExperiments = _EnabledExperiments
			}()

			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clientsMock.Config.ExperimentsFlag = tt.experiments
			rootCmd := &cobra.Command{
				Use:     "root",
				Aliases: []string{"su"},
				Run:     func(cmd *cobra.Command, args []string) {},
				Example: style.ExampleCommandsf(tt.exampleCommands),
			}
			subCommand := &cobra.Command{
				Use:   "demo",
				Short: "A short command",
				Run:   func(cmd *cobra.Command, args []string) {},
			}
			rootCmd.AddCommand(subCommand)
			rootCmd.SetContext(context.Background())
			rootCmd.Flags().Bool("help", true, "mock help flag")
			clientsMock.Config.SetFlags(rootCmd)
			testutil.MockCmdIO(clientsMock.IO, rootCmd)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			helpFunc := HelpFunc(clients, map[string]string{})
			helpFunc(rootCmd, []string{})
			for _, expectedString := range tt.expectedOutput {
				assert.Contains(t, clientsMock.GetStdoutOutput(), expectedString)
			}
		})
	}
}
