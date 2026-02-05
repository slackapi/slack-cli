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

package apps

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Add will add an app
func Add(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, auth types.SlackAuth, app types.App, orgGrantWorkspaceID string) (types.InstallState, types.App, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pkg.apps.add")
	defer span.Finish()

	// Validate the auth
	ctx, authSession, err := getAuthSession(ctx, clients, auth)
	if err != nil {
		return "", types.App{}, slackerror.Wrap(err, slackerror.ErrAddAppToProject)
	}
	log.Data["teamName"] = *authSession.TeamName

	log.Info("on_apps_add_init")

	// Add app remotely via Slack API
	installState, app, err := addAppRemotely(ctx, clients, log, auth, app, orgGrantWorkspaceID)
	if err != nil {
		return "", types.App{}, slackerror.Wrap(err, slackerror.ErrAddAppToProject)
	}

	// Add app to apps.json
	if !clients.Config.SkipLocalFs() {
		if _, err := addAppLocally(ctx, clients, log, app); err != nil {
			return installState, types.App{}, slackerror.Wrap(err, slackerror.ErrAddAppToProject)
		}
	}

	return installState, app, nil
}

// addAppLocally will add the app to the project's apps file with an empty AppID, TeamID, etc
func addAppLocally(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, app types.App) (types.App, error) {
	log.Info("on_apps_add_local")

	app, err := clients.AppClient().NewDeployed(ctx, app.TeamID)
	if err != nil {
		if !strings.Contains(err.Error(), slackerror.ErrAppFound) { // Ignore the error when the app already exists
			return types.App{}, slackerror.Wrap(err, slackerror.ErrAddAppToProject)
		}
	}
	return app, nil
}

// addAppRemotely will create the app manifest using the current auth account's team
func addAppRemotely(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, auth types.SlackAuth, app types.App, orgGrantWorkspaceID string) (types.InstallState, types.App, error) {
	log.Info("on_apps_add_remote_init")

	if app.TeamID == "" {
		// App hasn't been created yet and
		// so the target team ID set by the auth
		app.TeamID = auth.TeamID
	}

	app, installState, err := Install(ctx, clients, log, auth, CreateAppManifestAndInstall, app, orgGrantWorkspaceID)
	if err != nil {
		return installState, types.App{}, slackerror.Wrap(err, slackerror.ErrAppAdd)
	}

	log.Info("on_apps_add_remote_success")

	if !clients.Config.SkipLocalFs() {
		app, err = clients.AppClient().GetDeployed(ctx, app.TeamID)
	}

	return installState, app, err
}
