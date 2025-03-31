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

package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/auth"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"golang.org/x/mod/semver"
)

const InvalidNoPromptFlags = "Invalid arguments, both --ticket and --challenge flag values are required"

// LoginWithClients ...
func LoginWithClients(ctx context.Context, clients *shared.ClientFactory, userToken string, noRotation bool) (auth types.SlackAuth, credentialsPath string, err error) {
	return Login(ctx, clients.ApiInterface(), clients.AuthInterface(), clients.IO, userToken, noRotation)
}

// Login takes the user through the Slack CLI login process
func Login(ctx context.Context, apiClient api.ApiInterface, authClient auth.AuthInterface, io iostreams.IOStreamer, userToken string, noRotation bool) (auth types.SlackAuth, credentialsPath string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.login")
	defer span.Finish()

	// An XOXP token was provided via the "--token" flag
	if userToken != "" {
		io.PrintDebug(ctx, "user token (xoxp-) provided with the --token flag")
		return createNewLoginWithUserToken(ctx, apiClient, authClient, userToken, noRotation)
	}

	return createNewAuth(ctx, apiClient, authClient, io, noRotation)
}

// createNewLoginWithUserToken function takes in an User Token (XOXP) and uses it to grab an existing auth
// TODO (@Sarah) Remove this command
func createNewLoginWithUserToken(ctx context.Context, apiClient api.ApiInterface, authClient auth.AuthInterface, userToken string, noRotation bool) (auth types.SlackAuth, credentialsPath string, err error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "createNewLoginWithUserToken")
	defer span.Finish()

	var authSession api.AuthSession

	// validates the supplied user token
	if userToken != "" {
		var err error
		authSession, err = apiClient.ValidateSession(ctx, userToken)

		if err != nil {
			return types.SlackAuth{}, "", slackerror.New(slackerror.ErrInvalidAuth).WithRootCause(err)
		}
	} else {
		return types.SlackAuth{}, "", slackerror.New("A user token was not provided")
	}

	var newAuth = types.SlackAuth{
		Token:       userToken,
		TeamID:      *authSession.TeamID,
		UserID:      *authSession.UserID,
		LastUpdated: time.Now(),
	}

	if !authClient.IsApiHostSlackProd(apiClient.Host()) {
		// if we don't have a production apihost then save it in the credentials
		var apiHost = apiClient.Host()
		newAuth.ApiHost = &apiHost
	}

	// Grabbing the TeamDomain from the subdomain of the URL. Seems to work reliably on dev and prod workspaces
	var TeamDomain string

	// We remove the protocol part of the url, turning https://teamdomain.slack.com -> teamdomain.slack.com
	urlWithoutProtocol := strings.SplitAfter(*authSession.URL, "//")
	urlParts := strings.Split(urlWithoutProtocol[1], ".")

	// We grab the subdomain of the URL here
	TeamDomain = urlParts[0]
	newAuth.TeamDomain = TeamDomain

	// TODO: ignoring error?
	// FIXME: This is an unsafe action, since team domain does not guarantee unique auth
	// So returned auth struct may be unexpected
	auth, _ = authClient.AuthWithTeamDomain(ctx, TeamDomain)

	if auth.RefreshToken != "" {
		newAuth.RefreshToken = auth.RefreshToken
	}

	if auth.ExpiresAt != 0 {
		newAuth.ExpiresAt = auth.ExpiresAt
	}

	// We don't want save new long-lived token auth into credentials file or overwrite rotatable token with long-lived token
	if strings.HasPrefix(userToken, "xoxp-") && !strings.HasPrefix(auth.Token, "xoxp-") {
		return types.SlackAuth{}, "", nil
	}

	// Save the new auth
	auth, location, err := authClient.SetAuth(ctx, newAuth)
	if err != nil {
		return types.SlackAuth{}, "", slackerror.Wrap(err, "Error saving credentials")
	}

	span.SetTag("user", authSession.UserID)
	span.SetTag("team", authSession.TeamID)
	return auth, location, nil
}

// createNewAuth takes the user through the main Slack CLI Auth steps:
//  1. Requests an AuthTicket (base64 encoded UUID) from Slack which must be submitted in a valid Slack team
//  2. Wait for a challenge code to be submitted
//  3. Submit a request to exchange the ticket for an Auth response containing an access token
//  4. Saves auth as a credential
func createNewAuth(ctx context.Context, apiClient api.ApiInterface, authClient auth.AuthInterface, io iostreams.IOStreamer, noRotation bool) (auth types.SlackAuth, credentialsPath string, err error) {
	authTicket, err := requestAuthTicket(ctx, apiClient, io, noRotation)
	if err != nil {
		return types.SlackAuth{}, "", err
	}

	challengeCode, err := promptForChallengeCode(ctx, io)
	if err != nil {
		return types.SlackAuth{}, "", err
	}

	authExchangeRes, err := apiClient.ExchangeAuthTicket(ctx, authTicket, challengeCode, version.Get())
	if err != nil {
		return types.SlackAuth{}, "", err
	}

	return saveNewAuth(ctx, apiClient, authClient, authExchangeRes, noRotation)
}

