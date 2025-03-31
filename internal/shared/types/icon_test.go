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

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Icons_MarshalJSON(t *testing.T) {
	tests := []struct {
		name              string
		icons             *Icons
		expectedErrorType error
		expectedBlobs     []string
	}{
		{
			name:              "Marshal no icons",
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedBlobs:     []string{},
		},
		{
			name:              "Marshal 1 icon",
			icons:             &Icons{"image_96": "path/to/image_96.png"},
			expectedErrorType: nil,
			expectedBlobs: []string{
				`"image_96":"path/to/image_96.png"`,
			},
		},
		{
			name:              "Marshal 2 icons",
			icons:             &Icons{"image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png"},
			expectedErrorType: nil,
			expectedBlobs: []string{
				`"image_96":"path/to/image_96.png"`,
				`"image_192":"path/to/image_192.png"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedBlob, err := json.Marshal(tt.icons)

			require.IsType(t, err, tt.expectedErrorType)
			for _, expectedBlob := range tt.expectedBlobs {
				require.Contains(t, string(returnedBlob), expectedBlob)
			}
		})
	}
}

func Test_Icons_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name              string
		blob              string
		icons             *Icons
		expectedErrorType error
		expectedIcons     *Icons
	}{
		{
			name:              "JSON unmarshal error",
			blob:              `{ "image_96": 100 }`, // expects type to be string not int
			icons:             &Icons{},
			expectedErrorType: &json.UnmarshalTypeError{},
			expectedIcons:     &Icons{},
		},
		{
			name:              "image_96 and image_192 do not exist",
			blob:              `{ "cat": "meow", "dog": "bark" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{},
		},
		{
			name:              "image_96 exists",
			blob:              `{ "image_96": "path/to/image_96.png" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png"},
		},
		{
			name:              "image_192 exists",
			blob:              `{ "image_192": "path/to/image_192.png" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_192": "path/to/image_192.png"},
		},
		{
			name:              "image_96 and image_192 exist",
			blob:              `{ "image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png"},
		},
		{
			name:              "image_96 exists and unsupported properties exists",
			blob:              `{ "image_96": "path/to/image_96.png", "foo": "bar" }`,
			icons:             &Icons{},
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png", "foo": "bar"},
		},
		{
			name:              "Icons is nil, image_96 and image_192 exist",
			blob:              `{ "image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png" }`,
			icons:             nil,
			expectedErrorType: nil,
			expectedIcons:     &Icons{"image_96": "path/to/image_96.png", "image_192": "path/to/image_192.png"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.blob), &tt.icons)

			require.IsType(t, err, tt.expectedErrorType)
			require.Equal(t, tt.expectedIcons, tt.icons)
		})
	}
}
