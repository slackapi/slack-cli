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

package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const scriptTemplate = "#!/bin/bash\necho %s\n"

type testFile struct {
	path    string
	content string
}

type testData struct {
	archiveFiles []testFile
	compressFn   func([]testFile, string, *testing.T)
	archiveName  string
}

func TestDownload(t *testing.T) {
	badCLIDownloadURL := "https://downloads.slack-edge.com/slack-cli/fake.zip"
	dir := t.TempDir()
	dstFilePath := filepath.Join(dir, "slack")
	err := download(badCLIDownloadURL, dstFilePath)
	require.Error(t, err, "Should return an error")
}

func TestUpgradeFromLocalFile_multiFile(t *testing.T) {
	oldVersion := "v1.0.1"
	newVersion := "v2.0.0"

	filesToArchive := []testFile{
		{
			path:    "bin/slack",
			content: fmt.Sprintf(scriptTemplate, newVersion),
		},
		{
			path:    "README.md",
			content: "readme file",
		},
		{
			path:    "LICENSE",
			content: "MIT",
		},
	}

	testCases := []testData{
		{
			archiveFiles: filesToArchive,
			archiveName:  "slack-v2.0.0.zip",
			compressFn:   zipArchive,
		},
		{
			archiveFiles: filesToArchive,
			archiveName:  "slack-v2.0.0.tar.gz",
			compressFn:   tarGzArchive,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.archiveName, func(t *testing.T) {
			dir := t.TempDir()
			archivePath := filepath.Join(dir, testCase.archiveName)
			testCase.compressFn(testCase.archiveFiles, archivePath, t)
			oldBinary := filepath.Join(dir, "slack")
			err := os.WriteFile(oldBinary, []byte(fmt.Sprint(oldVersion)), 0700)
			if err != nil {
				t.Errorf("failed to write old binary, %s", err)
				return
			}
			cliUpgradeData := cliUpgrade{
				UpgradeArchivePath: archivePath,
				ExistingBinaryPath: oldBinary,
				ExecutablePath:     oldBinary,
				CurrentVersion:     oldVersion,
				NewVersion:         newVersion,
				PrintDebug:         func(format string, a ...interface{}) {},
			}
			err = UpgradeFromLocalFile(cliUpgradeData)
			if err != nil {
				t.Errorf("upgrade failed: %s", err)
			}

			verifyFileContainsVersion(oldBinary, "2.0.0", t)
			// TODO(mcodik) should this just be bin/2.0.0/slack? it keeps bin/ since its in that dir in the archive
			// compare with TestAutoUpdate_zipFile_singleFile
			fileInVersionedFolder := filepath.Join(dir, "bin", "v2.0.0", "bin", "slack")
			verifyFileContainsVersion(fileInVersionedFolder, "2.0.0", t)
			fileInBackupFolder := filepath.Join(dir, "bin", "backups", "v1.0.1", "slack")
			verifyFileContainsVersion(fileInBackupFolder, "1.0.1", t)
		})
	}
}

func TestUpgradeFromLocalFile_singleFile(t *testing.T) {
	oldVersion := "v1.0.1"
	newVersion := "v2.0.0"

	filesToArchive := []testFile{
		{
			path:    "slack",
			content: fmt.Sprintf(scriptTemplate, newVersion),
		},
	}

	testCases := []testData{
		{
			archiveFiles: filesToArchive,
			archiveName:  "slack-v2.0.0.zip",
			compressFn:   zipArchive,
		},
		{
			archiveFiles: filesToArchive,
			archiveName:  "slack-v2.0.0.tar.gz",
			compressFn:   tarGzArchive,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.archiveName, func(t *testing.T) {
			dir := t.TempDir()
			archivePath := filepath.Join(dir, testCase.archiveName)
			testCase.compressFn(testCase.archiveFiles, archivePath, t)
			oldBinary := filepath.Join(dir, "slack")
			err := os.WriteFile(oldBinary, []byte(fmt.Sprint(oldVersion)), 0700)
			if err != nil {
				t.Errorf("failed to write old binary, %s", err)
				return
			}
			cliUpgradeData := cliUpgrade{
				UpgradeArchivePath: archivePath,
				ExistingBinaryPath: oldBinary,
				ExecutablePath:     oldBinary,
				CurrentVersion:     oldVersion,
				NewVersion:         newVersion,
				PrintDebug:         func(format string, a ...interface{}) {},
			}
			err = UpgradeFromLocalFile(cliUpgradeData)
			if err != nil {
				t.Errorf("upgrade failed: %s", err)
			}

			verifyFileContainsVersion(oldBinary, "2.0.0", t)
			fileInVersionedFolder := filepath.Join(dir, "bin", "v2.0.0", "slack")
			verifyFileContainsVersion(fileInVersionedFolder, "2.0.0", t)
			fileInBackupFolder := filepath.Join(dir, "bin", "backups", "v1.0.1", "slack")
			verifyFileContainsVersion(fileInBackupFolder, "1.0.1", t)
		})
	}
}

