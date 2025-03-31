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

var Put = datastore.Put

func NewPutCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put <expression> [flags]",
		Short: "Create or replace an item in a datastore",
		Long: strings.Join([]string{
			"Create or replace an item in a datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Add a new entry to the datastore",
				Command: `datastore put --datastore tasks '{"item": {"id": "42", "description": "Create a PR", "status": "Done"}}'`,
			},
			{
				Meaning: "Add a new entry to the datastore with an expression",
				Command: `datastore put '{"datastore": "tasks", "item": {"id": "42", "description": "Create a PR", "status": "Done"}}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunPutCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastorePut

			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "put")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("put", true))
			}

			// Get the selection from the flag or prompt
			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}
			ctx = config.SetContextToken(ctx, selection.Auth.Token)

			// Build the query if it wasn't passed by argument
			if len(args) == 0 && unstableFlag {
				query, err = promptDatastorePutRequest(ctx, clients, selection.App, selection.Auth)
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

			// Perform the put
			log := newPutLogger(clients, cmd)
			event, err := Put(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastorePutSuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)
	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)

	return cmd
}

// preRunPutCommandFunc determines if the command is supported for a project and
// configures flags
func preRunPutCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

func newPutLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_put_result":
				putResult := types.AppDatastorePutResult{}
				if event.Data["putResult"] != nil {
					putResult = event.Data["putResult"].(types.AppDatastorePutResult)
				}
				if cmd != nil {
					// TODO: this can raise an error on indentation failures, but not sure how to handle that using logger
					_ = printPutResult(clients, cmd, putResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printPutResult(clients *shared.ClientFactory, cmd *cobra.Command, putResult types.AppDatastorePutResult) error {
	var datastore = putResult.Datastore
	var item = putResult.Item
	cmd.Printf(
		style.Bold("%s Stored below record in the datastore: %s\n\n"),
		style.Emoji("tada"),
		datastore,
	)
	b, err := goutils.JsonMarshalUnescapedIndent(item)
	if err != nil {
		return slackerror.New("Error during output indentation").WithRootCause(err)
	}
	cmd.Printf(
		"%s\n",
		string(b),
	)
	return nil
}

func printDatastorePutSuccess(cmd *cobra.Command, event *logger.LogEvent) {
	commandText := style.Commandf("datastore query <expression>", true)
	if cmd != nil {
		cmd.Printf(
			"To inspect the datastore after updates, run %s\n",
			commandText,
		)
	}
}

// promptDatastorePutRequest constructs a datastore put expression by prompting
func promptDatastorePutRequest(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) (
	types.AppDatastorePut,
	error,
) {
	var query types.AppDatastorePut

	// Collect datastore information from the manifest
	yaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return types.AppDatastorePut{}, err
	}

	var datastores = []string{}
	for name := range yaml.Datastores {
		datastores = append(datastores, name)
	}
	if len(datastores) <= 0 {
		return types.AppDatastorePut{}, slackerror.New(slackerror.ErrDatastoreNotFound).
			WithMessage("No datastores are associated with this app")
	}
	sort.Strings(datastores)

	selection, err := clients.IO.SelectPrompt(ctx, "Select a datastore", datastores, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("datastore"),
		Required: true,
	})
	if err != nil {
		return types.AppDatastorePut{}, err
	} else if yaml.Datastores[selection.Option].PrimaryKey == "" {
		return types.AppDatastorePut{}, slackerror.New(slackerror.ErrDatastoreNotFound)
	} else {
		query.Datastore = selection.Option
	}

	// Gather necessary information for the expression
	fields := yaml.Datastores[query.Datastore].Attributes
	if len(fields) > 0 {
		query.Item = make(map[string]interface{})
	}

	// Prompt for the primary key first
	primaryKey := yaml.Datastores[query.Datastore].PrimaryKey
	primaryKeyPrompt := fmt.Sprintf("Enter a value for '%s':", primaryKey)
	recordId, err := clients.IO.InputPrompt(ctx, primaryKeyPrompt, iostreams.InputPromptConfig{
		Required: true,
	})
	if err != nil {
		return types.AppDatastorePut{}, err
	}
	query.Item[primaryKey] = recordId
	delete(fields, primaryKey)

	for field := range fields {
		fieldPrompt := fmt.Sprintf("Enter a value for '%s':", field)
		query.Item[field], err = clients.IO.InputPrompt(ctx, fieldPrompt, iostreams.InputPromptConfig{
			Required: false,
		})
		if err != nil {
			return types.AppDatastorePut{}, err
		}
	}

	return query, nil
}
