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
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NewCLIDependency(t *testing.T) {
	tests := map[string]struct {
		version string
	}{
		"stores the clients and the current version": {
			version: "v3.4.5",
		},
		"stores an empty version when none is known": {
			version: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cliDependency := NewCLIDependency(clients, tc.version)
			assert.Same(t, clients, cliDependency.clients)
			assert.Equal(t, tc.version, cliDependency.version)
			assert.Nil(t, cliDependency.releaseInfo)
		})
	}
}

func Test_CLIDependency_HasUpdate(t *testing.T) {
	tests := map[string]struct {
		releaseInfo    *LatestCLIRelease
		expectedUpdate bool
	}{
		"returns false when no newer release was found": {
			releaseInfo:    nil,
			expectedUpdate: false,
		},
		"returns true when a newer release was found": {
			releaseInfo:    &LatestCLIRelease{Version: "v4.0.0"},
			expectedUpdate: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cliDependency := &CLIDependency{
				releaseInfo: tc.releaseInfo,
			}
			hasUpdate, err := cliDependency.HasUpdate()
			assert.Equal(t, tc.expectedUpdate, hasUpdate)
			assert.NoError(t, err)
		})
	}
}

func Test_CLIDependency_PrintUpdateNotification(t *testing.T) {
	tests := map[string]struct {
		executablePath     string
		expectedSelfUpdate bool
	}{
		"offers a self-update to binaries placed by the install script": {
			executablePath:     ".slack/bin/slack",
			expectedSelfUpdate: true,
		},
		"skips the self-update for binaries placed elsewhere": {
			executablePath:     ".slack/dev-build/bin/slack",
			expectedSelfUpdate: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			homeDir := t.TempDir()
			executablePath := filepath.Join(homeDir, tc.executablePath)
			require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
			require.NoError(t, os.WriteFile(executablePath, []byte{}, 0o755))
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
			clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.Anything)
			clientsMock.Os.On("Executable").Return(executablePath, nil)
			clientsMock.Os.On("UserHomeDir").Return(homeDir, nil)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cmd := &cobra.Command{}
			cmd.SetContext(ctx)
			testutil.MockCmdIO(clients.IO, cmd)
			cliDependency := CLIDependency{
				clients:     clients,
				version:     "v4.10.0",
				releaseInfo: &LatestCLIRelease{Version: "4.11.0"},
			}
			selfUpdate, err := cliDependency.PrintUpdateNotification(cmd)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedSelfUpdate, selfUpdate)
			assert.Contains(t, clientsMock.GetCombinedOutput(), "https://docs.slack.dev/tools/slack-cli")
			if tc.expectedSelfUpdate {
				clientsMock.IO.AssertCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
			} else {
				clientsMock.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
