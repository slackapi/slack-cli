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

package goutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldIgnore(t *testing.T) {
	tests := map[string]struct {
		val        string
		list       []string
		ignoreFunc func(string) bool
		expected   bool
	}{
		"returns true when value is in list": {
			val:      "node_modules",
			list:     []string{"node_modules", ".git"},
			expected: true,
		},
		"returns false when value is not in list": {
			val:      "src",
			list:     []string{"node_modules", ".git"},
			expected: false,
		},
		"returns false with empty list and nil ignoreFunc": {
			val:      "anything",
			list:     []string{},
			expected: false,
		},
		"returns false with nil list and nil ignoreFunc": {
			val:      "anything",
			list:     nil,
			expected: false,
		},
		"returns true when ignoreFunc returns true": {
			val:  "secret.txt",
			list: []string{},
			ignoreFunc: func(s string) bool {
				return s == "secret.txt"
			},
			expected: true,
		},
		"returns false when ignoreFunc returns false": {
			val:  "readme.md",
			list: []string{},
			ignoreFunc: func(s string) bool {
				return s == "secret.txt"
			},
			expected: false,
		},
		"list match takes priority over ignoreFunc": {
			val:  "node_modules",
			list: []string{"node_modules"},
			ignoreFunc: func(s string) bool {
				return false
			},
			expected: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := ShouldIgnore(tc.val, tc.list, tc.ignoreFunc)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCopy(t *testing.T) {
	t.Run("copies a regular file", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		srcFile := filepath.Join(srcDir, "test.txt")
		dstFile := filepath.Join(dstDir, "test.txt")

		err := os.WriteFile(srcFile, []byte("hello world"), 0644)
		require.NoError(t, err)

		err = Copy(srcFile, dstFile)
		require.NoError(t, err)

		content, err := os.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(content))
	})

	t.Run("returns error for non-existent source", func(t *testing.T) {
		dstDir := t.TempDir()
		err := Copy("/nonexistent/file.txt", filepath.Join(dstDir, "file.txt"))
		assert.Error(t, err)
	})
}

func TestCopySymLink(t *testing.T) {
	t.Run("copies a symbolic link", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		// Create a target file and a symlink to it
		targetFile := filepath.Join(srcDir, "target.txt")
		err := os.WriteFile(targetFile, []byte("target content"), 0644)
		require.NoError(t, err)

		srcLink := filepath.Join(srcDir, "link.txt")
		err = os.Symlink(targetFile, srcLink)
		require.NoError(t, err)

		dstLink := filepath.Join(dstDir, "link.txt")
		err = CopySymLink(srcLink, dstLink)
		require.NoError(t, err)

		// Verify the symlink target is the same
		linkTarget, err := os.Readlink(dstLink)
		require.NoError(t, err)
		assert.Equal(t, targetFile, linkTarget)
	})
}

func TestCopyDirectory(t *testing.T) {
	t.Run("copies directory structure", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		// Create source structure
		err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("file1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("file2"), 0644)
		require.NoError(t, err)

		err = CopyDirectory(CopyDirectoryOpts{Src: srcDir, Dst: dstDir})
		require.NoError(t, err)

		// Verify files were copied
		content, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
		require.NoError(t, err)
		assert.Equal(t, "file1", string(content))

		content, err = os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
		require.NoError(t, err)
		assert.Equal(t, "file2", string(content))
	})

	t.Run("ignores specified files", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		err := os.WriteFile(filepath.Join(srcDir, "keep.txt"), []byte("keep"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(srcDir, "ignore.txt"), []byte("ignore"), 0644)
		require.NoError(t, err)

		err = CopyDirectory(CopyDirectoryOpts{
			Src:         srcDir,
			Dst:         dstDir,
			IgnoreFiles: []string{"ignore.txt"},
		})
		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(dstDir, "keep.txt"))
		assert.NoFileExists(t, filepath.Join(dstDir, "ignore.txt"))
	})

	t.Run("ignores specified directories", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "dst")

		err := os.MkdirAll(filepath.Join(srcDir, "keep_dir"), 0755)
		require.NoError(t, err)
		err = os.MkdirAll(filepath.Join(srcDir, "node_modules"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(srcDir, "keep_dir", "file.txt"), []byte("keep"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(srcDir, "node_modules", "file.txt"), []byte("ignore"), 0644)
		require.NoError(t, err)

		err = CopyDirectory(CopyDirectoryOpts{
			Src:               srcDir,
			Dst:               dstDir,
			IgnoreDirectories: []string{"node_modules"},
		})
		require.NoError(t, err)

		assert.FileExists(t, filepath.Join(dstDir, "keep_dir", "file.txt"))
		assert.NoDirExists(t, filepath.Join(dstDir, "node_modules"))
	})

	t.Run("returns error for non-existent source", func(t *testing.T) {
		dstDir := filepath.Join(t.TempDir(), "dst")
		err := CopyDirectory(CopyDirectoryOpts{
			Src: "/nonexistent/path",
			Dst: dstDir,
		})
		assert.Error(t, err)
	})
}
