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
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	maxImportBulkSize = 25
	maxImportItems    = 5000
)

var readFromFileFlag string
var readFromFileUsage = "store multiple items from a file of JSON Lines"

var BulkPut = datastore.BulkPut
var importProgressSpinner *style.Spinner

func NewBulkPutCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-put <expression> [flags]",
		Short: "Create or replace a list of items in a datastore",
		Long: strings.Join([]string{
			"Create or replace a list of items in a datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Create or replace two new entries in the datastore",
				Command: `datastore bulk-put --datastore tasks '{"items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'`,
			},
			{
				Meaning: "Create or replace two new entries in the datastore with an expression",
				Command: `datastore bulk-put '{"datastore": "tasks", "items": [{"id": "12", "description": "Create a PR", "status": "Done"}, {"id": "42", "description": "Approve a PR", "status": "Pending"}]}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunBulkPutCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastoreBulkPut

			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "bulk-put")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("bulk-put", true))
			}

			// Get the selection from the flag or prompt
			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
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

			if readFromFileFlag != "" {
				if len(query.Items) != 0 {
					return slackerror.New(slackerror.ErrMismatchedFlags).
						WithMessage("items field for --to-file cannot be specified")
				}

				// primary key is needed to match failed items returned by bulkPut API to items being sent in the request body
				primaryKey, err := getPrimaryKey(ctx, clients, selection.App, selection.Auth, query.Datastore)
				if err != nil {
					return err
				}
				return startBulkPutImport(ctx, clients, cmd, query, primaryKey)
			}

			// Perform the put
			log := newBulkPutLogger(clients, cmd)
			event, err := BulkPut(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastoreBulkPutSuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)
	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)
	cmd.Flags().StringVar(&readFromFileFlag, "from-file", "", readFromFileUsage)

	return cmd
}

// preRunBulkPutCommandFunc determines if the command is supported for a project
// and configures flags
func preRunBulkPutCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

func newBulkPutLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_bulk_put_result":
				bulkPutResult := types.AppDatastoreBulkPutResult{}
				if event.Data["bulkPutResult"] != nil {
					bulkPutResult = event.Data["bulkPutResult"].(types.AppDatastoreBulkPutResult)
				}
				if cmd != nil {
					// TODO: this can raise an error on indentation failures, but not sure how to handle that using logger
					_ = printBulkPutResult(clients, cmd, bulkPutResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printBulkPutResult(clients *shared.ClientFactory, cmd *cobra.Command, putResult types.AppDatastoreBulkPutResult) error {
	var datastore = putResult.Datastore
	cmd.Printf(
		style.Bold("%s Stored items in the datastore: %s\n\n"),
		style.Emoji("tada"),
		datastore,
	)

	var failed_items = putResult.FailedItems
	if len(failed_items) > 0 {
		cmd.Printf(
			style.Bold("%s Some items failed to be inserted and should be retried: \n\n"),
			style.Emoji("warning"),
		)

		b, err := goutils.JsonMarshalUnescapedIndent(failed_items)
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

func printDatastoreBulkPutSuccess(cmd *cobra.Command, event *logger.LogEvent) {
	commandText := style.Commandf("datastore query <expression>", true)
	if cmd != nil {
		cmd.Printf(
			"To inspect the datastore after updates, run %s\n",
			commandText,
		)
	}
}

func startBulkPutImport(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, query types.AppDatastoreBulkPut, primaryKey string) error {
	currentBatch := []map[string]interface{}{}
	totalSuccessfulItems := 0
	totalFailedItems := 0
	totalPutItems := 0

	itemsFile, err := clients.Fs.Open(readFromFileFlag)
	if err != nil {
		return err
	}
	defer itemsFile.Close()
	scanner := bufio.NewScanner(itemsFile)

	logFolder, err := clients.Config.SystemConfig.LogsDir(ctx)
	if err != nil {
		return err
	}
	var currentTime = time.Now().UTC()
	var filename = "slack-bulk-put-errors-" + currentTime.Format("20060102150405") + ".log"
	errorLogFilePath := filepath.Join(logFolder, filename)
	errorLogFile, err := clients.Fs.Create(errorLogFilePath)
	if err != nil {
		return err
	}
	defer errorLogFile.Close()

	itemsFilePath, err := filepath.Abs(itemsFile.Name())
	if err != nil {
		return err
	}

	token := config.GetContextToken(ctx)

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "file_cabinet",
		Text:  "Importing datastore items from a file",
		Secondary: []string{
			fmt.Sprintf("Items will be read from %s", style.HomePath(itemsFilePath)),
		},
	}))

	importProgressSpinner = style.NewSpinner(cmd.OutOrStdout())
	defer importProgressSpinner.Stop()

	for {
		totalPutItems = totalSuccessfulItems + totalFailedItems
		update := fmt.Sprintf("Imported (%d) items. So far (%d) items failed to be imported. Total processed items is (%d).", totalSuccessfulItems, totalFailedItems, totalPutItems)

		importProgressSpinner.Update(update, "").Start()

		maxItemsToRead := min(maxImportBulkSize, maxImportItems-totalPutItems)
		if len(currentBatch) < maxItemsToRead && scanner.Scan() {
			nextItem := scanner.Text()
			if nextItem == "" {
				continue
			}
			var parsedNextItem map[string]interface{}
			err = goutils.JsonUnmarshal([]byte(nextItem), &parsedNextItem)
			if err != nil {
				err = logBulkPutImportError(errorLogFile, nextItem, "item couldn't be parsed as JSON")
				if err != nil {
					return err
				}
				totalFailedItems++
				continue
			}
			if _, exists := parsedNextItem[primaryKey]; !exists {
				err = logBulkPutImportError(errorLogFile, nextItem, "primary key not found")
				if err != nil {
					return err
				}
				totalFailedItems++
				continue
			}

			idx := slices.IndexFunc(currentBatch, func(item map[string]interface{}) bool { return item[primaryKey] == parsedNextItem[primaryKey] })
			if idx != -1 {
				err = logBulkPutImportError(errorLogFile, nextItem, "item with the same primary key already exists")
				if err != nil {
					return err
				}
				totalFailedItems++
				continue
			}

			currentBatch = append(currentBatch, parsedNextItem)
			continue
		}

		if len(currentBatch) == 0 {
			update = fmt.Sprintf("Successfully imported (%d) items! (%d) items failed to be imported. Total processed items is (%d)", totalSuccessfulItems, totalFailedItems, totalPutItems)
			importProgressSpinner.Update(update, "tada").Stop()
			break
		}

		query.Items = currentBatch
		bulkPutResult, err := clients.ApiInterface().AppsDatastoreBulkPut(ctx, token, query)
		if err != nil {
			if len(err.(*slackerror.Error).Details) == 0 {
				return err
			}
			for _, errorDetail := range err.(*slackerror.Error).Details {
				idx := slices.IndexFunc(currentBatch, func(item map[string]interface{}) bool { return item[primaryKey] == errorDetail.Item[primaryKey] })
				if idx == -1 {
					return err //defensive coding, this shouldn't happen
				}
				currentBatch = slices.Delete(currentBatch, idx, idx+1)

				stringItem, err := goutils.JsonMarshalUnescaped(errorDetail.Item)
				if err != nil {
					return err
				}
				err = logBulkPutImportError(errorLogFile, stringItem, errorDetail.Message)
				if err != nil {
					return err
				}
				totalFailedItems++
			}
			continue
		}
		totalSuccessfulItems += len(currentBatch) - len(bulkPutResult.FailedItems)
		currentBatch = bulkPutResult.FailedItems
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if totalPutItems >= maxImportItems {
		clients.IO.PrintInfo(ctx, false, "%sImport will be limited to the first %d items in the file.", style.Emoji("warning"), maxImportItems)
	}
	if totalFailedItems != 0 {
		clients.IO.PrintInfo(ctx, false, "%sSome items failed to be imported. Check %s for more details.", style.Emoji("warning"), errorLogFilePath)
	} else {
		if err = clients.Fs.Remove(errorLogFilePath); err != nil {
			return err
		}
	}
	cmd.Println()

	return nil
}

// logBulkPutImportError saves failed item to file along with reason. Failed items are saved as strings rather than json
// because the item might not be a valid json after all
func logBulkPutImportError(file afero.File, item string, reason string) error {
	failedItem := struct {
		Item   string `json:"item"`
		Reason string `json:"reason"`
	}{
		Item:   item,
		Reason: reason,
	}
	stringFailedItem, err := goutils.JsonMarshalUnescaped(failedItem)
	if err != nil {
		return err
	}
	_, err = file.WriteString(stringFailedItem)
	return err
}
