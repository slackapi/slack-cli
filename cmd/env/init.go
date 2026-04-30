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
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdotenv"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewEnvInitCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize environment variables from placeholders",
		Long: strings.Join([]string{
			`Initialize the project ".env" file by copying from a template placeholder file.`,
			"",
			`Copies content from either the ".env.sample" or ".env.example" file to the`,
			`project ".env" file if those project environment variables don't already exist.`,
			"",
			fmt.Sprintf("Apps using ROSI features should set environment variables with %s.", style.Commandf("env set", false)),
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Initialize environment variables from template placeholders",
				Command: "env init",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvInitCommandFunc(ctx, clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvInitCommandFunc(clients, cmd)
		},
	}

	return cmd
}

// preRunEnvInitCommandFunc determines if the command is run in a valid project
func preRunEnvInitCommandFunc(_ context.Context, clients *shared.ClientFactory) error {
	return cmdutil.IsValidProjectDirectory(clients)
}

// runEnvInitCommandFunc copies a sample .env file to .env
func runEnvInitCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Hosted apps manage environment variables through the API, not .env files.
	hosted := isHostedRuntime(ctx, clients)
	if hosted {
		selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
		if err != nil {
			return err
		}
		if !selection.App.IsDev {
			clients.IO.PrintTrace(ctx, slacktrace.EnvInitSuccess)
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "evergreen_tree",
				Text:  "Environment Initialize",
				Secondary: []string{
					fmt.Sprintf("Set environment variables for apps using ROSI features with %s", style.Commandf("env set", false)),
				},
			}))
			return nil
		}
	}

	source, err := slackdotenv.Init(clients.Fs)
	if err != nil {
		switch slackerror.ToSlackError(err).Code {
		case slackerror.ErrDotEnvFileAlreadyExists:
			clients.IO.PrintTrace(ctx, slacktrace.EnvInitSuccess)
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "evergreen_tree",
				Text:  "Environment Initialize",
				Secondary: []string{
					`A project ".env" file already exists and was left unchanged`,
					fmt.Sprintf("Set environment variables with %s", style.Commandf("env set", false)),
				},
			}))
			return nil
		case slackerror.ErrDotEnvPlaceholderNotFound:
			clients.IO.PrintTrace(ctx, slacktrace.EnvInitSuccess)
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
				Emoji: "evergreen_tree",
				Text:  "Environment Initialize",
				Secondary: []string{
					`No template placeholder was found for environment variables in this project`,
					fmt.Sprintf("Set environment variables with %s", style.Commandf("env set", false)),
				},
			}))
			return nil
		default:
			return err
		}
	}

	clients.IO.PrintTrace(ctx, slacktrace.EnvInitSuccess)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "evergreen_tree",
		Text:  "Environment Initialize",
		Secondary: []string{
			fmt.Sprintf(`Placeholders were copied from "%s" to a project ".env" file`, source),
			`This new ".env" file shouldn't be added to version control`,
		},
	}))
	return nil
}
