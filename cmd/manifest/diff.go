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
	"github.com/slackapi/slack-cli/internal/manifest"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// NewDiffCommand implements the "manifest diff" command, which prints the
// differences between the project manifest and the app settings on Slack.
func NewDiffCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Show differences between the project manifest and app settings",
		Long:  "Compare the project manifest with app settings and print any differences.",
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "manifest diff", Meaning: "Show differences between project manifest and app settings"},
			{Command: "manifest diff --app A0123456789 --token xoxp-...", Meaning: "Show manifest differences without prompts"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.IsValidProjectDirectory(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.manifest.diff")
			defer span.Finish()

			selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
			if err != nil {
				return err
			}

			clients.Config.ManifestEnv = app.SetManifestEnvTeamVars(clients.Config.ManifestEnv, selection.App.TeamDomain, selection.App.IsDev)

			localManifest, err := clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
			if err != nil {
				return err
			}

			remoteManifest, err := clients.AppClient().Manifest.GetManifestRemote(ctx, selection.Auth.Token, selection.App.AppID)
			if err != nil {
				return err
			}

			diffs, err := manifest.Diff(localManifest.AppManifest, remoteManifest.AppManifest, selection.App.IsDev)
			if err != nil {
				return err
			}

			if !diffs.HasDifferences() {
				clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
					Emoji:     "books",
					Text:      "Manifest Diff",
					Secondary: []string{"Project manifest and app settings match"},
				}))
				return nil
			}

			displayDiffs(ctx, clients.IO, diffs)
			return nil
		},
	}
}
