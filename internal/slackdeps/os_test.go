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

package slackdeps

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Os_Getenv(t *testing.T) {
	tests := map[string]struct {
		key      string
		value    string
		expected string
	}{
		"returns set env var": {
			key:      "SLACK_TEST_OS_GETENV",
			value:    "hello",
			expected: "hello",
		},
		"returns empty for unset env var": {
			key:      "SLACK_TEST_OS_GETENV_UNSET",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			o := NewOs()
			if tc.value != "" {
				t.Setenv(tc.key, tc.value)
			}
			require.Equal(t, tc.expected, o.Getenv(tc.key))
		})
	}
}

func Test_Os_SetenvUnsetenv(t *testing.T) {
	o := NewOs()
	key := "SLACK_TEST_OS_SETENV"

	err := o.Setenv(key, "test_value")
	require.NoError(t, err)
	require.Equal(t, "test_value", o.Getenv(key))

	err = o.Unsetenv(key)
	require.NoError(t, err)
	require.Equal(t, "", o.Getenv(key))
}

func Test_Os_LookupEnv(t *testing.T) {
	tests := map[string]struct {
		key             string
		value           string
		setValue        bool
		expectedValue   string
		expectedPresent bool
	}{
		"returns value and true for set env var": {
			key:             "SLACK_TEST_OS_LOOKUPENV",
			value:           "present",
			setValue:        true,
			expectedValue:   "present",
			expectedPresent: true,
		},
		"returns empty and false for unset env var": {
			key:             "SLACK_TEST_OS_LOOKUPENV_UNSET",
			setValue:        false,
			expectedValue:   "",
			expectedPresent: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			o := NewOs()
			if tc.setValue {
				t.Setenv(tc.key, tc.value)
			}
			val, present := o.LookupEnv(tc.key)
			require.Equal(t, tc.expectedValue, val)
			require.Equal(t, tc.expectedPresent, present)
		})
	}
}

func Test_Os_Getwd(t *testing.T) {
	o := NewOs()
	dir, err := o.Getwd()
	require.NoError(t, err)
	require.NotEmpty(t, dir)
}

func Test_Os_UserHomeDir(t *testing.T) {
	o := NewOs()
	home, err := o.UserHomeDir()
	require.NoError(t, err)
	require.NotEmpty(t, home)
}

func Test_Os_IsNotExist(t *testing.T) {
	tests := map[string]struct {
		err      error
		expected bool
	}{
		"returns true for os.ErrNotExist": {
			err:      os.ErrNotExist,
			expected: true,
		},
		"returns false for other errors": {
			err:      os.ErrPermission,
			expected: false,
		},
		"returns false for nil": {
			err:      nil,
			expected: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			o := NewOs()
			require.Equal(t, tc.expected, o.IsNotExist(tc.err))
		})
	}
}

func Test_Os_Glob(t *testing.T) {
	o := NewOs()
	// Create a temp file to glob for
	tmpDir := t.TempDir()
	f, err := os.CreateTemp(tmpDir, "glob_test_*.txt")
	require.NoError(t, err)
	f.Close()

	matches, err := o.Glob(tmpDir + "/*.txt")
	require.NoError(t, err)
	require.NotEmpty(t, matches)
}

func Test_Os_Stdout(t *testing.T) {
	o := NewOs()
	stdout := o.Stdout()
	require.NotNil(t, stdout)
}

func Test_Os_GetExecutionDir(t *testing.T) {
	tests := map[string]struct {
		executionDirPathAbs         string
		expectedExecutionDirPathAbs string
	}{
		"When unset should return blank": {
			executionDirPathAbs:         "",
			expectedExecutionDirPathAbs: "",
		},
		"When set should return path": {
			executionDirPathAbs:         "/path/to/execution/dir",
			expectedExecutionDirPathAbs: "/path/to/execution/dir",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			os := NewOs()
			os.executionDirPathAbs = tc.executionDirPathAbs

			// Run the test
			actualExecutionDirPathAbs := os.GetExecutionDir()

			// Assertions
			require.Equal(t, tc.expectedExecutionDirPathAbs, actualExecutionDirPathAbs)
		})
	}
}

func Test_Os_SetExecutionDir(t *testing.T) {
	tests := map[string]struct {
		executionDirPathAbs         string
		setExecutionDirPathAbs      string
		expectedExecutionDirPathAbs string
	}{
		"Successful set path": {
			executionDirPathAbs:         "",
			setExecutionDirPathAbs:      "/path/to/execution/dir",
			expectedExecutionDirPathAbs: "/path/to/execution/dir",
		},
		"Successfully overwrite path": {
			executionDirPathAbs:         "/some/original/path",
			setExecutionDirPathAbs:      "/path/to/execution/dir",
			expectedExecutionDirPathAbs: "/path/to/execution/dir",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			os := NewOs()
			os.executionDirPathAbs = tc.executionDirPathAbs

			// Run the test
			os.SetExecutionDir(tc.setExecutionDirPathAbs)
			actualExecutionDirPathAbs := os.executionDirPathAbs

			// Assertions
			require.Equal(t, tc.expectedExecutionDirPathAbs, actualExecutionDirPathAbs)
		})
	}
}
