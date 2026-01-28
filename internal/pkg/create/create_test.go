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

package create

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackhttp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	assert.True(t, true, "should be true")
}

func TestGetProjectDirectoryName(t *testing.T) {
	var appName string
	var err error

	// Test without app name test removed because more than one possible default name
	// Test with app name
	appName, err = getAppDirName("my-app")
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, appName, "my-app", "should return 'my-app'")

	// Test with a dot in the app name
	appName, err = getAppDirName(".my-app")
	assert.NoError(t, err, "should not return an error")
	assert.Equal(t, appName, ".my-app", "should return '.my-app'")
}

func TestGetAvailableDirectory(t *testing.T) {
	var exists bool
	var err error

	// Current directory should exist
	exists, err = parentDirExists("pkg")
	assert.True(t, exists, "should exist")
	assert.Nil(t, err, "should not return an error")

	// Random directory should not exist
	exists, err = parentDirExists(`path/to/my-app`)
	assert.False(t, exists, "should not exist")
	assert.Error(t, err, "should return an error")

	// Dot notation for current directory (.) should exist
	exists, err = parentDirExists(`.`)
	assert.True(t, exists, "should exist")
	assert.Nil(t, err, "should not return an error")

	// Dot notation for parent directory (..) should exist
	exists, err = parentDirExists(`..`)
	assert.True(t, exists, "should exist")
	assert.Nil(t, err, "should not return an error")
}

func Test_generateGitZipFileURL(t *testing.T) {
	tests := map[string]struct {
		templateURL         string
		gitBranch           string
		expectedURL         string
		setupHTTPClientMock func(*slackhttp.HTTPClientMock)
	}{
		"Returns the zip URL using the main branch when no branch is provided": {
			templateURL: "https://github.com/slack-samples/deno-starter-template",
			gitBranch:   "",
			expectedURL: "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/main.zip",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusOK, "OK")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
		"Returns the zip URL using the master branch when no branch is provided and main branch doesn't exist": {
			templateURL: "https://github.com/slack-samples/deno-starter-template",
			gitBranch:   "",
			expectedURL: "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/master.zip",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusOK, "OK")
				httpClientMock.On("Head", "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/main.zip").Return(nil, fmt.Errorf("HttpClient error"))
				httpClientMock.On("Head", "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/master.zip").Return(res, nil)
			},
		},
		"Returns the zip URL using the specified branch when a branch is provided": {
			templateURL: "https://github.com/slack-samples/deno-starter-template",
			gitBranch:   "pre-release-0316",
			expectedURL: "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/pre-release-0316.zip",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusOK, "OK")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
		"Returns an empty string when the HTTP status code is not 200": {
			templateURL: "https://github.com/slack-samples/deno-starter-template",
			gitBranch:   "",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusNotFound, "Not Found")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
		"Returns an empty string when the HTTPClient has an error": {
			templateURL: "https://github.com/slack-samples/deno-starter-template",
			gitBranch:   "",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				httpClientMock.On("Head", mock.Anything).Return(nil, fmt.Errorf("HTTPClient error"))
			},
		},
		"Returns the zip URL with .git suffix removed": {
			templateURL: "https://github.com/slack-samples/deno-starter-template.git",
			gitBranch:   "",
			expectedURL: "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/main.zip",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusOK, "OK")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
		"Returns the zip URL with .git inside URL preserved": {
			templateURL: "https://github.com/slack-samples/deno.git-starter-template",
			gitBranch:   "",
			expectedURL: "https://github.com/slack-samples/deno.git-starter-template/archive/refs/heads/main.zip",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusOK, "OK")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mocks
			httpClientMock := &slackhttp.HTTPClientMock{}
			tc.setupHTTPClientMock(httpClientMock)

			// Execute
			url := generateGitZipFileURL(httpClientMock, tc.templateURL, tc.gitBranch)

			// Assertions
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}

func TestCreateGitArgs(t *testing.T) {
	var testGitArgs, expectedArgs []string

	templatePath := "git://github.com:slackapi/bolt-js-getting-started-app"

	// TemplateURLFlag
	testGitArgs = createGitArgs(templatePath, "./", "")
	expectedArgs = []string{"clone", "--depth=1", "git://github.com:slackapi/bolt-js-getting-started-app", "./"}
	assert.Equal(t, expectedArgs, testGitArgs)

	// GitBranchFlag
	testGitArgs = createGitArgs(templatePath, "./", "test-branch")
	expectedArgs = []string{"clone", "--depth=1", "git://github.com:slackapi/bolt-js-getting-started-app", "./", "--branch", "test-branch"}
	assert.Equal(t, expectedArgs, testGitArgs)

	// GitBranchFlag as empty string
	testGitArgs = createGitArgs(templatePath, "./", "    ")
	expectedArgs = []string{"clone", "--depth=1", "git://github.com:slackapi/bolt-js-getting-started-app", "./"}
	assert.Equal(t, expectedArgs, testGitArgs)
}

