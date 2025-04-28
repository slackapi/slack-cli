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

package function

import (
	"context"
	"fmt"
	"sort"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/pflag"
)

// chooseFunctionPrompt displays a selection prompt that lists all available functions and returns the function callback ID.
func chooseFunctionPrompt(ctx context.Context, clients *shared.ClientFactory, functions []types.Function) (string, error) {
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].CallbackID < functions[j].CallbackID
	})

	functionLabels := []string{}
	for _, fn := range functions {
		functionLabels = append(functionLabels, fmt.Sprintf("%s (%s)", fn.CallbackID, fn.Title))
	}

	var selectedFunction string
	selection, err := clients.IO.SelectPrompt(ctx, "Choose a function", functionLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("name"),
		Required: true,
	})
	if err != nil {
		return "", err
	} else if selection.Flag {
		selectedFunction = selection.Option
	} else if selection.Prompt {
		selectedFunction = functions[selection.Index].CallbackID
	}

	return selectedFunction, nil
}

// chooseDistributionPrompt displays a selection prompt that lists distribution options for the function and optionally
// gives app collaborators access when a more restrictive distribution type is chosen. It returns the chosen distribution type.
func chooseDistributionPrompt(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	token string,
) (types.Permission, error) {

	// Get the function's active distribution type
	currentDist, _, err := clients.APIInterface().FunctionDistributionList(ctx, functionFlag, app.AppID)
	if err != nil {
		return "", err
	}

	labels, distributions := prompts.AccessLabels(currentDist)

	// Execute the prompt
	var selectedDistribution types.Permission
	selection, err := clients.IO.SelectPrompt(ctx, "Who would you like to have access to your function?", labels, iostreams.SelectPromptConfig{
		Flags: []*pflag.Flag{
			clients.Config.Flags.Lookup("users"),
			clients.Config.Flags.Lookup("app-collaborators"),
			clients.Config.Flags.Lookup("everyone"),
		},
		Required: true,
	})
	if err != nil {
		// Prefer named entities on mismatch then prompt for collaborators later
		if slackerror.ToSlackError(err).Code == slackerror.ErrMismatchedFlags &&
			!clients.Config.Flags.Lookup("everyone").Changed {
			selectedDistribution = types.NAMED_ENTITIES
		} else {
			return "", err
		}
	} else if selection.Flag {
		switch {
		case clients.Config.Flags.Lookup("app-collaborators").Changed:
			selectedDistribution = types.APP_COLLABORATORS
		case clients.Config.Flags.Lookup("everyone").Changed:
			selectedDistribution = types.EVERYONE
		case clients.Config.Flags.Lookup("users").Changed:
			selectedDistribution = types.NAMED_ENTITIES
		}
	} else if selection.Prompt {
		selectedDistribution = distributions[selection.Index]
	}

	// Optional follow-up: if the function is moving from an access type where collaborators have access,
	// to named_entities where they do not unless explicitly added, offer to add them automatically
	if (currentDist == types.APP_COLLABORATORS || currentDist == types.EVERYONE) &&
		selectedDistribution == types.NAMED_ENTITIES {
		err := addCollaboratorsToNamedEntitiesPrompt(ctx, clients, app, token)
		if err != nil {
			return "", err
		}
	}

	return selectedDistribution, nil
}

// addCollaboratorsToNamedEntitiesPrompt displays a confirmation prompt for adding app collaborators to a
// function's access list when the distribution type is changing to named_entities from a distribution
// type where collaborators had access by default
func addCollaboratorsToNamedEntitiesPrompt(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	token string,
) error {
	shouldAddCollaborators, err := clients.IO.ConfirmPrompt(ctx, "Do you want your app collaborators to maintain their current access?", true)
	if err != nil {
		return err
	}
	if shouldAddCollaborators {
		return AddCollaboratorsToNamedEntities(ctx, clients, app, token)
	}
	return nil
}
