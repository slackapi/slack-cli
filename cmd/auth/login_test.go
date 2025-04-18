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
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/iostreams"
	authpkg "github.com/slackapi/slack-cli/internal/pkg/auth"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

var mockChallengeCode = "1234"

var mockOrgAuth = types.SlackAuth{
	Token:               "token",
	TeamID:              "E123",
	UserID:              "U123",
	TeamDomain:          "org",
	IsEnterpriseInstall: true,
}
var mockOrgAuthURL = "https://url.com"

func TestLoginCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors when the challenge flag is provided without the ticket flag": {
			CmdArgs:              []string{"--challenge=achallengestring"},
			ExpectedErrorStrings: []string{authpkg.InvalidNoPromptFlags},
		},
		"errors when the ticket flag is provided without the challenge flag": {
			CmdArgs:              []string{"--ticket=aticketstring"},
			ExpectedErrorStrings: []string{authpkg.InvalidNoPromptFlags},
		},
		"deprecated auth flag is noted in outputs": {
			CmdArgs:         []string{"--auth", "xoxp-example"},
			ExpectedOutputs: []string{deprecatedUserTokenMessage},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.ApiInterface.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{
					UserID:   &mockOrgAuth.UserID,
					TeamID:   &mockOrgAuth.TeamID,
					TeamName: &mockOrgAuth.TeamDomain,
					URL:      &mockOrgAuthURL,
				}, nil)
				cm.AuthInterface.On("IsApiHostSlackProd", mock.Anything).Return(true)
				cm.AuthInterface.On(
					"AuthWithTeamDomain",
					mock.Anything,
					mock.Anything,
				).Return(
					types.SlackAuth{},
					nil,
				)
				cm.AddDefaultMocks()
			},
		},
		"deprecated auth flag errors with the token flag": {
			CmdArgs: []string{"--auth", "xoxp-example", "--token", "xoxb-example"},
			ExpectedError: slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage("The --auth and --token flags cannot be used together. Please use --token <token>"),
		},
		"suggests creating a new app if not in a project": {
			CmdArgs:               []string{"--ticket=example", "--challenge=tictactoe"},
			ExpectedStdoutOutputs: []string{"Get started by creating a new app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.ApiInterface.On(
					"ExchangeAuthTicket",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExchangeAuthTicketResult{
						IsReady: true,
						Token:   "xoxp-example",
					},
					nil,
				)
				cm.AuthInterface.On("IsApiHostSlackProd", mock.Anything).Return(true)
				cm.AuthInterface.On(
					"SetAuth",
					mock.Anything,
					mock.Anything,
				).Return(
					types.SlackAuth{Token: "xoxp-example"},
					"",
					nil,
				)
				cm.AddDefaultMocks()
			},
		},
		"suggests listing existing apps from the project": {
			CmdArgs:               []string{"--ticket", "example", "--challenge", "tictactoe"},
			ExpectedStdoutOutputs: []string{"Review existing installations of the app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.ApiInterface.On(
					"ExchangeAuthTicket",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					api.ExchangeAuthTicketResult{
						IsReady: true,
						Token:   "xoxp-example",
					},
					nil,
				)
				cm.AuthInterface.On("IsApiHostSlackProd", mock.Anything).Return(true)
				cm.AuthInterface.On(
					"SetAuth",
					mock.Anything,
					mock.Anything,
				).Return(
					types.SlackAuth{Token: "xoxp-example"},
					"",
					nil,
				)
				cm.AddDefaultMocks()
				cf.SDKConfig.WorkingDirectory = "."
			},
		},
		"happy path login with prompt flow should pass challenge code to ExchangeAuthTicket API": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.ApiInterface.On("GenerateAuthTicket", mock.Anything, mock.Anything, mock.Anything).Return(api.GenerateAuthTicketResult{}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter challenge code", iostreams.InputPromptConfig{
					Required: true,
				}).Return(mockChallengeCode, nil)
				cm.ApiInterface.On("ExchangeAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.ExchangeAuthTicketResult{}, nil)
				cm.AuthInterface.On("IsApiHostSlackProd", mock.Anything).Return(true)
				cm.AuthInterface.On(
					"SetAuth",
					mock.Anything,
					mock.Anything,
				).Return(
					types.SlackAuth{},
					"",
					nil,
				)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertCalled(t, "ExchangeAuthTicket", mock.Anything, mock.Anything, mockChallengeCode, mock.Anything)
			},
		},
		"should explode if ExchangeAuthTicket API fails": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.ApiInterface.On("GenerateAuthTicket", mock.Anything, mock.Anything, mock.Anything).Return(api.GenerateAuthTicketResult{}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter challenge code", iostreams.InputPromptConfig{
					Required: true,
				}).Return(mockChallengeCode, nil)
				cm.ApiInterface.On("ExchangeAuthTicket", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.ExchangeAuthTicketResult{}, slackerror.New(slackerror.ErrHttpResponseInvalid))
			},
			ExpectedError: slackerror.New(slackerror.ErrHttpResponseInvalid),
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewLoginCommand(cf)
	})
}
