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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EmitClaudeCodePluginHint(t *testing.T) {
	tests := map[string]struct {
		claudeCode string
		expected   string
	}{
		"emits the hint on its own line inside Claude Code": {
			claudeCode: "1",
			expected:   claudeCodePluginHint + "\n",
		},
		"emits nothing when CLAUDECODE is unset": {
			claudeCode: "",
			expected:   "",
		},
		"emits nothing when CLAUDECODE is set to another value": {
			claudeCode: "true",
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
