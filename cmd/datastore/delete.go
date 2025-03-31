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

var Delete = datastore.Delete

func NewDeleteCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <expression> [flags]",
		Short: "Delete an item from a datastore",
		Long: strings.Join([]string{
			"Delete an item from a datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Remove an item from the datastore",
				Command: `datastore delete --datastore tasks '{"id": "42"}'`,
			},
			{
				Meaning: "Remove an item from the datastore with an expression",
				Command: `datastore delete '{"datastore": "tasks", "id": "42"}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunDeleteCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastoreDelete

			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "delete")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("delete", true))
			}

			// Get the app auth selection from the flag or prompt
			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}
			ctx = config.SetContextToken(ctx, selection.Auth.Token)

			// Build the query if it wasn't passed by argument
			if len(args) == 0 && unstableFlag {
				query, err = promptDatastoreDeleteRequest(ctx, clients, selection.App, selection.Auth)
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

			// Perform the delete
			log := newDeleteLogger(clients, cmd)
			event, err := Delete(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastoreDeleteSuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)
	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)

	return cmd
}

// preRunDeleteCommandFunc determines if the command is supported for a project
// and configures flags
func preRunDeleteCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

func newDeleteLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_delete_result":
				deleteResult := types.AppDatastoreDeleteResult{}
				if event.Data["deleteResult"] != nil {
					deleteResult = event.Data["deleteResult"].(types.AppDatastoreDeleteResult)
				}
				if cmd != nil {
					printDeleteResult(clients, cmd, deleteResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printDeleteResult(clients *shared.ClientFactory, cmd *cobra.Command, deleteResult types.AppDatastoreDeleteResult) {
	var datastore = deleteResult.Datastore
	var id = deleteResult.ID
	cmd.Printf(
		style.Bold("%s Deleted from datastore: %s\n\n"),
		style.Emoji("tada"),
		datastore,
	)
	cmd.Printf(
		"primary_key: %s\n",
		id,
	)
}

func printDatastoreDeleteSuccess(cmd *cobra.Command, event *logger.LogEvent) {
	commandText := style.Commandf("datastore query <expression>", true)
	if cmd != nil {
		cmd.Printf(
			"To inspect the datastore after updates, run %s\n",
			commandText,
		)
	}
}

// promptDatastoreDeleteRequest constructs a datastore delete expression by prompting
func promptDatastoreDeleteRequest(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) (
	types.AppDatastoreDelete,
	error,
) {
	var query types.AppDatastoreDelete

	// Collect datastore information from the manifest
	yaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return types.AppDatastoreDelete{}, err
	}

	var datastores = []string{}
	for name := range yaml.Datastores {
		datastores = append(datastores, name)
	}
	if len(datastores) <= 0 {
		return types.AppDatastoreDelete{}, slackerror.New(slackerror.ErrDatastoreNotFound).
			WithMessage("No datastores are associated with this app")
	}
	sort.Strings(datastores)

	// Prompt for information to create the query
	selection, err := clients.IO.SelectPrompt(ctx, "Select a datastore", datastores, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("datastore"),
		Required: true,
	})
	if err != nil {
		return types.AppDatastoreDelete{}, err
	} else if yaml.Datastores[selection.Option].PrimaryKey == "" {
		return types.AppDatastoreDelete{}, slackerror.New(slackerror.ErrDatastoreNotFound)
	} else {
		query.Datastore = selection.Option
	}

	primaryKeyPrompt := fmt.Sprintf("Enter a %s", yaml.Datastores[query.Datastore].PrimaryKey)
	query.ID, err = clients.IO.InputPrompt(ctx, primaryKeyPrompt, iostreams.InputPromptConfig{
		Required: true,
	})
	if err != nil {
		return types.AppDatastoreDelete{}, err
	}

	return query, nil
}
