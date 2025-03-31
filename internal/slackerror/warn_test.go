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

package slackerror

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/style"
	"github.com/stretchr/testify/assert"
)

func Test_Warning(t *testing.T) {
	tests := map[string]struct {
		message  string
		warnings Warnings
		verbose  bool
		expected string
	}{
		"formats nothing if nothing is provided": {
			warnings: []Warning{},
		},
		"formats the message if only a message without details": {
			message:  "something strange happened",
			warnings: []Warning{},
			expected: "something strange happened",
		},
		"formats the message with provided warning details": {
			message: "something strange happened",
			warnings: []Warning{
				{
					Code:    "warn_something",
					Message: "something is not allowed",
				},
			},
			expected: "something strange happened\n\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is not allowed (warn_something)",
				}),
		},
		"formats the message without a blank additional details": {
			message: "something strange happened",
			warnings: []Warning{
				{
					Code:        "warn_something",
					Message:     "something is not allowed",
					Pointer:     " ",
					Remediation: "  \n",
				},
			},
			expected: "something strange happened\n\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is not allowed (warn_something)",
				}),
		},
		"formats details about the warning with a message": {
			message: "something strange happened",
			warnings: []Warning{
				{
					Code:        "warn_something",
					Message:     "something is not allowed",
					Pointer:     "this.example.function",
					Remediation: "change that one thing",
				},
			},
			expected: "something strange happened\n\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is not allowed (warn_something)",
					Secondary: []string{
						"Source: this.example.function",
						"Suggestion: change that one thing",
					},
				}),
		},
		"orders groups of identical errors with unique sources": {
			message: "something strange happened",
			warnings: []Warning{
				{
					Code:    "warn_details",
					Message: "another example here",
				},
				{
					Code:    "warn_details",
					Message: "a changed message",
					Pointer: "included after above",
				},
				{
					Code:        "warn_something",
					Message:     "something is not allowed",
					Pointer:     "this.example.function.0",
					Remediation: "change that one thing",
				},
				{
					Code:        "warn_something",
					Message:     "something is not allowed",
					Pointer:     "this.example.function.1",
					Remediation: "change that one thing",
				},
				{
					Code:        "warn_something",
					Message:     "something else is not allowed here",
					Pointer:     "this.example.function.2",
					Remediation: "change that one thing",
				},
				{
					Code:    "warn_something",
					Message: "something is missing details",
				},
				{
					Code:    "warn_nothing",
					Message: "nothing else in this example",
				},
			},
			expected: "something strange happened\n\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "another example here (warn_details)",
				}) +
				"\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "a changed message (warn_details)",
					Secondary: []string{
						"Source: included after above",
					},
				}) +
				"\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "nothing else in this example (warn_nothing)",
				}) +
				"\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is missing details (warn_something)",
				}) +
				"\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is not allowed (warn_something)",
					Secondary: []string{
						"Source: this.example.function.0",
						"Source: this.example.function.1",
						"Suggestion: change that one thing",
					},
				}) +
				"\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something else is not allowed here (warn_something)",
					Secondary: []string{
						"Source: this.example.function.2",
						"Suggestion: change that one thing",
					},
				}),
		},
		"multiple of the same errors are hidden without verbose settings": {
			message: "something strange happened",
			verbose: false,
			warnings: []Warning{
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.0",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.1",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.2",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.3",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.4",
				},
			},
			expected: "something strange happened\n\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is not allowed (warn_something)",
					Secondary: []string{
						"Source: this.example.function.0",
						"Source: this.example.function.1",
						"Similar warnings from '3' other sources can be revealed with --verbose",
					},
				}),
		},
		"multiple of the same errors are shown with verbose settings": {
			message: "something strange happened",
			verbose: true,
			warnings: []Warning{
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.0",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.1",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.2",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.3",
				},
				{
					Code:    "warn_something",
					Message: "something is not allowed",
					Pointer: "this.example.function.4",
				},
			},
			expected: "something strange happened\n\n" +
				style.Sectionf(style.TextSection{
					Emoji: "memo",
					Text:  "something is not allowed (warn_something)",
					Secondary: []string{
						"Source: this.example.function.0",
						"Source: this.example.function.1",
						"Source: this.example.function.2",
						"Source: this.example.function.3",
						"Source: this.example.function.4",
					},
				}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tt.warnings.Warning(tt.verbose, tt.message)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
