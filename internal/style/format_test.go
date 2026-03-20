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
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetKeyLengthZero(t *testing.T) {
	var keys = map[string]string{
		"": "the zero key",
	}

	if getKeyLength(keys) != 0 {
		t.Error("the longest key has a length greater than zero")
	}
}

func TestGetKeyLengthMatched(t *testing.T) {
	var keys = map[string]string{
		"key1": "unlocks the building",
		"key2": "unlocks the room",
	}

	if getKeyLength(keys) != 4 {
		t.Error("the longest key should have length 4")
	}
}

func TestGetKeyLengthLong(t *testing.T) {
	var keys = map[string]string{
		"longer_key1":    "locks the building",
		"very_long_key2": "locks the room",
	}

	if getKeyLength(keys) != 14 {
		t.Error("the longest key `very_long_key2` should have length 14")
	}
}

func TestGetKeyLengthFirst(t *testing.T) {
	var keys = map[string]string{
		"longest_key1": "short value",
		"short_key2":   "longer value",
	}

	if getKeyLength(keys) != 12 {
		t.Error("the longest key `longest_key1` should have length 12")
	}
}

// Verify no text is output with an empty input text
func TestSectionfEmpty(t *testing.T) {
	formattedText := Sectionf(TextSection{
		Emoji:     "",
		Text:      "",
		Secondary: []string{},
	})
	if formattedText != "" {
		t.Error("non-zero text returned when none was expected")
	}
}

// Verify no text is output with an empty input text
func TestSectionfHeader(t *testing.T) {
	expected := Emoji("tada") + "Congrats\n" + Indent(Secondary("You did it")) + "\n"
	formattedText := Sectionf(TextSection{
		Emoji:     "tada",
		Text:      "Congrats",
		Secondary: []string{"You did it"},
	})
	if formattedText != expected {
		t.Error("section is not formatted as expected")
	}
}

// Verify text begins immediately if no emoji is input
func TestSectionfEmptyEmoji(t *testing.T) {
	text := "On the left. Where I like it."
	formattedText := Sectionf(TextSection{
		Emoji:     "",
		Text:      text,
		Secondary: []string{},
	})

	if formattedText != text+"\n" {
		t.Error("additional spacing added to text")
	}
}

// Verify no text is output with an empty input text
func TestSectionHeaderfEmpty(t *testing.T) {
	text := ""
	formattedText := SectionHeaderf("tada", text)
	if formattedText != "" {
		t.Error("non-zero text returned when none was expected")
	}
}

// Verify no text is output with an empty input
func TestSectionSecondaryfEmpty(t *testing.T) {
	text := ""
	formattedText := SectionSecondaryf("%s", text)
	if formattedText != "" {
		t.Log(formattedText)
		t.Error("non-zero text returned when none was expected")
	}
}

// Verify plain string is preserved and properly indented
func TestSectionSecondaryfPlain(t *testing.T) {
	text := "If you have a moment, go grab a glass of water!"
	formattedText := SectionSecondaryf("%s", text)
	if !strings.Contains(formattedText, text) {
		t.Error("input text is not preserved")
	}
	if formattedText != Indent(Secondary(text))+"\n" {
		t.Error("output is not indented")
	}
}

// Verify string formats input variables
func TestSectionSecondaryfFormat(t *testing.T) {
	text := "App ID: %s\tStatus: %s"
	appID := "A123456"
	status := "Installed"
	formattedText := SectionSecondaryf(text, appID, status)
	if !strings.Contains(formattedText, "App ID: A123456\tStatus: Installed") {
		t.Error("formatted string does not contain variables")
	}
}

// Verify multi-line input is properly indented
func TestSectionSecondaryfIndent(t *testing.T) {
	text := "L1\nL2\nL3"
	formattedText := SectionSecondaryf("%s", text)

	for i, line := range strings.Split(text, "\n") {
		lines := strings.Split(formattedText, "\n")
		if strings.Compare(lines[i], Indent(Secondary(line))) != 0 {
			t.Errorf("new line not properly indented\n"+
				"expect: *%s*\nactual: *%s*", Indent(Secondary(line)), lines[i])
		}
	}
}

// Verify a `process command`-like string is presented
func TestCommandfPrimary(t *testing.T) {
	// rename the process for fuzz-like testing
	processTemp := os.Args[0]
	process := "renamed-slack-command"
	os.Args[0] = "renamed-slack-command"
	command := "feedback"

	formatted := Commandf(command, true)
	if !strings.Contains(formatted, process+" "+command) {
		t.Errorf("a `process command`-like string is not present in output:\n%s", formatted)
	}

	os.Args[0] = processTemp
}

// Verify a "process command"-like string is presented
func TestCommandfSecondary(t *testing.T) {
	// Rename the process for fuzzy testing
	processTemp := os.Args[0]
	process := "a-renamed-slack-cli"
	os.Args[0] = "a-renamed-slack-cli"
	command := "feedback"

	formatted := Commandf(command, false)
	if !strings.Contains(formatted, process+" "+command) {
		t.Errorf("a `process command`-like string is not present")
	}

	os.Args[0] = processTemp
}

