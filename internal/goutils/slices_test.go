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

package goutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AppendStringIfNotMember(t *testing.T) {
	tests := []struct {
		name          string
		newElement    string
		originalSlice []string
		expectedSlice []string
	}{
		{
			name:          "Append new element",
			newElement:    "four",
			originalSlice: []string{"one", "two", "three"},
			expectedSlice: []string{"one", "two", "three", "four"},
		},
		{
			name:          "Do not append existing element",
			newElement:    "two",
			originalSlice: []string{"one", "two", "three"},
			expectedSlice: []string{"one", "two", "three"},
		},
		{
			name:          "Append new element to empty list",
			newElement:    "one",
			originalSlice: []string{},
			expectedSlice: []string{"one"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualSlice := AppendStringIfNotMember(tt.originalSlice, tt.newElement)
			require.ElementsMatch(t, tt.expectedSlice, actualSlice)
		})
	}
}

func Test_Contains(t *testing.T) {
	tests := []struct {
		name            string
		listToCheck     []string
		toFind          string
		isCaseSensitive bool
		want            bool
	}{
		{name: "not_case_sensitive_success", listToCheck: []string{"hi", "hey"}, toFind: "Hey", isCaseSensitive: false, want: true},
		{name: "case_sensitive_success", listToCheck: []string{"hi", "Hey"}, toFind: "Hey", isCaseSensitive: true, want: true},
		{name: "case_sensitive_fail", listToCheck: []string{"hi", "hey", "hello", "apple", "pear"}, toFind: "Hey", isCaseSensitive: true, want: false},
		{name: "not_case_sensitive_fail", listToCheck: []string{"hi", "hey", "hello", "apple", "pear"}, toFind: "Peach", isCaseSensitive: false, want: false},
		{name: "not_case_sensitive_substring", listToCheck: []string{"hi", "hey hello"}, toFind: "hey", isCaseSensitive: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.listToCheck, tt.toFind, tt.isCaseSensitive); got != tt.want {
				t.Errorf("method() = %v, want %v", got, tt.want)
			}
		})
	}
}
