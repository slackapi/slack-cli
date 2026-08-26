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

package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// requestTimeFormat displays the moment a request changed
const requestTimeFormat = "2006-01-02 15:04:05 Z07:00"

// Handle to a function used for testing
var requestAppSelectPromptFunc = prompts.AppSelectPrompt

// Handle to a function used for testing
var requestTeamSelectPromptFunc = prompts.PromptTeamSlackAuth

// Flags
type requestCmdFlags struct {
	workspaceIDs []string
}

var requestFlags requestCmdFlags

// NewRequestCommand returns a new Cobra command
func NewRequestCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "request [flags]",
		Aliases: []string{"requests"},
		Short:   "Check approval requests to install the app",
		Long: strings.Join([]string{
			"Check the status of your most recent request to have the app approved for",
			"install.",
			"",
			"Requests are searched on the team of the authenticated account. An account of",
			"a workspace that belongs to an organization also searches that organization,",
			"while an account of an organization searches the organization alone.",
			"",
			"Other workspaces of an organization can be searched with the --workspace-ids",
			"flag.",
			"",
			"Searches are made with the credentials of an authenticated account chosen",
			"with the --team flag or a prompt.",
			"",
			"Apps saved to a project are chosen with a prompt, but any app can be checked",
			"by app ID with the --app flag, which does not require a project.",
		}, "\n"),
		Hidden: true,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app request", Meaning: "Check requests to install an app"},
			{Command: "app request --app A0123456789", Meaning: "Check requests for an app outside a project"},
			{Command: "app request --workspace-ids T0123456789,T9876543210", Meaning: "Check requests on certain workspaces of an organization"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !clients.Config.WithExperimentOn(experiment.AppApprovalStatus) {
				return slackerror.New(slackerror.ErrExperimentRequired).
					WithRemediation("Enable the %s experiment with %s",
						style.Highlight(string(experiment.AppApprovalStatus)),
						style.CommandText("--experiment app-approval-status"),
					)
			}
			clients.Config.SetFlags(cmd)
			// An app named by ID is checked without the apps of a project
			if types.IsAppID(clients.Config.AppFlag) {
				return nil
			}
			// Verify command is run in a project directory
			if err := cmdutil.IsValidProjectDirectory(clients); err != nil {
				invalid := slackerror.ToSlackError(err)
				return invalid.WithRemediation("%s\n\nApps of other projects can be checked with %s",
					invalid.Remediation,
					style.CommandText("--app A0123456789"),
				)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRequestCommand(cmd, clients)
		},
	}

	cmd.Flags().StringSliceVar(&requestFlags.workspaceIDs, "workspace-ids", nil, "also check these workspaces of an organization,\nwith a maximum of 50 workspaces")

	return cmd
}

// runRequestCommand will execute the request command
func runRequestCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.app.request")
	defer span.Finish()

	appID, auth, err := requestAppSelection(ctx, clients)
	if err != nil {
		return err
	}

	result, err := clients.API().ListAppApprovalRequests(ctx, auth.Token, appID, requestFlags.workspaceIDs)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "lock",
		Text:      "App Install Approval Requests",
		Secondary: FormatRequestSuccess(appID, requestTeamNames(auth), result.Requests),
	}))
	return nil
}

// requestTeamNames collects the names of searched teams that are known.
//
// Requests are returned with team IDs alone, so only the team of the
// authenticated account is named. Other teams of an organization are not
// looked up to avoid another API call.
func requestTeamNames(auth types.SlackAuth) map[string]string {
	if auth.TeamID == "" || auth.TeamDomain == "" {
		return nil
	}
	return map[string]string{auth.TeamID: auth.TeamDomain}
}

