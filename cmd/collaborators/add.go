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
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// addCmdFlags contains the flag set for this command
type addCmdFlags struct {
	permissionType string
}

// addFlags implements values of the command flag set
var addFlags addCmdFlags

func NewAddCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [email|user_id]",
		Aliases: []string{"new", "include"},
		Short:   "Add a new collaborator to the app",
		Long:    "Add a collaborator to your app by Slack email address or user ID",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "collaborator add", Meaning: "Add a collaborator via prompt"},
			{Command: "collaborator add bot@slack.com", Meaning: "Add a collaborator from email"},
			{Command: "collaborator add USLACKBOT", Meaning: "Add a collaborator by user ID"},
		}),
		Args: cobra.RangeArgs(0, 1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runAddCommandFunc(ctx, clients, cmd, args)
		},
	}
	cmd.Flags().StringVarP(&addFlags.permissionType, "permission-type", "P", "", "collaborator permission type: reader, owner")
	cmd.Flag("permission-type").Hidden = true
	return cmd
}

// runAddCommandFunc adds a user as an app collaborator using inputs and the API
func runAddCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "cmd.Collaborators.Add")
	defer span.Finish()

	// Get the app auth selection from the flag or prompt
	selection, err := teamAppSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return err
	}
	if err = cmdutil.AppExists(selection.App, selection.Auth); err != nil {
		return err
	}
	slackUser, err := promptCollaboratorsAdd(ctx, clients, args, selection)
	if err != nil {
		return err
	}
	err = clients.APIInterface().AddCollaborator(ctx, selection.Auth.Token, selection.App.AppID, slackUser)
	if err != nil {
		if clients.Config.WithExperimentOn(experiment.ReadOnlyAppCollaborators) && strings.Contains(err.Error(), "user_already_owner") {
			cmd.Println()
			cmd.Println(style.Sectionf(style.TextSection{
				Emoji: "bulb",
				Text:  fmt.Sprintf("Use the %s command to update an existing collaborator", style.Highlight("collaborator update")),
			}))
		}
		return slackerror.Wrap(err, "Error adding collaborator")
	}
	printCollaboratorsAddSuccess(ctx, clients, selection.App.AppID, slackUser)
	return nil
}

// promptCollaboratorsAdd decides a user with permissions to add to an app
func promptCollaboratorsAdd(
	ctx context.Context,
	clients *shared.ClientFactory,
	args []string,
	selection prompts.SelectedApp,
) (
	slackUser types.SlackUser,
	err error,
) {
	switch {
	case len(args) > 0:
		slackUser, err = promptCollaboratorsAddSlackUserArguments(ctx, clients, args[0])
	case clients.IO.IsTTY():
		slackUser, err = promptCollaboratorsAddSlackUserPrompts(ctx, clients, selection)
	default:
		return types.SlackUser{}, slackerror.New(slackerror.ErrMissingInput).
			WithMessage("No collaborator was provided").
			WithRemediation("Include a collaborator ID in the command arguments")
	}
	if err != nil {
		return types.SlackUser{}, err
	}
	switch clients.Config.Flags.Lookup("permission-type").Changed {
	case true:
		slackUser.PermissionType, err = promptCollaboratorsAddPermissionFlags(ctx, clients, addFlags.permissionType)
	default:
		slackUser.PermissionType, err = promptCollaboratorsAddPermissionPrompts(ctx, clients)
	}
	if err != nil {
		return types.SlackUser{}, err
	}
	return slackUser, nil
}

// promptCollaboratorsAddArguments gathers a collaborator ID or email from input
// and sets the permission type
func promptCollaboratorsAddSlackUserArguments(
	ctx context.Context,
	clients *shared.ClientFactory,
	input string,
) (
	slackUser types.SlackUser,
	err error,
) {
	if parsedEmail, err := mail.ParseAddress(input); err != nil {
		slackUser.ID = input
	} else {
		slackUser.Email = parsedEmail.Address
	}
	return slackUser, nil
}

// promptCollaboratorsAddSlackUserPrompts gathers a collaborator from prompts
func promptCollaboratorsAddSlackUserPrompts(
	ctx context.Context,
	clients *shared.ClientFactory,
	selection prompts.SelectedApp,
) (
	slackUser types.SlackUser,
	err error,
) {
	response, err := clients.IO.InputPrompt(
		ctx,
		"Provide a new collaborator email or user ID",
		iostreams.InputPromptConfig{
			Required: true,
		},
	)
	if err != nil {
		return types.SlackUser{}, err
	}
	if parsedEmail, err := mail.ParseAddress(response); err != nil {
		slackUser.ID = response
	} else {
		slackUser.Email = parsedEmail.Address
	}
	slackUser.PermissionType = types.OWNER
	return slackUser, nil
}

// promptCollaboratorsAddPermissionPrompts gathers the collaborator permission
// from selection if the experiment allows
func promptCollaboratorsAddPermissionPrompts(
	ctx context.Context,
	clients *shared.ClientFactory,
) (
	permission types.AppCollaboratorPermission,
	err error,
) {
	switch clients.Config.WithExperimentOn(experiment.ReadOnlyAppCollaborators) {
	case false:
		return types.OWNER, nil
	default:
		permissionLabels := []string{
			"owner",
			"reader",
		}
		response, err := clients.IO.SelectPrompt(
			ctx,
			"Decide the collaborator permission",
			permissionLabels,
			iostreams.SelectPromptConfig{
				Required: true,
			},
		)
		if err != nil {
			return "", err
		}
		return types.StringToAppCollaboratorPermission(response.Option)
	}
}

// promptCollaboratorsAddPermissionFlags gathers the collaborator permission
// from flags if the experiment allows
func promptCollaboratorsAddPermissionFlags(
	ctx context.Context,
	clients *shared.ClientFactory,
	flag string,
) (
	permission types.AppCollaboratorPermission,
	err error,
) {
	switch clients.Config.WithExperimentOn(experiment.ReadOnlyAppCollaborators) {
	case true:
		return types.StringToAppCollaboratorPermission(addFlags.permissionType)
	default:
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "construction",
			Text:  fmt.Sprintf("This command is under construction. Use at your own risk %s", style.Emoji("skull")),
			Secondary: []string{
				fmt.Sprintf("Bypass this message with the %s flag", style.Highlight("--experiment read-only-collaborators")),
			},
		}))
		return "", slackerror.New(slackerror.ErrMissingExperiment)
	}
}

// printCollaboratorsAddSuccess outputs a message when addition is done
func printCollaboratorsAddSuccess(
	ctx context.Context,
	clients *shared.ClientFactory,
	appID string,
	collaborator types.SlackUser,
) {
	clients.IO.PrintTrace(ctx, slacktrace.CollaboratorAddSuccess)
	clients.IO.PrintTrace(ctx,
		slacktrace.CollaboratorAddCollaborator,
		collaborator.ShorthandF(),
		string(collaborator.PermissionType),
	)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "bust_in_silhouette",
		Text: fmt.Sprintf(
			"Member '%s' is now %s",
			collaborator.ShorthandF(),
			collaborator.PermissionType.AppCollaboratorPermissionF(),
		),
		Secondary: []string{
			fmt.Sprintf("App '%s' had a collaborator added", appID),
		},
	}))
}
