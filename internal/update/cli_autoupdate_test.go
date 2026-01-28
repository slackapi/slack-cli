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

package update

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CLI_getUpdateFileName(t *testing.T) {
	tests := map[string]struct {
		version          string
		operatingSystem  string
		architecture     string
		expectedFilename string
		expectedErrorF   string
	}{
		"darwin production x86_64": {
			version:          "3.4.5",
			operatingSystem:  "darwin",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5_macOS_amd64.zip",
		},
		"darwin development x86_64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "darwin",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_macOS_amd64.zip",
		},
		"darwin production aarch64": {
			version:          "3.4.5",
			operatingSystem:  "darwin",
			architecture:     "arm64",
			expectedFilename: "slack_cli_3.4.5_macOS_arm64.zip",
		},
		"darwin development aarch64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "darwin",
			architecture:     "arm64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_macOS_arm64.zip",
		},
		"darwin production universal": {
			version:          "3.4.5",
			operatingSystem:  "darwin",
			architecture:     "fallback",
			expectedFilename: "slack_cli_3.4.5_macOS_64-bit.zip",
		},
		"darwin development universal": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "darwin",
			architecture:     "fallback",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_macOS_64-bit.zip",
		},
		"linux production x86_64": {
			version:          "3.4.5",
			operatingSystem:  "linux",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5_linux_64-bit.tar.gz",
		},
		"linux development x86_64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "linux",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_linux_64-bit.tar.gz",
		},
		"windows production x86_64": {
			version:          "3.4.5",
			operatingSystem:  "windows",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5_windows_64-bit.zip",
		},
		"windows development x86_64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "windows",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_windows_64-bit.zip",
		},
		"unknown production errors": {
			version:         "3.4.5",
			operatingSystem: "dragonfly",
			architecture:    "amd64",
			expectedErrorF:  "auto-updating for the operating system (dragonfly) is unsupported",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			filename, err := getUpdateFileName(tc.version, tc.operatingSystem, tc.architecture)
			if tc.expectedErrorF != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectedErrorF)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedFilename, filename)
			}
		})
	}
}
