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

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Icons_MarshalJSON(t *testing.T) {
	tests := map[string]struct {
		icons             *Icons
		expectedErrorType error
		expectedBlobs     []string
	}{
		"Marshal no icons": {
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedBlobs:     []string{},
		},
		"Marshal 1 icon": {
			icons:             &Icons{"image_96": "path/to/image_96.png"},
			expectedErrorType: nil,
			expectedBlobs: []string{
				`"image_96":"path/to/image_96.png"`,
			},
		},
		"Marshal 2 icons": {
			icons:             &Icons{"image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png"},
			expectedErrorType: nil,
			expectedBlobs: []string{
				`"image_96":"path/to/image_96.png"`,
				`"image_192":"path/to/image_192.png"`,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedBlob, err := json.Marshal(tc.icons)

			require.IsType(t, err, tc.expectedErrorType)
			for _, expectedBlob := range tc.expectedBlobs {
				require.Contains(t, string(returnedBlob), expectedBlob)
			}
		})
	}
}

func Test_Icons_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		blob              string
		icons             *Icons
		expectedErrorType error
		expectedIcons     *Icons
	}{
		"JSON unmarshal error": {
			blob:              `{ "image_96": 100 }`, // expects type to be string not int
			icons:             &Icons{},
			expectedErrorType: &json.UnmarshalTypeError{},
			expectedIcons:     &Icons{},
		},
		"image_96 and image_192 do not exist": {
			blob:              `{ "cat": "meow", "dog": "bark" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{},
		},
		"image_96 exists": {
			blob:              `{ "image_96": "path/to/image_96.png" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png"},
		},
		"image_192 exists": {
			blob:              `{ "image_192": "path/to/image_192.png" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_192": "path/to/image_192.png"},
		},
		"image_96 and image_192 exist": {
			blob:              `{ "image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png"},
		},
		"image_96 exists and unsupported properties exists": {
			blob:              `{ "image_96": "path/to/image_96.png", "foo": "bar" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png", "foo": "bar"},
		},
		"Icons is nil, image_96 and image_192 exist": {
			blob:              `{ "image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png" }`,
			icons:             nil,
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tc.blob), &tc.icons)

			require.IsType(t, err, tc.expectedErrorType)
			require.Equal(t, tc.expectedIcons, tc.icons)
		})
	}
}
