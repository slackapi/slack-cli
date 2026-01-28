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
	"context"
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/cobra"
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

func Test_Update_HasUpdate(t *testing.T) {
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
			var enabled = true
			if clients.Config.SkipUpdateFlag || clients.Config.TokenFlag != "" {
				enabled = false
			}
			// Create updateNotification
			updateNotification = &UpdateNotification{
				clients:      &clients,
				enabled:      enabled,
				envDisabled:  "SLACK_SKIP_UPDATE",
				hoursToWait:  defaultHoursToWait,
				dependencies: dependencies,
			}

			// Test
			require.Equal(t, tc.expectedReturnValue, updateNotification.HasUpdate())
		})
	}
}

func Test_Update_isIgnoredCommand(t *testing.T) {
	for name, tc := range map[string]struct {
		command  string
		expected bool
	}{
		"No command": {
			command:  "",
			expected: false,
		},
		"fingerprint command": {
			command:  "_fingerprint",
			expected: true,
		},
		"version command": {
			command:  "version",
			expected: true,
		},
		"auth command": {
			command:  "auth",
			expected: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if tc.command != "" {
				os.Args = []string{"placeholder", tc.command}
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