func TestUpgradeFromLocalFile_restoresAfterBadUpgrade(t *testing.T) {
	oldVersion := "v1.0.0"
	filesToArchive := []testFile{
		{
			path:    "slack",
			content: fmt.Sprintf(scriptTemplate, "vBadVersion"),
		},
	}
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "slack-v2.0.0.zip")
	zipArchive(filesToArchive, archivePath, t)
	oldBinary := filepath.Join(dir, "slack")
	err := os.WriteFile(oldBinary, []byte(fmt.Sprint(oldVersion)), 0700)
	if err != nil {
		t.Errorf("failed to write old binary, %s", err)
		return
	}
	cliUpgradeData := cliUpgrade{
		UpgradeArchivePath: archivePath,
		ExistingBinaryPath: oldBinary,
		ExecutablePath:     oldBinary,
		CurrentVersion:     oldVersion,
		NewVersion:         "v2.0.0",
		PrintDebug:         func(format string, a ...interface{}) {},
	}
	err = UpgradeFromLocalFile(cliUpgradeData)
	if err == nil {
		t.Errorf("upgrade should have failed due to version mismatch")
	}
	require.Contains(t, err.Error(), "versions did not match")
	verifyFileContainsVersion(oldBinary, oldVersion, t)
}

func verifyFileContainsVersion(file, version string, t *testing.T) {
	contents, err := os.ReadFile(file)
	if err != nil {
		t.Errorf("couldnt read file: %s", err)
		return
	}
	strContents := string(contents)
	require.Contains(t, strContents, version,
		"expected file %s to contain %s, but instead was '%s'", file, version, strContents)
}

// zip creates a new zip archive at destination containing a single
// file at path with the given contents
func zipArchive(files []testFile, destination string, t *testing.T) {
	destFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		t.Errorf("failed open zip: %s", err.Error())
		return
	}
	w := zip.NewWriter(destFile)
	for _, testFile := range files {
		header := &zip.FileHeader{}
		header.Name = testFile.path
		header.SetMode(0777)
		f, err := w.CreateHeader(header)
		if err != nil {
			t.Errorf("failed to add to zip: %s", err.Error())
			return
		}
		_, err = f.Write([]byte(testFile.content))
		if err != nil {
			t.Errorf("failed to write file: %s", err.Error())
			return
		}
	}
	err = w.Close()
	if err != nil {
		t.Errorf("failed to close file: %s", err.Error())
	}
}

func tarGzArchive(files []testFile, destination string, t *testing.T) {
	destFile, err := os.Create(destination)
	if err != nil {
		t.Errorf("failed open tar.gz: %s", err.Error())
		return
	}
	defer destFile.Close()
	gw := gzip.NewWriter(destFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		contentBytes := []byte(file.content)
		header := &tar.Header{
			Name: file.path,
			Size: int64(len(contentBytes)),
			Mode: 0777,
		}
		err = tw.WriteHeader(header)
		if err != nil {
			t.Errorf("error writing header: %s", err)
			return
		}
		rdr := bytes.NewReader(contentBytes)
		_, err = io.Copy(tw, rdr)
		if err != nil {
			t.Errorf("error writing file: %s", err)
			return
		}
	}
}
