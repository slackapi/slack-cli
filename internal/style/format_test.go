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

	"github.com/AlecAivazis/survey/v2/core"
	"github.com/stretchr/testify/assert"
)

func TestGetKeyLength(t *testing.T) {
	tests := map[string]struct {
		keys     map[string]string
		expected int
	}{
		"empty key has zero length": {
			keys:     map[string]string{"": "the zero key"},
			expected: 0,
		},
		"equal length keys return that length": {
			keys:     map[string]string{"key1": "unlocks the building", "key2": "unlocks the room"},
			expected: 4,
		},
		"returns length of longest key": {
			keys:     map[string]string{"longer_key1": "locks the building", "very_long_key2": "locks the room"},
			expected: 14,
		},
		"longest key is first": {
			keys:     map[string]string{"longest_key1": "short value", "short_key2": "longer value"},
			expected: 12,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getKeyLength(tc.keys))
		})
	}
}

func TestSectionf(t *testing.T) {
	tests := map[string]struct {
		section  TextSection
		expected string
	}{
		"empty text returns empty string": {
			section:  TextSection{Emoji: "", Text: "", Secondary: []string{}},
			expected: "",
		},
		"header with emoji and secondary text": {
			section:  TextSection{Emoji: "tada", Text: "Congrats", Secondary: []string{"You did it"}},
			expected: Emoji("tada") + "Congrats\n" + Indent(Secondary("You did it")) + "\n",
		},
		"no emoji starts text immediately": {
			section:  TextSection{Emoji: "", Text: "On the left. Where I like it.", Secondary: []string{}},
			expected: "On the left. Where I like it.\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Sectionf(tc.section))
		})
	}
}

func TestSectionHeaderfEmpty(t *testing.T) {
	assert.Equal(t, "", SectionHeaderf("tada", ""))
}

func TestSectionSecondaryf(t *testing.T) {
	tests := map[string]struct {
		format   string
		args     []interface{}
		validate func(t *testing.T, result string)
	}{
		"empty input returns empty string": {
			format: "%s",
			args:   []interface{}{""},
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "", result)
			},
		},
		"plain text is preserved and indented": {
			format: "%s",
			args:   []interface{}{"If you have a moment, go grab a glass of water!"},
			validate: func(t *testing.T, result string) {
				text := "If you have a moment, go grab a glass of water!"
				assert.Contains(t, result, text)
				assert.Equal(t, Indent(Secondary(text))+"\n", result)
			},
		},
		"formats input variables": {
			format: "App ID: %s\tStatus: %s",
			args:   []interface{}{"A123456", "Installed"},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "App ID: A123456\tStatus: Installed")
			},
		},
		"multi-line input is properly indented": {
			format: "%s",
			args:   []interface{}{"L1\nL2\nL3"},
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				for i, line := range strings.Split("L1\nL2\nL3", "\n") {
					assert.Equal(t, Indent(Secondary(line)), lines[i])
				}
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := SectionSecondaryf(tc.format, tc.args...)
			tc.validate(t, result)
		})
	}
}

func TestCommandf(t *testing.T) {
	tests := map[string]struct {
		process   string
		command   string
		isPrimary bool
	}{
		"primary command contains process and command": {
			process:   "renamed-slack-command",
			command:   "feedback",
			isPrimary: true,
		},
		"secondary command contains process and command": {
			process:   "a-renamed-slack-cli",
			command:   "feedback",
			isPrimary: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			processTemp := os.Args[0]
			os.Args[0] = tc.process
			defer func() { os.Args[0] = processTemp }()

			formatted := Commandf(tc.command, tc.isPrimary)
			assert.Contains(t, formatted, tc.process+" "+tc.command)
		})
	}
}

func TestIndent(t *testing.T) {
	text := "a few spaces are expected at the start of this line, but no other changes"
	indented := Indent(text)
	assert.Contains(t, indented, text)
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

func TestSurveyIcons(t *testing.T) {
	tests := map[string]struct {
		styleEnabled bool
	}{
		"styles are not enabled": {
			styleEnabled: false,
		},
		"styles are enabled": {
			styleEnabled: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			core.DisableColor = false
			isStyleEnabled = tc.styleEnabled

			_ = SurveyIcons()
			assert.NotEqual(t, tc.styleEnabled, core.DisableColor)
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
			ToggleCharm(tc.charmEnabled)
			defer func() {
				ToggleStyles(false)
				ToggleCharm(false)
			}()
			actual := StyleFlags(tc.input)
			assert.Equal(t, tc.expectedFunc(), actual)
		})
	}
}

func Test_ExampleCommandsf(t *testing.T) {
	tests := map[string]struct {
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
		ToggleCharm(false)
	}()
	ToggleStyles(true)
	ToggleCharm(true)

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
