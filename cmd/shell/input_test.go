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

package shell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestInputModel(t *testing.T) {
	tests := map[string]struct {
		setup    func() inputModel
		actions  func(inputModel) inputModel
		assertFn func(t *testing.T, m inputModel)
	}{
		"enter submits text": {
			setup: func() inputModel {
				return newInputModel(nil, "")
			},
			actions: func(m inputModel) inputModel {
				for _, r := range "deploy" {
					updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
					m = updated.(inputModel)
				}
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
				m = updated.(inputModel)
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.True(t, m.done)
				assert.Equal(t, "deploy", m.value)
			},
		},
		"ctrl+c returns exit": {
			setup: func() inputModel {
				return newInputModel(nil, "")
			},
			actions: func(m inputModel) inputModel {
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
				m = updated.(inputModel)
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.True(t, m.done)
				assert.Equal(t, "exit", m.value)
			},
		},
		"up arrow recalls history": {
			setup: func() inputModel {
				return newInputModel([]string{"deploy", "run"}, "")
			},
			actions: func(m inputModel) inputModel {
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
				m = updated.(inputModel)
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.Equal(t, "run", m.textInput.Value())
				assert.Equal(t, 1, m.histIndex)
			},
		},
		"up arrow twice recalls older history": {
			setup: func() inputModel {
				return newInputModel([]string{"deploy", "run"}, "")
			},
			actions: func(m inputModel) inputModel {
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
				m = updated.(inputModel)
				updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
				m = updated.(inputModel)
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.Equal(t, "deploy", m.textInput.Value())
				assert.Equal(t, 0, m.histIndex)
			},
		},
		"down arrow restores saved text": {
			setup: func() inputModel {
				m := newInputModel([]string{"deploy", "run"}, "")
				for _, r := range "ver" {
					updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
					m = updated.(inputModel)
				}
				return m
			},
			actions: func(m inputModel) inputModel {
				// Go up - saves "ver", shows "run"
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
				m = updated.(inputModel)
				// Go back down - restores "ver"
				updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
				m = updated.(inputModel)
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.Equal(t, "ver", m.textInput.Value())
				assert.Equal(t, 2, m.histIndex)
			},
		},
		"up at oldest entry does nothing": {
			setup: func() inputModel {
				return newInputModel([]string{"deploy"}, "")
			},
			actions: func(m inputModel) inputModel {
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
				m = updated.(inputModel)
				updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
				m = updated.(inputModel)
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.Equal(t, "deploy", m.textInput.Value())
				assert.Equal(t, 0, m.histIndex)
			},
		},
		"view renders border": {
			setup: func() inputModel {
				return newInputModel(nil, "")
			},
			actions: func(m inputModel) inputModel {
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				view := ansi.Strip(m.View())
				assert.Contains(t, view, "─")
				assert.Contains(t, view, "❯")
			},
		},
		"view renders banner when version set": {
			setup: func() inputModel {
				m := newInputModel(nil, "v1.0.0")
				m.width = 40
				return m
			},
			actions: func(m inputModel) inputModel { return m },
			assertFn: func(t *testing.T, m inputModel) {
				view := ansi.Strip(m.View())
				assert.Contains(t, view, "Slack CLI Shell")
				assert.Contains(t, view, "v1.0.0")
				assert.Contains(t, view, "❯")
			},
		},
		"view renders no banner when version empty": {
			setup: func() inputModel {
				m := newInputModel(nil, "")
				m.width = 40
				return m
			},
			actions: func(m inputModel) inputModel { return m },
			assertFn: func(t *testing.T, m inputModel) {
				view := ansi.Strip(m.View())
				assert.NotContains(t, view, "Slack CLI Shell")
				assert.Contains(t, view, "❯")
			},
		},
		"view is empty when done": {
			setup: func() inputModel {
				m := newInputModel(nil, "")
				m.done = true
				return m
			},
			actions: func(m inputModel) inputModel {
				return m
			},
			assertFn: func(t *testing.T, m inputModel) {
				assert.Equal(t, "", m.View())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			m := tc.setup()
			m = tc.actions(m)
			tc.assertFn(t, m)
		})
	}
}
