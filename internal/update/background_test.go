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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AutoUpdate(t *testing.T) {
	tests := map[string]struct {
		installedVersion string
		lockHeld         bool
		movedBinary      bool
		installErr       error
		attempts         int
		expectedCalls    int
		expectedResult   *AutoUpdateResult
		expectedErr      error
	}{
		"records the new version after installing": {
			installedVersion: "v4.11.0",
			attempts:         1,
			expectedCalls:    1,
			expectedResult:   &AutoUpdateResult{From: "v4.11.0", To: "v4.12.0"},
		},
		"records the error when the install fails": {
			installedVersion: "v4.11.0",
			installErr:       errors.New("permission denied"),
			attempts:         1,
			expectedCalls:    1,
			expectedResult:   &AutoUpdateResult{From: "v4.11.0", To: "v4.12.0", Error: "permission denied"},
			expectedErr:      errors.New("permission denied"),
		},
		"skips while another update holds the lock": {
			installedVersion: "v4.11.0",
			lockHeld:         true,
			attempts:         1,
			expectedCalls:    0,
			expectedResult:   nil,
		},
		"skips an update that starts after another one installed": {
			installedVersion: "v4.11.0",
			attempts:         2,
			expectedCalls:    1,
			expectedResult:   &AutoUpdateResult{From: "v4.11.0", To: "v4.12.0"},
		},
		"skips when another update moved the binary": {
			installedVersion: "v4.11.0",
			lockHeld:         true,
			movedBinary:      true,
			attempts:         1,
			expectedCalls:    0,
			expectedResult:   nil,
			expectedErr:      errors.New("the running binary is not in a folder that the installation scripts write to"),
		},
		"skips when a newer version is installed": {
			installedVersion: "v4.13.0",
			attempts:         1,
			expectedCalls:    0,
			expectedResult:   nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			homeDir := t.TempDir()
			binaryPath := filepath.Join(homeDir, ".slack", "bin", "slack")
			executablePath := binaryPath
			if tc.movedBinary {
				executablePath = filepath.Join(homeDir, ".slack", "bin", "backups", "v4.11.0", "slack")
			}
			require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
			require.NoError(t, os.WriteFile(executablePath, fmt.Appendf(nil, scriptTemplate, "Using slack "+tc.installedVersion), 0o755))
			lockPath := filepath.Join(filepath.Dir(binaryPath), "autoupdate.lock")
			clientsMock := shared.NewClientsMock()
			clientsMock.Os.On("Executable").Return(executablePath, nil)
			clientsMock.Os.On("UserHomeDir").Return(homeDir, nil)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			holder := flock.New(lockPath)
			if tc.lockHeld {
				locked, err := holder.TryLock()
				require.NoError(t, err)
				require.True(t, locked)
			}
			calls := 0
			originalInstallUpdateFunc := installUpdateFunc
			t.Cleanup(func() {
				installUpdateFunc = originalInstallUpdateFunc
			})
			installUpdateFunc = func(ctx context.Context, c *CLIDependency) error {
				calls++
				assert.Equal(t, "v4.11.0", c.version)
				assert.Equal(t, "4.12.0", c.releaseInfo.Version)
				_, hasDeadline := ctx.Deadline()
				assert.True(t, hasDeadline, "the install has a deadline")
				locked, err := flock.New(lockPath).TryLock()
				require.NoError(t, err)
				assert.False(t, locked, "the lock is held while installing")
				if tc.installErr != nil {
					return tc.installErr
				}
				return os.WriteFile(binaryPath, fmt.Appendf(nil, scriptTemplate, "Using slack v4.12.0"), 0o755)
			}
			var err error
			for range tc.attempts {
				err = AutoUpdate(ctx, clients, "v4.11.0", "4.12.0")
			}
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedCalls, calls)
			result, err := readAutoUpdateResult(ctx, clients)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
			require.NoError(t, holder.Unlock())
			assert.FileExists(t, lockPath, "the lock file stays after unlocking")
			locked, err := flock.New(lockPath).TryLock()
			require.NoError(t, err)
			assert.True(t, locked, "the lock is released afterwards")
		})
	}
}
