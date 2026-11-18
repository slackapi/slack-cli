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
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
)

// UpdateNotification checks for an update (non-blocking in the background or blocking).
// It provides the release information for the latest update of each dependency.
type UpdateNotification struct {
	clients      *shared.ClientFactory
	hoursToWait  float64
	wg           sync.WaitGroup
	dependencies []Dependency
}

type Dependency interface {
	CheckForUpdate(ctx context.Context) error
	PrintUpdateNotification(cmd *cobra.Command) (bool, error)
	HasUpdate() (bool, error)
	InstallUpdate(ctx context.Context) error
}

// defaultHoursWaitInterval is the default number of hours to wait between checking for updates.
const defaultHoursToWait = 24

// retryHoursToWait is the number of hours to wait before checking again after the Slack CLI check fails.
// It keeps commands run while offline from each waiting on a check.
const retryHoursToWait = 1

// envSkipUpdate is the environment variable that skips checking for updates.
const envSkipUpdate = "SLACK_SKIP_UPDATE"

// New is the constructor for UpdateNotification that sets good defaults.
func New(clients *shared.ClientFactory, cliVersion string) *UpdateNotification {
	return &UpdateNotification{
		clients:     clients,
		hoursToWait: defaultHoursToWait,
		dependencies: []Dependency{
			NewCLIDependency(clients, cliVersion),
			NewSDKDependency(clients),
		},
	}
}

// Dependencies returns the dependencies to check for updates.
func (u *UpdateNotification) Dependencies() []Dependency {
	return u.dependencies
}

