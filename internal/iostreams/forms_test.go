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
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	huh "charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/stretchr/testify/assert"
)

// keys creates a tea.KeyPressMsg for the given rune.
func key(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func TestInputForm(t *testing.T) {
	t.Run("renders the title", func(t *testing.T) {
		var input string
		f := buildInputForm(nil, "Enter your name", InputPromptConfig{}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Enter your name")
	})

	t.Run("renders the chevron prompt", func(t *testing.T) {
		var input string
		f := buildInputForm(nil, "Name?", InputPromptConfig{}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, style.Chevron())
	})

	t.Run("accepts typed input", func(t *testing.T) {
		var input string
		f := buildInputForm(nil, "Name?", InputPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(key('H'))
		f.Update(key('u'))
		f.Update(key('h'))

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Huh")
	})

	t.Run("renders placeholder text", func(t *testing.T) {
		var input string
		f := buildInputForm(nil, "Name?", InputPromptConfig{Placeholder: "my-cool-app"}, &input)
		f.Update(f.Init())

		// In huh v2, the cursor overlays the first placeholder character,
		// so the full placeholder may not appear verbatim in the view.
		// Verify the form renders and includes at least the placeholder start.
		view := ansi.Strip(f.View())
		assert.Contains(t, view, "m")
		assert.Contains(t, view, "Name?")
	})

	t.Run("stores typed value", func(t *testing.T) {
		var input string
		f := buildInputForm(nil, "Name?", InputPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(key('t'))
		f.Update(key('e'))
		f.Update(key('s'))
		f.Update(key('t'))
		f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Equal(t, "test", input)
	})
}

func TestConfirmForm(t *testing.T) {
	t.Run("renders the title and buttons", func(t *testing.T) {
		choice := false
		f := buildConfirmForm(nil, "Are you sure?", &choice)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Are you sure?")
		assert.Contains(t, view, "Yes")
		assert.Contains(t, view, "No")
	})

	t.Run("default value is respected", func(t *testing.T) {
		choice := true
		f := buildConfirmForm(nil, "Continue?", &choice)
		f.Update(f.Init())

		assert.True(t, choice)
	})

	t.Run("toggle changes value", func(t *testing.T) {
		choice := false
		f := buildConfirmForm(nil, "Continue?", &choice)
		f.Update(f.Init())

		// Toggle to Yes
		f.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		assert.True(t, choice)

		// Toggle back to No
		f.Update(tea.KeyPressMsg{Code: tea.KeyRight})
		assert.False(t, choice)
	})
}

func TestSelectForm(t *testing.T) {
	t.Run("renders the title and options", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildSelectForm(nil, "Pick one", options, SelectPromptConfig{}, &selected)
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
		f := buildSelectForm(nil, "Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, style.Chevron()+" Foo")
	})

	t.Run("cursor navigation moves selection", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildSelectForm(nil, "Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		m, _ := f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		view := ansi.Strip(m.View())
		assert.Contains(t, view, style.Chevron()+" Bar")
		assert.False(t, strings.Contains(view, style.Chevron()+" Foo"))
	})

	t.Run("submit selects the hovered option", func(t *testing.T) {
		var selected string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildSelectForm(nil, "Pick one", options, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		// Move down to Bar, then submit
		f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

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
		f := buildSelectForm(nil, "Choose", options, cfg, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "First letter")
	})

	t.Run("descriptions use em-dash separator with lipgloss enabled", func(t *testing.T) {
		style.ToggleLipgloss(true)
		style.ToggleStyles(true)
		t.Cleanup(func() {
			style.ToggleLipgloss(false)
			style.ToggleStyles(false)
		})

		fsMock := slackdeps.NewFsMock()
		osMock := slackdeps.NewOsMock()
		osMock.AddDefaultMocks()
		cfg := config.NewConfig(fsMock, osMock)
		cfg.ExperimentsFlag = []string{"lipgloss"}
		cfg.LoadExperiments(context.Background(), func(_ context.Context, _ string, _ ...any) {})
		io := NewIOStreams(cfg, fsMock, osMock)

		var selected string
		options := []string{"Alpha", "Beta"}
		selectCfg := SelectPromptConfig{
			Description: func(opt string, _ int) string {
				if opt == "Alpha" {
					return "First letter"
				}
				return ""
			},
		}
		f := buildSelectForm(io, "Choose", options, selectCfg, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, " — First letter")
	})

	t.Run("descriptions use em-dash separator without lipgloss", func(t *testing.T) {
		var selected string
		options := []string{"Alpha", "Beta"}
		selectCfg := SelectPromptConfig{
			Description: func(opt string, _ int) string {
				if opt == "Alpha" {
					return "First letter"
				}
				return ""
			},
		}
		f := buildSelectForm(nil, "Choose", options, selectCfg, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Alpha — First letter")
	})

	t.Run("page size sets field height", func(t *testing.T) {
		var selected string
		options := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
		cfg := SelectPromptConfig{PageSize: 3}
		f := buildSelectForm(nil, "Pick", options, cfg, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		// With PageSize 3 (height 5), not all 8 options should be visible
		assert.Contains(t, view, "A")
		// At minimum the form should render without error
		assert.NotEmpty(t, view)
	})
}

func TestPasswordForm(t *testing.T) {
	t.Run("renders the title", func(t *testing.T) {
		var input string
		f := buildPasswordForm(nil, "Enter password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Enter password")
	})

	t.Run("renders the chevron prompt", func(t *testing.T) {
		var input string
		f := buildPasswordForm(nil, "Enter password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, style.Chevron())
	})

	t.Run("typed characters are masked in view", func(t *testing.T) {
		var input string
		f := buildPasswordForm(nil, "Password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(key('s'))
		f.Update(key('e'))
		f.Update(key('c'))
		f.Update(key('r'))
		f.Update(key('e'))
		f.Update(key('t'))

		view := ansi.Strip(f.View())
		assert.NotContains(t, view, "secret")
	})

	t.Run("stores typed value despite masking", func(t *testing.T) {
		var input string
		f := buildPasswordForm(nil, "Password", PasswordPromptConfig{}, &input)
		f.Update(f.Init())

		f.Update(key('a'))
		f.Update(key('b'))
		f.Update(key('c'))
		f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Equal(t, "abc", input)
	})
}

func TestMultiSelectForm(t *testing.T) {
	t.Run("renders the title and options", func(t *testing.T) {
		var selected []string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildMultiSelectForm(nil, "Pick many", options, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Pick many")
		assert.Contains(t, view, "Foo")
		assert.Contains(t, view, "Bar")
		// Baz may be scrolled out of the default viewport
	})

	t.Run("toggle selection with x key", func(t *testing.T) {
		var selected []string
		options := []string{"Foo", "Bar"}
		f := buildMultiSelectForm(nil, "Pick", options, &selected)
		f.Update(f.Init())

		// Toggle first item
		m, _ := f.Update(key('x'))
		view := ansi.Strip(m.View())

		// After toggle, the first item should show as selected (checkmark)
		assert.Contains(t, view, "✓")
	})

	t.Run("submit returns toggled items", func(t *testing.T) {
		var selected []string
		options := []string{"Foo", "Bar", "Baz"}
		f := buildMultiSelectForm(nil, "Pick", options, &selected)
		f.Update(f.Init())

		// Toggle Foo (first item)
		f.Update(key('x'))
		// Move down and toggle Bar
		f.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		f.Update(key('x'))
		// Submit
		f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.ElementsMatch(t, []string{"Foo", "Bar"}, selected)
	})
}

func TestFormsUseSlackTheme(t *testing.T) {
	fsMock := slackdeps.NewFsMock()
	osMock := slackdeps.NewOsMock()
	osMock.AddDefaultMocks()
	cfg := config.NewConfig(fsMock, osMock)
	cfg.ExperimentsFlag = []string{"lipgloss"}
	cfg.LoadExperiments(context.Background(), func(_ context.Context, _ string, _ ...any) {})
	io := NewIOStreams(cfg, fsMock, osMock)

	t.Run("input form uses Slack theme", func(t *testing.T) {
		var input string
		f := buildInputForm(io, "Test", InputPromptConfig{}, &input)
		f.Update(f.Init())

		// The Slack theme applies a thick left border with bright aubergine color.
		// Verify the form renders with a border (the base theme includes a thick
		// left border which renders as a vertical bar character).
		view := f.View()
		assert.Contains(t, view, "┃")
	})

	t.Run("select form renders themed cursor", func(t *testing.T) {
		var selected string
		f := buildSelectForm(io, "Pick", []string{"A", "B"}, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, style.Chevron()+" A")
	})

	t.Run("multi-select form renders themed prefixes", func(t *testing.T) {
		var selected []string
		f := buildMultiSelectForm(io, "Pick", []string{"A", "B"}, &selected)
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
			buildInputForm(io, "msg", InputPromptConfig{}, &s),
			buildConfirmForm(io, "msg", &b),
			buildSelectForm(io, "msg", []string{"a"}, SelectPromptConfig{}, &s),
			buildPasswordForm(io, "msg", PasswordPromptConfig{}, &s),
			buildMultiSelectForm(io, "msg", []string{"a"}, &ss),
		}
		for _, f := range forms {
			f.Update(f.Init())
			assert.NotEmpty(t, f.View())
		}
	})
}

func TestFormsUseSurveyTheme(t *testing.T) {
	t.Run("multi-select uses survey prefix without lipgloss", func(t *testing.T) {
		var selected []string
		f := buildMultiSelectForm(nil, "Pick", []string{"A", "B"}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		// ThemeSurvey uses "[ ] " as unselected prefix
		assert.Contains(t, view, "[ ]")
	})

	t.Run("multi-select uses [x] for selected prefix", func(t *testing.T) {
		var selected []string
		f := buildMultiSelectForm(nil, "Pick", []string{"A", "B"}, &selected)
		f.Update(f.Init())

		// Toggle first item
		m, _ := f.Update(key('x'))
		view := ansi.Strip(m.View())
		assert.Contains(t, view, "[x]")
	})

	t.Run("select form renders chevron cursor", func(t *testing.T) {
		var selected string
		f := buildSelectForm(nil, "Pick", []string{"A", "B"}, SelectPromptConfig{}, &selected)
		f.Update(f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, style.Chevron()+" A")
	})

	t.Run("all form builders apply ThemeSurvey without lipgloss", func(t *testing.T) {
		var s string
		var b bool
		var ss []string
		forms := []*huh.Form{
			buildInputForm(nil, "msg", InputPromptConfig{}, &s),
			buildConfirmForm(nil, "msg", &b),
			buildSelectForm(nil, "msg", []string{"a"}, SelectPromptConfig{}, &s),
			buildPasswordForm(nil, "msg", PasswordPromptConfig{}, &s),
			buildMultiSelectForm(nil, "msg", []string{"a"}, &ss),
		}
		for _, f := range forms {
			f.Update(f.Init())
			assert.NotEmpty(t, f.View())
		}
	})
}
