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
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCountCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count [expression]",
		Short: "Count the number of items in a datastore",
		Long: strings.Join([]string{
			"Count the number of items in a datastore that match a query expression or just",
			"all of the items in the datastore.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Count all items in a datastore",
				Command: `datastore count --datastore tasks`,
			},
			{
				Meaning: "Count number of items in datastore that match a query",
				Command: `datastore count '{"datastore": "tasks", "expression": "#status = :status", "expression_attributes": {"#status": "status"}, "expression_values": {":status": "In Progress"}}'`,
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunCountCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var ctx = cmd.Context()
			return runCountCommandFunc(ctx, clients, args)
		},
	}
	cmd.Flags().StringVar(&datastoreFlag, "datastore", "", datastoreUsage)
	cmd.Flags().BoolVar(&showExpressionFlag, "show", false, showExpressionUsage)

	cmd.Flags().BoolVar(&unstableFlag, "unstable", false, unstableUsage)
	cmd.Flags().StringVar(&attributeFlag, "attributes", "", attributeUsage)

	cmd.Flag("attributes").Hidden = true // Hide while unstable is present

	return cmd
}

// preRunCountCommandFunc determines if the command is supported for a project
// and configures flags
func preRunCountCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

// runCountCommandFunc processes input to construct a query expression then
// calls the API
func runCountCommandFunc(
	ctx context.Context,
	clients *shared.ClientFactory,
	args []string,
) error {
	var count types.AppDatastoreCount

	if len(args) > 0 {
		err := setQueryExpression(clients, &count, args[0], "count")
		if err != nil {
			return err
		}
	} else if len(args) == 0 && !unstableFlag {
		err := setQueryExpression(clients, &count, "{}", "count")
		if err != nil {
			return err
		}
		if count.Datastore == "" {
			return slackerror.New(slackerror.ErrDatastoreNotFound).
				WithMessage("No datastore was specified").
				WithRemediation("Provide a datastore with the %s flag", style.Highlight("--datastore"))
		}
	}

	// Get the app from the flag or prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Build the count if it wasn't passed by argument
	if len(args) == 0 && unstableFlag {
		count, err = promptDatastoreCountRequest(ctx, clients, selection.App, selection.Auth)
		if err != nil {
			return err
		}
	}

	// Set the app ID from the selected workspace
	count.App = selection.App.AppID

	// Optionally display the JSON expression and exit
	if showExpressionFlag {
		return printDatastoreExpressionMarshal(ctx, clients, count)
	}

	// Perform the count
	countResult, err := clients.APIInterface().AppsDatastoreCount(ctx, selection.Auth.Token, count)
	if err != nil {
		return err
	}
	return printCountResult(ctx, clients, countResult)
}

// printCountResult outputs information about the count command result
func printCountResult(ctx context.Context, clients *shared.ClientFactory, countResult types.AppDatastoreCountResult) error {
	clients.IO.PrintTrace(ctx, slacktrace.DatastoreCountSuccess)
	clients.IO.PrintTrace(ctx, slacktrace.DatastoreCountTotal, fmt.Sprintf("%d", countResult.Count))
	clients.IO.PrintTrace(ctx, slacktrace.DatastoreCountDatastore, countResult.Datastore)
	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji: "tada",
		Text: fmt.Sprintf(
			"Counted %d matching items from datastore: %s",
			countResult.Count,
			countResult.Datastore,
		),
	}))
	clients.IO.PrintInfo(ctx, false, "Create or update existing items with %s",
		style.Commandf("datastore put <expression>", true),
	)
	return nil
}

// promptDatastoreCountRequest constructs a datastore count expression by prompting
func promptDatastoreCountRequest(
	ctx context.Context,
	clients *shared.ClientFactory,
	app types.App,
	auth types.SlackAuth,
) (
	types.AppDatastoreCount,
	error,
) {
	var count types.AppDatastoreCount

	// Collect datastore information from the manifest
	yaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return types.AppDatastoreCount{}, err
	}

	var datastores []string
	for name := range yaml.Datastores {
		datastores = append(datastores, name)
	}
	if len(datastores) <= 0 {
		return types.AppDatastoreCount{}, slackerror.New(slackerror.ErrDatastoreNotFound).
			WithMessage("No datastores are associated with this app")
	}
	sort.Strings(datastores)

	selection, err := clients.IO.SelectPrompt(ctx, "Select a datastore", datastores, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("name"),
		Required: true,
	})
	if err != nil {
		return types.AppDatastoreCount{}, err
	} else if yaml.Datastores[selection.Option].PrimaryKey == "" {
		return types.AppDatastoreCount{}, slackerror.New(slackerror.ErrDatastoreNotFound)
	} else {
		count.Datastore = selection.Option
	}

	// Display a hint for writing expressions
	clients.IO.PrintInfo(ctx, false, "")
	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji: "bulb",
		Text:  "Expressions should use the following format",
		Secondary: []string{
			fmt.Sprintf(`Attributes begin with "#" (Example: "#%s")`, yaml.Datastores[count.Datastore].PrimaryKey),
			`Values begin with ":" (Example: ":num")`,
			fmt.Sprintf(`Example expression: #%s < :num`, yaml.Datastores[count.Datastore].PrimaryKey),
		},
	}))

	// Gather the expression and necessary information for the expression
	expression, err := clients.IO.InputPrompt(ctx, "Enter an expression", iostreams.InputPromptConfig{
		Required: false,
	})
	if err != nil {
		return types.AppDatastoreCount{}, err
	}
	count.Expression = expression
	attributes, values := getExpressionPatterns(expression)

	fields := yaml.Datastores[count.Datastore].Attributes
	var fieldSlice []string
	for field := range fields {
		fieldSlice = append(fieldSlice, field)
	}

	if len(attributes) > 0 {
		count.ExpressionAttributes = make(map[string]interface{})
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
				return types.AppDatastoreCount{}, err
			} else if selection.Flag {
				if attributes, err := mapAttributeFlag(selection.Option); err != nil {
					return types.AppDatastoreCount{}, err
				} else {
					count.ExpressionAttributes = attributes
					break
				}
			} else if selection.Prompt {
				count.ExpressionAttributes[attribute] = selection.Option
			}
		} else {
			count.ExpressionAttributes[attribute] = field
		}
	}

	if len(values) > 0 {
		count.ExpressionValues = make(map[string]interface{})
	}
	for _, value := range values {
		field := strings.TrimPrefix(value, ":")
		valuePrompt := fmt.Sprintf("Enter a value for ':%s'", field)
		field, err = clients.IO.InputPrompt(ctx, valuePrompt, iostreams.InputPromptConfig{
			Required: true,
		})
		if err != nil {
			return types.AppDatastoreCount{}, err
		}
		count.ExpressionValues[value] = field
	}

	return count, nil
}
