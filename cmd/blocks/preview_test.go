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

package blocks

import (
	"bytes"
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/useragent"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// stubAIAgent stubs the detected AI coding tool and returns a function that
// restores the original detection.
func stubAIAgent(agent *useragent.AIAgent) func() {
	original := aiAgentFunc
	aiAgentFunc = func() *useragent.AIAgent { return agent }
	return func() { aiAgentFunc = original }
}

// stubTeamAuth stubs the team selection to return the provided auth
func stubTeamAuth(auth *types.SlackAuth) func() {
	original := promptTeamSlackAuthFunc
	promptTeamSlackAuthFunc = func(ctx context.Context, clients *shared.ClientFactory, promptText string, promptConfig *prompts.PromptTeamSlackAuthConfig) (*types.SlackAuth, error) {
		return auth, nil
	}
	return func() { promptTeamSlackAuthFunc = original }
}

func Test_Blocks_PreviewCommand(t *testing.T) {
	var restore func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"opens the builder with blocks from the --blocks flag": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.BlocksPreviewSuccess, []string{expectedURL})
			},
			Teardown: func() { restore() },
		},
		"opens the builder with blocks from stdin via the - sentinel": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			Teardown: func() { restore() },
		},
		"accepts a blocks object payload": {
			CmdArgs: []string{"--blocks", `{"blocks":[{"type":"divider"}]}`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			Teardown: func() { restore() },
		},
		"errors when no blocks are provided": {
			ExpectedErrorStrings: []string{slackerror.ErrMissingInput, "No blocks were provided"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"errors when the --blocks flag is empty": {
			CmdArgs:              []string{"--blocks", ""},
			ExpectedErrorStrings: []string{slackerror.ErrMissingInput, "No blocks were provided"},
		},
		"errors when reading from stdin on an interactive terminal": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsStdinTTY").Return(true)
			},
			ExpectedErrorStrings: []string{slackerror.ErrMissingInput, "standard input"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"errors when no teams are logged in": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)
			},
			ExpectedErrorStrings: []string{slackerror.ErrCredentialsNotFound},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"errors when the blocks are not valid json": {
			CmdArgs:              []string{"--blocks", `not json`},
			ExpectedErrorStrings: []string{slackerror.ErrUnableToParseJSON},
		},
		"errors when the json is not a blocks payload": {
			CmdArgs:              []string{"--blocks", `{"foo":"bar"}`},
			ExpectedErrorStrings: []string{slackerror.ErrInvalidBlocks},
		},
		"errors when reading blocks from stdin with multiple teams and no --team flag": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{TeamID: "T123", TeamDomain: "team-a"},
					{TeamID: "T456", TeamDomain: "team-b"},
				}, nil)
			},
			ExpectedErrorStrings: []string{slackerror.ErrMissingFlag, "--team"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"opens the builder when reading blocks from stdin with the --team flag set": {
			CmdArgs: []string{"--blocks", "-", "--team", "T123"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/T123/builder")
				}))
			},
			Teardown: func() { restore() },
		},
		"uses the enterprise id for enterprise installs": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: true})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/E456/builder")
				}))
			},
			Teardown: func() { restore() },
		},
		"uses the team id for org-grid workspace installs": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: false})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/T123/builder")
				}))
			},
			Teardown: func() { restore() },
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewPreviewCommand(cf)
	})
}

func Test_buildBlockKitBuilderURL(t *testing.T) {
	tests := map[string]struct {
		apiHost     string
		id          string
		blocksJSON  string
		expected    string
		expectedErr string
	}{
		"production host": {
			apiHost:    "https://slack.com",
			id:         "T123",
			blocksJSON: `{"blocks":[]}`,
			expected:   "https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"developer host": {
			apiHost:    "https://dev1234.slack.com",
			id:         "E456",
			blocksJSON: `{"blocks":[]}`,
			expected:   "https://app.dev1234.slack.com/block-kit-builder/E456/builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"empty host": {
			apiHost:     "",
			id:          "T123",
			blocksJSON:  `{"blocks":[]}`,
			expectedErr: slackerror.ErrInvalidArguments,
		},
		"scheme-less host": {
			apiHost:     "app.slack.com",
			id:          "T123",
			blocksJSON:  `{"blocks":[]}`,
			expectedErr: slackerror.ErrInvalidArguments,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := buildBlockKitBuilderURL(tc.apiHost, tc.id, tc.blocksJSON)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, slackerror.ToSlackError(err).Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func Test_normalizeBlocksPayload(t *testing.T) {
	tests := map[string]struct {
		input       string
		expected    string
		expectedErr string
	}{
		"wraps a bare array": {
			input:    `[{"type":"divider"}]`,
			expected: `{"blocks":[{"type":"divider"}]}`,
		},
		"passes through a blocks object": {
			input:    `{"blocks":[{"type":"divider"}]}`,
			expected: `{"blocks":[{"type":"divider"}]}`,
		},
		"compacts whitespace": {
			input:    "[\n  {\n    \"type\": \"divider\"\n  }\n]",
			expected: `{"blocks":[{"type":"divider"}]}`,
		},
		"preserves key order when wrapping a bare array": {
			input:    `[{"type":"section","text":{"type":"mrkdwn","text":"hi"}}]`,
			expected: `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"hi"}}]}`,
		},
		"preserves key order in a blocks object": {
			input:    `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"hi"}}]}`,
			expected: `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"hi"}}]}`,
		},
		"rejects invalid json": {
			input:       `not json`,
			expectedErr: slackerror.ErrUnableToParseJSON,
		},
		"rejects an object without blocks": {
			input:       `{"foo":"bar"}`,
			expectedErr: slackerror.ErrInvalidBlocks,
		},
		"rejects a non array blocks value": {
			input:       `{"blocks":"nope"}`,
			expectedErr: slackerror.ErrInvalidBlocks,
		},
		"rejects a scalar value": {
			input:       `42`,
			expectedErr: slackerror.ErrInvalidBlocks,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := normalizeBlocksPayload(tc.input)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, slackerror.ToSlackError(err).Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
