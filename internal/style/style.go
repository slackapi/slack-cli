// Copyright 2022-2025 Salesforce, Inc.
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

	"github.com/kyokomi/emoji/v2"
	"github.com/logrusorgru/aurora/v4"
)

// isStyleEnabled specifies if styles should be shown in outputs
var isStyleEnabled = false

// isColorShown specifies if colors should be displayed in outputs
var isColorShown = isStyleEnabled

// isLinkShown specifies if hyperlinks should be formatted
var isLinkShown = isStyleEnabled

// ANSI escape sequence color code
//
// Non-grayscale codes are selected from
// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
//
// Grayscale codes might be ANSI or selected from
// https://github.com/logrusorgru/aurora#grayscale
//
// TODO: check whether tty supports 256; if not, simplify to top 8 colors
// https://unix.stackexchange.com/questions/9957/how-to-check-if-bash-can-print-colors
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

func Styler() *aurora.Aurora {
	config := aurora.NewConfig()
	config.Colors = isColorShown
	config.Hyperlinks = isLinkShown

	return aurora.New(config.Options()...)
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
	return Styler().Gray(grayLight, text).String()
}

// CommandText emphasizes command text
func CommandText(text string) string {
	return Styler().Index(blue, text).Bold().String()
}

// LinkText underlines and formats the provided path
func LinkText(path string) string {
	return Styler().Gray(grayLight, path).Underline().String()
}

func Selector(text string) string {
	return Styler().Index(green, text).Bold().String()
}

func Error(text string) string {
	return Styler().Index(red, text).Bold().String()
}

func Warning(text string) string {
	return Styler().Index(yellow, text).Bold().String()
}

func Header(text string) string {
	return Styler().Bold(strings.ToUpper(text)).String()
}

func Input(text string) string {
	return Styler().Index(blue, text).String()
}

/*
Text styles
*/

// Bright is a strong bold version of the text
func Bright(text string) string {
	return Styler().Bold(text).String()
}

// Bold brightly emboldens the provided text
func Bold(text string) string {
	return Styler().Gray(whiteOffset, text).Bold().String()
}

// Darken adds a bold gray shade to text
func Darken(text string) string {
	return Styler().Index(gray, text).Bold().String()
}

// Faint resets all effects then decreases text intensity
func Faint(text string) string {
	if !isColorShown {
		return text
	}
	return "\x1b[0;2m" + text + "\x1b[0m"
}

// Highlight adds emphasis to text
func Highlight(text string) string {
	return Styler().Bold(text).String()
}

// Underline underscores the given text
func Underline(text string) string {
	return Styler().Underline(text).String()
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
