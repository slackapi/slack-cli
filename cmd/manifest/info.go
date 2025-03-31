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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// TODO - In the future, we can support the following:
//    --format flag which will determine if the user wants the printout in yaml or json format.  By default, we will use json.

// NewInfoCommand implements the "manifest info" command
func NewInfoCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print the app manifest of a project or app",
		Long: strings.Join([]string{
			fmt.Sprintf("Get the manifest of an app using either the \"%s\" values on app settings", config.MANIFEST_SOURCE_REMOTE.String()),
			fmt.Sprintf("or from the \"%s\" configurations.", config.MANIFEST_SOURCE_LOCAL.String()),
			"",
			"The manifest on app settings represents the latest version of the manifest.",
			"",
			fmt.Sprintf("Project configurations use the \"get-manifest\" hook from \"%s\".", config.GetProjectHooksJSONFilePath()),
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Print the app manifest from project configurations",
				Command: "manifest info",
			},
			{
				Meaning: "Print the remote manifest of an app",
				Command: "manifest info --app A0123456789",
			},
			{
				Meaning: "Print the app manifest gathered from App Config",
				Command: "manifest info --source remote",
			},
		}),
		Aliases: []string{"show", "list"},
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfoCommand(cmd, clients)
		},
	}
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

// runInfoCommand performs the "manifest info" command
func runInfoCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()
	info, err := getManifestInfo(ctx, clients, cmd)
	if err != nil {
		return err
	}
	manifest, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	clients.IO.PrintInfo(ctx, false, string(manifest))
	return nil
}

// getManifestInfo gathers app manifest information from the specified source
func getManifestInfo(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) (types.AppManifest, error) {
	source, err := getManifestSource(ctx, clients, cmd)
	if err != nil {
		return types.AppManifest{}, err
	}
	switch {
	case source.Equals(config.MANIFEST_SOURCE_LOCAL):
		return getManifestInfoProject(clients)
	case source.Equals(config.MANIFEST_SOURCE_REMOTE):
		return getManifestInfoRemote(ctx, clients)
	default:
		return types.AppManifest{}, slackerror.New(slackerror.ErrInvalidManifestSource)
	}
}

// getManifestInfoProject gathers app manifest information from "get-manifest"
func getManifestInfoProject(clients *shared.ClientFactory) (types.AppManifest, error) {
	slackManifest, err := clients.AppClient().Manifest.GetManifestLocal(
		clients.SDKConfig,
		clients.HookExecutor,
	)
	if err != nil {
		return types.AppManifest{}, err
	}
	return slackManifest.AppManifest, nil
}

// getManifestInfoRemote gathers app manifest information from app settings
func getManifestInfoRemote(ctx context.Context, clients *shared.ClientFactory) (types.AppManifest, error) {
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAndUninstalledApps)
	if err != nil {
		return types.AppManifest{}, err
	}
	slackManifest, err := clients.AppClient().Manifest.GetManifestRemote(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
	)
	if err != nil {
		return types.AppManifest{}, err
	}
	return slackManifest.AppManifest, nil
}
