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
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewRevokeCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [flags]",
		Short: "Revoke an authentication token",
		Long:  "Revoke an authentication token",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "auth revoke --token xoxp-1-4921830...", Meaning: "Revoke a service token"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.PrintTrace(ctx, slacktrace.AuthRevokeStart)

			token, err := promptAuthToken(ctx, clients)
			if err != nil {
				return slackerror.New("Failed to collect a token to revoke").WithCode(slackerror.ErrNoTokenFound).WithRootCause(err)
			}

			if err := clients.AuthInterface().RevokeToken(ctx, token); err != nil {
				return err
			}
			printRevokeSuccess(ctx, clients)
			return nil
		},
	}
	return cmd
}

func promptAuthToken(ctx context.Context, clients *shared.ClientFactory) (string, error) {
	response, err := clients.IO.PasswordPrompt(ctx, "Enter a token to revoke", iostreams.PasswordPromptConfig{
		Required: true,
		Flag:     clients.Config.Flags.Lookup("token"),
	})
	if err != nil {
		return "", err
	}
	return response.Value, nil
}

func printRevokeSuccess(ctx context.Context, clients *shared.ClientFactory) {
	revokedAuthText := "Authorization successfully revoked"
	logoutNextSteps := []string{
		fmt.Sprintf("Login to a new team with %s", style.Commandf("login", false)),
		fmt.Sprintf("Create a new token with %s", style.Commandf("auth token", false)),
	}

	clients.IO.PrintTrace(ctx, slacktrace.AuthRevokeSuccess)
	clients.IO.PrintInfo(ctx, false, fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
		Emoji:     "wastebasket",
		Text:      revokedAuthText,
		Secondary: logoutNextSteps,
	})))
}
