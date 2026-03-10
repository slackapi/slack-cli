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
	"regexp"
	"runtime"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/kyokomi/emoji/v2"
	"github.com/logrusorgru/aurora/v4"
)

// isStyleEnabled specifies if styles should be shown in outputs
var isStyleEnabled = false

// isColorShown specifies if colors should be displayed in outputs
var isColorShown = isStyleEnabled

// isLinkShown specifies if hyperlinks should be formatted
var isLinkShown = isStyleEnabled

// isCharmEnabled specifies if lipgloss/charm styling should be used instead of aurora
var isCharmEnabled = false

// RemoveANSI uses regex to strip ANSI colour codes
//
// Shamelessly stolen from https://github.com/acarl005/stripansi/blob/master/stripansi.go
func RemoveANSI(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var ansiRegex = regexp.MustCompile(ansi)
	return ansiRegex.ReplaceAllString(str, "")
}

// ToggleStyles sets styles and formatting values to the active state
func ToggleStyles(active bool) {
	isStyleEnabled = active
	isColorShown = active
	isLinkShown = active
}

// ToggleCharm enables lipgloss-based styling when set to true
func ToggleCharm(active bool) {
	isCharmEnabled = active
}

// render applies a lipgloss style to text, returning plain text when colors are disabled.
func render(s lipgloss.Style, text string) string {
	if !isColorShown {
		return text
	}
	return s.Render(text)
}

func Emoji(alias string) string {
	if !isColorShown || strings.TrimSpace(alias) == "" {
		return ""
	}

	// windows terminal does not support emoji's yet.
	// We have to override the value to a simple text to avoid
	// weird character printouts
	if runtime.GOOS == "windows" {
		return CommandText("> ")
	}

	// Add padding to some emojis
	var padding string

	switch alias {
	case "cloud_with_lightning":
		padding = " "
	case "file_cabinet":
		padding = " "
	case "gear":
		padding = " "
	case "hook":
		padding = " "
	case "house_buildings":
		padding = " "
	case "label":
		padding = " "
	case "potted_plant":
		padding = " "
	case "wastebasket":
		padding = " "
	}

	return emoji.Sprint(":"+alias+":") + padding
}

/*
Color styles
*/

// Secondary dims the displayed text
func Secondary(text string) string {
	if !isCharmEnabled {
		return legacySecondary(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackDescriptionText), text)
}

// CommandText emphasizes command text
func CommandText(text string) string {
	if !isCharmEnabled {
		return legacyCommandText(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackBlue).Bold(true), text)
}

// LinkText underlines and formats the provided path
func LinkText(path string) string {
	if !isCharmEnabled {
		return legacyLinkText(path)
	}
	return render(lipgloss.NewStyle().Foreground(slackDescriptionText).Underline(true), path)
}

func Selector(text string) string {
	if !isCharmEnabled {
		return legacySelector(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackGreen).Bold(true), text)
}

func Error(text string) string {
	if !isCharmEnabled {
		return legacyError(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackRed).Bold(true), text)
}

func Warning(text string) string {
	if !isCharmEnabled {
		return legacyWarning(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackYellow).Bold(true), text)
}

func Header(text string) string {
	if !isCharmEnabled {
		return legacyHeader(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackAubergine).Bold(true), strings.ToUpper(text))
}

func Input(text string) string {
	if !isCharmEnabled {
		return legacyInput(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackBlue), text)
}

// Green applies green color to text without bold
func Green(text string) string {
	if !isCharmEnabled {
		return legacyGreen(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackGreen), text)
}

// Red applies red color to text without bold
func Red(text string) string {
	if !isCharmEnabled {
		return legacyRed(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackRedDark), text)
}

// Yellow applies yellow color to text without bold
func Yellow(text string) string {
	if !isCharmEnabled {
		return legacyYellow(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackYellow), text)
}

