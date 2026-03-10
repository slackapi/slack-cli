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

	huh "charm.land/huh/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

// styles is a helper that resolves the ThemeSlack for testing (isDark=false).
func styles() *huh.Styles {
	return ThemeSlack().Theme(false)
}

func TestThemeSlack(t *testing.T) {
	t.Run("returns a non-nil theme", func(t *testing.T) {
		s := styles()
		assert.NotNil(t, s)
	})

	t.Run("focused title is bold", func(t *testing.T) {
		s := styles()
		assert.True(t, s.Focused.Title.GetBold())
	})

	t.Run("focused title uses aubergine foreground", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#7C2852"), s.Focused.Title.GetForeground())
	})

	t.Run("focused select selector renders cursor", func(t *testing.T) {
		s := styles()
		rendered := s.Focused.SelectSelector.Render()
		assert.Contains(t, rendered, "❱")
	})

	t.Run("focused multi-select selected prefix renders checkmark", func(t *testing.T) {
		s := styles()
		rendered := s.Focused.SelectedPrefix.Render()
		assert.Contains(t, rendered, "✓")
	})

	t.Run("focused multi-select unselected prefix renders brackets", func(t *testing.T) {
		s := styles()
		rendered := s.Focused.UnselectedPrefix.Render()
		assert.Contains(t, rendered, "[ ]")
	})

	t.Run("focused error message uses red foreground", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#e01e5a"), s.Focused.ErrorMessage.GetForeground())
	})

	t.Run("focused button uses aubergine background", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#7C2852"), s.Focused.FocusedButton.GetBackground())
	})

	t.Run("focused button is bold", func(t *testing.T) {
		s := styles()
		assert.True(t, s.Focused.FocusedButton.GetBold())
	})

	t.Run("blurred select selector is blank", func(t *testing.T) {
		s := styles()
		rendered := s.Blurred.SelectSelector.Render()
		assert.Contains(t, rendered, "  ")
		assert.NotContains(t, rendered, "❱")
	})

	t.Run("blurred multi-select selector is blank", func(t *testing.T) {
		s := styles()
		rendered := s.Blurred.MultiSelectSelector.Render()
		assert.Contains(t, rendered, "  ")
		assert.NotContains(t, rendered, "❱")
	})

	t.Run("blurred border is hidden", func(t *testing.T) {
		s := styles()
		borderStyle := s.Blurred.Base.GetBorderStyle()
		assert.Equal(t, lipgloss.HiddenBorder(), borderStyle)
	})

	t.Run("focused border uses aubergine", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#7C2852"), s.Focused.Base.GetBorderLeftForeground())
	})

	t.Run("focused text input prompt uses blue", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#36c5f0"), s.Focused.TextInput.Prompt.GetForeground())
	})

	t.Run("focused text input cursor uses yellow", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#ecb22e"), s.Focused.TextInput.Cursor.GetForeground())
	})

	t.Run("focused selected option uses green", func(t *testing.T) {
		s := styles()
		assert.Equal(t, lipgloss.Color("#2eb67d"), s.Focused.SelectedOption.GetForeground())
	})
}
