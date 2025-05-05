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
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewRemoveCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:     "remove [email|user_id]",
		Aliases: []string{"delete"},
		Short:   "Remove a collaborator from an app",
		Long:    "Remove a collaborator from an app by Slack email address or user ID",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "collaborator remove", Meaning: "Remove collaborator on prompt"},
			{Command: "collaborator remove bot@slack.com", Meaning: "Remove collaborator by email"},
			{Command: "collaborator remove USLACKBOT", Meaning: "Remove collaborator using ID"},
		}),
		Args: cobra.RangeArgs(0, 1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runRemoveCommandFunc(ctx, clients, cmd, args)
		},
	}
}

// runRemoveCommandFunc removes a user as an app collaborator from an app
func runRemoveCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "cmd.Collaborators.Remove")
	defer span.Finish()
	selection, err := teamAppSelectPromptFunc(ctx, clients, prompts.ShowHostedOnly, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return err
	}
	slackUser, err := promptCollaboratorsRemoveSlackUser(ctx, clients, args, selection)
	if err != nil {
		return err
	}
	if err = cmdutil.AppExists(selection.App, selection.Auth); err != nil {
		return err
	}
	warnings, err := clients.API().RemoveCollaborator(ctx, selection.Auth.Token, selection.App.AppID, slackUser)
	if err != nil {
		return err
	}
	printCollaboratorsRemoveSuccess(ctx, clients, selection.App.AppID, slackUser)
	printWarnings(ctx, clients, warnings)
	return nil
}

// promptCollaboratorsRemoveSlackUser determines which user to remove as an app
// collaborator
func promptCollaboratorsRemoveSlackUser(
	ctx context.Context,
	clients *shared.ClientFactory,
	args []string,
	selection prompts.SelectedApp,
) (
	types.SlackUser,
	error,
) {
	if len(args) > 0 {
		return promptCollaboratorsRemoveSlackUserArguments(args[0])
	} else if clients.IO.IsTTY() {
		return promptCollaboratorsRemoveSlackUserPrompts(ctx, clients, selection)
	}
	return types.SlackUser{}, slackerror.New(slackerror.ErrMissingInput).
		WithMessage("No collaborator was provided").
		WithRemediation("Include a collaborator ID in the command arguments")
}

// promptCollaboratorsRemoveSlackUserArguments gathers an ID or email from input
func promptCollaboratorsRemoveSlackUserArguments(input string) (types.SlackUser, error) {
	var slackUser types.SlackUser
	if parsedEmail, err := mail.ParseAddress(input); err != nil {
		slackUser.ID = input
	} else {
		slackUser.Email = parsedEmail.Address
	}
	return slackUser, nil
}

// promptCollaboratorsRemoveSlackUserPrompts gathers a collaborator from prompts
func promptCollaboratorsRemoveSlackUserPrompts(
	ctx context.Context,
	clients *shared.ClientFactory,
	selection prompts.SelectedApp,
) (
	slackUser types.SlackUser,
	err error,
) {
	collaborators, err := clients.API().ListCollaborators(ctx, selection.Auth.Token, selection.App.AppID)
	if err != nil {
		return types.SlackUser{}, err
	}
	sortCollaboratorsList(collaborators)
	var collaboratorLabels []string
	for _, collaborator := range collaborators {
		collaboratorLabels = append(collaboratorLabels, fmt.Sprintf(
			"%s%s",
			collaborator.UserName,
			style.Secondary(strings.TrimLeft(collaborator.String(), collaborator.UserName)),
		))
	}
	response, err := clients.IO.SelectPrompt(
		ctx,
		"Remove a collaborator",
		collaboratorLabels,
		iostreams.SelectPromptConfig{
			Required: true,
		},
	)
	if err != nil {
		return types.SlackUser{}, err
	}
	collaborator := collaborators[response.Index]
	if collaborator.ID == selection.Auth.UserID && !clients.Config.ForceFlag {
		confirm, err := clients.IO.ConfirmPrompt(
			ctx,
			"Are you sure you want to remove yourself?",
			false,
		)
		if err != nil {
			return types.SlackUser{}, err
		}
		if !confirm {
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "ring_buoy",
				Text:  "No collaborators were changed in the happenings of this process",
			}))
			return types.SlackUser{}, slackerror.New(slackerror.ErrProcessInterrupted)
		}
	}
	return collaborator, nil
}

// printCollaboratorsRemoveSuccess outputs a message when deletion is done
func printCollaboratorsRemoveSuccess(
	ctx context.Context,
	clients *shared.ClientFactory,
	appID string,
	collaborator types.SlackUser,
) {
	clients.IO.PrintTrace(ctx, slacktrace.CollaboratorRemoveSuccess)
	clients.IO.PrintTrace(ctx,
		slacktrace.CollaboratorRemoveCollaborator,
		collaborator.ShorthandF(),
	)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "bust_in_silhouette",
		Text: fmt.Sprintf(
			"Member '%s' is no longer %s",
			collaborator.ShorthandF(),
			collaborator.PermissionType.AppCollaboratorPermissionF(),
		),
		Secondary: []string{
			fmt.Sprintf("App '%s' had a collaborator removed", appID),
		},
	}))
}

func printWarnings(ctx context.Context,
	clients *shared.ClientFactory, warnings slackerror.Warnings) {
	if len(warnings) <= 0 {
		return
	}
	warningMsgs := []string{}
	for _, warning := range warnings {
		warningMsgs = append(warningMsgs, warning.Message)
	}
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji:     "bulb",
		Text:      "Warnings",
		Secondary: warningMsgs,
	}))
}
