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
	"fmt"
	"sort"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/pkg/apps"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Handle to client's function used for testing
var listFunc = apps.List

// Flags

type listCmdFlags struct {
	displayAllOrgGrants bool
}

var listFlags listCmdFlags

// NewListCommand returns a new Cobra command
func NewListCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"ls"},
		Short:   "List teams with the app installed",
		Long:    "List all teams that have installed the app",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "app list", Meaning: "List all teams with the app installed"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	cmd.Flags().BoolVar(&listFlags.displayAllOrgGrants, "all-org-workspace-grants", false, "display all workspace grants for an app\ninstalled to an organization")

	return cmd
}

// runListCommand will execute the list command
func runListCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	envs, _, err := listFunc(ctx, clients)
	if err != nil {
		return err
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "house_buildings",
		Text:      "Apps",
		Secondary: formatListSuccess(envs),
	}))
	return nil
}

// formatListSuccess formats details about the list of project apps
func formatListSuccess(apps []types.App) (secondaryText []string) {
	for _, app := range apps {
		if app.AppID == "" {
			continue
		}
		teamDomain := app.TeamDomain
		if app.IsDev && !strings.HasSuffix(teamDomain, style.LocalRunNameTag) {
			teamDomain = style.LocalRunDisplayName(teamDomain)
		}
		secondaryText = append(secondaryText, fmt.Sprintf(
			style.Bold("%s:"), teamDomain))
		secondaryText = append(secondaryText, fmt.Sprintf(
			style.Indent(style.Secondary("App  ID: %s")), app.AppID))
		secondaryText = append(secondaryText, fmt.Sprintf(
			style.Indent(style.Secondary("Team ID: %s")), app.TeamID))
		if app.UserID != "" {
			secondaryText = append(secondaryText, fmt.Sprintf(
				style.Indent(style.Secondary("User ID: %s")), app.UserID))
		}
		secondaryText = append(secondaryText, fmt.Sprintf(
			style.Indent(style.Secondary("Status:  %s")), app.InstallStatus.String()))
		if app.IsEnterpriseApp() && len(app.EnterpriseGrants) > 0 {
			secondaryText = appendEnterpriseWorkspaceGrantInfo(secondaryText, app)
		}
	}
	if len(secondaryText) <= 0 {
		secondaryText = append(secondaryText, "This project has no apps")
	}
	return
}

// Append workspace grant information for the enterprise app to the display text
func appendEnterpriseWorkspaceGrantInfo(secondaryText []string, app types.App) []string {

	// Sort workspace grants into alphabetical order by domain
	sort.Slice(app.EnterpriseGrants, func(i, j int) bool {
		return app.EnterpriseGrants[i].WorkspaceDomain < app.EnterpriseGrants[j].WorkspaceDomain
	})

	spacerText := " "
	if len(app.EnterpriseGrants) > 1 {
		spacerText = "\n     "
	}

	secondaryText = append(secondaryText, fmt.Sprintf(style.Indent(style.Secondary("Workspace %s:%s%s (%s)")),
		style.Pluralize("Grant", "Grants", len(app.EnterpriseGrants)),
		spacerText,
		app.EnterpriseGrants[0].WorkspaceDomain,
		app.EnterpriseGrants[0].WorkspaceID))

	// Print workspace names. Defaults to a limit of 3 but can be overridden by the displayAllOrgGrants flag.
	limit := 3
	if listFlags.displayAllOrgGrants {
		limit = len(app.EnterpriseGrants)
	}
	for i := 1; i < limit; i++ {
		if i >= len(app.EnterpriseGrants) {
			break
		}
		secondaryText = append(secondaryText, style.Indent(fmt.Sprintf("  %s (%s)",
			app.EnterpriseGrants[i].WorkspaceDomain,
			app.EnterpriseGrants[i].WorkspaceID)))
	}

	if len(app.EnterpriseGrants) > limit && !listFlags.displayAllOrgGrants {
		remaining := len(app.EnterpriseGrants) - limit
		secondaryText = append(secondaryText, style.Indent(fmt.Sprintf("  ... and %d other %s", remaining, style.Pluralize("workspace", "workspaces", remaining))))
	}

	return secondaryText
}
