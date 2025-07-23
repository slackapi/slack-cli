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

var Get = datastore.Get

func NewGetCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <expression> [flags]",
		Short: "Get an item from a datastore",
		Long: strings.Join([]string{
			"Get an item from a datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Get an item from the datastore",
				Command: `datastore get --datastore tasks '{"id": "42"}'`,
			},
			{
				Meaning: "Get an item from the datastore with an expression",
				Command: `datastore get '{"datastore": "tasks", "id": "42"}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunGetCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			var query types.AppDatastoreGet

			// TODO: almost all the code below here except for the actual API call is identical across all datastore commands. can we DRY this up / is it worth it?
			if len(args) > 0 {
				err := setQueryExpression(clients, &query, args[0], "get")
				if err != nil {
					return err
				}
			} else if len(args) == 0 && !unstableFlag {
				return slackerror.New(slackerror.ErrInvalidDatastoreExpression).
					WithMessage("No expression was provided").
					WithRemediation("%s", datastoreExpressionRemediation("get", true))
			}

			// Get the app auth selection from the flag or prompt
			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}
			ctx = config.SetContextToken(ctx, selection.Auth.Token)

			// Build the query if it wasn't passed by argument
			if len(args) == 0 && unstableFlag {
				query, err = promptDatastoreGetRequest(ctx, clients, selection.App, selection.Auth)
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

			// Perform the get
			log := newGetLogger(clients, cmd)
			event, err := Get(ctx, clients, log, query)
			if err != nil {
				return err
			}
			printDatastoreGetSuccess(cmd, event)
			return nil
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().StringVar(&outputFlag, "output", "text", outputUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)
	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)

	return cmd
}

// preRunGetCommandFunc determines if the command is supported for a project and
// configures flags
func preRunGetCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

func newGetLogger(clients *shared.ClientFactory, cmd *cobra.Command) *logger.Logger {
	return logger.New(
		// OnEvent
		func(event *logger.LogEvent) {
			switch event.Name {
			case "on_get_result":
				getResult := types.AppDatastoreGetResult{}
				if event.Data["getResult"] != nil {
					getResult = event.Data["getResult"].(types.AppDatastoreGetResult)
				}
				if cmd != nil {
					// TODO: this can raise an error on indentation failures, but not sure how to handle that using logger
					_ = printGetResult(clients, cmd, getResult)
				}
			default:
				// Ignore the event
			}
		},
	)
}

func printGetResult(clients *shared.ClientFactory, cmd *cobra.Command, getResult types.AppDatastoreGetResult) error {
	var datastore = getResult.Datastore
	var item = getResult.Item

	if outputFlag == "text" {
		cmd.Printf(
			style.Bold("%s Get from Datastore: %s\n\n"),
			style.Emoji("tada"),
			datastore,
		)
	}

	var b []byte
	var err error
	if len(item) == 0 {
		b = []byte("Not found")
	} else {
		b, err = json.MarshalIndent(item, "", "  ")
		if err != nil {
			return slackerror.New("Error during output indentation").WithRootCause(err)
		}
	}

	cmd.Printf(
		"%s\n",
		string(b),
	)
	return nil
}

func printDatastoreGetSuccess(cmd *cobra.Command, event *logger.LogEvent) {
	if outputFlag == "text" {
		commandText := style.Commandf("datastore get <expression>", true)
		if cmd != nil {
			cmd.Printf(
				"To inspect the datastore after updates, run %s\n",
				commandText,
			)
		}
	}
}

// promptDatastoreCountRequest constructs a datastore get expression by prompting
func promptDatastoreGetRequest(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) (
	types.AppDatastoreGet,
	error,
) {
	var query types.AppDatastoreGet

	// Collect datastore information from the manifest
	yaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return types.AppDatastoreGet{}, err
	}

	var datastores = []string{}
	for name := range yaml.Datastores {
		datastores = append(datastores, name)
	}
	if len(datastores) <= 0 {
		return types.AppDatastoreGet{}, slackerror.New(slackerror.ErrDatastoreNotFound).
			WithMessage("No datastores are associated with this app")
	}
	sort.Strings(datastores)

	// Prompt for information to create the query
	selection, err := clients.IO.SelectPrompt(ctx, "Select a datastore", datastores, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("datastore"),
		Required: true,
	})
	if err != nil {
		return types.AppDatastoreGet{}, err
	} else if yaml.Datastores[selection.Option].PrimaryKey == "" {
		return types.AppDatastoreGet{}, slackerror.New(slackerror.ErrDatastoreNotFound)
	} else {
		query.Datastore = selection.Option
	}

	primaryKeyPrompt := fmt.Sprintf("Enter a %s", yaml.Datastores[query.Datastore].PrimaryKey)
	query.ID, err = clients.IO.InputPrompt(ctx, primaryKeyPrompt, iostreams.InputPromptConfig{
		Required: true,
	})
	if err != nil {
		return types.AppDatastoreGet{}, err
	}

	return query, nil
}
