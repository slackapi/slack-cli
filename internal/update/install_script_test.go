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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_isInstallScript(t *testing.T) {
	tests := map[string]struct {
		goos           string
		executablePath string
		expected       bool
	}{
		"binary placed by the install script": {
			goos:           "darwin",
			executablePath: "home/.slack/bin/slack",
			expected:       true,
		},
		"symbolic link to the binary placed by the install script": {
			goos:           "darwin",
			executablePath: "usr/local/bin/slack",
			expected:       true,
		},
		"binary placed by the development install script": {
			goos:           "darwin",
			executablePath: "home/.slack/dev-build/bin/slack",
			expected:       false,
		},
		"binary placed in a subfolder of the install script folder": {
			goos:           "darwin",
			executablePath: "home/.slack/bin/4.14.0/slack",
			expected:       false,
		},
		"symbolic link to a binary placed by a package manager": {
			goos:           "darwin",
			executablePath: "opt/homebrew/bin/slack",
			expected:       false,
		},
		"binary placed by the windows install script": {
			goos:           "windows",
			executablePath: "home/AppData/Local/slack-cli/bin/slack.exe",
			expected:       true,
		},
		"binary placed in the fallback folder of the windows install script": {
			goos:           "windows",
			executablePath: "home/.slack-cli/bin/slack.exe",
			expected:       true,
		},
		"binary that does not exist": {
			goos:           "darwin",
			executablePath: "home/.slack/bin/missing",
			expected:       false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			for _, file := range []string{
				"home/.slack/bin/slack",
				"home/.slack/bin/4.14.0/slack",
				"home/.slack/dev-build/bin/slack",
				"home/.slack-cli/bin/slack.exe",
				"home/AppData/Local/slack-cli/bin/slack.exe",
				"opt/homebrew/Caskroom/slack-cli/4.14.0/bin/slack",
			} {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(file)), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, file), []byte{}, 0o755))
			}
			for link, target := range map[string]string{
				"opt/homebrew/bin/slack": "opt/homebrew/Caskroom/slack-cli/4.14.0/bin/slack",
				"usr/local/bin/slack":    "home/.slack/bin/slack",
			} {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(link)), 0o755))
				require.NoError(t, os.Symlink(filepath.Join(dir, target), filepath.Join(dir, link)))
			}
			actual := isInstallScript(filepath.Join(dir, tc.executablePath), filepath.Join(dir, "home"), tc.goos)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
