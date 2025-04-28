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

package triggers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "List details of existing triggers",
		Long:  "List details of existing triggers",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "trigger access", Meaning: "Select who can run a trigger"},
			{Command: "trigger create", Meaning: "Create a new trigger"},
			{Command: "trigger delete --trigger-id Ft01234ABCD", Meaning: "Delete an existing trigger"},
			{Command: "trigger info --trigger-id Ft01234ABCD", Meaning: "Get details for a trigger"},
			{Command: "trigger list", Meaning: "List details for all existing triggers"},
			{Command: "trigger update --trigger-id Ft01234ABCD", Meaning: "Update a trigger definition"},
		}),
		Aliases: []string{"triggers"},
		Args:    cobra.NoArgs,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify command is run in a project directory
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	// Add child commands
	cmd.AddCommand(NewListCommand(clients))
	cmd.AddCommand(NewCreateCommand(clients))
	cmd.AddCommand(NewDeleteCommand(clients))
	cmd.AddCommand(NewUpdateCommand(clients))
	cmd.AddCommand(NewAccessCommand(clients))
	cmd.AddCommand(NewInfoCommand(clients))

	return cmd
}

// sprintTrigger converts a trigger into a readable format
func sprintTrigger(ctx context.Context, t types.DeployedTrigger, clients *shared.ClientFactory, singleTriggerInfo bool, app types.App) ([]string, error) {
	timeFormat := "2006-01-02 15:04:05 Z07:00"
	var triggerText = []string{""}

	triggerType := fmt.Sprintf("(%s)", t.Type)

	if t.Name != "" {
		triggerText = append(triggerText, fmt.Sprintf(
			style.Indent("%s %s %s"),
			t.Name,
			style.Faint(t.ID),
			style.Secondary(triggerType),
		))
	} else {
		triggerText = append(triggerText, fmt.Sprintf(
			style.Indent("Trigger ID: %s %s"),
			t.ID,
			style.Secondary(triggerType),
		))
	}

	timeAgoTag := fmt.Sprintf("(%s)", style.TimeAgo(t.DateCreated))
	triggerText = append(triggerText, fmt.Sprintf(
		style.Indent(style.Secondary("Created: %s %s")),
		style.Faint(time.Unix(int64(t.DateCreated), 0).Format(timeFormat)),
		style.Secondary(timeAgoTag),
	))

	if t.DateUpdated != t.DateCreated {
		timeAgoTag := fmt.Sprintf("(%s)", style.TimeAgo(t.DateUpdated))
		triggerText = append(triggerText, fmt.Sprintf(
			style.Indent(style.Secondary("Updated: %s %s")),
			style.Faint(time.Unix(int64(t.DateUpdated), 0).Format(timeFormat)),
			style.Secondary(timeAgoTag),
		))
	}

	token := config.GetContextToken(ctx)

	// Get app owners & collaborators
	collaborators, err := clients.ApiInterface().ListCollaborators(ctx, token, app.AppID)
	if err != nil {
		return []string{}, err
	}
	if len(collaborators) > 0 {
		triggerText = append(triggerText, fmt.Sprint(
			style.Indent(style.Secondary("Collaborators:")),
		))
		for _, collaborator := range collaborators {
			userInfo, err := clients.ApiInterface().UsersInfo(ctx, token, collaborator.ID)
			if err != nil {
				return []string{}, err
			}
			triggerText = append(triggerText, fmt.Sprintf(
				style.Indent("  %s %s %s"),
				userInfo.RealName, style.Secondary("@"+userInfo.Profile.DisplayName), style.Secondary(userInfo.ID),
			))
		}
	}
	// Get trigger's ACL type
	accessType, entitiesAccessList, err := clients.ApiInterface().TriggerPermissionsList(ctx, token, t.ID)
	if err != nil {
		return []string{}, err
	}
	// Get trigger's ACL entities details
	if singleTriggerInfo {
		if accessType != types.EVERYONE && len(entitiesAccessList) <= 0 {
			triggerText = append(triggerText, fmt.Sprintf(
				style.Indent(style.Secondary("  %s")),
				"nobody",
			))
		} else {
			if accessType == types.EVERYONE {
				var everyoneAccessTypeDescription = types.GetAccessTypeDescriptionForEveryone(app)
				triggerText = append(triggerText, fmt.Sprintf(
					style.Indent(style.Secondary("Can be found and used by:\n  %s")),
					style.Indent(style.Secondary(everyoneAccessTypeDescription)),
				))
			} else if accessType == types.APP_COLLABORATORS {
				triggerText = append(triggerText, fmt.Sprint(
					style.Indent(style.Secondary("Can be found and used by:")),
				))
				for _, entity := range entitiesAccessList {
					userInfo, err := clients.ApiInterface().UsersInfo(ctx, token, entity)
					if err != nil {
						return []string{}, err
					}
					triggerText = append(triggerText, fmt.Sprintf(
						style.Indent("  %s %s %s"),
						userInfo.RealName, style.Secondary("@"+userInfo.Profile.DisplayName), style.Secondary(userInfo.ID),
					))
				}
			} else if accessType == types.NAMED_ENTITIES {
				triggerText = append(triggerText, fmt.Sprint(
					style.Indent(style.Secondary("Can be found and used by:")),
				))
				namedEntitiesAccessMap := namedEntitiesAccessMap(entitiesAccessList)
				if len(namedEntitiesAccessMap["users"]) > 0 {

					for _, entity := range namedEntitiesAccessMap["users"] {
						userInfo, err := clients.ApiInterface().UsersInfo(ctx, token, entity)
						if err != nil {
							return []string{}, err
						}
						triggerText = append(triggerText, fmt.Sprintf(
							style.Indent("  %s %s %s"),
							userInfo.RealName, style.Secondary("@"+userInfo.Profile.DisplayName), style.Secondary(userInfo.ID),
						))
					}
				}
				if len(namedEntitiesAccessMap["channels"]) > 0 {
					for _, entity := range namedEntitiesAccessMap["channels"] {
						channelInfo, err := clients.ApiInterface().ChannelsInfo(ctx, token, entity)
						if err != nil {
							return []string{}, err
						}
						triggerText = append(triggerText, fmt.Sprintf(
							style.Indent("  %s #%s %s"),
							style.Secondary("members of"), channelInfo.Name, style.Secondary(channelInfo.ID),
						))
					}
				}
				if len(namedEntitiesAccessMap["teams"]) > 0 {
					for _, entity := range namedEntitiesAccessMap["teams"] {
						teamInfo, err := clients.ApiInterface().TeamsInfo(ctx, token, entity)
						if err != nil {
							return []string{}, err
						}
						triggerText = append(triggerText, fmt.Sprintf(
							style.Indent("  %s %s %s"),
							style.Secondary("members of workspace"), teamInfo.Name, style.Secondary(teamInfo.ID),
						))
					}
				}
				if len(namedEntitiesAccessMap["organizations"]) > 0 {
					for _, entity := range namedEntitiesAccessMap["organizations"] {
						orgInfo, err := clients.ApiInterface().TeamsInfo(ctx, token, entity)
						if err != nil {
							return []string{}, err
						}
						triggerText = append(triggerText, fmt.Sprintf(
							style.Indent("  %s %s %s"),
							style.Secondary("members of organization"), orgInfo.Name, style.Secondary(orgInfo.ID),
						))
					}
				}
			}
		}
	} else {
		accessTypeDescription := accessType.ToString()
		if accessType == types.EVERYONE {
			accessTypeDescription = types.GetAccessTypeDescriptionForEveryone(app)
		}

		triggerText = append(triggerText, fmt.Sprintf(
			style.Indent(style.Secondary("Can be found and used by:\n  %s")),
			style.Indent(style.Secondary(accessTypeDescription)),
		))
	}

	if t.Webhook != "" {
		triggerText = append(triggerText,
			style.Indent(style.Faint(style.Underline(t.Webhook))))
	}

	if t.ShortcutUrl != "" {
		triggerText = append(triggerText,
			style.Indent(style.Faint(style.Underline(t.ShortcutUrl))))
	}

	if t.Type == "event" {
		triggerText = append(triggerText, fmt.Sprintf(
			style.Indent(style.Secondary("Hint:\n  %s")),
			style.Indent(style.Secondary("Invite your app to the channel to receive the events")),
		))

		triggerText = append(triggerText, fmt.Sprintf(
			style.Indent(style.Secondary("Warning:\n  %s")),
			style.Indent(style.Secondary("Slack Connect channels are unsupported")),
		))
	}

	return triggerText, nil
}

