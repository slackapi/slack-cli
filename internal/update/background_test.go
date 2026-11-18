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
		lockHeld       bool
		installErr     error
		expectedCalls  int
		expectedResult *AutoUpdateResult
		expectedErr    error
	}{
		"records the new version after installing": {
			expectedCalls:  1,
			expectedResult: &AutoUpdateResult{From: "v4.11.0", To: "v4.12.0"},
		},
		"records the error when the install fails": {
			installErr:     errors.New("permission denied"),
			expectedCalls:  1,
			expectedResult: &AutoUpdateResult{From: "v4.11.0", To: "v4.12.0", Error: "permission denied"},
			expectedErr:    errors.New("permission denied"),
		},
		"skips while another update holds the lock": {
			lockHeld:       true,
			expectedCalls:  0,
			expectedResult: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			configDir := t.TempDir()
			lockPath := filepath.Join(configDir, "autoupdate.lock")
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			require.NoError(t, clients.Fs.MkdirAll(configDir, 0o755))
			clients.Config.SystemConfig.SetCustomConfigDirPath(configDir)
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
				locked, err := flock.New(lockPath).TryLock()
				require.NoError(t, err)
				assert.False(t, locked, "the lock is held while installing")
				return tc.installErr
			}
			err := AutoUpdate(ctx, clients, "v4.11.0", "4.12.0")
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
			locked, err := flock.New(lockPath).TryLock()
			require.NoError(t, err)
			assert.True(t, locked, "the lock is released afterwards")
		})
	}
}
