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
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/slackapi/slack-cli/internal/shared"
)

const autoUpdateLockFileName = "autoupdate.lock"

const downloadTimeout = 10 * time.Minute

const lockRetryDelay = 250 * time.Millisecond

var installUpdateFunc = func(ctx context.Context, c *CLIDependency) error {
	return c.installUpdate(ctx)
}

var versionPattern = regexp.MustCompile(`v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?`)

// AutoUpdate installs the targetVersion of the CLI over the currentVersion and records the
// result for the next command to report. It returns without installing or recording while
// another update holds the lock next to the binary, or when the binary is already updated.
func AutoUpdate(ctx context.Context, clients *shared.ClientFactory, currentVersion string, targetVersion string) error {
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()
	lock, binaryPath, err := newUpdateLock(ctx, clients)
	if err != nil {
		return err
	}
	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		clients.IO.PrintDebug(ctx, "Skipping the update since another update is installing")
		return nil
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			clients.IO.PrintDebug(ctx, "Failed to release the update lock: %s", err)
		}
	}()
	if isInstalled(ctx, binaryPath, targetVersion) {
		clients.IO.PrintDebug(ctx, "Skipping the update since %s is already installed", targetVersion)
		return nil
	}

	cli := NewCLIDependency(clients, currentVersion)
	cli.releaseInfo = &LatestCLIRelease{Version: targetVersion}
	installErr := installUpdateFunc(ctx, cli)

	result := AutoUpdateResult{
		From: ensureVPrefix(currentVersion),
		To:   ensureVPrefix(targetVersion),
	}
	if installErr != nil {
		result.Error = installErr.Error()
	}
	writeErr := writeAutoUpdateResult(ctx, clients, result)
	return errors.Join(installErr, writeErr)
}

// newUpdateLock returns the lock that an update holds while it installs, next to the binary in
// the folder that the installation scripts write to, and the path to that binary. It errors when
// the running binary resolves outside of those folders, as it does on Linux after another update
// moved the running binary to the backups folder.
func newUpdateLock(ctx context.Context, clients *shared.ClientFactory) (*flock.Flock, string, error) {
	binaryPath, ok := installScriptPath(ctx, clients)
	if !ok {
		return nil, "", errors.New("the running binary is not in a folder that the installation scripts write to")
	}
	return flock.New(filepath.Join(filepath.Dir(binaryPath), autoUpdateLockFileName)), binaryPath, nil
}

// isInstalled reports whether the binary is at the targetVersion or newer, so an update that
// starts right after another one finished doesn't install again. SLACK_TEST_VERSION is left
// out so the binary reports its own version. A version that can't be read reports false.
func isInstalled(ctx context.Context, binaryPath string, targetVersion string) bool {
	command := exec.CommandContext(ctx, binaryPath, "--version")
	command.Env = slices.DeleteFunc(os.Environ(), func(envar string) bool {
		return strings.HasPrefix(envar, "SLACK_TEST_VERSION=")
	})
	output, err := command.Output()
	if err != nil {
		return false
	}
	installedVersion := versionPattern.FindString(string(output))
	newer, err := SemVerGreaterThan(targetVersion, installedVersion)
	return err == nil && !newer
}
