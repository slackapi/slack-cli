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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDependency struct {
	mock.Mock
}

func (m *mockDependency) CheckForUpdate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDependency) PrintUpdateNotification(cmd *cobra.Command) (bool, error) {
	args := m.Called(cmd)
	return args.Bool(0), args.Error(1)
}

func (m *mockDependency) HasUpdate() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

func (m *mockDependency) InstallUpdate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func Test_UpdateNotification_CheckForUpdate(t *testing.T) {
	now := time.Now()
	lastUpdateCheckedAt := now.Add(-48 * time.Hour)
	tests := map[string]struct {
		forced                          bool
		status                          int
		unreachable                     bool
		hookErr                         error
		lastUpdateCheckFailedAt         time.Time
		expectedErr                     string
		expectedRequests                int32
		expectedLastUpdateCheckedAt     time.Time
		expectedLastUpdateCheckFailedAt time.Time
	}{
		"returns no error when a forced check succeeds": {
			forced:                          true,
			status:                          http.StatusOK,
			lastUpdateCheckFailedAt:         now.Add(-30 * time.Minute),
			expectedRequests:                1,
			expectedLastUpdateCheckedAt:     now,
			expectedLastUpdateCheckFailedAt: time.Time{},
		},
		"returns the error when a forced check gets an error response": {
			forced:                          true,
			status:                          http.StatusInternalServerError,
			expectedErr:                     slackerror.ErrCLIUpdateCheck,
			expectedRequests:                1,
			expectedLastUpdateCheckedAt:     lastUpdateCheckedAt,
			expectedLastUpdateCheckFailedAt: now,
		},
		"returns the error when a forced check can't reach the server": {
			forced:                          true,
			unreachable:                     true,
			expectedErr:                     slackerror.ErrCLIUpdateCheck,
			expectedRequests:                0,
			expectedLastUpdateCheckedAt:     lastUpdateCheckedAt,
			expectedLastUpdateCheckFailedAt: now,
		},
		"returns no error when a background check gets an error response": {
			forced:                          false,
			status:                          http.StatusInternalServerError,
			lastUpdateCheckFailedAt:         now.Add(-2 * time.Hour),
			expectedRequests:                1,
			expectedLastUpdateCheckedAt:     lastUpdateCheckedAt,
			expectedLastUpdateCheckFailedAt: now,
		},
		"skips a background check within an hour of a failed check": {
			forced:                          false,
			status:                          http.StatusOK,
			lastUpdateCheckFailedAt:         now.Add(-30 * time.Minute),
			expectedRequests:                0,
			expectedLastUpdateCheckedAt:     lastUpdateCheckedAt,
			expectedLastUpdateCheckFailedAt: now.Add(-30 * time.Minute),
		},
		"counts a failed check-update hook as a successful check": {
			forced:                          false,
			status:                          http.StatusOK,
			hookErr:                         slackerror.New("The check-update hook failed"),
			lastUpdateCheckFailedAt:         now.Add(-2 * time.Hour),
			expectedRequests:                1,
			expectedLastUpdateCheckedAt:     now,
			expectedLastUpdateCheckFailedAt: time.Time{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, env := range []string{envSkipUpdate, "CI", "BUILD_NUMBER", "RUN_ID"} {
				t.Setenv(env, "")
			}
			args := os.Args
			t.Cleanup(func() { os.Args = args })
			os.Args = []string{"placeholder"}
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"slack-cli": {"releases": [{"version": "4.10.0"}]}}`))
			}))
			if tc.unreachable {
				server.Close()
			} else {
				t.Cleanup(server.Close)
			}
			metadataURL := MetadataURL
			t.Cleanup(func() { MetadataURL = metadataURL })
			MetadataURL = server.URL
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clientsMock.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return("", tc.hookErr)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			if tc.hookErr != nil {
				clients.SDKConfig.Hooks.CheckUpdate = hooks.HookScript{Name: "CheckUpdate", Command: "check-update"}
			}
			_, err := clients.Config.SystemConfig.SetLastUpdateCheckedAt(ctx, lastUpdateCheckedAt)
			require.NoError(t, err)
			_, err = clients.Config.SystemConfig.SetLastUpdateCheckFailedAt(ctx, tc.lastUpdateCheckFailedAt)
			require.NoError(t, err)
			u := New(clients, "v4.10.0")
			err = u.CheckForUpdate(ctx, tc.forced)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, slackerror.ToSlackError(err).Code)
			} else {
				require.NoError(t, err)
			}
			actualLastUpdateCheckedAt, err := clients.Config.SystemConfig.GetLastUpdateCheckedAt(ctx)
			require.NoError(t, err)
			assert.WithinDuration(t, tc.expectedLastUpdateCheckedAt, actualLastUpdateCheckedAt, time.Minute)
			actualLastUpdateCheckFailedAt, err := clients.Config.SystemConfig.GetLastUpdateCheckFailedAt(ctx)
			require.NoError(t, err)
			assert.WithinDuration(t, tc.expectedLastUpdateCheckFailedAt, actualLastUpdateCheckFailedAt, time.Minute)
			// A background check right after is skipped
			require.NoError(t, u.CheckForUpdate(ctx, false))
			assert.Equal(t, tc.expectedRequests, requests.Load())
		})
	}
}

func Test_UpdateNotification_Dependencies(t *testing.T) {
	dep := &mockDependency{}
	u := &UpdateNotification{
		dependencies: []Dependency{dep},
	}
	require.Len(t, u.Dependencies(), 1)
	require.Equal(t, dep, u.Dependencies()[0])
}

func Test_UpdateNotification_HasUpdate(t *testing.T) {
	for name, tc := range map[string]struct {
		dependencyHasUpdate []bool
		expectedReturnValue bool
	}{
		"No updates": {
			dependencyHasUpdate: []bool{false, false, false},
			expectedReturnValue: false,
		},
		"First dependency has update": {
			dependencyHasUpdate: []bool{true, false, false},
			expectedReturnValue: true,
		},
		"Middle dependency has update": {
			dependencyHasUpdate: []bool{false, true, false},
			expectedReturnValue: true,
		},
		"Last dependency has update": {
			dependencyHasUpdate: []bool{false, false, true},
			expectedReturnValue: true,
		},
		"All dependencies have updates": {
			dependencyHasUpdate: []bool{true, true, true},
			expectedReturnValue: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Setup mock dependencies
			var dependencies []Dependency
			for _, hasUpdate := range tc.dependencyHasUpdate {
				dependency := mockDependency{}
				dependency.On("HasUpdate").Return(hasUpdate, nil)
				dependencies = append(dependencies, &dependency)
			}

			// Create clients
			clients := shared.ClientFactory{
				Config:    config.NewConfig(slackdeps.NewFsMock(), slackdeps.NewOsMock()),
				SDKConfig: hooks.NewSDKConfigMock(),
			}
			// Create updateNotification
			u := &UpdateNotification{
				clients:      &clients,
				hoursToWait:  defaultHoursToWait,
				dependencies: dependencies,
			}

			// Test
			require.Equal(t, tc.expectedReturnValue, u.HasUpdate())
		})
	}
}

func Test_UpdateNotification_PrintAndPromptUpdates(t *testing.T) {
	tests := map[string]struct {
		checkInBackground bool
		hasUpdate         bool
		shouldSelfUpdate  bool
		expectedInstalls  int
	}{
		"returns without a background check": {
			hasUpdate:        false,
			expectedInstalls: 0,
		},
		"installs an accepted update": {
			hasUpdate:        true,
			shouldSelfUpdate: true,
			expectedInstalls: 1,
		},
		"skips a declined update": {
			hasUpdate:        true,
			shouldSelfUpdate: false,
			expectedInstalls: 0,
		},
		"waits for a running background check": {
			checkInBackground: true,
			hasUpdate:         true,
			shouldSelfUpdate:  true,
			expectedInstalls:  1,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, env := range []string{envSkipUpdate, "CI", "BUILD_NUMBER", "RUN_ID"} {
				t.Setenv(env, "")
			}
			args := os.Args
			t.Cleanup(func() { os.Args = args })
			os.Args = []string{"placeholder"}
			ctx := slackcontext.MockContext(t.Context())
			cmd := &cobra.Command{}
			cmd.SetContext(ctx)
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			release := make(chan struct{})
			var checked atomic.Bool
			dependency := &mockDependency{}
			dependency.On("CheckForUpdate", mock.Anything).Run(func(mock.Arguments) {
				<-release
				checked.Store(true)
			}).Return(nil)
			dependency.On("HasUpdate").Run(func(mock.Arguments) {
				assert.Equal(t, tc.checkInBackground, checked.Load(), "HasUpdate ran before the background check finished")
			}).Return(tc.hasUpdate, nil)
			dependency.On("PrintUpdateNotification", cmd).Return(tc.shouldSelfUpdate, nil)
			dependency.On("InstallUpdate", mock.Anything).Return(nil)
			u := &UpdateNotification{
				clients:      clients,
				hoursToWait:  defaultHoursToWait,
				dependencies: []Dependency{dependency},
			}
			if tc.checkInBackground {
				u.CheckForUpdateInBackground(ctx)
				// Finish the check after PrintAndPromptUpdates has started
				time.AfterFunc(10*time.Millisecond, func() { close(release) })
			}
			done := make(chan error, 1)
			go func() {
				done <- u.PrintAndPromptUpdates(cmd)
			}()
			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(time.Second):
				require.Fail(t, "PrintAndPromptUpdates did not return")
			}
			dependency.AssertNumberOfCalls(t, "InstallUpdate", tc.expectedInstalls)
		})
	}
}

func Test_UpdateNotification_isEnabled(t *testing.T) {
	tests := map[string]struct {
		args           []string
		command        string
		env            map[string]string
		skipUpdateFlag bool
		tokenFlag      string
		expected       bool
	}{
		"checks for updates on other commands": {
			args:     []string{"auth", "list"},
			command:  "auth list",
			expected: true,
		},
		"skips the upgrade command with flags before it": {
			args:     []string{"--verbose", "upgrade"},
			command:  "upgrade",
			expected: false,
		},
		"skips with the skip update flag": {
			args:           []string{"auth", "list", "--skip-update"},
			command:        "auth list",
			skipUpdateFlag: true,
			expected:       false,
		},
		"skips with a token flag": {
			args:      []string{"auth", "list", "--token", "xoxb-test-token"},
			command:   "auth list",
			tokenFlag: "xoxb-test-token",
			expected:  false,
		},
		"skips with the skip update environment variable": {
			args:     []string{"auth", "list"},
			command:  "auth list",
			env:      map[string]string{envSkipUpdate: "true"},
			expected: false,
		},
		"skips in continuous integration": {
			args:     []string{"auth", "list"},
			command:  "auth list",
			env:      map[string]string{"CI": "true"},
			expected: false,
		},
		"skips ignored commands": {
			args:     []string{"version"},
			command:  "version",
			expected: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, env := range []string{envSkipUpdate, "CI", "BUILD_NUMBER", "RUN_ID"} {
				t.Setenv(env, tc.env[env])
			}
			args := os.Args
			t.Cleanup(func() { os.Args = args })
			os.Args = append([]string{"placeholder"}, tc.args...)
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			clients.Config.Command = tc.command
			clients.Config.SkipUpdateFlag = tc.skipUpdateFlag
			clients.Config.TokenFlag = tc.tokenFlag
			u := New(clients, "v4.10.0")
			require.Equal(t, tc.expected, u.isEnabled())
		})
	}
}

func Test_UpdateNotification_isIgnoredCommand(t *testing.T) {
	for name, tc := range map[string]struct {
		command  string
		expected bool
	}{
		"No command": {
			command:  "",
			expected: false,
		},
		"api command": {
			command:  "api",
			expected: true,
		},
		"api command with method": {
			command:  "api chat.postMessage",
			expected: true,
		},
		"api command with method and args": {
			command:  "api chat.postMessage channel=C123 text=hello",
			expected: true,
		},
		"api command with flags": {
			command:  "api chat.postMessage --json {}",
			expected: true,
		},
		"fingerprint command": {
			command:  "_fingerprint",
			expected: true,
		},
		"version command": {
			command:  "version",
			expected: true,
		},
		"manifest command": {
			command:  "manifest",
			expected: true,
		},
		"manifest command with flags": {
			command:  "manifest --source local",
			expected: true,
		},
		"manifest info command": {
			command:  "manifest info",
			expected: true,
		},
		"manifest info command with flags": {
			command:  "manifest info --source local",
			expected: true,
		},
		"manifest validate command": {
			command:  "manifest validate",
			expected: false,
		},
		"auth command": {
			command:  "auth",
			expected: false,
		},
		"skill command": {
			command:  "skill",
			expected: true,
		},
		"skills alias": {
			command:  "skills",
			expected: true,
		},
		"skill flag": {
			command:  "--skill",
			expected: true,
		},
		"skill flag with flags": {
			command:  "--skill --verbose",
			expected: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if tc.command != "" {
				os.Args = append([]string{"placeholder"}, strings.Split(tc.command, " ")...)
			} else {
				os.Args = []string{"placeholder"}
			}
			// Test
			updateNotification := &UpdateNotification{}
			actual := updateNotification.isIgnoredCommand()
			require.Equal(t, tc.expected, actual)
		})
	}
}
