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
	"os"
	"path/filepath"
	"runtime"

	"github.com/slackapi/slack-cli/internal/shared"
)

// isInstallScriptFunc checks if the running binary was placed by an installation script
var isInstallScriptFunc = func(clients *shared.ClientFactory) bool {
	executablePath, err := os.Executable()
	if err != nil {
		return false
	}
	homeDir, err := clients.Os.UserHomeDir()
	if err != nil {
		return false
	}
	return isInstallScript(executablePath, homeDir, runtime.GOOS)
}

// isInstallScript checks if the executable resolves to a binary in a folder that
// the installation scripts write to. Binaries placed by package managers or other
// tools are managed by those tools and should not be replaced by a self-update.
func isInstallScript(executablePath string, homeDir string, goos string) bool {
	binaryPath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return false
	}
	for _, dir := range installScriptDirs(homeDir, goos) {
		installDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			continue
		}
		if filepath.Dir(binaryPath) == installDir {
			return true
		}
	}
	return false
}

// installScriptDirs returns the folders where the installation scripts place the binary
//
// These match scripts/install.sh and scripts/install-windows.ps1
func installScriptDirs(homeDir string, goos string) []string {
	if goos == "windows" {
		return []string{
			filepath.Join(homeDir, "AppData", "Local", "slack-cli", "bin"),
			filepath.Join(homeDir, ".slack-cli", "bin"),
		}
	}
	return []string{
		filepath.Join(homeDir, ".slack", "bin"),
	}
}
