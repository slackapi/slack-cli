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
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/pkg/datastore"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

const (
	maxExportQueryLimit = 100
	maxExportItems      = 10000
)

var attributeFlag string
var attributeUsage = "attribute pairings for the expression"

var saveToFileFlag string
var saveToFileUsage = "save items directly to a file as JSON Lines"

var Query = datastore.Query
var exportProgressSpinner *style.Spinner

func NewQueryCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <expression> [flags]",
		Short: "Query a datastore for items",
		Long: strings.Join([]string{
			"Query a datastore for items.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Collect a limited set of items from the datastore",
				Command: `datastore query --datastore tasks '{"limit": 8}' --output json`,
			},
			{
				Meaning: "Collect items from the datastore starting at a cursor",
				Command: `datastore query --datastore tasks '{"cursor": "eyJfX2NWaV..."}'`,
			},
			{
				Meaning: "Query the datastore for specific items",
				Command: `datastore query --datastore tasks '{"expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'`,
			},
			{
				Meaning: "Query the datastore for specific items with only an expression",
				Command: `datastore query '{"datastore": "tasks", "expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunQueryCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastoreQuery

			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "query")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("query", true))
			}

			// Get the selection from the flag or prompt
			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}
			ctx = config.SetContextToken(ctx, selection.Auth.Token)

			// Build the query if it wasn't passed by argument
			if len(args) == 0 && unstableFlag {
				query, err = promptDatastoreQueryRequest(ctx, clients, selection.App, selection.Auth)
				if err != nil {
					return err
				}
			}

			// Set the app ID from the selected workspace
			query.App = selection.App.AppID

			// Optionally display the JSON expression and exit
			if showExpressionFlag {
				return printDatastoreExpressionMarshal(ctx, clients, query)
			}

			if saveToFileFlag != "" {
				return startQueryExport(ctx, clients, cmd, query)
			}

			// Perform the query
			log := newQueryLogger(clients, cmd)
			event, err := Query(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastoreQuerySuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&outputFlag, "output", "text", outputUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)

	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().StringVar(&attributeFlag, "attributes", "", attributeUsage)

	cmd.Flags().StringVar(&saveToFileFlag, "to-file", "", saveToFileUsage)

	cmd.Flag("attributes").Hidden = true // Hide while unstable is present

	return cmd
}

// preRunQueryCommandFunc determines if the command is supported for a project
// and configures flags
func preRunQueryCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	if clients.Config.Flags.Lookup("to-file").Changed &&
		clients.Config.Flags.Lookup("output").Changed {
		return slackerror.New(slackerror.ErrMismatchedFlags).
			WithMessage("Output type for --to-file cannot be specified")
	}
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return err
	}
	if clients.Config.ForceFlag {
		return nil
	}
	return cmdutil.IsSlackHostedProject(ctx, clients)
}

func newQueryLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_query_result":
				queryResult := types.AppDatastoreQueryResult{}
				if event.Data["queryResult"] != nil {
					queryResult = event.Data["queryResult"].(types.AppDatastoreQueryResult)
				}
				if cmd != nil {
					// TODO: this can raise an error on indentation failures, but not sure how to handle that using logger
					_ = printQueryResult(clients, cmd, queryResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printQueryResult(clients *shared.ClientFactory, cmd *cobra.Command, queryResult types.AppDatastoreQueryResult) error {
	var datastore = queryResult.Datastore
	switch outputFlag {
	case "text":
		var items = queryResult.Items
		cmd.Printf(
			style.Bold("%s Retrieved %d items from datastore: %s\n\n"),
			style.Emoji("tada"),
			len(items),
			datastore,
		)
		for _, item := range items {
			b, err := goutils.JSONMarshalUnescapedIndent(item)
			if err != nil {
				return slackerror.New("Error during output indentation").WithRootCause(err)
			}
			cmd.Printf(
				"%s\n",
				string(b),
			)
		}
	case "json":
		b, err := goutils.JSONMarshalUnescapedIndent(queryResult)
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

func printDatastoreQuerySuccess(cmd *cobra.Command, event *logger.LogEvent) {
	if outputFlag == "text" {
		commandText := style.Commandf("datastore put <expression>", true)
		if cmd != nil {
			cmd.Printf(
				"To create or update existing items run %s\n",
				commandText,
			)
		}
	}
}

// promptDatastoreQueryRequest constructs a datastore query expression by prompting
func promptDatastoreQueryRequest(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) (
	types.AppDatastoreQuery,
	error,
) {
	var query types.AppDatastoreQuery

	// Collect datastore information from the manifest
	yaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return types.AppDatastoreQuery{}, err
	}

	var datastores []string
	for name := range yaml.Datastores {
		datastores = append(datastores, name)
	}
	if len(datastores) <= 0 {
		return types.AppDatastoreQuery{}, slackerror.New(slackerror.ErrDatastoreNotFound).
			WithMessage("No datastores are associated with this app")
	}
	sort.Strings(datastores)

	selection, err := clients.IO.SelectPrompt(ctx, "Select a datastore", datastores, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("datastore"),
		Required: true,
	})
	if err != nil {
		return types.AppDatastoreQuery{}, err
	} else if yaml.Datastores[selection.Option].PrimaryKey == "" {
		return types.AppDatastoreQuery{}, slackerror.New(slackerror.ErrDatastoreNotFound)
	} else {
		query.Datastore = selection.Option
	}

	// Display a hint for writing expressions
	clients.IO.PrintInfo(ctx, false, "")
	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji: "bulb",
		Text:  "Expressions should use the following format",
		Secondary: []string{
			fmt.Sprintf(`Attributes begin with "#" (Example: "#%s")`, yaml.Datastores[query.Datastore].PrimaryKey),
			`Values begin with ":" (Example: ":num")`,
			fmt.Sprintf(`Example expression: #%s < :num`, yaml.Datastores[query.Datastore].PrimaryKey),
		},
	}))

	// Gather the expression and necessary information for the expression
	expression, err := clients.IO.InputPrompt(ctx, "Enter an expression", iostreams.InputPromptConfig{
		Required: false,
	})
	if err != nil {
		return types.AppDatastoreQuery{}, err
	}
	query.Expression = expression
	attributes, values := getExpressionPatterns(expression)

	fields := yaml.Datastores[query.Datastore].Attributes
	var fieldSlice []string
	for field := range fields {
		fieldSlice = append(fieldSlice, field)
	}

	if len(attributes) > 0 {
		query.ExpressionAttributes = make(map[string]interface{})
	}

	// Prompt for individual attribute selections or gather from the flag
	for _, attribute := range attributes {
		field := strings.TrimPrefix(attribute, "#")
		if _, ok := fields[field]; !ok {
			attributePrompt := fmt.Sprintf("Select an attribute for '#%s'", field)
			selection, err = clients.IO.SelectPrompt(ctx, attributePrompt, fieldSlice, iostreams.SelectPromptConfig{
				Flag:     clients.Config.Flags.Lookup("attributes"),
				Required: false,
			})
			if err != nil {
				return types.AppDatastoreQuery{}, err
			} else if selection.Flag {
				if attributes, err := mapAttributeFlag(selection.Option); err != nil {
					return types.AppDatastoreQuery{}, err
				} else {
					query.ExpressionAttributes = attributes
					break
				}
			} else if selection.Prompt {
				query.ExpressionAttributes[attribute] = selection.Option
			}
		} else {
			query.ExpressionAttributes[attribute] = field
		}
	}

	if len(values) > 0 {
		query.ExpressionValues = make(map[string]interface{})
	}
	for _, value := range values {
		field := strings.TrimPrefix(value, ":")
		valuePrompt := fmt.Sprintf("Enter a value for ':%s'", field)
		field, err = clients.IO.InputPrompt(ctx, valuePrompt, iostreams.InputPromptConfig{
			Required: true,
		})
		if err != nil {
			return types.AppDatastoreQuery{}, err
		}
		query.ExpressionValues[value] = field
	}

	return query, nil
}

