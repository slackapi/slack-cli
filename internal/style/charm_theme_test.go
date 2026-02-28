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

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestThemeSlack(t *testing.T) {
	t.Run("returns a non-nil theme", func(t *testing.T) {
		theme := ThemeSlack()
		assert.NotNil(t, theme)
	})

	t.Run("focused title is bold", func(t *testing.T) {
		theme := ThemeSlack()
		assert.True(t, theme.Focused.Title.GetBold())
	})

	t.Run("focused title uses aubergine foreground", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#7C2852"), theme.Focused.Title.GetForeground())
	})

	t.Run("focused select selector renders cursor", func(t *testing.T) {
		theme := ThemeSlack()
		rendered := theme.Focused.SelectSelector.Render()
		assert.Contains(t, rendered, "❱")
	})

	t.Run("focused multi-select selected prefix renders checkmark", func(t *testing.T) {
		theme := ThemeSlack()
		rendered := theme.Focused.SelectedPrefix.Render()
		assert.Contains(t, rendered, "✓")
	})

	t.Run("focused multi-select unselected prefix renders brackets", func(t *testing.T) {
		theme := ThemeSlack()
		rendered := theme.Focused.UnselectedPrefix.Render()
		assert.Contains(t, rendered, "[ ]")
	})

	t.Run("focused error message uses red foreground", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#e01e5a"), theme.Focused.ErrorMessage.GetForeground())
	})

	t.Run("focused button uses aubergine background", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#7C2852"), theme.Focused.FocusedButton.GetBackground())
	})

	t.Run("focused button is bold", func(t *testing.T) {
		theme := ThemeSlack()
		assert.True(t, theme.Focused.FocusedButton.GetBold())
	})

	t.Run("blurred select selector is blank", func(t *testing.T) {
		theme := ThemeSlack()
		rendered := theme.Blurred.SelectSelector.Render()
		assert.Contains(t, rendered, "  ")
		assert.NotContains(t, rendered, "❱")
	})

	t.Run("blurred multi-select selector is blank", func(t *testing.T) {
		theme := ThemeSlack()
		rendered := theme.Blurred.MultiSelectSelector.Render()
		assert.Contains(t, rendered, "  ")
		assert.NotContains(t, rendered, "❱")
	})

	t.Run("blurred border is hidden", func(t *testing.T) {
		theme := ThemeSlack()
		borderStyle := theme.Blurred.Base.GetBorderStyle()
		assert.Equal(t, lipgloss.HiddenBorder(), borderStyle)
	})

	t.Run("focused border uses aubergine", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#7C2852"), theme.Focused.Base.GetBorderLeftForeground())
	})

	t.Run("focused text input prompt uses blue", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#36c5f0"), theme.Focused.TextInput.Prompt.GetForeground())
	})

	t.Run("focused text input cursor uses yellow", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#ecb22e"), theme.Focused.TextInput.Cursor.GetForeground())
	})

	t.Run("focused selected option uses green", func(t *testing.T) {
		theme := ThemeSlack()
		assert.Equal(t, lipgloss.Color("#2eb67d"), theme.Focused.SelectedOption.GetForeground())
	})
}
