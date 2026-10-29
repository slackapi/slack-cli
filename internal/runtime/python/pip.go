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

package python

import (
	"regexp"
	"strings"
)

// packageNameRegex matches text shaped like a Python package name
//
// https://packaging.python.org/en/latest/specifications/name-normalization/#name-format
var packageNameRegex = regexp.MustCompile(`(?i)[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?`)

// packageNameSeparatorRegex matches runs of "-", "_", and "." which are equivalent in Python package names
var packageNameSeparatorRegex = regexp.MustCompile(`[-_.]+`)

// normalizePackageName lowercases a Python package name and replaces runs of separators with "-"
//
// https://packaging.python.org/en/latest/specifications/name-normalization/#name-normalization
func normalizePackageName(name string) string {
	return packageNameSeparatorRegex.ReplaceAllString(strings.ToLower(name), "-")
}

// containsPackage returns if the file data names the package with any spelling
// that normalizes to the same name, such as slack_cli_hooks for slack-cli-hooks
func containsPackage(fileData string, pkgName string) bool {
	pkgName = normalizePackageName(pkgName)
	for _, name := range packageNameRegex.FindAllString(fileData, -1) {
		if normalizePackageName(name) == pkgName {
			return true
		}
	}
	return false
}
