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
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/slackapi/slack-cli/internal/shared"
)

// autoUpdateLockFileName is the file in the Slack config directory that a background update
// locks while it installs, so only one update installs at a time
const autoUpdateLockFileName = "autoupdate.lock"

// installUpdateFunc installs an update of the CLI and is replaced in tests
var installUpdateFunc = func(ctx context.Context, c *CLIDependency) error {
	return c.InstallUpdate(ctx)
}

// AutoUpdate installs the targetVersion of the CLI over the currentVersion and records the
// result for the next command to report. It returns right away while another update holds the lock.
func AutoUpdate(ctx context.Context, clients *shared.ClientFactory, currentVersion string, targetVersion string) error {
	configDir, err := clients.Config.SystemConfig.SlackConfigDir(ctx)
	if err != nil {
		return err
	}
	// The lock is an OS file lock, so it's released if this process exits early. The lock file
	// stays on disk, since removing it could let another update lock a new file at the same path.
	lock := flock.New(filepath.Join(configDir, autoUpdateLockFileName))
	locked, err := lock.TryLock()
	if err != nil {
		return err
	}
	if !locked {
		clients.IO.PrintDebug(ctx, "Skipping the update since another update is installing")
		return nil
	}
	defer func() {
		_ = lock.Unlock()
	}()

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
