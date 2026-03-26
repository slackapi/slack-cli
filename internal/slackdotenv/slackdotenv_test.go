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

package slackdotenv

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_Read(t *testing.T) {
	tests := map[string]struct {
		fs          afero.Fs
		dotenv      string
		writeDotenv bool
		expected    map[string]string
		expectErr   bool
	}{
		"returns nil when fs is nil": {
			fs:       nil,
			expected: nil,
		},
		"returns nil when .env file does not exist": {
			fs:       afero.NewMemMapFs(),
			expected: nil,
		},
		"returns empty map for empty .env file": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "",
			writeDotenv: true,
			expected:    map[string]string{},
		},
		"parses single variable": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "FOO=bar\n",
			writeDotenv: true,
			expected:    map[string]string{"FOO": "bar"},
		},
		"parses multiple variables": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "FOO=bar\nBAZ=qux\n",
			writeDotenv: true,
			expected:    map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		"parses quoted values": {
			fs:          afero.NewMemMapFs(),
			dotenv:      `TOKEN="my secret token"` + "\n",
			writeDotenv: true,
			expected:    map[string]string{"TOKEN": "my secret token"},
		},
		"skips comment lines": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "# this is a comment\nFOO=bar\n",
			writeDotenv: true,
			expected:    map[string]string{"FOO": "bar"},
		},
		"handles values with equals signs": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "URL=https://example.com?foo=bar&baz=qux\n",
			writeDotenv: true,
			expected:    map[string]string{"URL": "https://example.com?foo=bar&baz=qux"},
		},
		"handles empty values": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "EMPTY=\n",
			writeDotenv: true,
			expected:    map[string]string{"EMPTY": ""},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.writeDotenv && tc.fs != nil {
				_ = afero.WriteFile(tc.fs, ".env", []byte(tc.dotenv), 0644)
			}
			result, err := Read(tc.fs)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}
