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
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EmitClaudeCodePluginHint(t *testing.T) {
	tests := map[string]struct {
		claudeCode string
		expected   string
	}{
		"emits the hint on its own line inside claude code": {
			claudeCode: "1",
			expected:   claudeCodePluginHint + "\n",
		},
		"emits the hint for any non-empty CLAUDECODE value": {
			claudeCode: "true",
			expected:   claudeCodePluginHint + "\n",
		},
		"emits nothing when CLAUDECODE is unset": {
			claudeCode: "",
			expected:   "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clearEnvVars(t)
			t.Setenv("CLAUDECODE", tc.claudeCode)

			var buf bytes.Buffer
			EmitClaudeCodePluginHint(&buf)

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func Test_EmitAgentPluginHint(t *testing.T) {
	tests := map[string]struct {
		envVars     map[string]string
		processName string
		expected    string
	}{
		"emits the hint on its own line inside codex": {
			envVars:  map[string]string{"CODEX_CI": "1"},
			expected: fmt.Sprintf(agentPluginHint, "slack") + "\n",
		},
		"emits the hint on its own line inside cursor": {
			envVars:  map[string]string{"CURSOR_AGENT": "1"},
			expected: fmt.Sprintf(agentPluginHint, "slack") + "\n",
		},
		"emits the hint for a generic agent": {
			envVars:  map[string]string{"AGENT": "amp"},
			expected: fmt.Sprintf(agentPluginHint, "slack") + "\n",
		},
		"emits the hint on its own line inside claude code": {
			envVars:  map[string]string{"CLAUDECODE": "1"},
			expected: fmt.Sprintf(agentPluginHint, "slack") + "\n",
		},
		"emits the hint for any non-empty CLAUDECODE value": {
			envVars:  map[string]string{"CLAUDECODE": "true"},
			expected: fmt.Sprintf(agentPluginHint, "slack") + "\n",
		},
		"emits the hint with the process name of the cli": {
			envVars:     map[string]string{"CODEX_CI": "1"},
			processName: "slack-dev",
			expected:    fmt.Sprintf(agentPluginHint, "slack-dev") + "\n",
		},
		"emits nothing when no agent is detected": {
			envVars:  map[string]string{},
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clearEnvVars(t)
			for key, value := range tc.envVars {
				t.Setenv(key, value)
			}

			processName := tc.processName
			if processName == "" {
				processName = "slack"
			}
			var buf bytes.Buffer
			EmitAgentPluginHint(&buf, processName)

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}
