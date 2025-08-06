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
	"encoding/json"
	"fmt"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// SDKInstallUpdateResponse contains information following
// the attempted updating of dependencies used within a
// given project
type SDKInstallUpdateResponse struct {
	Name    string                      `json:"name"`
	Updates []SDKInstallUpdateComponent `json:"updates"`
	Error   struct {
		Message string `json:"message,omitempty"`
	} `json:"error"`
}

// SDKInstallUpdateComponent contains the information
// about a given SDK dependency's install update
type SDKInstallUpdateComponent struct {
	Name             string `json:"name"`
	PreviousVersion  string `json:"previous"`
	InstalledVersion string `json:"installed"`
	Error            struct {
		Message string `json:"message,omitempty"`
	} `json:"error"`
}

// SDKDependency contains the latest release
// information for dependencies used by
// a given project
type SDKDependency struct {
	clients     *shared.ClientFactory
	releaseInfo SDKReleaseInfo
}

// SDKReleaseInfo contains the information about updates
// available for the dependencies used within a
// given project
type SDKReleaseInfo struct {
	Name     string `json:"name"`
	Breaking bool
	Update   bool
	Releases []SDKReleaseComponent `json:"releases"`
	Message  string                `json:"message"`
	URL      string                `json:"url"`
	Error    struct {
		Message string `json:"message"`
	} `json:"error"`
}

