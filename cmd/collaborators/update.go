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

package collaborators

import (
	"fmt"
	"net/mail"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type updateCmdFlags struct {
	permissionType string
}

var updateFlags updateCmdFlags

func NewUpdateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <email|user_id>",
		Hidden:  true,
		Aliases: []string{},
		Short:   "Experimental command to update a collaborator's permission",
		Long:    "Experimental command to update the type of access a collaborator has to an app",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "collaborator update", Meaning: "Update a collaborator's permission"},
		}),
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !clients.Config.WithExperimentOn(experiment.ReadOnlyAppCollaborators) {
				cmd.Println()
				cmd.Println(style.Sectionf(style.TextSection{
					Emoji: "construction",
					Text:  fmt.Sprintf("This command is under construction. Use at your own risk %s", style.Emoji("skull")),
					Secondary: []string{
						fmt.Sprintf("Bypass this message with the %s flag", style.Highlight("--experiment read-only-collaborators")),
					},
				}))
				return nil
			}

			return runUpdateCommand(cmd, clients, args)
		},
	}

	cmd.Flags().StringVarP(&updateFlags.permissionType, "permission-type", "P", "", "collaborator permission type: reader, owner")

	return cmd
}

// runUpdateCommand will execute the update command
func runUpdateCommand(cmd *cobra.Command, clients *shared.ClientFactory, args []string) error {
	var span opentracing.Span
	ctx := cmd.Context()
	span, ctx = opentracing.StartSpanFromContext(ctx, "cmd.Collaborators.Update")
	defer span.Finish()

	var slackUser types.SlackUser
	parsedEmail, err := mail.ParseAddress(args[0])
	if err != nil {
		slackUser.ID = args[0]
	} else {
		slackUser.Email = parsedEmail.Address
	}

	if updateFlags.permissionType != "" {
		slackUser.PermissionType, err = types.StringToAppCollaboratorPermission(updateFlags.permissionType)
		if err != nil {
			return err
		}
	} else {
		cmd.Println(fmt.Sprintf("\n%s Specify a permission type for your collaborator with the %s flag\n", style.Emoji("warning"), style.Highlight("--permission-type")))
		return nil
	}

	// Get the app auth selection from the flag or prompt
	selection, err := teamAppSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return err
	}

	app := selection.App

	if err = cmdutil.AppExists(app, selection.Auth); err != nil {
		return err
	}

	err = clients.ApiInterface().UpdateCollaborator(ctx, selection.Auth.Token, app.AppID, slackUser)
	if err != nil {
		return slackerror.Wrap(err, "Error updating collaborator")
	}

	printSuccess(ctx, clients.IO, slackUser, "updated")

	return nil
}
