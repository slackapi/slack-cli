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
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSignupCommand(t *testing.T) {
	testToken := "xoxb-test-token"

	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors when the sandbox-signup experiment is off": {
			CmdArgs: []string{"--token", testToken, "--accept-tos"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{"experiment_required", "--experiment sandbox-signup"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Auth.AssertNotCalled(t, "AuthWithToken", mock.Anything, mock.Anything)
				cm.API.AssertNotCalled(t, "Signup", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"creates an account with details provided via flag": {
			CmdArgs: []string{
				"--token", testToken,
				"--first-name", "Ada",
				"--last-name", "Lovelace",
				"--postal-code", "94105",
				"--country", "us",
				"--newsletter",
				"--accept-tos",
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.API.On("UsersInfo", mock.Anything, testToken, "U123").Return(&types.UserInfo{
					RealName: "Someone Else",
					Profile:  types.UserProfile{Email: "ada@example.com"},
				}, nil)
				cm.API.On("Signup", mock.Anything, testToken, mock.Anything).Return(&api.SignupResponse{DeveloperID: "D123"}, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedStdoutOutputs: []string{
				"Name: Ada Lovelace",
				"Email: ada@example.com",
				"Country: US",
				"Postal Code: 94105",
				"Newsletter: true",
				"Welcome to the Slack Developer Program!",
				"sandbox create",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "Signup", mock.Anything, testToken, api.SignupRequest{
					FirstName:              "Ada",
					LastName:               "Lovelace",
					PostalCode:             "94105",
					CountryOption:          "US",
					NewsletterSubscription: true,
					AcceptTermsOfService:   true,
				})
				cm.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"prompts for auth and terms then prefills the name from the profile": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-one", TeamID: "T1", TeamDomain: "one", UserID: "U1"},
					{Token: "xoxp-two", TeamID: "T2", TeamDomain: "two", UserID: "U2"},
				}, nil)
				cm.API.On("UsersInfo", mock.Anything, "xoxp-one", "U1").Return(&types.UserInfo{
					Profile: types.UserProfile{Email: "one@example.com"},
				}, nil)
				cm.API.On("UsersInfo", mock.Anything, "xoxp-two", "U2").Return(&types.UserInfo{
					RealName: "Grace Brewster Hopper",
					Profile:  types.UserProfile{Email: "grace@example.com"},
				}, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Select an email for your developer program account", []string{
					"one@example.com " + style.Secondary("(one T1)"),
					"grace@example.com " + style.Secondary("(two T2)"),
				}, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 1}, nil)
				cm.IO.On("ConfirmPrompt", mock.Anything, "Do you agree to these terms?", acceptTermsPromptConfig).Return(true, nil)
				cm.API.On("Signup", mock.Anything, "xoxp-two", mock.Anything).Return(&api.SignupResponse{DeveloperID: "D456"}, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedStdoutOutputs: []string{
				"Terms of service",
				"Name: Grace Brewster Hopper",
				"Email: grace@example.com",
				"Newsletter: false",
				"Welcome to the Slack Developer Program!",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "Signup", mock.Anything, "xoxp-two", api.SignupRequest{
					FirstName:            "Grace",
					LastName:             "Brewster Hopper",
					AcceptTermsOfService: true,
				})
				cm.API.AssertNumberOfCalls(t, "UsersInfo", 2)
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"uses the team flag without fetching profiles for other logins": {
			CmdArgs: []string{"--team", "T2", "--accept-tos", "--newsletter=false"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-one", TeamID: "T1", TeamDomain: "one", UserID: "U1"},
					{Token: "xoxp-two", TeamID: "T2", TeamDomain: "two", UserID: "U2"},
				}, nil)
				cm.API.On("UsersInfo", mock.Anything, "xoxp-two", "U2").Return(&types.UserInfo{
					RealName: "Grace Hopper",
					Profile:  types.UserProfile{Email: "grace@example.com"},
				}, nil)
				cm.API.On("Signup", mock.Anything, "xoxp-two", mock.Anything).Return(&api.SignupResponse{}, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedStdoutOutputs: []string{"Name: Grace Hopper", "Email: grace@example.com"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				cm.API.AssertNotCalled(t, "UsersInfo", mock.Anything, "xoxp-one", "U1")
				cm.API.AssertNumberOfCalls(t, "UsersInfo", 1)
				cm.API.AssertCalled(t, "Signup", mock.Anything, "xoxp-two", api.SignupRequest{
					FirstName:            "Grace",
					LastName:             "Hopper",
					AcceptTermsOfService: true,
				})
			},
		},
		"errors when the team flag does not match a login": {
			CmdArgs: []string{"--team", "T9", "--accept-tos"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{Token: "xoxp-one", TeamID: "T1", TeamDomain: "one", UserID: "U1"},
				}, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedErrorStrings: []string{slackerror.ErrCredentialsNotFound},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "UsersInfo", mock.Anything, mock.Anything, mock.Anything)
				cm.API.AssertNotCalled(t, "Signup", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"prompts for name when we cannot fetch it from the user's profile": {
			CmdArgs: []string{"--token", testToken, "--accept-tos"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.API.On("UsersInfo", mock.Anything, testToken, "U123").Return(&types.UserInfo{}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter your first name", mock.Anything).Return("Alan", nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter your last name", mock.Anything).Return("Turing", nil)
				cm.API.On("Signup", mock.Anything, testToken, mock.Anything).Return(&api.SignupResponse{}, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedStdoutOutputs: []string{"Name: Alan Turing"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "Signup", mock.Anything, testToken, api.SignupRequest{
					FirstName:            "Alan",
					LastName:             "Turing",
					AcceptTermsOfService: true,
				})
			},
		},
		"cancels when the terms are declined": {
			CmdArgs: []string{"--token", testToken},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.IO.On("ConfirmPrompt", mock.Anything, "Do you agree to these terms?", acceptTermsPromptConfig).Return(false, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedError: slackerror.New(slackerror.ErrProcessInterrupted),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNotCalled(t, "Signup", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"errors when the force flag is used without accepting the terms": {
			CmdArgs: []string{"--token", testToken, "--force"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedErrorStrings: []string{
				"missing_flag",
				"The terms of service must be accepted",
				"--accept-tos",
				developerProgramAgreementURL,
				slackPrivacyPolicyURL,
				salesforcePrivacyStatementURL,
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
				cm.API.AssertNotCalled(t, "Signup", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"signs up without prompts when the force and accept-tos flags are used": {
			CmdArgs: []string{"--token", testToken, "--first-name", "Ada", "--last-name", "Lovelace", "--force", "--accept-tos"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsTTY").Return(true)
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.API.On("UsersInfo", mock.Anything, testToken, "U123").Return(&types.UserInfo{}, nil)
				cm.API.On("Signup", mock.Anything, testToken, mock.Anything).Return(&api.SignupResponse{}, nil)
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedStdoutOutputs: []string{"Welcome to the Slack Developer Program!"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertCalled(t, "Signup", mock.Anything, testToken, api.SignupRequest{
					FirstName:            "Ada",
					LastName:             "Lovelace",
					AcceptTermsOfService: true,
				})
				cm.IO.AssertNotCalled(t, "ConfirmPrompt", mock.Anything, mock.Anything, mock.Anything)
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"returns an API error": {
			CmdArgs: []string{"--token", testToken, "--first-name", "Ada", "--last-name", "Lovelace", "--accept-tos", "--newsletter=false"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken, UserID: "U123"}, nil)
				cm.API.On("UsersInfo", mock.Anything, testToken, "U123").Return(&types.UserInfo{}, nil)
				cm.API.On("Signup", mock.Anything, testToken, mock.Anything).
					Return(nil, slackerror.NewAPIError(slackerror.ErrEmailAlreadyInUse, "", nil, "developer.account.createVerified"))
				enableSandboxSignupExperiment(ctx, cm, cf)
			},
			ExpectedErrorStrings: []string{
				"email_already_in_use",
				"An account already exists for this email address",
				"sandbox create",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewSignupCommand(cf)
	})
}

func TestSignupCommandHidden(t *testing.T) {
	clients := shared.NewClientFactory(shared.NewClientsMock().MockClientFactory())
	assert.True(t, NewSignupCommand(clients).Hidden)
}

// acceptTermsPromptConfig matches the terms prompt config that suggests the --accept-tos flag
var acceptTermsPromptConfig = mock.MatchedBy(func(cfg iostreams.ConfirmPromptConfig) bool {
	return !cfg.Default && cfg.Flag != nil && cfg.Flag.Name == "accept-tos"
})

// enableSandboxSignupExperiment turns on the experiment that gates the signup command
func enableSandboxSignupExperiment(ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	cm.AddDefaultMocks()
	cf.Config.ExperimentsFlag = []string{string(experiment.SandboxSignup)}
	cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
}

func TestSplitName(t *testing.T) {
	tests := map[string]struct {
		input string
		first string
		last  string
	}{
		"empty":        {input: "", first: "", last: ""},
		"single name":  {input: "Cher", first: "Cher", last: ""},
		"two names":    {input: "Ada Lovelace", first: "Ada", last: "Lovelace"},
		"many names":   {input: "Grace Brewster Hopper", first: "Grace", last: "Brewster Hopper"},
		"extra spaces": {input: "  Alan   Turing ", first: "Alan", last: "Turing"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			first, last := splitName(tc.input)
			assert.Equal(t, tc.first, first)
			assert.Equal(t, tc.last, last)
		})
	}
}
