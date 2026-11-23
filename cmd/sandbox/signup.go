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

// Links to the terms and policies accepted when signing up for the Developer Program
const (
	developerProgramAgreementURL  = "https://slack.com/terms-of-service/slack-developer-program"
	slackAPITermsOfServiceURL     = "https://slack.com/terms-of-service/api"
	slackPrivacyPolicyURL         = "https://slack.com/trust/privacy/privacy-policy"
	salesforcePrivacyStatementURL = "https://www.salesforce.com/company/legal/privacy/"
)

// developerProfileSettingsURL is where developers can view or update their user data for their Developer Program profile
const developerProfileSettingsURL = "https://api.slack.com/developer-program/settings/profile"

// developerProgramTermsOfServiceVersion is the current version of the terms of service links
const developerProgramTermsOfServiceVersion = "may2024"

type signupFlags struct {
	firstName                   string
	lastName                    string
	postalCode                  string
	countryOption               string
	newsletterSubscription      bool
	acceptTermsOfServiceVersion string
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
			{Command: "sandbox signup --team T0123456 --postal-code 90210 --country US --newsletter", Meaning: "Sign up, authenticated as your user on the given team with the provided details"},
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
	cmd.Flags().StringVar(&signupCmdFlags.acceptTermsOfServiceVersion, "accept-terms", "", acceptTermsOfServiceUsage())
	// The flag without a value accepts the version of the terms that this CLI knows
	cmd.Flag("accept-terms").NoOptDefVal = developerProgramTermsOfServiceVersion

	// Rebuild the flag usage when help is shown so the links are styled once styles are toggled
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		style.ToggleStyles(clients.IO.IsTTY() && !clients.Config.NoColor)
		cmd.Flag("accept-terms").Usage = acceptTermsOfServiceUsage()
		cmd.Parent().HelpFunc()(cmd, args)
	})

	return cmd
}

// acceptTermsOfServiceUsage returns the usage text for the --accept-terms flag
func acceptTermsOfServiceUsage() string {
	return "Agree to the terms and acknowledge the privacy policies:\n" +
		"  " + style.LinkText(developerProgramAgreementURL) + "\n" +
		"  " + style.LinkText(slackAPITermsOfServiceURL) + "\n" +
		"  " + style.LinkText(slackPrivacyPolicyURL) + "\n" +
		"  " + style.LinkText(salesforcePrivacyStatementURL) + "\n"
}

// termsOfServiceLinks returns a bulleted list of the terms and policies accepted at signup.
// The terms and policies from the API are preferred, with known links used as a fallback.
func termsOfServiceLinks(policies []api.TermsAndPolicy) []string {
	if len(policies) > 0 {
		links := make([]string, 0, len(policies))
		for _, policy := range policies {
			links = append(links, fmt.Sprintf("• %s: %s", policy.Name, style.LinkText(policy.URL)))
		}
		return links
	}
	return []string{
		fmt.Sprintf("• Developer Program Agreement: %s", style.LinkText(developerProgramAgreementURL)),
		fmt.Sprintf("• Slack API Terms of Service: %s", style.LinkText(slackAPITermsOfServiceURL)),
		fmt.Sprintf("• Slack Privacy Policy: %s", style.LinkText(slackPrivacyPolicyURL)),
		fmt.Sprintf("• Salesforce Privacy Statement: %s", style.LinkText(salesforcePrivacyStatementURL)),
	}
}

func runSignupCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	// A version passed with --accept-terms is sent as is, so a newer version can be accepted if this
	// CLI is outdated. The flag without a version sets the value to developerProgramTermsOfServiceVersion.
	acceptedVersion := strings.TrimSpace(signupCmdFlags.acceptTermsOfServiceVersion)
	if cmd.Flag("accept-terms").Changed && acceptedVersion == "" {
		acceptedVersion = developerProgramTermsOfServiceVersion
	}

	auth, profile, err := selectAuthForSignup(ctx, clients)
	if err != nil {
		return err
	}

	email := prefillUserInfoFromAuth(ctx, clients, auth, profile, &signupCmdFlags)

	info, err := collectUserInfo(ctx, clients, &signupCmdFlags)
	if err != nil {
		return err
	}
	info.email = email

	req := api.SignupRequest{
		FirstName:                   info.firstName,
		LastName:                    info.lastName,
		PostalCode:                  info.postalCode,
		CountryOption:               info.countryOption,
		NewsletterSubscription:      signupCmdFlags.newsletterSubscription,
		AcceptTermsOfServiceVersion: acceptedVersion,
	}

	printSignupDetails(ctx, clients, info, signupCmdFlags.newsletterSubscription)

	apiClient := apiClientForAuth(clients, auth)

	// Terms that were accepted before are used when no version is included in the request,
	// otherwise the current terms are returned to be accepted before signing up again
	resp, err := apiClient.Signup(ctx, auth.Token, req)
	if err != nil && slackerror.Is(err, slackerror.ErrTosNotAccepted) {
		termsVersion := developerProgramTermsOfServiceVersion
		var policies []api.TermsAndPolicy
		if resp != nil {
			if resp.TermsOfServiceVersion != "" {
				termsVersion = resp.TermsOfServiceVersion
			}
			policies = resp.TermsAndPolicies
		}
		version, termsErr := acceptTermsOfService(ctx, clients, termsVersion, policies, acceptedVersion)
		if termsErr != nil {
			return termsErr
		}
		req.AcceptTermsOfServiceVersion = version
		_, err = apiClient.Signup(ctx, auth.Token, req)
	}
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji: "tada",
		Text:  "Welcome to the Slack Developer Program!",
		Secondary: []string{
			fmt.Sprintf("Manage your profile at %s", style.LinkText(developerProfileSettingsURL)),
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

	// If only one login is present, default to that auth without prompting
	if len(allAuths) == 1 {
		return &allAuths[0], nil, nil
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

// printTermsOfService shows the terms and policies that must be accepted at signup
func printTermsOfService(ctx context.Context, clients *shared.ClientFactory, policies []api.TermsAndPolicy) {
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji: "scroll",
		Text:  "Terms of service",
		Secondary: append(
			[]string{"By signing up, you agree to the following terms and acknowledge the privacy policies:"},
			termsOfServiceLinks(policies)...,
		),
	}))
}

// acceptTermsOfService confirms agreement to the current version of the terms of service and returns
// the version that was accepted. A version that was already provided with the --accept-terms flag is
// outdated if the terms must still be accepted, so it errors instead of agreeing to terms that changed.
func acceptTermsOfService(ctx context.Context, clients *shared.ClientFactory, version string, policies []api.TermsAndPolicy, flagVersion string) (string, error) {
	acceptTermsRemediation := fmt.Sprintf("Include %s to agree to the following terms and acknowledge the privacy policies:\n%s",
		style.Highlight(fmt.Sprintf("--accept-terms=%s", version)),
		strings.Join(termsOfServiceLinks(policies), "\n"),
	)

	if flagVersion != "" {
		return "", slackerror.New(slackerror.ErrTosNotAccepted).
			WithMessage("The accepted terms of service version %s does not match the current version %s", flagVersion, version).
			WithRemediation("%s", acceptTermsRemediation)
	}

	// The --force flag skips prompts but never accepts the terms on its own
	if clients.Config.ForceFlag || !clients.IO.IsTTY() {
		return "", slackerror.New(slackerror.ErrMissingFlag).
			WithMessage("The terms of service must be accepted to sign up for the Developer Program").
			WithRemediation("%s", acceptTermsRemediation)
	}

	printTermsOfService(ctx, clients, policies)
	acceptTerms, err := clients.IO.ConfirmPrompt(ctx, "Do you agree to these terms?", iostreams.ConfirmPromptConfig{Default: false})
	if err != nil {
		return "", handlePromptError(clients, err)
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
		return "", slackerror.New(slackerror.ErrProcessInterrupted)
	}
	return version, nil
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
	newsletterStatus := "opted out"
	if newsletter {
		newsletterStatus = "opted in"
	}
	details = append(details, fmt.Sprintf("Newsletter: %s", newsletterStatus))
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji:     "clipboard",
		Text:      "Developer Program Signup",
		Secondary: details,
	}))
}
