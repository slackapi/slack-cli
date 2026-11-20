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

package upgrade

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/update"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type UpdatePkgMock struct {
	mock.Mock
}

func (m *UpdatePkgMock) CheckForUpdates(clients *shared.ClientFactory, cmd *cobra.Command) error {
	args := m.Called(clients, cmd)
	return args.Error(0)
}

func TestUpgradeCommand(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Create the command
	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	updatePkgMock := new(UpdatePkgMock)
	checkForUpdatesFunc = updatePkgMock.CheckForUpdates

	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything).Return(nil)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade had unexpected error")
	}

	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything)
}

func Test_checkForUpdates(t *testing.T) {
	tests := map[string]struct {
		status         int
		release        string
		hookResponse   string
		hookErr        error
		installScript  bool
		experiments    []string
		expectedErr    string
		expectedOutput string
		expectedPrompt bool
	}{
		"prints the latest version when no update is found": {
			status:         http.StatusOK,
			release:        "0.0.0",
			expectedOutput: "You are using the latest Slack CLI version",
		},
		"returns the error when the update check fails": {
			status:      http.StatusInternalServerError,
			release:     "0.0.0",
			expectedErr: slackerror.ErrCLIUpdateCheck,
		},
		"prints the update before the error from the check-update hook": {
			status:         http.StatusOK,
			release:        "99.0.0",
			hookErr:        slackerror.New("The check-update hook failed"),
			expectedErr:    "The check-update hook failed",
			expectedOutput: "A new version of the Slack CLI is available",
		},
		"prompts to install the update without the auto-update experiment": {
			status:         http.StatusOK,
			release:        "99.0.0",
			installScript:  true,
			expectedOutput: "A new version of the Slack CLI is available",
			expectedPrompt: true,
		},
		"skips the prompt with the auto-update experiment": {
			status:         http.StatusOK,
			release:        "99.0.0",
			installScript:  true,
			experiments:    []string{string(experiment.AutoUpdate)},
			expectedOutput: "You are using the latest Slack CLI version",
			expectedPrompt: false,
		},
		"returns the error from the check-update hook": {
			status:       http.StatusOK,
			release:      "0.0.0",
			hookResponse: `{"error": {"message": "The SDK could not be checked"}}`,
			expectedErr:  "The SDK could not be checked",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"slack-cli": {"releases": [{"version": "` + tc.release + `"}]}}`))
			}))
			t.Cleanup(server.Close)
			metadataURL := update.MetadataURL
			t.Cleanup(func() { update.MetadataURL = metadataURL })
			update.MetadataURL = server.URL
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			if tc.installScript {
				homeDir := t.TempDir()
				executablePath := filepath.Join(homeDir, ".slack", "bin", "slack")
				require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
				require.NoError(t, os.WriteFile(executablePath, []byte("#!/bin/bash\necho Using slack v"+tc.release+"\n"), 0o755))
				clientsMock.Os.On("Executable").Return(executablePath, nil)
				clientsMock.Os.On("UserHomeDir").Return(homeDir, nil)
			}
			clientsMock.AddDefaultMocks()
			clientsMock.IO.On("ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
			clientsMock.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return(tc.hookResponse, tc.hookErr)
			clientsMock.Config.ExperimentsFlag = tc.experiments
			clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			if tc.hookResponse != "" || tc.hookErr != nil {
				clients.SDKConfig.Hooks.CheckUpdate = hooks.HookScript{Name: "CheckUpdate", Command: "check-update"}
			}
			cmd := &cobra.Command{}
			cmd.SetContext(ctx)
			testutil.MockCmdIO(clients.IO, cmd)
			err := checkForUpdates(clients, cmd)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, slackerror.ToSlackError(err).Error(), tc.expectedErr)
				assert.NotContains(t, clientsMock.GetCombinedOutput(), "You are using the latest")
			} else {
				require.NoError(t, err)
			}
			assert.Contains(t, clientsMock.GetCombinedOutput(), tc.expectedOutput)
			if tc.expectedPrompt {
				clientsMock.IO.AssertCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
			} else {
				clientsMock.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
