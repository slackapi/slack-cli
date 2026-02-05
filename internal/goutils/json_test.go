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

import (
	"encoding/json"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mergeJSON(t *testing.T) {
	type testConfig struct {
		One   string `json:"one,omitempty"`
		Two   string `json:"two,omitempty"`
		Three string `json:"three,omitempty"`
	}

	for name, tc := range map[string]struct {
		defaultJSON   string
		customJSON    string
		expectedError error
		expectedJSON  string
	}{
		"default does not exist": {
			defaultJSON:   ``,
			customJSON:    `{}`,
			expectedError: nil,
			expectedJSON:  `{}`,
		},
		"custom does not exist": {
			defaultJSON:   `{}`,
			customJSON:    ``,
			expectedError: nil,
			expectedJSON:  `{}`,
		},
		"default has no properties": {
			defaultJSON:   `{}`,
			customJSON:    `{"one":"1-custom","two":"2-custom"}`,
			expectedError: nil,
			expectedJSON:  `{"one":"1-custom","two":"2-custom"}`,
		},
		"inherit default properties": {
			defaultJSON:   `{"three":"3-default"}`,
			customJSON:    `{"one":"1-custom","two":"2-custom"}`,
			expectedError: nil,
			expectedJSON:  `{"one":"1-custom","two":"2-custom","three":"3-default"}`,
		},
		"override default properties": {
			defaultJSON:   `{"one":"1-default"}`,
			customJSON:    `{"one":"1-custom"}`,
			expectedError: nil,
			expectedJSON:  `{"one":"1-custom"}`,
		},
		"override default properties and add values": {
			defaultJSON:   `{"one":"1-default"}`,
			customJSON:    `{"one":"1-custom","two":"2-custom"}`,
			expectedError: nil,
			expectedJSON:  `{"one":"1-custom","two":"2-custom"}`,
		},
		"overrides default properties, adds values, and inherits default values": {
			defaultJSON:   `{"one":"1-default","two":"2-default"}`,
			customJSON:    `{"one":"1-custom","three":"3-custom"}`,
			expectedError: nil,
			expectedJSON:  `{"one":"1-custom","two":"2-default","three":"3-custom"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var actualConfig testConfig
			actualError := MergeJSON(tc.defaultJSON, tc.customJSON, &actualConfig)
			require.Equal(t, tc.expectedError, actualError)

			b, err := json.Marshal(actualConfig)
			require.Nil(t, err)

			var actualConfigJSON = string(b)
			require.Equal(t, tc.expectedJSON, actualConfigJSON)
		})
	}
}

func Test_JSONMarshalUnescaped(t *testing.T) {
	type TestInput struct {
		Data    string
		Numbers []int
	}
	type TestStruct struct {
		Input    TestInput
		Expected string
	}

	for name, tc := range map[string]TestStruct{
		"encode a struct with a slice and escapable data string": {
			Input: TestInput{
				Data:    ":num < 3",
				Numbers: []int{1, 2, 3, 4, 6},
			},
			Expected: "{\"Data\":\":num < 3\",\"Numbers\":[1,2,3,4,6]}\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			buff, err := JSONMarshalUnescaped(tc.Input)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.Expected, buff)
			}
		})
	}
}

func Test_JSONMarshalUnescapedIndent(t *testing.T) {
	type TestInput struct {
		Data    string
		Numbers []int
	}
	type TestStruct struct {
		Input    TestInput
		Expected string
	}

	for name, tc := range map[string]TestStruct{
		"format a struct with a slice and escapable data string": {
			Input: TestInput{
				Data:    ":num < 3",
				Numbers: []int{1, 2, 3, 4, 6},
			},
			Expected: `{
  "Data": ":num < 3",
  "Numbers": [
    1,
    2,
    3,
    4,
    6
  ]
}
`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			buff, err := JSONMarshalUnescapedIndent(tc.Input)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.Expected, buff)
			}
		})
	}
}

func Test_UnmarshalJSON(t *testing.T) {
	type testConfig struct {
		One string `json:"one,omitempty"`
	}

	for name, tc := range map[string]struct {
		data          string
		expectedError error
		expectAll     []string
	}{
		"valid json": {
			data:          `{"one": "1"}`,
			expectedError: nil,
		},
		"invalid json": {
			data:      `}{`,
			expectAll: []string{slackerror.ErrUnableToParseJSON},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var v testConfig
			err := JSONUnmarshal([]byte(tc.data), &v)
			if tc.expectedError == nil && len(tc.expectAll) == 0 {
				require.Nil(t, err)
			} else {
				require.Contains(t, err.Error(), slackerror.ErrUnableToParseJSON)
				for _, s := range tc.expectAll {
					require.Contains(t, err.Error(), s)
				}
			}
		})
	}
}
