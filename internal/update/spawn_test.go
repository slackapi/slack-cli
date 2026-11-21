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

package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_backgroundUpdateArgs(t *testing.T) {
	tests := map[string]struct {
		configDir string
		expected  []string
	}{
		"uses the target version": {
			expected: []string{"_autoupdate", "4.12.0"},
		},
		"keeps a custom config directory": {
			configDir: "/tmp/custom",
			expected:  []string{"_autoupdate", "4.12.0", "--config-dir", "/tmp/custom"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, backgroundUpdateArgs("4.12.0", tc.configDir))
		})
	}
}

func Test_backgroundUpdateConfigDir(t *testing.T) {
	workingDir, err := os.Getwd()
	require.NoError(t, err)
	customConfigDir, err := filepath.Abs(filepath.Join(string(filepath.Separator), "tmp", "custom"))
	require.NoError(t, err)
	tests := map[string]struct {
		configDirFlag string
		expected      string
	}{
		"returns nothing without a custom config directory": {
			expected: "",
		},
		"keeps an absolute config directory": {
			configDirFlag: customConfigDir,
			expected:      customConfigDir,
		},
		"resolves a relative config directory from the working directory": {
			configDirFlag: " custom ",
			expected:      filepath.Join(workingDir, "custom"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			clients.Config.ConfigDirFlag = tc.configDirFlag
			configDir, err := backgroundUpdateConfigDir(clients)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, configDir)
		})
	}
}