// SDKReleaseComponent contains the information about
// a given SDK dependency's update status
type SDKReleaseComponent struct {
	Name     string `json:"name"`
	Current  string `json:"current"`
	Latest   string `json:"latest"`
	Breaking bool   `json:"breaking"`
	Update   bool   `json:"update"`
	Message  string `json:"message"`
	URL      string `json:"url"`
	Error    struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewSDKDependency creates and returns a new instance of SDKDependency
func NewSDKDependency(clients *shared.ClientFactory) *SDKDependency {
	sdkDependency := &SDKDependency{
		clients: clients,
	}
	return sdkDependency
}

// CheckForUpdate executes the `check-update` hook, if available, and
// sets the release information to the SDKDependency instance
func (c *SDKDependency) CheckForUpdate(ctx context.Context) error {
	var err error
	c.releaseInfo, err = CheckUpdateHook(ctx, c.clients)
	c.releaseInfo.Update = false
	c.releaseInfo.Breaking = false
	if err != nil {
		if slackerror.ToSlackError(err).Code == slackerror.ErrSDKHookNotFound {
			c.clients.IO.PrintDebug(ctx, `"check-update" hook value not found in %s when checking for SDK updates`, config.GetProjectHooksJSONFilePath())
		} else {
			return err
		}
	}
	for _, dependency := range c.releaseInfo.Releases {
		if dependency.Update {
			c.releaseInfo.Update = true
		}
		if dependency.Breaking {
			c.releaseInfo.Breaking = true
		}
	}
	return nil
}

// CheckUpdateHook returns the response from the check update hook
func CheckUpdateHook(ctx context.Context, clients *shared.ClientFactory) (SDKReleaseInfo, error) {
	var checkUpdate SDKReleaseInfo
	if !clients.SDKConfig.Hooks.CheckUpdate.IsAvailable() {
		return SDKReleaseInfo{}, slackerror.New(slackerror.ErrSDKHookNotFound).
			WithMessage("The `check-update` hook was not found").
			WithRemediation("Debug responses from the Slack hooks file (%s)", config.GetProjectHooksJSONFilePath())
	}
	var hookExecOpts = hooks.HookExecOpts{
		Hook: clients.SDKConfig.Hooks.CheckUpdate,
	}
	checkUpdateResponse, err := clients.HookExecutor.Execute(ctx, hookExecOpts)
	if err != nil {
		return SDKReleaseInfo{}, err
	}
	err = json.Unmarshal([]byte(checkUpdateResponse), &checkUpdate)
	if err != nil {
		return SDKReleaseInfo{}, slackerror.New(slackerror.ErrUnableToParseJSON).
			WithMessage("Failed to parse response from check-update hook").
			WithRootCause(err)
	}
	return checkUpdate, nil
}

// HasUpdate returns true if the SDK has an update available
func (c *SDKDependency) HasUpdate() (bool, error) {
	var err error
	if c.releaseInfo.Error.Message != "" {
		err = slackerror.New(c.releaseInfo.Error.Message)
	}
	return c.releaseInfo.Update, err
}

// InstallUpdate executes the `install-update` hook, if available, and
// prints information about the execution of available updates
func (c *SDKDependency) InstallUpdate(ctx context.Context) error {
	if c.clients.SDKConfig.Hooks.InstallUpdate.IsAvailable() {
		var hookExecOpts = hooks.HookExecOpts{
			Hook: c.clients.SDKConfig.Hooks.InstallUpdate,
			//Name: "install-update",
		}

		fmt.Print(style.SectionSecondaryf("Starting the auto-update..."))

		installUpdateJSON, err := c.clients.HookExecutor.Execute(ctx, hookExecOpts)
		if err != nil {
			return err
		}

		fmt.Print(style.SectionSecondaryf("Wrapping up updates..."))

		var installUpdateInfo SDKInstallUpdateResponse
		err = json.Unmarshal([]byte(installUpdateJSON), &installUpdateInfo)
		if err != nil {
			return err
		}

		printInstallUpdateResponse(installUpdateInfo)
	} else {
		c.clients.IO.PrintDebug(ctx, `"install-update" hook value not found in %s when attempting to update SDK`, config.GetProjectHooksJSONFilePath())
	}

	return nil
}

// PrintUpdateNotification prints out the update message returned from the
// SDK, including formatting and language that indicates if a breaking change
// is included or an error has occurred
func (c *SDKDependency) PrintUpdateNotification(cmd *cobra.Command) (bool, error) {
	ctx := cmd.Context()

	if c.releaseInfo.Update {
		// Standard "update is available" message
		cmd.Printf(
			style.Indent("\n%s\n\n"),
			style.Bold(
				fmt.Sprintf("%s An update from %s is available:",
					style.Emoji("hammer_and_wrench"),
					c.releaseInfo.Name,
				)),
		)

		// The update(s) includes a breaking change
		if c.releaseInfo.Breaking {
			cmd.Printf(
				style.Indent("%s\n\n"),
				style.Warning("Warning: this update contains a breaking change!"),
			)
		}
	}

	// Print SDK update message
	if c.releaseInfo.Message != "" {
		cmd.Printf(
			style.Indent("%s\n\n"),
			c.releaseInfo.Message,
		)
	}

	// Print out each individual dependency update detail
	for _, dependency := range c.releaseInfo.Releases {
		// The dependency encountered an error
		if dependency.Error.Message != "" {
			cmd.Printf(
				style.Indent(" %s %s\n      %s\n\n"),
				style.Error("✖"),
				style.Bold(dependency.Name),
				style.Error(dependency.Error.Message),
			)
		} else {
			// Dependency has breaking change
			if dependency.Breaking {
				cmd.Printf(
					style.Indent(" %s %s\n      %s → %s\n\n"),
					"›",
					style.Bold(dependency.Name),
					style.Secondary(dependency.Current),
					style.Warning(fmt.Sprintf("%s (breaking)", dependency.Latest)),
				)
			} else {
				// Latest version is unavailable
				if dependency.Latest == "" {

					if dependency.Current != "" {
						cmd.Printf(
							style.Indent(" %s %s\n      %s\n\n"),
							"›",
							style.Bold(dependency.Name),
							style.Secondary(fmt.Sprintf("Current version: %s", dependency.Current)),
						)
					} else {
						cmd.Printf(
							style.Indent(" %s %s\n\n"),
							"›",
							style.Bold(dependency.Name),
						)
					}
				} else {

					if dependency.Current == "" {
						cmd.Printf(
							style.Indent(" %s %s\n      %s\n\n"),
							"›",
							style.Bold(dependency.Name),
							style.Secondary(fmt.Sprintf("Latest version: %s", dependency.Latest)),
						)
					} else {
						if dependency.Update {
							cmd.Printf(
								style.Indent(" %s %s\n      %s → %s\n\n"),
								"›",
								style.Bold(dependency.Name),
								style.Secondary(dependency.Current),
								style.CommandText(dependency.Latest),
							)
						}
					}
				}
			}
		}

		// A message accompanies the dependency update
		if dependency.Message != "" {
			cmd.Printf(
				style.Indent("   %s\n\n"),
				style.Secondary(dependency.Message),
			)
		}

		// A URL accompanies the dependency update
		if dependency.URL != "" {
			cmd.Printf(
				style.Indent("   %s\n   %s\n\n"),
				style.Secondary("Learn more at:"),
				style.Indent(style.Bold(style.Secondary(dependency.URL))),
			)
		}
	}

	// A reference URL for the update has been provided
	if c.releaseInfo.URL != "" {
		cmd.Printf(
			style.Indent("%s\n%s\n\n"),
			"For more information about this update, visit:",
			style.Indent(style.CommandText(c.releaseInfo.URL)),
		)
	}

	// The update(s) includes an error
	if c.releaseInfo.Error.Message != "" {
		c.clients.IO.PrintError(ctx,
			style.Indent("%s\n%s\n"),
			style.Error("Error:"),
			style.Indent(c.releaseInfo.Error.Message),
		)
	}

	// If `install-update` hook available, prompt to auto-update
	if c.clients.SDKConfig.Hooks.InstallUpdate.IsAvailable() {
		// Check for sdk flag from upgrade command
		if cmd.Name() == "upgrade" && cmd.Flags().Changed("sdk") {
			sdk, _ := cmd.Flags().GetBool("sdk")
			if sdk {
				return true, nil
			}
		}

		autoUpdatePrompt := fmt.Sprintf("%sDo you want to auto-update to the latest versions now?", style.Emoji("rocket"))
		return c.clients.IO.ConfirmPrompt(ctx, autoUpdatePrompt, false)
	}

	return false, nil
}

// printInstallUpdateResponse prints out the dependency installation results
// returned from the `install-update` hook, featuring formatting and language that indicates
// if each update was successful or if an error occurred
func printInstallUpdateResponse(updateInfo SDKInstallUpdateResponse) {
	// If individual update information has been returned
	if len(updateInfo.Updates) > 0 {
		// Standard "these dependencies were involved" message
		fmt.Printf(
			style.Indent("\n%s\n\n"),
			style.Bold(
				fmt.Sprintf("%sThe following dependencies have been updated:",
					style.Emoji("sparkles"),
				)),
		)

		// Print out each individual dependency update detail
		for _, update := range updateInfo.Updates {

			// The installation encountered an error
			if update.Error.Message != "" {
				fmt.Printf(
					style.Indent(" %s %s\n      %s\n\n"),
					style.Error("✖"),
					style.Bold(update.Name),
					style.Error(update.Error.Message),
				)
			} else {
				if update.PreviousVersion != update.InstalledVersion {
					fmt.Printf(
						style.Indent(" %s %s\n      %s → %s\n\n"),
						style.Styler().Green("✔"),
						style.Bold(update.Name),
						style.Secondary(update.PreviousVersion),
						style.Styler().Green(update.InstalledVersion),
					)
				}
			}
		}
	}

	// The installation(s) encountered an error
	if updateInfo.Error.Message != "" {
		fmt.Printf(
			style.Indent("%s\n%s\n\n"),
			style.Error("Error:"),
			style.Indent(updateInfo.Error.Message),
		)
	}

	fmt.Printf(
		style.Indent("%s\n\n"),
		style.Secondary("Please review any file changes before continuing development."),
	)
}
