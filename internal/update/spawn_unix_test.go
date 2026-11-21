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

//go:build !windows

package update

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

const envHelperProcess = "SLACK_TEST_HELPER_PROCESS"

type helperProcessReport struct {
	Args         []string `json:"args"`
	ConfigDirEnv string   `json:"config_dir_env"`
	Detached     bool     `json:"detached"`
	Dir          string   `json:"dir"`
}

func TestMain(m *testing.M) {
	if dir := os.Getenv(envHelperProcess); dir != "" {
		os.Exit(runHelperProcess(dir))
	}
	os.Exit(m.Run())
}

func Test_detach(t *testing.T) {
	command := exec.Command("slack")
	detach(command)
	require.NotNil(t, command.SysProcAttr)
	assert.True(t, command.SysProcAttr.Setsid, "the process starts in a new session")
}

func Test_startBackgroundUpdate(t *testing.T) {
	workingDir, err := os.Getwd()
	require.NoError(t, err)
	tests := map[string]struct {
		executablePath    string
		configDirFlag     string
		expectedArgs      []string
		expectedConfigDir string
		expectedErr       string
	}{
		"starts the binary in the install script folder": {
			executablePath: "home/.slack/bin/slack",
			expectedArgs:   []string{"_autoupdate", "4.12.0"},
		},
		"starts the binary in the install script folder from a symbolic link": {
			executablePath: "usr/local/bin/slack",
			expectedArgs:   []string{"_autoupdate", "4.12.0"},
		},
		"starts the binary with a custom config directory": {
			executablePath:    "home/.slack/bin/slack",
			configDirFlag:     "custom",
			expectedArgs:      []string{"_autoupdate", "4.12.0", "--config-dir", filepath.Join(workingDir, "custom")},
			expectedConfigDir: filepath.Join(workingDir, "custom"),
		},
		"skips a running binary that moved to the backups folder": {
			executablePath: "home/.slack/bin/backups/v4.11.0/slack",
			expectedErr:    "not in a folder that the installation scripts write to",
		},
		"skips a binary placed by a package manager": {
			executablePath: "opt/homebrew/bin/slack",
			expectedErr:    "not in a folder that the installation scripts write to",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			testBinary, err := os.Executable()
			require.NoError(t, err)
			installPath := filepath.Join(dir, "home/.slack/bin/slack")
			for _, file := range []string{
				"home/.slack/bin/slack",
				"home/.slack/bin/backups/v4.11.0/slack",
				"opt/homebrew/bin/slack",
			} {
				copyFile(t, testBinary, filepath.Join(dir, file))
			}
			linkPath := filepath.Join(dir, "usr/local/bin/slack")
			require.NoError(t, os.MkdirAll(filepath.Dir(linkPath), 0o755))
			require.NoError(t, os.Symlink(installPath, linkPath))
			logsDir := filepath.Join(dir, "home/.slack/logs")
			reportDir := filepath.Join(dir, "report")
			require.NoError(t, os.MkdirAll(reportDir, 0o755))
			t.Setenv(envHelperProcess, reportDir)
			t.Setenv("SLACK_CONFIG_DIR", tc.configDirFlag)
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Os.On("Executable").Return(filepath.Join(dir, tc.executablePath), nil)
			clientsMock.Os.On("UserHomeDir").Return(filepath.Join(dir, "home"), nil)
			clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.Anything)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			clients.Config.ConfigDirFlag = tc.configDirFlag
			systemConfigMock := &config.SystemConfigMock{}
			systemConfigMock.On("LogsDir", mock.Anything).Return(logsDir, nil)
			clients.Config.SystemConfig = systemConfigMock
			require.NoError(t, os.MkdirAll(logsDir, 0o755))
			if tc.expectedErr == "" {
				require.NoError(t, os.WriteFile(filepath.Join(logsDir, "autoupdate.log"), []byte("output of an earlier update\n"), 0o600))
			}

			err = startBackgroundUpdate(ctx, clients, "4.12.0")
			require.NoError(t, os.WriteFile(filepath.Join(reportDir, "started"), []byte{}, 0o600))

			reportPath := filepath.Join(reportDir, "finished")
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.NoFileExists(t, filepath.Join(logsDir, "autoupdate.log"), "no binary was started")
				return
			}
			require.NoError(t, err)
			require.Eventually(t, func() bool {
				_, err := os.Stat(reportPath)
				return err == nil
			}, 10*time.Second, 10*time.Millisecond, "the binary finishes after the start returned")
			data, err := os.ReadFile(reportPath)
			require.NoError(t, err)
			var report helperProcessReport
			require.NoError(t, json.Unmarshal(data, &report))
			resolvedInstallPath, err := filepath.EvalSymlinks(installPath)
			require.NoError(t, err)
			assert.Equal(t, resolvedInstallPath, report.Args[0])
			assert.Equal(t, tc.expectedArgs, report.Args[1:])
			resolvedDir, err := filepath.EvalSymlinks(report.Dir)
			require.NoError(t, err)
			assert.Equal(t, filepath.Dir(resolvedInstallPath), resolvedDir)
			assert.True(t, report.Detached, "the binary runs detached from the terminal")
			assert.Equal(t, tc.expectedConfigDir, report.ConfigDirEnv)
			require.Eventually(t, func() bool {
				logs, err := os.ReadFile(filepath.Join(logsDir, "autoupdate.log"))
				return err == nil && string(logs) == "helper process output\n"
			}, 10*time.Second, 10*time.Millisecond, "the log holds only the output of this update")
		})
	}
}

func copyFile(t *testing.T, source string, destination string) {
	require.NoError(t, os.MkdirAll(filepath.Dir(destination), 0o755))
	in, err := os.Open(source)
	require.NoError(t, err)
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	require.NoError(t, err)
	defer out.Close()
	_, err = io.Copy(out, in)
	require.NoError(t, err)
}

// runHelperProcess stands in for the binary that startBackgroundUpdate starts. It prints to
// the output, waits for the test to signal that the start returned, then reports its
// arguments, config directory, working directory, and whether it is detached. A start that
// waits for the process never signals it.
func runHelperProcess(dir string) int {
	_, _ = os.Stdout.WriteString("helper process output\n")
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(dir, "started")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			return 1
		}
		time.Sleep(10 * time.Millisecond)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return 1
	}
	data, err := json.Marshal(helperProcessReport{
		Args:         os.Args,
		ConfigDirEnv: os.Getenv("SLACK_CONFIG_DIR"),
		Detached:     isDetached(),
		Dir:          workingDir,
	})
	if err != nil {
		return 1
	}
	if err := os.WriteFile(filepath.Join(dir, "finished.tmp"), data, 0o600); err != nil {
		return 1
	}
	if err := os.Rename(filepath.Join(dir, "finished.tmp"), filepath.Join(dir, "finished")); err != nil {
		return 1
	}
	return 0
}

func isDetached() bool {
	sid, err := unix.Getsid(0)
	return err == nil && sid == os.Getpid()
}
