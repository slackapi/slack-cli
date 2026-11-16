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
			assert.Nil(t, cliDependency.releaseError)
		})
	}
}

func Test_CLIDependency_PrintUpdateNotification(t *testing.T) {
	tests := map[string]struct {
		isInstallScript    bool
		expectedSelfUpdate bool
	}{
		"offers a self-update to binaries placed by the install script": {
			isInstallScript:    true,
			expectedSelfUpdate: true,
		},
		"skips the self-update for binaries placed elsewhere": {
			isInstallScript:    false,
			expectedSelfUpdate: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cmd := &cobra.Command{}
			cmd.SetContext(ctx)
			testutil.MockCmdIO(clients.IO, cmd)
			originalIsInstallScriptFunc := isInstallScriptFunc
			isInstallScriptFunc = func(_ *shared.ClientFactory) bool {
				return tc.isInstallScript
			}
			t.Cleanup(func() { isInstallScriptFunc = originalIsInstallScriptFunc })
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
