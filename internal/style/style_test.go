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

package style

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveANSI(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"normal strings are left unmodified": {
			input:    "sometimes strings are plain",
			expected: "sometimes strings are plain",
		},
		"stylized strings have ansi removed": {
			input:    "sometimes \x1b[1;30mstrings are\x1b[0m bold",
			expected: "sometimes strings are bold",
		},
		"sequences without space are removed": {
			input:    "executable file not found in $PATH\x1b[1;38;5;178m\x1b[0m",
			expected: "executable file not found in $PATH",
		},
		"characters before ansi are included": {
			input:    "exit status 1\x1b[1;38;5;178m\x1b[0m",
			expected: "exit status 1",
		},
		"uppercase escapes are removed": {
			input:    "exit status \x1B[1mnone\x1B[0m",
			expected: "exit status none",
		},
		"unexpected spacing is still escaped": {
			input:    "script was not found\x1b[1;38;5;178m (sdk_hook_not_found)\x1b[0m",
			expected: "script was not found (sdk_hook_not_found)",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := RemoveANSI(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestToggleStyles(t *testing.T) {
	defer func() {
		ToggleStyles(false)
	}()
	t.Run("false sets to false", func(t *testing.T) {
		isStyleEnabled = true
		isColorShown = true
		isLinkShown = true
		ToggleStyles(false)
		assert.False(t, isStyleEnabled)
		assert.False(t, isColorShown)
		assert.False(t, isLinkShown)
	})
	t.Run("true sets to true", func(t *testing.T) {
		isStyleEnabled = false
		isColorShown = false
		isLinkShown = false
		ToggleStyles(true)
		assert.True(t, isStyleEnabled)
		assert.True(t, isColorShown)
		assert.True(t, isLinkShown)
	})
}

func TestPluralize(t *testing.T) {
	tests := map[string]struct {
		singular       string
		plural         string
		count          int
		expectedResult string
	}{
		"0 should be plural": {
			singular:       "cat",
			plural:         "cats",
			count:          0,
			expectedResult: "cats",
		},
		"1 should be singular": {
			singular:       "workspace",
			plural:         "workspaces",
			count:          1,
			expectedResult: "workspace",
		},
		"2 should be plural": {
			singular:       "shoe",
			plural:         "shoes",
			count:          2,
			expectedResult: "shoes",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if s := Pluralize(tt.singular, tt.plural, tt.count); s != tt.expectedResult {
				t.Errorf("expected: %s, actual: %s", tt.expectedResult, s)
			}
		})
	}
}

// Verify no text is output when no emoji is given
func TestEmojiEmpty(t *testing.T) {
	alias := ""
	emoji := Emoji(alias)
	if emoji != "" {
		t.Errorf("non-empty text returned, when none was expected")
	}
}
