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

package python

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_containsPackage(t *testing.T) {
	tests := map[string]struct {
		fileData string
		expected bool
	}{
		"Matches the normalized name": {
			fileData: "slack-cli-hooks<1.0.0",
			expected: true,
		},
		"Matches dots and mixed case": {
			fileData: "Slack.CLI-Hooks",
			expected: true,
		},
		"Matches underscores with extras and markers": {
			fileData: `slack_cli_hooks[x]>=0.3; python_version>"3.9"`,
			expected: true,
		},
		"Skips a longer package name": {
			fileData: "slack-cli-hooks-example==1.0.0",
			expected: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, containsPackage(tc.fileData, "slack-cli-hooks"))
		})
	}
}
