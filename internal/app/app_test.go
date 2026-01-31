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
	tests := []struct {
		name        string
		src         []byte
		appName     string
		expectedSrc []byte
	}{
		{
			name:        "manifest.json is validate",
			src:         testdata.ManifestJSON,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestJSONAppName,
		},
		{
			name:        "manifest.js is validate",
			src:         testdata.ManifestJS,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestJSAppName,
		},
		{
			name:        "manifest.ts is validate",
			src:         testdata.ManifestTS,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestTSAppName,
		},
		{
			name:        "manifest.ts with sdk is validate",
			src:         testdata.ManifestSDKTS,
			appName:     "vibrant-butterfly-1234",
			expectedSrc: testdata.ManifestSDKTSAppName,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualSrc := regexReplaceAppNameInManifest(tc.src, tc.appName)
			require.Equal(t, tc.expectedSrc, actualSrc)
		})
	}
}
