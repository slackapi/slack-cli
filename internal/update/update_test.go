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
