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

package archiveutil

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip decompresses a zip file and returns extracted files list
func Unzip(src, dest string) ([]string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	_ = os.MkdirAll(dest, 0755)

	extractedFiles := []string{}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()
		extractedFile, err := ExtractAndWriteFile(rc, dest, f.Name, f.FileInfo().IsDir(), f.FileInfo().Mode())
		if err != nil {
			return nil, err
		}
		extractedFiles = append(extractedFiles, extractedFile)
	}

	return extractedFiles, nil
}

// UntarGzip decompresses a gzip'ed tar archive and returns extracted files list
func UntarGzip(src, dest string) ([]string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	gzipReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(gzipReader)
	extractedFiles := []string{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeDir {
			extractedFile, err := ExtractAndWriteFile(tarReader, dest, header.Name, header.Typeflag == tar.TypeDir, header.FileInfo().Mode())
			if err != nil {
				return nil, err
			}
			extractedFiles = append(extractedFiles, extractedFile)
		}
	}

	return extractedFiles, nil
}

// ExtractAndWriteFile extracts the file backing the io.Reader into the right location. It is the caller's responsibility to close that
// the reader if needed (which is why this method doesnt take an io.ReadCloser)
func ExtractAndWriteFile(r io.Reader, destDir, fileName string, isDir bool, mode os.FileMode) (string, error) {
	path := filepath.Join(destDir, fileName)

	// Check for ZipSlip (Directory traversal)
	if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal file path: %s", path)
	}

	if isDir {
		_ = os.MkdirAll(path, 0700)
	} else {
		_ = os.MkdirAll(filepath.Dir(path), 0700)
		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return "", err
		}
		defer func() {
			if err := destFile.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(destFile, r)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}
