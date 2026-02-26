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

package app

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/test/testdata"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_App_UpdateDefaultProjectFiles(t *testing.T) {
	tests := map[string]struct {
		appDirName        string
		existingFiles     map[string]string
		expectedFiles     map[string]string
		expectedErrorType error
	}{
		"manifest.json file exists": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"manifest.json": string(testdata.ManifestJSON),
			},
			expectedFiles: map[string]string{
				"manifest.json": string(testdata.ManifestJSONAppName),
			},
			expectedErrorType: nil,
		},
		"manifest.js file exists": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"manifest.js": string(testdata.ManifestJS),
			},
			expectedFiles: map[string]string{
				"manifest.js": string(testdata.ManifestJSAppName),
			},
			expectedErrorType: nil,
		},
		"manifest.ts file exists": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"manifest.ts": string(testdata.ManifestTS),
			},
			expectedFiles: map[string]string{
				"manifest.ts": string(testdata.ManifestTSAppName),
			},
			expectedErrorType: nil,
		},
		"Multiple manifest files exist": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"manifest.json": string(testdata.ManifestJSON),
				"manifest.ts":   string(testdata.ManifestTS),
			},
			expectedFiles: map[string]string{
				"manifest.json": string(testdata.ManifestJSONAppName),
				"manifest.ts":   string(testdata.ManifestTSAppName),
			},
			expectedErrorType: nil,
		},
		"package.json file exists": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"package.json": string(testdata.PackageJSON),
			},
			expectedFiles: map[string]string{
				"package.json": string(testdata.PackageJSONAppName),
			},
			expectedErrorType: nil,
		},
		"pyproject.toml file exists": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"pyproject.toml": string(testdata.PyprojectTOML),
			},
			expectedFiles: map[string]string{
				"pyproject.toml": string(testdata.PyprojectTOMLAppName),
			},
			expectedErrorType: nil,
		},
		"Multiple project files exist": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"manifest.json":  string(testdata.ManifestJSON),
				"package.json":   string(testdata.PackageJSON),
				"pyproject.toml": string(testdata.PyprojectTOML),
			},
			expectedFiles: map[string]string{
				"manifest.json":  string(testdata.ManifestJSONAppName),
				"package.json":   string(testdata.PackageJSONAppName),
				"pyproject.toml": string(testdata.PyprojectTOMLAppName),
			},
			expectedErrorType: nil,
		},
		"No manifest files exist": {
			appDirName:        "vibrant-butterfly-1234",
			existingFiles:     map[string]string{},
			expectedFiles:     map[string]string{},
			expectedErrorType: nil,
		},
		"WriteFile error": {
			appDirName: "vibrant-butterfly-1234",
			existingFiles: map[string]string{
				"manifest.json": string(testdata.ManifestJSON),
			},
			expectedFiles: map[string]string{
				"manifest.json": string(testdata.ManifestJSONAppName),
			},
			expectedErrorType: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup parameters for test
			fs := slackdeps.NewFsMock()
			projectDirPath := "/path/to/project-name"

			// Create files
			for filePath, fileData := range tc.existingFiles {
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

			// Run the tests
			err := UpdateDefaultProjectFiles(fs, projectDirPath, tc.appDirName)

			// Assertions
			require.IsType(t, err, tc.expectedErrorType)

			for filePath, fileData := range tc.expectedFiles {
				filePathAbs := filepath.Join(projectDirPath, filePath)
				d, err := afero.ReadFile(fs, filePathAbs)
				require.NoError(t, err)
				require.Equal(t, fileData, string(d))
			}
		})
	}
}

func Test_RegexReplaceAppNameInManifest(t *testing.T) {
	tests := map[string]struct {
		src         []byte
		appName     string
		expectedSrc []byte
	}{
		"manifest.json is validate": {
			src:         testdata.ManifestJSON,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestJSONAppName,
		},
		"manifest.js is validate": {
			src:         testdata.ManifestJS,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestJSAppName,
		},
		"manifest.ts is validate": {
			src:         testdata.ManifestTS,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestTSAppName,
		},
		"manifest.ts with sdk is validate": {
			src:         testdata.ManifestSDKTS,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestSDKTSAppName,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSrc := regexReplaceAppNameInManifest(tc.src, tc.appName)
			require.Equal(t, tc.expectedSrc, actualSrc)
		})
	}
}

