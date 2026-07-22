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

	internalmanifest "github.com/slackapi/slack-cli/internal/manifest"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
)

func Test_displayDiffs(t *testing.T) {
	tests := map[string]struct {
		diffs            []internalmanifest.FieldDiff
		expectedSubstrs  []string
		forbiddenSubstrs []string
	}{
		"modified field shows both values side-by-side": {
			diffs: []internalmanifest.FieldDiff{
				{
					Path:        "display_information.name",
					Type:        internalmanifest.DiffModified,
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
			diffs: []internalmanifest.FieldDiff{
				{
					Path:       "functions.greet.title",
					Type:       internalmanifest.DiffLocalOnly,
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
			diffs: []internalmanifest.FieldDiff{
				{
					Path:        "settings.is_mcp_enabled",
					Type:        internalmanifest.DiffRemoteOnly,
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
		"multiple diffs are sorted by path and rendered in order": {
			diffs: []internalmanifest.FieldDiff{
				{Path: "features.bot_user.display_name", Type: internalmanifest.DiffModified, LocalValue: "App", RemoteValue: "App (local)"},
				{Path: "display_information.name", Type: internalmanifest.DiffModified, LocalValue: "App", RemoteValue: "App (local)"},
			},
			expectedSubstrs: []string{
				"Found 2 differences between project and app settings",
				"display_information.name",
				"features.bot_user.display_name",
			},
		},
		"empty result prints nothing": {
			diffs: nil,
			forbiddenSubstrs: []string{
				"Found",
				"Manifest Diff",
				"Project:",
				"App settings:",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()

			displayDiffs(ctx, cm.IO, &internalmanifest.DiffResult{Diffs: tc.diffs})

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

func Test_formatValue(t *testing.T) {
	tests := map[string]struct {
		input    any
		expected string
	}{
		"nil renders as (not present)": {
			input:    nil,
			expected: "(not present)",
		},
		"strings are quoted": {
			input:    "hello",
			expected: `"hello"`,
		},
		"booleans are JSON-encoded": {
			input:    false,
			expected: "false",
		},
		"long strings are rune-truncated": {
			input:    "This is a very long description that exceeds the eighty character limit for displayed values in the manifest diff",
			expected: `"This is a very long description that exceeds the eighty character limit for ...`,
		},
		"long non-string values are rune-truncated": {
			input: []string{
				"alpha", "bravo", "charlie", "delta", "echo", "foxtrot",
				"golf", "hotel", "india", "juliet", "kilo",
			},
			expected: `["alpha","bravo","charlie","delta","echo","foxtrot","golf","hotel","india","j...`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, formatValue(tc.input))
		})
	}
}

func Test_displayPath(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"plain path passes through unchanged": {
			input:    "display_information.name",
			expected: "display_information.name",
		},
		"escaped dots are unescaped for display": {
			input:    `functions.slack\.users\.lookup.title`,
			expected: "functions.slack.users.lookup.title",
		},
		"escaped backslashes are preserved": {
			input:    `path.key\\with\\backslashes.field`,
			expected: `path.key\with\backslashes.field`,
		},
		"escaped backslash before dot is handled correctly": {
			input:    `path.segment\\.next`,
			expected: `path.segment\.next`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, displayPath(tc.input))
		})
	}
}
