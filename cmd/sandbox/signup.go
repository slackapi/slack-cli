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
	"github.com/spf13/pflag"
)

// Links to the terms and policies accepted when signing up for the Developer Program
const (
	developerProgramAgreementURL  = "https://slack.com/terms-of-service/slack-developer-program"
	slackPrivacyPolicyURL         = "https://slack.com/trust/privacy/privacy-policy"
	salesforcePrivacyStatementURL = "https://www.salesforce.com/company/privacy/"
)

// developerProfileSettingsURL is where developers can view or update their user data for their Developer Program profile
const developerProfileSettingsURL = "https://api.slack.com/developer-program/settings/profile"

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
and testing Slack apps isolated from your production workspace.

Learn more at {{LinkText "https://api.slack.com/developer-program"}}`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox signup", Meaning: "Sign up for the Developer Program"},
			{Command: "sandbox signup --first-name Anna --last-name Banana --postal-code 94105 --country US", Meaning: "Sign up with the provided user details"},
			{Command: "sandbox signup --team T0123456 --accept-tos", Meaning: "Sign up, authenticated as your user on the given team, accept the terms of service and acknowledge the privacy policies"},
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
	cmd.Flags().BoolVar(&signupCmdFlags.acceptTermsOfService, "accept-tos", false, acceptTermsOfServiceUsage())

	// Rebuild the flag usage when help is shown so the links are styled once styles are toggled
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		style.ToggleStyles(clients.IO.IsTTY() && !clients.Config.NoColor)
		cmd.Flag("accept-tos").Usage = acceptTermsOfServiceUsage()
		cmd.Parent().HelpFunc()(cmd, args)
	})

	return cmd
}

// acceptTermsOfServiceUsage returns the usage text for the --accept-tos flag
func acceptTermsOfServiceUsage() string {
	return "Confirm agreement to the Slack Developer Program Agreement\n" +
		"  " + style.LinkText(developerProgramAgreementURL) + "\n" +
		"  and acknowledge the Slack Privacy Policy\n" +
		"  " + style.LinkText(slackPrivacyPolicyURL) + "\n" +
		"  and Salesforce Privacy Statement.\n" +
		"  " + style.LinkText(salesforcePrivacyStatementURL) + "\n" +
		"  Pass this flag or confirm via the interactive prompt."
}

// termsOfServiceLinks returns a bulleted list of the terms and policies accepted at signup
func termsOfServiceLinks() []string {
	return []string{
		fmt.Sprintf("• Developer Program Agreement: %s", style.LinkText(developerProgramAgreementURL)),
		fmt.Sprintf("• Slack Privacy Policy: %s", style.LinkText(slackPrivacyPolicyURL)),
		fmt.Sprintf("• Salesforce Privacy Statement: %s", style.LinkText(salesforcePrivacyStatementURL)),
	}
}

func runSignupCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	// The --force flag skips prompts but never accepts the terms on its own
	if clients.Config.ForceFlag && !signupCmdFlags.acceptTermsOfService {
		return slackerror.New(slackerror.ErrMissingFlag).
			WithMessage("The terms of service must be accepted to sign up for the Developer Program").
			WithRemediation("Include the %s flag to agree to the following terms and acknowledge the privacy policies:\n%s",
				style.Highlight("--accept-tos"),
				strings.Join(termsOfServiceLinks(), "\n"),
			)
	}

	auth, profile, err := selectAuthForSignup(ctx, clients)
	if err != nil {
		return err
	}

	// Always show the terms, but only prompt for acceptance if the --accept-tos flag is not set
	printTermsOfService(ctx, clients)
	if !signupCmdFlags.acceptTermsOfService {
		if err := promptTermsAcceptance(ctx, clients, cmd.Flag("accept-tos")); err != nil {
			return err
		}
	}

	email := prefillUserInfoFromAuth(ctx, clients, auth, profile, &signupCmdFlags)

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
		AcceptTermsOfService:   true, // Terms were accepted via the --accept-tos flag or prompt
	}

	printSignupDetails(ctx, clients, info, signupCmdFlags.newsletterSubscription)

	_, err = apiClientForAuth(clients, auth).Signup(ctx, auth.Token, req)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji: "tada",
		Text:  "Welcome to the Slack Developer Program!",
		Secondary: []string{
			fmt.Sprintf("Manage your profile details at %s", style.LinkText(developerProfileSettingsURL)),
		},
	}))

	clients.IO.PrintInfo(ctx, false, "Run %s to create a developer sandbox", style.Commandf("sandbox create", false))
	return nil
}

// handlePromptError sets the cancel exit code when a prompt is interrupted, such as with Ctrl+C
func handlePromptError(clients *shared.ClientFactory, err error) error {
	if slackerror.Is(err, slackerror.ErrProcessInterrupted) {
		clients.IO.SetExitCode(iostreams.ExitCancel)
	}
	return err
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

// selectAuthForSignup returns the auth to link to the Developer Program account and, when it
// was already fetched for the prompt, the profile of that user.
// Uses the global --token or --team flag if present, otherwise prompts the user to select an account.
func selectAuthForSignup(ctx context.Context, clients *shared.ClientFactory) (*types.SlackAuth, *types.UserInfo, error) {
	if clients.Config.TokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return nil, nil, err
		}
		return &auth, nil, nil
	}

	allAuths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return nil, nil, err
	}
	if len(allAuths) == 0 {
		return nil, nil, slackerror.New(slackerror.ErrNotAuthed).
			WithMessage("You must be logged in to sign up for the Developer Program").
			WithRemediation("Run %s to sign in to a workspace, then try again", style.Commandf("login", false))
	}

	// Match the --team flag without fetching profiles for the other logins
	if clients.Config.TeamFlag != "" {
		for i, auth := range allAuths {
			if clients.Config.TeamFlag == auth.TeamID || clients.Config.TeamFlag == auth.TeamDomain {
				return &allAuths[i], nil, nil
			}
		}
		return nil, nil, slackerror.New(slackerror.ErrCredentialsNotFound)
	}

	teamFlag := clients.Config.Flags.Lookup("team")
	labels := make([]string, 0, len(allAuths))
	options := make([]iostreams.PromptOption, 0, len(allAuths))
	profiles := make([]*types.UserInfo, len(allAuths))
	for i, auth := range allAuths {
		label := style.TeamSelectLabel(auth.TeamDomain, auth.TeamID)
		if userInfo, err := apiClientForAuth(clients, &auth).UsersInfo(ctx, auth.Token, auth.UserID); err == nil {
			profiles[i] = userInfo
			if userInfo.Profile.Email != "" {
				label = fmt.Sprintf("%s %s", userInfo.Profile.Email, style.Faint(fmt.Sprintf("(%s %s)", auth.TeamDomain, auth.TeamID)))
			}
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
		Help:     "The selected login will be used to fetch your user details and create your account",
	})
	if err != nil {
		return nil, nil, handlePromptError(clients, err)
	}

	if selection.Prompt {
		return &allAuths[selection.Index], profiles[selection.Index], nil
	}
	for i, auth := range allAuths {
		if selection.Option == auth.TeamID || selection.Option == auth.TeamDomain {
			return &allAuths[i], profiles[i], nil
		}
	}
	return nil, nil, slackerror.New(slackerror.ErrCredentialsNotFound)
}

// prefillUserInfoFromAuth fills in any unset name flags from the user's Slack profile.
// The profile is fetched when it wasn't already fetched while selecting the auth.
// Returns the email address of the profile, if available.
func prefillUserInfoFromAuth(ctx context.Context, clients *shared.ClientFactory, auth *types.SlackAuth, userInfo *types.UserInfo, flags *signupFlags) string {
	if userInfo == nil {
		if auth.UserID == "" {
			return ""
		}
		var err error
		userInfo, err = apiClientForAuth(clients, auth).UsersInfo(ctx, auth.Token, auth.UserID)
		if err != nil {
			return ""
		}
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

// printTermsOfService shows the terms and policies accepted at signup
func printTermsOfService(ctx context.Context, clients *shared.ClientFactory) {
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "scroll",
		Text:  "Terms of service",
		Secondary: append(
			[]string{"By signing up, you agree to the following terms and acknowledge the privacy policies:"},
			termsOfServiceLinks()...,
		),
	}))
}

// promptTermsAcceptance prompts the user to accept the terms.
// The acceptTermsFlag is suggested when prompts aren't possible.
func promptTermsAcceptance(ctx context.Context, clients *shared.ClientFactory, acceptTermsFlag *pflag.Flag) error {
	acceptTerms, err := clients.IO.ConfirmPrompt(ctx, "Do you agree to these terms?", iostreams.ConfirmPromptConfig{Default: false, Flag: acceptTermsFlag})
	if err != nil {
		return handlePromptError(clients, err)
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
// Postal code and country are optional flags.
func collectUserInfo(ctx context.Context, clients *shared.ClientFactory, flags *signupFlags) (*userInfo, error) {
	var err error
	info := &userInfo{
		firstName:     flags.firstName,
		lastName:      flags.lastName,
		postalCode:    strings.TrimSpace(flags.postalCode),
		countryOption: strings.ToUpper(strings.TrimSpace(flags.countryOption)),
	}

	// Name the missing flags when run from a non-interactive terminal
	if !clients.IO.IsTTY() && (info.firstName == "" || info.lastName == "") {
		return nil, slackerror.New(slackerror.ErrMissingFlag).
			WithMessage("A first and last name are required to sign up").
			WithRemediation("Include the %s and %s flags", style.Highlight("--first-name"), style.Highlight("--last-name"))
	}

	if info.firstName == "" {
		info.firstName, err = clients.IO.InputPrompt(ctx, "Enter your first name", iostreams.InputPromptConfig{Required: true})
		if err != nil {
			return nil, handlePromptError(clients, err)
		}
	}

	if info.lastName == "" {
		info.lastName, err = clients.IO.InputPrompt(ctx, "Enter your last name", iostreams.InputPromptConfig{Required: true})
		if err != nil {
			return nil, handlePromptError(clients, err)
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
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji:     "clipboard",
		Text:      "Developer Program Signup",
		Secondary: details,
	}))
}
