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

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestLogoutCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"competing flags errors": {
			CmdArgs:       []string{"--team", "team1", "--all"},
			ExpectedError: slackerror.New(slackerror.ErrMismatchedFlags),
		},
		"logout of all teams": {
			CmdArgs: []string{"--all"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api.slack.com")
				clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("logstash.slack.com")
				clientsMock.Auth.On("Auths", mock.Anything).Return(fakeAuthsByTeamSlice, nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].Token).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].RefreshToken).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].Token).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].RefreshToken).Return(nil)
				clientsMock.Auth.On("DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[0]).Return(types.SlackAuth{}, nil)
				clientsMock.Auth.On("DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[1]).Return(types.SlackAuth{}, nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked for all teams"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].Token)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].RefreshToken)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].Token)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].RefreshToken)
				clients.Auth.AssertCalled(t, "DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[0])
				clients.Auth.AssertCalled(t, "DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[1])
				clients.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"logout of single team by named domain via flag": {
			CmdArgs: []string{"--team", "team1"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api.slack.com")
				clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("logstash.slack.com")
				clientsMock.Auth.On("Auths", mock.Anything).Return(fakeAuthsByTeamSlice, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("team"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "team1",
				}, nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].Token).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].RefreshToken).Return(nil)
				clientsMock.Auth.On("DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[0]).Return(types.SlackAuth{}, nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked for team1"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].Token)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].RefreshToken)
				clients.Auth.AssertCalled(t, "DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[0])
				clients.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"logout of single team by id via flag": {
			CmdArgs: []string{"--team", "T2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api.slack.com")
				clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("logstash.slack.com")
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("team"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "T2",
				}, nil)
				clientsMock.Auth.On("Auths", mock.Anything).Return(fakeAuthsByTeamSlice, nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].Token).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].RefreshToken).Return(nil)
				clientsMock.Auth.On("DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[1]).Return(types.SlackAuth{}, nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked for team2"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].Token)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].RefreshToken)
				clients.Auth.AssertCalled(t, "DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[1])
				clients.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"require a team value with the flag": {
			CmdArgs:              []string{"--team", ""},
			ExpectedErrorStrings: []string{"The argument is missing from the --team flag (missing_flag)"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertNotCalled(t, "RevokeToken")
				clients.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select an workspace authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"require a known team value is used in flag": {
			CmdArgs: []string{"--team", "randomteamdomain"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("Auths", mock.Anything).Return(fakeAuthsByTeamSlice, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("team"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "randomteamdomain",
				}, nil)
			},
			ExpectedErrorStrings: []string{"invalid_auth", "Cannot revoke authentication tokens for 'randomteamdomain'"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertNotCalled(t, "RevokeToken")
				clients.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"logout of a workspace by prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api.slack.com")
				clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("logstash.slack.com")
				clientsMock.Auth.On("Auths", mock.Anything).Return(fakeAuthsByTeamSlice, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an authorization to revoke", []string{
					FormatAuthLabel(fakeAuthsByTeamSlice[0]),
					FormatAuthLabel(fakeAuthsByTeamSlice[1]),
				}, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("team"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: fakeAuthsByTeamSlice[1].TeamDomain,
					Index:  1,
				}, nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].Token).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].RefreshToken).Return(nil)
				clientsMock.Auth.On("DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[1]).Return(types.SlackAuth{}, nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked for team2"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].Token)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[1].RefreshToken)
				clients.Auth.AssertCalled(t, "DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[1])
				clients.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"automatically logout of the only available workspace available": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api.slack.com")
				clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("logstash.slack.com")
				clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					fakeAuthsByTeamSlice[0],
				}, nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].Token).Return(nil)
				clientsMock.Auth.On("RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].RefreshToken).Return(nil)
				clientsMock.Auth.On("DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[0]).Return(types.SlackAuth{}, nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked for team1"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].Token)
				clients.Auth.AssertCalled(t, "RevokeToken", mock.Anything, fakeAuthsByTeamSlice[0].RefreshToken)
				clients.Auth.AssertCalled(t, "DeleteAuth", mock.Anything, fakeAuthsByTeamSlice[0])
				clients.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"verify the only available auth matches a team flag": {
			CmdArgs: []string{"--team", "anotherteam"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					fakeAuthsByTeamSlice[0],
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("team"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "anotherteam",
				}, nil)
			},
			ExpectedErrorStrings: []string{"invalid_auth", "Cannot revoke authentication tokens for 'anotherteam'"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertNotCalled(t, "RevokeToken")
				clients.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"confirm authorizations are revoked if none exist": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)
			},
			ExpectedOutputs: []string{"All authorizations successfully revoked"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertNotCalled(t, "RevokeToken")
				clients.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
		},
		"error if a team flag is provided without auths": {
			CmdArgs: []string{"--team", "someteam"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("team"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "someteam",
				}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clients *shared.ClientsMock) {
				clients.Auth.AssertNotCalled(t, "RevokeToken")
				clients.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an authorization to revoke", mock.Anything, mock.Anything)
			},
			ExpectedErrorStrings: []string{"Cannot revoke authentication tokens for 'someteam'"},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		return NewLogoutCommand(clients)
	})
}
