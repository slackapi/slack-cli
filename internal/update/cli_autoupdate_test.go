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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/flock"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CLIDependency_InstallUpdate(t *testing.T) {
	tests := map[string]struct {
		canceled         bool
		installedVersion string
		lockHeld         bool
		lockUnavailable  bool
		movedBinary      bool
		releasedVersion  string
		expectedErr      string
		expectedInstalls int
		expectedOutput   []string
	}{
		"installs the update while holding the lock": {
			installedVersion: "v4.11.0",
			expectedInstalls: 1,
		},
		"installs when the lock can't be created": {
			installedVersion: "v4.11.0",
			lockUnavailable:  true,
			expectedInstalls: 1,
		},
		"returns without installing when canceled while waiting": {
			installedVersion: "v4.11.0",
			lockHeld:         true,
			canceled:         true,
			expectedInstalls: 0,
			expectedOutput:   []string{"Waiting for another update to finish"},
			expectedErr:      "context canceled",
		},
		"returns without installing when another update moved the binary": {
			installedVersion: "v4.11.0",
			lockHeld:         true,
			movedBinary:      true,
			expectedInstalls: 0,
			expectedErr:      "the running binary is not in a folder that the installation scripts write to",
		},
		"skips the install when the binary reports the target": {
			installedVersion: "v4.12.0",
			expectedInstalls: 0,
			expectedOutput:   []string{"You are using the latest Slack CLI version"},
		},
		"waits for another update and skips the install after it finishes": {
			installedVersion: "v4.11.0",
			lockHeld:         true,
			releasedVersion:  "v4.12.0",
			expectedInstalls: 0,
			expectedOutput:   []string{"Waiting for another update to finish", "You are using the latest Slack CLI version"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(slackcontext.MockContext(t.Context()))
			defer cancel()
			homeDir := t.TempDir()
			binaryPath := filepath.Join(homeDir, ".slack", "bin", "slack")
			executablePath := binaryPath
			if tc.movedBinary {
				executablePath = filepath.Join(homeDir, ".slack", "bin", "backups", "v4.11.0", "slack")
			}
			require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
			require.NoError(t, os.WriteFile(executablePath, fmt.Appendf(nil, scriptTemplate, "Using slack "+tc.installedVersion), 0o755))
			lockPath := filepath.Join(filepath.Dir(binaryPath), "autoupdate.lock")
			if tc.lockUnavailable {
				require.NoError(t, os.Symlink(filepath.Join(t.TempDir(), "missing", "autoupdate.lock"), lockPath))
			}
			holder := flock.New(lockPath)
			if tc.lockHeld {
				locked, err := holder.TryLock()
				require.NoError(t, err)
				require.True(t, locked)
				t.Cleanup(func() { _ = holder.Unlock() })
			}
			if tc.releasedVersion != "" {
				time.AfterFunc(100*time.Millisecond, func() {
					assert.NoError(t, os.WriteFile(binaryPath, fmt.Appendf(nil, scriptTemplate, "Using slack "+tc.releasedVersion), 0o755))
					assert.NoError(t, holder.Unlock())
				})
			}
			if tc.canceled {
				time.AfterFunc(100*time.Millisecond, cancel)
			}
			clientsMock := shared.NewClientsMock()
			clientsMock.Os.On("Executable").Return(executablePath, nil)
			clientsMock.Os.On("UserHomeDir").Return(homeDir, nil)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			installs := 0
			originalInstallUpdateFunc := installUpdateFunc
			t.Cleanup(func() {
				installUpdateFunc = originalInstallUpdateFunc
			})
			installUpdateFunc = func(ctx context.Context, c *CLIDependency) error {
				installs++
				if !tc.lockUnavailable {
					locked, err := flock.New(lockPath).TryLock()
					require.NoError(t, err)
					assert.False(t, locked, "the lock is held while installing")
				}
				return nil
			}
			cli := NewCLIDependency(clients, "v4.11.0")
			cli.releaseInfo = &LatestCLIRelease{Version: "4.12.0"}
			err := cli.InstallUpdate(ctx)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedInstalls, installs)
			for _, expectedOutput := range tc.expectedOutput {
				assert.Contains(t, clientsMock.GetCombinedOutput(), expectedOutput)
			}
		})
	}
}

