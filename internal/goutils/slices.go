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

import "strings"

// AppendStringIfNotMember only appends if not already a member
func AppendStringIfNotMember(list []string, item string) []string {
	for _, member := range list {
		if member == item {
			return list
		}
	}
	return append(list, item)
}

// Contains returns true if a string can be found in a list of strings, false otherwise
// isCaseSensitive defines whether the string needs to match capitalization to return true
// case sensitivity based on Unicode case-folding, more general form of case-insensitivity
func Contains(listToCheck []string, toFind string, isCaseSensitive bool) bool {
	for _, s := range listToCheck {
		if isCaseSensitive {
			if s == toFind {
				return true
			}
		} else {
			if strings.EqualFold(s, toFind) {
				return true
			}
		}
	}
	return false
}
