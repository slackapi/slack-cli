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
	"fmt"
	"sort"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// List returns a list of the apps (only includes dev apps if there is a valid session)
func List(ctx context.Context, clients *shared.ClientFactory) ([]types.App, string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "pkg.apps.list")
	defer span.Finish()

	// List all deployed apps
	apps, defaultEnvName, err := clients.AppClient().GetDeployedAll(ctx)
	if err != nil {
		return []types.App{}, "", slackerror.Wrap(err, slackerror.ErrAppsList)
	}

	// List all local run apps
	devApps, err := clients.AppClient().GetLocalAll(ctx)
	if err != nil {
		return []types.App{}, "", slackerror.Wrap(err, slackerror.ErrAppsList)
	}
	// Update local run app names to include domain (when possible)
	for i, devApp := range devApps {
		if auth, err := clients.AuthInterface().AuthWithTeamID(ctx, devApp.TeamID); err == nil {
			// modify for display
			devApps[i].TeamDomain = fmt.Sprintf("%s %s", auth.TeamDomain, style.LocalRunNameTag)

			// Handle legacy apps.dev.json format
			// Legacy dev apps have team_domain as "dev"
			// instead of the team domain of the team they
			// were created in. Selector UI that relies on the
			// app's team domain will incorrectly display "dev"
			// So we override the TeamDomain when the auth
			// context is known and after we've confirmed that
			// auth.TeamID matches the app.TeamID
			devApp.TeamDomain = auth.TeamDomain
			_ = clients.AppClient().SaveLocal(ctx, devApp)
		}
	}

	apps = append(apps, devApps...)

	appsWithInstallStatus, err := FetchAppInstallStates(ctx, clients, apps)
	if err != nil {
		return nil, "", err
	}

	// Sort apps into alphabetical order by team domain
	sort.Slice(appsWithInstallStatus, func(i, j int) bool {
		return appsWithInstallStatus[i].TeamDomain < appsWithInstallStatus[j].TeamDomain
	})

	return appsWithInstallStatus, defaultEnvName, nil
}

// FetchAppInstallStates fetches app installation status from the backend and sets the values on the given apps
func FetchAppInstallStates(ctx context.Context, clients *shared.ClientFactory, apps []types.App) ([]types.App, error) {
	// Sort apps by team and ID
	appIDsByTeamID := map[string][]string{}
	appIDsByEnterpriseTeamID := map[string][]string{}
	appsByAppID := map[string]types.App{}
	for _, a := range apps {
		// Add app by its team id
		if appIDsByTeamID[a.TeamID] == nil {
			appIDsByTeamID[a.TeamID] = []string{}
		}
		appIDsByTeamID[a.TeamID] = append(appIDsByTeamID[a.TeamID], a.AppID)
		appsByAppID[a.AppID] = a

		// If app is an enterprise workspace app, also add it by its enterprise teamID
		if a.IsEnterpriseWorkspaceApp() {
			if appIDsByEnterpriseTeamID[a.EnterpriseID] == nil {
				appIDsByEnterpriseTeamID[a.EnterpriseID] = []string{}
			}
			appIDsByEnterpriseTeamID[a.EnterpriseID] = append(appIDsByEnterpriseTeamID[a.EnterpriseID], a.AppID)
			appsByAppID[a.AppID] = a
		}
	}

	// Get all available authed workspaces
	auths, err := clients.AuthInterface().Auths(ctx)
	if err != nil {
		return []types.App{}, err
	}
	if clients.Config.TokenFlag != "" {
		if tokenAuth, err := clients.AuthInterface().AuthWithToken(ctx, clients.Config.TokenFlag); err != nil {
			return []types.App{}, err
		} else {
			auths = append(auths, tokenAuth)
		}
	}

	appToInstallState := map[string]types.AppInstallationStatus{}
	appToEnterpriseGrants := map[string][]types.EnterpriseGrant{}
	for _, auth := range auths {

		if len(appIDsByTeamID[auth.TeamID]) == 0 && len(appIDsByEnterpriseTeamID[auth.TeamID]) == 0 {
			continue
		}

		apiClient := clients.APIInterface()
		if auth.APIHost != nil {
			// Most internal/api methods do not explicitly require the host to be set.
			// Rather, they rely implicitly on host being set on the apiClient when the instance
			// is created, (see internal/shared/clients.go). The value that the host is set
			// to today, in turn relies on the global clients.Config.APIHostResolved value
			// which in most cases is resolved once at the root of the command.
			//
			// For most cases, commands only require that a apiHost be set once. But in some cases,
			// such in list, we must potentially request to a different Slack API host for each of
			// the CLI's potential saved authorizations' apiHost values. (e.g. dev.slack.com, slack.com,
			// or number development instances such as dev123.slack.com)
			//
			// Refer to types.SlackAuth where we optionally represent this apiHost value.
			//
			// It is an anti-pattern to set and reset a global APIHostResolved value
			// for each authorization that we must make a Slack API call for, since:
			//
			//  1. setting global value impacts other future instances of apiClient
			//     at instantiation if the value is not reset correctly
			//  2. developers working on this codebase must have implicit knowledge
			//     about the way apiHost gets resolved to know to set and reset this global
			//     value
			//  3. Since each new apiClient that is instantiated will default to the existing
			//     global resolved host anyway, we can instead SetHost here without having to reset it
			//
			// So here we modify the host of this APIClient instance,
			// for each GetAppStatus (POST) request it makes.
			apiClient.SetHost(*auth.APIHost)
		}

		var appStatusResponse api.GetAppStatusResult
		if auth.IsEnterpriseInstall {
			var allApps = append(appIDsByEnterpriseTeamID[auth.TeamID], appIDsByTeamID[auth.TeamID]...)
			appStatusResponse, err = apiClient.GetAppStatus(ctx, auth.Token, allApps, auth.TeamID)
		} else {
			appStatusResponse, err = apiClient.GetAppStatus(ctx, auth.Token, appIDsByTeamID[auth.TeamID], auth.TeamID)
		}

		if err != nil {
			clients.IO.PrintDebug(ctx, "error fetching installation status for apps %v: %s", appIDsByTeamID[auth.TeamID], err.Error())
			continue
		}

		for _, a := range appStatusResponse.Apps {
			if a.Installed {
				appToInstallState[a.AppID] = types.AppStatusInstalled
			} else {
				appToInstallState[a.AppID] = types.AppStatusUninstalled
			}
			appToEnterpriseGrants[a.AppID] = a.EnterpriseGrants
		}
	}

	updatedApps := []types.App{}
	for _, a := range apps {
		if _, ok := appToInstallState[a.AppID]; ok {
			a.InstallStatus = appToInstallState[a.AppID]
		} else {
			a.InstallStatus = types.AppInstallationStatusUnknown
		}
		if _, ok := appToEnterpriseGrants[a.AppID]; ok {
			a.EnterpriseGrants = appToEnterpriseGrants[a.AppID]
		}
		updatedApps = append(updatedApps, a)
	}

	return updatedApps, nil
}
