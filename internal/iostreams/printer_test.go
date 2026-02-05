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

package iostreams

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PrintDebug(t *testing.T) {
	tests := map[string]struct {
		format    string
		arguments []any
		expected  []string
	}{
		"prints a formatted debug to stdout": {
			format:    "hello %s - noon is %d",
			arguments: []any{"world", 12},
			expected: []string{
				"hello world - noon is 12",
			},
		},
		"prints a multiline debug to stdout": {
			format: "something\nstrange\nhappened",
			expected: []string{
				"something",
				"strange",
				"happened",
			},
		},
		"prints unformatted debug to stdout": {
			format: "something strange happened",
			expected: []string{
				"something strange happened",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			config := config.NewConfig(fsMock, osMock)
			config.DebugEnabled = true
			io := NewIOStreams(config, fsMock, osMock)
			stdoutBuffer := bytes.Buffer{}
			stdoutLogger := log.Logger{}
			stdoutLogger.SetOutput(&stdoutBuffer)
			io.Stdout = &stdoutLogger
			io.PrintDebug(ctx, tc.format, tc.arguments...)

			// Assert output lines match the pattern: [YYYY-MM-DD HH:MM:SS] line
			// With a trailing blank newline at the end.
			actual := strings.Split(stdoutBuffer.String(), "\n")
			require.Len(
				t,
				actual,
				len(tc.expected)+1,
				"an unexpected amount of lines were output!\nactual:\n%s\nexpected:\n%s\n",
				strings.Join(actual, "\n"),
				strings.Join(tc.expected, "\n"),
			)
			for ii, line := range tc.expected {
				pattern := regexp.MustCompile(
					fmt.Sprintf(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] %s`, line),
				)
				assert.True(
					t,
					pattern.MatchString(actual[ii]),
					"outputs do not match!\nactual:\n%s\nexpected:\n%s\n",
					actual[ii],
					tc.expected[ii],
				)
			}
		})
	}
}

func Test_PrintWarning(t *testing.T) {
	tests := map[string]struct {
		format    string
		arguments []any
		expected  string
	}{
		"prints a formatted warning to stderr": {
			format:    "hello %s - noon is %d",
			arguments: []any{"world", 12},
			expected: strings.Join([]string{
				fmt.Sprintf("Check /Users/user.name/.slack/logs/%s for error logs", filename),
				"",
				"hello world - noon is 12",
				"",
			}, "\n"),
		},
		"prints unformatted warning to stderr": {
			format: "something strange happened",
			expected: strings.Join([]string{
				fmt.Sprintf("Check /Users/user.name/.slack/logs/%s for error logs", filename),
				"",
				"something strange happened",
				"",
			}, "\n"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			config := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(config, fsMock, osMock)
			stderrBuffer := bytes.Buffer{}
			stderrLogger := log.Logger{}
			stderrLogger.SetOutput(&stderrBuffer)
			io.Stderr = &stderrLogger
			io.PrintWarning(ctx, tc.format, tc.arguments...)
			assert.Equal(t, tc.expected, stderrBuffer.String())
		})
	}
}

func Test_IOStreams_PrintTrace(t *testing.T) {
	tests := map[string]struct {
		traceID        string
		traceValues    []string
		traceEnabled   bool
		expectedOutput string
	}{
		"Trace Disabled": {
			traceID:        "TRACE_ID_1",
			traceValues:    []string{},
			traceEnabled:   false,
			expectedOutput: "",
		},
		"Trace ID": {
			traceID:        "TRACE_ID_1",
			traceValues:    []string{},
			traceEnabled:   true,
			expectedOutput: "TRACE_ID_1\n",
		},
		"Trace ID and one value": {
			traceID:        "TRACE_ID_1",
			traceValues:    []string{"VALUE_1"},
			traceEnabled:   true,
			expectedOutput: "TRACE_ID_1=VALUE_1\n",
		},
		"Trace ID and many values": {
			traceID:        "TRACE_ID_1",
			traceValues:    []string{"VALUE_1", "VALUE_2"},
			traceEnabled:   true,
			expectedOutput: "TRACE_ID_1=VALUE_1,VALUE_2\n",
		},
		"Trim whitespace": {
			traceID:        "  TRACE_ID_1   ",
			traceValues:    []string{"  VALUE_1    ", "  VALUE_2   "},
			traceEnabled:   true,
			expectedOutput: "TRACE_ID_1=VALUE_1,VALUE_2\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			var fsMock = slackdeps.NewFsMock()
			var osMock = slackdeps.NewOsMock()

			var config = config.NewConfig(fsMock, osMock)
			config.SlackTestTraceFlag = tc.traceEnabled
			var io = NewIOStreams(config, fsMock, osMock)

			var stdoutBuffer = bytes.Buffer{}
			var stdoutLogger = log.Logger{}
			stdoutLogger.SetOutput(&stdoutBuffer)
			io.Stdout = &stdoutLogger

			// Execute
			io.PrintTrace(ctx, tc.traceID, tc.traceValues...)

			// Read output
			var actualOutput string
			if stdout, ok := io.WriteOut().(*bytes.Buffer); ok {
				actualOutput = stdout.String()
			}

			// Assert
			assert.Equal(t, tc.expectedOutput, actualOutput)
		})
	}
}

func Test_IOStreams_SprintF(t *testing.T) {
	tests := map[string]struct {
		format   string
		args     []any
		expected string
	}{
		"the standard string is unchanged without args": {
			format:   "standard",
			expected: "standard",
		},
		"the formatted string is unchanged without args": {
			format:   "formatte%d",
			expected: "formatte%d",
		},
		"the formatted string is changed with args": {
			format:   "%s: %d",
			args:     []any{"number", 12},
			expected: "number: 12",
		},
		"the missing argument is noticed with args": {
			format:   "%s: %d",
			args:     []any{"number"},
			expected: "number: %!d(MISSING)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := sprintF(tc.format, tc.args...)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
