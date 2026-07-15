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
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_DisplayDiffs(t *testing.T) {
	tests := map[string]struct {
		diffs             *DiffResult
		expectedOutputs   []string
		unexpectedOutputs []string
	}{
		"no differences prints nothing": {
			diffs:             &DiffResult{},
			unexpectedOutputs: []string{"App Manifest", "difference"},
		},
		"modified field shows both values": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "display_information.description", Type: DiffModified, LocalValue: "Local desc", RemoteValue: "Remote desc"},
			}},
			expectedOutputs: []string{
				"1 difference(s)",
				"display_information.description",
				"Project:",
				"Local desc",
				"App settings:",
				"Remote desc",
			},
		},
		"local-only field shows only-in-project label": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "functions.greet.title", Type: DiffLocalOnly, LocalValue: "Greet"},
			}},
			expectedOutputs: []string{
				"only in project",
				"functions.greet.title",
				"Greet",
			},
		},
		"remote-only field shows only-in-app-settings label": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "oauth_config.scopes.bot", Type: DiffRemoteOnly, RemoteValue: []any{"chat:write"}},
			}},
			expectedOutputs: []string{
				"only in app settings",
				"oauth_config.scopes.bot",
			},
		},
		"multiple diffs are sorted alphabetically by path": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "z_last", Type: DiffModified, LocalValue: "a", RemoteValue: "b"},
				{Path: "a_first", Type: DiffModified, LocalValue: "c", RemoteValue: "d"},
			}},
			expectedOutputs: []string{"2 difference(s)"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()

			DisplayDiffs(ctx, clientsMock.IO, tc.diffs)

			output := clientsMock.GetStdoutOutput()
			for _, expected := range tc.expectedOutputs {
				assert.Contains(t, output, expected)
			}
			for _, unexpected := range tc.unexpectedOutputs {
				assert.NotContains(t, output, unexpected)
			}
		})
	}
}

func Test_PromptResolutionStrategy(t *testing.T) {
	tests := map[string]struct {
		selectIndex int
		selectErr   error
		expected    MergeStrategy
		expectErr   bool
	}{
		"index 0 returns MergeAllLocal": {
			selectIndex: 0,
			expected:    MergeAllLocal,
		},
		"index 1 returns MergeAllRemote": {
			selectIndex: 1,
			expected:    MergeAllRemote,
		},
		"index 2 returns MergePerField": {
			selectIndex: 2,
			expected:    MergePerField,
		},
		"error from prompt is propagated": {
			selectErr: fmt.Errorf("user cancelled"),
			expectErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(iostreams.SelectPromptResponse{Index: tc.selectIndex}, tc.selectErr)

			result, err := PromptResolutionStrategy(ctx, clientsMock.IO)

			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_PromptFieldResolutions(t *testing.T) {
	tests := map[string]struct {
		diffs      *DiffResult
		selections []int
		selectErr  error
		expected   []FieldResolution
		expectErr  bool
	}{
		"modified field resolved as local": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "display_information.name", Type: DiffModified, LocalValue: "Local", RemoteValue: "Remote"},
			}},
			selections: []int{0},
			expected: []FieldResolution{
				{Path: "display_information.name", Resolution: ResolveLocal},
			},
		},
		"modified field resolved as remote": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "display_information.name", Type: DiffModified, LocalValue: "Local", RemoteValue: "Remote"},
			}},
			selections: []int{1},
			expected: []FieldResolution{
				{Path: "display_information.name", Resolution: ResolveRemote},
			},
		},
		"local-only field keep means resolve local": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "functions.greet", Type: DiffLocalOnly, LocalValue: "val"},
			}},
			selections: []int{0},
			expected: []FieldResolution{
				{Path: "functions.greet", Resolution: ResolveLocal},
			},
		},
		"local-only field remove means resolve remote": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "functions.greet", Type: DiffLocalOnly, LocalValue: "val"},
			}},
			selections: []int{1},
			expected: []FieldResolution{
				{Path: "functions.greet", Resolution: ResolveRemote},
			},
		},
		"remote-only field remove means resolve local": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "oauth_config.token", Type: DiffRemoteOnly, RemoteValue: "tok"},
			}},
			selections: []int{0},
			expected: []FieldResolution{
				{Path: "oauth_config.token", Resolution: ResolveLocal},
			},
		},
		"remote-only field keep means resolve remote": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "oauth_config.token", Type: DiffRemoteOnly, RemoteValue: "tok"},
			}},
			selections: []int{1},
			expected: []FieldResolution{
				{Path: "oauth_config.token", Resolution: ResolveRemote},
			},
		},
		"prompt error propagates": {
			diffs: &DiffResult{Diffs: []FieldDiff{
				{Path: "a", Type: DiffModified, LocalValue: "x", RemoteValue: "y"},
			}},
			selectErr: fmt.Errorf("cancelled"),
			expectErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()

			if tc.selectErr != nil {
				clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{}, tc.selectErr)
			} else {
				for _, idx := range tc.selections {
					clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(iostreams.SelectPromptResponse{Index: idx}, nil).Once()
				}
			}

			result, err := PromptFieldResolutions(ctx, clientsMock.IO, tc.diffs)

			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_formatValue(t *testing.T) {
	tests := map[string]struct {
		input    any
		expected string
	}{
		"nil returns not present": {
			input:    nil,
			expected: "(not present)",
		},
		"string is quoted": {
			input:    "hello world",
			expected: `"hello world"`,
		},
		"integer is marshalled as JSON": {
			input:    float64(42),
			expected: "42",
		},
		"map is marshalled as JSON": {
			input:    map[string]any{"key": "value"},
			expected: `{"key":"value"}`,
		},
		"long value is truncated at 80 chars": {
			input: map[string]any{
				"a_very_long_key_name_that_exceeds": "a_value_that_when_combined_with_the_key_will_definitely_push_past_eighty_chars_total",
			},
			expected: "...",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := formatValue(tc.input)
			assert.Contains(t, result, tc.expected)
		})
	}

	t.Run("long value result is exactly 80 chars", func(t *testing.T) {
		longMap := map[string]any{
			"a_very_long_key_name_that_exceeds": "a_value_that_when_combined_with_the_key_will_definitely_push_past_eighty_chars_total",
		}
		result := formatValue(longMap)
		assert.Len(t, result, 80)
		assert.True(t, result[len(result)-3:] == "...")
	})
}
