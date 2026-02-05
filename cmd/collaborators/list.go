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

package collaborators

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewListCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{},
		Short:   "List all collaborators of an app",
		Long:    "List all collaborators of an app",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "collaborator list", Meaning: "List all of the collaborators"},
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
}

// runListCommand will execute the list command
func runListCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	span, _ := opentracing.StartSpanFromContext(ctx, "cmd.Collaborators.List")
	defer span.Finish()

	// Get the app auth selection from the flag or prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return err
	}
	app := selection.App
	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}
	collaborators, err := clients.API().ListCollaborators(ctx, selection.Auth.Token, app.AppID)
	if err != nil {
		return slackerror.Wrap(err, "Error listing collaborators")
	}
	sortCollaboratorsList(collaborators)
	printCollaboratorsListSuccess(ctx, clients, app.AppID, collaborators)
	return nil
}

// sortCollaboratorsList orders collaborators by permission then username
func sortCollaboratorsList(collaborators []types.SlackUser) {
	slices.SortFunc(collaborators, func(a types.SlackUser, b types.SlackUser) int {
		switch {
		case a.PermissionType == types.OWNER && b.PermissionType != types.OWNER:
			return -1
		case a.PermissionType != types.OWNER && b.PermissionType == types.OWNER:
			return 1
		default:
			return strings.Compare(a.UserName, b.UserName)
		}
	})
}

// printCollaboratorsListSuccess outputs a list of collaborators on an app
func printCollaboratorsListSuccess(
	ctx context.Context,
	clients *shared.ClientFactory,
	appID string,
	collaborators []types.SlackUser,
) {
	clients.IO.PrintTrace(ctx, slacktrace.CollaboratorListSuccess)
	clients.IO.PrintTrace(ctx, slacktrace.CollaboratorListCount, fmt.Sprintf("%d", len(collaborators)))
	var list []string
	for _, collaborator := range collaborators {
		list = append(list, collaborator.String())
		clients.IO.PrintTrace(ctx,
			slacktrace.CollaboratorListCollaborator,
			collaborator.ID,
			collaborator.UserName,
			collaborator.Email,
			string(collaborator.PermissionType),
		)
	}
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "house",
		Text: fmt.Sprintf(
			"The app '%s' has %d %s",
			appID,
			len(collaborators),
			style.Pluralize("collaborator", "collaborators", len(collaborators)),
		),
		Secondary: list,
	}))
}
