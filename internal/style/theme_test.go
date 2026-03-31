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
	"runtime"
	"testing"

	"github.com/AlecAivazis/survey/v2/core"
	"github.com/stretchr/testify/assert"
)

func TestThemeSlack(t *testing.T) {
	theme := ThemeSlack().Theme(false)
	tests := map[string]struct {
		rendered   string
		expected   []string
		unexpected []string
	}{
		"focused title renders text": {
			rendered: theme.Focused.Title.Render("x"),
			expected: []string{"x"},
		},
		"focused error message renders text": {
			rendered: theme.Focused.ErrorMessage.Render("err"),
			expected: []string{"err"},
		},
		"focused select selector renders chevron": {
			rendered: theme.Focused.SelectSelector.Render(),
			expected: []string{Chevron()},
		},
		"focused multi-select selected prefix has checkmark": {
			rendered: theme.Focused.SelectedPrefix.Render(),
			expected: []string{"✓"},
		},
		"focused multi-select unselected prefix has brackets": {
			rendered: theme.Focused.UnselectedPrefix.Render(),
			expected: []string{"[ ]"},
		},
		"focused button renders text": {
			rendered: theme.Focused.FocusedButton.Render("OK"),
			expected: []string{"OK"},
		},
		"blurred select selector is blank": {
			rendered:   theme.Blurred.SelectSelector.Render(),
			expected:   []string{"  "},
			unexpected: []string{Chevron()},
		},
		"blurred multi-select selector is blank": {
			rendered:   theme.Blurred.MultiSelectSelector.Render(),
			expected:   []string{"  "},
			unexpected: []string{Chevron()},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, exp := range tc.expected {
				assert.Contains(t, tc.rendered, exp)
			}
			for _, unexp := range tc.unexpected {
				assert.NotContains(t, tc.rendered, unexp)
			}
		})
	}
}

func TestThemeSurvey(t *testing.T) {
	theme := ThemeSurvey().Theme(false)
	tests := map[string]struct {
		rendered   string
		expected   []string
		unexpected []string
	}{
		"focused title renders text": {
			rendered: theme.Focused.Title.Render("x"),
			expected: []string{"x"},
		},
		"focused error message renders text": {
			rendered: theme.Focused.ErrorMessage.Render("err"),
			expected: []string{"err"},
		},
		"focused select selector renders chevron": {
			rendered: theme.Focused.SelectSelector.Render(),
			expected: []string{Chevron()},
		},
		"focused multi-select selected prefix has [x]": {
			rendered: theme.Focused.SelectedPrefix.Render(),
			expected: []string{"[x]"},
		},
		"focused multi-select unselected prefix has brackets": {
			rendered: theme.Focused.UnselectedPrefix.Render(),
			expected: []string{"[ ]"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, exp := range tc.expected {
				assert.Contains(t, tc.rendered, exp)
			}
			for _, unexp := range tc.unexpected {
				assert.NotContains(t, tc.rendered, unexp)
			}
		})
	}
}

func TestThemePlain(t *testing.T) {
	theme := ThemePlain().Theme(false)
	tests := map[string]struct {
		rendered string
		expected string
	}{
		"title renders plain text": {
			rendered: theme.Focused.Title.Render("x"),
			expected: "x",
		},
		"error message renders plain text": {
			rendered: theme.Focused.ErrorMessage.Render("err"),
			expected: "err",
		},
		"select selector renders plain >": {
			rendered: theme.Focused.SelectSelector.Render(),
			expected: "> ",
		},
		"selected prefix renders [x]": {
			rendered: theme.Focused.SelectedPrefix.Render(),
			expected: "[x] ",
		},
		"unselected prefix renders [ ]": {
			rendered: theme.Focused.UnselectedPrefix.Render(),
			expected: "[ ] ",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.rendered)
		})
	}
}

func TestChevron(t *testing.T) {
	tests := map[string]struct {
		styleEnabled bool
		expected     string
	}{
		"styles disabled returns plain chevron": {
			styleEnabled: false,
			expected:     ">",
		},
		"styles enabled returns fancy chevron": {
			styleEnabled: true,
			expected: func() string {
				if runtime.GOOS == "windows" {
					return ">"
				}
				return "❱"
			}(),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			prev := isStyleEnabled
			isStyleEnabled = tc.styleEnabled
			defer func() {
				isStyleEnabled = prev
			}()
			assert.Equal(t, tc.expected, Chevron())
		})
	}
}

func TestSurveyIcons(t *testing.T) {
	tests := map[string]struct {
		styleEnabled bool
	}{
		"styles are not enabled": {
			styleEnabled: false,
		},
		"styles are enabled": {
			styleEnabled: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			core.DisableColor = false
			isStyleEnabled = tc.styleEnabled

			_ = SurveyIcons()
			assert.NotEqual(t, tc.styleEnabled, core.DisableColor)
		})
	}
}
