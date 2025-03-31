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
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Uninstall will uninstall the app that belongs to the teamDomain from the backend.
// It will not modify the local project files (apps.json).
func Uninstall(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, teamDomain string, app types.App, auth types.SlackAuth) (types.App, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pkg.apps.uninstall")
	defer span.Finish()

	// Set selected team as the current team domain
	clients.Config.TeamFlag = teamDomain

	// Get Team Name
	ctx, authSession, err := getAuthSession(ctx, clients, auth)
	if err != nil {
		return types.App{}, slackerror.Wrap(err, slackerror.ErrAppRemove)
	}
	if authSession.TeamName != nil {
		log.Data["teamName"] = *authSession.TeamName
	}

	// Emit starting to remove app (requires teamName)
	log.Info("on_apps_uninstall_init")

	if app.AppID == "" {
		err = slackerror.New("app for team " + teamDomain + " not found")
		return types.App{}, slackerror.Wrap(err, slackerror.ErrAppRemove)
	}

	// Get token
	token := config.GetContextToken(ctx)

	// Uninstall the app
	log.Info("on_apps_uninstall_app_init")
	err = clients.ApiInterface().UninstallApp(ctx, token, app.AppID, app.TeamID)
	if err != nil {
		return app, err
	}
	log.Info("on_apps_uninstall_app_success")

	// Not modifying apps.json / apps.dev.json

	return app, nil
}