// requestAuthTicket requests an auth ticket from Slack. The ticket must be submitted
// by a valid user within their Slack workspace, then permissions must be granted
func requestAuthTicket(ctx context.Context, apiClient api.ApiInterface, io iostreams.IOStreamer, noRotation bool) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "requestAuthTicket")
	defer span.Finish()

	cliVersion := semver.MajorMinor(version.Get())

	// Get a ticket from Slack
	if authTicketResult, err := apiClient.GenerateAuthTicket(ctx, cliVersion, noRotation); err != nil {
		return "", err
	} else {
		printAuthTicketSubmissionInstructions(ctx, io, authTicketResult.Ticket)
		return authTicketResult.Ticket, nil
	}
}

func printAuthTicketSubmissionInstructions(ctx context.Context, IO iostreams.IOStreamer, authTicket string) {
	slashCommandDetails := style.Sectionf(style.TextSection{
		Emoji: "clipboard",
		Text:  "Run the following slash command in any Slack channel or DM",
		Secondary: []string{
			"This will open a modal with user permissions for you to approve",
			"Once approved, a challenge code will be generated in Slack",
		},
	})

	// Be careful when formatting slash commands:
	// Issue #99  - When pasting slash commands with indentation, the slash command will not be detected and fail to execute.
	// Issue #129 - When pasting bolded text, some users have experienced the Slack composer bolding the slash command.
	//              This causes an error preventing the slash command from executing.
	authTicketText := fmt.Sprintf(
		"/slackauthticket %s",
		authTicket,
	)

	IO.PrintInfo(ctx, false, "\n%s\n%s\n", slashCommandDetails, authTicketText)

	IO.PrintTrace(ctx, slacktrace.AuthLoginStart)
}

// promptForChallengeCode asks user to submit the valid challenge code received from the Slack authorization
func promptForChallengeCode(ctx context.Context, IO iostreams.IOStreamer) (string, error) {
	return IO.InputPrompt(ctx, "Enter challenge code", iostreams.InputPromptConfig{
		Required: true,
	})
}

// saveNewAuth saves a new auth to the credentials and returns the auth
func saveNewAuth(ctx context.Context, apiClient api.ApiInterface, authClient auth.AuthInterface, authResult api.ExchangeAuthTicketResult, noRotation bool) (auth types.SlackAuth, credentialsPath string, err error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "saveNewAuth")
	defer span.Finish()

	// Add new auth
	newAuth := types.SlackAuth{
		ExpiresAt:           authResult.ExpiresAt,
		LastUpdated:         time.Now(),
		RefreshToken:        authResult.RefreshToken,
		Token:               authResult.Token,
		TeamDomain:          authResult.TeamDomain,
		TeamID:              authResult.TeamID,
		UserID:              authResult.UserID,
		IsEnterpriseInstall: authResult.IsEnterpriseInstall,
		EnterpriseID:        authResult.EnterpriseID,
	}

	if !authClient.IsApiHostSlackProd(apiClient.Host()) {
		// If not using a production api host then add to newAuth as apiHost
		var apiHost = apiClient.Host()
		newAuth.ApiHost = &apiHost
	}

	// Write to credentials json if serviceTokenFlag is false
	var filePath string = ""
	if !noRotation {
		_, credentialsLocation, err := authClient.SetAuth(ctx, newAuth)
		if err != nil {
			return types.SlackAuth{}, "", err
		}
		filePath = credentialsLocation
	}
	// utility logging
	span.SetTag("user", authResult.UserID)
	span.SetTag("team", authResult.TeamID)

	return newAuth, filePath, nil
}

// LoginNoPrompt initiates a login flow with no prompts
func LoginNoPrompt(ctx context.Context, clients *shared.ClientFactory, ticketArg string, challengeCodeArg string, noRotation bool) (auth types.SlackAuth, credentialsPath string, err error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "authNoPrompt")
	defer span.Finish()

	// existing ticket request, try to exchange
	if ticketArg != "" && challengeCodeArg != "" {
		authExchangeRes, err := clients.ApiInterface().ExchangeAuthTicket(ctx, ticketArg, challengeCodeArg, version.Get())
		if err != nil || !authExchangeRes.IsReady {
			return types.SlackAuth{}, "", err
		}
		savedAuth, credentialsPath, err := saveNewAuth(ctx, clients.ApiInterface(), clients.AuthInterface(), authExchangeRes, noRotation)
		if err != nil {
			return types.SlackAuth{}, "", err
		}
		return savedAuth, credentialsPath, err
	}

	// brand new login
	if ticketArg == "" && challengeCodeArg == "" {
		_, err := requestAuthTicket(ctx, clients.ApiInterface(), clients.IO, noRotation)
		if err != nil {
			return types.SlackAuth{}, "", err
		}
		return types.SlackAuth{}, "", err
	}

	// If we get to here then invalid flags have been supplied
	return types.SlackAuth{}, "", slackerror.New(slackerror.ErrMismatchedFlags).WithMessage(InvalidNoPromptFlags)
}