func Test_Create_installProjectDependencies(t *testing.T) {
	tests := map[string]struct {
		experiments            []string
		runtime                string
		manifestSource         config.ManifestSource
		existingFiles          map[string]string
		expectedOutputs        []string
		unexpectedOutputs      []string
		expectedVerboseOutputs []string
	}{
		"Should output added .slack, hooks.json, .gitignore, and caching": {
			expectedOutputs: []string{
				"Added project-name/.slack",
				"Added project-name/.slack/.gitignore",
				"Added project-name/.slack/hooks.json",
				"Cached dependencies with deno cache import_map.json",
			},
			expectedVerboseOutputs: []string{
				"Detected a project using Deno",
			},
		},
		"When hooks.json exists, should output found .slack and hooks.json": {
			existingFiles: map[string]string{
				".slack/hooks.json": "{}",
			},
			expectedOutputs: []string{
				"Found project-name/.slack",
				"Found project-name/.slack/hooks.json",
				"Cached dependencies with deno cache import_map.json",
			},
			unexpectedOutputs: []string{
				"Added project-name/.slack", // Already exists
				"Error adding the directory project-name/.slack",
			},
			expectedVerboseOutputs: []string{
				"Detected a project using Deno",
			},
		},
		"When slack.json exists, should output added .slack": {
			existingFiles: map[string]string{
				"slack.json": "{}",
			},
			expectedOutputs: []string{
				"Added project-name/.slack",
				"Found project-name/slack.json", // DEPRECATED(semver:major): Now use hooks.json
				"Cached dependencies with deno cache import_map.json",
			},
			expectedVerboseOutputs: []string{
				"Detected a project using Deno",
			},
		},
		"When no manifest source, default to project (local)": {
			expectedOutputs: []string{
				`Updated config.json manifest source to "project" (local)`,
			},
		},
		"When manifest source is provided, should set it": {
			manifestSource: config.ManifestSourceRemote,
			expectedOutputs: []string{
				`Updated config.json manifest source to "app settings" (remote)`,
			},
		},
		"When bolt-install experiment and Deno project, should set manifest source to project (local)": {
			experiments: []string{"bolt-install"},
			expectedOutputs: []string{
				`Updated config.json manifest source to "project" (local)`,
			},
		},
		"When bolt-install experiment and non-Deno project, should set manifest source to app settings (remote)": {
			experiments: []string{"bolt-install"},
			runtime:     "node",
			expectedOutputs: []string{
				`Updated config.json manifest source to "app settings" (remote)`,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Remove any enabled experiments during the test and restore afterward
			var _EnabledExperiments = experiment.EnabledExperiments
			experiment.EnabledExperiments = []experiment.Experiment{}
			defer func() {
				// Restore original EnabledExperiments
				experiment.EnabledExperiments = _EnabledExperiments
			}()

			// Setup parameters for test
			projectDirPath := "/path/to/project-name"

			// Create mocks
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Os.On("Getwd").Return(projectDirPath, nil)
			clientsMock.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return(`{}`, nil)
			clientsMock.AddDefaultMocks()

			// Set experiment flag
			clientsMock.Config.ExperimentsFlag = append(clientsMock.Config.ExperimentsFlag, tc.experiments...)
			clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)

			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Set runtime to be Deno (or node or whatever)
			clients.SDKConfig.Runtime = "deno"
			if tc.runtime != "" {
				clients.SDKConfig.Runtime = tc.runtime
			}

			// Create project directory
			if err := clients.Fs.MkdirAll(filepath.Dir(projectDirPath), 0755); err != nil {
				require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", projectDirPath))
			}

			// Create files
			for filePath, fileData := range tc.existingFiles {
				filePathAbs := filepath.Join(projectDirPath, filePath)
				// Create the directory
				if err := clients.Fs.MkdirAll(filepath.Dir(filePathAbs), 0755); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", filePath))
				}
				// Create the file
				if err := afero.WriteFile(clients.Fs, filePathAbs, []byte(fileData), 0644); err != nil {
					require.FailNow(t, fmt.Sprintf("Failed to create the file %s in the memory-based file system", filePath))
				}
			}

			// Run the test
			outputs := InstallProjectDependencies(ctx, clients, projectDirPath, tc.manifestSource)

			// Assertions
			for _, expectedOutput := range tc.expectedOutputs {
				require.Contains(t, outputs, expectedOutput)
			}
			for _, unexpectedOutput := range tc.unexpectedOutputs {
				require.NotContains(t, outputs, unexpectedOutput)
			}
			for _, expectedVerboseOutput := range tc.expectedVerboseOutputs {
				clientsMock.IO.AssertCalled(t, "PrintDebug", mock.Anything, expectedVerboseOutput, mock.MatchedBy(func(args ...any) bool { return true }))
			}
			assert.NotEmpty(t, clients.Config.ProjectID, "config.project_id")
			// output := clientsMock.GetCombinedOutput()
			// assert.Contains(t, output, tc.expectedOutputs)
		})
	}
}
