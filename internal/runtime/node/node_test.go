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

package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Node_New(t *testing.T) {
	tests := []struct {
		name         string
		expectedNode *Node
	}{
		{
			name:         "New Node instance",
			expectedNode: &Node{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New()
			require.IsType(t, Node{}, *n)
		})
	}
}

func Test_Node_IgnoreDirectories(t *testing.T) {
	tests := []struct {
		name                      string
		expectedIgnoreDirectories []string
	}{
		{
			name:                      "Ignore node modules",
			expectedIgnoreDirectories: []string{"node_modules"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New()
			require.Equal(t, tt.expectedIgnoreDirectories, n.IgnoreDirectories())
		})
	}
}

func Test_Node_InstallProjectDependencies(t *testing.T) {
	tests := []struct {
		name                   string
		hookExecutorError      []error
		hookInstallationStdout string
		hookInstallationStderr string
		expectedDebug          string
		expectedError          error
		expectedOutputs        []string
		npmMock                func() NPM
	}{
		{
			name: "When @slack/cli-hooks found then skip installing it and continue",
			expectedOutputs: []string{
				"Found package @slack/cli-hooks@1.1.2",
				"Installed dependencies",
			},
			expectedError: nil,
			npmMock: func() NPM {
				npmMock := &NPMMock{}
				// @slack/cli-hooks found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("@slack/cli-hooks@1.1.2", true).Once()
				// npm install is successful
				npmMock.On("InstallAllPackages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil).Once()
				return npmMock
			},
		},
		{
			name: "When @slack/cli-hooks not found then install it and continue",
			expectedOutputs: []string{
				"Added package @slack/cli-hooks@1.1.2",
				"Installed dependencies",
			},
			expectedError: nil,
			npmMock: func() NPM {
				npmMock := &NPMMock{}
				// @slack/cli-hooks not found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", false).Once()
				// @slack/cli-hooks installed
				npmMock.On("InstallDevPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil).Once()
				// @slack/cli-hooks found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("@slack/cli-hooks@1.1.2", true)
				// npm install is successful
				npmMock.On("InstallAllPackages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				return npmMock
			},
		},
		{
			name: "When @slack/cli-hooks install fails then error and continue",
			expectedOutputs: []string{
				"Error adding package @slack/cli-hooks",
				"Installed dependencies",
			},
			expectedError: errors.New("super error"),
			npmMock: func() NPM {
				npmMock := &NPMMock{}
				// @slack/cli-hooks not found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", false).Once()
				// @slack/cli-hooks install error
				npmMock.On("InstallDevPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("super error")).Once()
				// @slack/cli-hooks not found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", false).Once()
				// npm install is successful
				npmMock.On("InstallAllPackages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				return npmMock
			},
		},
		{
			name: "When npm install successful",
			expectedOutputs: []string{
				"Found package @slack/cli-hooks@1.1.2",
				"Installed dependencies",
			},
			expectedError: nil,
			npmMock: func() NPM {
				npmMock := &NPMMock{}
				// @slack/cli-hooks found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("@slack/cli-hooks@1.1.2", true).Once()
				// npm install is successful
				npmMock.On("InstallAllPackages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil).Once()
				return npmMock
			},
		},
		{
			name: "When npm install fails return error",
			expectedOutputs: []string{
				"Found package @slack/cli-hooks@1.1.2",
				"Error installing dependencies",
			},
			expectedError: errors.New("npm install error"),
			npmMock: func() NPM {
				npmMock := &NPMMock{}
				// @slack/cli-hooks found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("@slack/cli-hooks@1.1.2", true).Once()
				// npm install error
				npmMock.On("InstallAllPackages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("npm install error")).Once()
				return npmMock
			},
		},
		{
			name: "When @slack/cli-hooks and npm install both fail return first error",
			expectedOutputs: []string{
				"Error adding package @slack/cli-hooks",
				"Error installing dependencies",
			},
			expectedError: errors.New("@slack/cli-hooks error"),
			npmMock: func() NPM {
				npmMock := &NPMMock{}
				// @slack/cli-hooks not found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", false).Once()
				// @slack/cli-hooks install error
				npmMock.On("InstallDevPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("@slack/cli-hooks error")).Once()
				// @slack/cli-hooks not found
				npmMock.On("ListPackage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", false).Once()
				// npm install error
				npmMock.On("InstallAllPackages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("npm install error")).Once()
				return npmMock
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			projectDirPath := "/path/to/project-name"

			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			cfg := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(cfg, fs, os)

			mockHookExecutor := &hooks.MockHookExecutor{}
			npmMock := tt.npmMock()

			// Test
			n := New()
			n.npmClient = npmMock
			output, err := n.InstallProjectDependencies(ctx, projectDirPath, mockHookExecutor, ios, fs, os)

			// Assertions
			for _, expectedOutput := range tt.expectedOutputs {
				require.Contains(t, output, expectedOutput)
			}
			require.Equal(t, tt.expectedError, err)
		})
	}
}

func Test_Node_Name(t *testing.T) {
	n := New()
	require.Equal(t, "Node.js", n.Name())
}

func Test_Node_Version(t *testing.T) {
	tests := []struct {
		name            string
		expectedVersion string
	}{
		{
			name:            "Default version",
			expectedVersion: "node",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New()
			require.Equal(t, tt.expectedVersion, n.Version())
		})
	}
}

func Test_Node_SetVersion(t *testing.T) {
	// Unsupported feature, but calling for test coverage
	n := New()
	n.SetVersion("unsupported")
	require.True(t, true)
}

func Test_Node_HooksJSONTemplate(t *testing.T) {
	tests := []struct {
		name              string
		hooksJSONTemplate []byte
		expectedErrorType error
	}{
		{
			name:              "HooksJSONTemplate() should be valid JSON",
			hooksJSONTemplate: New().HooksJSONTemplate(),
			expectedErrorType: nil,
		},
		{
			name:              "Should fail on invalid JSON",
			hooksJSONTemplate: []byte(`}{`),
			expectedErrorType: &json.SyntaxError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var anyJSON map[string]interface{}

			// Test
			err := json.Unmarshal(tt.hooksJSONTemplate, &anyJSON)

			// Assertions
			require.IsType(t, tt.expectedErrorType, err)
		})
	}
}

func Test_Node_PreparePackage(t *testing.T) {
	tests := []struct {
		name                        string
		hookExecutorError           error
		expectedPreparePackageError error
	}{
		{
			name:                        "Should return no error because unsupported",
			hookExecutorError:           nil,
			expectedPreparePackageError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// Setup SDKConfig
			mockSDKConfig := hooks.NewSDKConfigMock()
			mockSDKConfig.Hooks.BuildProject = hooks.HookScript{
				Name:    "BuildProject",
				Command: "./hook-script/build-project",
			}

			// Setup HookExecutor
			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).Return("text output", tt.hookExecutorError)

			// Setup
			mockOpts := types.PreparePackageOpts{}
			mockOpts.AuthTokens = "auth-token-xxx"
			mockOpts.SrcDirPath = "src/dir/path"
			mockOpts.DstDirPath = "dst/dir/path"

			// Run tests
			d := New()
			err := d.PreparePackage(ctx, mockSDKConfig, mockHookExecutor, mockOpts)

			// Assertions
			require.Equal(t, tt.expectedPreparePackageError, err)
		})
	}
}

func Test_Node_IsRuntimeForProject(t *testing.T) {
	tests := []struct {
		name              string
		sdkConfigRuntime  string
		existingFilePaths []string
		expectedBool      bool
	}{
		{
			name:              "Not a Node.js project",
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{},
			expectedBool:      false,
		},
		{
			name:              "SDKConfig Runtime is Node.js",
			sdkConfigRuntime:  "node",
			existingFilePaths: []string{}, // Unset to check SDKConfig
			expectedBool:      true,
		},
		{
			name:              "package.json file exists",
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"package.json"},
			expectedBool:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			projectDirPath := "/path/to/project-name"

			// Create files
			for _, filePath := range tt.existingFilePaths {
				filePathAbs := filepath.Join(projectDirPath, filePath)
				// Create the directory
				if err := fs.MkdirAll(filepath.Dir(filePathAbs), 0755); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", filePath))
				}
				// Create the file
				if err := afero.WriteFile(fs, filePathAbs, []byte("mock file data"), 0644); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to create the file %s in the memory-based file system", filePath))
				}
			}

			// Test
			b := IsRuntimeForProject(ctx, fs, projectDirPath, hooks.SDKCLIConfig{Runtime: tt.sdkConfigRuntime})

			// Assertions
			require.Equal(t, tt.expectedBool, b)
		})
	}
}
