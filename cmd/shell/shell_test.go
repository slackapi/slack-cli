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
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// mockBuildCommandTree returns a command tree with a single "version" command
// that prints a known string, for testing command execution.
func mockBuildCommandTree(_ context.Context, clients *shared.ClientFactory) *cobra.Command {
	root := &cobra.Command{Use: "slack", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("v0.0.0-test")
			return nil
		},
	})
	return root
}

// mockReadLines returns a readLineFunc that yields lines sequentially, then io.EOF.
func mockReadLines(lines []string) func(*shared.ClientFactory, []string, string) (string, error) {
	i := 0
	return func(_ *shared.ClientFactory, _ []string, _ string) (string, error) {
		if i >= len(lines) {
			return "", io.EOF
		}
		line := lines[i]
		i++
		return line, nil
	}
}

func setupTest(isTTY bool, charmEnabled bool) (*shared.ClientsMock, *cobra.Command) {
	cm := shared.NewClientsMock()
	cm.AddDefaultMocks()
	cm.IO.On("IsTTY").Unset()
	cm.IO.On("IsTTY").Return(isTTY)

	clients := shared.NewClientFactory(cm.MockClientFactory())

	if charmEnabled {
		clients.Config.ExperimentsFlag = []string{string(experiment.Charm)}
		clients.Config.LoadExperiments(context.Background(), func(_ context.Context, _ string, _ ...interface{}) {})
	}

	cmd := NewCommand(clients)
	cmd.SetContext(context.Background())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	return cm, cmd
}

func TestShellCommand(t *testing.T) {
	// Save and restore BuildCommandTree and readLineFunc
	origBuild := BuildCommandTree
	origReadLine := readLineFunc
	defer func() {
		BuildCommandTree = origBuild
		readLineFunc = origReadLine
	}()
	BuildCommandTree = mockBuildCommandTree

	tests := map[string]struct {
		input       []string
		isTTY       bool
		charm       bool
		expectErr   bool
		errContains string
		outContains string
		outExcludes string
	}{
		"requires charm experiment": {
			input:       []string{"exit"},
			isTTY:       true,
			charm:       false,
			expectErr:   true,
			errContains: "charm experiment",
		},
		"requires TTY": {
			input:       []string{"exit"},
			isTTY:       false,
			charm:       true,
			expectErr:   true,
			errContains: "interactive terminal",
		},
		"exit command exits cleanly": {
			input:       []string{"exit"},
			isTTY:       true,
			charm:       true,
			expectErr:   false,
			outContains: "Goodbye",
		},
		"quit command exits cleanly": {
			input:       []string{"quit"},
			isTTY:       true,
			charm:       true,
			expectErr:   false,
			outContains: "Goodbye",
		},
		"shell recursion warning": {
			input:       []string{"shell", "exit"},
			isTTY:       true,
			charm:       true,
			expectErr:   false,
			outContains: "Already in shell mode",
		},
		"empty input continues": {
			input:     []string{"", "exit"},
			isTTY:     true,
			charm:     true,
			expectErr: false,
		},
		"command execution": {
			input:       []string{"version", "exit"},
			isTTY:       true,
			charm:       true,
			expectErr:   false,
			outContains: "v0.0.0-test",
		},
		"help maps to --help": {
			input:       []string{"help", "exit"},
			isTTY:       true,
			charm:       true,
			expectErr:   false,
			outContains: "slack",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			readLineFunc = mockReadLines(tc.input)
			cm, cmd := setupTest(tc.isTTY, tc.charm)
			stdout := &bytes.Buffer{}
			cmd.SetOut(stdout)

			err := cmd.Execute()

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			output := cm.GetCombinedOutput() + stdout.String()
			if tc.outContains != "" {
				assert.Contains(t, output, tc.outContains)
			}
			if tc.outExcludes != "" {
				assert.NotContains(t, output, tc.outExcludes)
			}
		})
	}
}
