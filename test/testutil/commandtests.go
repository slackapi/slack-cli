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

package testutil

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CommandTests describes a single test case for a command.
type CommandTests map[string]struct {
	Setup                 func(*testing.T, context.Context, *shared.ClientsMock, *shared.ClientFactory) // Optional
	Teardown              func()                                                                        // Optional
	CmdArgs               []string                                                                      // Required, Example: ["my-app", "--template", "slack-samples/deno-starter-template", "--verbose"]
	ExpectedOutputs       []string                                                                      // Optional
	ExpectedStdoutOutputs []string                                                                      // Optional
	ExpectedAsserts       func(*testing.T, *shared.ClientsMock)                                         // Optional
	ExpectedError         error                                                                         // Optional
	ExpectedErrorStrings  []string                                                                      // Optional
}

// TableTestCommand will run a table test collection defined by commandTests for a command created by newCommandFunc
func TableTestCommand(t *testing.T, commandTests CommandTests, newCommandFunc func(*shared.ClientFactory) *cobra.Command) {
	for name, tt := range commandTests {
		t.Run(name, func(t *testing.T) {
			// Create mocks
			ctxMock := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()

			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Create the command
			cmd := newCommandFunc(clients)
			clientsMock.Config.InitializeGlobalFlags(cmd)
			MockCmdIO(clients.IO, cmd)

			// Setup custom mocks (higher priority than default mocks)
			if tt.Setup != nil {
				tt.Setup(t, ctxMock, clientsMock, clients)
			}

			// Setup default mock actions
			clientsMock.AddDefaultMocks()

			// Execute the command
			cmd.SetArgs(tt.CmdArgs)
			err := cmd.ExecuteContext(ctxMock)

			// Test failure mode
			if tt.ExpectedError != nil || tt.ExpectedErrorStrings != nil {
				if err == nil {
					var expectedErr string
					if tt.ExpectedError != nil {
						// ExpectedError set on test
						expectedErr = tt.ExpectedError.Error()
					} else {
						// ExpectedErrorStrings set on test
						expectedErr = strings.Join(tt.ExpectedErrorStrings, ", ")
					}
					assert.Fail(t, `cmd.Execute had expected error(s) "%s", but none was returned`, expectedErr)
				} else {
					if tt.ExpectedError != nil {
						// ExpectedError set on test
						assert.Contains(t, err.Error(), tt.ExpectedError.Error())
					} else {
						// ExpectedErrorStrings set on test
						for _, expectedErr := range tt.ExpectedErrorStrings {
							requireOutputContains(t, "error", err.Error(), expectedErr)
						}
					}
				}
			} else {
				require.NoError(t, err)
			}
			// Assert command output
			combinedOutput := clientsMock.GetCombinedOutput()
			for _, expectedOutput := range tt.ExpectedOutputs {
				requireOutputContains(t, "combined", combinedOutput, expectedOutput)
			}
			stdoutOutput := clientsMock.GetStdoutOutput()
			for _, expectedOutput := range tt.ExpectedStdoutOutputs {
				requireOutputContains(t, "stdout", stdoutOutput, expectedOutput)
			}

			// Assert mocks or other custom assertions
			if tt.ExpectedAsserts != nil {
				tt.ExpectedAsserts(t, clientsMock)
			}

			if tt.Teardown != nil {
				tt.Teardown()
			}
		})
	}
}

// requireOutputContains verifies that output contains a string and fails if not
func requireOutputContains(t *testing.T, fd string, output string, contains string) {
	message := strings.Join([]string{
		fmt.Sprintf("The %s output is missing a string!", fd),
		"Expected:",
		fmt.Sprintf("  %s", strings.ReplaceAll(contains, "\n", "\n  ")),
		"Actual:",
		fmt.Sprintf("  %s", strings.ReplaceAll(output, "\n", "\n  ")),
	}, "\n")
	if assert.Contains(t, output, contains, message) {
		return
	}
	t.FailNow()
}
