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

package hooks

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_HookExecOpts_ShellEnv(t *testing.T) {
	tests := map[string]struct {
		env    map[string]string
		osenv  map[string]string
		setup  func(afero.Fs)
		assert func(t *testing.T, result []string)
	}{
		"includes opts.Env variables": {
			env: map[string]string{"SLACK_CLI_XAPP": "xapp-token"},
			assert: func(t *testing.T, result []string) {
				assert.Contains(t, result, "SLACK_CLI_XAPP=xapp-token")
			},
		},
		"includes dotenv variables": {
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, ".env", []byte("PROJECT_VAR=hello\n"), 0600)
			},
			assert: func(t *testing.T, result []string) {
				assert.Contains(t, result, "PROJECT_VAR=hello")
			},
		},
		"includes os environment variables": {
			osenv: map[string]string{"SHELL_ENV_TEST_VAR": "from-os"},
			assert: func(t *testing.T, result []string) {
				assert.Contains(t, result, "SHELL_ENV_TEST_VAR=from-os")
			},
		},
		"dotenv has higher precedence than opts.Env": {
			env: map[string]string{"DUPLICATE": "from-opts"},
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, ".env", []byte("DUPLICATE=from-dotenv\n"), 0600)
			},
			assert: func(t *testing.T, result []string) {
				assert.Equal(t, "DUPLICATE=from-opts", result[0])
				assert.Equal(t, "DUPLICATE=from-dotenv", result[1])
			},
		},
		"works without dotenv file": {
			env: map[string]string{"HOOK_VAR": "value"},
			assert: func(t *testing.T, result []string) {
				assert.Contains(t, result, "HOOK_VAR=value")
			},
		},
		"continues when dotenv file has invalid syntax": {
			setup: func(fs afero.Fs) {
				_ = afero.WriteFile(fs, ".env", []byte(`KEY="unclosed`), 0600)
			},
			assert: func(t *testing.T, result []string) {
				// Should still include os.Environ() even if .env parsing fails
				assert.NotEmpty(t, result)
			},
		},
		"works with nil env map": {
			assert: func(t *testing.T, result []string) {
				assert.NotEmpty(t, result)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			fs := slackdeps.NewFsMock()
			cfg := config.NewConfig(fs, os)
			io := iostreams.NewIOStreamsMock(cfg, fs, os)
			io.AddDefaultMocks()
			ctx := context.Background()

			for k, v := range tc.osenv {
				t.Setenv(k, v)
			}
			if tc.setup != nil {
				tc.setup(fs)
			}

			opts := HookExecOpts{Env: tc.env}
			result := opts.ShellEnv(ctx, fs, io)
			tc.assert(t, result)
		})
	}
}
