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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Flatten converts a manifest (as JSON-serializable struct) into a flat map
// where keys are dot-notation paths and values are the leaf values.
// Arrays are treated as leaf values (not recursed into individually) because
// array element identity is ambiguous without a key field.
//
// Keys that contain literal dots (e.g. function IDs like "slack.users.lookup")
// have those dots backslash-escaped in the path so flatten/unflatten round-
// trip cleanly. splitPath honors the same escape sequence.
func Flatten(manifest any) (map[string]any, error) {
	data, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}
	result := make(map[string]any)
	flattenRecursive("", raw, result)
	return result, nil
}

func flattenRecursive(prefix string, data map[string]any, result map[string]any) {
	for key, value := range data {
		fullKey := escapePathSegment(key)
		if prefix != "" {
			fullKey = prefix + "." + fullKey
		}
		switch v := value.(type) {
		case map[string]any:
			flattenRecursive(fullKey, v, result)
		default:
			result[fullKey] = value
		}
	}
}

// escapePathSegment backslash-escapes the path delimiter and the escape
// character so a key containing a literal dot survives flatten/splitPath.
func escapePathSegment(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `.`, `\.`)
	return s
}

// SortedKeys returns the keys of a flat map in sorted order for deterministic output.
func SortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
