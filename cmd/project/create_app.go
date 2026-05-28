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

package project

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

// resolveAuthForApp finds an authenticated workspace that has access to the given app ID.
func resolveAuthForApp(ctx context.Context, clients *shared.ClientFactory, appID string) (types.SlackAuth, error) {
	if clients.Config.TokenFlag != "" {
		auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return types.SlackAuth{}, slackerror.Wrap(err, slackerror.ErrNotAuthed)
		}
		return auth, nil
	}

	allAuths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return types.SlackAuth{}, slackerror.Wrap(err, slackerror.ErrNotAuthed)
	}

	if len(allAuths) == 0 {
		return types.SlackAuth{}, slackerror.New(slackerror.ErrNotAuthed).
			WithMessage("No workspaces connected").
			WithRemediation("Run %s to sign in to a workspace that has access to app %s", style.Commandf("login", false), appID)
	}

	if clients.Config.TeamFlag != "" {
		for i := range allAuths {
			if allAuths[i].TeamID == clients.Config.TeamFlag || allAuths[i].TeamDomain == clients.Config.TeamFlag {
				if _, err := clients.API().GetAppStatus(ctx, allAuths[i].Token, []string{appID}, allAuths[i].TeamID); err == nil {
					return allAuths[i], nil
				}
			}
		}
		return types.SlackAuth{}, slackerror.New(slackerror.ErrTeamNotFound).
			WithMessage("The specified team does not have access to app %s", appID).
			WithRemediation("Run %s to sign in to the workspace that owns this app", style.Commandf("login", false))
	}

	for i := range allAuths {
		if _, err := clients.API().GetAppStatus(ctx, allAuths[i].Token, []string{appID}, allAuths[i].TeamID); err == nil {
			return allAuths[i], nil
		}
	}

	return types.SlackAuth{}, slackerror.New(slackerror.ErrAppNotFound).
		WithMessage("No authenticated workspace has access to app %s", appID).
		WithRemediation("Run %s to sign in to the workspace that owns this app", style.Commandf("login", false))
}

// fetchRemoteManifest retrieves the app manifest from the platform via apps.manifest.export.
func fetchRemoteManifest(ctx context.Context, clients *shared.ClientFactory, token string, appID string) (types.SlackYaml, error) {
	manifest, err := clients.AppClient().Manifest.GetManifestRemote(ctx, token, appID)
	if err != nil {
		return types.SlackYaml{}, slackerror.New(slackerror.ErrInvalidManifest).
			WithMessage("Failed to fetch manifest for app %s", appID)
	}
	return manifest, nil
}

// writeManifestToProject writes the fetched manifest JSON to the project directory.
func writeManifestToProject(fs afero.Fs, projectPath string, manifest types.SlackYaml) error {
	manifestData, err := json.MarshalIndent(manifest.AppManifest, "", "  ")
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrProjectFileUpdate).
			WithMessage("Failed to serialize app manifest")
	}

	manifestPath := filepath.Join(projectPath, "manifest.json")
	if err := afero.WriteFile(fs, manifestPath, append(manifestData, '\n'), 0644); err != nil {
		return slackerror.Wrap(err, slackerror.ErrProjectFileUpdate).
			WithMessage("Failed to write manifest to project")
	}
	return nil
}

// linkAppToProject saves the app to the project's apps JSON file.
// Defaults to local/dev unless the manifest explicitly uses a hosted runtime.
func linkAppToProject(ctx context.Context, clients *shared.ClientFactory, auth types.SlackAuth, appID string, manifest types.SlackYaml) error {
	app := types.App{
		AppID:        appID,
		TeamID:       auth.TeamID,
		TeamDomain:   auth.TeamDomain,
		EnterpriseID: auth.EnterpriseID,
	}

	if manifest.IsFunctionRuntimeSlackHosted() {
		return clients.AppClient().SaveDeployed(ctx, app)
	}
	app.IsDev = true
	app.UserID = auth.UserID
	return clients.AppClient().SaveLocal(ctx, app)
}
