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

package project

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/ansi"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/stretchr/testify/assert"
)

// doAllUpdates recursively processes all commands returned by form updates,
// including batch messages from OptionsFunc evaluations and group transitions.
// This mirrors the helper in huh's own test suite.
func doAllUpdates(f *huh.Form, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	var cmds []tea.Cmd
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, subcommand := range msg {
			doAllUpdates(f, subcommand)
		}
		return
	default:
		_, result := f.Update(msg)
		cmds = append(cmds, result)
	}
	doAllUpdates(f, tea.Batch(cmds...))
}

func TestBuildTemplateSelectionForm(t *testing.T) {
	t.Run("renders category and template on one screen", func(t *testing.T) {
		cm := shared.NewClientsMock()
		cm.AddDefaultMocks()
		clients := shared.NewClientFactory(cm.MockClientFactory())

		var category, template string
		f := buildTemplateSelectionForm(clients, &category, &template)
		doAllUpdates(f, f.Init())

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Select an app:")
		assert.Contains(t, view, "Starter app")
		assert.Contains(t, view, "AI Agent app")
		assert.Contains(t, view, "Automation app")
		assert.Contains(t, view, "View more samples")
		assert.Contains(t, view, "Select a language:")
	})

	t.Run("selecting a category updates template options", func(t *testing.T) {
		cm := shared.NewClientsMock()
		cm.AddDefaultMocks()
		clients := shared.NewClientFactory(cm.MockClientFactory())

		var category, template string
		f := buildTemplateSelectionForm(clients, &category, &template)
		doAllUpdates(f, f.Init())

		// Submit first option (Starter app -> getting-started)
		_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
		doAllUpdates(f, cmd)

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Bolt for JavaScript")
		assert.Contains(t, view, "Bolt for Python")
	})

	t.Run("selecting view more samples shows browse option", func(t *testing.T) {
		cm := shared.NewClientsMock()
		cm.AddDefaultMocks()
		clients := shared.NewClientFactory(cm.MockClientFactory())

		var category, template string
		f := buildTemplateSelectionForm(clients, &category, &template)
		doAllUpdates(f, f.Init())

		// Navigate down to "View more samples" (4th option, index 3)
		_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyDown})
		doAllUpdates(f, cmd)
		_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyDown})
		doAllUpdates(f, cmd)
		_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyDown})
		doAllUpdates(f, cmd)
		_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyEnter})
		doAllUpdates(f, cmd)

		assert.Equal(t, viewMoreSamples, category)
		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Browse sample gallery...")
	})

	t.Run("automation category shows Deno option", func(t *testing.T) {
		cm := shared.NewClientsMock()
		cm.AddDefaultMocks()
		clients := shared.NewClientFactory(cm.MockClientFactory())

		var category, template string
		f := buildTemplateSelectionForm(clients, &category, &template)
		doAllUpdates(f, f.Init())

		// Navigate to Automation app (3rd option, index 2) and submit
		_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyDown})
		doAllUpdates(f, cmd)
		_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyDown})
		doAllUpdates(f, cmd)
		_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyEnter})
		doAllUpdates(f, cmd)

		view := ansi.Strip(f.View())
		assert.Contains(t, view, "Deno Slack SDK")
	})

	t.Run("complete flow selects a template", func(t *testing.T) {
		cm := shared.NewClientsMock()
		cm.AddDefaultMocks()
		clients := shared.NewClientFactory(cm.MockClientFactory())

		var category, template string
		f := buildTemplateSelectionForm(clients, &category, &template)
		doAllUpdates(f, f.Init())

		// Select first category (Starter app)
		_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})
		doAllUpdates(f, cmd)
		// Select first template (Bolt for JavaScript)
		_, cmd = f.Update(tea.KeyMsg{Type: tea.KeyEnter})
		doAllUpdates(f, cmd)

		assert.Equal(t, "slack-cli#getting-started", category)
		assert.Equal(t, "slack-samples/bolt-js-starter-template", template)
	})

	t.Run("uses Slack theme", func(t *testing.T) {
		cm := shared.NewClientsMock()
		cm.AddDefaultMocks()
		clients := shared.NewClientFactory(cm.MockClientFactory())

		var category, template string
		f := buildTemplateSelectionForm(clients, &category, &template)
		doAllUpdates(f, f.Init())

		view := f.View()
		assert.Contains(t, view, "┃")
	})
}