// requestAppSelection decides the app to check and the account to search with.
//
// An app named by ID with the app flag is checked without a project so that
// apps missing from a project can be checked too. The team of that app is
// gathered from the authenticated accounts instead of the project apps.
func requestAppSelection(ctx context.Context, clients *shared.ClientFactory) (appID string, auth types.SlackAuth, err error) {
	if types.IsAppID(clients.Config.AppFlag) {
		selected, err := requestTeamSelectPromptFunc(ctx, clients, "Select an account to search with", nil)
		if err != nil {
			return "", types.SlackAuth{}, err
		}
		if selected == nil || selected.Token == "" {
			return "", types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound)
		}
		return clients.Config.AppFlag, *selected, nil
	}
	selection, err := requestAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return "", types.SlackAuth{}, err
	}
	if selection.App.AppID == "" {
		return "", types.SlackAuth{}, slackerror.New(slackerror.ErrAppNotFound)
	}
	return selection.App.AppID, selection.Auth, nil
}

// FormatRequestSuccess formats the install request of each team for an app.
// Teams found in teamNames are titled by name while others are titled by ID.
func FormatRequestSuccess(appID string, teamNames map[string]string, requests []api.AppsApprovalsRequest) (secondaryText []string) {
	sorted := slices.SortedFunc(slices.Values(requests), func(a api.AppsApprovalsRequest, b api.AppsApprovalsRequest) int {
		return strings.Compare(a.TeamID, b.TeamID)
	})
	field := func(label string, value string) string {
		return fmt.Sprintf(style.Indent(style.Secondary("%-13s %s")), label+":", value)
	}
	if appID != "" {
		secondaryText = append(secondaryText, fmt.Sprintf(style.Bold("%-13s %s"), "App ID:", appID))
	}
	// Requests are gathered apart from the app to know when none were made
	requestsText := []string{}
	for _, request := range sorted {
		requestsText = append(requestsText, fmt.Sprintf(style.Bold("%s:"), formatRequestTeam(teamNames, request.TeamID)))
		requestsText = append(requestsText, field("Request ID", request.ID))
		requestsText = append(requestsText, field("Status", formatRequestStatus(request.Status)))
		requestsText = append(requestsText, field("Requested", formatRequestTime(request.DateCreated)))
		if request.DateResolved > 0 {
			requestsText = append(requestsText, field("Resolved", formatRequestTime(request.DateResolved)))
		}
		if request.CancelledBy != "" {
			requestsText = append(requestsText, field("Cancelled by", formatRequestCancelledBy(request.CancelledBy)))
		}
		if request.CanSelfApprove {
			requestsText = append(requestsText, style.Indent(style.Secondary("You can install this app without approval. Please cancel the request.")))
		}
	}
	if len(requestsText) <= 0 {
		requestsText = append(requestsText, "You have not requested to install this app")
	}
	secondaryText = append(secondaryText, requestsText...)
	return
}

// formatRequestTeam titles a team by name and ID when the name is known
func formatRequestTeam(teamNames map[string]string, teamID string) string {
	if name, ok := teamNames[teamID]; ok {
		return fmt.Sprintf("%s (%s)", name, teamID)
	}
	return teamID
}

// formatRequestTime displays a Unix timestamp in the local timezone
func formatRequestTime(timestamp int64) string {
	if timestamp <= 0 {
		return "unknown"
	}
	return time.Unix(timestamp, 0).Format(requestTimeFormat)
}

// formatRequestCancelledBy names the kind of actor that cancelled a request.
// Every returned request was made by the authenticated account, so a request
// cancelled by a user was withdrawn by that same account.
func formatRequestCancelledBy(actor api.AppsApprovalsRequestCancelledBy) string {
	switch actor {
	case api.AppsApprovalsRequestCancelledByAdmin:
		return "an admin"
	case api.AppsApprovalsRequestCancelledBySystem:
		return "the system"
	case api.AppsApprovalsRequestCancelledByUser:
		return "you"
	default:
		return string(actor)
	}
}

// formatRequestStatus styles a status by how much attention it deserves
func formatRequestStatus(status api.AppsApprovalsRequestStatus) string {
	switch status {
	case api.AppsApprovalsRequestStatusApproved:
		return style.Green(string(status))
	case api.AppsApprovalsRequestStatusCancelled:
		return style.Secondary(string(status))
	case api.AppsApprovalsRequestStatusDenied:
		return style.Red(string(status))
	case api.AppsApprovalsRequestStatusPending:
		return style.Yellow(string(status))
	default:
		return string(status)
	}
}
