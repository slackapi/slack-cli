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

package config

import "encoding/json"

// unmarshalExperimentsField handles backwards-compatible unmarshaling of the
// experiments field. It accepts both the old array format (["charm", "sandboxes"])
// and the new map format ({"charm": true, "sandboxes": true}).
func unmarshalExperimentsField(raw json.RawMessage) map[string]bool {
	if len(raw) == 0 {
		return nil
	}

	// Try new map format first
	var mapFormat map[string]bool
	if err := json.Unmarshal(raw, &mapFormat); err == nil {
		return mapFormat
	}

	// Fall back to old array format
	var arrayFormat []string
	if err := json.Unmarshal(raw, &arrayFormat); err == nil {
		if len(arrayFormat) == 0 {
			return nil
		}
		result := make(map[string]bool, len(arrayFormat))
		for _, exp := range arrayFormat {
			result[exp] = true
		}
		return result
	}

	return nil
}
