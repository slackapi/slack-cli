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
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
)

const LocalRunNameTag = "(local)"

type TextSection struct {
	Emoji     string
	Text      string
	Secondary []string
}

/*
Create inline padding to display map rows in clean way
*/
func Mapf(m map[string]string) string {
	var formattedText strings.Builder
	labelLength := getKeyLength(m)

	// Keep track of last element for appending newline
	i := 1
	newLineChar := "\n"

	for label, text := range m {
		if i == len(m) {
			newLineChar = ""
		}
		row := Secondary(fmt.Sprintf("%-*s:  %s%s", labelLength, label, text, newLineChar))
		formattedText.WriteString(row)
		i++
	}

	return formattedText.String()
}

// getKeyLength returns the length of longest key in map
func getKeyLength(m map[string]string) int {
	length := 0
	// Iterate through keys
	for k := range m {
		if l := len(k); l > length {
			length = l
		}
	}
	return length
}

/*
Standardize text section formatting
*/
func Sectionf(section TextSection) string {
	var textSection strings.Builder
	if section.Text == "" {
		return ""
	}

	// Header
	textSection.WriteString(SectionHeaderf(section.Emoji, section.Text))

	// Secondary
	for _, str := range section.Secondary {
		if str != "" {
			text := SectionSecondaryf("%s", str)
			textSection.WriteString(text)
		}
	}

	return textSection.String()
}

// SectionHeaderf returns a standard formatted section header
// TODO - support the the arg text ...string to do formatting
func SectionHeaderf(emoji, text string) string {
	if text == "" {
		return ""
	}

	emoji = Emoji(emoji)
	return fmt.Sprintf("%s%s\n", emoji, text)
}

// SectionSecondaryf returns a standard formatted section secondary text
func SectionSecondaryf(format string, a ...interface{}) string {
	if format == "" {
		return ""
	}

	// Format the text
	text := fmt.Sprintf(format, a...)
	if text == "" {
		return ""
	}

	// Indent new lines of secondary output
	lines := []string{}
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			lines = append(lines, Indent(Secondary(line))+"\n")
		} else {
			lines = append(lines, "\n")
		}
	}

	return strings.Join(lines, "")
}

// TODO: this is reusing logic from process.go, but can't import because of circular dependencies
func Commandf(command string, isPrimary bool) string {
	commandText := processName() + " " + command

	if !isColorShown {
		return fmt.Sprintf("`%s`", commandText)
	}

	if !isPrimary {
		commandText = Highlight(commandText)
	} else {
		commandText = CommandText(commandText)
	}
	return commandText
}

func Indent(text string) string {
	return "   " + text
}

// Tracef prepares the output string for a test trace ID with optional values
//
// Formatted as TRACE_ID or TRACE_ID=VALUE or TRACE_ID=VALUE,VALUE,VALUE
func Tracef(traceID string, traceValues ...string) string {
	traceID = strings.TrimSpace(traceID)
	for i, traceValue := range traceValues {
		traceValues[i] = strings.TrimSpace(traceValue)
	}
	var s string
	if len(traceValues) > 0 {
		s = fmt.Sprintf("%s=%s", traceID, strings.Join(traceValues, ","))
	} else {
		s = traceID
	}
	return s
}

func processName() string {
	return filepath.Base(os.Args[0])
}

// HomePath replaces the home directory with the shorthand "~" in a filepath
func HomePath(filepath string) string {
	if runtime.GOOS == "windows" {
		return filepath
	}

	homeDirPath, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(filepath, homeDirPath) {
		filepath = strings.Replace(filepath, homeDirPath, "~", 1)
	}
	return filepath
}

// SurveyIcons returns customizations to the appearance of prompts
func SurveyIcons() survey.AskOpt {
	if !isStyleEnabled {
		core.DisableColor = true
	}

	cursor := ">"
	// Unfortunately "❱" does not display on Windows Powershell
	// Limit "❱" to macOS until support is known for other operating systems
	if isStyleEnabled && runtime.GOOS == "darwin" {
		cursor = "❱"
	}

	// Customize the appearance of each survey prompt
	return survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Text = cursor
		icons.SelectFocus.Format = fmt.Sprintf("%d+b", blue)
		icons.MarkedOption.Format = fmt.Sprintf("%d+b", blue)

		icons.Question.Text = "?"
		icons.Question.Format = fmt.Sprintf("%d+hb", gray)
	})
}

// ExampleCommand contains a command with a descriptive meaning
type ExampleCommand struct {
	// Command is the command to be run, not including the process name
	Command string
	// Meaning plainly describes the result of running the command
	Meaning string
}

