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
	"encoding/json"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/datastore"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var BulkGet = datastore.BulkGet

func NewBulkGetCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-get <expression> [flags]",
		Short: "Get multiple items from a datastore",
		Long: strings.Join([]string{
			"Get multiple items from a datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Get two items from datastore",
				Command: `datastore bulk-get --datastore tasks '{"ids": ["12", "42"]}'`,
			},
			{
				Meaning: "Get two items from datastore with an expression",
				Command: `datastore bulk-get '{"datastore": "tasks", "ids": ["12", "42"]}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunBulkGetCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastoreBulkGet

			// TODO: almost all the code below here except for the actual API call is identical across all datastore commands. can we DRY this up / is it worth it?
			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "bulk-get")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("bulk-get", true))
			}

			// Get the app auth selection from the flag or prompt
			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}
			ctx = config.SetContextToken(ctx, selection.Auth.Token)

			// Set the app ID from the selected workspace
			query.App = selection.App.AppID

			// Optionally display the JSON expression and exit
			if showExpressionFlag {
				return printDatastoreExpressionMarshal(ctx, clients, query)
			}

			// Perform the get
			log := newBulkGetLogger(clients, cmd, query)
			event, err := BulkGet(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastoreBulkGetSuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().StringVar(&outputFlag, "output", "text", outputUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)
	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)

	return cmd
}

// preRunBulkGetCommandFunc determines if the command is supported for a project
// and configures flags
func preRunBulkGetCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return err
	}
	if clients.Config.ForceFlag {
		return nil
	}
	return cmdutil.IsSlackHostedProject(ctx, clients)
}

func newBulkGetLogger(clients *shared.ClientFactory, cmd *cobra.Command, request types.AppDatastoreBulkGet) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_bulk_get_result":
				getResult := types.AppDatastoreBulkGetResult{}
				if event.Data["bulkGetResult"] != nil {
					getResult = event.Data["bulkGetResult"].(types.AppDatastoreBulkGetResult)
				}
				if cmd != nil {
					// TODO: this can raise an error on indentation failures, but not sure how to handle that using logger
					_ = printBulkGetResult(clients, cmd, request, getResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printBulkGetResult(clients *shared.ClientFactory, cmd *cobra.Command, request types.AppDatastoreBulkGet, getResult types.AppDatastoreBulkGetResult) error {
	var datastore = getResult.Datastore
	var items = getResult.Items

	missingIDsMessage := ""
	if len(request.IDs) != len(items)+len(getResult.FailedItems) {
		missingIDsMessage = " Not all IDs were found"
	}

	if outputFlag == "text" {
		cmd.Printf(
			style.Bold("%s Get from Datastore: %s.%s\n\n"),
			style.Emoji("tada"),
			datastore,
			missingIDsMessage,
		)
	}

	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return slackerror.New("Error during output indentation").WithRootCause(err)
	}

	cmd.Printf(
		"%s\n\n",
		string(b),
	)

	var failedItems = getResult.FailedItems
	if len(failedItems) > 0 {
		cmd.Printf(
			style.Bold("%s Some items failed to be retrieved and should be retried: \n\n"),
			style.Emoji("warning"),
		)

		b, err := goutils.JSONMarshalUnescapedIndent(failedItems)
		if err != nil {
			return slackerror.New("Error during output indentation").WithRootCause(err)
		}
		cmd.Printf(
			"%s\n",
			string(b),
		)
	}
	return nil
}

func printDatastoreBulkGetSuccess(cmd *cobra.Command, event *logger.LogEvent) {
}
