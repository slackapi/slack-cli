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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ProjectConfig_NewProjectConfig(t *testing.T) {
	t.Run("NewProjectConfig(fs, os) when arguments nil", func(t *testing.T) {
		var fs afero.Fs
		var os types.Os
		fs = nil
		os = nil

		var config = NewProjectConfig(fs, os)
		require.Equal(t, nil, config.fs)
		require.Equal(t, nil, config.os)
	})

	// Test: when arguments exist
	t.Run("NewProjectConfig(fs, os) when arguments exist", func(t *testing.T) {
		var fs = slackdeps.NewFsMock()
		var os = slackdeps.NewOsMock()
		os.AddDefaultMocks()

		var projectConfig = NewProjectConfig(fs, os)
		require.Equal(t, fs, projectConfig.fs)
		require.Equal(t, os, projectConfig.os)
	})
}

func Test_ProjectConfig_InitProjectID(t *testing.T) {
	t.Run("When not a project directory, should return an error", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()
		// Do not add project mocks

		projectConfig := NewProjectConfig(fs, os)
		projectID, err := projectConfig.InitProjectID(ctx, false)

		require.Error(t, err)
		require.Empty(t, projectID)
	})

	t.Run("When project_id is empty, should init project_id", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		projectConfig := NewProjectConfig(fs, os)
		projectID, err := projectConfig.InitProjectID(ctx, false)

		require.NoError(t, err)
		require.NotEmpty(t, projectID)
	})

	t.Run("When project_id exists, should not overwrite project_id (overwrite: false)", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		projectConfig := NewProjectConfig(fs, os)

		// Set a project_id that should not be overwritten
		expectedProjectID := uuid.New().String()
		_, err := WriteProjectConfigFile(ctx, fs, os, ProjectConfig{ProjectID: expectedProjectID})
		require.NoError(t, err)

		projectID, err := projectConfig.InitProjectID(ctx, false)

		require.NoError(t, err)
		require.Equal(t, expectedProjectID, projectID)
	})

	t.Run("When project_id exists, should overwrite project_id (overwrite: true)", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		projectConfig := NewProjectConfig(fs, os)

		// Set a project_id that should be overwritten
		expectedProjectID := uuid.New().String()
		_, err := WriteProjectConfigFile(ctx, fs, os, ProjectConfig{ProjectID: expectedProjectID})
		require.NoError(t, err)

		projectID, err := projectConfig.InitProjectID(ctx, true)

		require.NoError(t, err)
		require.NotEqual(t, expectedProjectID, projectID)
		require.NotEmpty(t, projectID)
	})
}

func Test_ProjectConfig_GetProjectID(t *testing.T) {
	t.Run("When not a project directory, should return an error", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()
		// Do not add project mocks

		projectConfig := NewProjectConfig(fs, os)
		projectID, err := projectConfig.GetProjectID(ctx)

		require.Error(t, err)
		require.Empty(t, projectID)
	})

	t.Run("When a project directory, should return trimmed project_id", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		projectConfig := NewProjectConfig(fs, os)

		// Set a project_id that has whitespace padding
		const paddedProjectID = "   abc-123   "
		_, err := WriteProjectConfigFile(ctx, fs, os, ProjectConfig{ProjectID: paddedProjectID})
		require.NoError(t, err)

		projectID, err := projectConfig.GetProjectID(ctx)

		require.NoError(t, err)
		require.Equal(t, "abc-123", projectID)
	})
}

func Test_ProjectConfig_SetProjectID(t *testing.T) {
	t.Run("When not a project directory, should return an error", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()
		// Do not create the project fixtures

		projectConfig := NewProjectConfig(fs, os)
		projectID, err := projectConfig.SetProjectID(ctx, uuid.New().String())

		require.Error(t, err)
		require.Equal(t, slackerror.ToSlackError(err).Code, slackerror.ErrInvalidAppDirectory)
		require.Empty(t, projectID)
	})

	t.Run("When a project directory, should update the project_id", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		projectConfig := NewProjectConfig(fs, os)

		// Set a project_id that will be replaced later
		_, err := WriteProjectConfigFile(ctx, fs, os, ProjectConfig{ProjectID: uuid.New().String()})
		require.NoError(t, err)

		var expectedProjectID = uuid.New().String()
		projectID, err := projectConfig.SetProjectID(ctx, expectedProjectID)

		require.NoError(t, err)
		require.Equal(t, expectedProjectID, projectID)
	})
}

