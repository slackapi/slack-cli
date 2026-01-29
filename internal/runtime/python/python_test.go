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

package python

import (
	"encoding/json"
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

func Test_Python_New(t *testing.T) {
	tests := []struct {
		name           string
		expectedPython *Python
	}{
		{
			name:           "New Python instance",
			expectedPython: &Python{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			require.Equal(t, tt.expectedPython, p)
		})
	}
}

func Test_Python_IgnoreDirectories(t *testing.T) {
	tests := []struct {
		name                      string
		expectedIgnoreDirectories []string
	}{
		{
			name:                      "No directories",
			expectedIgnoreDirectories: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			require.Equal(t, tt.expectedIgnoreDirectories, p.IgnoreDirectories())
		})
	}
}

func Test_Python_InstallProjectDependencies(t *testing.T) {
	tests := []struct {
		name               string
		existingFiles      map[string]string
		expectedFiles      map[string]string
		expectedOutputs    []string
		notExpectedOutputs []string
		expectedError      bool
	}{
		{
			name:            "Error when requirements.txt is missing",
			existingFiles:   map[string]string{}, // No files
			expectedOutputs: []string{"Error"},
			expectedError:   true,
		},
		{
			name: "Skip when requirements.txt contains slack-cli-hooks",
			existingFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks\npytest==8.3.2\nruff==0.7.2",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks\npytest==8.3.2\nruff==0.7.2",
			},
			expectedOutputs: []string{"Found"},
			expectedError:   false,
		},
		{
			name: "Skip when requirements.txt contains slack-cli-hooks<1.0.0",
			existingFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks<1.0.0\npytest==8.3.2\nruff==0.7.2",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks<1.0.0\npytest==8.3.2\nruff==0.7.2",
			},
			expectedOutputs: []string{"Found"},
			expectedError:   false,
		},
		{
			name: "Update when requirements.txt contain slack-bolt at top of file",
			existingFiles: map[string]string{
				"requirements.txt": "slack-bolt==2.31.2\npytest==8.3.2\nruff==0.7.2",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "slack-bolt==2.31.2\nslack-cli-hooks<1.0.0\npytest==8.3.2\nruff==0.7.2",
			},
			expectedOutputs: []string{"Updated"},
			expectedError:   false,
		},
		{
			name: "Update when requirements.txt contain slack-bolt at middle of file",
			existingFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nslack-bolt==2.31.2\nruff==0.7.2",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nslack-bolt==2.31.2\nslack-cli-hooks<1.0.0\nruff==0.7.2",
			},
			expectedOutputs: []string{"Updated"},
			expectedError:   false,
		},
		{
			name: "Update when requirements.txt contain slack-bolt at bottom of file",
			existingFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nruff==0.7.2\nslack-bolt==2.31.2",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nruff==0.7.2\nslack-bolt==2.31.2\nslack-cli-hooks<1.0.0",
			},
			expectedOutputs: []string{"Updated"},
			expectedError:   false,
		},
		{
			name: "Update when requirements.txt does not contain slack-bolt",
			existingFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nruff==0.7.2",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nruff==0.7.2\nslack-cli-hooks<1.0.0",
			},
			expectedOutputs: []string{"Updated"},
			expectedError:   false,
		},
		{
			name: "Update when requirements.txt with trailing whitespace does not contain slack-bolt",
			existingFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nruff==0.7.2\n\n\n    ",
			},
			expectedFiles: map[string]string{
				"requirements.txt": "pytest==8.3.2\nruff==0.7.2\nslack-cli-hooks<1.0.0",
			},
			expectedOutputs: []string{"Updated"},
			expectedError:   false,
		},
		{
			name: "Should output help text because installing project dependencies is unsupported",
			existingFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks\npytest==8.3.2\nruff==0.7.2",
			},
			expectedOutputs: []string{"Manually setup a Python virtual environment"},
			expectedError:   false,
		},
		{
			name: "Should output pip install -r requirements.txt when only requirements.txt exists",
			existingFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks\npytest==8.3.2",
			},
			expectedOutputs:    []string{"pip install -r requirements.txt"},
			notExpectedOutputs: []string{"pip install -e ."},
			expectedError:      false,
		},
		{
			name: "Should output pip install -e . when only pyproject.toml exists",
			existingFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = ["slack-cli-hooks<1.0.0"]`,
			},
			expectedOutputs:    []string{"pip install -e ."},
			notExpectedOutputs: []string{"pip install -r requirements.txt"},
			expectedError:      false,
		},
		{
			name: "Should output both install commands when both files exist",
			existingFiles: map[string]string{
				"requirements.txt": "slack-cli-hooks\npytest==8.3.2",
				"pyproject.toml": `[project]
name = "my-app"
dependencies = ["slack-cli-hooks<1.0.0"]`,
			},
			expectedOutputs: []string{"pip install -r requirements.txt", "pip install -e ."},
			expectedError:   false,
		},
		{
			name: "Error when neither requirements.txt nor pyproject.toml exists",
			existingFiles: map[string]string{
				"main.py": "# some python code",
			},
			expectedOutputs: []string{"Error: no Python dependency file found"},
			expectedError:   true,
		},
		{
			name: "Skip when pyproject.toml contains slack-cli-hooks",
			existingFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "slack-cli-hooks<1.0.0",
    "pytest==8.3.2",
]`,
			},
			expectedFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "slack-cli-hooks<1.0.0",
    "pytest==8.3.2",
]`,
			},
			expectedOutputs: []string{"Found pyproject.toml"},
			expectedError:   false,
		},
		{
			name: "Update when pyproject.toml contains slack-bolt",
			existingFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "slack-bolt>=1.0.0",
    "pytest==8.3.2",
]`,
			},
			expectedFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "slack-bolt>=1.0.0",
    "pytest==8.3.2",
    "slack-cli-hooks<1.0.0",
]`,
			},
			expectedOutputs: []string{"Updated pyproject.toml"},
			expectedError:   false,
		},
		{
			name: "Update when pyproject.toml does not contain slack-bolt",
			existingFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "pytest==8.3.2",
]`,
			},
			expectedFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "pytest==8.3.2",
    "slack-cli-hooks<1.0.0",
]`,
			},
			expectedOutputs: []string{"Updated pyproject.toml"},
			expectedError:   false,
		},
		{
			name: "Update both requirements.txt and pyproject.toml when both exist",
			existingFiles: map[string]string{
				"requirements.txt": "slack-bolt==2.31.2\npytest==8.3.2",
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "slack-bolt>=1.0.0",
]`,
			},
			expectedFiles: map[string]string{
				"requirements.txt": "slack-bolt==2.31.2\nslack-cli-hooks<1.0.0\npytest==8.3.2",
				"pyproject.toml": `[project]
name = "my-app"
dependencies = [
    "slack-bolt>=1.0.0",
    "slack-cli-hooks<1.0.0",
]`,
			},
			expectedOutputs: []string{"Updated requirements.txt"},
			expectedError:   false,
		},
		{
			name: "Error when pyproject.toml has no dependencies array",
			existingFiles: map[string]string{
				"pyproject.toml": `[project]
name = "my-app"`,
			},
			expectedOutputs: []string{"Error: pyproject.toml missing dependencies array"},
			expectedError:   true,
		},
		{
			name: "Error when pyproject.toml has no [project] section",
			existingFiles: map[string]string{
				"pyproject.toml": `[tool.black]
line-length = 88`,
			},
			expectedOutputs: []string{"Error: pyproject.toml missing project section"},
			expectedError:   true,
		},
		{
			name: "Error when pyproject.toml is invalid TOML",
			existingFiles: map[string]string{
				"pyproject.toml": `[project
name = "broken`,
			},
			expectedOutputs: []string{"Error parsing pyproject.toml"},
			expectedError:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			cfg := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(cfg, fs, os)

			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).Return("text output", nil)

			projectDirPath := "/path/to/project-name"

			// Create files
			for filePath, fileData := range tt.existingFiles {
				filePathAbs := filepath.Join(projectDirPath, filePath)
				// Create the directory
				if err := fs.MkdirAll(filepath.Dir(filePathAbs), 0755); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", filePath))
				}
				// Create the file
				if err := afero.WriteFile(fs, filePathAbs, []byte(fileData), 0644); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to create the file %s in the memory-based file system", filePath))
				}
			}

			// Test
			p := New()
			outputs, err := p.InstallProjectDependencies(ctx, projectDirPath, mockHookExecutor, ios, fs, os)

			// Assertions
			for filePath, fileData := range tt.expectedFiles {
				filePathAbs := filepath.Join(projectDirPath, filePath)
				d, err := afero.ReadFile(fs, filePathAbs)
				require.NoError(t, err)
				require.Equal(t, fileData, string(d))
			}

			for _, expected := range tt.expectedOutputs {
				require.Contains(t, outputs, expected)
			}

			for _, notExpected := range tt.notExpectedOutputs {
				require.NotContains(t, outputs, notExpected)
			}

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_Python_Name(t *testing.T) {
	p := New()
	require.Equal(t, "Python", p.Name())
}

