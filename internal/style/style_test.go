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

package style

import (
	"testing"

	lipgloss "charm.land/lipgloss/v2"
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
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := RemoveANSI(tc.input)
			assert.Equal(t, tc.expected, actual)
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

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if s := Pluralize(tc.singular, tc.plural, tc.count); s != tc.expectedResult {
				t.Errorf("expected: %s, actual: %s", tc.expectedResult, s)
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

func TestToggleCharm(t *testing.T) {
	tests := map[string]struct {
		initial  bool
		toggle   bool
		expected bool
	}{
		"enables charm styling": {
			initial:  false,
			toggle:   true,
			expected: true,
		},
		"disables charm styling": {
			initial:  true,
			toggle:   false,
			expected: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			isCharmEnabled = tc.initial
			defer func() { isCharmEnabled = false }()
			ToggleCharm(tc.toggle)
			assert.Equal(t, tc.expected, isCharmEnabled)
		})
	}
}

// testStyleFunc verifies a style function returns the original text (stripped of ANSI)
// and behaves correctly across all three modes: colors off, legacy aurora, and charm lipgloss.
func testStyleFunc(t *testing.T, name string, fn func(string) string) {
	t.Helper()
	defer func() {
		ToggleStyles(false)
		ToggleCharm(false)
	}()

	input := "hello"

	t.Run(name+" returns plain text when colors are off", func(t *testing.T) {
		ToggleStyles(false)
		ToggleCharm(false)
		result := fn(input)
		assert.Equal(t, input, RemoveANSI(result))
	})

	t.Run(name+" returns styled text with legacy aurora", func(t *testing.T) {
		ToggleStyles(true)
		ToggleCharm(false)
		result := fn(input)
		assert.Contains(t, RemoveANSI(result), input)
	})

	t.Run(name+" returns styled text with charm lipgloss", func(t *testing.T) {
		ToggleStyles(true)
		ToggleCharm(true)
		result := fn(input)
		assert.Contains(t, RemoveANSI(result), input)
	})
}

func TestColorStyleFunctions(t *testing.T) {
	testStyleFunc(t, "Secondary", Secondary)
	testStyleFunc(t, "CommandText", CommandText)
	testStyleFunc(t, "LinkText", LinkText)
	testStyleFunc(t, "Selector", Selector)
	testStyleFunc(t, "Error", Error)
	testStyleFunc(t, "Warning", Warning)
	testStyleFunc(t, "Input", Input)
	testStyleFunc(t, "Green", Green)
	testStyleFunc(t, "Red", Red)
	testStyleFunc(t, "Yellow", Yellow)
	testStyleFunc(t, "Gray", Gray)
}

func TestTextStyleFunctions(t *testing.T) {
	testStyleFunc(t, "Bright", Bright)
	testStyleFunc(t, "Bold", Bold)
	testStyleFunc(t, "Darken", Darken)
	testStyleFunc(t, "Highlight", Highlight)
	testStyleFunc(t, "Underline", Underline)
}

func TestHeader(t *testing.T) {
	defer func() {
		ToggleStyles(false)
		ToggleCharm(false)
	}()

	t.Run("uppercases text", func(t *testing.T) {
		ToggleStyles(true)
		ToggleCharm(true)
		result := Header("commands")
		assert.Contains(t, RemoveANSI(result), "COMMANDS")
	})

	t.Run("uppercases text with legacy", func(t *testing.T) {
		ToggleStyles(true)
		ToggleCharm(false)
		result := Header("commands")
		assert.Contains(t, RemoveANSI(result), "COMMANDS")
	})
}

func TestFaint(t *testing.T) {
	defer func() {
		ToggleStyles(false)
		ToggleCharm(false)
	}()

	t.Run("returns plain text when colors are off", func(t *testing.T) {
		ToggleStyles(false)
		result := Faint("hello")
		assert.Equal(t, "hello", result)
	})

	t.Run("returns styled text with legacy", func(t *testing.T) {
		ToggleStyles(true)
		ToggleCharm(false)
		result := Faint("hello")
		assert.Contains(t, result, "hello")
		assert.NotEqual(t, "hello", result)
	})

	t.Run("returns styled text with charm", func(t *testing.T) {
		ToggleStyles(true)
		ToggleCharm(true)
		result := Faint("hello")
		assert.Contains(t, RemoveANSI(result), "hello")
	})
}

func TestRender(t *testing.T) {
	defer func() {
		ToggleStyles(false)
	}()

	t.Run("returns plain text when colors are off", func(t *testing.T) {
		ToggleStyles(false)
		result := render(lipgloss.NewStyle().Bold(true), "test")
		assert.Equal(t, "test", result)
	})

	t.Run("returns styled text when colors are on", func(t *testing.T) {
		ToggleStyles(true)
		result := render(lipgloss.NewStyle().Bold(true), "test")
		assert.Contains(t, RemoveANSI(result), "test")
	})
}

func TestStyler(t *testing.T) {
	t.Run("returns an aurora instance", func(t *testing.T) {
		s := Styler()
		assert.NotNil(t, s)
	})
}

func TestEmoji(t *testing.T) {
	defer func() {
		ToggleStyles(false)
	}()

	t.Run("returns empty when colors are off", func(t *testing.T) {
		ToggleStyles(false)
		assert.Equal(t, "", Emoji("gear"))
	})

	t.Run("returns empty for whitespace alias", func(t *testing.T) {
		ToggleStyles(true)
		assert.Equal(t, "", Emoji("  "))
	})

	t.Run("returns emoji with padding for known aliases", func(t *testing.T) {
		ToggleStyles(true)
		result := Emoji("gear")
		assert.NotEmpty(t, result)
	})
}
