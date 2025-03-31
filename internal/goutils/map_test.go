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

	"github.com/stretchr/testify/assert"
)

func Test_MapToStringSlice(t *testing.T) {
	tests := map[string]struct {
		inputMap      map[string]string
		prefix        string
		expectedSlice []string
	}{
		"no prefix": {
			inputMap: map[string]string{
				"color": "purple",
				"size":  "L",
			},
			prefix:        "",
			expectedSlice: []string{`color="purple"`, `size="L"`},
		},
		"with prefix": {
			inputMap: map[string]string{
				"color": "purple",
				"size":  "L",
			},
			prefix:        "--",
			expectedSlice: []string{`--color="purple"`, `--size="L"`},
		},
		"with double quotes": {
			inputMap: map[string]string{
				"quotes": `some "values" are important`,
			},
			prefix:        "--",
			expectedSlice: []string{`--quotes="some \"values\" are important"`},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actualSlice := MapToStringSlice(tt.inputMap, tt.prefix)
			assert.ElementsMatch(t, tt.expectedSlice, actualSlice)
		})
	}
}

func Test_MapToStringSliceKeys(t *testing.T) {
	var inputMap = map[string]struct{}{
		"color": {},
		"size":  {},
	}
	var expectedSlice = []string{"color", "size"}

	t.Run("Returns all keys in the map", func(t *testing.T) {
		actualSlice := MapToStringSliceKeys(inputMap)
		assert.ElementsMatch(t, expectedSlice, actualSlice)
	})
}
