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

package goutils

import "testing"

func Test_Contains(t *testing.T) {
	tests := map[string]struct {
		listToCheck     []string
		toFind          string
		isCaseSensitive bool
		want            bool
	}{
		"not_case_sensitive_success": {
			listToCheck:     []string{"hi", "hey"},
			toFind:          "Hey",
			isCaseSensitive: false,
			want:            true,
		},
		"case_sensitive_success": {
			listToCheck:     []string{"hi", "Hey"},
			toFind:          "Hey",
			isCaseSensitive: true,
			want:            true,
		},
		"case_sensitive_fail": {
			listToCheck:     []string{"hi", "hey", "hello", "apple", "pear"},
			toFind:          "Hey",
			isCaseSensitive: true,
			want:            false,
		},
		"not_case_sensitive_fail": {
			listToCheck:     []string{"hi", "hey", "hello", "apple", "pear"},
			toFind:          "Peach",
			isCaseSensitive: false,
			want:            false,
		},
		"not_case_sensitive_substring": {
			listToCheck:     []string{"hi", "hey hello"},
			toFind:          "hey",
			isCaseSensitive: false,
			want:            false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Contains(tc.listToCheck, tc.toFind, tc.isCaseSensitive); got != tc.want {
				t.Errorf("method() = %v, want %v", got, tc.want)
			}
		})
	}
}
