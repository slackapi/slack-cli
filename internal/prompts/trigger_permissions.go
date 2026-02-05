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

package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/pflag"
)

func TriggerAccessLabels(current types.Permission) ([]string, []types.Permission) {
	// Prepare the prompt labels, with the active distribution listed first
	distributions := []types.Permission{}
	distributionLabels := []string{}
	optionLabels := map[types.Permission]string{
		types.PermissionAppCollaborators: "app collaborators only",
		types.PermissionEveryone:         "everyone",
		types.PermissionNamedEntities:    "specific entities",
	}

	distributionLabels = append(distributionLabels, fmt.Sprintf("%s (current)", optionLabels[current]))
	distributions = append(distributions, current)
	delete(optionLabels, current)
	for dist, ol := range optionLabels {
		distributionLabels = append(distributionLabels, ol)
		distributions = append(distributions, dist)
	}

	return distributionLabels, distributions
}

// TriggerChooseNamedEntityPrompt displays a selection prompt that lists available actions one can take to manage
// the list of users or channels with access to the trigger. This function returns the chosen action and user(s) or channel(s) it
// should be applied to but does not execute the action.
func TriggerChooseNamedEntityPrompt(ctx context.Context, clients *shared.ClientFactory, accessAction string, currentAccessType types.Permission, hasIncludeAppCollabFlag bool) (string, string, bool, error) {
	actions := []string{
		"add_user",
		"remove_user",
		"add_channel",
		"remove_channel",
		"add_workspace",
		"remove_workspace",
		"add_organization",
		"remove_organization",
		"cancel",
	}
	actionLabels := []string{
		"grant a user access",
		"revoke a user's access",
		"grant a channel access",
		"revoke a channel's access",
		"grant a workspace access",
		"revoke a workspace's access",
		"grant an organization access",
		"revoke an organization's access",
		"cancel",
	}

	switch accessAction {
	case "grant":
		actions = []string{
			"add_user",
			"add_channel",
			"add_workspace",
			"add_organization",
			"cancel",
		}
		actionLabels = []string{
			"grant a user access",
			"grant a channel access",
			"grant a workspace access",
			"grant an organization access",
			"cancel",
		}
	case "revoke":
		actions = []string{
			"remove_user",
			"remove_channel",
			"remove_workspace",
			"remove_organization",
			"cancel",
		}
		actionLabels = []string{
			"revoke a user's access",
			"revoke a channel's access",
			"revoke a workspace's access",
			"revoke an organization's access",
			"cancel",
		}
	}

	var selectedAction string
	selection, err := clients.IO.SelectPrompt(ctx, "Choose an action", actionLabels, iostreams.SelectPromptConfig{
		Flags: []*pflag.Flag{
			clients.Config.Flags.Lookup("grant"),
			clients.Config.Flags.Lookup("revoke"),
		},
		Required: true,
	})
	if err != nil {
		return "", "", false, err
	} else if selection.Flag && clients.Config.Flags.Lookup("grant").Changed {
		selectedAction = "grant"
	} else if selection.Flag && clients.Config.Flags.Lookup("revoke").Changed {
		selectedAction = "revoke"
	} else if selection.Prompt {
		selectedAction = actions[selection.Index]
	}

	if selectedAction == "cancel" {
		return selectedAction, "", false, nil
	}

	shouldAddCollaborators := false
	if strings.HasPrefix(selectedAction, "add_") && currentAccessType != types.PermissionNamedEntities && !hasIncludeAppCollabFlag {
		shouldAddCollaborators, err = AddAppCollaboratorsToNamedEntitiesPrompt(ctx, clients.IO)
		if err != nil {
			return "", "", false, err
		}
	}

	// ask for specific ID types
	var prompt string
	if strings.HasSuffix(selectedAction, "_user") {
		prompt = "Provide the ID(s) of one or more users in your workspace (e.g. 'U00001,U00002'):"
	} else if strings.HasSuffix(selectedAction, "_channel") {
		prompt = "Provide the ID(s) of one or more channels in your workspace (e.g. 'C00001, C00002'):"
	} else if strings.HasSuffix(selectedAction, "_workspace") {
		prompt = "Provide the ID(s) of one or more workspaces (e.g. 'T00001,T00002'):"
	} else if strings.HasSuffix(selectedAction, "_organization") {
		prompt = "Provide the ID(s) of one or more organizations (e.g. 'E00001,E00002'):"
	}

	ids, err := clients.IO.InputPrompt(ctx, prompt, iostreams.InputPromptConfig{
		Required: true,
	})
	if err != nil {
		return "", "", false, err
	}
	ids = goutils.UpperCaseTrimAll(ids) // trim white space eg. 'ID1, ID2' -> 'ID1,ID2'
	return selectedAction, ids, shouldAddCollaborators, nil
}

func TriggerChooseNamedEntityActionPrompt(ctx context.Context, clients *shared.ClientFactory) (string, error) {
	actions := []string{
		"grant",
		"revoke",
		"cancel",
	}
	actionLabels := []string{
		"grant access",
		"revoke access",
		"cancel",
	}

	var selectedAction string
	selection, err := clients.IO.SelectPrompt(ctx, "Choose an action", actionLabels, iostreams.SelectPromptConfig{
		Flags: []*pflag.Flag{
			clients.Config.Flags.Lookup("grant"),
			clients.Config.Flags.Lookup("revoke"),
		},
		Required: true,
	})
	if err != nil {
		return "", err
	} else if selection.Flag && clients.Config.Flags.Lookup("grant").Changed {
		selectedAction = "grant"
	} else if selection.Flag && clients.Config.Flags.Lookup("revoke").Changed {
		selectedAction = "revoke"
	} else if selection.Prompt {
		selectedAction = actions[selection.Index]
	}

	return selectedAction, nil
}

// AddAppCollaboratorsToNamedEntitiesPrompt displays a confirmation prompt for adding app collaborators to trigger access named entities list
func AddAppCollaboratorsToNamedEntitiesPrompt(ctx context.Context, IO iostreams.IOStreamer) (bool, error) {
	IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "warning",
		Text:  style.Bold("Warning"),
		Secondary: []string{
			"App collaborators will lose their current access if not included in the list of entities",
		},
	}))
	return IO.ConfirmPrompt(ctx, "Include app collaborators?", true)
}
