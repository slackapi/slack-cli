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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var updateScenarios = []struct {
	Name           string
	SDKReleaseInfo SDKReleaseInfo
}{
	{
		Name: "No updates available",
		SDKReleaseInfo: SDKReleaseInfo{
			Message:  "",
			Releases: []SDKReleaseComponent{},
			Error: struct {
				Message string "json:\"message\""
			}{Message: ""},
		},
	},
	{
		Name: "Update without breaking changes",
		SDKReleaseInfo: SDKReleaseInfo{
			Message: "Update without breaking change",
			Releases: []SDKReleaseComponent{
				{
					Name:     "test_dependency_1",
					Current:  "1.0.0",
					Latest:   "1.0.1",
					Breaking: false,
					Update:   true,
					Error: struct {
						Message string "json:\"message\""
					}{Message: ""},
				},
			},
			Error: struct {
				Message string "json:\"message\""
			}{Message: ""},
		},
	},
	{
		Name: "Update with breaking changes",
		SDKReleaseInfo: SDKReleaseInfo{
			Message: "Update with breaking change",
			Releases: []SDKReleaseComponent{
				{
					Name:     "test_dependency_1",
					Current:  "1.0.0",
					Latest:   "2.0.0",
					Breaking: true,
					Update:   true,
					Error: struct {
						Message string "json:\"message\""
					}{Message: ""},
				},
			},
			Error: struct {
				Message string "json:\"message\""
			}{Message: ""},
		},
	},
	{
		Name: "Update with occurrence of error",
		SDKReleaseInfo: SDKReleaseInfo{
			Message: "Update with occurrence of error",
			Releases: []SDKReleaseComponent{
				{
					Name:     "test_dependency_1",
					Current:  "1.0.0",
					Latest:   "1.0.1",
					Breaking: false,
					Update:   true,
					Error: struct {
						Message string "json:\"message\""
					}{Message: ""},
				},
				{
					Name:     "test_dependency_2",
					Current:  "1.0.0",
					Latest:   "",
					Breaking: false,
					Update:   false,
					Error: struct {
						Message string "json:\"message\""
					}{Message: "Test error message (test_dependency_2)"},
				},
			},
			Error: struct {
				Message string "json:\"message\""
			}{Message: "Test error message"},
		},
	},
	{
		Name: "No update with occurrence of error",
		SDKReleaseInfo: SDKReleaseInfo{
			Message: "No update with occurrence of error",
			Releases: []SDKReleaseComponent{
				{
					Name:     "test_dependency_1",
					Current:  "1.0.0",
					Latest:   "1.0.0",
					Breaking: false,
					Update:   false,
					Error: struct {
						Message string "json:\"message\""
					}{Message: ""},
				},
				{
					Name:     "test_dependency_2",
					Current:  "1.0.0",
					Latest:   "",
					Breaking: false,
					Update:   false,
					Error: struct {
						Message string "json:\"message\""
					}{Message: "Test error message (test_dependency_2)"},
				},
			},
			Error: struct {
				Message string "json:\"message\""
			}{Message: "Test error message"},
		},
	},
}

var installScenarios = []struct {
	Name                     string
	SDKInstallUpdateResponse SDKInstallUpdateResponse
}{
	{
		Name: "No updates available",
		SDKInstallUpdateResponse: SDKInstallUpdateResponse{
			Name:    "the Slack SDK",
			Updates: []SDKInstallUpdateComponent{},
			Error: struct {
				Message string `json:"message,omitempty"`
			}{Message: ""},
		},
	},
	{
		Name: "Updates without errors",
		SDKInstallUpdateResponse: SDKInstallUpdateResponse{
			Name: "the Slack SDK",
			Updates: []SDKInstallUpdateComponent{
				{
					Name:             "test_dependency_1",
					PreviousVersion:  "1.0.0",
					InstalledVersion: "1.0.1",
					Error: struct {
						Message string `json:"message,omitempty"`
					}{Message: ""},
				},
			},
			Error: struct {
				Message string `json:"message,omitempty"`
			}{Message: ""},
		},
	},
	{
		Name: "Updates with installation errors",
		SDKInstallUpdateResponse: SDKInstallUpdateResponse{
			Name: "the Slack SDK",
			Updates: []SDKInstallUpdateComponent{
				{
					Name:             "test_dependency_1",
					PreviousVersion:  "1.0.0",
					InstalledVersion: "1.0.1",
					Error: struct {
						Message string `json:"message,omitempty"`
					}{Message: "An error occurred"},
				},
				{
					Name:             "test_dependency_2",
					PreviousVersion:  "1.0.0",
					InstalledVersion: "1.0.1",
					Error: struct {
						Message string `json:"message,omitempty"`
					}{Message: "An error occurred while installing the update"},
				},
			},
			Error: struct {
				Message string `json:"message,omitempty"`
			}{Message: "There was an error installing one or more updates"},
		},
	},
	{
		Name: "Updates with no installation errors, but top-level error",
		SDKInstallUpdateResponse: SDKInstallUpdateResponse{
			Name: "the Slack SDK",
			Updates: []SDKInstallUpdateComponent{
				{
					Name:             "test_dependency_1",
					PreviousVersion:  "1.0.0",
					InstalledVersion: "1.0.1",
					Error: struct {
						Message string `json:"message,omitempty"`
					}{Message: ""},
				},
				{
					Name:             "test_dependency_2",
					PreviousVersion:  "1.0.0",
					InstalledVersion: "1.0.1",
					Error: struct {
						Message string `json:"message,omitempty"`
					}{Message: ""},
				},
			},
			Error: struct {
				Message string `json:"message,omitempty"`
			}{Message: "There was an error building the project"},
		},
	},
}

