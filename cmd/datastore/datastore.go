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

package datastore

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// appSelectPromptFunc is a handle to the AppSelectPrompt that can be mocked in tests
var appSelectPromptFunc = prompts.AppSelectPrompt

var datastoreFlag string
var datastoreUsage = "the datastore used to store items"

var outputFlag string
var outputUsage = "output format: text, json"

var showExpressionFlag bool
var showExpressionUsage = "only construct a JSON expression"

var unstableFlag bool
var unstableUsage = "kick the tires of experimental features"

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datastore <subcommand> <expression> [flags]",
		Short: "Interact with an app's datastore",
		Long: strings.Join([]string{
			"Interact with the items stored in an app's datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
			"",
			`Discover the datastores: {{LinkText "https://tools.slack.dev/deno-slack-sdk/guides/using-datastores"}}`,
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Add a new entry to the datastore",
				Command: `datastore put --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'`,
			},
			{
				Meaning: "Add two new entries to the datastore",
				Command: `datastore bulk-put --datastore tasks '{"items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'`,
			},
			{
				Meaning: "Update the entry in the datastore",
				Command: `datastore update --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'`,
			},
			{
				Meaning: "Get an item from the datastore",
				Command: `datastore get --datastore tasks '{"id": "42"}'`,
			},
			{
				Meaning: "Get two items from datastore",
				Command: `datastore bulk-get --datastore tasks '{"ids": ["12", "42"]}'`,
			},
			{
				Meaning: "Remove an item from the datastore",
				Command: `datastore delete --datastore tasks '{"id": "42"}'`,
			},
			{
				Meaning: "Remove two items from the datastore",
				Command: `datastore bulk-delete --datastore tasks '{"ids": ["12", "42"]}'`,
			},
			{
				Meaning: "Query the datastore for specific items",
				Command: `datastore query --datastore tasks '{"expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'`,
			},
			{
				Meaning: "Count number of items in datastore",
				Command: `datastore count --datastore tasks`,
			},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewPutCommand(clients))
	cmd.AddCommand(NewBulkPutCommand(clients))
	cmd.AddCommand(NewUpdateCommand(clients))
	cmd.AddCommand(NewGetCommand(clients))
	cmd.AddCommand(NewBulkGetCommand(clients))
	cmd.AddCommand(NewDeleteCommand(clients))
	cmd.AddCommand(NewBulkDeleteCommand(clients))
	cmd.AddCommand(NewQueryCommand(clients))
	cmd.AddCommand(NewCountCommand(clients))

	return cmd
}

// setQueryExpression validates the provided expression and sets query values
func setQueryExpression(clients *shared.ClientFactory, query types.Datastorer, expression string, method string) error {
	err := goutils.JsonUnmarshal([]byte(expression), query)
	if err != nil {
		return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
			WithRootCause(err).
			WithRemediation("%s", datastoreExpressionRemediation(method, false))
	}
	// FIXME: do not error on (local|deploy) flag if it matches the query.App environment
	if query.AppID() != "" && clients.Config.AppFlag != "" && clients.Config.AppFlag != query.AppID() {
		return slackerror.New(slackerror.ErrInvalidAppFlag).
			WithRemediation("Check that the app flag matches the app ID in your expression")
	} else if query.AppID() != "" {
		clients.Config.AppFlag = query.AppID()
	}
	if clients.Config.Flags.Lookup("datastore").Changed {
		if query.Name() != "" && query.Name() != clients.Config.Flags.Lookup("datastore").Value.String() {
			return slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage("Provided datastore name from flag does not match query")
		} else {
			query.SetName(clients.Config.Flags.Lookup("datastore").Value.String())
		}
	}
	return nil
}

// datastoreExpressionRemediation returns a command-specific message to display
// on invalid expressions
func datastoreExpressionRemediation(command string, isEmpty bool) string {
	var remediationStep string
	if isEmpty {
		remediationStep = "Provide a JSON expression in the command arguments"
	} else {
		remediationStep = "Verify the expression you provided is valid JSON surrounded by quotations"
	}
	helpCommand := fmt.Sprintf("datastore %s --help", command)
	showCommand := fmt.Sprintf("datastore %s --show --unstable", command)

	return strings.Join([]string{
		remediationStep,
		fmt.Sprintf("Find an example with %s", style.Commandf(helpCommand, false)),
		fmt.Sprintf("Build a new expression using %s", style.Commandf(showCommand, false)),
	}, "\n")
}

// printDatastoreExpressionMarshal displays a message with the query encoded as JSON
func printDatastoreExpressionMarshal(ctx context.Context, clients *shared.ClientFactory, query interface{}) error {
	expression, err := goutils.JSONMarshalUnescaped(query)
	if err != nil {
		return err
	}
	clients.IO.PrintInfo(ctx, false, "")
	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji: "open_file_folder",
		Text:  "This expression can be represented by the following JSON:",
	}))
	clients.IO.PrintInfo(ctx, false, style.Secondary(expression))
	return nil
}

// getPrimaryKey return the primary key of a datastore for an installed app
func getPrimaryKey(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
	datastoreName string,
) (
	string,
	error,
) {
	yaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return "", err
	}

	datastore, exists := yaml.Datastores[datastoreName]
	if !exists {
		return "", slackerror.New(slackerror.ErrDatastoreNotFound)
	}
	if datastore.PrimaryKey == "" {
		return "", slackerror.New(slackerror.ErrDatastoreMissingPrimaryKey)
	}
	return datastore.PrimaryKey, nil
}