func sprintTriggers(ctx context.Context, triggers []types.DeployedTrigger, clients *shared.ClientFactory, app types.App) ([]string, error) {
	var formattedText = []string{}

	var eventTriggers []types.DeployedTrigger
	var scheduledTriggers []types.DeployedTrigger
	var shortcutTriggers []types.DeployedTrigger
	var webhookTriggers []types.DeployedTrigger

	for _, t := range triggers {
		switch t.Type {
		case "event":
			eventTriggers = append(eventTriggers, t)
		case "scheduled":
			scheduledTriggers = append(scheduledTriggers, t)
		case "shortcut":
			shortcutTriggers = append(shortcutTriggers, t)
		case "webhook":
			webhookTriggers = append(webhookTriggers, t)
		}
	}

	if len(eventTriggers) > 0 {
		formattedText = append(formattedText, fmt.Sprintf("\n%s%s", style.Emoji("mailbox"),
			style.Pluralize("Event trigger:", "Event triggers:", len(eventTriggers))))
		sort.Slice(eventTriggers, func(i, j int) bool {
			return eventTriggers[i].DateCreated < eventTriggers[j].DateCreated
		})
		for _, t := range eventTriggers {
			trigs, err := sprintTrigger(ctx, t, clients, false, app)
			if err != nil {
				return []string{}, err
			}
			formattedText = append(formattedText, trigs...)
		}
	}

	if len(scheduledTriggers) > 0 {
		formattedText = append(formattedText, fmt.Sprintf("\n%s%s", style.Emoji("calendar"),
			style.Pluralize("Scheduled trigger:", "Scheduled triggers:", len(scheduledTriggers))))
		sort.Slice(scheduledTriggers, func(i, j int) bool {
			return scheduledTriggers[i].DateCreated < scheduledTriggers[j].DateCreated
		})
		for _, t := range scheduledTriggers {
			trigs, err := sprintTrigger(ctx, t, clients, false, app)
			if err != nil {
				return []string{}, err
			}
			formattedText = append(formattedText, trigs...)
		}
	}

	if len(shortcutTriggers) > 0 {
		formattedText = append(formattedText, fmt.Sprintf("\n%s%s", style.Emoji("link"),
			style.Pluralize("Shortcut trigger:", "Shortcut triggers:", len(shortcutTriggers))))
		sort.Slice(shortcutTriggers, func(i, j int) bool {
			return shortcutTriggers[i].DateCreated < shortcutTriggers[j].DateCreated
		})
		for _, t := range shortcutTriggers {
			trigs, err := sprintTrigger(ctx, t, clients, false, app)
			if err != nil {
				return []string{}, err
			}
			formattedText = append(formattedText, trigs...)
		}
	}

	if len(webhookTriggers) > 0 {
		formattedText = append(formattedText, fmt.Sprintf("\n%s%s", style.Emoji("hook"),
			style.Pluralize("Webhook trigger:", "Webhook triggers:", len(webhookTriggers))))
		sort.Slice(webhookTriggers, func(i, j int) bool {
			return webhookTriggers[i].DateCreated < webhookTriggers[j].DateCreated
		})
		for _, t := range webhookTriggers {
			trigs, err := sprintTrigger(ctx, t, clients, false, app)
			if err != nil {
				return []string{}, err
			}
			formattedText = append(formattedText, trigs...)
		}
	}

	return formattedText, nil
}

