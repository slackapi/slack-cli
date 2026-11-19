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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