func Test_ProjectConfig_ManifestSource(t *testing.T) {
	tests := map[string]struct {
		mockManifestSource            ManifestSource
		expectedManifestSourceDefault ManifestSource
		expectedManifestSource        ManifestSource
		expectedError                 error
	}{
		"saves manifest.source remote to project configs": {
			mockManifestSource:            ManifestSourceRemote,
			expectedManifestSourceDefault: ManifestSourceLocal,
			expectedManifestSource:        ManifestSourceRemote,
		},
		"saves manifest.source local to project configs": {
			mockManifestSource:            ManifestSourceLocal,
			expectedManifestSourceDefault: ManifestSourceLocal,
			expectedManifestSource:        ManifestSourceLocal,
		},
		"errors if an unknown manifest.source is provided": {
			mockManifestSource:            ManifestSource("upstream"),
			expectedManifestSourceDefault: ManifestSourceLocal,
			expectedError:                 slackerror.New(slackerror.ErrProjectConfigManifestSource),
		},
		"defaults to local manifest without manifest.source": {
			mockManifestSource:            ManifestSource(""),
			expectedManifestSourceDefault: ManifestSourceLocal,
			expectedManifestSource:        ManifestSourceLocal,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			addProjectMocks(t, fs)
			config := NewProjectConfig(fs, os)
			initial, err := config.GetManifestSource(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedManifestSourceDefault, initial)
			err = SetManifestSource(ctx, fs, os, tt.mockManifestSource)
			require.NoError(t, err)
			actual, err := config.GetManifestSource(ctx)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedManifestSource, actual)
		})
	}
}

func Test_ProjectConfig_ReadProjectConfigFile(t *testing.T) {
	t.Run("When not a project directory, should return an error", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()
		// Do not add the project mocks

		projectConfigData, err := ReadProjectConfigFile(ctx, fs, os)
		require.Error(t, err)
		require.Empty(t, projectConfigData)
	})

	t.Run("When project directory doesn't have a .slack/config.json, should return a default config.json", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		// Assert project-level .slack/config.json does not exist
		_, err := fs.Stat(GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory))
		require.True(t, os.IsNotExist(err))

		projectConfigFile, err := ReadProjectConfigFile(ctx, fs, os)
		require.NoError(t, err)
		require.Empty(t, projectConfigFile) // Currently, the default is an empty config.json ("{}") but this may change in the future
	})

	t.Run("When a project directory has a .slack/config.json, should return config.json", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		expectedProjectConfig := ProjectConfig{
			ProjectID: uuid.New().String(),
			Surveys:   map[string]SurveyConfig{},
		}

		_, err := WriteProjectConfigFile(ctx, fs, os, expectedProjectConfig)
		require.NoError(t, err)

		projectConfigFile, err := ReadProjectConfigFile(ctx, fs, os)
		require.NoError(t, err)
		require.Equal(t, expectedProjectConfig, projectConfigFile)
	})

	t.Run("errors on invalid formatting of project config file", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		os.AddDefaultMocks()
		addProjectMocks(t, fs)

		projectDirPath, err := GetProjectDirPath(fs, os)
		require.NoError(t, err)
		projectConfigFilePath := GetProjectConfigJSONFilePath(projectDirPath)

		expectedConfigFileData := "{\"hello\""
		err = afero.WriteFile(fs, projectConfigFilePath, []byte(expectedConfigFileData), 0600)
		require.NoError(t, err)

		_, err = ReadProjectConfigFile(ctx, fs, os)
		require.Error(t, err)
		assert.Equal(t, slackerror.ToSlackError(err).Code, slackerror.ErrUnableToParseJSON)
		assert.Equal(t, slackerror.ToSlackError(err).Message, "Failed to parse contents of project-level config file")
	})
}

func Test_ProjectConfig_WriteProjectConfigFile(t *testing.T) {
	t.Run("When not a project directory, should return an error", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()
		// Do not add project mocks

		defaultProjectConfig := ProjectConfig{
			ProjectID: uuid.New().String(),
		}

		projectConfigData, err := WriteProjectConfigFile(ctx, fs, os, defaultProjectConfig)
		require.Error(t, err)
		require.Empty(t, projectConfigData)
	})

	t.Run("When a project directory, should write the .slack/config.json file", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		expectedProjectConfig := ProjectConfig{
			ProjectID: uuid.New().String(),
			Surveys:   map[string]SurveyConfig{},
		}

		// Assert writing the config file
		_, err := WriteProjectConfigFile(ctx, fs, os, expectedProjectConfig)
		require.NoError(t, err)

		// Assert reading the written file contents
		actualProjectConfig, err := ReadProjectConfigFile(ctx, fs, os)
		require.NoError(t, err)

		// Assert the written file has the same content as the original
		require.Equal(t, expectedProjectConfig, actualProjectConfig)
	})
}

