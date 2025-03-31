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

package goutils

// Copied from https://github.com/plus3it/gorecurcopy/edit/master/gorecurcopy.go
// Tweaked to ignore node_modules

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type CopyDirectoryOpts struct {
	Src                   string
	Dst                   string
	IgnoreFiles           []string
	IgnoreDirectories     []string
	IgnoreFunc            func(string) bool
	RelativeParentDirPath string
}

// CopyDirectory recursively copies a src directory to a destination.
func CopyDirectory(opts CopyDirectoryOpts) error {
	// ensure Dst exists
	var err = os.MkdirAll(opts.Dst, 0755)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(opts.Src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(opts.Src, entry.Name())
		destPath := filepath.Join(opts.Dst, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		var entryName = filepath.Join(opts.RelativeParentDirPath, entry.Name())
		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if ShouldIgnore(entryName, opts.IgnoreDirectories, opts.IgnoreFunc) {
				continue
			}
			if err := createDir(destPath, 0755); err != nil {
				return err
			}
			var relativeParentDirPath = filepath.Join(opts.RelativeParentDirPath, entry.Name())
			subOpts := CopyDirectoryOpts{Src: sourcePath, Dst: destPath, IgnoreDirectories: opts.IgnoreDirectories, IgnoreFiles: opts.IgnoreFiles, IgnoreFunc: opts.IgnoreFunc, RelativeParentDirPath: relativeParentDirPath}
			if err := CopyDirectory(subOpts); err != nil {
				return err
			}
		case os.ModeSymlink:
			if ShouldIgnore(entryName, opts.IgnoreFiles, opts.IgnoreFunc) {
				continue
			}
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if ShouldIgnore(entryName, opts.IgnoreFiles, opts.IgnoreFunc) {
				continue
			}
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		// `go test` fails on Windows even with this `if` supposedly
		// protecting the `syscall.Stat_t` and `os.Lchown` calls (not
		// available on windows). why?
		/*
			if runtime.GOOS != "windows" {
					stat, ok := fileInfo.Sys().(*syscall.Stat_t)
					if !ok {
						return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
					}
					if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
						return err
					}
			}
		*/

		info, err := entry.Info()
		if err != nil {
			return err
		}

		isSymlink := info.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Copy copies a src file to a dst file where src and dst are regular files.
func Copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func createDir(dir string, perm os.FileMode) error {
	if exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

// CopySymLink copies a symbolic link from src to dst.
func CopySymLink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

// ShouldIgnore returns true if a file or directory is in the exclusion list or ignore function
func ShouldIgnore(val string, list []string, ignoreFunc func(string) bool) bool {
	for _, listVal := range list {
		if val == listVal {
			return true
		}
	}
	if ignoreFunc != nil && ignoreFunc(val) {
		return true
	}
	return false
}
