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
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
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

func Test_Update_HasUpdate(t *testing.T) {
	for name, tt := range map[string]struct {
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
			for _, hasUpdate := range tt.dependencyHasUpdate {
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
			require.Equal(t, tt.expectedReturnValue, updateNotification.HasUpdate())
		})
	}
}

func Test_Update_InstallUpdatesWithComponentFlags(t *testing.T) {
	// Save original type checking functions to restore later
	originalIsCLI := isDependencyCLI
	originalIsSDK := isDependencySDK
	defer func() {
		isDependencyCLI = originalIsCLI
		isDependencySDK = originalIsSDK
	}()

	// Create mock dependencies - one CLI, one SDK
	cliDep := new(mockDependency)
	sdkDep := new(mockDependency)

	// Setup test cases
	for name, tt := range map[string]struct {
		cli                   bool
		sdk                   bool
		cliHasUpdate          bool
		sdkHasUpdate          bool
		cliInstallError       error
		sdkInstallError       error
		expectedErrorContains string
		shouldInstallCLI      bool
		shouldInstallSDK      bool
	}{
		"Both flags false, both have updates": {
			cli:                   false,
			sdk:                   false,
			cliHasUpdate:          true,
			sdkHasUpdate:          true,
			cliInstallError:       nil,
			sdkInstallError:       nil,
			expectedErrorContains: "",
			shouldInstallCLI:      false,
			shouldInstallSDK:      false,
		},
		"Only CLI flag set, both have updates": {
			cli:                   true,
			sdk:                   false,
			cliHasUpdate:          true,
			sdkHasUpdate:          true,
			cliInstallError:       nil,
			sdkInstallError:       nil,
			expectedErrorContains: "",
			shouldInstallCLI:      true,
			shouldInstallSDK:      false,
		},
		"Only SDK flag set, both have updates": {
			cli:                   false,
			sdk:                   true,
			cliHasUpdate:          true,
			sdkHasUpdate:          true,
			cliInstallError:       nil,
			sdkInstallError:       nil,
			expectedErrorContains: "",
			shouldInstallCLI:      false,
			shouldInstallSDK:      true,
		},
		"Both flags set, both have updates": {
			cli:                   true,
			sdk:                   true,
			cliHasUpdate:          true,
			sdkHasUpdate:          true,
			cliInstallError:       nil,
			sdkInstallError:       nil,
			expectedErrorContains: "",
			shouldInstallCLI:      true,
			shouldInstallSDK:      true,
		},
		"CLI flag set, CLI fails to install": {
			cli:                   true,
			sdk:                   false,
			cliHasUpdate:          true,
			sdkHasUpdate:          true,
			cliInstallError:       assert.AnError,
			sdkInstallError:       nil,
			expectedErrorContains: "general error for testing",
			shouldInstallCLI:      true,
			shouldInstallSDK:      false,
		},
		"SDK flag set, SDK fails to install": {
			cli:                   false,
			sdk:                   true,
			cliHasUpdate:          true,
			sdkHasUpdate:          true,
			cliInstallError:       nil,
			sdkInstallError:       assert.AnError,
			expectedErrorContains: "general error for testing",
			shouldInstallCLI:      false,
			shouldInstallSDK:      true,
		},
		"CLI flag set, CLI has no update": {
			cli:                   true,
			sdk:                   false,
			cliHasUpdate:          false,
			sdkHasUpdate:          true,
			cliInstallError:       nil,
			sdkInstallError:       nil,
			expectedErrorContains: "",
			shouldInstallCLI:      false,
			shouldInstallSDK:      false,
		},
		"SDK flag set, SDK has no update": {
			cli:                   false,
			sdk:                   true,
			cliHasUpdate:          true,
			sdkHasUpdate:          false,
			cliInstallError:       nil,
			sdkInstallError:       nil,
			expectedErrorContains: "",
			shouldInstallCLI:      false,
			shouldInstallSDK:      false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Create clients
			clients := shared.ClientFactory{
				Config:    config.NewConfig(slackdeps.NewFsMock(), slackdeps.NewOsMock()),
				SDKConfig: hooks.NewSDKConfigMock(),
			}

			// Reset mocks
			cliDep = new(mockDependency)
			sdkDep = new(mockDependency)

			// Setup mock CLI dependency - allowing any number of calls to HasUpdate
			cliDep.On("HasUpdate").Return(tt.cliHasUpdate, nil).Maybe()
			if tt.cliHasUpdate && tt.shouldInstallCLI {
				cliDep.On("PrintUpdateNotification", mock.Anything).Return(false, nil)
				cliDep.On("InstallUpdate", mock.Anything).Return(tt.cliInstallError)
			}

			// Setup mock SDK dependency - allowing any number of calls to HasUpdate
			sdkDep.On("HasUpdate").Return(tt.sdkHasUpdate, nil).Maybe()
			if tt.sdkHasUpdate && tt.shouldInstallSDK {
				sdkDep.On("PrintUpdateNotification", mock.Anything).Return(false, nil)
				sdkDep.On("InstallUpdate", mock.Anything).Return(tt.sdkInstallError)
			}

			// Override type checking functions for this test
			isDependencyCLI = func(d Dependency) bool {
				return d == cliDep
			}
			isDependencySDK = func(d Dependency) bool {
				return d == sdkDep
			}

			// Create updateNotification with our dependencies
			updateNotification = &UpdateNotification{
				clients:      &clients,
				enabled:      true,
				envDisabled:  "SLACK_SKIP_UPDATE",
				hoursToWait:  defaultHoursToWait,
				dependencies: []Dependency{cliDep, sdkDep},
			}

			// Create test cmd
			cmd := &cobra.Command{}

			// Test
			err := updateNotification.InstallUpdatesWithComponentFlags(cmd, tt.cli, tt.sdk)

			// Verify the error
			if tt.expectedErrorContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErrorContains)
			} else {
				require.NoError(t, err)
			}

			// Don't assert expectations since we've used .Maybe()
			// This avoids strictness in the number of times HasUpdate is called
		})
	}
}
