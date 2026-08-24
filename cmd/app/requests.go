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
	"sort"
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

// requestsTeamsLimit is the most teams the API searches in a single call
const requestsTeamsLimit = 50

// requestsTimeFormat displays the moment a request changed
const requestsTimeFormat = "2006-01-02 15:04:05 Z07:00"

// Handle to a function used for testing
var requestsAppSelectPromptFunc = prompts.AppSelectPrompt

// Handle to a function used for testing
var requestsTeamSelectPromptFunc = prompts.PromptTeamSlackAuth

// Flags

type requestsCmdFlags struct {
	teamIDs []string
}

var requestsFlags requestsCmdFlags

// NewRequestsCommand returns a new Cobra command
func NewRequestsCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "requests [flags]",
		Aliases: []string{"approval-requests", "approvals"},
		Short:   "Check requests to install the app",
		Long: strings.Join([]string{
			"Check the status of your most recent request to have the app approved for",
			"install.",
			"",
			"Requests are searched on the team of the authenticated account. An account of",
			"a workspace that belongs to an organization also searches that organization,",
			"while an account of an organization searches the organization alone.",
			"",
			"Other workspaces of an organization can be searched with the --team-ids flag.",
			"",
			"Apps saved to a project are chosen with a prompt, but any app can be checked",
			"by app ID with the --app flag, which does not require a project.",
		}, "\n"),
		Hidden: true,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app requests", Meaning: "Check requests to install an app"},
			{Command: "app requests --app A0123456789", Meaning: "Check requests for an app outside a project"},
			{Command: "app requests --team-ids T0123456789,T9876543210", Meaning: "Check requests on certain teams of an organization"},
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
			if len(requestsFlags.teamIDs) > requestsTeamsLimit {
				return slackerror.New(slackerror.ErrInvalidArguments).
					WithMessage("The %s flag accepts at most %d teams",
						style.CommandText("--team-ids"),
						requestsTeamsLimit,
					)
			}
			clients.Config.SetFlags(cmd)
			// An app named by ID is checked without the apps of a project
			if types.IsAppID(clients.Config.AppFlag) {
				return nil
			}
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRequestsCommand(cmd, clients)
		},
	}

	cmd.Flags().StringSliceVar(&requestsFlags.teamIDs, "team-ids", nil, "also check these teams of an organization,\nwith a maximum of 50 teams")

	return cmd
}

// runRequestsCommand will execute the requests command
func runRequestsCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.app.requests")
	defer span.Finish()

	appID, token, err := requestsAppSelection(ctx, clients)
	if err != nil {
		return err
	}

	result, err := clients.API().ListAppApprovalRequests(ctx, token, appID, requestsFlags.teamIDs)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "lock",
		Text:      "App Requests",
		Secondary: FormatRequestsSuccess(result.Requests),
	}))
	return nil
}

// requestsAppSelection decides the app to check and a token of the app team.
//
// An app named by ID with the app flag is checked without a project so that
// apps missing from a project can be checked too. The team of that app is
// gathered from the authenticated accounts instead of the project apps.
func requestsAppSelection(ctx context.Context, clients *shared.ClientFactory) (appID string, token string, err error) {
	if types.IsAppID(clients.Config.AppFlag) {
		auth, err := requestsTeamSelectPromptFunc(ctx, clients, "Select the team of the app", nil)
		if err != nil {
			return "", "", err
		}
		if auth == nil || auth.Token == "" {
			return "", "", slackerror.New(slackerror.ErrCredentialsNotFound)
		}
		clients.Auth().SetSelectedAuth(ctx, *auth, clients.Config, clients.Os)
		return clients.Config.AppFlag, auth.Token, nil
	}
	selection, err := requestsAppSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return "", "", err
	}
	if selection.App.AppID == "" {
		return "", "", slackerror.New(slackerror.ErrAppNotFound)
	}
	return selection.App.AppID, selection.Auth.Token, nil
}

// FormatRequestsSuccess formats the install request of each team
func FormatRequestsSuccess(requests []api.AppsApprovalsRequest) (secondaryText []string) {
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].TeamID < requests[j].TeamID
	})
	field := func(label string, value string) string {
		return fmt.Sprintf(style.Indent(style.Secondary("%-13s %s")), label+":", value)
	}
	for _, request := range requests {
		secondaryText = append(secondaryText, fmt.Sprintf(style.Bold("%s:"), request.TeamID))
		secondaryText = append(secondaryText, field("Request ID", request.ID))
		secondaryText = append(secondaryText, field("Status", formatRequestStatus(request.Status)))
		secondaryText = append(secondaryText, field("Requested", formatRequestTime(request.DateCreated)))
		if request.DateResolved > 0 {
			secondaryText = append(secondaryText, field("Resolved", formatRequestTime(request.DateResolved)))
		}
		if request.CancelledBy != "" {
			secondaryText = append(secondaryText, field("Cancelled by", formatRequestCancelledBy(request.CancelledBy)))
		}
		if request.CanSelfApprove {
			secondaryText = append(secondaryText, style.Indent(style.Secondary("You can install this app without approval. Please cancel the request.")))
		}
	}
	if len(secondaryText) <= 0 {
		secondaryText = append(secondaryText, "You have not requested to install this app")
	}
	return
}

// formatRequestTime displays a Unix timestamp in the local timezone
func formatRequestTime(timestamp int64) string {
	if timestamp <= 0 {
		return "unknown"
	}
	return time.Unix(timestamp, 0).Format(requestsTimeFormat)
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
