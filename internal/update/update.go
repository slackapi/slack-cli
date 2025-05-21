// Copyright 2022-2025 Salesforce, Inc.
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
	"sync"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
)

// UpdateNotification checks for an update (non-blocking in the background or blocking).
// It provides the release information for the latest update of each dependency.
type UpdateNotification struct {
	clients         *shared.ClientFactory
	enabled         bool
	envDisabled     string
	hoursToWait     float64
	checkUpdateChan chan []Dependency
	dependencies    []Dependency
}

type Dependency interface {
	CheckForUpdate(ctx context.Context) error
	PrintUpdateNotification(cmd *cobra.Command) (bool, error)
	HasUpdate() (bool, error)
	InstallUpdate(ctx context.Context) error
}

// defaultHoursWaitInterval is the default number of hours to wait between checking for updates.
const defaultHoursToWait = 24

// We use the singleton pattern with UpdateNotification
var updateNotification *UpdateNotification

// Object that performs an action once. Used as a thread-safe way to implement the singleton pattern.
var once sync.Once

// New is the constructor for UpdateNotification that sets good defaults.
func New(clients *shared.ClientFactory, cliVersion string, envDisabled string) *UpdateNotification {
	var enabled = true
	if clients.Config.SkipUpdateFlag || clients.Config.TokenFlag != "" {
		enabled = false
	}
	// The native sync library provides a thread safe way to implement the Singleton pattern. We use this to ensure that only one instance of the UpdateNotification is running at a time.
	// https://progolang.com/how-to-implement-singleton-pattern-in-go/ and https://medium.com/golang-issue/how-singleton-pattern-works-with-golang-2fdd61cd5a7f
	once.Do(func() {
		updateNotification = &UpdateNotification{
			clients:     clients,
			enabled:     enabled,
			envDisabled: envDisabled,
			hoursToWait: defaultHoursToWait,
			dependencies: []Dependency{
				NewCLIDependency(clients, cliVersion),
				NewSDKDependency(clients),
			},
		}
	})

	return updateNotification
}

// SetEnabled enables or disables checking for the latest version. Default is true.
func (u *UpdateNotification) SetEnabled(b bool) {
	u.enabled = b
}

// Enabled returns true if checking for the latest version is enabled.
func (u *UpdateNotification) Enabled() bool {
	return u.enabled
}

// SetEnv sets an environment variable to disable update notifications.
func (u *UpdateNotification) SetEnv(s string) {
	u.envDisabled = s
}

// Env returns the environment variable to disable update notifications.
func (u *UpdateNotification) Env() string {
	return u.envDisabled
}

// SetHours sets the number of hours to wait between checking for updates.
func (u *UpdateNotification) SetHours(h float64) {
	u.hoursToWait = h
}

// Hours returns the number of hours to wait between checking for updates.
func (u *UpdateNotification) Hours() float64 {
	return u.hoursToWait
}

// Dependencies returns the dependencies to check for updates.
func (u *UpdateNotification) Dependencies() []Dependency {
	return u.dependencies
}

// PrintUpdates displays an update message after the command runs and prompts the user if they want to update, if applicable
// Invoked from root command's post-run method. If an error occurs, we return it so it is raised to the user.
func (u *UpdateNotification) PrintAndPromptUpdates(cmd *cobra.Command, cliVersion string) error {
	ctx := cmd.Context()

	if updateNotification.WaitForCheckForUpdateInBackground() {
		for _, dependency := range updateNotification.Dependencies() {
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
	}
	return nil
}

// CheckForUpdateInBackground starts a Go routine that checks for an update every hoursSinceLastUpdateCheck (e.g. 24 hours).
// The shouldForceUpdateCheck param can be used to force the update check, regardless of the time since the last check.
func (u *UpdateNotification) CheckForUpdateInBackground(ctx context.Context, shouldForceUpdateCheck bool) {
	// This function is run before every command as its used in the PersistentPreRun in root.go
	// In that case, we must clear the existing updateNotification instance out of the channel
	if u.checkUpdateChan != nil {
		<-u.checkUpdateChan
	}

	u.checkUpdateChan = make(chan []Dependency)
	go func() {
		_ = u.CheckForUpdate(ctx, shouldForceUpdateCheck)
		u.checkUpdateChan <- u.Dependencies()
	}()
}

// WaitForCheckForUpdateInBackground must be called after CheckForUpdateInBackground and will block until the go routine finishes.
// Returns boolean that indicates if updates have been retrieved.
func (u *UpdateNotification) WaitForCheckForUpdateInBackground() bool {
	dependencies := <-u.checkUpdateChan
	return len(dependencies) > 0
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
func (u *UpdateNotification) CheckForUpdate(ctx context.Context, shouldForceUpdateCheck bool) error {
	if !u.shouldCheckForUpdate(ctx, shouldForceUpdateCheck) {
		return nil
	}

	// Each dependency self-checks for an update
	for _, v := range u.Dependencies() {
		err := v.CheckForUpdate(ctx)
		if err != nil {
			u.clients.IO.PrintDebug(ctx, "CheckForUpdate for %s failed: %v", reflect.TypeOf(v), err)
		}
	}

	_, _ = u.clients.Config.SystemConfig.SetLastUpdateCheckedAt(ctx, time.Now())

	return nil
}

// shouldCheckForUpdate ensures that update checks are only run on supported environments.
func (u *UpdateNotification) shouldCheckForUpdate(ctx context.Context, shouldForceUpdateCheck bool) bool {
	if shouldForceUpdateCheck {
		return true
	}

	if (u.envDisabled != "" && os.Getenv(u.envDisabled) != "") || !u.enabled || u.isCI() || u.isIgnoredCommand() {
		return false
	}

	if !u.isLastUpdateCheckedAtGreaterThan(ctx, u.hoursToWait) {
		return false
	}

	return true
}

// isCI returns true when running on a continuous integration service.
// Borrowed from https://github.com/watson/ci-info/blob/HEAD/index.js
func (u *UpdateNotification) isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

// isIgnoredCommand returns true when the process is in the list of commands.
func (u *UpdateNotification) isIgnoredCommand() bool {
	ignoredCommands := []string{"version"}
	osStr := os.Args[0:]
	if len(osStr) < 2 {
		return false
	}
	commandName := osStr[1]
	return goutils.Contains(ignoredCommands, commandName, true)
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
