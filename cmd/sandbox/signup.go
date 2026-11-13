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
	"strings"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type signupFlags struct {
	firstName              string
	lastName               string
	postalCode             string
	countryOption          string
	newsletterSubscription bool
	acceptTermsOfService   bool
}

var signupCmdFlags signupFlags

func NewSignupCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signup [flags]",
		Short: "Sign up for the Slack Developer Program",
		Long: `Sign up for the Slack Developer Program!

The developer program gives you access to sandbox environments for building
and testing Slack apps isolated from your production workspace. To get started,
log in to your production workspace via the 'login' command.

Learn more at {{LinkText "https://api.slack.com/developer-program"}}`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox signup", Meaning: "Sign up for the Developer Program"},
			{Command: "sandbox signup --first-name Anna --last-name Banana --postal-code 94105 --country US", Meaning: "Sign up with the provided details"},
			{Command: "sandbox signup --team T0123456 --accept-tos", Meaning: "Sign up with a specific team and accept the terms"},
		}),
		Args:   cobra.NoArgs,
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !clients.Config.WithExperimentOn(experiment.SandboxSignup) {
				return slackerror.New(slackerror.ErrExperimentRequired).
					WithRemediation("Enable the %s experiment with %s",
						style.Highlight(string(experiment.SandboxSignup)),
						style.CommandText("--experiment sandbox-signup"),
					)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSignupCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&signupCmdFlags.firstName, "first-name", "", "Your first name")
	cmd.Flags().StringVar(&signupCmdFlags.lastName, "last-name", "", "Your last name")
	cmd.Flags().StringVar(&signupCmdFlags.postalCode, "postal-code", "", "Your postal code")
	cmd.Flags().StringVar(&signupCmdFlags.countryOption, "country", "", "Your country (ISO 3166-1 alpha-2 format, e.g. US)")
	cmd.Flags().BoolVar(&signupCmdFlags.newsletterSubscription, "newsletter", false, "Subscribe to the developer newsletter")
	cmd.Flags().BoolVar(&signupCmdFlags.acceptTermsOfService, "accept-tos", false, "Accept the Privacy Policy, API Terms of Service, and Developer Program Agreement")

	return cmd
}

func runSignupCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	auth, err := selectAuthForSignup(ctx, clients)
	if err != nil {
		return err
	}

	// Prompt user to accept terms unless --accept-tos or --force flag is used
	if !signupCmdFlags.acceptTermsOfService && !clients.Config.ForceFlag {
		if err := promptTermsAcceptance(ctx, clients); err != nil {
			return err
		}
	}

	email := prefillUserInfoFromAuth(ctx, clients, auth, &signupCmdFlags)

	info, err := collectUserInfo(ctx, clients, &signupCmdFlags)
	if err != nil {
		return err
	}
	info.email = email

	req := api.SignupRequest{
		FirstName:              info.firstName,
		LastName:               info.lastName,
		PostalCode:             info.postalCode,
		CountryOption:          info.countryOption,
		NewsletterSubscription: signupCmdFlags.newsletterSubscription,
		AcceptTermsOfService:   true, // Terms were accepted via the --accept-tos flag, --force flag, or prompt
	}

	printSignupDetails(ctx, clients, info, signupCmdFlags.newsletterSubscription)

	_, err = apiClientForAuth(clients, auth).SignupDeveloperProgramVerified(ctx, auth.Token, req)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "tada",
		Text:  "Welcome to the Slack Developer Program!",
	}))

	clients.IO.PrintInfo(ctx, false, "\nRun %s to create a developer sandbox", style.Commandf("sandbox create", false))
	return nil
}

// apiClientForAuth returns an API client targeting the host for the given auth.
// An explicit --apihost (or --slackdev) flag takes precedence over the auth's stored API host,
// since clients.API() is already configured with the resolved flag value.
func apiClientForAuth(clients *shared.ClientFactory, auth *types.SlackAuth) api.APIInterface {
	apiClient := clients.API()
	if clients.Config.APIHostFlag != "" || clients.Config.SlackDevFlag {
		return apiClient
	}
	if auth != nil && auth.APIHost != nil {
		apiClient.SetHost(*auth.APIHost)
	}
	return apiClient
}

// selectAuthForSignup returns the auth to link to the Developer Program account.
// Uses the global --token or --team flag if present, otherwise prompts the user to select an account.
func selectAuthForSignup(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, error) {
	if clients.Config.TokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return nil, err
		}
		return &auth, nil
	}

	allAuths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return nil, err
	}
	if len(allAuths) == 0 {
		return nil, slackerror.New(slackerror.ErrNotAuthed).
			WithMessage("You must be logged in to sign up for the Developer Program").
			WithRemediation("Run %s to sign in to a workspace, then try again", style.Commandf("login", false))
	}

	teamFlag := clients.Config.Flags.Lookup("team")
	labels := make([]string, 0, len(allAuths))
	options := make([]iostreams.PromptOption, 0, len(allAuths))
	for _, auth := range allAuths {
		label := style.TeamSelectLabel(auth.TeamDomain, auth.TeamID)
		if userInfo, err := apiClientForAuth(clients, &auth).UsersInfo(ctx, auth.Token, auth.UserID); err == nil && userInfo.Profile.Email != "" {
			label = fmt.Sprintf("%s %s", userInfo.Profile.Email, style.Faint(fmt.Sprintf("(%s %s)", auth.TeamDomain, auth.TeamID)))
		}
		labels = append(labels, label)
		options = append(options, iostreams.PromptOption{
			Label: label,
			Flag:  teamFlag,
			Value: auth.TeamID,
		})
	}

	selection, err := clients.IO.SelectPrompt(ctx, "Select an email for your developer program account", labels, iostreams.SelectPromptConfig{
		Required: true,
		Flag:     teamFlag,
		Options:  options,
		Help:     "The selected login wil be used to fetch your user details and create your account",
	})
	if err != nil {
		if slackerror.Is(err, slackerror.ErrProcessInterrupted) {
			clients.IO.SetExitCode(iostreams.ExitCancel)
		}
		return nil, err
	}

	if selection.Prompt {
		return &allAuths[selection.Index], nil
	}
	for i, auth := range allAuths {
		if selection.Option == auth.TeamID || selection.Option == auth.TeamDomain {
			return &allAuths[i], nil
		}
	}
	return nil, slackerror.New(slackerror.ErrCredentialsNotFound)
}