// ExampleCommandsf returns a multiline string of commands and comments
func ExampleCommandsf(commands []ExampleCommand) string {
	longestExample := 0
	longestLine := 0

	// Find the longest command and line for splitting measurements
	for _, cmd := range commands {
		command := fmt.Sprintf("%s %s", processName(), cmd.Command)
		if len(command) > longestExample {
			longestExample = len(command)
		}
		unformattedLength := len(command) + len(cmd.Meaning)
		if unformattedLength > longestLine {
			longestLine = unformattedLength
		}
	}

	examples := []string{}
	for _, cmd := range commands {
		command := fmt.Sprintf("$ %s %s", processName(), cmd.Command)
		spaces := strings.Repeat(" ", longestExample-len(command)+4)
		meaning := fmt.Sprintf("# %s", cmd.Meaning)

		// Split long examples into multiple lines
		if 2+len(command)+len(spaces)+len(meaning)+4 > 80 {
			examples = append(examples, fmt.Sprintf("\n%s\n%s", meaning, command))
		} else {
			examples = append(examples, fmt.Sprintf("%s%s%s", command, spaces, meaning))
		}
	}
	output := strings.TrimSpace(strings.Join(examples, "\n"))
	return output
}

// ExampleTemplatef indents and styles command examples for the help messages
func ExampleTemplatef(template string) string {
	lines := strings.Split(template, "\n")
	re := regexp.MustCompile(`(^#.*$)|(  #.*$)`)
	examples := []string{}
	for _, cmd := range lines {
		example := ""
		if cmd != "" {
			styled := re.ReplaceAllStringFunc(cmd, Secondary)
			example = fmt.Sprintf("  %s", styled)
		}
		examples = append(examples, example)
	}
	return strings.Join(examples, "\n")
}

// LocalRunDisplayName appends the (local) tag to apps created by the run command
func LocalRunDisplayName(name string) string {
	return name + " " + LocalRunNameTag
}

// AppIDLabel formats the appID to indicate the installation status
func AppIDLabel(appID string, isUninstalled bool) string {
	if appID != "" && isUninstalled {
		return "\x1b[0m" + Secondary(appID) + Secondary(" (uninstalled)")
	}

	return "\x1b[0m" + Selector(appID)
}

// AppSelectLabel formats a label with the environment and installed status of an app
func AppSelectLabel(environment string, appID string, isUninstalled bool) string {
	return fmt.Sprintf("%s %s", environment, AppIDLabel(appID, isUninstalled))
}

// TeamAppSelectLabel formats an app select label for a team app.
// The environment of this app is implied by the prompting function.
func TeamAppSelectLabel(teamDomain string, teamID string, appID string, isUninstalled bool) string {
	return fmt.Sprintf("%s %s", TeamSelectLabel(teamDomain, teamID), AppIDLabel(appID, isUninstalled))
}

// TeamSelectLabel formats a team label with the teamID
func TeamSelectLabel(teamDomain string, teamID string) string {
	return fmt.Sprintf("%s %s", teamDomain, Faint(teamID))
}

// TimeAgo formats the time since the provided datetime into a friendly format
func TimeAgo(datetime int) string {
	now := time.Now()
	ago := time.Unix(int64(datetime), 0)
	dir := "ago"

	difference := now.Sub(ago)
	if difference < 0 {
		difference = -difference
		dir = "until"
	}

	seconds := int(math.Round(difference.Seconds()))
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case years > 0:
		durr := Pluralize("year", "years", years)
		return fmt.Sprintf("%d %s %s", years, durr, dir)
	case months > 0:
		durr := Pluralize("month", "months", months)
		return fmt.Sprintf("%d %s %s", months, durr, dir)
	case weeks > 0:
		durr := Pluralize("week", "weeks", weeks)
		return fmt.Sprintf("%d %s %s", weeks, durr, dir)
	case days > 0:
		durr := Pluralize("day", "days", days)
		return fmt.Sprintf("%d %s %s", days, durr, dir)
	case hours > 0:
		durr := Pluralize("hour", "hours", hours)
		return fmt.Sprintf("%d %s %s", hours, durr, dir)
	case minutes > 0:
		durr := Pluralize("minute", "minutes", minutes)
		return fmt.Sprintf("%d %s %s", minutes, durr, dir)
	default:
		durr := Pluralize("second", "seconds", seconds)
		return fmt.Sprintf("%d %s %s", seconds, durr, dir)
	}
}