// mapAttributeFlag converts a flag value into attributes for a query
func mapAttributeFlag(flag string) (map[string]interface{}, error) {
	var attributes map[string]interface{}
	if err := json.Unmarshal([]byte(flag), &attributes); err != nil {
		return attributes, err
	}
	return attributes, nil
}

// getExpressionPatterns returns the attributes and values in an expression in
// order of appearance
func getExpressionPatterns(expression string) ([]string, []string) {
	attributePattern := regexp.MustCompile(`#([a-zA-Z0-9_]+)`)
	valuePattern := regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
	expressionAttributes := attributePattern.FindAllString(expression, -1)
	expressionValues := valuePattern.FindAllString(expression, -1)
	return expressionAttributes, expressionValues
}

func startQueryExport(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command, query types.AppDatastoreQuery) error {
	totalExportedItems := 0
	userPreferredLimit := maxExportQueryLimit
	if query.Limit > 0 {
		userPreferredLimit = query.Limit
	}

	itemsFile, err := clients.Fs.Create(saveToFileFlag)
	if err != nil {
		return err
	}
	defer itemsFile.Close()

	itemsFilePath, err := filepath.Abs(itemsFile.Name())
	if err != nil {
		return err
	}

	token := config.GetContextToken(ctx)

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "file_cabinet",
		Text:  "Exporting datastore items to a file",
		Secondary: []string{
			fmt.Sprintf("Items will be saved to %s", style.HomePath(itemsFilePath)),
		},
	}))

	exportProgressSpinner = style.NewSpinner(cmd.OutOrStdout())
	defer exportProgressSpinner.Stop()

	for {
		maxItemsToRead := min(maxExportQueryLimit, userPreferredLimit, maxExportItems-totalExportedItems)

		if maxItemsToRead <= 0 {
			break
		}

		query.Limit = maxItemsToRead
		queryResult, err := clients.API().AppsDatastoreQuery(ctx, token, query)
		if err != nil {
			return err
		}

		for _, element := range queryResult.Items {
			stringItem, err := goutils.JSONMarshalUnescaped(element)
			if err != nil {
				return err
			}
			_, err = itemsFile.WriteString(stringItem)
			if err != nil {
				return err
			}
		}

		totalExportedItems += len(queryResult.Items)

		update := fmt.Sprintf("Exported (%d) items.", totalExportedItems)
		exportProgressSpinner.Update(update, "").Start()

		if queryResult.NextCursor == "" {
			break
		}
		query.Cursor = queryResult.NextCursor

	}

	exportProgressSpinner.Update(fmt.Sprintf("Successfully exported (%d) items!", totalExportedItems), "tada").Stop()

	if totalExportedItems >= maxExportItems {
		clients.IO.PrintInfo(ctx, false, "%sExport will be limited to the first %d items in the datastore", style.Emoji("warning"), maxExportItems)
	}
	cmd.Println()

	return nil
}
