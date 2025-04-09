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
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func TestRevokeCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"revoke a token passed by flag": {
			CmdArgs: []string{"--token", "xoxp-example-1234"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter a token to revoke", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("token"),
				})).Return(iostreams.PasswordPromptResponse{
					Flag:  true,
					Value: "xoxp-example-1234",
				}, nil)
				clientsMock.AuthInterface.On("RevokeToken", mock.Anything, "xoxp-example-1234").Return(nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked"},
			ExpectedAsserts: func(t *testing.T, clients *shared.ClientsMock) {
				clients.AuthInterface.AssertCalled(t, "RevokeToken", mock.Anything, "xoxp-example-1234")
			},
		},
		"require a set token value with the flag": {
			CmdArgs: []string{"--token", ""},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter a token to revoke", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("token"),
				})).Return(iostreams.PasswordPromptResponse{}, slackerror.New(slackerror.ErrMissingFlag))
			},
			ExpectedErrorStrings: []string{"Failed to collect a token to revoke", "no_token_found", "missing_flag"},
			ExpectedAsserts: func(t *testing.T, clients *shared.ClientsMock) {
				clients.AuthInterface.AssertNotCalled(t, "RevokeToken")
			},
		},
		"revoke a token input by prompt": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter a token to revoke", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("token"),
				})).Return(iostreams.PasswordPromptResponse{
					Prompt: true,
					Value:  "xoxp-example-1234",
				}, nil)
				clientsMock.AuthInterface.On("RevokeToken", mock.Anything, "xoxp-example-1234").Return(nil)
			},
			ExpectedOutputs: []string{"Authorization successfully revoked"},
			ExpectedAsserts: func(t *testing.T, clients *shared.ClientsMock) {
				clients.AuthInterface.AssertCalled(t, "RevokeToken", mock.Anything, "xoxp-example-1234")
			},
		},
		"verify errors are gracefully handled during the revoke": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter a token to revoke", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("token"),
				})).Return(iostreams.PasswordPromptResponse{
					Prompt: true,
					Value:  "xoxp-example-1234",
				}, nil)
				clientsMock.AuthInterface.On("RevokeToken", mock.Anything, "xoxp-example-1234").Return(slackerror.New(slackerror.ErrNotAuthed))
			},
			ExpectedError: slackerror.New(slackerror.ErrNotAuthed),
			ExpectedAsserts: func(t *testing.T, clients *shared.ClientsMock) {
				clients.AuthInterface.AssertCalled(t, "RevokeToken", mock.Anything, "xoxp-example-1234")
				clients.IO.AssertNotCalled(t, "PrintTrace", mock.Anything, slacktrace.AuthRevokeSuccess)
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		return NewRevokeCommand(clients)
	})
}
