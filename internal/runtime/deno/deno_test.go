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

package deno

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Deno_New(t *testing.T) {
	tests := map[string]struct {
		expectedDeno *Deno
	}{
		"New Deno instance": {
			expectedDeno: &Deno{version: defaultVersion},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := New()
			require.Equal(t, tc.expectedDeno, d)
		})
	}
}

func Test_Deno_IgnoreDirectories(t *testing.T) {
	tests := map[string]struct {
		expectedIgnoreDirectories []string
	}{
		"No directories": {
			expectedIgnoreDirectories: []string{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := New()
			require.Equal(t, tc.expectedIgnoreDirectories, d.IgnoreDirectories())
		})
	}
}

func Test_Deno_InstallProjectDependencies(t *testing.T) {
	tests := map[string]struct {
		projectDirPath            string
		lookPathError             error
		hookExecutorError         error
		existingFilePaths         []string
		expectedError             error
		expectedResponse          string
		expectedHookExecutorCalls int
	}{
		"Deno executable not found": {
			projectDirPath:            "/path/to/project-name",
			lookPathError:             exec.ErrNotFound,
			hookExecutorError:         nil,
			existingFilePaths:         []string{},
			expectedError:             slackerror.Wrap(exec.ErrNotFound, slackerror.ErrDenoNotFound),
			expectedResponse:          "",
			expectedHookExecutorCalls: 0,
		},
		"No manifest files": {
			projectDirPath:            "/path/to/project-name",
			lookPathError:             nil,
			hookExecutorError:         nil,
			existingFilePaths:         []string{}, // No manifest files declared
			expectedError:             nil,
			expectedResponse:          "",
			expectedHookExecutorCalls: 0,
		},
		"InstallProjectDependencies cache dependencies when manifest file exists": {
			projectDirPath:            "/path/to/project-name",
			lookPathError:             nil,
			hookExecutorError:         nil,
			existingFilePaths:         []string{"manifest.ts"}, // Manifest file exists
			expectedError:             nil,
			expectedResponse:          "",
			expectedHookExecutorCalls: 1, // Cache dependencies script executed
		},
		"InstallProjectDependencies cache dependencies when multiple manifest files exist": {
			projectDirPath:            "/path/to/project-name",
			lookPathError:             nil,
			hookExecutorError:         nil,
			existingFilePaths:         []string{"manifest.ts", "manifest.js"}, // Manifest files exist
			expectedError:             nil,
			expectedResponse:          "",
			expectedHookExecutorCalls: 2, // Cache dependencies script executed multiple times (manifest.ts, manifest.json)
		},
		"InstallProjectDependencies should not error when cache dependencies fails": {
			projectDirPath:            "/path/to/project-name",
			lookPathError:             nil,
			hookExecutorError:         slackerror.New(slackerror.ErrSDKHookNotFound), // Cache dependencies script error
			existingFilePaths:         []string{"manifest.ts"},                       // Manifest file to execute hook script
			expectedError:             nil,                                           // Hook script error is ignored
			expectedResponse:          "",
			expectedHookExecutorCalls: 1, // Cache dependencies script executed
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			projectDirPath := "/path/to/project-name"

			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			os.On("LookPath", mock.Anything).Return("", tc.lookPathError)
			os.AddDefaultMocks()

			cfg := config.NewConfig(fs, os)

			ios := iostreams.NewIOStreamsMock(cfg, fs, os)
			ios.AddDefaultMocks()

			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).Return("text output", tc.hookExecutorError)

			// Create files
			for _, filePath := range tc.existingFilePaths {
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
			d := New()
			response, err := d.InstallProjectDependencies(ctx, projectDirPath, mockHookExecutor, ios, fs, os)

			// Assertions
			require.Contains(t, response, tc.expectedResponse)
			require.Equal(t, tc.expectedError, err)
			mockHookExecutor.AssertNumberOfCalls(t, "Execute", tc.expectedHookExecutorCalls)
		})
	}
}

func Test_Deno_Name(t *testing.T) {
	d := New()
	require.Equal(t, "Deno", d.Name())
}

