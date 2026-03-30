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

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Read(t *testing.T) {
	tests := map[string]struct {
		fs          afero.Fs
		dotenv      string
		writeDotenv bool
		expected    map[string]string
		expectErr   string
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
		"strips inline comments": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "FOO=bar # this is a comment\n",
			writeDotenv: true,
			expected:    map[string]string{"FOO": "bar"},
		},
		"handles empty values": {
			fs:          afero.NewMemMapFs(),
			dotenv:      "EMPTY=\n",
			writeDotenv: true,
			expected:    map[string]string{"EMPTY": ""},
		},
		"returns parse error for invalid syntax": {
			fs:          afero.NewMemMapFs(),
			dotenv:      `KEY="unclosed`,
			writeDotenv: true,
			expectErr:   slackerror.ErrDotEnvFileParse,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.writeDotenv && tc.fs != nil {
				_ = afero.WriteFile(tc.fs, ".env", []byte(tc.dotenv), 0600)
			}
			result, err := Read(tc.fs)
			if tc.expectErr != "" {
				var slackErr *slackerror.Error
				require.ErrorAs(t, err, &slackErr)
				assert.Equal(t, tc.expectErr, slackErr.Code)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_Set(t *testing.T) {
	tests := map[string]struct {
		existingEnv   string
		writeExisting bool
		name          string
		value         string
		expectedFile  string
		expectErr     string
	}{
		"creates .env file when it does not exist": {
			name:         "FOO",
			value:        "bar",
			expectedFile: "FOO=\"bar\"\n",
		},
		"adds a variable to an empty .env file": {
			existingEnv:   "",
			writeExisting: true,
			name:          "FOO",
			value:         "bar",
			expectedFile:  "FOO=\"bar\"\n",
		},
		"adds a variable preserving existing variables": {
			existingEnv:   "EXISTING=value\n",
			writeExisting: true,
			name:          "NEW_VAR",
			value:         "new_value",
			expectedFile:  "EXISTING=value\nNEW_VAR=\"new_value\"\n",
		},
		"adds a variable preserving comments and blank lines": {
			existingEnv:   "# Database config\nDB_HOST=localhost\n\n# API keys\nAPI_KEY=secret\n",
			writeExisting: true,
			name:          "NEW_VAR",
			value:         "new_value",
			expectedFile:  "# Database config\nDB_HOST=localhost\n\n# API keys\nAPI_KEY=secret\nNEW_VAR=\"new_value\"\n",
		},
		"updates an existing unquoted variable in-place": {
			existingEnv:   "# Config\nFOO=old_value\nBAR=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "# Config\nFOO=\"new_value\"\nBAR=keep\n",
		},
		"updates an existing quoted variable in-place": {
			existingEnv:   "FOO=\"old_value\"\nBAR=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "FOO=\"new_value\"\nBAR=keep\n",
		},
		"updates a variable with export prefix": {
			existingEnv:   "export SECRET=old_secret\nOTHER=keep\n",
			writeExisting: true,
			name:          "SECRET",
			value:         "new_secret",
			expectedFile:  "export SECRET=\"new_secret\"\nOTHER=keep\n",
		},
		"escapes special characters in values": {
			name:         "SPECIAL",
			value:        "has \"quotes\" and $vars and \\ backslash",
			expectedFile: "SPECIAL=\"has \\\"quotes\\\" and \\$vars and \\\\ backslash\"\n",
		},
		"replaces a multiline value in-place": {
			existingEnv:   "export DB_KEY=\"---START---\npassword\n---END---\"\nOTHER=keep\n",
			writeExisting: true,
			name:          "DB_KEY",
			value:         "new_key",
			expectedFile:  "export DB_KEY=\"new_key\"\nOTHER=keep\n",
		},
		"returns error for value that cannot round-trip": {
			name:      "KEY",
			value:     `idk\`,
			expectErr: slackerror.ErrDotEnvVarMarshal,
		},
		"round-trips through Read": {
			name:         "ROUND_TRIP",
			value:        "hello world",
			expectedFile: "ROUND_TRIP=\"hello world\"\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tc.writeExisting {
				err := afero.WriteFile(fs, ".env", []byte(tc.existingEnv), 0600)
				assert.NoError(t, err)
			}
			err := Set(fs, tc.name, tc.value)
			if tc.expectErr != "" {
				var slackErr *slackerror.Error
				require.ErrorAs(t, err, &slackErr)
				assert.Equal(t, tc.expectErr, slackErr.Code)
				return
			}
			assert.NoError(t, err)
			content, err := afero.ReadFile(fs, ".env")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFile, string(content))
		})
	}
}