func Test_Python_Version(t *testing.T) {
	tests := []struct {
		name            string
		expectedVersion string
	}{
		{
			name:            "Default version",
			expectedVersion: "python",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			require.Equal(t, tt.expectedVersion, p.Version())
		})
	}
}

func Test_Python_SetVersion(t *testing.T) {
	// Unsupported feature, but calling for test coverage
	p := New()
	p.SetVersion("unsupported")
	require.True(t, true)
}

func Test_Python_HooksJSONTemplate(t *testing.T) {
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

func Test_Python_PreparePackage(t *testing.T) {
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
			p := New()
			err := p.PreparePackage(ctx, mockSDKConfig, mockHookExecutor, mockOpts)

			// Assertions
			require.Equal(t, tt.expectedPreparePackageError, err)
		})
	}
}

func Test_Python_IsRuntimeForProject(t *testing.T) {
	tests := []struct {
		name              string
		sdkConfigRuntime  string
		existingFilePaths []string
		expectedBool      bool
	}{
		{
			name:              "Not a Python project",
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{},
			expectedBool:      false,
		},
		{
			name:              "SDKConfig Runtime is Python",
			sdkConfigRuntime:  "python",
			existingFilePaths: []string{}, // Unset to check SDKConfig
			expectedBool:      true,
		},
		{
			name:              "requirements.txt file exists",
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"requirements.txt"},
			expectedBool:      true,
		},
		{
			name:              "pyproject.toml file exists",
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"pyproject.toml"},
			expectedBool:      true,
		},
		{
			name:              "both requirements.txt and pyproject.toml exist",
			sdkConfigRuntime:  "", // Unset to check for file
			existingFilePaths: []string{"requirements.txt", "pyproject.toml"},
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

func Test_Python_getProjectDirRelPath(t *testing.T) {
	tests := map[string]struct {
		currentDirPath  string
		projectDirPath  string
		getwdPath       string
		getwdError      error
		expectedRelPath string
		expectedError   error
	}{
		"When currentDirPath missing and Getwd returns an error": {
			currentDirPath:  "",
			projectDirPath:  "path/to/my-project",
			getwdPath:       "",
			getwdError:      fmt.Errorf("Something went wrong"),
			expectedRelPath: "path/to/my-project",
			expectedError:   fmt.Errorf("Something went wrong"),
		},
		"When Rel returns an error": {
			currentDirPath:  "//host/some/remote/volume",
			projectDirPath:  "path/to/my-project",
			getwdPath:       "",  // Not called
			getwdError:      nil, // Not called
			expectedRelPath: "path/to/my-project",
			expectedError:   fmt.Errorf("Rel: can't make path/to/my-project relative to //host/some/remote/volume"),
		},
		"When current working directory outside the projectDirPath": {
			currentDirPath:  "path/to",
			projectDirPath:  "path/to/my-project",
			getwdPath:       "",  // Not called
			getwdError:      nil, // Not called
			expectedRelPath: "my-project",
			expectedError:   nil,
		},
		"When current working directory equals the projectDirPath": {
			currentDirPath:  "path/to/my-project",
			projectDirPath:  "path/to/my-project",
			getwdPath:       "",  // Not called
			getwdError:      nil, // Not called
			expectedRelPath: ".",
			expectedError:   nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mocks
			osMock := slackdeps.NewOsMock()
			osMock.On("Getwd").Return(tt.getwdPath, tt.getwdError)
			osMock.AddDefaultMocks()

			// Run the test
			actualRelPath, actualErr := getProjectDirRelPath(osMock, tt.currentDirPath, tt.projectDirPath)

			// Assertions
			require.Equal(t, tt.expectedRelPath, actualRelPath)
			require.Equal(t, tt.expectedError, actualErr)
		})
	}
}
