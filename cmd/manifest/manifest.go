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

package manifest

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// manifestFlagSet contains persistent flag values for this command
type manifestFlagSet struct {
	source string
}

// manifestFlags has the set flag values
var manifestFlags manifestFlagSet

// manifestFlagSource possible values for the "source" flag
const (
	manifestFlagSource = "source"
)

// appSelectPromptFunc provides a handle for stubbing app selections
var appSelectPromptFunc = prompts.AppSelectPrompt

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Print the app manifest of a project or app",
		Long: strings.Join([]string{
			fmt.Sprintf("Get the manifest of an app using either the \"%s\" values on app settings", config.MANIFEST_SOURCE_REMOTE.String()),
			fmt.Sprintf("or from the \"%s\" configurations.", config.MANIFEST_SOURCE_LOCAL.String()),
			"",
			"Subcommands unlock additional engagements and interactions with the manifest.",
			"",
			"The manifest on app settings represents the latest version of the manifest.",
			"",
			fmt.Sprintf("Project configurations use the \"get-manifest\" hook from \"%s\".", config.GetProjectHooksJSONFilePath()),
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Display the app manifest for the current project",
				Command: "manifest info",
			},
			{
				Meaning: "Validate the app manifest generated by a project",
				Command: "manifest validate",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfoCommand(cmd, clients)
		},
	}

	// Add child commands
	cmd.AddCommand(NewInfoCommand(clients))
	cmd.AddCommand(NewValidateCommand(clients))

	cmd.Flags().StringVar(
		&manifestFlags.source,
		manifestFlagSource,
		config.MANIFEST_SOURCE_LOCAL.String(),
		fmt.Sprintf(
			"source of the app manifest (\"%s\" or \"%s\")",
			config.MANIFEST_SOURCE_LOCAL.String(),
			config.MANIFEST_SOURCE_REMOTE.String(),
		),
	)

	return cmd
}

// getManifestSource determines if either the "local" or "remote" manifest
// should be used based on the "--source" flag
func getManifestSource(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) (config.ManifestSource, error) {
	if clients.Config.AppFlag != "" {
		if cmd.Flag(manifestFlagSource).Changed && !config.ManifestSource(manifestFlags.source).Equals(config.MANIFEST_SOURCE_REMOTE) {
			return "", slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage(
					"The \"--%s\" flag must be \"%s\" when using \"--app\"",
					manifestFlagSource,
					config.MANIFEST_SOURCE_REMOTE.String(),
				)
		}
		return config.MANIFEST_SOURCE_REMOTE, nil
	}
	if clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		if !cmd.Flag(manifestFlagSource).Changed {
			manifestConfigSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
			if err != nil {
				return "", err
			}
			switch {
			case manifestConfigSource.Equals(config.MANIFEST_SOURCE_LOCAL):
				return manifestConfigSource, nil
			case manifestConfigSource.Equals(config.MANIFEST_SOURCE_REMOTE):
				return "", slackerror.New(slackerror.ErrInvalidManifestSource).
					WithMessage(`Cannot get manifest info from the "%s" source`, config.MANIFEST_SOURCE_REMOTE).
					WithRemediation("%s", strings.Join([]string{
						fmt.Sprintf("Find the current manifest on app settings: %s", style.LinkText("https://api.slack.com/apps")),
						fmt.Sprintf("Set \"manifest.source\" to \"%s\" in \"%s\" to continue", config.MANIFEST_SOURCE_LOCAL, filepath.Join(".slack", "config.json")),
						fmt.Sprintf("Read about manifest sourcing with %s", style.Commandf("manifest info --help", false)),
					}, "\n"))
			}
		}
	}
	switch {
	case config.ManifestSource(manifestFlags.source).Equals(config.MANIFEST_SOURCE_LOCAL):
		return config.MANIFEST_SOURCE_LOCAL, nil
	case config.ManifestSource(manifestFlags.source).Equals(config.MANIFEST_SOURCE_REMOTE):
		return config.MANIFEST_SOURCE_REMOTE, nil
	default:
		return "", slackerror.New(slackerror.ErrInvalidFlag).
			WithMessage(
				"The \"--%s\" flag must be \"%s\" or \"%s\"",
				manifestFlagSource,
				config.MANIFEST_SOURCE_LOCAL,
				config.MANIFEST_SOURCE_REMOTE,
			)
	}
}
