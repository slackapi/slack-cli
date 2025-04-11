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
	"fmt"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

const metaDataUrl = "https://api.slack.com/slackcli/metadata.json"

// CLIDependency contains information about the
// current version and the latest CLI release version
// from Slack CLI metadata
type CLIDependency struct {
	clients      *shared.ClientFactory
	releaseError error
	version      string
	releaseInfo  *LatestCLIRelease
}

// NewCLIDependency creates and returns a new instance of CLIDependency
func NewCLIDependency(clients *shared.ClientFactory, version string) *CLIDependency {
	cliDependency := &CLIDependency{
		clients: clients,
		version: version,
	}
	return cliDependency
}

// CheckForUpdate retrieves and sets the LatestCLIRelease that describes
// the latest update available directly to the CLIDependency instance
// Synchronous version of CheckForUpdateInBackground
func (c *CLIDependency) CheckForUpdate(ctx context.Context) error {
	httpClient, err := newHttpClient()
	if err != nil {
		return err
	}

	metadata := Metadata{httpClient: httpClient}
	c.releaseInfo, err = metadata.CheckForUpdate(ctx, metaDataUrl, c.version)
	if err != nil {
		return err
	}

	return nil
}

func (c *CLIDependency) HasUpdate() (bool, error) {
	return c.releaseInfo != nil, c.releaseError
}

// PrintUpdateNotification notifies the user that a new version is available and provides upgrade instructions for Homebrew. Returns a bool representing whether the user wants the self-update to run
func (c *CLIDependency) PrintUpdateNotification(cmd *cobra.Command) (bool, error) {
	processName := cmdutil.GetProcessName()
	isHomebrew := IsHomebrew(processName)

	cmd.Printf(
		"\n%s\n   %s → %s\n\n%s\n   %s\n",
		style.Bold(fmt.Sprintf("%sA new version of the Slack CLI is available:", style.Emoji("seedling"))),
		style.Secondary(c.version),
		style.CommandText(c.releaseInfo.Version),
		"   You can read the release notes at:",
		style.CommandText("https://docs.slack.dev/changelog"),
	)

	if isHomebrew {
		cmd.Printf(
			"\n   To update with Homebrew, run: %s\n\n",
			style.CommandText(fmt.Sprintf("brew update && brew upgrade %s", processName)),
		)
	} else {
		cmd.Printf(
			"\n   To manually update, visit the download page:\n   %s\n\n",
			style.CommandText("https://tools.slack.dev/slack-cli"),
		)
		selfUpdatePrompt := fmt.Sprintf("%sDo you want to auto-update to the latest version now?", style.Emoji("rocket"))
		return c.clients.IO.ConfirmPrompt(cmd.Context(), selfUpdatePrompt, false)
	}

	return false, nil

	// TODO: Uncomment when open sourced to display the latest release URL that includes release notes
	// cmd.Printf(
	// 	"\n%s\n\n",
	// 	newRelease.URL,
	// )
}
