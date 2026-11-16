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
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/shared"
)

// isInstallScript checks if the running binary resolves to a binary in a folder
// that the installation scripts write to. Binaries placed by package managers or
// other tools are managed by those tools and should not be replaced by a self-update.
func isInstallScript(clients *shared.ClientFactory) bool {
	executablePath, err := clients.Os.Executable()
	if err != nil {
		return false
	}
	homeDir, err := clients.Os.UserHomeDir()
	if err != nil {
		return false
	}
	binaryPath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return false
	}
	installDirs := []string{
		filepath.Join(homeDir, ".slack", "bin"),                        // scripts/install.sh
		filepath.Join(homeDir, ".slack-cli", "bin"),                    // scripts/install-windows.ps1 fallback
		filepath.Join(homeDir, "AppData", "Local", "slack-cli", "bin"), // scripts/install-windows.ps1
	}
	for _, dir := range installDirs {
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
