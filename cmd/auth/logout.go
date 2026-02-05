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

package auth

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flags
var allFlag bool

func NewLogoutCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout [flags]",
		Short: "Log out of a team",
		Long:  "Log out of a team, removing any local credentials",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "auth logout", Meaning: "Select a team to log out of"},
			{Command: "auth logout --all", Meaning: "Log out of all team"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AuthLogoutStart)
			span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.logout")
			defer span.Finish()

			if allFlag && clients.Config.TeamFlag != "" {
				return slackerror.New(slackerror.ErrMismatchedFlags).
					WithDetails(slackerror.ErrorDetails{slackerror.ErrorDetail{
						Message: "The --team flag cannot be used with --all",
					}})
			}

			auths, err := promptUserLogout(ctx, clients, cmd)
			if err != nil {
				return err
			}

			if err := logout(ctx, clients, auths); err != nil {
				return err
			}
			printLogoutSuccess(ctx, clients, auths)
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&allFlag, "all", "A", false, "logout of all workspaces")

	return cmd
}

// logout revokes the user and refresh tokens for the provided auths
func logout(ctx context.Context, clients *shared.ClientFactory, auths []types.SlackAuth) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "logout")
	defer span.Finish()

	for _, auth := range auths {
		if err := handleAuthRemoval(ctx, clients, auth); err != nil {
			return err
		}
	}
	return nil
}

// handleAuthRemoval revokes the passed authorization and removes it from the credentials file
func handleAuthRemoval(ctx context.Context, clients *shared.ClientFactory, auth types.SlackAuth) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "handleAuthRemoval")
	defer span.Finish()

	// Update the API Host and Logstash Host to be the selected/default auth
	clients.Config.APIHostResolved = clients.Auth().ResolveAPIHost(ctx, clients.Config.APIHostFlag, &auth)
	clients.Config.LogstashHostResolved = clients.Auth().ResolveLogstashHost(ctx, clients.Config.APIHostResolved, clients.Config.Version)

	// First, try to revoke the xoxe-xoxp (auth) token credential
	var xoxpToken = auth.Token
	if err := clients.Auth().RevokeToken(ctx, xoxpToken); err != nil {
		return err
	}

	// Next, try to revoke the refresh token xoxe-1 credential
	var refreshToken = auth.RefreshToken
	if refreshToken != "" {
		if err := clients.Auth().RevokeToken(ctx, refreshToken); err != nil {
			return err
		}
	}

	// Once successfully revoked, remove from credentials.json
	if _, err := clients.Auth().DeleteAuth(ctx, auth); err != nil {
		return err
	}

	// Update the API Host and Logstash Host to be the selected/default auth
	clients.Config.APIHostResolved = clients.Auth().ResolveAPIHost(ctx, clients.Config.APIHostFlag, nil)
	clients.Config.LogstashHostResolved = clients.Auth().ResolveLogstashHost(ctx, clients.Config.APIHostResolved, clients.Config.Version)

	return nil
}

// promptUserLogout collects the authorizations to revoke
func promptUserLogout(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) ([]types.SlackAuth, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "promptUserLogout")
	defer span.Finish()

	if clients.Config.TeamFlag == "" && cmd.Flag("team").Changed {
		// Handle usage of empty --team flag
		return []types.SlackAuth{}, slackerror.New(slackerror.ErrMissingFlag).
			WithMessage("The argument is missing from the --team flag")
	}

	// Gather all available auths
	auths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return []types.SlackAuth{}, err
	}

	if allFlag || (len(auths) <= 1 && !cmd.Flag("team").Changed) {
		return auths, nil
	}

	// Create a list of teams for the prompt options
	authTeamDomains := make([]string, len(auths))
	authTeamIDs := make([]string, len(auths))
	authTeamDomainLabels := make([]string, len(auths))

	// Build labels
	for ii, auth := range auths {
		authTeamDomains[ii] = auth.TeamDomain
		authTeamDomainLabels[ii] = FormatAuthLabel(auth)
		authTeamIDs[ii] = auth.TeamID
	}
	// Sort options
	err = prompts.SortAlphaNumeric(authTeamDomainLabels, authTeamDomains, authTeamIDs)
	if err != nil {
		return []types.SlackAuth{}, err
	}

	// Collect the team to logout of
	var selectedTeamID string
	var selectedTeamDomain string
	selection, err := clients.IO.SelectPrompt(ctx, "Select an authorization to revoke", authTeamDomainLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("team"),
		Required: true,
	})
	if err != nil {
		return []types.SlackAuth{}, err
	} else {
		if selection.Flag {
			// flag was provided, that will be returned as the option
			selectedTeamDomain = selection.Option
		} else {
			selectedTeamDomain = authTeamDomains[selection.Index]
			selectedTeamID = authTeamIDs[selection.Index]
		}
	}

	// Find the matching authentication
	for _, auth := range auths {
		if auth.TeamID == selectedTeamID || auth.TeamID == clients.Config.TeamFlag || auth.TeamDomain == selectedTeamDomain {
			return []types.SlackAuth{auth}, nil
		}
	}
	return []types.SlackAuth{}, slackerror.New(slackerror.ErrInvalidAuth).
		WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{
				Message: fmt.Sprintf("Cannot revoke authentication tokens for '%s'", selectedTeamDomain),
				Code:    slackerror.ErrCredentialsNotFound,
			},
		}).
		WithRemediation("You might already be logged out of this team!")
}

// printLogoutSuccess outputs a message notifying of successful revokes
func printLogoutSuccess(ctx context.Context, clients *shared.ClientFactory, auths []types.SlackAuth) {
	revokedTeamDomain := "all teams"
	if len(auths) == 1 {
		revokedTeamDomain = auths[0].TeamDomain
	}
	revokedAuthText := fmt.Sprintf("Authorization successfully revoked for %s", style.Highlight(revokedTeamDomain))
	if len(auths) == 0 {
		revokedAuthText = "All authorizations successfully revoked!"
	}
	logoutNextSteps := []string{
		fmt.Sprintf("Login to a new team with %s", style.Commandf("login", false)),
	}

	clients.IO.PrintTrace(ctx, slacktrace.AuthLogoutSuccess)
	clients.IO.PrintInfo(ctx, false, fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji:     "wastebasket",
		Text:      revokedAuthText,
		Secondary: logoutNextSteps,
	})))
}

// FormatAuthLabel returns a formatted auth label for user selection during logout
func FormatAuthLabel(auth types.SlackAuth) string {
	return fmt.Sprintf("%s %s %s", auth.TeamDomain, style.Faint(auth.TeamID), style.Faint(fmt.Sprintf("(%s)", auth.AuthLevel())))
}