func Test_ProjectConfig_ProjectConfigJSONFileExists(t *testing.T) {
	t.Run("When .slack/config.json exists, should return true", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		// Create .slack/config.json
		err := afero.WriteFile(fs, GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory), []byte("{}\n"), 0600)
		require.NoError(t, err)

		exists := ProjectConfigJSONFileExists(fs, os, slackdeps.MockWorkingDirectory)
		require.True(t, exists)
	})

	t.Run("When .slack/config.json does not exist, should return false", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		// Remove .slack/config.json and ignore errors (errors when file does not exist)
		_ = fs.Remove(GetProjectConfigJSONFilePath(slackdeps.MockWorkingDirectory))

		exists := ProjectConfigJSONFileExists(fs, os, slackdeps.MockWorkingDirectory)
		require.False(t, exists)
	})
}

func Test_ProjectConfig_GetProjectDirPath(t *testing.T) {
	t.Run("When project directory is missing .slack/hooks.json, should return an error", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		// Delete .slack/hooks.json
		err := fs.Remove(GetProjectHooksJSONFilePath(slackdeps.MockWorkingDirectory))
		require.NoError(t, err)

		projectDirPath, err := GetProjectDirPath(fs, os)
		require.Error(t, err)
		require.Empty(t, projectDirPath)
	})

	t.Run("When project directory has .slack/hooks.json, should return directory path", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Use project mocks to create project in filesystem
		addProjectMocks(t, fs)

		projectDirPath, err := GetProjectDirPath(fs, os)
		require.NoError(t, err)
		require.Equal(t, slackdeps.MockWorkingDirectory, projectDirPath) // MockWorkingDirectory is the test's project directory
	})
}

func Test_ProjectConfig_Cache(t *testing.T) {
	tests := map[string]struct {
		mockAppID    string
		mockHash     cache.Hash
		expectedHash cache.Hash
	}{
		"creates an empty cache": {
			mockAppID: "A0123456789",
		},
		"persists cache changes": {
			mockAppID:    "A0123456789",
			mockHash:     "xoxo",
			expectedHash: "xoxo",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			addProjectMocks(t, fs)
			projectConfig := NewProjectConfig(fs, os)
			cache := projectConfig.Cache()
			if !tt.mockHash.Equals("") {
				err := cache.SetManifestHash(ctx, tt.mockAppID, tt.mockHash)
				require.NoError(t, err)
			}
			hash, err := cache.GetManifestHash(ctx, tt.mockAppID)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHash, hash)
		})
	}
}

func Test_Config_GetProjectConfigDirPath(t *testing.T) {
	tests := map[string]struct {
		projectDirPath        string
		expectedConfigDirPath string
	}{
		"Absolute project path": {
			projectDirPath:        "/path/to/project-name",
			expectedConfigDirPath: "/path/to/project-name/.slack",
		},
		"Empty project path": {
			projectDirPath:        "",
			expectedConfigDirPath: ".slack",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			configDirPath := GetProjectConfigDirPath(tt.projectDirPath)
			require.Equal(t, configDirPath, tt.expectedConfigDirPath)
		})
	}
}

func Test_Config_CreateProjectConfigDir(t *testing.T) {
	tests := map[string]struct {
		projectDirPath        string
		existingDir           bool
		expectedConfigDirPath string
		expectedError         error
	}{
		"Valid project path": {
			projectDirPath:        "/path/to/project-name",
			expectedConfigDirPath: "/path/to/project-name/.slack",
			expectedError:         nil,
		},
		"Existing dotslack directory": {
			projectDirPath:        "/path/to/project-name",
			existingDir:           true,
			expectedConfigDirPath: "/path/to/project-name/.slack",
			expectedError:         os.ErrExist,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := afero.NewMemMapFs()
			if tt.existingDir {
				err := fs.MkdirAll(tt.projectDirPath+"/.slack", 0755)
				require.NoError(t, err)
			}
			projectConfigDirPath, err := CreateProjectConfigDir(ctx, fs, tt.projectDirPath)
			require.Equal(t, err, tt.expectedError)
			require.Equal(t, projectConfigDirPath, tt.expectedConfigDirPath)
		})
	}
}