func Test_RegexReplaceAppNameInPackageJSON(t *testing.T) {
	tests := map[string]struct {
		src         []byte
		appName     string
		expectedSrc []byte
	}{
		"package.json name is replaced": {
			src:         testdata.PackageJSON,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.PackageJSONAppName,
		},
		"only top-level name is replaced not nested config name": {
			src: []byte(`{
  "name": "bolt-app-template",
  "version": "1.0.0",
  "description": "A Slack app built with Bolt",
  "main": "app.js",
  "scripts": {
    "start": "node app.js"
  },
  "dependencies": {
    "@slack/bolt": "^4.0.0"
  },
  "config": {
    "name": "local-server-name",
    "host": "localhost",
    "port": "8080"
  }
}
`),
			appName: "vibrant-butterfly-1234",
			expectedSrc: []byte(`{
  "name": "vibrant-butterfly-1234",
  "version": "1.0.0",
  "description": "A Slack app built with Bolt",
  "main": "app.js",
  "scripts": {
    "start": "node app.js"
  },
  "dependencies": {
    "@slack/bolt": "^4.0.0"
  },
  "config": {
    "name": "local-server-name",
    "host": "localhost",
    "port": "8080"
  }
}
`),
		},
		"no name field leaves input unchanged": {
			src: []byte(`{
  "version": "1.0.0"
}
`),
			appName: "my-app",
			expectedSrc: []byte(`{
  "version": "1.0.0"
}
`),
		},
		"empty name value is replaced": {
			src:         []byte("{\n  \"name\": \"\"\n}\n"),
			appName:     "my-app",
			expectedSrc: []byte("{\n  \"name\": \"my-app\"\n}\n"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSrc := regexReplaceAppNameInPackageJSON(tc.src, tc.appName)
			require.Equal(t, tc.expectedSrc, actualSrc)
		})
	}
}

func Test_RegexReplaceAppNameInPyprojectToml(t *testing.T) {
	tests := map[string]struct {
		src         []byte
		appName     string
		expectedSrc []byte
	}{
		"pyproject.toml name is replaced": {
			src:         testdata.PyprojectTOML,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.PyprojectTOMLAppName,
		},
		"only project section name is replaced not project.scripts name": {
			src: []byte(`[project]
name = "bolt-python-ai-agent-template"
version = "0.1.0"
requires-python = ">=3.9"
dependencies = [
    "slack-sdk==3.40.0",
    "slack-bolt==1.27.0",
    "slack-cli-hooks<1.0.0",
]

[tool.ruff]
[tool.ruff.lint]
[tool.ruff.format]

[tool.pytest.ini_options]
testpaths = ["tests"]

[project.scripts]
name = "my_package.name:main_function"
`),
			appName: "vibrant-butterfly-1234",
			expectedSrc: []byte(`[project]
name = "vibrant-butterfly-1234"
version = "0.1.0"
requires-python = ">=3.9"
dependencies = [
    "slack-sdk==3.40.0",
    "slack-bolt==1.27.0",
    "slack-cli-hooks<1.0.0",
]

[tool.ruff]
[tool.ruff.lint]
[tool.ruff.format]

[tool.pytest.ini_options]
testpaths = ["tests"]

[project.scripts]
name = "my_package.name:main_function"
`),
		},
		"no project section leaves input unchanged": {
			src: []byte(`[tool.ruff]
name = "should-not-change"
`),
			appName: "my-app",
			expectedSrc: []byte(`[tool.ruff]
name = "should-not-change"
`),
		},
		"empty name value is replaced": {
			src:         []byte(`[project]` + "\n" + `name = ""` + "\n"),
			appName:     "my-app",
			expectedSrc: []byte(`[project]` + "\n" + `name = "my-app"` + "\n"),
		},
		"extra whitespace around equals sign": {
			src:         []byte(`[project]` + "\n" + `name  =  "old-name"` + "\n"),
			appName:     "new-name",
			expectedSrc: []byte(`[project]` + "\n" + `name  =  "new-name"` + "\n"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSrc := regexReplaceAppNameInPyprojectToml(tc.src, tc.appName)
			require.Equal(t, tc.expectedSrc, actualSrc)
		})
	}
}
