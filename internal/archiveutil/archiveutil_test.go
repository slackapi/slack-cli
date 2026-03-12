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

package archiveutil

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Refactor to use afero.Fs once ExtractAndWriteFile accepts it. Currently uses t.TempDir() which is safe.
func Test_ExtractAndWriteFile(t *testing.T) {
	tests := map[string]struct {
		fileName string
		content  string
		isDir    bool
	}{
		"extracts a regular file": {
			fileName: "hello.txt",
			content:  "hello world",
			isDir:    false,
		},
		"creates a directory": {
			fileName: "subdir/",
			isDir:    true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			destDir := t.TempDir()
			reader := strings.NewReader(tc.content)

			path, err := ExtractAndWriteFile(reader, destDir, tc.fileName, tc.isDir, 0644)
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(path, destDir))

			if tc.isDir {
				info, err := os.Stat(path)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			} else {
				content, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.Equal(t, tc.content, string(content))
			}
		})
	}

	t.Run("rejects path traversal", func(t *testing.T) {
		destDir := t.TempDir()
		reader := strings.NewReader("malicious")
		_, err := ExtractAndWriteFile(reader, destDir, "../../../etc/passwd", false, 0644)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "illegal file path")
	})
}

// TODO: Refactor to use afero.Fs once Unzip accepts it. Currently uses t.TempDir() which is safe.
func Test_Unzip(t *testing.T) {
	t.Run("extracts a zip archive", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		// Create a zip file
		zipPath := filepath.Join(srcDir, "test.zip")
		zipFile, err := os.Create(zipPath)
		require.NoError(t, err)

		w := zip.NewWriter(zipFile)
		f, err := w.Create("test.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("zip content"))
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)
		err = zipFile.Close()
		require.NoError(t, err)

		files, err := Unzip(zipPath, destDir)
		require.NoError(t, err)
		assert.NotEmpty(t, files)

		content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "zip content", string(content))
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := Unzip("/nonexistent.zip", t.TempDir())
		assert.Error(t, err)
	})
}

// TODO: Refactor to use afero.Fs once UntarGzip accepts it. Currently uses t.TempDir() which is safe.
func Test_UntarGzip(t *testing.T) {
	t.Run("extracts a tar.gz archive", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		// Create a tar.gz file
		tgzPath := filepath.Join(srcDir, "test.tar.gz")
		tgzFile, err := os.Create(tgzPath)
		require.NoError(t, err)

		gw := gzip.NewWriter(tgzFile)
		tw := tar.NewWriter(gw)

		content := []byte("tar content")
		hdr := &tar.Header{
			Name:     "test.txt",
			Mode:     0644,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		err = tw.WriteHeader(hdr)
		require.NoError(t, err)
		_, err = tw.Write(content)
		require.NoError(t, err)

		err = tw.Close()
		require.NoError(t, err)
		err = gw.Close()
		require.NoError(t, err)
		err = tgzFile.Close()
		require.NoError(t, err)

		files, err := UntarGzip(tgzPath, destDir)
		require.NoError(t, err)
		assert.NotEmpty(t, files)

		data, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "tar content", string(data))
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := UntarGzip("/nonexistent.tar.gz", t.TempDir())
		assert.Error(t, err)
	})
}