func Test_Config_GetProjectConfigJSONFilePath(t *testing.T) {
	tests := map[string]struct {
		projectDirPath             string
		expectedConfigJSONFilePath string
	}{
		"Valid project path": {
			projectDirPath:             "/path/to/project-name",
			expectedConfigJSONFilePath: "/path/to/project-name/.slack/config.json",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			configJSONFilePath := GetProjectConfigJSONFilePath(tt.projectDirPath)
			require.Equal(t, configJSONFilePath, tt.expectedConfigJSONFilePath)
		})
	}
}

func Test_Config_CreateProjectConfigJSONFile(t *testing.T) {
	tests := map[string]struct {
		projectDirPath             string
		existingConfigJSONData     string
		expectedError              error
		expectedConfigJSONFilePath string
		expectedConfigJSONFileData string
	}{
		"Create a new file": {
			projectDirPath:             "/path/to/project-name",
			existingConfigJSONData:     "",
			expectedError:              nil,
			expectedConfigJSONFilePath: "/path/to/project-name/.slack/config.json",
			expectedConfigJSONFileData: `{}`,
		},
		"Existing slack/hooks.json": {
			projectDirPath:             "/path/to/project-name",
			existingConfigJSONData:     `{ "project_id": "abc123" }`,
			expectedError:              os.ErrExist,
			expectedConfigJSONFilePath: "/path/to/project-name/.slack/config.json",
			expectedConfigJSONFileData: `{ "project_id": "abc123" }`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := afero.NewMemMapFs()

			// Create the project directory and .slack directory
			if _, err := CreateProjectConfigDir(ctx, fs, tt.projectDirPath); err != nil {
				require.Fail(t, "Failed to create the project's .slack/ directory: %s", err)
			}

			// Create existing .slack/config.json (optional)
			if tt.existingConfigJSONData != "" {
				configJSONFilePath := GetProjectConfigJSONFilePath(tt.projectDirPath)
				if err := afero.WriteFile(fs, configJSONFilePath, []byte(tt.existingConfigJSONData), 0o0644); err != nil {
					require.Fail(t, "Failed to setup the test by creating an existing .slack/config.json: %s", err)
				}
			}

			// Run the test
			configJSONFilePath, err := CreateProjectConfigJSONFile(fs, tt.projectDirPath)
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, configJSONFilePath, tt.expectedConfigJSONFilePath)

			// Assert .slack/config.json data is correct
			configJSONData, err := afero.ReadFile(fs, configJSONFilePath)
			require.Equal(t, err, nil)
			require.Equal(t, string(configJSONData), tt.expectedConfigJSONFileData)
		})
	}
}

func Test_Config_GetProjectHooksJSONFilePath(t *testing.T) {
	tests := map[string]struct {
		projectDirPath            string
		expectedHooksJSONFilePath string
	}{
		"Valid project path": {
			projectDirPath:            "/path/to/project-name",
			expectedHooksJSONFilePath: "/path/to/project-name/.slack/hooks.json",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hooksJSONFilePath := GetProjectHooksJSONFilePath(tt.projectDirPath)
			require.Equal(t, hooksJSONFilePath, tt.expectedHooksJSONFilePath)
		})
	}
}

func Test_Config_CreateProjectHooksJSONFile(t *testing.T) {
	tests := map[string]struct {
		projectDirPath            string
		hooksJSONData             string
		existingHooksJSONData     string
		expectedError             error
		expectedHooksJSONFilePath string
		expectedHooksJSONFileData string
	}{
		"Create a new file": {
			projectDirPath:            "/path/to/project-name",
			hooksJSONData:             `{ "valid": "json" }`,
			existingHooksJSONData:     "",
			expectedError:             nil,
			expectedHooksJSONFilePath: "/path/to/project-name/.slack/hooks.json",
			expectedHooksJSONFileData: `{ "valid": "json" }`,
		},
		"Existing slack/hooks.json": {
			projectDirPath:            "/path/to/project-name",
			hooksJSONData:             `{ "valid": "json" }`,
			existingHooksJSONData:     `{ "existing": "file" }`,
			expectedError:             os.ErrExist,
			expectedHooksJSONFilePath: "/path/to/project-name/.slack/hooks.json",
			expectedHooksJSONFileData: `{ "existing": "file" }`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			// Create existing .slack/hooks.json (optional)
			if tt.existingHooksJSONData != "" {
				hooksJSONFilePath := GetProjectHooksJSONFilePath(tt.projectDirPath)
				if err := afero.WriteFile(fs, hooksJSONFilePath, []byte(tt.existingHooksJSONData), 0644); err != nil {
					require.Fail(t, "Failed to setup the test by creating an existing .slack/hooks.json: "+err.Error())
				}
			}

			// Run the test
			hooksJSONFilePath, err := CreateProjectHooksJSONFile(fs, tt.projectDirPath, []byte(tt.hooksJSONData))
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, hooksJSONFilePath, tt.expectedHooksJSONFilePath)

			// Assert .slack/hooks.json data is correct
			hooksJSONData, err := afero.ReadFile(fs, hooksJSONFilePath)
			require.Equal(t, err, nil)
			require.Equal(t, string(hooksJSONData), tt.expectedHooksJSONFileData)
		})
	}
}

