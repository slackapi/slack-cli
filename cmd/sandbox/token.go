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
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// getSandboxToken returns the token to use for sandbox API operations.
// It uses the --token flag if provided, otherwise resolves from stored credentials.
func getSandboxToken(ctx context.Context, clients *shared.ClientFactory, tokenFlag string) (string, error) {
	token, _, err := getSandboxTokenAndAuth(ctx, clients, tokenFlag)
	return token, err
}

// getSandboxTokenAndAuth returns the token and auth used for sandbox API operations.
// When --token is provided, auth is resolved via AuthWithToken (may have limited fields).
// Otherwise auth comes from stored credentials.
func getSandboxTokenAndAuth(ctx context.Context, clients *shared.ClientFactory, tokenFlag string) (string, *types.SlackAuth, error) {
	if tokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, tokenFlag)
		if err != nil {
			return "", nil, err
		}
		return tokenFlag, &auth, nil
	}

	auth, err := resolveAuthForSandbox(ctx, clients)
	if err != nil {
		return "", nil, err
	}

	return auth.Token, auth, nil
}

// resolveAuthForSandbox gets the appropriate auth for sandbox operations.
// If the global --token flag is set, that is used. Otherwise, if --team is set,
// uses the auth that matches that team; else the first available auth.
func resolveAuthForSandbox(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, error) {
	// Check persistent token flag (from root)
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

	// If --team flag is set, find matching auth
	if clients.Config.TeamFlag != "" {
		for _, auth := range auths {
			if auth.TeamID == clients.Config.TeamFlag || auth.TeamDomain == clients.Config.TeamFlag {
				return &auth, nil
			}
		}
		return nil, slackerror.New(slackerror.ErrTeamNotFound).
			WithMessage("No auth found for team: " + clients.Config.TeamFlag).
			WithRemediation("Run 'slack auth list' to see your authorized workspaces")
	}

	// Use first auth
	return &auths[0], nil
}

// parseLabels parses a comma-separated key=value string into a map
func parseLabels(labelsStr string) map[string]string {
	if labelsStr == "" {
		return nil
	}

	labels := make(map[string]string)
	for _, pair := range strings.Split(labelsStr, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return labels
}