// prefillUserInfoFromAuth fetches the user's Slack profile and fills in any unset name flags.
// Returns the email address of the profile, if available.
func prefillUserInfoFromAuth(ctx context.Context, clients *shared.ClientFactory, auth *types.SlackAuth, flags *signupFlags) string {
	if auth.UserID == "" {
		return ""
	}
	userInfo, err := apiClientForAuth(clients, auth).UsersInfo(ctx, auth.Token, auth.UserID)
	if err != nil {
		return ""
	}

	// Prefer the profile name fields, falling back to splitting the real name
	realFirst, realLast := splitName(userInfo.RealName)
	if flags.firstName == "" {
		flags.firstName = userInfo.Profile.FirstName
		if flags.firstName == "" {
			flags.firstName = realFirst
		}
	}
	if flags.lastName == "" {
		flags.lastName = userInfo.Profile.LastName
		if flags.lastName == "" {
			flags.lastName = realLast
		}
	}

	return userInfo.Profile.Email
}

// splitName splits a full name into first and last name
func splitName(fullName string) (firstName, lastName string) {
	parts := strings.SplitN(strings.TrimSpace(fullName), " ", 2)
	firstName = parts[0]
	if len(parts) == 2 {
		lastName = strings.TrimSpace(parts[1])
	}
	return firstName, lastName
}

// userInfo holds the collected user information for signup
type userInfo struct {
	firstName     string
	lastName      string
	email         string
	postalCode    string
	countryOption string
}

// promptTermsAcceptance shows the terms and prompts the user to accept them
func promptTermsAcceptance(ctx context.Context, clients *shared.ClientFactory) error {
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "scroll",
		Text:  "Terms and Conditions",
		Secondary: []string{
			"By signing up, you agree to the following:",
			fmt.Sprintf("• Privacy Policy: %s", style.LinkText("https://slack.com/trust/privacy/privacy-policy")),
			fmt.Sprintf("• API Terms of Service: %s", style.LinkText("https://slack.com/terms-of-service/api")),
			fmt.Sprintf("• Developer Program Agreement: %s", style.LinkText("https://slack.com/terms-of-service/slack-developer-program")),
		},
	}))

	acceptTerms, err := clients.IO.ConfirmPrompt(ctx, "Do you agree to these terms?", iostreams.ConfirmPromptConfig{Default: true})
	if err != nil {
		if slackerror.Is(err, slackerror.ErrProcessInterrupted) {
			clients.IO.SetExitCode(iostreams.ExitCancel)
		}
		return err
	}
	if !acceptTerms {
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
			Emoji: "information_source",
			Text:  "Signup Cancelled",
			Secondary: []string{
				"You must agree to the terms to join the Developer Program.",
			},
		}))
		clients.IO.SetExitCode(iostreams.ExitCancel)
		return slackerror.New(slackerror.ErrProcessInterrupted)
	}
	return nil
}

// collectUserInfo prompts for a first and last name when not already set.
// Postal code and country are optional for authenticated signups so are only read from flags.
func collectUserInfo(ctx context.Context, clients *shared.ClientFactory, flags *signupFlags) (*userInfo, error) {
	var err error
	info := &userInfo{
		firstName:     flags.firstName,
		lastName:      flags.lastName,
		postalCode:    flags.postalCode,
		countryOption: strings.ToUpper(strings.TrimSpace(flags.countryOption)),
	}

	if info.firstName == "" {
		info.firstName, err = clients.IO.InputPrompt(ctx, "Enter your first name", iostreams.InputPromptConfig{Required: true})
		if err != nil {
			return nil, err
		}
	}

	if info.lastName == "" {
		info.lastName, err = clients.IO.InputPrompt(ctx, "Enter your last name", iostreams.InputPromptConfig{Required: true})
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

// printSignupDetails shows the signup details
func printSignupDetails(ctx context.Context, clients *shared.ClientFactory, info *userInfo, newsletter bool) {
	details := []string{fmt.Sprintf("Name: %s %s", info.firstName, info.lastName)}
	if info.email != "" {
		details = append(details, fmt.Sprintf("Email: %s", info.email))
	}
	if info.countryOption != "" {
		details = append(details, fmt.Sprintf("Country: %s", info.countryOption))
	}
	if info.postalCode != "" {
		details = append(details, fmt.Sprintf("Postal Code: %s", info.postalCode))
	}
	details = append(details, fmt.Sprintf("Newsletter: %v", newsletter))
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "clipboard",
		Text:      "Developer Program Signup",
		Secondary: details,
	}))
}
