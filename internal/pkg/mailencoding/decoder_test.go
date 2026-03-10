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

package mailencoding

import (
	"testing"
)

func TestDecodeRFC2047(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "日本語UTF-8 Base64",
			input:    "=?UTF-8?B?44GE44KT44G744Gm44GE44KT?=",
			expected: "日本語日本語",
		},
		{
			name:     "プレーンテキスト",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "複合テキスト",
			input:    "Re: =?UTF-8?B?44GE44KT44G744Gm?= job opportunity",
			expected: "Re: 日本語 job opportunity",
		},
		{
			name:     "シフトJIS",
			input:    "=?SHIFT_JIS?B?k/qWe5f6?=",
			expected: "日本",
		},
		{
			name:     "デコード不要のテキスト",
			input:    "Subject: Regular email",
			expected: "Subject: Regular email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeRFC2047(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeRFC2047(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