// Verify the text indented is not modified
func TestIndent(t *testing.T) {
	text := "a few spaces are expected at the start of this line, but no other changes"
	indented := Indent(text)
	if !strings.Contains(indented, text) {
		t.Error("original text is not preserved")
	}
}

func TestTracef(t *testing.T) {
	tests := map[string]struct {
		traceID     string
		traceValues []string
		expected    string
	}{
		"only the trace id is provided": {
			traceID:  "TRACE_ID_1",
			expected: "TRACE_ID_1",
		},
		"a single value with the trace id": {
			traceID:     "TRACE_ID_2",
			traceValues: []string{"VALUE_A"},
			expected:    "TRACE_ID_2=VALUE_A",
		},
		"multiple values and the trace id": {
			traceID:     "TRACE_ID_3",
			traceValues: []string{"VALUE_A", "VALUE_B"},
			expected:    "TRACE_ID_3=VALUE_A,VALUE_B",
		},
		"surrounding spaces are removed": {
			traceID:     "  TRACE_ID_4    ",
			traceValues: []string{" VALUE_A", "   VALUE_B   \t"},
			expected:    "TRACE_ID_4=VALUE_A,VALUE_B",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			trace := Tracef(tc.traceID, tc.traceValues...)
			assert.Equal(t, tc.expected, trace)
		})
	}
}

/*
* Example commands
 */

func TestStyleFlags(t *testing.T) {
	tests := map[string]struct {
		charmEnabled bool
		input        string
		expectedFunc func() string
	}{
		"short and long flag with type and description": {
			charmEnabled: true,
			input:        "  -s, --long string   Description text",
			expectedFunc: func() string { return Yellow("  -s, --long string   ") + Secondary("Description text") },
		},
		"long-only flag with description": {
			charmEnabled: true,
			input:        "      --verbose       Enable verbose output",
			expectedFunc: func() string { return Yellow("      --verbose       ") + Secondary("Enable verbose output") },
		},
		"plain text without flag pattern returned unchanged": {
			charmEnabled: true,
			input:        "some plain text",
			expectedFunc: func() string { return "some plain text" },
		},
		"empty string returned unchanged": {
			charmEnabled: true,
			input:        "",
			expectedFunc: func() string { return "" },
		},
		"multiline flag output": {
			charmEnabled: true,
			input:        "  -a, --all           Show all\n      --verbose       Enable verbose",
			expectedFunc: func() string {
				return Yellow("  -a, --all           ") + Secondary("Show all") + "\n" + Yellow("      --verbose       ") + Secondary("Enable verbose")
			},
		},
		"charm disabled returns input unchanged": {
			charmEnabled: false,
			input:        "  -s, --long string   Description text",
			expectedFunc: func() string { return "  -s, --long string   Description text" },
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ToggleStyles(tc.charmEnabled)
			ToggleLipgloss(tc.charmEnabled)
			defer func() {
				ToggleStyles(false)
				ToggleLipgloss(false)
			}()
			actual := StyleFlags(tc.input)
			assert.Equal(t, tc.expectedFunc(), actual)
		})
	}
}

func Test_ExampleCommandsf(t *testing.T) {
	tests := map[string]struct {
		name     string
		commands []ExampleCommand
		expected []string
	}{
		"verify short outputs are listed inline": {
			commands: []ExampleCommand{
				{Command: "create", Meaning: "Make a project"},
				{Command: "branch", Meaning: "Use a branch"},
				{Command: "template", Meaning: "Choose a template"},
			},
			expected: []string{
				"$ slack create    # Make a project",
				"$ slack branch    # Use a branch",
				"$ slack template  # Choose a template",
			},
		},
		"verify line breaks occur with a long command and meaning": {
			commands: []ExampleCommand{
				{Command: "create", Meaning: "Create a new project from a selected template"},
				{Command: "create my-project -t sample/repo-url", Meaning: "Create a named project from a given template"},
			},
			expected: []string{
				"# Create a new project from a selected template",
				"$ slack create",
				"",
				"# Create a named project from a given template",
				"$ slack create my-project -t sample/repo-url",
			},
		},
	}
	for name, tc := range tests {
		commandName := os.Args[0]
		os.Args[0] = "slack"
		defer func() {
			os.Args[0] = commandName
		}()
		t.Run(name, func(t *testing.T) {
			actual := ExampleCommandsf(tc.commands)
			assert.Equal(t, tc.expected, strings.Split(actual, "\n"))
		})
	}
}

