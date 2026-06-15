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

package manifest

import (
	"testing"
	"unicode/utf8"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
)

func Test_DisplayDiffs(t *testing.T) {
	tests := map[string]struct {
		diffs            []FieldDiff
		expectedSubstrs  []string
		forbiddenSubstrs []string
	}{
		"modified field shows both values side-by-side": {
			diffs: []FieldDiff{
				{
					Path:        "display_information.name",
					Type:        DiffModified,
					LocalValue:  "Project Name",
					RemoteValue: "Remote Name",
				},
			},
			expectedSubstrs: []string{
				"display_information.name",
				`Project:      "Project Name"`,
				`App settings: "Remote Name"`,
			},
			forbiddenSubstrs: []string{"(only in", "Value:", "(not set)"},
		},
		"local-only field shows (not set) on the app settings side": {
			diffs: []FieldDiff{
				{
					Path:       "functions.greet.title",
					Type:       DiffLocalOnly,
					LocalValue: "Greet",
				},
			},
			expectedSubstrs: []string{
				"functions.greet.title",
				`Project:      "Greet"`,
				"App settings: (not set)",
			},
			forbiddenSubstrs: []string{"(only in", "Value:"},
		},
		"remote-only field shows (not set) on the project side": {
			diffs: []FieldDiff{
				{
					Path:        "settings.is_mcp_enabled",
					Type:        DiffRemoteOnly,
					RemoteValue: false,
				},
			},
			expectedSubstrs: []string{
				"settings.is_mcp_enabled",
				"Project:      (not set)",
				"App settings: false",
			},
			forbiddenSubstrs: []string{"(only in", "Value:"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()

			DisplayDiffs(ctx, cm.IO, &DiffResult{Diffs: tc.diffs})

			out := cm.GetStdoutOutput()
			for _, want := range tc.expectedSubstrs {
				assert.Contains(t, out, want, "expected output to contain %q", want)
			}
			for _, forbidden := range tc.forbiddenSubstrs {
				assert.NotContains(t, out, forbidden, "expected output not to contain %q", forbidden)
			}
		})
	}
}

func Test_truncateRunes(t *testing.T) {
	tests := map[string]struct {
		input    string
		max      int
		expected string
	}{
		"shorter than max returns unchanged": {
			input:    "hello",
			max:      80,
			expected: "hello",
		},
		"exactly max runes returns unchanged": {
			input:    "abcdefghij",
			max:      10,
			expected: "abcdefghij",
		},
		"longer than max truncates with ellipsis": {
			input:    "abcdefghijklmno",
			max:      10,
			expected: "abcdefg...",
		},
		"multi-byte runes are not cut mid-character": {
			// Each emoji is 4 bytes in UTF-8 but one rune. Byte-based
			// slicing would split the middle emoji.
			input:    "🐶🐱🐭🐹🐰🦊🐻🐼🐨🐯🦁🐮",
			max:      6,
			expected: "🐶🐱🐭...",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := truncateRunes(tc.input, tc.max)
			assert.Equal(t, tc.expected, got)
			// In every case the result must remain valid UTF-8.
			assert.True(t, utf8.ValidString(got), "result is not valid UTF-8: %q", got)
		})
	}
}
