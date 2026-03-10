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

package sandbox

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// getSandboxAuth returns the auth used for sandbox management.
// Uses the global --token flag if present, otherwise delegates to resolveAuthForSandbox.
func getSandboxAuth(ctx context.Context, clients *shared.ClientFactory) (string, *types.SlackAuth, error) {
	auth, err := resolveAuthForSandbox(ctx, clients)
	if err != nil {
		return "", nil, err
	}

	clients.Config.APIHostResolved = clients.Auth().ResolveAPIHost(ctx, clients.Config.APIHostFlag, auth)
	clients.Config.LogstashHostResolved = clients.Auth().ResolveLogstashHost(ctx, clients.Config.APIHostResolved, clients.CLIVersion)

	return auth.Token, auth, nil
}

// resolveAuthForSandbox determines what auth to use for sandbox operations.
// If the global --token flag is set, we will use that.
// Else if the global --team flag is set, we will use the associated token for that team.
// Else if the user is only logged in to one team, we will default to that team's auth.
// Else we will prompt the user to select a team to use for authentication.
func resolveAuthForSandbox(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, error) {
	// Check for the global --token flag
	if clients.Config.TokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return nil, err
		}
		return &auth, nil
	}

	auths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return nil, err
	}

	if len(auths) == 0 {
		return nil, slackerror.New(slackerror.ErrCredentialsNotFound).
			WithMessage("You must be logged in to manage sandboxes").
			WithRemediation("Run 'slack login' to authenticate, or use --token for CI/CD")
	}

	// Check for the global --team flag
	if clients.Config.TeamFlag != "" {
		for _, auth := range auths {
			if auth.TeamID == clients.Config.TeamFlag || auth.TeamDomain == clients.Config.TeamFlag {
				return &auth, nil
			}
		}
		return nil, slackerror.New(slackerror.ErrTeamNotFound).
			WithMessage("No auth found for team: %s", clients.Config.TeamFlag).
			WithRemediation("Run 'slack auth list' to see your authorized workspaces")
	}

	// Prompt the user to select a team to use for authentication (if there are multiple auths), or default to the user's only auth
	if len(auths) == 1 {
		return &auths[0], nil
	}
	type authOption struct {
		auth  types.SlackAuth
		label string
	}
	options := make([]authOption, 0, len(auths))
	for _, a := range auths {
		options = append(options, authOption{
			auth:  a,
			label: fmt.Sprintf("%s %s", a.TeamDomain, style.Secondary(a.TeamID)),
		})
	}
	slices.SortFunc(options, func(a, b authOption) int {
		if c := strings.Compare(a.auth.TeamDomain, b.auth.TeamDomain); c != 0 {
			return c
		}
		return strings.Compare(a.auth.TeamID, b.auth.TeamID)
	})
	labels := make([]string, 0, len(options))
	for _, opt := range options {
		labels = append(labels, opt.label)
	}
	selection, err := clients.IO.SelectPrompt(ctx, "Select a team for authentication", labels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("team"),
		Required: true,
	})
	if err != nil {
		return nil, err
	}
	switch {
	case selection.Flag:
		for _, opt := range options {
			if opt.auth.TeamID == selection.Option || opt.auth.TeamDomain == selection.Option {
				return &opt.auth, nil
			}
		}
		return nil, slackerror.New(slackerror.ErrTeamNotFound).
			WithMessage("No auth found for team: %s", selection.Option).
			WithRemediation("Run 'slack auth list' to see your authorized workspaces")
	case selection.Prompt:
		return &options[selection.Index].auth, nil
	default:
		return nil, slackerror.New(slackerror.ErrInvalidAuth)
	}
}