type promptForTriggerIDLabelOption int

const (
	defaultLabels promptForTriggerIDLabelOption = iota
	// labelsIncludeAccessType indicates that trigger access type should be included in the prompt labels
	labelsIncludeAccessType
)

func promptForTriggerID(ctx context.Context, cmd *cobra.Command, clients *shared.ClientFactory, app types.App, token string, labelOption promptForTriggerIDLabelOption) (string, error) {
	args := api.TriggerListRequest{
		AppID: app.AppID,
		Limit: 0,     // 0 means no pagation
		Type:  "all", // all means showing all types of triggers
	}
	triggers, _, err := clients.ApiInterface().WorkflowsTriggersList(ctx, token, args)
	if err != nil {
		return "", err
	}

	if len(triggers) == 0 {
		return "", slackerror.New(slackerror.ErrNoTriggers)
	}

	// Sort triggers by name (or ID if names are equal)
	sort.Slice(triggers, func(i, j int) bool {
		if triggers[i].Name == triggers[j].Name {
			return triggers[i].ID < triggers[j].ID
		}
		return triggers[i].Name < triggers[j].Name
	})

	triggerLabels := []string{}
	for _, tr := range triggers {
		if labelOption == labelsIncludeAccessType {
			accessType, _, err := clients.ApiInterface().TriggerPermissionsList(ctx, token, tr.ID)
			if err != nil {
				return "", err
			}
			triggerLabels = append(triggerLabels, fmt.Sprintf("%s %s %s", tr.Name, style.Secondary(tr.ID), style.Secondary(accessType.ToString())))
		} else {
			triggerLabels = append(triggerLabels, fmt.Sprintf("%s %s", tr.Name, style.Secondary(tr.ID)))
		}
	}

	var selectedTriggerID string
	selection, err := clients.IO.SelectPrompt(ctx, "Choose a trigger:", triggerLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("trigger-id"),
		Required: true,
		PageSize: 4,
	})
	if err != nil {
		return "", err
	} else if selection.Flag {
		selectedTriggerID = selection.Option
	} else if selection.Prompt {
		selectedTriggerID = triggers[selection.Index].ID
	}

	return selectedTriggerID, nil
}

func printNoTriggersMessage(ctx context.Context, IO iostreams.IOStreamer) {
	fmt.Println()
	IO.PrintInfo(ctx, true, style.Sectionf(style.TextSection{
		Emoji: "zap",
		Text:  "There are no triggers installed for this app",
	}))
}
