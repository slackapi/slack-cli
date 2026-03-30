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

	"github.com/joho/godotenv"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdotenv"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func NewEnvAddCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <value> [flags]",
		Short: "Add an environment variable to the app",
		Long: strings.Join([]string{
			"Add an environment variable to the app.",
			"",
			"If a name or value is not provided, you will be prompted to provide these.",
			"",
			"Commands that run in the context of a project source environment variables from",
			"the \".env\" file. This includes the \"run\" command.",
			"",
			"The \"deploy\" command gathers environment variables from the \".env\" file as well",
			"unless the app is using ROSI features.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Prompt for an environment variable",
				Command: "env add",
			},
			{
				Meaning: "Add an environment variable",
				Command: "env add MAGIC_PASSWORD abracadbra",
			},
			{
				Meaning: "Prompt for an environment variable value",
				Command: "env add SECRET_PASSWORD",
			},
		}),
		Args: cobra.MaximumNArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunEnvAddCommandFunc(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnvAddCommandFunc(clients, cmd, args)
		},
	}

	cmd.Flags().StringVar(&variableValueFlag, "value", "", "set the environment variable value")

	return cmd
}

// preRunEnvAddCommandFunc determines if the command is run in a valid project
// and configures flags
func preRunEnvAddCommandFunc(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	return cmdutil.IsValidProjectDirectory(clients)
}

// runEnvAddCommandFunc sets an app environment variable to given values
func runEnvAddCommandFunc(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get the workspace from the flag or prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Get the variable name from the args or prompt
	var variableName string
	if len(args) < 1 {
		variableName, err = clients.IO.InputPrompt(ctx, "Variable name", iostreams.InputPromptConfig{
			Required: false,
		})
		if err != nil {
			return err
		}
	} else {
		variableName = args[0]

		// Display the variable name before getting the variable value
		if len(args) < 2 && !clients.Config.Flags.Lookup("value").Changed {
			mimickedInput := iostreams.MimicInputPrompt("Variable name", variableName)
			clients.IO.PrintInfo(ctx, false, "%s", mimickedInput)
		}
	}

	// Get the variable value from the args or prompt
	var variableValue string
	if len(args) < 2 {
		response, err := clients.IO.PasswordPrompt(ctx, "Variable value", iostreams.PasswordPromptConfig{
			Flag: clients.Config.Flags.Lookup("value"),
		})
		if err != nil {
			return err
		} else {
			variableValue = response.Value
		}
	} else {
		variableValue = args[1]
	}

	// Add the environment variable using either the Slack API method or the
	// project ".env" file depending on the app hosting.
	if !selection.App.IsDev && cmdutil.IsSlackHostedProject(ctx, clients) == nil {
		err = clients.API().AddVariable(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
			variableName,
			variableValue,
		)
		if err != nil {
			return err
		}
	} else {
		err = setDotEnv(clients.Fs, variableName, variableValue)
		if err != nil {
			return err
		}
	}

	clients.IO.PrintTrace(ctx, slacktrace.EnvAddSuccess)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "evergreen_tree",
		Text:  "App Environment",
		Secondary: []string{
			fmt.Sprintf(
				"Successfully added \"%s\" as an environment variable",
				variableName,
			),
		},
	}))
	return nil
}

// setDotEnv sets a single environment variable in the .env file, preserving
// comments, blank lines, and other formatting. If the key already exists its
// value is replaced in-place. Otherwise the entry is appended. The file is
// created if it does not exist.
func setDotEnv(fs afero.Fs, name string, value string) error {
	newEntry, err := godotenv.Marshal(map[string]string{name: value})
	if err != nil {
		return err
	}

	// Check for an existing .env file and parse it to detect existing keys.
	existing, err := slackdotenv.Read(fs)
	if err != nil {
		return err
	}

	// If the file does not exist or the key is new, append the entry.
	if existing == nil {
		return afero.WriteFile(fs, ".env", []byte(newEntry+"\n"), 0644)
	}

	oldValue, found := existing[name]
	if !found {
		raw, err := afero.ReadFile(fs, ".env")
		if err != nil {
			return err
		}
		content := string(raw)
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return afero.WriteFile(fs, ".env", []byte(content+newEntry+"\n"), 0644)
	}

	// The key exists — replace the old entry in the raw file content. Build
	// possible representations of the old entry to find and replace in the raw
	// bytes, since the file format may differ from what Marshal produces.
	raw, err := afero.ReadFile(fs, ".env")
	if err != nil {
		return err
	}

	// Strip the "NAME=" prefix from the marshaled new entry to get just the
	// value portion, then build the old entry alternatives using the same
	// prefix variations the file might contain.
	marshaledValue := strings.TrimPrefix(newEntry, name+"=")
	oldMarshaledValue := marshaledValue
	oldEntry, err := godotenv.Marshal(map[string]string{name: oldValue})
	if err == nil {
		oldMarshaledValue = strings.TrimPrefix(oldEntry, name+"=")
	}

	// Try each possible form of the old entry, longest (most specific) first.
	entries := []string{
		"export " + name + "=" + oldMarshaledValue,
		"export " + name + "=" + oldValue,
		"export " + name + "=" + "'" + oldValue + "'",
		name + "=" + oldMarshaledValue,
		name + "=" + oldValue,
		name + "=" + "'" + oldValue + "'",
	}

	content := string(raw)
	replaced := false
	for _, entry := range entries {
		if strings.Contains(content, entry) {
			replacement := newEntry
			if strings.HasPrefix(entry, "export ") {
				replacement = "export " + newEntry
			}
			content = strings.Replace(content, entry, replacement, 1)
			replaced = true
			break
		}
	}
	if !replaced {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += newEntry + "\n"
	}
	return afero.WriteFile(fs, ".env", []byte(content), 0644)
}
