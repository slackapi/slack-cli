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

	"github.com/slackapi/slack-cli/internal/iostreams"
	authpkg "github.com/slackapi/slack-cli/internal/pkg/auth"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// flag values
var deprecatedUserTokenArg string
var tokenFlag string
var ticketArg string
var challengeCodeArg string
var noPromptFlag bool
var serviceTokenFlag bool

const invalidFlagComboMessage = "The --auth and --token flags cannot be used together. Please use"
const deprecatedUserTokenMessage = "The --auth flag has been removed"

func NewLoginCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to a Slack account",
		Long:  "Log in to a Slack account in your team",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "auth login", Meaning: "Login to a Slack account with prompts"},
			{Command: "auth login --no-prompt", Meaning: "Login to a Slack account without prompts, this returns a ticket"},
			{Command: "auth login --challenge 6d0a31c9 --ticket ISQWLiZT0OtMLO3YWNTJO0...", Meaning: "Complete login using ticket and challenge code"},
			{Command: "auth login --token xoxp-...", Meaning: "Login with a user token"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := RunLoginCommand(clients, cmd)
			return err
		},
	}

	// Support for login with existing token via --auth flag is now deprecated, recommended to use --token now
	var authFlagName = "auth"
	cmd.Flags().StringVarP(&deprecatedUserTokenArg, authFlagName, "", "", "provide a user token for pre-authenticated login")
	_ = cmd.Flags().MarkHidden(authFlagName)

	cmd.Flags().StringVarP(&tokenFlag, "token", "", "", "provide a token for a pre-authenticated login")

	// Support login in promptless fashion
	cmd.Flags().BoolVarP(&noPromptFlag, "no-prompt", "", false, "login without prompts using ticket and challenge code")
	cmd.Flags().StringVarP(&ticketArg, "ticket", "", "", "provide an auth ticket value")
	cmd.Flags().StringVarP(&challengeCodeArg, "challenge", "", "", "provide a challenge code for pre-authenticated login")

	return cmd
}

// RunLoginCommand prompts the user to select a team or login to a new team
func RunLoginCommand(clients *shared.ClientFactory, cmd *cobra.Command) (types.SlackAuth, error) {
	if cmd == nil {
		return types.SlackAuth{}, slackerror.New("Login command is nil")
	}
	ctx := cmd.Context()

	// --auth and --token flags cannot be used together
	if tokenFlag != "" && deprecatedUserTokenArg != "" {
		invalidFlagComboRemediation := fmt.Sprintf("%s %s", invalidFlagComboMessage, style.Highlight("--token <token>"))
		return types.SlackAuth{}, slackerror.New(slackerror.ErrMismatchedFlags).
			WithMessage("%s", invalidFlagComboRemediation)
	}

	if deprecatedUserTokenArg != "" {
		// Deprecation warning for --auth flag. Recommend to use --token instead
		cmd.PrintErrf("\n%s", style.Sectionf(style.TextSection{
			Emoji: "construction",
			Text:  "Deprecation of --auth",
			Secondary: []string{
				deprecatedUserTokenMessage,
				fmt.Sprintf("Specify a token for use with the token flag %s", style.Highlight("--token <token>")),
			},
		}))

		tokenFlag = deprecatedUserTokenArg
		proceedMessage := fmt.Sprintf("\n%s %s", "Proceeding with ", style.Highlight("--token <token>"))
		cmd.Print(style.SectionSecondaryf("%s", proceedMessage))
	}

	// When --no-prompt flag supplied OR --ticket and --challenge code flags provided
	// attempt to login in a promptless fashion
	if (noPromptFlag) || (ticketArg != "" || challengeCodeArg != "") {
		selectedAuth, credentialsPath, err := authpkg.LoginNoPrompt(ctx, clients, ticketArg, challengeCodeArg, serviceTokenFlag)
		if err != nil {
			return types.SlackAuth{}, err
		}
		if selectedAuth.Token != "" {
			printAuthSuccess(cmd, clients.IO, credentialsPath, selectedAuth.Token)
			printAuthNextSteps(ctx, clients)
		}
		return selectedAuth, err
	}

	selectedAuth, credentialsPath, err := authpkg.LoginWithClients(ctx, clients, tokenFlag, serviceTokenFlag)
	if err != nil {
		return types.SlackAuth{}, err
	} else {
		printAuthSuccess(cmd, clients.IO, credentialsPath, selectedAuth.Token)
		printAuthNextSteps(ctx, clients)
	}

	return selectedAuth, nil
}

func printAuthSuccess(cmd *cobra.Command, IO iostreams.IOStreamer, credentialsPath string, token string) {
	var secondaryLog string
	if credentialsPath != "" {
		secondaryLog = fmt.Sprintf("Authorization data was saved to %s", style.HomePath(credentialsPath))
	} else if serviceTokenFlag && tokenFlag == "" {
		secondaryLog = fmt.Sprintf("Service token:\n\n  %s\n\nMake sure to copy the token now and save it safely.", token)
	}

	ctx := cmd.Context()
	IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "key",
		Text:      "You've successfully authenticated!",
		Secondary: []string{secondaryLog},
	}))
	IO.PrintTrace(ctx, slacktrace.AuthLoginSuccess)
}

// printAuthNextSteps suggests possible commands to run after logging in
func printAuthNextSteps(ctx context.Context, clients *shared.ClientFactory) {
	_, project := clients.SDKConfig.Exists()
	if !project {
		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji: "bulb",
			Text:  fmt.Sprintf("Get started by creating a new app with %s", style.Commandf("create my-app", true)),
			Secondary: []string{
				fmt.Sprintf("Explore the details of available commands with %s", style.Commandf("help", false)),
			},
		}))
	} else {
		clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
			Emoji: "bulb",
			Text:  fmt.Sprintf("Review existing installations of the app with %s", style.Commandf("app list", false)),
			Secondary: []string{
				fmt.Sprintf("Explore the details of available commands with %s", style.Commandf("help", false)),
			},
		}))
	}
}
