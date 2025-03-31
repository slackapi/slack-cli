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
	"fmt"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/auth"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Create handle to the function for testing
// TODO - Stopgap until we learn the correct way to structure our code for testing.
var listFunc = auth.List

// NewListCommand creates the Cobra command for listing authorized accounts
func NewListCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all authorized accounts",
		Long:  "List all authorized accounts",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "auth list", Meaning: "List all authorized accounts"},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}
}

// runListCommand will execute the list command
func runListCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	log := newListLogger(cmd, clients.IO)
	userAuthList, err := listFunc(cmd.Context(), clients, log)
	if err != nil {
		return err
	}
	printAuthListSuccess(cmd, clients.IO, userAuthList)
	return nil
}

// newListLogger creates a logger instance to receive event notifications
func newListLogger(cmd *cobra.Command, IO iostreams.IOStreamer) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_auth_list":
				userAuthList := []types.SlackAuth{}
				if event.Data["userAuthList"] != nil {
					userAuthList = event.Data["userAuthList"].([]types.SlackAuth)
				}
				printAuthList(cmd, IO, userAuthList)
			default:
				// Ignore the event
			}
		},
	)
}

// printAuthList will display a list of all authorizations available and highlight the default.
// The output will be a list formatted as:
//
// trashpanda-dev (Team ID: T01TY782HVV)
// User ID: U01PHTW5JFR
// API Host: https://dev.slack.com (optional, only shown for custom API Hosts)
// Last Updated: 2021-03-12 11:18:00 -0700
func printAuthList(cmd *cobra.Command, IO iostreams.IOStreamer, userAuthList []types.SlackAuth) {
	// Based on loosely on time.RFC3339
	timeFormat := "2006-01-02 15:04:05 Z07:00"

	cmd.Println()
	// Display each authorization
	for _, authInfo := range userAuthList {
		cmd.Printf(
			style.Bold("%s (Team ID: %s)\n"),
			authInfo.TeamDomain,
			authInfo.TeamID,
		)
		cmd.Printf(
			style.Secondary("User ID: %s\n"),
			authInfo.UserID,
		)
		if authInfo.ApiHost != nil {
			cmd.Printf(
				style.Secondary("API Host: %s\n"),
				*authInfo.ApiHost,
			)
		}
		cmd.Printf(
			style.Secondary("Last Updated: %s\n"),
			authInfo.LastUpdated.Format(timeFormat),
		)
		caser := cases.Title(language.English)
		cmd.Printf(
			style.Secondary("Authorization Level: %s\n"),
			caser.String(authInfo.AuthLevel()),
		)

		cmd.Println()

		// Print a trace with info about the authorization
		IO.PrintTrace(cmd.Context(), slacktrace.AuthListInfo, authInfo.UserID, authInfo.TeamID)
	}

	// When there are no authorizations
	if len(userAuthList) <= 0 {
		cmd.Printf("%s", style.Secondary("You are not logged in to any Slack accounts\n\n"))
	}

	// Print a trace with the total number of authorized workspaces
	IO.PrintTrace(cmd.Context(), slacktrace.AuthListCount, fmt.Sprint(len(userAuthList)))
}

// printAuthListSuccess is displayed at the very end and helps guide the developer toward next steps.
func printAuthListSuccess(cmd *cobra.Command, IO iostreams.IOStreamer, userAuthList []types.SlackAuth) {
	commandText := style.Commandf("login", true)

	// When there are no authorizations, guide the user to creating an authorization.
	if len(userAuthList) <= 0 {
		cmd.Printf(
			"To login to a Slack account, run %s\n\n",
			commandText,
		)
	}

	IO.PrintTrace(cmd.Context(), slacktrace.AuthListSuccess)
}
