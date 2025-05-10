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

package upgrade

import (
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/internal/update"
	"github.com/spf13/cobra"
)

// checkForUpdatesFunc is a function pointer for tests to mock the checkForUpdates function
var checkForUpdatesFunc = checkForUpdates

const changelogURL = "https://docs.slack.dev/changelog"

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:     "upgrade",
		Aliases: []string{"update"},
		Short:   "Checks for available updates to the CLI or SDK",
		Long: strings.Join([]string{
			"Checks for available updates to the CLI or the SDKs of a project",
			"",
			"If there are any, then you will be prompted to upgrade",
			"",
			fmt.Sprintf(`The changelog can be found at {{LinkText "%s"}}`, changelogURL),
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "upgrade", Meaning: "Check for any available updates"},
			{Command: "upgrade --auto-approve", Meaning: "Check for updates and automatically upgrade without confirmation"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkForUpdatesFunc(clients, cmd, autoApprove)
		},
	}

	cmd.Flags().BoolVar(&autoApprove, "auto-approve", false, "automatically approve and install updates without prompting")

	return cmd
}

// checkForUpdates will check for CLI/SDK updates and print a message when no updates are available.
// When there are updates, the function will *not* print a message because the root command handles printing update notifications.
func checkForUpdates(clients *shared.ClientFactory, cmd *cobra.Command, autoApprove bool) error {
	ctx := cmd.Context()
	updateNotification := update.New(clients, version.Get(), "SLACK_SKIP_UPDATE")

	// TODO(@mbrooks) This update check is happening at the same time as the root command's `CheckForUpdateInBackground`.
	//                The difference between the two is that this update check is forced while the root command runs every 24 hours.
	//                If both find an update, only 1 notification is displayed.
	//                How can we improve this to avoid doing 2 update network requests/checks?
	//
	// Force an update check that is blocking and synchronous
	if err := updateNotification.CheckForUpdate(ctx, true); err != nil {
		return err
	}

	// Update notification messages are printed by the root command's persistent post-run (cmd/root.go).
	// So this command only needs to print a message when everything is up-to-date.
	if updateNotification.HasUpdate() {
		if autoApprove {
			// Automatically install updates without prompting when auto-approve flag is set
			if err := updateNotification.InstallUpdatesWithoutPrompt(cmd); err != nil {
				return err
			}
			return nil
		}
		return nil
	}

	if clients.SDKConfig.Hooks.CheckUpdate.IsAvailable() {
		cmd.Printf("%s You are using the latest Slack CLI and SDK versions\n", style.Styler().Green("✔").String())
	} else {
		cmd.Printf("%s You are using the latest Slack CLI version\n", style.Styler().Green("✔").String())
	}

	return nil
}
