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
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/slackapi/slack-cli/internal/shared"
)

// inputModel wraps a bubbles textinput with shell history navigation.
type inputModel struct {
	textInput     textinput.Model
	history       []string
	histIndex     int    // len(history) = "new line" position
	saved         string // user's in-progress text before navigating history
	done          bool
	value         string
	width         int
	bannerVersion string // non-empty = show banner above input
}

// readLine runs a short-lived bubbletea program to collect one line of input.
func readLine(clients *shared.ClientFactory, history []string, bannerVersion string) (string, error) {
	m := newInputModel(history, bannerVersion)
	p := tea.NewProgram(m,
		tea.WithInput(clients.IO.ReadIn()),
		tea.WithOutput(clients.IO.WriteOut()),
	)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	final := result.(inputModel)
	return final.value, nil
}

func newInputModel(history []string, bannerVersion string) inputModel {
	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.PromptStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#e8a400"))
	ti.Focus()

	// Disable built-in suggestion navigation (we use Up/Down for history)
	ti.KeyMap.NextSuggestion.SetEnabled(false)
	ti.KeyMap.PrevSuggestion.SetEnabled(false)

	return inputModel{
		textInput:     ti,
		history:       history,
		histIndex:     len(history),
		bannerVersion: bannerVersion,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			m.value = m.textInput.Value()
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.done = true
			m.value = "exit"
			return m, tea.Quit
		case tea.KeyUp:
			if m.histIndex > 0 {
				// Save current text on first Up press
				if m.histIndex == len(m.history) {
					m.saved = m.textInput.Value()
				}
				m.histIndex--
				m.textInput.SetValue(m.history[m.histIndex])
				m.textInput.CursorEnd()
			}
			return m, nil
		case tea.KeyDown:
			if m.histIndex < len(m.history) {
				m.histIndex++
				if m.histIndex == len(m.history) {
					m.textInput.SetValue(m.saved)
				} else {
					m.textInput.SetValue(m.history[m.histIndex])
				}
				m.textInput.CursorEnd()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.done {
		return ""
	}
	w := m.width
	if w <= 0 {
		w = 80
	}
	line := lipgloss.NewStyle().Foreground(colorPool).Render(strings.Repeat("─", w))
	content := " " + m.textInput.View()
	input := line + "\n" + content + "\n" + line

	if m.bannerVersion != "" {
		return bannerView(w, m.bannerVersion) + "\n" + input
	}
	return input
}