// Gray applies a subdued gray color to text
func Gray(text string) string {
	if !isCharmEnabled {
		return legacyGray(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackLegalGray), text)
}

/*
Text styles
*/

// Bright is a strong bold version of the text
func Bright(text string) string {
	if !isCharmEnabled {
		return legacyBright(text)
	}
	return render(lipgloss.NewStyle().Bold(true), text)
}

// Bold brightly emboldens the provided text
func Bold(text string) string {
	if !isCharmEnabled {
		return legacyBold(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackOptionText).Bold(true), text)
}

// Darken adds a bold gray shade to text
func Darken(text string) string {
	if !isCharmEnabled {
		return legacyDarken(text)
	}
	return render(lipgloss.NewStyle().Foreground(slackPlaceholderText).Bold(true), text)
}

// Faint resets all effects then decreases text intensity
func Faint(text string) string {
	if !isColorShown {
		return text
	}
	if !isCharmEnabled {
		return legacyFaint(text)
	}
	return lipgloss.NewStyle().Faint(true).Render(text)
}

// Highlight adds emphasis to text
func Highlight(text string) string {
	if !isCharmEnabled {
		return legacyHighlight(text)
	}
	return render(lipgloss.NewStyle().Bold(true), text)
}

// Underline underscores the given text
func Underline(text string) string {
	if !isCharmEnabled {
		return legacyUnderline(text)
	}
	return render(lipgloss.NewStyle().Underline(true), text)
}

/*
Lexical
*/
func Pluralize(singular string, plural string, count int) string {
	if count == 1 {
		return singular
	}
	return plural
}

// ════════════════════════════════════════════════════════════════════════════════
// DEPRECATED: Legacy aurora styling
//
// Delete this entire section, the aurora import, and the ANSI color constants
// when the charm experiment is permanently enabled.
// ════════════════════════════════════════════════════════════════════════════════

const (
	blueDark    = 32
	blue        = 39
	grayDark    = 236
	gray        = 246
	grayLight   = 12
	green       = 29
	red         = 196
	redDark     = 1
	whiteOffset = 21 // 235 in ANSI
	yellow      = 178
)

// DEPRECATED: Styler returns an aurora instance for legacy styling.
// Use the style functions (Secondary, CommandText, Error, etc.) instead.
func Styler() *aurora.Aurora {
	config := aurora.NewConfig()
	config.Colors = isColorShown
	config.Hyperlinks = isLinkShown
	return aurora.New(config.Options()...)
}

func legacySecondary(text string) string   { return Styler().Gray(grayLight, text).String() }
func legacyCommandText(text string) string { return Styler().Index(blue, text).Bold().String() }
func legacyLinkText(path string) string    { return Styler().Gray(grayLight, path).Underline().String() }
func legacySelector(text string) string    { return Styler().Index(green, text).Bold().String() }
func legacyError(text string) string       { return Styler().Index(red, text).Bold().String() }
func legacyWarning(text string) string     { return Styler().Index(yellow, text).Bold().String() }
func legacyHeader(text string) string      { return Styler().Bold(strings.ToUpper(text)).String() }
func legacyInput(text string) string       { return Styler().Index(blue, text).String() }
func legacyGreen(text string) string       { return Styler().Index(green, text).String() }
func legacyRed(text string) string         { return Styler().Index(redDark, text).String() }
func legacyYellow(text string) string      { return Styler().Index(yellow, text).String() }
func legacyGray(text string) string        { return Styler().Gray(13, text).String() }
func legacyBright(text string) string      { return Styler().Bold(text).String() }
func legacyBold(text string) string        { return Styler().Gray(whiteOffset, text).Bold().String() }
func legacyDarken(text string) string      { return Styler().Index(gray, text).Bold().String() }
func legacyFaint(text string) string       { return "\x1b[0;2m" + text + "\x1b[0m" }
func legacyHighlight(text string) string   { return Styler().Bold(text).String() }
func legacyUnderline(text string) string   { return Styler().Underline(text).String() }