// PrintAndPromptUpdates displays an update message and prompts the user if they want to update, if applicable
// Invoked from root command's post-run method and the upgrade command. If an error occurs, we return it so it is raised to the user.
func (u *UpdateNotification) PrintAndPromptUpdates(cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Wait for the background check, if one was started
	u.wg.Wait()

	for _, dependency := range u.Dependencies() {
		hasUpdate, err := dependency.HasUpdate()
		if err != nil {
			return slackerror.Wrap(err, "An error occurred while fetching a dependency")
		}

		if hasUpdate {
			shouldSelfUpdate, err := dependency.PrintUpdateNotification(cmd)
			if err != nil {
				return err
			}
			if shouldSelfUpdate {
				if err := dependency.InstallUpdate(ctx); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CheckForUpdateInBackground starts a Go routine that checks for an update every hoursSinceLastUpdateCheck (e.g. 24 hours).
func (u *UpdateNotification) CheckForUpdateInBackground(ctx context.Context) {
	u.wg.Go(func() {
		_ = u.CheckForUpdate(ctx, false)
	})
}

// HasUpdate returns true if any dependency has an update available.
func (u *UpdateNotification) HasUpdate() bool {
	for _, dependency := range u.Dependencies() {
		if hasUpdate, _ := dependency.HasUpdate(); hasUpdate {
			return true
		}
	}
	return false
}

// CheckForUpdate has each dependency perform a self-check for available updates.
// This is the synchronous version of CheckForUpdateInBackground
//
// A forced check returns the first error from a dependency, so the upgrade command
// doesn't report the latest version when it couldn't check. Other checks only log
// errors, so a command that runs offline doesn't fail.
func (u *UpdateNotification) CheckForUpdate(ctx context.Context, shouldForceUpdateCheck bool) error {
	if !u.shouldCheckForUpdate(ctx, shouldForceUpdateCheck) {
		return nil
	}

	// Each dependency self-checks for an update
	var checkErr error
	var cliCheckFailed bool
	for _, v := range u.Dependencies() {
		err := v.CheckForUpdate(ctx)
		if err != nil {
			u.clients.IO.PrintDebug(ctx, "CheckForUpdate for %s failed: %v", reflect.TypeOf(v), err)
			if checkErr == nil {
				checkErr = err
			}
			if _, ok := v.(*CLIDependency); ok {
				cliCheckFailed = true
			}
		}
	}

	// A failed check of the Slack CLI is retried sooner than usual, such as after going offline.
	// Errors from the check-update hook are specific to a project, so those wait as usual.
	if cliCheckFailed {
		_, _ = u.clients.Config.SystemConfig.SetLastUpdateCheckFailedAt(ctx, time.Now())
	} else {
		_, _ = u.clients.Config.SystemConfig.SetLastUpdateCheckedAt(ctx, time.Now())
		_, _ = u.clients.Config.SystemConfig.SetLastUpdateCheckFailedAt(ctx, time.Time{})
	}

	if shouldForceUpdateCheck {
		return checkErr
	}
	return nil
}

// shouldCheckForUpdate ensures that update checks are only run on supported environments.
func (u *UpdateNotification) shouldCheckForUpdate(ctx context.Context, shouldForceUpdateCheck bool) bool {
	if shouldForceUpdateCheck {
		return true
	}

	if !u.isEnabled() {
		return false
	}

	if u.isLastUpdateCheckFailedAtLessThan(ctx, retryHoursToWait) {
		return false
	}

	if !u.isLastUpdateCheckedAtGreaterThan(ctx, u.hoursToWait) {
		return false
	}

	return true
}

// isEnabled returns false when update checks are skipped by the environment, a flag, or the command.
func (u *UpdateNotification) isEnabled() bool {
	// The upgrade command runs its own check. This matches the parsed command because
	// os.Args can't tell "slack --verbose upgrade" apart from other commands.
	if u.clients.Config.Command == "upgrade" {
		return false
	}
	return os.Getenv(envSkipUpdate) == "" &&
		!u.clients.Config.SkipUpdateFlag &&
		u.clients.Config.TokenFlag == "" &&
		!isCI() &&
		!u.isIgnoredCommand()
}

// isCI returns true when running on a continuous integration service.
// Borrowed from https://github.com/watson/ci-info/blob/HEAD/index.js
func isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

// isIgnoredCommand returns true when the process is in the list of commands.
func (u *UpdateNotification) isIgnoredCommand() bool {
	// Commands that skip update checks only when used alone or with flags
	// "manifest" is included because it's an alias that runs "manifest info"
	ignoredCommands := []string{"--skill", "_fingerprint", "manifest", "manifest info", "skill", "skills", "version"}
	// Commands that skip update checks regardless of any arguments
	// "api" is a leaf command that accepts positional args, not subcommands
	ignoredPrefixes := []string{"api"}
	if len(os.Args) < 2 {
		return false
	}
	commandStr := strings.Join(os.Args[1:], " ")
	for _, ignored := range ignoredCommands {
		if commandStr == ignored {
			return true
		}
		// Match commands with additional flags (e.g. "manifest info --source local")
		// but not subcommands (e.g. "manifest validate" should not match "manifest")
		if strings.HasPrefix(commandStr, ignored+" ") {
			rest := commandStr[len(ignored)+1:]
			if strings.HasPrefix(rest, "-") {
				return true
			}
			// Allow prefix match for multi-word commands (e.g. "manifest info --flag")
			if strings.Contains(ignored, " ") {
				return true
			}
		}
	}
	for _, prefix := range ignoredPrefixes {
		if commandStr == prefix || strings.HasPrefix(commandStr, prefix+" ") {
			return true
		}
	}
	return false
}

// isLastUpdateCheckFailedAtLessThan returns true when the last failed check of the Slack CLI happened
// less than the hours parameter ago.
func (u *UpdateNotification) isLastUpdateCheckFailedAtLessThan(ctx context.Context, hours float64) bool {
	lastUpdateCheckFailedAt, err := u.clients.Config.SystemConfig.GetLastUpdateCheckFailedAt(ctx)
	if err != nil || lastUpdateCheckFailedAt.IsZero() {
		return false
	}

	return time.Since(lastUpdateCheckFailedAt).Hours() < hours
}

// isLastUpdateCheckedAtGreaterThan returns true when the time since the last update check is greater
// than the hours parameter.
func (u *UpdateNotification) isLastUpdateCheckedAtGreaterThan(ctx context.Context, hours float64) bool {
	lastUpdateCheckedAt, err := u.clients.Config.SystemConfig.GetLastUpdateCheckedAt(ctx)
	if err != nil {
		return false
	}

	if time.Since(lastUpdateCheckedAt).Hours() > hours {
		return true
	}

	return false
}

// newHTTPClient returns an http.Client for checking the latest release on github.com.
func newHTTPClient() (*http.Client, error) {
	return api.NewHTTPClient(api.HTTPClientOptions{TotalTimeOut: 60 * time.Second}), nil
}
