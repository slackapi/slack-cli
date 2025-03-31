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

package hooks

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var MOCK_BOUNDARY_STRING = "boundary-string"
var SIXTY_FOUR_KB_STRING = string(make([]byte, (64*1024)+1))
var FIVE_HUNDRED_TWELVE_KB_STRING = string(make([]byte, (512*1024)+1))

// mockBoundaryStringGenerator returns a random string for finding in tests
func mockBoundaryStringGenerator() string {
	return MOCK_BOUNDARY_STRING
}

func Test_Hook_Execute_V2_Protocol(t *testing.T) {
	tests := map[string]struct {
		opts  HookExecOpts
		check func(*testing.T, string, error, ExecInterface)
	}{
		"error if hook command unavailable": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "batman"},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.Equal(t, slackerror.New(slackerror.ErrSDKHookNotFound).WithMessage("The command for 'batman' was not found"), err)
			},
		},
		"successful execution": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Env: map[string]string{
					"batman": "robin",
					"yin":    "yang",
				},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte(MOCK_BOUNDARY_STRING + `{"message": "hello world"}` + MOCK_BOUNDARY_STRING),
						Err:        nil,
					},
				},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.Equal(t, `{"message": "hello world"}`, response)
				require.Equal(t, nil, err)
				require.Contains(t, mockExec.(*MockExec).mockCommand.Env, `batman="robin"`)
				require.Contains(t, mockExec.(*MockExec).mockCommand.Env, `yin="yang"`)
			},
		},
		"successful execution with payload > 64kb": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Env: map[string]string{
					"batman": "robin",
					"yin":    "yang",
				},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						StdoutIO: io.NopCloser(strings.NewReader(MOCK_BOUNDARY_STRING + SIXTY_FOUR_KB_STRING + MOCK_BOUNDARY_STRING)),
						Err:      nil,
					},
				},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.Equal(t, SIXTY_FOUR_KB_STRING, response)
				require.Equal(t, nil, err)
				require.Contains(t, mockExec.(*MockExec).mockCommand.Env, `batman="robin"`)
				require.Contains(t, mockExec.(*MockExec).mockCommand.Env, `yin="yang"`)
			},
		},
		"successful execution with payload > 512kb": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						StdoutIO: io.NopCloser(strings.NewReader("before" + MOCK_BOUNDARY_STRING + FIVE_HUNDRED_TWELVE_KB_STRING + MOCK_BOUNDARY_STRING + "after")),
						Err:      nil,
					},
				},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.NoError(t, err)
				require.Equal(t, FIVE_HUNDRED_TWELVE_KB_STRING, response)
			},
		},
		"failed command execution": {
			opts: HookExecOpts{
				Hook: HookScript{Command: "boom", Name: "sadpath"},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						Err: errors.New("explosion"),
					},
				},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.Equal(t, slackerror.New(slackerror.ErrSDKHookInvocationFailed).
					WithMessage("Error running 'sadpath' command: explosion"), err)
			},
		},
		"fail to parse payload due to improper boundary strings": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Env:  map[string]string{},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						StdoutIO: io.NopCloser(strings.NewReader("diagnostic info" + MOCK_BOUNDARY_STRING + MOCK_BOUNDARY_STRING + `{"message": "hello world"}` + MOCK_BOUNDARY_STRING)),
						StderrIO: io.NopCloser(strings.NewReader(``)),
						Err:      nil,
					},
				},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.Equal(t, "", response)
				require.Equal(t, nil, err)
			},
		},
		"fail to parse payload due to missing boundary strings": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Env:  map[string]string{},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte(`{"message": "hello world"}`),
						Err:        nil,
					},
				},
			},
			check: func(t *testing.T, response string, err error, mockExec ExecInterface) {
				require.Equal(t, "", response)
				require.Equal(t, nil, err)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			generateBoundary = mockBoundaryStringGenerator
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			config := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(config, fs, os)
			ios.AddDefaultMocks()
			hookExecutor := &HookExecutorMessageBoundaryProtocol{
				IO: ios,
			}
			response, err := hookExecutor.Execute(tt.opts)
			tt.check(t, response, err, tt.opts.Exec)
		})
	}
}

func Test_Hook_Execute_V2_GenerateMD5FromRandomString(t *testing.T) {
	randomString1 := generateMD5FromRandomString()
	randomString2 := generateMD5FromRandomString()

	assert.NotEqual(t, randomString1, randomString2)
	assert.GreaterOrEqual(t, len(randomString1), 10)
	assert.GreaterOrEqual(t, len(randomString2), 10)
}
