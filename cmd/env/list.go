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

package env

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewEnvListCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List all environment variables for the app",
		Long: strings.Join([]string{
			"List all of the environment variables of an app deployed to Slack managed",
			"infrastructure.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "List all environment variables",
				Command: "env list",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvListCommandFunc(ctx, clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvListCommandFunc(clients, cmd)
		},
	}

	return cmd
}

// preRunEnvListCommandFunc determines if the command is supported for a project
// and configures flags
func preRunEnvListCommandFunc(ctx context.Context, clients *shared.ClientFactory) error {
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return err
	}
	if clients.Config.ForceFlag {
		return nil
	}
	return cmdutil.IsSlackHostedProject(ctx, clients)
}

// runEnvListCommandFunc outputs environment variables for a selected app
func runEnvListCommandFunc(
	clients *shared.ClientFactory,
	cmd *cobra.Command,
) error {
	ctx := cmd.Context()

	selection, err := teamAppSelectPromptFunc(
		ctx,
		clients,
		prompts.ShowHostedOnly,
		prompts.ShowInstalledAppsOnly,
	)
	if err != nil {
		return err
	}

	variableNames, err := clients.API().ListVariables(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
	)
	if err != nil {
		return err
	}

	count := len(variableNames)
	clients.IO.PrintTrace(ctx, slacktrace.EnvListCount, strconv.Itoa(count))
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "evergreen_tree",
		Text:  "App Environment",
		Secondary: []string{
			fmt.Sprintf(
				"There %s %d %s stored in this environment",
				style.Pluralize("is", "are", count),
				count,
				style.Pluralize("variable", "variables", count),
			),
		},
	}))

	if len(variableNames) <= 0 {
		return nil
	}
	sort.Strings(variableNames)
	variableLabel := []string{}
	for _, v := range variableNames {
		variableLabel = append(
			variableLabel,
			fmt.Sprintf("%s: %s", v, style.Secondary("***")),
		)
	}
	clients.IO.PrintTrace(ctx, slacktrace.EnvListVariables, variableNames...)
	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji:     "evergreen_tree",
		Text:      "App Environment",
		Secondary: variableLabel,
	}))

	return nil
}
