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
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_clearAutoUpdateResult(t *testing.T) {
	tests := map[string]struct {
		savedResult *AutoUpdateResult
	}{
		"removes the saved result": {
			savedResult: &AutoUpdateResult{From: "v4.11.0", To: "v4.12.0"},
		},
		"succeeds without a saved result": {
			savedResult: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			if tc.savedResult != nil {
				require.NoError(t, writeAutoUpdateResult(ctx, clients, *tc.savedResult))
			}
			err := clearAutoUpdateResult(ctx, clients)
			require.NoError(t, err)
			resultPath := filepath.Join(slackdeps.MockHomeDirectory, ".slack", "autoupdate.json")
			exists, err := afero.Exists(clients.Fs, resultPath)
			require.NoError(t, err)
			assert.False(t, exists)
		})
	}
}

func Test_readAutoUpdateResult(t *testing.T) {
	tests := map[string]struct {
		expectedResult *AutoUpdateResult
	}{
		"returns nil without a saved result": {
			expectedResult: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			result, err := readAutoUpdateResult(ctx, clients)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func Test_writeAutoUpdateResult(t *testing.T) {
	tests := map[string]struct {
		configDirFlag string
		result        AutoUpdateResult
		expectedDir   string
	}{
		"saves the versions of an update": {
			result:      AutoUpdateResult{From: "v4.11.0", To: "v4.12.0"},
			expectedDir: filepath.Join(slackdeps.MockHomeDirectory, ".slack"),
		},
		"saves the error of a failed update": {
			result:      AutoUpdateResult{From: "v4.11.0", To: "v4.12.0", Error: "permission denied"},
			expectedDir: filepath.Join(slackdeps.MockHomeDirectory, ".slack"),
		},
		"saves to a custom config directory": {
			configDirFlag: filepath.Join("tmp", "custom"),
			result:        AutoUpdateResult{From: "v4.11.0", To: "v4.12.0"},
			expectedDir:   filepath.Join("tmp", "custom"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			if tc.configDirFlag != "" {
				require.NoError(t, clients.Fs.MkdirAll(tc.configDirFlag, 0o755))
				clients.Config.SystemConfig.SetCustomConfigDirPath(tc.configDirFlag)
			}
			err := writeAutoUpdateResult(ctx, clients, tc.result)
			require.NoError(t, err)
			matches, err := afero.Glob(clients.Fs, filepath.Join(tc.expectedDir, "autoupdate.json*"))
			require.NoError(t, err)
			assert.Equal(t, []string{filepath.Join(tc.expectedDir, "autoupdate.json")}, matches)
			result, err := readAutoUpdateResult(ctx, clients)
			require.NoError(t, err)
			assert.Equal(t, &tc.result, result)
		})
	}
}
