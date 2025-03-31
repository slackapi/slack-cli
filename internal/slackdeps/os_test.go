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

package slackdeps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			os := NewOs()
			os.executionDirPathAbs = tt.executionDirPathAbs

			// Run the test
			actualExecutionDirPathAbs := os.GetExecutionDir()

			// Assertions
			require.Equal(t, tt.expectedExecutionDirPathAbs, actualExecutionDirPathAbs)
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			os := NewOs()
			os.executionDirPathAbs = tt.executionDirPathAbs

			// Run the test
			os.SetExecutionDir(tt.setExecutionDirPathAbs)
			actualExecutionDirPathAbs := os.executionDirPathAbs

			// Assertions
			require.Equal(t, tt.expectedExecutionDirPathAbs, actualExecutionDirPathAbs)
		})
	}
}
