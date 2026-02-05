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

package datastore

import (
	"context"
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

var BulkDelete = datastore.BulkDelete

func NewBulkDeleteCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-delete <expression> [flags]",
		Short: "Delete multiple items from a datastore",
		Long: strings.Join([]string{
			"Delete multiple items from a datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Delete two items from the datastore",
				Command: `datastore bulk-delete --datastore tasks '{"ids": ["12", "42"]}'`,
			},
			{
				Meaning: "Delete two items from the datastore with an expression",
				Command: `datastore bulk-delete '{"datastore": "tasks", "ids": ["12", "42"]}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunBulkDeleteCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastoreBulkDelete

			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "bulk-delete")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("bulk-delete", true))
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

			// Perform the delete
			log := newBulkDeleteLogger(clients, cmd)
			event, err := BulkDelete(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastoreBulkDeleteSuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)
	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)

	return cmd
}

// preRunBulkDeleteCommandFunc determines if the command is supported for a
// project and configures flags
func preRunBulkDeleteCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

func newBulkDeleteLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_bulk_delete_result":
				deleteResult := types.AppDatastoreBulkDeleteResult{}
				if event.Data["bulkDeleteResult"] != nil {
					deleteResult = event.Data["bulkDeleteResult"].(types.AppDatastoreBulkDeleteResult)
				}
				if cmd != nil {
					_ = printBulkDeleteResult(clients, cmd, deleteResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printBulkDeleteResult(clients *shared.ClientFactory, cmd *cobra.Command, deleteResult types.AppDatastoreBulkDeleteResult) error {
	var datastore = deleteResult.Datastore
	cmd.Printf(
		style.Bold("%s Deleted from datastore: %s\n\n"),
		style.Emoji("tada"),
		datastore,
	)

	var failedItems = deleteResult.FailedItems
	if len(failedItems) > 0 {
		cmd.Printf(
			style.Bold("%s Some items failed to be deleted and should be retried: \n\n"),
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

func printDatastoreBulkDeleteSuccess(cmd *cobra.Command, event *logger.LogEvent) {
	commandText := style.Commandf("datastore query <expression>", true)
	if cmd != nil {
		cmd.Printf(
			"To inspect the datastore after updates, run %s\n",
			commandText,
		)
	}
}
