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

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared"
)

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
	configDir, err := backgroundUpdateConfigDir(clients)
	if err != nil {
		return err
	}
	logsDir, err := clients.Config.SystemConfig.LogsDir(ctx)
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(logsDir, autoUpdateLogFileName), os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	command := exec.Command(binaryPath, backgroundUpdateArgs(targetVersion, configDir)...)
	command.Dir = filepath.Dir(binaryPath)
	if configDir != "" {
		command.Env = append(os.Environ(), config.SlackConfigDirEnv+"="+configDir)
	}
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

// backgroundUpdateArgs returns the arguments of the command that installs the targetVersion,
// with the custom config directory when one is set.
func backgroundUpdateArgs(targetVersion string, configDir string) []string {
	args := []string{AutoUpdateCommandName, targetVersion}
	if configDir == "" {
		return args
	}
	return append(args, "--config-dir", configDir)
}

// backgroundUpdateConfigDir returns the custom config directory as an absolute path, so the
// update is recorded where the next command reads it, or an empty string when none is set.
// The process also gets it as SLACK_CONFIG_DIR, since that variable overrides the flag and a
// relative value would resolve from the folder the process runs in.
func backgroundUpdateConfigDir(clients *shared.ClientFactory) (string, error) {
	configDirFlag := strings.TrimSpace(clients.Config.ConfigDirFlag)
	if configDirFlag == "" {
		return "", nil
	}
	return filepath.Abs(configDirFlag)
}
