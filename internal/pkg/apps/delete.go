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

package apps

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Delete will delete the app for this teamDomain both remotely (API) and locally (project)
func Delete(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, teamDomain string, app types.App, auth types.SlackAuth) (types.App, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pkg.apps.delete")
	defer span.Finish()

	// Set selected team as the current team
	clients.Config.TeamFlag = teamDomain

	// Get Team Name
	ctx, authSession, err := getAuthSession(ctx, clients, auth)
	if err != nil {
		return types.App{}, slackerror.Wrap(err, slackerror.ErrAppRemove)
	}

	// Is this needed anymore?
	if authSession.TeamName != nil {
		log.Data["teamName"] = *authSession.TeamName
	}

	// Emit starting to remove app (requires teamName)
	log.Info("on_apps_delete_init")

	if app.AppID == "" {
		err = slackerror.New(slackerror.ErrAppNotFound).WithMessage("App not found for team '%s'", teamDomain)
		return types.App{}, slackerror.Wrap(err, slackerror.ErrAppRemove)
	}

	log.Info("on_apps_delete_app_init")

	// Delete app remotely via Slack API
	err = clients.API().DeleteApp(ctx, config.GetContextToken(ctx), app.AppID)
	if err != nil {
		return app, err
	}
	log.Info("on_apps_delete_app_success")

	// Remove the saved app from project files
	log.Info("on_apps_remove_project")
	removedApp, err := clients.AppClient().Remove(ctx, app)
	if err != nil {
		return types.App{}, err
	}
	if removedApp.IsNew() {
		clients.IO.PrintDebug(
			ctx,
			"deleted app \"%s\" of team \"%s\" was not found in project files",
			app.AppID,
			app.TeamID,
		)
	}
	return app, nil
}

// getAuthSession return the api.AuthSession for the current auth
func getAuthSession(ctx context.Context, clients *shared.ClientFactory, auth types.SlackAuth) (context.Context, api.AuthSession, error) {
	// Should we be setting context token before we've validated it?
	var token = auth.Token
	ctx = config.SetContextToken(ctx, token)

	// Update the APIHost with the selected login, this is important for commands that use the Login to temporarily
	// get an auth without updating the default auth. It's less important for the Login command that terminals afterward,
	// because on start up, the root command resolves the auth's current APIHost.
	clients.Config.APIHostResolved = clients.Auth().ResolveAPIHost(ctx, clients.Config.APIHostFlag, &auth)
	clients.Config.LogstashHostResolved = clients.Auth().ResolveLogstashHost(ctx, clients.Config.APIHostResolved, clients.Config.Version)

	authSession, err := clients.API().ValidateSession(ctx, token)
	if err != nil {
		return ctx, api.AuthSession{}, slackerror.Wrap(err, slackerror.ErrInvalidAuth)
	}

	if authSession.UserID != nil {
		ctx = config.SetContextUserID(ctx, *authSession.UserID)
		clients.EventTracker.SetAuthUserID(*authSession.UserID)
	}

	return ctx, authSession, err
}