func Test_CLI_getUpdateFileName(t *testing.T) {
	tests := map[string]struct {
		version          string
		operatingSystem  string
		architecture     string
		expectedFilename string
		expectedErrorF   string
	}{
		"darwin production x86_64": {
			version:          "3.4.5",
			operatingSystem:  "darwin",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5_macOS_amd64.zip",
		},
		"darwin development x86_64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "darwin",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_macOS_amd64.zip",
		},
		"darwin production aarch64": {
			version:          "3.4.5",
			operatingSystem:  "darwin",
			architecture:     "arm64",
			expectedFilename: "slack_cli_3.4.5_macOS_arm64.zip",
		},
		"darwin development aarch64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "darwin",
			architecture:     "arm64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_macOS_arm64.zip",
		},
		"darwin production universal": {
			version:          "3.4.5",
			operatingSystem:  "darwin",
			architecture:     "fallback",
			expectedFilename: "slack_cli_3.4.5_macOS_64-bit.zip",
		},
		"darwin development universal": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "darwin",
			architecture:     "fallback",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_macOS_64-bit.zip",
		},
		"linux production x86_64": {
			version:          "3.4.5",
			operatingSystem:  "linux",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5_linux_amd64.tar.gz",
		},
		"linux development x86_64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "linux",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_linux_amd64.tar.gz",
		},
		"linux production arm64": {
			version:          "3.4.5",
			operatingSystem:  "linux",
			architecture:     "arm64",
			expectedFilename: "slack_cli_3.4.5_linux_arm64.tar.gz",
		},
		"linux development arm64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "linux",
			architecture:     "arm64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_linux_arm64.tar.gz",
		},
		"linux production unknown arch fallback": {
			version:          "3.4.5",
			operatingSystem:  "linux",
			architecture:     "riscv64",
			expectedFilename: "slack_cli_3.4.5_linux_64-bit.tar.gz",
		},
		"windows production x86_64": {
			version:          "3.4.5",
			operatingSystem:  "windows",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5_windows_64-bit.zip",
		},
		"windows development x86_64": {
			version:          "3.4.5-6-badaabad",
			operatingSystem:  "windows",
			architecture:     "amd64",
			expectedFilename: "slack_cli_3.4.5-6-badaabad_windows_64-bit.zip",
		},
		"unknown production errors": {
			version:         "3.4.5",
			operatingSystem: "dragonfly",
			architecture:    "amd64",
			expectedErrorF:  "auto-updating for the operating system (dragonfly) is unsupported",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			filename, err := getUpdateFileName(tc.version, tc.operatingSystem, tc.architecture)
			if tc.expectedErrorF != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expectedErrorF)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedFilename, filename)
			}
		})
	}
}

func Test_download(t *testing.T) {
	tests := map[string]struct {
		handler         http.HandlerFunc
		timeout         time.Duration
		expectedContent string
		expectedErr     error
	}{
		"saves the downloaded file": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("example-archive"))
			},
			timeout:         time.Minute,
			expectedContent: "example-archive",
		},
		"stops a stalled download once the context expires": {
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			},
			timeout:     10 * time.Millisecond,
			expectedErr: context.DeadlineExceeded,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()
			dstFilePath := filepath.Join(t.TempDir(), "slack.tar.gz")
			err := download(ctx, server.URL, dstFilePath)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)
			content, err := os.ReadFile(dstFilePath)
			require.NoError(t, err)
			require.Equal(t, tc.expectedContent, string(content))
		})
	}
}
