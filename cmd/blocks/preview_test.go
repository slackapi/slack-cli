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

func Test_Blocks_PreviewCommand(t *testing.T) {
	// teamlessURL is the Block Kit Builder URL rendered when no team can be
	// resolved. The browser resolves the team from its own session.
	const teamlessURL = `https://app.slack.com/block-kit-builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
	// teamURL is the Block Kit Builder URL scoped to team T123.
	const teamURL = `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
	testutil.TableTestCommand(t, testutil.CommandTests{
		"opens the builder with blocks from the --blocks flag": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.BlocksPreviewSuccess, []string{teamURL})
			},
		},
		"opens the builder with blocks from stdin via the - sentinel": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamURL)
			},
		},
		"opens the builder with blocks from stdin when the --blocks flag is omitted": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamURL)
			},
		},
		"accepts a blocks object payload": {
			CmdArgs: []string{"--blocks", `{"blocks":[{"type":"divider"}]}`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamURL)
			},
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
		"opens the team-less builder when no teams are logged in": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamlessURL)
			},
		},
		"opens the team-less builder when the credentials cannot be read": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound))
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamlessURL)
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
		"opens the team-less builder with multiple teams and no --team flag": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{TeamID: "T123", TeamDomain: "team-a"},
					{TeamID: "T456", TeamDomain: "team-b"},
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamlessURL)
			},
		},
		"opens the builder when reading blocks from stdin with the --team flag set": {
			CmdArgs: []string{"--blocks", "-", "--team", "T123"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{TeamID: "T123", TeamDomain: "team-a"},
					{TeamID: "T456", TeamDomain: "team-b"},
				}, nil)
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", teamURL)
			},
		},
		"errors when the --team flag has no matching auth": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`, "--team", "T999"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{{TeamID: "T123"}}, nil)
			},
			ExpectedErrorStrings: []string{slackerror.ErrTeamNotFound},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"errors when the --team flag is set but the credentials cannot be read": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`, "--team", "T123"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound))
			},
			ExpectedErrorStrings: []string{slackerror.ErrCredentialsNotFound},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"uses the enterprise id for enterprise installs": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: true},
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/E456/builder")
				}))
			},
		},
		"uses the team id for org-grid workspace installs": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: false},
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/T123/builder")
				}))
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewPreviewCommand(cf)
	})
}

func Test_buildBlockKitBuilderURL(t *testing.T) {
	tests := map[string]struct {
		apiHost            string
		teamOrEnterpriseID string
		blocksJSON         string
		expected           string
		expectedErr        string
	}{
		"production host": {
			apiHost:            "https://slack.com",
			teamOrEnterpriseID: "T123",
			blocksJSON:         `{"blocks":[]}`,
			expected:           "https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"developer host": {
			apiHost:            "https://dev1234.slack.com",
			teamOrEnterpriseID: "E456",
			blocksJSON:         `{"blocks":[]}`,
			expected:           "https://app.dev1234.slack.com/block-kit-builder/E456/builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"team-less builder when the id is empty": {
			apiHost:            "https://slack.com",
			teamOrEnterpriseID: "",
			blocksJSON:         `{"blocks":[]}`,
			expected:           "https://app.slack.com/block-kit-builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"empty host": {
			apiHost:            "",
			teamOrEnterpriseID: "T123",
			blocksJSON:         `{"blocks":[]}`,
			expectedErr:        slackerror.ErrInvalidArguments,
		},
		"scheme-less host": {
			apiHost:            "app.slack.com",
			teamOrEnterpriseID: "T123",
			blocksJSON:         `{"blocks":[]}`,
			expectedErr:        slackerror.ErrInvalidArguments,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := buildBlockKitBuilderURL(tc.apiHost, tc.teamOrEnterpriseID, tc.blocksJSON)
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

func Test_teamOrEnterpriseID(t *testing.T) {
	tests := map[string]struct {
		auth     *types.SlackAuth
		expected string
	}{
		"returns an empty string when the auth is nil": {
			auth:     nil,
			expected: "",
		},
		"returns the team id for workspace installs": {
			auth:     &types.SlackAuth{TeamID: "T123", EnterpriseID: "E456"},
			expected: "T123",
		},
		"returns the enterprise id for enterprise installs": {
			auth:     &types.SlackAuth{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: true},
			expected: "E456",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, teamOrEnterpriseID(tc.auth))
		})
	}
}
