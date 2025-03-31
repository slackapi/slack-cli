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

package deputil

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

// GetGitVersion returns the installed version of Git
func GetGitVersion() (string, error) {
	version, err := exec.Command("git", "--version").Output()
	if err != nil {
		return "", slackerror.New(slackerror.ErrGitNotFound)
	}
	return string(version), nil
}

// GetDenoVersion shells out to `deno` to determine the version of the runtime. This should only be called once in root.go!
func GetDenoVersion() string {
	denoVersion, err := exec.Command("deno", "--version").Output()
	var denoVersionFormatted string
	if err != nil {
		return "unknown"
	}
	// Get Deno version
	// Before Deno v1.5.1, information didn't include OS
	if strings.Contains(string(denoVersion), "(") {
		denoVersionFormatted = string(denoVersion)[5:strings.IndexByte(string(denoVersion), '(')]
	} else {
		var eol = "\n"
		if runtime.GOOS == "windows" {
			eol = "\r\n"
		}
		denoVersionFormatted = strings.Split(string(denoVersion), eol)[0][5:]
	}
	denoVersionFormatted = "v" + strings.TrimSpace(denoVersionFormatted)

	return denoVersionFormatted
}