func Test_Deno_Version(t *testing.T) {
	tests := map[string]struct {
		deno            *Deno
		expectedVersion string
	}{
		"Default version": {
			deno:            New(),
			expectedVersion: defaultVersion,
		},
		"Custom version": {
			deno:            &Deno{version: "deno@2"},
			expectedVersion: "deno@2",
		},
		"Undefined version": {
			deno:            &Deno{version: ""},
			expectedVersion: defaultVersion,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expectedVersion, tc.deno.Version())
		})
	}
}

func Test_Deno_SetVersion(t *testing.T) {
	tests := map[string]struct {
		version         string
		expectedVersion string
	}{
		"Default version": {
			version:         "",
			expectedVersion: defaultVersion,
		},
		"Custom version": {
			version:         "deno@2",
			expectedVersion: "deno@2",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := New()
			d.SetVersion(tc.version)
			require.Equal(t, tc.expectedVersion, d.Version())
		})
	}
}

func Test_Deno_HooksJSONTemplate(t *testing.T) {
	tests := map[string]struct {
		hooksJSONTemplate []byte
		expectedErrorType error
	}{
		"HooksJSONTemplate() should be valid JSON": {
			hooksJSONTemplate: New().HooksJSONTemplate(),
			expectedErrorType: nil,
		},
		"Should fail on invalid JSON": {
			hooksJSONTemplate: []byte(`}{`),
			expectedErrorType: &json.SyntaxError{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			var anyJSON map[string]interface{}

			// Test
			err := json.Unmarshal(tc.hooksJSONTemplate, &anyJSON)

			// Assertions
			require.IsType(t, tc.expectedErrorType, err)
		})
	}
}

func Test_Deno_PreparePackage(t *testing.T) {
	tests := map[string]struct {
		hookExecutorError           error
		expectedPreparePackageError error
	}{
		"Hook successful": {
			hookExecutorError:           nil,
			expectedPreparePackageError: nil,
		},
		"Hook error": {
			hookExecutorError:           slackerror.New(slackerror.ErrSDKHookInvocationFailed),
			expectedPreparePackageError: slackerror.New(slackerror.ErrSDKHookInvocationFailed),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// Setup SDKConfig
			mockSDKConfig := hooks.NewSDKConfigMock()
			mockSDKConfig.Hooks.BuildProject = hooks.HookScript{
				Name:    "BuildProject",
				Command: "./hook-script/build-project",
			}

			// Setup HookExecutor
			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).Return("text output", tc.hookExecutorError)

			// Setup
			mockOpts := types.PreparePackageOpts{}
			mockOpts.AuthTokens = "auth-token-xxx"
			mockOpts.SrcDirPath = "src/dir/path"
			mockOpts.DstDirPath = "dst/dir/path"

			// Run tests
			d := New()
			err := d.PreparePackage(ctx, mockSDKConfig, mockHookExecutor, mockOpts)

			// Assertions
			require.Equal(t, tc.expectedPreparePackageError, err)
		})
	}
}

func Test_Deno_IsRuntimeForProject(t *testing.T) {
	tests := map[string]struct {
		sdkConfigRuntime  string
		existingFilePaths []string
		expectedBool      bool
	}{
		"SDKConfig Runtime is Deno": {
			sdkConfigRuntime:  "deno",
			existingFilePaths: []string{}, // Unset to check SDKConfig
			expectedBool:      true,
		},
		"deno.json file exists": {
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"deno.json"},
			expectedBool:      true,
		},
		"deno.jsonc file exists": {
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"deno.jsonc"},
			expectedBool:      true,
		},
		"import_map.json file exists": {
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"import_map.json"},
			expectedBool:      true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			projectDirPath := "/path/to/project-name"

			// Create files
			for _, filePath := range tc.existingFilePaths {
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
			b := IsRuntimeForProject(ctx, fs, projectDirPath, hooks.SDKCLIConfig{Runtime: tc.sdkConfigRuntime})

			// Assertions
			require.Equal(t, tc.expectedBool, b)
		})
	}
}
