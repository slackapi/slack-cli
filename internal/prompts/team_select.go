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
	"slices"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// PromptTeamSlackAuth prompts the user to select a team that they're logged in to and returns the auth information.
// If the user is only logged in to one team, we return it by default.
func PromptTeamSlackAuth(ctx context.Context, clients *shared.ClientFactory, promptText string) (*types.SlackAuth, error) {

	allAuths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return &types.SlackAuth{}, err
	}

	if len(allAuths) == 1 {
		return &allAuths[0], nil
	}

	slices.SortFunc(allAuths, func(i, j types.SlackAuth) int {
		if i.TeamDomain == j.TeamDomain {
			return strings.Compare(i.TeamID, j.TeamID)
		}
		return strings.Compare(i.TeamDomain, j.TeamDomain)
	})

	var teamLabels []string
	for _, auth := range allAuths {
		teamLabels = append(
			teamLabels,
			style.TeamSelectLabel(auth.TeamDomain, auth.TeamID),
		)
	}

	selection, err := clients.IO.SelectPrompt(
		ctx,
		promptText,
		teamLabels,
		iostreams.SelectPromptConfig{
			Required: true,
			Flag:     clients.Config.Flags.Lookup("team"),
		},
	)
	if err != nil {
		return &types.SlackAuth{}, err
	}

	if selection.Prompt {
		clients.Auth().SetSelectedAuth(ctx, allAuths[selection.Index], clients.Config, clients.Os)
		return &allAuths[selection.Index], nil
	}

	teamMatch := false
	teamIndex := -1
	for ii, auth := range allAuths {
		if selection.Option == auth.TeamID || selection.Option == auth.TeamDomain {
			if teamMatch {
				return &types.SlackAuth{}, slackerror.New(slackerror.ErrMissingAppTeamID).
					WithMessage("The team cannot be determined by team domain").
					WithRemediation("Provide the team ID for the installed app")
			}
			teamMatch = true
			teamIndex = ii
		}
	}
	if !teamMatch {
		return &types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound)
	}

	clients.Auth().SetSelectedAuth(ctx, allAuths[teamIndex], clients.Config, clients.Os)
	return &allAuths[teamIndex], nil
}
