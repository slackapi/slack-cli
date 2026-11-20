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
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
)

// AutoUpdateCommandName is the hidden command that installs an update in the background.
const AutoUpdateCommandName = "_autoupdate"

const autoUpdateLogFileName = "autoupdate.log"

// startBackgroundUpdate starts a detached process that installs the targetVersion and returns
// without waiting for it. The process runs the binary only from a folder that the installation
// scripts write to, so nothing starts when the running path follows a binary that an update
// moved to the backups folder. It runs from that folder so a project in the current directory
// isn't read or changed. Output goes to the autoupdate log in the logs folder, which is opened
// with os because the process writes to a file descriptor.
func startBackgroundUpdate(ctx context.Context, clients *shared.ClientFactory, targetVersion string) error {
	binaryPath, ok := installScriptPath(ctx, clients)
	if !ok {
		return errors.New("the running binary is not in a folder that the installation scripts write to")
	}
	args, err := backgroundUpdateArgs(clients, targetVersion)
	if err != nil {
		return err
	}
	logsDir, err := clients.Config.SystemConfig.LogsDir(ctx)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(logsDir, autoUpdateLogFileName), os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	command := exec.Command(binaryPath, args...)
	command.Dir = filepath.Dir(binaryPath)
	command.Stdout = logFile
	command.Stderr = logFile
	detach(command)
	if err := command.Start(); err != nil {
		return err
	}
	if err := command.Process.Release(); err != nil {
		clients.IO.PrintDebug(ctx, "Failed to release the update process: %s", err)
	}
	return nil
}

// backgroundUpdateArgs returns the arguments of the command that installs the targetVersion.
// A custom config directory is kept as an absolute path so the update is recorded where the
// next command reads it.
func backgroundUpdateArgs(clients *shared.ClientFactory, targetVersion string) ([]string, error) {
	args := []string{AutoUpdateCommandName, targetVersion}
	configDirFlag := strings.TrimSpace(clients.Config.ConfigDirFlag)
	if configDirFlag == "" {
		return args, nil
	}
	configDir, err := filepath.Abs(configDirFlag)
	if err != nil {
		return nil, err
	}
	return append(args, "--config-dir", configDir), nil
}
