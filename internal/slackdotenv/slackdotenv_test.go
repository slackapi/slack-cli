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
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Init(t *testing.T) {
	tests := map[string]struct {
		files       map[string]string
		expected    string
		expectedEnv string
		expectErr   string
	}{
		"copies .env.sample to .env": {
			files:       map[string]string{".env.sample": "FOO=bar\n"},
			expected:    ".env.sample",
			expectedEnv: "FOO=bar\n",
		},
		"copies .env.example to .env": {
			files:       map[string]string{".env.example": "BAZ=qux\n"},
			expected:    ".env.example",
			expectedEnv: "BAZ=qux\n",
		},
		"prefers .env.sample over .env.example": {
			files: map[string]string{
				".env.sample":  "FROM_SAMPLE=1\n",
				".env.example": "FROM_EXAMPLE=1\n",
			},
			expected:    ".env.sample",
			expectedEnv: "FROM_SAMPLE=1\n",
		},
		"returns error when .env already exists": {
			files:     map[string]string{".env": "EXISTING=value\n"},
			expectErr: slackerror.ErrDotEnvFileAlreadyExists,
		},
		"returns error when placeholder file cannot be parsed": {
			files:     map[string]string{".env.sample": "INVALID LINE WITHOUT EQUALS\n"},
			expectErr: slackerror.ErrDotEnvFileParse,
		},
		"returns error when no sample file exists": {
			files:     map[string]string{},
			expectErr: slackerror.ErrDotEnvPlaceholderNotFound,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				err := afero.WriteFile(fs, path, []byte(content), 0600)
				require.NoError(t, err)
			}
			source, err := Init(fs)
			if tc.expectErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, slackerror.ToSlackError(err).Code)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, source)
			content, err := afero.ReadFile(fs, ".env")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedEnv, string(content))
		})
	}
}

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
		"adds a variable preserving newline comments and blank lines": {
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
		"updates a variable with spaces around equals": {
			existingEnv:   "BEFORE=keep\nFOO = old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates a variable with space before equals": {
			existingEnv:   "BEFORE=keep\nFOO =old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates a variable with space after equals": {
			existingEnv:   "BEFORE=keep\nFOO= old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates an existing empty value": {
			existingEnv:   "BEFORE=keep\nFOO=\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates an existing empty value with spaces": {
			existingEnv:   "BEFORE=keep\nFOO = \nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates export variable with spaces around equals": {
			existingEnv:   "BEFORE=keep\nexport FOO = old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nexport FOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates a variable with leading spaces": {
			existingEnv:   "BEFORE=keep\n  FOO=old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates a variable with leading tab": {
			existingEnv:   "BEFORE=keep\n\tFOO=old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\"\nAFTER=keep\n",
		},
		"updates export variable with leading spaces": {
			existingEnv:   "BEFORE=keep\n  export FOO=old_value\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nexport FOO=\"new_value\"\nAFTER=keep\n",
		},
		"preserves inline comment on unquoted value": {
			existingEnv:   "BEFORE=keep\nFOO=old_value # important note\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\" # important note\nAFTER=keep\n",
		},
		"preserves inline comment on quoted value": {
			existingEnv:   "BEFORE=keep\nFOO=\"old_value\" # important note\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nFOO=\"new_value\" # important note\nAFTER=keep\n",
		},
		"preserves inline comment on export variable": {
			existingEnv:   "BEFORE=keep\nexport FOO=old_value # important note\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			value:         "new_value",
			expectedFile:  "BEFORE=keep\nexport FOO=\"new_value\" # important note\nAFTER=keep\n",
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
				require.Error(t, err)
				assert.Equal(t, tc.expectErr, slackerror.ToSlackError(err).Code)
				return
			}
			assert.NoError(t, err)
			content, err := afero.ReadFile(fs, ".env")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFile, string(content))
		})
	}
}

func Test_Unset(t *testing.T) {
	tests := map[string]struct {
		existingEnv   string
		writeExisting bool
		name          string
		expectedFile  string
	}{
		"no-op when .env file does not exist": {
			name:         "FOO",
			expectedFile: "",
		},
		"no-op when key does not exist": {
			existingEnv:   "OTHER=value\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "OTHER=value\n",
		},
		"removes a simple key-value pair": {
			existingEnv:   "FOO=bar\nBAZ=qux\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BAZ=qux\n",
		},
		"removes a quoted value": {
			existingEnv:   "TOKEN=\"my secret\"\nOTHER=keep\n",
			writeExisting: true,
			name:          "TOKEN",
			expectedFile:  "OTHER=keep\n",
		},
		"removes a key with export prefix": {
			existingEnv:   "export SECRET=mysecret\nOTHER=keep\n",
			writeExisting: true,
			name:          "SECRET",
			expectedFile:  "OTHER=keep\n",
		},
		"preserves comments and blank lines": {
			existingEnv:   "# Config\nFOO=bar\n\n# Keys\nAPI_KEY=secret\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "# Config\n\n# Keys\nAPI_KEY=secret\n",
		},
		"removes the only variable": {
			existingEnv:   "FOO=bar\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "",
		},
		"removes a multiline value": {
			existingEnv:   "export DB_KEY=\"---START---\npassword\n---END---\"\nOTHER=keep\n",
			writeExisting: true,
			name:          "DB_KEY",
			expectedFile:  "OTHER=keep\n",
		},
		"removes a variable with spaces around equals": {
			existingEnv:   "BEFORE=keep\nFOO = bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes a variable with space before equals": {
			existingEnv:   "BEFORE=keep\nFOO =bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes a variable with space after equals": {
			existingEnv:   "BEFORE=keep\nFOO= bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes an empty value": {
			existingEnv:   "BEFORE=keep\nFOO=\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes an empty value with spaces": {
			existingEnv:   "BEFORE=keep\nFOO = \nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes export variable with spaces around equals": {
			existingEnv:   "BEFORE=keep\nexport FOO = bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes a variable with leading spaces": {
			existingEnv:   "BEFORE=keep\n  FOO=bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes a variable with leading tab": {
			existingEnv:   "BEFORE=keep\n\tFOO=bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes export variable with leading spaces": {
			existingEnv:   "BEFORE=keep\n  export FOO=bar\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes a variable with inline comment": {
			existingEnv:   "BEFORE=keep\nFOO=bar # important note\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes a quoted variable with inline comment": {
			existingEnv:   "BEFORE=keep\nFOO=\"bar\" # important note\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
		"removes an export variable with inline comment": {
			existingEnv:   "BEFORE=keep\nexport FOO=bar # important note\nAFTER=keep\n",
			writeExisting: true,
			name:          "FOO",
			expectedFile:  "BEFORE=keep\nAFTER=keep\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tc.writeExisting {
				err := afero.WriteFile(fs, ".env", []byte(tc.existingEnv), 0600)
				assert.NoError(t, err)
			}
			err := Unset(fs, tc.name)
			assert.NoError(t, err)
			if !tc.writeExisting {
				_, err := fs.Stat(".env")
				assert.True(t, os.IsNotExist(err))
				return
			}
			content, err := afero.ReadFile(fs, ".env")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFile, string(content))
		})
	}
}
