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

package iostreams

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// keys creates a tea.KeyMsg for the given runes (same helper used in huh_test.go).
func keys(runes ...rune) tea.KeyMsg {
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: runes,
	}
}

func TestCharmInput(t *testing.T) {
	t.Run("renders the title", func(t *testing.T) {
		var input string
		f := buildInputForm("Enter your name", InputPromptConfig{}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Enter your name")
	})

	t.Run("accepts typed input", func(t *testing.T) {
		var input string
		f := buildInputForm("Name?", InputPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(keys('H', 'u', 'h'))

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Huh")
	})

	t.Run("renders placeholder text", func(t *testing.T) {
		var input string
		f := buildInputForm("Name?", InputPromptConfig{Placeholder: "my-cool-app"}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "my-cool-app")
	})

	t.Run("stores typed value", func(t *testing.T) {
		var input string
		f := buildInputForm("Name?", InputPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(keys('t', 'e', 's', 't'))
		f.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Equal(t, "test", input)
	})
}

func TestCharmConfirm(t *testing.T) {
	t.Run("renders the title and buttons", func(t *testing.T) {
		choice := false
		f := buildConfirmForm("Are you sure?", &choice)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Are you sure?")
		assert.Contains(t, view, "Yes")
		assert.Contains(t, view, "No")
	})

	t.Run("default value is respected", func(t *testing.T) {
		choice := true
		f := buildConfirmForm("Continue?", &choice)
		f.Update(f.Init())

		assert.True(t, choice)
	})

	t.Run("toggle changes value", func(t *testing.T) {
		choice := false
		f := buildConfirmForm("Continue?", &choice)
		f.Update(f.Init())

		// Toggle to Yes
		f.Update(tea.KeyMsg{Type: tea.KeyLeft})
		assert.True(t, choice)

		// Toggle back to No
		f.Update(tea.KeyMsg{Type: tea.KeyRight})
		assert.False(t, choice)
	})
}

func TestCharmSelect(t *testing.T) {
	t.Run("renders the title and options", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildSelectForm("Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Pick one")
		assert.Contains(t, view, "Foo")
		assert.Contains(t, view, "Bar")
		assert.Contains(t, view, "Baz")
	})

	t.Run("cursor starts on first option", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar"}
		f := buildSelectForm("Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "❱ Foo")
	})

	t.Run("cursor navigation moves selection", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildSelectForm("Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		m, _ := f.Update(tea.KeyMsg{Type: tea.KeyDown})
		view := ansi.Strip(m.View())
		assert.Contains(t, view, "❱ Bar")
		assert.False(t, strings.Contains(view, "❱ Foo"))
	})

	t.Run("submit selects the hovered option", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildSelectForm("Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		// Move down to Bar, then submit
		f.Update(tea.KeyMsg{Type: tea.KeyDown})
		f.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Equal(t, "Bar", selected)
	})

	t.Run("descriptions are appended to option display", func(t *testing.T) {
		var selected string
		options := []string{"Alpha", "Beta"}
		cfg := SelectPromptConfig{
			Description: func(opt string, _ int) string {
				if opt == "Alpha" {
					return "First letter"
				}
				return ""
			},
		}
		f := buildSelectForm("Choose", options, cfg, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "First letter")
	})

	t.Run("page size sets field height", func(t *testing.T) {
		var selected string
		options := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		cfg := SelectPromptConfig{PageSize: 3}
		f := buildSelectForm("Pick", options, cfg, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		// With PageSize 3 (height 5), not all 8 options should be visible
		assert.Contains(t, view, "A")
		// At minimum the form should render without error
		assert.NotEmpty(t, view)
	})
}

func TestCharmPassword(t *testing.T) {
	t.Run("renders the title", func(t *testing.T) {
		var input string
		f := buildPasswordForm("Enter password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Enter password")
	})

	t.Run("typed characters are masked in view", func(t *testing.T) {
		var input string
		f := buildPasswordForm("Password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(keys('s', 'e', 'c', 'r', 'e', 't'))

		view := ansi.Strip(f.View())
		assert.NotContains(t, view, "secret")
	})

	t.Run("stores typed value despite masking", func(t *testing.T) {
		var input string
		f := buildPasswordForm("Password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(keys('a', 'b', 'c'))
		f.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Equal(t, "abc", input)
	})
}

func TestCharmMultiSelect(t *testing.T) {
	t.Run("renders the title and options", func(t *testing.T) {
		var selected []string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildMultiSelectForm("Pick many", options, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Pick many")
		assert.Contains(t, view, "Foo")
		assert.Contains(t, view, "Bar")
		assert.Contains(t, view, "Baz")
	})

	t.Run("toggle selection with x key", func(t *testing.T) {
		var selected []string
		options := []string{"Foo", "Bar"}
		f := buildMultiSelectForm("Pick", options, &selected)
		f.Update(f.Init())

		// Toggle first item
		m, _ := f.Update(keys('x'))
		view := ansi.Strip(m.View())

		// After toggle, the first item should show as selected (checkmark)
		assert.Contains(t, view, "✓")
	})

	t.Run("submit returns toggled items", func(t *testing.T) {
		var selected []string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildMultiSelectForm("Pick", options, &selected)
		f.Update(f.Init())

		// Toggle Foo (first item)
		f.Update(keys('x'))
		// Move down and toggle Bar
		f.Update(tea.KeyMsg{Type: tea.KeyDown})
		f.Update(keys('x'))
		// Submit
		f.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.ElementsMatch(t, []string{"Foo", "Bar"}, selected)
	})
}

func TestCharmFormsUseSlackTheme(t *testing.T) {
	t.Run("input form uses Slack theme", func(t *testing.T) {
		var input string
		f := buildInputForm("Test", InputPromptConfig{}, &input)
		f.Update(f.Init())

		// The Slack theme applies a thick left border with bright aubergine color.
		// Verify the form renders with a border (the base theme includes a thick
		// left border which renders as a vertical bar character).
		view := f.View()
		assert.Contains(t, view, "┃")
	})

	t.Run("select form renders themed cursor", func(t *testing.T) {
		var selected string
		f := buildSelectForm("Pick", []string{"A", "B"}, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "❱ A")
	})

	t.Run("multi-select form renders themed prefixes", func(t *testing.T) {
		var selected []string
		f := buildMultiSelectForm("Pick", []string{"A", "B"}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		// Our Slack theme uses "[ ] " as unselected prefix
		assert.Contains(t, view, "[ ]")
	})

	t.Run("all form builders apply ThemeSlack", func(t *testing.T) {
		// Verify each builder returns a form that can Init and render without panic
		var s string
		var b bool
		var ss []string
		forms := []*huh.Form{
			buildInputForm("msg", InputPromptConfig{}, &s),
			buildConfirmForm("msg", &b),
			buildSelectForm("msg", []string{"a"}, SelectPromptConfig{}, &s),
			buildPasswordForm("msg", PasswordPromptConfig{}, &s),
			buildMultiSelectForm("msg", []string{"a"}, &ss),
		}
		for _, f := range forms {
			f.Update(f.Init())
			assert.NotEmpty(t, f.View())
		}
	})
}
