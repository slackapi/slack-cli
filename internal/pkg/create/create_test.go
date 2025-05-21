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
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
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
	url := generateGitZipFileURL("https://github.com/slack-samples/deno-starter-template", "pre-release-0316")
	assert.Equal(t, "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/pre-release-0316.zip", url, "should return zip download link with branch")

	url = generateGitZipFileURL("https://github.com/slack-samples/deno-starter-template", "")
	assert.Equal(t, "https://github.com/slack-samples/deno-starter-template/archive/refs/heads/main.zip", url, "should return zip download link with main")

	// TODO - We should mock the `deputil.URLChecker` HTTP request so that the unit test is not dependent on the network activity and repo configuration
	url = generateGitZipFileURL("https://github.com/google/uuid", "")
	assert.Equal(t, "https://github.com/google/uuid/archive/refs/heads/master.zip", url, "should return zip download link with 'master' when 'main' branch doesn't exist")

	url = generateGitZipFileURL("fake_url", "")
	assert.Equal(t, "", url, "should return empty string when url is invalid")
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
		"When no bolt experiment and hooks.json exists, should output found .slack and caching steps": {
			existingFiles: map[string]string{
				".slack/hooks.json": "{}", // Included with the template
			},
			expectedOutputs: []string{
				"Found project-name/.slack",
				"Cached dependencies with deno cache import_map.json",
			},
			unexpectedOutputs: []string{
				"Found project-name/.slack/hooks.json", // Behind bolt experiment
				"project-name/slack.json",              // DEPRECATED(semver:major): Now use hooks.json
			},
			expectedVerboseOutputs: []string{
				"Detected a project using Deno",
			},
		},
		"When no bolt experiment and slack.json exists, should output adding .slack and caching steps": {
			existingFiles: map[string]string{
				"slack.json": "{}", // DEPRECATED(semver:major): Included with the template (deprecated path)
			},
			expectedOutputs: []string{
				"Added project-name/.slack",
				"Cached dependencies with deno cache import_map.json",
			},
			unexpectedOutputs: []string{
				"project-name/.slack/hooks.json", // Behind bolt experiment, file doesn't exist
				"project-name/slack.json",        // Behind bolt experiment, file exists
			},
			expectedVerboseOutputs: []string{
				"Detected a project using Deno",
			},
		},
		"When bolt experiment, should output added .slack, hooks.json, .gitignore, and caching": {
			experiments: []string{"bolt"},
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
		"When bolt experiment and hooks.json exists, should output found .slack and hooks.json": {
			experiments: []string{"bolt"},
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
		"When bolt experiment and slack.json exists, should output added .slack": {
			experiments: []string{"bolt"},
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
			experiments: []string{"bolt"},
			expectedOutputs: []string{
				`Updated config.json manifest source to "project" (local)`,
			},
		},
		"When manifest source is provided, should set it": {
			experiments:    []string{"bolt"},
			manifestSource: config.ManifestSourceRemote,
			expectedOutputs: []string{
				`Updated config.json manifest source to "app settings" (remote)`,
			},
		},
		"When bolt + bolt-install experiment and Deno project, should set manifest source to project (local)": {
			experiments: []string{"bolt", "bolt-install"},
			expectedOutputs: []string{
				`Updated config.json manifest source to "project" (local)`,
			},
		},
		"When bolt + bolt-install experiment and non-Deno project, should set manifest source to app settings (remote)": {
			experiments: []string{"bolt", "bolt-install"},
			runtime:     "node",
			expectedOutputs: []string{
				`Updated config.json manifest source to "app settings" (remote)`,
			},
		},
	}
	for name, tt := range tests {
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
			clientsMock.Config.ExperimentsFlag = append(clientsMock.Config.ExperimentsFlag, tt.experiments...)
			clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)

			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Set runtime to be Deno (or node or whatever)
			clients.SDKConfig.Runtime = "deno"
			if tt.runtime != "" {
				clients.SDKConfig.Runtime = tt.runtime
			}

			// Create project directory
			if err := clients.Fs.MkdirAll(filepath.Dir(projectDirPath), 0755); err != nil {
				require.FailNow(t, fmt.Sprintf("Failed to create the directory %s in the memory-based file system", projectDirPath))
			}

			// Create files
			for filePath, fileData := range tt.existingFiles {
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
			outputs := InstallProjectDependencies(ctx, clients, projectDirPath, tt.manifestSource)

			// Assertions
			for _, expectedOutput := range tt.expectedOutputs {
				require.Contains(t, outputs, expectedOutput)
			}
			for _, unexpectedOutput := range tt.unexpectedOutputs {
				require.NotContains(t, outputs, unexpectedOutput)
			}
			for _, expectedVerboseOutput := range tt.expectedVerboseOutputs {
				clientsMock.IO.AssertCalled(t, "PrintDebug", mock.Anything, expectedVerboseOutput, mock.MatchedBy(func(args ...any) bool { return true }))
			}
			assert.NotEmpty(t, clients.Config.ProjectID, "config.project_id")
			// output := clientsMock.GetCombinedOutput()
			// assert.Contains(t, output, tt.expectedOutputs)
		})
	}
}
