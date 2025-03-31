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

package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/pflag"
)

func AccessLabels(current types.Permission) ([]string, []types.Permission) {
	// Prepare the prompt labels, with the active distribution listed first
	distributions := []types.Permission{}
	distributionLabels := []string{}
	optionLabels := map[types.Permission]string{
		types.APP_COLLABORATORS: "app collaborators only",
		types.EVERYONE:          "everyone",
		types.NAMED_ENTITIES:    "specific users",
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

// ChooseNamedEntityPrompt displays a selection prompt that lists available actions one can take to manage
// the list of users with access to the function. This function returns the chosen action and user(s) it
// should be applied to but does not execute the action.
func ChooseNamedEntityPrompt(ctx context.Context, clients *shared.ClientFactory) (string, string, error) {

	actions := []string{
		"add",
		"remove",
		"cancel",
	}
	actionLabels := []string{
		"grant a user access",
		"revoke a user's access",
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
		return "", "", err
	} else if selection.Flag && clients.Config.Flags.Lookup("grant").Changed {
		selectedAction = "add"
	} else if selection.Flag && clients.Config.Flags.Lookup("revoke").Changed {
		selectedAction = "remove"
	} else if selection.Prompt {
		selectedAction = actions[selection.Index]
	}

	if selectedAction == "cancel" {
		return selectedAction, "", nil
	}

	// ask for user IDs
	users, err := clients.IO.InputPrompt(ctx, "Provide the ID(s) of one or more users in your workspace (eg. 'U00001,U00002'):", iostreams.InputPromptConfig{
		Required: true,
	})
	if err != nil {
		return "", "", err
	}

	// trim white space eg. 'ID1, ID2' -> 'ID1,ID2'
	users = strings.ReplaceAll(users, " ", "")

	return selectedAction, users, nil
}
