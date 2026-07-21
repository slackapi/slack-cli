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

package manifest

import (
	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/manifest"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var manifestSyncFunc = manifest.Sync

func NewSyncCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "sync",
		Short:  "Sync the app manifest between project and app settings",
		Long:   "Compare the local project manifest with app settings, resolve differences, and sync both to the same state.",
		Hidden: true,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "manifest sync", Meaning: "Sync project manifest with app settings"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !clients.Config.WithExperimentOn(experiment.ManifestSync) {
				return slackerror.New(slackerror.ErrExperimentRequired).
					WithRemediation("Enable the %s experiment with %s",
						style.Highlight(string(experiment.ManifestSync)),
						style.CommandText("--experiment manifest-sync"),
					)
			}
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.manifest.sync")
			defer span.Finish()

			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}

			clients.Config.ManifestEnv = app.SetManifestEnvTeamVars(clients.Config.ManifestEnv, selection.App.TeamDomain, selection.App.IsDev)

			_, err = manifestSyncFunc(ctx, clients, selection.App, selection.Auth)
			return err
		},
	}
	return cmd
}