func Test_SDK_NewSDKDependency(t *testing.T) {
	clients := shared.ClientFactory{SDKConfig: hooks.NewSDKConfigMock()}
	sdkDependency := NewSDKDependency(&clients)
	assert.Equal(t, &SDKDependency{clients: &clients}, sdkDependency)
}

func Test_SDK_HasUpdate(t *testing.T) {
	for _, s := range updateScenarios {
		t.Run(s.Name, func(t *testing.T) {

			SDKDependencyMock := SDKDependency{
				releaseInfo: s.SDKReleaseInfo,
			}

			hasUpdate, err := SDKDependencyMock.HasUpdate()
			if err != nil {
				assert.Contains(t, err.Error(), s.SDKReleaseInfo.Error.Message)
			}

			if s.SDKReleaseInfo.Error.Message == "" {
				assert.Equal(t, nil, err)
			}

			assert.Equal(t, s.SDKReleaseInfo.Update, hasUpdate)
		})
	}
}

// TODO :: Test cases need to be implemented once it's possible
// to read the output from fmt.Print or by refactoring the code
// to produce a returned structure that is printed elsewhere
func Test_SDK_InstallUpdate(t *testing.T) {
	for _, s := range installScenarios {
		// Create mocks
		clientsMock := shared.NewClientsMock()

		// Create clients that is mocked for testing
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		mockInstallUpdateJSON, _ := json.Marshal(s.SDKInstallUpdateResponse)

		// Mock the returned value from executing the `install-update` hook
		mockInstallUpdateHook := hooks.HookScript{Command: fmt.Sprintf(`echo %s`, string(mockInstallUpdateJSON))}
		clients.SDKConfig.Hooks.InstallUpdate = mockInstallUpdateHook
		clientsMock.HookExecutor.On("Execute", mock.Anything).Return(string(mockInstallUpdateJSON), nil)

		// Execute `install-update` hook
		_, err := clients.HookExecutor.Execute(hooks.HookExecOpts{Hook: clients.SDKConfig.Hooks.InstallUpdate})
		if err != nil {
			assert.Fail(t, "Running the `install-update` encountered an unexpected error")
		}
		// Create the command
		cmd := &cobra.Command{}
		testutil.MockCmdIO(clients.IO, cmd)
		ctx := cmd.Context()

		// output := clientsMock.GetCombinedOutput()

		// For each installation update in scenario
		t.Run(s.Name, func(t *testing.T) {

			SDKDependencyMock := SDKDependency{
				clients: clients,
			}

			err := SDKDependencyMock.InstallUpdate(ctx)
			if err != nil {
				assert.Fail(t, "InstallUpdate had unexpected error")
			}

			clientsMock.HookExecutor.AssertCalled(t, "Execute", mock.Anything)

			// TODO :: Test Case: `install-update` hook is available
			// == TODO :: Assert:  Updates are present; printed output contains updates

			// TODO :: Test Case: Updates are not present; no printed output
			// == TODO :: Assert: `install-update` hook is not available
			// == TODO :: Assert: .Execute() was not called
		})
	}
}

func Test_SDK_PrintUpdateNotification(t *testing.T) {
	for i, s := range updateScenarios {
		// Create mocks
		clientsMock := shared.NewClientsMock()

		// Create clients that is mocked for testing
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())

		// Mock SDKConfig `install-update` hook script
		clients.SDKConfig.Hooks.InstallUpdate = hooks.HookScript{}

		// Create the command
		cmd := &cobra.Command{}
		testutil.MockCmdIO(clients.IO, cmd)

		t.Run(s.Name, func(t *testing.T) {

			SDKDependencyMock := SDKDependency{
				clients:     clients,
				releaseInfo: s.SDKReleaseInfo,
			}

			printResp, err := SDKDependencyMock.PrintUpdateNotification(cmd)
			if err != nil {
				assert.Fail(t, "PrintUpdateNotification had unexpected error")
			}

			err = cmd.Execute()
			if err != nil {
				assert.Fail(t, "cmd.Execute had unexpected error")
			}
			output := clientsMock.GetCombinedOutput()

			assert.Contains(t, output, updateScenarios[i].SDKReleaseInfo.Message)

			// Updates, not breaking
			if SDKDependencyMock.releaseInfo.Update && !SDKDependencyMock.releaseInfo.Breaking {
				assert.NotContains(t, output, "Warning: this update contains a breaking change!")
			}

			// Updates, breaking
			if SDKDependencyMock.releaseInfo.Update && SDKDependencyMock.releaseInfo.Breaking {
				assert.Contains(t, output, "Warning: this update contains a breaking change!")
			}

			// Updates, error
			if SDKDependencyMock.releaseInfo.Update && SDKDependencyMock.releaseInfo.Error.Message != "" {
				assert.Contains(t, output, "Error:")
				assert.Contains(t, output, SDKDependencyMock.releaseInfo.Error.Message)
			}

			// No updates, error
			if !SDKDependencyMock.releaseInfo.Update && SDKDependencyMock.releaseInfo.Error.Message != "" {
				assert.Contains(t, output, "Error:")
				assert.Contains(t, output, SDKDependencyMock.releaseInfo.Error.Message)
			}

			assert.Equal(t, false, printResp)
		})
	}
}
