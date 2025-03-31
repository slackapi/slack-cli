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
	"fmt"
	"runtime"
	"strings"
)

// MapToStringSlice converts a map[string]string to a slice of strings with
// elements key="value" from the map with an optional prefix
func MapToStringSlice(args map[string]string, prefix string) []string {
	var res = []string{}
	for name, value := range args {
		if len(value) > 0 {
			var escapedValue string
			if runtime.GOOS == "windows" {
				escapedValue = strings.ReplaceAll(value, `"`, "`\"")
			} else {
				escapedValue = strings.ReplaceAll(value, `"`, `\"`)
			}
			res = append(res, fmt.Sprintf(`%s%s="%s"`, prefix, name, escapedValue))
		}
	}
	return res
}

// MapToStringSliceKeys takes a map[string]interface{} and returns its keys as a
// slice of strings
func MapToStringSliceKeys(args map[string]struct{}) []string {
	var keys = []string{}
	for key := range args {
		keys = append(keys, key)
	}
	return keys
}
