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

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())

	tmp, _ := os.MkdirTemp("", "")
	_ = os.Chdir(tmp)
	defer os.RemoveAll(tmp)

	// Get command
	cmd, _ := Init(ctx)

	// Create mocks
	clientsMock := shared.NewClientsMock()
	testutil.MockCmdIO(clientsMock.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}
	output := clientsMock.GetCombinedOutput()

	for _, topLevelCommand := range cmd.Commands() {
		// In our template, we don't print out the parent command's description if it has subcommands
		// Hence, in the case where there are subcommands, we should always rather be checking
		// for the child command's description in the printout.
		if topLevelCommand.HasSubCommands() {
			for _, subCommand := range topLevelCommand.Commands() {
				// We should also ensure that we are not showing a subcommand if the parent is supposed to be hidden.
				if subCommand.Hidden || subCommand.Parent().Hidden {
					// Since this subcommand is to be hidden, we should not be relying on the Name() of the command not being present.
					// Reason: one command name could be a substring of another.  A command's `Short` value is a more reliable value to check.
					assert.NotContains(t, output, subCommand.Short, fmt.Sprintf("should contain %s in help output", subCommand.Short))
				} else {
					assert.Contains(t, output, subCommand.Short, fmt.Sprintf("should contain %s in help output", subCommand.Short))
					assert.Contains(t, output, subCommand.Name(), fmt.Sprintf("should contain %s in help output", subCommand.Name()))
				}
			}
		} else {
			// Since the command does not have child commands, we should check for the top-level command's description instead
			if topLevelCommand.Hidden {
				assert.NotContains(t, output, topLevelCommand.Short, fmt.Sprintf("should contain %s in help output", topLevelCommand.Short))
			} else {
				assert.Contains(t, output, topLevelCommand.Name(), fmt.Sprintf("should contain %s in help output", topLevelCommand.Name()))
				assert.Contains(t, output, topLevelCommand.Short, fmt.Sprintf("should contain %s in help output", topLevelCommand.Short))
			}
		}
	}
}

func TestExecuteContext(t *testing.T) {
	tests := map[string]struct {
		expectedErr      error
		expectedExitCode iostreams.ExitCode
		expectedOutputs  []string
	}{
		"Command successfully executes": {
			expectedErr:      nil,
			expectedExitCode: iostreams.ExitOK,
		},
		"Command fails execution and returns an error": {
			expectedErr:      fmt.Errorf("command failed"),
			expectedExitCode: iostreams.ExitError,
			expectedOutputs: []string{
				"command failed",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// Mock clients
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Mock command
			cmd := &cobra.Command{
				Use: "mock [flags]",
				RunE: func(cmd *cobra.Command, args []string) error {
					return tt.expectedErr
				},
			}

			// Execute the command
			ExecuteContext(ctx, cmd, clients)
			output := clientsMock.GetCombinedOutput()

			// Assertions
			// TODO: Assert that the event tracker was called with the correct exit code
			require.Equal(t, tt.expectedExitCode, clients.IO.GetExitCode())

			for _, expectedOutput := range tt.expectedOutputs {
				require.Contains(t, output, expectedOutput)
			}
		})
	}
}

// FYI: do not try to run this test in vscode using the run/debug test inline test helper; as the assertions in this test will fail in that context
func TestVersionFlags(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())

	tmp, _ := os.MkdirTemp("", "")
	_ = os.Chdir(tmp)
	defer os.RemoveAll(tmp)

	var output string

	// Get command
	cmd, _ := Init(ctx)

	// Create mocks
	clientsMock := shared.NewClientsMock()
	testutil.MockCmdIO(clientsMock.IO, cmd)

	// Test --version
	cmd.SetArgs([]string{"--version"})
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error", err.Error())
	}
	output = clientsMock.GetCombinedOutput()
	assert.True(t, testutil.ContainsSemVer(output), `--version should output the version number but yielded "%s"`, output)

	// Test -v
	cmd.SetArgs([]string{"-v"})
	err2 := cmd.ExecuteContext(ctx)
	if err2 != nil {
		assert.Fail(t, "cmd.Execute had unexpected error", err.Error())
	}
	output = clientsMock.GetCombinedOutput()
	assert.True(t, testutil.ContainsSemVer(output), `-v should output the version number but yielded "%s"`, output)
}

func Test_NewSuggestion(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())

	tmp, _ := os.MkdirTemp("", "")
	_ = os.Chdir(tmp)
	defer os.RemoveAll(tmp)

	// Get command
	cmd, clients := Init(ctx)

	// Create mocks
	clientsMock := shared.NewClientsMock()
	clients.IO = clientsMock.IO
	testutil.MockCmdIO(clientsMock.IO, cmd)

	// Execute new command
	cmd.SetArgs([]string{"new"})
	err := cmd.ExecuteContext(ctx)

	require.Error(t, err, "should have error because command not found")
	require.Regexp(t, `Did you mean this\?\s+create`, err.Error(), "should suggest the create command")
}

func Test_Aliases(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())

	tmp, _ := os.MkdirTemp("", "")
	_ = os.Chdir(tmp)
	defer os.RemoveAll(tmp)

	Init(ctx)

	tests := map[string]struct {
		args     string
		expected string
	}{
		"List alias": {
			args:     "list --help",
			expected: "auth list",
		},
		"Login alias": {
			args:     "login --help",
			expected: "auth login",
		},
		"Logout alias": {
			args:     "logout --help",
			expected: "auth logout",
		},
		"Activity alias": {
			args:     "activity --help",
			expected: "platform activity",
		},
		"Deploy alias": {
			args:     "deploy --help",
			expected: "platform deploy",
		},
		"Run alias": {
			args:     "run --help",
			expected: "platform run",
		},
		"Install alias": {
			args:     "install --help",
			expected: "app install",
		},
		"Uninstall alias": {
			args:     "uninstall --help",
			expected: "app uninstall",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err, output := testExecCmd(ctx, strings.Fields(tt.args))
			require.NoError(t, err)
			require.Contains(t, output, tt.expected)
		})
	}
}

// testExecCmd will execute the root cobra command with args and return the output
func testExecCmd(ctx context.Context, args []string) (error, string) {
	// Get command
	cmd, clients := Init(ctx)

	// Create mocks
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients.IO = clientsMock.IO
	testutil.MockCmdIO(clientsMock.IO, cmd)

	cmd.SetArgs(args)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		return err, ""
	}
	return nil, clientsMock.GetCombinedOutput()
}
