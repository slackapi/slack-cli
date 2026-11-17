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
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Delays used while a newly created sandbox becomes available to the sandbox login API.
// These are variables so tests can skip the waiting.
var (
	sandboxLoginInitialDelay = 500 * time.Millisecond
	sandboxLoginRetryDelay   = 1 * time.Second
	sandboxLoginMaxAttempts  = 3
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox <subcommand> [flags]",
		Short: "Manage developer sandboxes",
		Long: `Manage Slack developer sandboxes without leaving your terminal.
Use the --team flag to select the authentication to use for these commands.

Prefer a UI? Head over to
{{LinkText "https://api.slack.com/developer-program/sandboxes"}}

New to the Developer Program? Sign up at
{{LinkText "https://api.slack.com/developer-program/join"}}`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{}),
		Aliases: []string{"sandboxes"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCreateCommand(clients))
	cmd.AddCommand(NewDeleteCommand(clients))
	cmd.AddCommand(NewListCommand(clients))
	cmd.AddCommand(NewSignupCommand(clients))

	return cmd
}

// getSandboxAuth returns the auth to be used for sandbox management.
// Uses the global --token or --team flag if present, otherwise prompts the user to select a team.
func getSandboxAuth(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, error) {
	// Check for the global --token flag
	if clients.Config.TokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return nil, err
		}
		return &auth, nil
	}

	// Prompt the user to select a team to use for authentication
	auth, err := prompts.PromptTeamSlackAuth(ctx, clients, "Select a team for authentication", &prompts.PromptTeamSlackAuthConfig{HelpText: "Your email address on the selected team should match your Slack developer account"})
	if err != nil {
		return nil, err
	}

	return auth, nil
}

// loginAndSaveSandboxAuth logs the user in to the sandbox with their token and saves the
// returned tokens as a new auth, so the user is logged in to the sandbox after it's created.
// Newly created sandboxes can take a moment to become available, so the login is retried
// with exponential backoff while the sandbox isn't found.
func loginAndSaveSandboxAuth(ctx context.Context, clients *shared.ClientFactory, apiClient api.APIInterface, token, sandboxTeamID string) error {
	time.Sleep(sandboxLoginInitialDelay)

	var result *api.SandboxLoginResult
	var err error
	for attempt := 0; attempt < sandboxLoginMaxAttempts; attempt++ {
		result, err = apiClient.SandboxLogin(ctx, token, sandboxTeamID)
		if err == nil {
			break
		}
		if !slackerror.Is(err, slackerror.ErrSandboxNotFound) {
			return err
		}
		// Don't wait after the last attempt
		if attempt < sandboxLoginMaxAttempts-1 {
			delay := sandboxLoginRetryDelay * time.Duration(1<<attempt) // 1s, 2s
			clients.IO.PrintDebug(ctx, "Sandbox not yet available, retrying in %v (attempt %d/%d)", delay, attempt+1, sandboxLoginMaxAttempts)
			time.Sleep(delay)
		}
	}
	if err != nil {
		return err
	}

	isEnterprise := types.IsEnterpriseTeamID(result.SandboxTeamID)
	sandboxAuth := types.SlackAuth{
		Token:               result.AccessToken,
		RefreshToken:        result.RefreshToken,
		ExpiresAt:           result.ExpiresAt,
		TeamID:              result.SandboxTeamID,
		TeamDomain:          result.SandboxDomain,
		UserID:              result.SandboxUserID,
		LastUpdated:         time.Now(),
		IsEnterpriseInstall: isEnterprise,
	}
	if isEnterprise {
		sandboxAuth.EnterpriseID = result.SandboxTeamID
	}
	if !clients.Auth().IsAPIHostSlackProd(apiClient.Host()) {
		// Save non-production API hosts so the sandbox auth keeps targeting the same host
		apiHost := apiClient.Host()
		sandboxAuth.APIHost = &apiHost
	}

	_, _, err = clients.Auth().SetAuth(ctx, sandboxAuth)
	return err
}
