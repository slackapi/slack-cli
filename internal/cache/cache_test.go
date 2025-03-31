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

package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_createCacheDir(t *testing.T) {
	tests := map[string]struct {
		existingFiles          map[string]string
		expectedError          error
		expectedOutputs        []string
		unexpectedOutputs      []string
		expectedVerboseOutputs []string
	}{
		"returns with an existing error if caches are found": {
			existingFiles: map[string]string{
				".slack/cache/manifest.json": "{}",
			},
			expectedError: os.ErrExist,
		},
		"creates a new cache directory if none exists": {
			expectedError: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			projectDirPath := "/path/to/project-name"
			err := fsMock.MkdirAll(filepath.Dir(projectDirPath), 0o755)
			require.NoError(t, err)
			for filePath, fileData := range tt.existingFiles {
				filePathAbs := filepath.Join(projectDirPath, filePath)
				err := fsMock.MkdirAll(filepath.Dir(filePathAbs), 0o755)
				require.NoError(t, err)
				err = afero.WriteFile(fsMock, filePathAbs, []byte(fileData), 0o644)
				require.NoError(t, err)
			}
			osMock := slackdeps.NewOsMock()
			cache := NewCache(fsMock, osMock, projectDirPath)
			cacheErr := cache.createCacheDir()
			assert.Equal(t, tt.expectedError, cacheErr)
			dir, err := fsMock.Stat(filepath.Join(projectDirPath, ".slack", "cache"))
			require.NoError(t, err)
			assert.True(t, dir.IsDir())
		})
	}
}
