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

package manifest

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func Test_truncateRunes(t *testing.T) {
	tests := map[string]struct {
		input    string
		max      int
		expected string
	}{
		"shorter than max returns unchanged": {
			input:    "hello",
			max:      80,
			expected: "hello",
		},
		"exactly max runes returns unchanged": {
			input:    "abcdefghij",
			max:      10,
			expected: "abcdefghij",
		},
		"longer than max truncates with ellipsis": {
			input:    "abcdefghijklmno",
			max:      10,
			expected: "abcdefg...",
		},
		"multi-byte runes are not cut mid-character": {
			// Each emoji is 4 bytes in UTF-8 but one rune. Byte-based
			// slicing would split the middle emoji.
			input:    "🐶🐱🐭🐹🐰🦊🐻🐼🐨🐯🦁🐮",
			max:      6,
			expected: "🐶🐱🐭...",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := truncateRunes(tc.input, tc.max)
			assert.Equal(t, tc.expected, got)
			// In every case the result must remain valid UTF-8.
			assert.True(t, utf8.ValidString(got), "result is not valid UTF-8: %q", got)
		})
	}
}
