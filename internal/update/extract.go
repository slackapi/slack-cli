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

package update

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/archiveutil"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// ExtractArchive extracts the contents of the archive of the archive at src into the dest directory and returns
// the extracted file that it believes is the newly upgraded binary. Supports zip and tar.gz formats, which
// is determined by the file extension.
func ExtractArchive(src, dest string) (string, error) {
	if strings.HasSuffix(src, ".zip") {
		extractedFiles, err := archiveutil.Unzip(src, dest)
		if err != nil {
			return "", slackerror.New(fmt.Sprintf("Could not extract file: %s", src))
		}
		return findBinary(dest, extractedFiles)
	}
	if strings.HasSuffix(src, ".tar.gz") {
		extractedFiles, err := archiveutil.UntarGzip(src, dest)
		if err != nil {
			return "", slackerror.New(fmt.Sprintf("Could not extract file: %s", src))
		}
		return findBinary(dest, extractedFiles)
	}
	return "", slackerror.New(fmt.Sprintf("Unrecognized extension for file: %s", src))
}

// findBinary detects which of the extractedFiles extracted into dest is the executable binary.
func findBinary(dest string, extractedFiles []string) (string, error) {
	var updatedBinaryPath string

	switch len(extractedFiles) {
	case 0:
		return "", slackerror.New("no files extracted")
	case 1:
		updatedBinaryPath = extractedFiles[0]

	default:
		for _, fullFilePath := range extractedFiles {
			relativeFilePath, err := filepath.Rel(dest, fullFilePath)
			if err != nil {
				continue
			}
			if strings.Contains(relativeFilePath, "bin") {
				updatedBinaryPath = fullFilePath
			}
		}
	}

	if updatedBinaryPath == "" {
		return "", slackerror.New(fmt.Sprintf("could not detect binary from extractedFiles: %v", extractedFiles))
	}

	return updatedBinaryPath, nil
}
