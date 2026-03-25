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

package env

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/hooks"
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
			"List environment variables available to the app at runtime.",
			"",
			"Commands that run in the context of a project source environment variables from",
			"the \".env\" file. This includes the \"run\" command.",
			"",
			"The \"deploy\" command gathers environment variables from the \".env\" file as well",
			"unless the app is using ROSI features.",
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

// preRunEnvListCommandFunc determines if the command is run in a valid project
func preRunEnvListCommandFunc(_ context.Context, clients *shared.ClientFactory) error {
	return cmdutil.IsValidProjectDirectory(clients)
}

// runEnvListCommandFunc outputs environment variables for a selected app
func runEnvListCommandFunc(
	clients *shared.ClientFactory,
	cmd *cobra.Command,
) error {
	ctx := cmd.Context()

	selection, err := appSelectPromptFunc(
		ctx,
		clients,
		prompts.ShowAllEnvironments,
		prompts.ShowInstalledAppsOnly,
	)
	if err != nil {
		return err
	}

	// Gather environment variables for either a ROSI app from the Slack API method
	// or read from project files.
	var variableNames []string
	if !selection.App.IsDev && cmdutil.IsSlackHostedProject(ctx, clients) == nil {
		variableNames, err = clients.API().ListVariables(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
		)
		if err != nil {
			return err
		}
	} else {
		dotEnv, err := hooks.LoadDotEnv(clients.Fs)
		if err != nil {
			return err
		}
		variableNames = make([]string, 0, len(dotEnv))
		for k := range dotEnv {
			variableNames = append(variableNames, k)
		}
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

	if count <= 0 {
		return nil
	}

	sort.Strings(variableNames)
	variableLabels := make([]string, 0, count)
	for _, v := range variableNames {
		variableLabels = append(
			variableLabels,
			fmt.Sprintf("%s: %s", v, style.Secondary("***")),
		)
	}
	clients.IO.PrintTrace(ctx, slacktrace.EnvListVariables, variableNames...)
	clients.IO.PrintInfo(ctx, false, "%s", style.Sectionf(style.TextSection{
		Emoji:     "evergreen_tree",
		Text:      "App Environment",
		Secondary: variableLabels,
	}))

	return nil
}