func Test_Config_GetProjectConfigDirDotGitIgnoreFilePath(t *testing.T) {
	tests := map[string]struct {
		projectDirPath               string
		expectedDotGitignoreFilePath string
	}{
		"Valid file path": {
			projectDirPath:               "/path/to/project-name",
			expectedDotGitignoreFilePath: "/path/to/project-name/.slack/.gitignore",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dotGitignoreFilePath := GetProjectConfigDirDotGitIgnoreFilePath(tt.projectDirPath)
			require.Equal(t, dotGitignoreFilePath, tt.expectedDotGitignoreFilePath)
		})
	}
}

func Test_Config_CreateProjectConfigDirDotGitIgnoreFile(t *testing.T) {
	tests := map[string]struct {
		projectDirPath               string
		existingDotGitIgnoreFileData string
		expectedError                error
		expectedDotGitIgnoreFilePath string
		expectedDotGitIgnoreFileData string
	}{
		"Create a new file": {
			projectDirPath:               "/path/to/project-name",
			existingDotGitIgnoreFileData: "",
			expectedError:                nil,
			expectedDotGitIgnoreFilePath: "/path/to/project-name/.slack/.gitignore",
			expectedDotGitIgnoreFileData: "apps.dev.json\ncache/\n",
		},
		"Existing slack/hooks.json": {
			projectDirPath:               "/path/to/project-name",
			existingDotGitIgnoreFileData: `# existing .gitignore file data`,
			expectedError:                os.ErrExist,
			expectedDotGitIgnoreFilePath: "/path/to/project-name/.slack/.gitignore",
			expectedDotGitIgnoreFileData: `# existing .gitignore file data`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			// Create existing .slack/.gitignore (optional)
			if tt.existingDotGitIgnoreFileData != "" {
				dotGitIgnoreFilePath := GetProjectConfigDirDotGitIgnoreFilePath(tt.projectDirPath)
				if err := afero.WriteFile(fs, dotGitIgnoreFilePath, []byte(tt.existingDotGitIgnoreFileData), 0644); err != nil {
					require.Fail(t, "Failed to setup the test by creating an existing .slack/.gitignore: "+err.Error())
				}
			}

			// Run the test
			dotGitIgnoreFilePath, err := CreateProjectConfigDirDotGitIgnoreFile(fs, tt.projectDirPath)
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, dotGitIgnoreFilePath, tt.expectedDotGitIgnoreFilePath)

			// Assert .slack/.gitignore data is correct
			fileData, err := afero.ReadFile(fs, dotGitIgnoreFilePath)
			require.Equal(t, err, nil)
			require.Equal(t, string(fileData), tt.expectedDotGitIgnoreFileData)
		})
	}
}

// addProjectMocks will create the project's required directory and files
func addProjectMocks(t require.TestingT, fs afero.Fs) {
	var err error

	// Fixture: project/ directory
	err = fs.Mkdir(slackdeps.MockWorkingDirectory, 0755)
	require.NoError(t, err)

	// Fixture: project/.slack/ directory
	err = fs.Mkdir(filepath.Join(slackdeps.MockWorkingDirectory, ProjectConfigDirName), 0755)
	require.NoError(t, err)

	// Fixture: project/.slack/hooks.json file
	err = afero.WriteFile(fs, GetProjectHooksJSONFilePath(slackdeps.MockWorkingDirectory), []byte("{}\n"), 0644)
	require.NoError(t, err)
}
