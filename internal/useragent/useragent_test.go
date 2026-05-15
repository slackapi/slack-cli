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

package useragent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func clearEnvVars(t *testing.T) {
	t.Helper()
	for _, env := range []string{"CLAUDECODE", "CLAUDE_CODE_ENTRYPOINT", "CODEX_CI", "GEMINI_CLI", "CLINE_ACTIVE", "CURSOR_AGENT", "AGENT"} {
		t.Setenv(env, "")
	}
}

func Test_UserAgent_BuildUserAgent(t *testing.T) {
	tests := map[string]struct {
		envVars  map[string]string
		contains string
		noAgent  bool
	}{
		"CLAUDECODE takes priority over AGENT": {
			envVars:  map[string]string{"CLAUDECODE": "1", "AGENT": "goose", "CLAUDE_CODE_ENTRYPOINT": "cli"},
			contains: "AI-Agent (name=claude-code, entry=cli)",
		},
		"includes AI-Agent suffix for AGENT env var": {
			envVars:  map[string]string{"AGENT": "goose"},
			contains: "AI-Agent (name=goose)",
		},
		"includes AI-Agent suffix for Claude Code with entrypoint": {
			envVars:  map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_ENTRYPOINT": "cli"},
			contains: "AI-Agent (name=claude-code, entry=cli)",
		},
		"includes AI-Agent suffix for Claude Code with vscode entrypoint": {
			envVars:  map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_ENTRYPOINT": "vscode"},
			contains: "AI-Agent (name=claude-code, entry=vscode)",
		},
		"includes AI-Agent suffix for Claude Code without entrypoint": {
			envVars:  map[string]string{"CLAUDECODE": "1"},
			contains: "AI-Agent (name=claude-code)",
		},
		"includes AI-Agent suffix for Cline": {
			envVars:  map[string]string{"CLINE_ACTIVE": "true"},
			contains: "AI-Agent (name=cline)",
		},
		"includes AI-Agent suffix for Codex": {
			envVars:  map[string]string{"CODEX_CI": "1"},
			contains: "AI-Agent (name=codex)",
		},
		"includes AI-Agent suffix for Cursor": {
			envVars:  map[string]string{"CURSOR_AGENT": "1"},
			contains: "AI-Agent (name=cursor)",
		},
		"includes AI-Agent suffix for Gemini CLI": {
			envVars:  map[string]string{"GEMINI_CLI": "1"},
			contains: "AI-Agent (name=gemini-cli)",
		},
		"includes AI-Agent suffix for unknown agent": {
			envVars:  map[string]string{"AGENT": "future-agent"},
			contains: "AI-Agent (name=future-agent)",
		},
		"no AI-Agent suffix when no agent detected": {
			envVars: map[string]string{},
			noAgent: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clearEnvVars(t)
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			ua := BuildUserAgent("2.38.1")
			assert.Contains(t, ua, "slack-cli/2.38.1")
			if tc.contains != "" {
				assert.Contains(t, ua, tc.contains)
			}
			if tc.noAgent {
				assert.NotContains(t, ua, "AI-Agent")
			}
		})
	}
}

func Test_UserAgent_Detect(t *testing.T) {
	tests := map[string]struct {
		envVars  map[string]string
		expected *AIAgent
	}{
		"CLAUDECODE takes priority over AGENT": {
			envVars:  map[string]string{"CLAUDECODE": "1", "AGENT": "goose"},
			expected: &AIAgent{Name: "claude-code", Entry: ""},
		},
		"CODEX_CI takes priority over AGENT": {
			envVars:  map[string]string{"CODEX_CI": "1", "AGENT": "goose"},
			expected: &AIAgent{Name: "codex"},
		},
		"detects agent via AGENT env var": {
			envVars:  map[string]string{"AGENT": "goose"},
			expected: &AIAgent{Name: "goose"},
		},
		"detects Claude Code": {
			envVars:  map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_ENTRYPOINT": "cli"},
			expected: &AIAgent{Name: "claude-code", Entry: "cli"},
		},
		"detects Claude Code with vscode entrypoint": {
			envVars:  map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_ENTRYPOINT": "vscode"},
			expected: &AIAgent{Name: "claude-code", Entry: "vscode"},
		},
		"detects Claude Code without entrypoint": {
			envVars:  map[string]string{"CLAUDECODE": "1"},
			expected: &AIAgent{Name: "claude-code", Entry: ""},
		},
		"detects Cline": {
			envVars:  map[string]string{"CLINE_ACTIVE": "true"},
			expected: &AIAgent{Name: "cline"},
		},
		"detects Codex": {
			envVars:  map[string]string{"CODEX_CI": "1"},
			expected: &AIAgent{Name: "codex"},
		},
		"detects Cursor": {
			envVars:  map[string]string{"CURSOR_AGENT": "1"},
			expected: &AIAgent{Name: "cursor"},
		},
		"detects Gemini CLI": {
			envVars:  map[string]string{"GEMINI_CLI": "1"},
			expected: &AIAgent{Name: "gemini-cli"},
		},
		"detects unknown agent via AGENT env var": {
			envVars:  map[string]string{"AGENT": "future-agent"},
			expected: &AIAgent{Name: "future-agent"},
		},
		"returns nil when no agent detected": {
			envVars:  map[string]string{},
			expected: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clearEnvVars(t)
			for k, v := range tc.envVars {
				t.Setenv(k, v)
			}
			result := Detect()
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
