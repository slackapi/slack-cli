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

// Slack brand theme for prompt styling.
// Uses official Slack brand colors defined in colors.go.

import (
	"runtime"

	huh "charm.land/huh/v2"
	lipgloss "charm.land/lipgloss/v2"
)

// ThemeSlack returns a huh Theme styled with Slack brand colors.
func ThemeSlack() huh.Theme {
	return huh.ThemeFunc(themeSlack)
}

// themeSlack builds Slack-branded huh styles.
func themeSlack(isDark bool) *huh.Styles {
	t := huh.ThemeBase(isDark)

	// Focused styles apply to the field the user is currently interacting with.
	// Blurred styles apply to visible fields that are not currently active.
	t.Focused.Base = t.Focused.Base.
		BorderForeground(slackAubergine)
	t.Focused.Title = lipgloss.NewStyle().
		Foreground(slackAubergine).
		Bold(true)
	t.Focused.Description = lipgloss.NewStyle().
		Foreground(slackDescriptionText)
	t.Focused.ErrorIndicator = lipgloss.NewStyle().
		Foreground(slackRed).
		SetString(" *")
	t.Focused.ErrorMessage = lipgloss.NewStyle().
		Foreground(slackRed)

	// Select styles
	t.Focused.SelectSelector = lipgloss.NewStyle().
		Foreground(slackBlue).
		SetString(Chevron() + " ")
	t.Focused.Option = lipgloss.NewStyle().
		Foreground(slackOptionText)
	t.Focused.NextIndicator = lipgloss.NewStyle().
		Foreground(slackPool).
		MarginLeft(1).
		SetString("↓")
	t.Focused.PrevIndicator = lipgloss.NewStyle().
		Foreground(slackPool).
		MarginRight(1).
		SetString("↑")

	// Multi-select styles
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		Foreground(slackYellow).
		SetString(Chevron() + " ")
	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(slackGreen)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		Foreground(slackGreen).
		SetString("[✓] ")
	t.Focused.UnselectedOption = lipgloss.NewStyle().
		Foreground(slackOptionText)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		Foreground(slackLegalGray).
		SetString("[ ] ")

	// Text input styles
	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(slackYellow)
	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(slackBlue)
	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(slackPlaceholderText)
	t.Focused.TextInput.Text = lipgloss.NewStyle().
		Foreground(slackOptionText)

	// Button styles
	button := lipgloss.NewStyle().
		Padding(0, 2).
		MarginRight(1)
	t.Focused.FocusedButton = button.
		Foreground(lipgloss.Color("#ffffff")).
		Background(slackAubergine).
		Bold(true)
	t.Focused.BlurredButton = button.
		Foreground(slackLegalGray).
		Background(lipgloss.Color("#f8f8f8"))

	// Blurred field styles — subdued version of focused
	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.
		BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.SelectSelector = lipgloss.NewStyle().SetString("  ")
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	return t
}

// Chevron returns the select chevron character for the current platform.
// Unfortunately "❱" does not display on Windows Powershell.
// Limit "❱" to non-Windows until support is known for other operating systems.
func Chevron() string {
	if !isStyleEnabled || runtime.GOOS == "windows" {
		return ">"
	}
	return "❱"
}

// ThemeSurvey returns a huh Theme that matches the legacy survey prompt styling.
// Applied when experiment.Lipgloss is off.
func ThemeSurvey() huh.Theme {
	return huh.ThemeFunc(themeSurvey)
}

// themeSurvey builds huh styles matching the survey library's appearance.
func themeSurvey(isDark bool) *huh.Styles {
	t := huh.ThemeBase(isDark)

	ansiBlue := lipgloss.ANSIColor(blue)
	ansiGray := lipgloss.ANSIColor(gray)
	ansiGreen := lipgloss.ANSIColor(green)
	ansiRed := lipgloss.ANSIColor(red)

	t.Focused.Title = lipgloss.NewStyle().
		Foreground(ansiGray).
		Bold(true)
	t.Focused.ErrorIndicator = lipgloss.NewStyle().
		Foreground(ansiRed).
		SetString(" *")
	t.Focused.ErrorMessage = lipgloss.NewStyle().
		Foreground(ansiRed)

	// Select styles
	t.Focused.SelectSelector = lipgloss.NewStyle().
		Foreground(ansiBlue).
		Bold(true).
		SetString(Chevron() + " ")
	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(ansiBlue).
		Bold(true)

	// Multi-select styles
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		Foreground(ansiBlue).
		Bold(true).
		SetString(Chevron() + " ")
	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		Foreground(ansiGreen).
		SetString("[x] ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		Bold(true).
		SetString("[ ] ")

	return t
}
