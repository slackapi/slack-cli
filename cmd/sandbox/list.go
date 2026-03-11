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

package sandbox

import (
	"fmt"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type listFlags struct {
	filter string
}

var listCmdFlags listFlags

func NewListCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List developer sandboxes",
		Long: strings.Join([]string{
			"List details about your developer sandboxes.",
			"",
			"The listed developer sandboxes belong to a developer program account",
			"that matches the email address of the authenticated user.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox list", Meaning: "List developer sandboxes"},
			{Command: "sandbox list --filter active", Meaning: "List active sandboxes only"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&listCmdFlags.filter, "filter", "", "Filter by status: active, archived")

	return cmd
}

func runListCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	auth, err := getSandboxAuth(ctx, clients)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "")
	err = printSandboxes(cmd, clients, auth.Token, auth)
	if err != nil {
		return err
	}

	return nil
}

func printSandboxes(cmd *cobra.Command, clients *shared.ClientFactory, token string, auth *types.SlackAuth) error {
	ctx := cmd.Context()

	sandboxes, err := clients.API().ListSandboxes(ctx, token, listCmdFlags.filter)
	if err != nil {
		return err
	}

	email := ""
	if auth != nil && auth.UserID != "" {
		if userInfo, err := clients.API().UsersInfo(ctx, token, auth.UserID); err == nil && userInfo.Profile.Email != "" {
			email = userInfo.Profile.Email
		}
	}

	section := style.TextSection{
		Emoji: "beach_with_umbrella",
		Text:  " Developer Sandboxes",
	}

	// Some users' logins may not include the scope needed to access the email address from the `users.info` method, so it may not be set
	// Learn more: https://docs.slack.dev/reference/methods/users.info/#email-addresses
	if email != "" {
		section.Secondary = []string{fmt.Sprintf("Owned by Slack developer account %s", email)}
	}

	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(section))

	if len(sandboxes) == 0 {
		clients.IO.PrintInfo(ctx, false, "   %s\n", style.Secondary("No sandboxes found"))
		return nil
	}

	timeFormat := "2006-01-02" // We only support the granularity of the day for now, rather than a more precise datetime
	for _, s := range sandboxes {
		clients.IO.PrintInfo(ctx, false, "  %s (%s)", style.Bold(s.SandboxName), s.SandboxTeamID)

		if s.SandboxDomain != "" {
			clients.IO.PrintInfo(ctx, false, "    %s", style.Secondary(fmt.Sprintf("URL: https://%s.slack.com", s.SandboxDomain)))
		}

		if s.Status != "" {
			status := style.Secondary(fmt.Sprintf("Status: %s", strings.ToTitle(s.Status)))
			if strings.EqualFold(s.Status, "archived") {
				clients.IO.PrintInfo(ctx, false, "    %s", style.Styler().Red(status))
			} else {
				clients.IO.PrintInfo(ctx, false, "    %s", style.Styler().Green(status))
			}
		}

		if s.DateCreated > 0 {
			clients.IO.PrintInfo(ctx, false, "    %s", style.Secondary(fmt.Sprintf("Created: %s", time.Unix(s.DateCreated, 0).Format(timeFormat))))
		}

		if s.DateArchived > 0 {
			archivedTime := time.Unix(s.DateArchived, 0).In(time.Local)
			now := time.Now()
			archivedDate := time.Date(archivedTime.Year(), archivedTime.Month(), archivedTime.Day(), 0, 0, 0, 0, time.Local)
			todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
			label := "Expires:"
			if archivedDate.Before(todayDate) {
				label = "Archived:"
			}
			clients.IO.PrintInfo(ctx, false, "    %s", style.Secondary(fmt.Sprintf("%s %s", label, archivedTime.Format(timeFormat))))
		}

		clients.IO.PrintInfo(ctx, false, "")
	}

	clients.IO.PrintInfo(ctx, false, "Learn more at %s", style.Secondary("https://docs.slack.dev/tools/developer-sandboxes"))

	return nil
}
