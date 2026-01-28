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
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Hook_Execute_Default_Protocol(t *testing.T) {
	tests := map[string]struct {
		opts             HookExecOpts
		handler          func(t *testing.T, ctx context.Context, executor HookExecutor, opts HookExecOpts)
		expectedError    error
		expectedResponse string
	}{
		"error if unavailable": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "batman"},
			},
			expectedError:    slackerror.New(slackerror.ErrSDKHookNotFound).WithMessage("The command for 'batman' was not found"),
			expectedResponse: "",
		},
		"successful execution (no env vars)": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte("test output"),
						Err:        nil,
					},
				},
			},
			expectedError:    nil,
			expectedResponse: "test output",
		},
		"successful execution (some env vars)": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "happypath", Command: "echo {}"},
				Env: map[string]string{
					"batman": "robin",
					"yin":    "yang",
				},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte("test output"),
						Err:        nil,
					},
				},
			},
			handler: func(t *testing.T, ctx context.Context, executor HookExecutor, opts HookExecOpts) {
				response, err := executor.Execute(ctx, opts)
				require.Equal(t, "test output", response)
				require.Equal(t, nil, err)
				require.Contains(t, opts.Exec.(*MockExec).mockCommand.Env, `batman="robin"`)
				require.Contains(t, opts.Exec.(*MockExec).mockCommand.Env, `yin="yang"`)
			},
		},
		"failed execution": {
			opts: HookExecOpts{
				Hook: HookScript{Command: "boom", Name: "sadpath"},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte("kapow"),
						MockStderr: []byte("there was a problem compiling your app"),
						Err:        errors.New("explosion"),
					},
				},
			},
			expectedError: slackerror.New(slackerror.ErrSDKHookInvocationFailed).
				WithMessage("Error running 'sadpath' command: explosion").
				WithDetails(slackerror.ErrorDetails{
					slackerror.ErrorDetail{
						Message: "there was a problem compiling your app",
					},
				}),
			expectedResponse: "",
		},
		"successful deploy command": {
			opts: HookExecOpts{
				Hook:   HookScript{Command: "echo lgtm!", Name: "Deploy"},
				Stdout: &bytes.Buffer{},
			},
			expectedError:    nil,
			expectedResponse: "lgtm!",
		},
		"failed deploy script": {
			opts: HookExecOpts{
				Hook: HookScript{Command: "./deployer.sh", Name: "Deploy"},
			},
			expectedError: slackerror.New(slackerror.ErrSDKHookInvocationFailed).
				WithMessage("Failed to successfully complete the 'Deploy' hook"),
			handler: func(t *testing.T, ctx context.Context, executor HookExecutor, opts HookExecOpts) {
				_, err := executor.Execute(ctx, opts)
				require.Error(t, err)
			},
		},
		"successful start command": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "Start", Command: "deno run some/path/to/start"},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte("start output line 1\nstart output line 2"),
						Err:        nil,
					},
				},
			},
			expectedError:    nil,
			expectedResponse: "start output line 2",
		},
		"successful start command should trim trailing new line and space": {
			opts: HookExecOpts{
				Hook: HookScript{Name: "Start", Command: "deno run some/path/to/start"},
				Exec: &MockExec{
					mockCommand: &MockCommand{
						MockStdout: []byte("start output line 1\nstart output line 2\n   \n\n\n\n"),
						Err:        nil,
					},
				},
			},
			expectedError:    nil,
			expectedResponse: "start output line 2",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			config := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(config, fs, os)
			ios.AddDefaultMocks()
			hookExecutor := &HookExecutorDefaultProtocol{
				IO: ios,
			}
			if tc.handler != nil {
				tc.handler(t, ctx, hookExecutor, tc.opts)
			} else {
				str, err := hookExecutor.Execute(ctx, tc.opts)
				assert.Contains(t, str, tc.expectedResponse)
				assert.Equal(t, tc.expectedError, err)
			}
		})
	}
}