func Test_ExampleTemplatef(t *testing.T) {
	globalColorShown := isColorShown
	isColorShown = true
	defer func() {
		isColorShown = globalColorShown
	}()
	tests := map[string]struct {
		withColorShown bool
		template       []string
		expected       []string
	}{
		"text is indented when present": {
			template: []string{
				"# Create a new project from a selected template",
				"$ slack create",
				"",
				"# Create a named project from a given template",
				"$ slack create my-project -t sample/repo-url",
			},
			expected: []string{
				"  # Create a new project from a selected template",
				"  $ slack create",
				"",
				"  # Create a named project from a given template",
				"  $ slack create my-project -t sample/repo-url",
			},
		},
		"only comments are highlighted": {
			withColorShown: true,
			template: []string{
				"# Standalone comment before a command",
				"$ slack create  # Comment following a command",
				"",
				"$ slack datastore '#status = :status'  # Ignore commands",
			},
			expected: []string{
				fmt.Sprintf("  %s", Secondary("# Standalone comment before a command")),
				fmt.Sprintf("  $ slack create%s", Secondary("  # Comment following a command")),
				"",
				fmt.Sprintf("  $ slack datastore '#status = :status'%s", Secondary("  # Ignore commands")),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			localColorStatus := isColorShown
			isColorShown = tc.withColorShown
			defer func() {
				isColorShown = localColorStatus
			}()
			actual := ExampleTemplatef(strings.Join(tc.template, "\n"))
			assert.Equal(t, strings.Join(tc.expected, "\n"), actual)
		})
	}
}

func Test_ExampleTemplatef_Charm(t *testing.T) {
	defer func() {
		ToggleStyles(false)
		ToggleLipgloss(false)
	}()
	ToggleStyles(true)
	ToggleLipgloss(true)

	template := []string{
		"# Create a new project from a selected template",
		"$ slack create",
		"",
		"$ slack create my-project -t sample/repo-url  # Create a named project",
	}
	expected := []string{
		fmt.Sprintf("  %s", Secondary("# Create a new project from a selected template")),
		fmt.Sprintf("  %s%s", Yellow("$ "), CommandText("slack create")),
		"",
		fmt.Sprintf("  %s%s%s", Yellow("$ "), CommandText("slack create my-project -t sample/repo-url"), Secondary("  # Create a named project")),
	}
	actual := ExampleTemplatef(strings.Join(template, "\n"))
	assert.Equal(t, strings.Join(expected, "\n"), actual)
}

func TestMapf(t *testing.T) {
	t.Run("formats a map with aligned keys", func(t *testing.T) {
		m := map[string]string{
			"key": "value",
		}
		result := Mapf(m)
		assert.Contains(t, result, "key")
		assert.Contains(t, result, "value")
	})

	t.Run("returns empty for empty map", func(t *testing.T) {
		m := map[string]string{}
		result := Mapf(m)
		assert.Empty(t, result)
	})
}

func TestHomePath(t *testing.T) {
	tests := map[string]struct {
		path     string
		contains string
	}{
		"non-home path is unchanged": {
			path:     "/tmp/some/path",
			contains: "/tmp/some/path",
		},
		"empty path is unchanged": {
			path:     "",
			contains: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := HomePath(tc.path)
			assert.Contains(t, result, tc.contains)
		})
	}
}

func TestTeamSelectLabel(t *testing.T) {
	tests := map[string]struct {
		teamDomain string
		teamID     string
	}{
		"formats team domain and ID": {
			teamDomain: "my-workspace",
			teamID:     "T12345",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := TeamSelectLabel(tc.teamDomain, tc.teamID)
			assert.Contains(t, result, tc.teamDomain)
			assert.Contains(t, result, tc.teamID)
		})
	}
}

func TestTimeAgo(t *testing.T) {
	now := int(time.Now().Unix())
	tests := map[string]struct {
		datetime int
		contains string
	}{
		"seconds ago": {
			datetime: now - 30,
			contains: "seconds ago",
		},
		"minutes ago": {
			datetime: now - 120,
			contains: "minutes ago",
		},
		"hours ago": {
			datetime: now - 7200,
			contains: "hours ago",
		},
		"days ago": {
			datetime: now - 86400*3,
			contains: "days ago",
		},
		"weeks ago": {
			datetime: now - 86400*14,
			contains: "weeks ago",
		},
		"months ago": {
			datetime: now - 86400*60,
			contains: "months ago",
		},
		"years ago": {
			datetime: now - 86400*800,
			contains: "years ago",
		},
		"future time": {
			datetime: now + 3600,
			contains: "until",
		},
		"singular minute": {
			datetime: now - 90,
			contains: "minute ago",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := TimeAgo(tc.datetime)
			assert.Contains(t, result, tc.contains)
		})
	}
}

/*
* App name formatting
 */

func TestLocalRunDisplayNamePlain(t *testing.T) {
	tests := map[string]struct {
		mockAppName     string
		expectedAppName string
	}{
		"the local tag is appended to a name": {
			mockAppName:     "bizz",
			expectedAppName: "bizz (local)",
		},
		"blank names still receive the tagging": {
			mockAppName:     "",
			expectedAppName: " (local)",
		},
		"the local tag is not appended to a name that already has it": {
			mockAppName:     "bizz (local)",
			expectedAppName: "bizz (local)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualAppName := LocalRunDisplayName(tc.mockAppName)
			assert.Equal(t, tc.expectedAppName, actualAppName)
		})
	}
}
