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

package prompts

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/iostreams"
	authpkg "github.com/slackapi/slack-cli/internal/pkg/auth"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

type AppInstallStatus int

const (
	// ShowAllApps shows all existing and potential apps; one can choose from installed apps, uninstalled apps, or create a new app
	ShowAllApps AppInstallStatus = iota
	// ShowInstalledAppsOnly filters to show installed apps only
	ShowInstalledAppsOnly
	// ShowInstalledAndUninstalledApps filters to show both installed and uninstalled apps
	ShowInstalledAndUninstalledApps
	// ShowInstalledAndNewApps shows installed apps and allows the user to create a new app
	ShowInstalledAndNewApps
)

type AppEnvironmentType int

const (
	// ShowAllEnvironments does not filter apps by environment
	ShowAllEnvironments AppEnvironmentType = iota
	// ShowHostedOnly filters to show deployed apps only
	ShowHostedOnly
	// ShowLocalOnly filters to show locally run apps only
	//
	// Note: This check might not be correct when using the "app" and "token" flag
	// but without a saved apps.dev.json file. The ErrDeployedAppNotSupported error
	// will be returned with the selection in these cases.
	ShowLocalOnly
)

// Equals returns true if the app environment type is equal
func (environment AppEnvironmentType) Equals(is AppEnvironmentType) bool {
	return environment == is
}

// SelectedApp contains information for the user chosen app
type SelectedApp struct {
	// The workspace auth corresponding to the selected app
	Auth types.SlackAuth
	// The app that was chosen
	App types.App
}

// Equals returns true if the fields of the two SelectedApp objects match
func (apps *SelectedApp) Equals(otherApps SelectedApp) bool {
	if !apps.App.Equals(otherApps.App) {
		return false
	}
	if apps.Auth != otherApps.Auth {
		return false
	}
	return true
}

// IsEmpty returns true if the selection has no app and no auth
func (apps *SelectedApp) IsEmpty() bool {
	return apps.Equals(SelectedApp{})
}

// TeamApps contains the apps (local and hosted), auth and information for a team
type TeamApps struct {
	Auth   types.SlackAuth
	Hosted SelectedApp
	Local  SelectedApp
}

// Equals returns true if the fields of the two TeamApps objects match
func (apps *TeamApps) Equals(otherApps TeamApps) bool {
	if !apps.Hosted.Equals(otherApps.Hosted) {
		return false
	}
	if !apps.Local.Equals(otherApps.Local) {
		return false
	}
	if apps.Auth != otherApps.Auth {
		return false
	}
	return true
}

// IsEmpty returns true if the object is empty (all fields are their zero value)
func (apps *TeamApps) IsEmpty() bool {
	return apps.Equals(TeamApps{})
}

// authOrAppTeamDomain greedily returns a team domain corresponding to the TeamApps
// depending on what app and auth information is available.
//
// * When Auth is known, returns auth's team domain IFF it matches either the Hosted
// or local app's team domain.
//
// * When Auth is unknown or doesn't match the Hosted or Local app's team ID,
// then it returns app's team domain.
//
// Basically, treat as a convenience getter intended for use in the context where
// you have a team you want to filter against and you don't care whether it's an
// auth or an app that corresponds. E.g. when you are comparing to --team flags
func (t *TeamApps) authOrAppTeamDomain() string {
	if t.Auth.TeamDomain != "" && (t.Auth.TeamID == t.Hosted.App.TeamID || t.Auth.TeamID == t.Local.App.TeamID) {
		// Auth whose team id matches either hosted or local app's team
		// Can be safely returned
		return t.Auth.TeamDomain
	}

	// If we get here we might be missing an auth OR the auth doesn't
	// match any included app team ids (that is the case when the auth is org
	// resolved for a workspace app)
	if t.Hosted.App.TeamID != "" {
		return t.Hosted.App.TeamDomain
	}
	return t.Local.App.TeamDomain
}

// authOrAppTeamID greedily returns a team ID corresponding to the TeamApps
// depending on what app and auth information is available.
//
// * When Auth is known, returns auth's team domain IFF it matches either the Hosted
// or local app's team ID.
//
// * When Auth is unknown or doesn't match the Hosted or Local app's team ID,
// then it returns app's team ID.
//
// Basically, treat as a convenience getter intended for use in the context where
// you have a team you want to filter against and you don't care whether it's an
// auth or an app that corresponds. E.g. when you are comparing to --team flags
func (t *TeamApps) authOrAppTeamID() string {
	if t.Auth.TeamID != "" && (t.Hosted.App.TeamID == t.Auth.TeamID || t.Local.App.TeamID == t.Auth.TeamID) {
		return t.Auth.TeamID
	}
	if t.Hosted.App.TeamID != "" {
		return t.Hosted.App.TeamID
	}
	return t.Local.App.TeamID
}

// appTransferDisclaimer contains a notice of lost app management permissions
// if the installed workspace is left
var appTransferDisclaimer = style.TextSection{
	Emoji: "bell",
	Text:  "If you leave this team, you can no longer manage the installed apps",
	Secondary: []string{
		"Installed apps will belong to the team if you leave the workspace",
	},
}

var appInstallPromptNew = "Install to a new team"
var SelectTeamPrompt = "Select a team"

// getApps returns the apps saved to files with known credentials
//
// We start with the known authentications and saved apps. Then details for the
// saved apps without credentials are gathered. Sometimes these have enterprise
// authentications that can be resolved. Installation status is determined near
// the end of processing.
func getApps(ctx context.Context, clients *shared.ClientFactory) (map[string]SelectedApp, error) {
	appIDs := map[string]SelectedApp{}
	allAuths, err := getAuths(ctx, clients)
	if err != nil {
		return nil, err
	}
	deployedApps, _, err := clients.AppClient().GetDeployedAll(ctx)
	if err != nil {
		return nil, err
	}
	localApps, err := clients.AppClient().GetLocalAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, auth := range allAuths {
		app, err := clients.AppClient().GetDeployed(ctx, auth.TeamID)
		if err != nil {
			return nil, err
		}
		selection := SelectedApp{
			Auth: auth,
			App:  app,
		}
		if appExists(selection.App) {
			appIDs[selection.App.AppID] = selection
		}
		app, err = clients.AppClient().GetLocal(ctx, auth.TeamID)
		if err != nil {
			return nil, err
		}
		selection = SelectedApp{
			Auth: auth,
			App:  app,
		}
		if appExists(selection.App) {
			appIDs[selection.App.AppID] = selection
		}
	}
	for _, saved := range deployedApps {
		if appIDs[saved.AppID].App.AppID != "" {
			continue
		}
		resolvedAuth, err := clients.Auth().AuthWithTeamID(ctx, saved.TeamID)
		if err != nil {
			if slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
				return map[string]SelectedApp{}, err
			}
			if saved.IsEnterpriseWorkspaceApp() {
				resolvedAuth, err = clients.Auth().AuthWithTeamID(ctx, saved.EnterpriseID)
				if err != nil && slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
					return map[string]SelectedApp{}, err
				}
				logResolvedEnterpriseAuth(ctx, clients, saved, resolvedAuth)
			}
		}
		selection := SelectedApp{
			Auth: resolvedAuth,
			App:  saved,
		}
		appIDs[selection.App.AppID] = selection
	}
	for _, saved := range localApps {
		if appIDs[saved.AppID].App.AppID != "" {
			continue
		}
		resolvedAuth, err := clients.Auth().AuthWithTeamID(ctx, saved.TeamID)
		if err != nil {
			if slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
				return map[string]SelectedApp{}, err
			}
			if saved.IsEnterpriseWorkspaceApp() {
				resolvedAuth, err = clients.Auth().AuthWithTeamID(ctx, saved.EnterpriseID)
				if err != nil && slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
					return map[string]SelectedApp{}, err
				}
				logResolvedEnterpriseAuth(ctx, clients, saved, resolvedAuth)
			}
		}
		selection := SelectedApp{
			Auth: resolvedAuth,
			App:  saved,
		}
		appIDs[selection.App.AppID] = selection
	}
	teamIDToAppIDs := map[string][]SelectedApp{}
	for _, app := range appIDs {
		if len(teamIDToAppIDs[app.App.TeamID]) > 0 {
			teamIDToAppIDs[app.App.TeamID] = append(teamIDToAppIDs[app.App.TeamID], app)
		} else {
			teamIDToAppIDs[app.App.TeamID] = []SelectedApp{app}
		}
	}
	for _, apps := range teamIDToAppIDs {
		if len(apps) <= 0 {
			continue
		}
		auth := apps[0].Auth
		apiHost := ""
		if auth.APIHost != nil {
			apiHost = *auth.APIHost
		}
		ids := []string{}
		for _, app := range apps {
			ids = append(ids, app.App.AppID)
		}
		statuses, err := getInstallationStatuses(ctx, clients, auth.Token, ids, auth.TeamID, apiHost)
		if err != nil {
			clients.IO.PrintDebug(
				ctx,
				"error fetching installation status for the following %s in team %s %v: %s",
				style.Pluralize("app", "apps", len(appIDs)),
				auth.TeamDomain,
				appIDs,
				err.Error(),
			)
		}
		for _, status := range statuses {
			app := appIDs[status.AppID]
			app.App.EnterpriseGrants = status.EnterpriseGrants
			app.App.InstallStatus = status.InstallationState
			appIDs[status.AppID] = app
		}
	}
	return appIDs, nil
}

// getAuths returns the available authentications for the selection
func getAuths(ctx context.Context, clients *shared.ClientFactory) ([]types.SlackAuth, error) {
	allAuths, err := clients.Auth().Auths(ctx)
	if err != nil {
		return nil, err
	}
	if clients.Config.TokenFlag != "" {
		tokenAuth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return []types.SlackAuth{}, err
		}
		_, err = clients.Auth().AuthWithTeamID(ctx, tokenAuth.TeamID)
		if err != nil {
			if slackerror.ToSlackError(err).Code == slackerror.ErrCredentialsNotFound {
				allAuths = append(allAuths, tokenAuth)
			} else {
				return []types.SlackAuth{}, err
			}
		}
	}
	if len(allAuths) == 0 {
		auth := types.SlackAuth{}
		err := validateAuth(ctx, clients, &auth)
		if err != nil {
			return nil, slackerror.New(slackerror.ErrNotAuthed)
		}
		allAuths = append(allAuths, auth)
	}
	return allAuths, nil
}

// getTeamApps creates a map from team ID to applications and authentications
//
// Details are collected from both the credentials.json and apps.*.json files.
// Existing apps or placeholder apps will be added to each team. Apps without
// authentications and authentications with placeholder apps are both possible.
//
// Additional filters should check that apps and authentications meet criteria.
//
// Note: Providing both the "app" and "token" flag should skip this check and
// instead use the flag options provided.
func getTeamApps(ctx context.Context, clients *shared.ClientFactory) (map[string]TeamApps, error) {
	teamApps := map[string]TeamApps{}

	allAuths, err := getAuths(ctx, clients)
	if err != nil {
		return nil, err
	}

	// Get deployed and local apps
	deployedApps, _, err := clients.AppClient().GetDeployedAll(ctx)
	if err != nil {
		return nil, err
	}
	localApps, err := clients.AppClient().GetLocalAll(ctx)
	if err != nil {
		return nil, err
	}

	// First add all the apps for the authed teams
	for _, auth := range allAuths {
		deployedApp, err := clients.AppClient().GetDeployed(ctx, auth.TeamID)
		if err != nil {
			return nil, err
		}
		hostedApp := SelectedApp{
			Auth: auth,
			App:  deployedApp,
		}

		app, err := clients.AppClient().GetLocal(ctx, auth.TeamID)
		if err != nil {
			return nil, err
		}
		localApp := SelectedApp{
			Auth: auth,
			App:  app,
		}

		if localApp.App.TeamDomain == "dev" && auth.TeamID == app.TeamID {
			// Handle legacy apps.dev.json format
			// Legacy dev apps have team_domain as "dev"
			// instead of the team domain of the team they
			// were created in. Selector UI that relies on the
			// app's team domain will incorrectly display "dev"
			// So we override the TeamDomain when the auth
			// context is known and after we've confirmed that
			// auth.TeamID matches the app.TeamID
			localApp.App.TeamDomain = auth.TeamDomain
			_ = clients.AppClient().SaveLocal(ctx, localApp.App)
		}

		selection := TeamApps{
			Auth:   auth,
			Hosted: hostedApp,
			Local:  localApp,
		}

		// Fetch installation status for the apps
		selection = appendAppInstallStatus(ctx, clients, auth, selection)

		teamApps[selection.authOrAppTeamID()] = selection
	}

	// Then add remaining "hosted" apps that did not have saved credentials and
	// were therefore not saved to team apps
	for _, deployedApp := range deployedApps {
		if teamApps[deployedApp.TeamID].Hosted.App.AppID != "" {
			continue
		}
		var resolvedAuth types.SlackAuth

		// Try to find an auth that matches the app's team id
		_, err := clients.Auth().AuthWithTeamID(ctx, deployedApp.TeamID)
		if err == nil {
			continue
		} else {
			if slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
				return map[string]TeamApps{}, err
			}
			if deployedApp.IsEnterpriseWorkspaceApp() {
				// We can search to see whether we can find an existing org-level auth
				// in credentials.json. If found, use that auth as the resolved auth
				// for this workspace app

				resolvedAuth, err = clients.Auth().AuthWithTeamID(ctx, deployedApp.EnterpriseID)
				if err != nil && slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
					// Fetching an auth by team id failed for some other reason than credentials not being found
					return map[string]TeamApps{}, err
				}

				// Do some debug logging
				logResolvedEnterpriseAuth(ctx, clients, deployedApp, resolvedAuth)
			}
		}
		// Set the Auth as resolved Auth
		hostedApp := SelectedApp{
			Auth: resolvedAuth,
			App:  deployedApp,
		}

		// Create a dummy local app
		localApp := SelectedApp{
			Auth: resolvedAuth,
			App:  types.App{TeamID: deployedApp.TeamID},
		}

		// Assume the caller is a collaborator of the local app
		for _, a := range localApps {
			if a.TeamID == deployedApp.TeamID {
				localApp.App = a
			}
		}

		selection := TeamApps{
			Auth:   resolvedAuth,
			Hosted: hostedApp,
			Local:  localApp,
		}

		selection = appendAppInstallStatus(ctx, clients, resolvedAuth, selection)

		teamApps[selection.authOrAppTeamID()] = selection
	}

	// Then add remaining "local" apps that did not have saved credentials and
	// were therefore not saved to team apps
	for _, localApp := range localApps {
		if teamApps[localApp.TeamID].Local.App.AppID != "" {
			continue
		}
		var resolvedAuth types.SlackAuth

		_, err = clients.Auth().AuthWithTeamID(ctx, localApp.TeamID)
		if err == nil {
			continue
		} else {
			if slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
				return map[string]TeamApps{}, err
			}

			if localApp.IsEnterpriseWorkspaceApp() {
				// We can search to see whether we can find an existing org-level auth
				// in credentials.json. If found, use that auth as the resolved auth
				// for this workspace app

				resolvedAuth, err = clients.Auth().AuthWithTeamID(ctx, localApp.EnterpriseID)

				if err != nil && slackerror.ToSlackError(err).Code != slackerror.ErrCredentialsNotFound {
					// Fetching an auth by team id failed for some other reason than credentials not being found
					return map[string]TeamApps{}, err
				}

				logResolvedEnterpriseAuth(ctx, clients, localApp, resolvedAuth)
			}
		}
		// Don't override any existing Hosted / deployed selections or auth
		var existingHosted SelectedApp
		var existingAuth types.SlackAuth

		_, ok := teamApps[localApp.TeamID]

		if ok {
			existingHosted = teamApps[localApp.TeamID].Hosted
			existingAuth = teamApps[localApp.TeamID].Auth
		}

		newLocal := SelectedApp{
			Auth: resolvedAuth,
			App:  localApp,
		}

		selection := TeamApps{
			Auth:   existingAuth,
			Hosted: existingHosted,
			Local:  newLocal,
		}

		selection = appendAppInstallStatus(ctx, clients, resolvedAuth, selection)

		teamApps[selection.authOrAppTeamID()] = selection
	}

	return teamApps, nil
}

// getTokenApp gathers app and auth info from the API using the token and app ID
func getTokenApp(ctx context.Context, clients *shared.ClientFactory, token string, appID string) (SelectedApp, error) {
	auth, err := clients.Auth().AuthWithToken(ctx, token)
	if err != nil {
		return SelectedApp{}, err
	}
	var appStatus api.AppStatusResultAppInfo
	if appStatusResult, err := clients.API().GetAppStatus(ctx, token, []string{appID}, auth.TeamID); err != nil {
		return SelectedApp{}, err
	} else if len(appStatusResult.Apps) != 1 || appStatusResult.Apps[0].AppID != appID {
		return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound)
	} else {
		appStatus = appStatusResult.Apps[0]
	}
	app := types.App{
		AppID:        appID,
		EnterpriseID: auth.EnterpriseID,
		TeamDomain:   auth.TeamDomain,
		TeamID:       auth.TeamID,
		UserID:       auth.UserID, // Set for applications not saved to apps.dev.json
	}
	if appStatus.Installed {
		app.InstallStatus = types.AppStatusInstalled
	} else {
		app.InstallStatus = types.AppStatusUninstalled
	}
	switch clients.Config.TeamFlag {
	case "", auth.TeamDomain, auth.TeamID:
		break
	default:
		return SelectedApp{}, slackerror.New(slackerror.ErrTeamNotFound)
	}
	saved, err := clients.AppClient().GetLocal(ctx, app.TeamID)
	if err == nil && saved.AppID == app.AppID {
		app.IsDev = true
	}
	return SelectedApp{Auth: auth, App: app}, nil
}

// appendAppInstallStatus gets and appends an apps installation status to the selections
func appendAppInstallStatus(ctx context.Context, clients *shared.ClientFactory, auth types.SlackAuth, selection TeamApps) TeamApps {
	appIDs := []string{}
	if appExists(selection.Local.App) {
		appIDs = append(appIDs, selection.Local.App.AppID)
	}
	if appExists(selection.Hosted.App) {
		appIDs = append(appIDs, selection.Hosted.App.AppID)
	}
	if len(appIDs) > 0 {
		var apiHost = ""
		if auth.APIHost != nil {
			apiHost = *auth.APIHost
		}
		appInfo, err := getInstallationStatuses(ctx, clients, auth.Token, appIDs, auth.TeamID, apiHost)
		if err != nil {
			clients.IO.PrintDebug(ctx, "error fetching installation status for the following %s in team %s %v: %s", style.Pluralize("app", "apps", len(appIDs)), auth.TeamDomain, appIDs, err.Error())
		}
		for _, i := range appInfo {
			if i.AppID == selection.Local.App.AppID {
				selection.Local.App.InstallStatus = i.InstallationState
				selection.Local.App.EnterpriseGrants = i.EnterpriseGrants
			}
			if i.AppID == selection.Hosted.App.AppID {
				selection.Hosted.App.InstallStatus = i.InstallationState
				selection.Hosted.App.EnterpriseGrants = i.EnterpriseGrants
			}
		}
	}
	return selection
}

// logResolvedEnterpriseAuth logs out which org auth was resolved for which enterprise workspace app
func logResolvedEnterpriseAuth(ctx context.Context, clients *shared.ClientFactory, app types.App, resolvedAuth types.SlackAuth) {
	clients.IO.PrintDebug(ctx, "Workspace token missing for Enterprise Workspace App ID: %s, Team: %s (%s)", app.AppID, app.TeamDomain, app.TeamID)
	clients.IO.PrintDebug(ctx, "Workspace: %s (%s) belongs to Enterprise: %s (%s)", app.TeamDomain, app.TeamID, resolvedAuth.TeamDomain, app.EnterpriseID)
	clients.IO.PrintDebug(ctx, "OK to resolve org token for %s (%s) to use with App ID: %s", resolvedAuth.TeamDomain, resolvedAuth.TeamID, app.AppID)
}

type AppStatus struct {
	AppID             string
	InstallationState types.AppInstallationStatus
	EnterpriseGrants  []types.EnterpriseGrant
}

// getInstallationStatuses fetches installation states for the apps.
func getInstallationStatuses(ctx context.Context, clients *shared.ClientFactory, token string, appIDs []string, teamID string, apiHost string) ([]AppStatus, error) {
	startTimer := time.Now()

	// Ensure that the client's host in this case is set to any apiHost provided
	apiClient := clients.API()
	if apiHost != "" {
		apiClient.SetHost(apiHost)
	}

	// Get the app status of appIDs that are sorted for stable mocking in tests
	slices.Sort(appIDs)
	appStatusResponse, err := apiClient.GetAppStatus(ctx, token, appIDs, teamID)
	if err != nil {
		return nil, err
	}
	clients.IO.PrintDebug(ctx, "GetAppStatus request for team %s took: %v\n", teamID, time.Since(startTimer).Round(time.Millisecond))

	appInfos := []AppStatus{}
	for _, a := range appStatusResponse.Apps {
		var installState types.AppInstallationStatus
		if a.Installed {
			installState = types.AppStatusInstalled
		} else {
			installState = types.AppStatusUninstalled
		}
		info := AppStatus{AppID: a.AppID, InstallationState: installState, EnterpriseGrants: a.EnterpriseGrants}
		appInfos = append(appInfos, info)
	}

	return appInfos, nil
}

// filterAuthsByToken returns any matching workspace authentication for the token flag
func filterAuthsByToken(ctx context.Context, clients *shared.ClientFactory, workspaceApps map[string]TeamApps) (types.SlackAuth, error) {
	var teamFlag = clients.Config.TeamFlag // team_id, domain of a workspace or an org, i.e. T123445678, 'acme-corp', 'acme-workspace'
	var appFlag = clients.Config.AppFlag   // an app_id, local, deploy, or deployed

	// Fetch an existing Auth that matches supplied token OR return a brand new Auth object
	auth, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
	if err != nil {
		return types.SlackAuth{}, err
	}

	var teamFlagIsAuthTeamDomain = auth.TeamDomain != "" && teamFlag == auth.TeamDomain
	var teamFlagIsAuthTeamID = auth.TeamID != "" && teamFlag == auth.TeamID

	if teamFlag != "" && !teamFlagIsAuthTeamDomain && !teamFlagIsAuthTeamID {
		// If a team flag is provided and it doesn't match either auth team domain or auth team id
		return types.SlackAuth{}, slackerror.New(slackerror.ErrInvalidToken).WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{Message: "The team flag is not associated with the provided token"},
		})
	}

	if types.IsAppID(appFlag) {
		filtered, err := filterByAppID(workspaceApps, appFlag)
		if err != nil {
			return types.SlackAuth{}, err
		}

		if auth.TeamID != "" && (auth.TeamID != filtered.App.TeamID) && (auth.TeamID != filtered.App.EnterpriseID) {
			// Auth team domain and app team domain don't match
			return types.SlackAuth{}, slackerror.New(slackerror.ErrInvalidToken).WithDetails(slackerror.ErrorDetails{
				slackerror.ErrorDetail{Message: "The app flag is not associated with the provided token"},
			})
		}
	}
	return auth, nil
}

// filterByAppID returns the app with a matching appID and errors if it cannot find an app by that ID
func filterByAppID(workspaceApps map[string]TeamApps, appID string) (SelectedApp, error) {
	for _, selection := range workspaceApps {
		if selection.Hosted.App.AppID == appID {
			return selection.Hosted, nil
		}
		if selection.Local.App.AppID == appID {
			return selection.Local, nil
		}
	}
	return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound)
}

// filterByTeamFlag returns the possible combos of auth and apps that correspond to a given team flag.
// Team flag can be team_domain or team_id
func filterByTeamFlag(workspaceApps map[string]TeamApps, teamFlag string) (TeamApps, error) {

	for _, selection := range workspaceApps {
		if teamFlag == selection.authOrAppTeamDomain() || teamFlag == selection.authOrAppTeamID() {

			return selection, nil
		}
	}
	return TeamApps{}, slackerror.New(slackerror.ErrTeamNotFound)
}

// selectAppWorkspace prompts for a choice of workspace with status apps
func selectAppWorkspace(ctx context.Context, clients *shared.ClientFactory, workspaceApps map[string]TeamApps, status AppInstallStatus) (TeamApps, error) {
	teamIDsWithApps, domainsWithApps, domainsWithAppsLabels := []string{}, []string{}, []string{}
	unusedTeamIDs, unusedDomains, unusedDomainsLabels := []string{}, []string{}, []string{}

	for _, workspace := range workspaceApps {

		if includeInAppSelect(workspace.Hosted.App, status) {
			domainsWithApps = append(domainsWithApps, workspace.Hosted.App.TeamDomain)
			teamIDsWithApps = append(teamIDsWithApps, workspace.Hosted.App.TeamID)

			// to show the user
			domainsWithAppsLabels = append(domainsWithAppsLabels, style.TeamSelectLabel(workspace.Hosted.App.TeamDomain, workspace.Hosted.App.TeamID))
			continue
		} else if includeInAppSelect(workspace.Local.App, status) {
			domainsWithApps = append(domainsWithApps, workspace.Local.App.TeamDomain)
			teamIDsWithApps = append(teamIDsWithApps, workspace.Local.App.TeamID)

			// to show the user
			domainsWithAppsLabels = append(domainsWithAppsLabels, style.TeamSelectLabel(workspace.Local.App.TeamDomain, workspace.Local.App.TeamID))
			continue
		}
		if workspace.Auth.TeamID != "" {
			unusedDomains = append(unusedDomains, workspace.Auth.TeamDomain)
			unusedTeamIDs = append(unusedTeamIDs, workspace.Auth.TeamID)
			unusedDomainsLabels = append(unusedDomainsLabels, style.TeamSelectLabel(workspace.Auth.TeamDomain, workspace.Auth.TeamID))
		}
	}

	if (status == ShowInstalledAppsOnly || status == ShowInstalledAndUninstalledApps) && len(domainsWithApps) == 0 {
		return TeamApps{}, slackerror.New(slackerror.ErrInstallationRequired)
	}

	// Prepare the options for the user
	if err := SortAlphaNumeric(domainsWithApps, domainsWithAppsLabels, teamIDsWithApps); err != nil {
		return TeamApps{}, err
	}
	if err := SortAlphaNumeric(unusedDomains, unusedDomainsLabels, unusedTeamIDs); err != nil {
		return TeamApps{}, err
	}

	// Optionally, add prompt for creating new app
	if status == ShowAllApps || status == ShowInstalledAndNewApps && len(unusedDomains) > 0 {
		installPrompt := style.Secondary(appInstallPromptNew)
		domainsWithAppsLabels = append(domainsWithAppsLabels, installPrompt)
	}

	var selectedTeamID string

	// Get the users selection of team
	selection, err := clients.IO.SelectPrompt(ctx, SelectTeamPrompt, domainsWithAppsLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("team"),
		Required: true,
	})
	if err != nil {
		return TeamApps{}, err
	}

	// If a flag was used, return the team selection by the flag
	if selection.Flag {
		return getTeamByFlag(selection.Option, workspaceApps, status)
	}

	if selection.Index < len(domainsWithApps) {
		// Selected an existing domain
		selectedTeamID = teamIDsWithApps[selection.Index]
	} else {
		// Somehow did not select an existing app domain
		// Would we ever fall into this section of code?
		selection, err = clients.IO.SelectPrompt(ctx, appInstallPromptNew, unusedDomainsLabels, iostreams.SelectPromptConfig{
			Required: true,
		})
		if err != nil {
			return TeamApps{}, err
		}
		if selection.Prompt {
			selectedTeamID = unusedTeamIDs[selection.Index]
		}
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(appTransferDisclaimer))
	}

	return workspaceApps[selectedTeamID], nil
}

// getTeamByFlag returns the matching workspace with a given status for a flag
func getTeamByFlag(flag string, workspaceApps map[string]TeamApps, status AppInstallStatus) (TeamApps, error) {
	var selectedTeam TeamApps
	for _, team := range workspaceApps {
		if team.authOrAppTeamID() == flag || team.authOrAppTeamDomain() == flag {
			selectedTeam = team
		}
	}
	if selectedTeam.IsEmpty() {
		return TeamApps{}, slackerror.New(slackerror.ErrCredentialsNotFound)
	}
	if (!selectedTeam.Hosted.IsEmpty() || !selectedTeam.Local.IsEmpty()) && status == ShowAllApps {
		return selectedTeam, nil
	}
	if ((!selectedTeam.Hosted.IsEmpty() && includeInAppSelect(selectedTeam.Hosted.App, status)) ||
		(selectedTeam.Local.IsEmpty() && includeInAppSelect(selectedTeam.Local.App, status))) &&
		(status == ShowInstalledAppsOnly || status == ShowInstalledAndUninstalledApps || status == ShowInstalledAndNewApps) {
		return selectedTeam, nil
	}
	return TeamApps{}, slackerror.New(slackerror.ErrAppNotFound)
}

// includeInAppSelect is a helper function for app selection that determines whether an existing app should be displayed given the install status
func includeInAppSelect(app types.App, status AppInstallStatus) bool {

	// Require that the app exists
	if !appExists(app) {
		return false
	}

	// App exists but we're unable to fetch installation status from backend
	if app.InstallStatus == types.AppInstallationStatusUnknown {
		return true
	}

	switch status {
	case ShowAllApps: // Status allows both installed and uninstalled apps; return true
		return true
	case ShowInstalledAndUninstalledApps: // Status allows both installed and uninstalled apps; return true
		return true
	case ShowInstalledAndNewApps: // Status indicates that uninstalled apps should be excluded
		return app.InstallStatus != types.AppStatusUninstalled
	case ShowInstalledAppsOnly: // Status indicates that only installed apps should be included
		return app.InstallStatus == types.AppStatusInstalled
	}

	return false
}

// showOptionsForNewAppCreation determines if an app should be shown in creation
func showOptionsForNewAppCreation(app types.App, status AppInstallStatus) bool {
	return !appExists(app) && (status == ShowAllApps || status == ShowInstalledAndNewApps)
}

// selectAppEnvironment prompts for the environment choice of a status app
func selectAppEnvironment(ctx context.Context, clients *shared.ClientFactory, workspace TeamApps, status AppInstallStatus) (SelectedApp, error) {

	showLocalApp := includeInAppSelect(workspace.Local.App, status) || showOptionsForNewAppCreation(workspace.Local.App, status)
	showDeployedApp := includeInAppSelect(workspace.Hosted.App, status) || showOptionsForNewAppCreation(workspace.Hosted.App, status)
	appOptions, appLabels := []SelectedApp{}, []string{}

	if showLocalApp {
		appOptions = append(appOptions, workspace.Local)
		appLabels = append(appLabels, style.AppSelectLabel("Local", workspace.Local.App.AppID, workspace.Local.App.IsUninstalled()))
	}
	if showDeployedApp {
		appOptions = append(appOptions, workspace.Hosted)
		appLabels = append(appLabels, style.AppSelectLabel("Deployed", workspace.Hosted.App.AppID, workspace.Hosted.App.IsUninstalled()))
	}

	var selectedApp SelectedApp
	selection, err := clients.IO.SelectPrompt(ctx, "Choose an app environment", appLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("app"),
		Required: true,
	})
	if err != nil {
		return SelectedApp{}, err
	} else if selection.Flag {
		switch {
		case showLocalApp && types.IsAppFlagLocal(selection.Option):
			selectedApp = workspace.Local
		case showDeployedApp && types.IsAppFlagDeploy(selection.Option):
			selectedApp = workspace.Hosted
		default:
			return SelectedApp{}, slackerror.New(slackerror.ErrInvalidAppFlag)
		}
	} else if selection.Prompt {
		selectedApp = appOptions[selection.Index]
	}
	return selectedApp, nil
}

type workspaceOptions struct {
	environment   AppEnvironmentType
	installStatus AppInstallStatus
}

// selectTeamApp prompts user to select a SelectedApp matching the provided options
func selectTeamApp(ctx context.Context, clients *shared.ClientFactory, workspaceApps map[string]TeamApps, options workspaceOptions) (string, SelectedApp, error) {

	includeApp := func(app types.App) bool {
		if app.IsDev && options.environment == ShowHostedOnly || !app.IsDev && options.environment == ShowLocalOnly {
			return false
		}
		return includeInAppSelect(app, options.installStatus)
	}

	teamIDsWithApps, domainsWithApps, domainsWithAppsLabels := []string{}, []string{}, []string{}
	unusedTeamIDs, unusedDomains, unusedDomainsLabels := []string{}, []string{}, []string{}
	for _, workspace := range workspaceApps {
		if includeApp(workspace.Hosted.App) {

			teamIDsWithApps = append(teamIDsWithApps, workspace.Hosted.App.TeamID)
			domainsWithApps = append(domainsWithApps, workspace.Hosted.App.TeamDomain)
			domainsWithAppsLabels = append(domainsWithAppsLabels,
				style.TeamAppSelectLabel(workspace.Hosted.App.TeamDomain, workspace.Hosted.App.TeamID, workspace.Hosted.App.AppID, workspace.Hosted.App.IsUninstalled()))

		} else if includeApp(workspace.Local.App) {

			teamIDsWithApps = append(teamIDsWithApps, workspace.Local.App.TeamID)
			domainsWithApps = append(domainsWithApps, workspace.Local.App.TeamDomain)
			domainsWithAppsLabels = append(domainsWithAppsLabels,
				style.TeamAppSelectLabel(workspace.Local.App.TeamDomain, workspace.Local.App.TeamID, workspace.Local.App.AppID, workspace.Local.App.IsUninstalled()))
		} else {
			// Neither hosted nor local app can be included, so this teamApp's remaining auth domain is unused
			if workspace.Auth.TeamID != "" {
				unusedTeamIDs = append(unusedTeamIDs, workspace.Auth.TeamID)
				unusedDomains = append(unusedDomains, workspace.Auth.TeamDomain)
				unusedDomainsLabels = append(unusedDomainsLabels, style.TeamSelectLabel(workspace.Auth.TeamDomain, workspace.Auth.TeamID))
			}
		}
	}

	if options.installStatus == ShowInstalledAppsOnly && len(domainsWithApps) == 0 {
		return "", SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
	}

	if err := SortAlphaNumeric(domainsWithApps, domainsWithAppsLabels, teamIDsWithApps); err != nil {
		return "", SelectedApp{}, err
	}
	if err := SortAlphaNumeric(unusedDomains, unusedDomainsLabels, unusedTeamIDs); err != nil {
		return "", SelectedApp{}, err
	}

	// Add prompt for creating new apps if unused domains exist
	if len(unusedDomains) > 0 && (options.installStatus == ShowAllApps || options.installStatus == ShowInstalledAndNewApps) {
		installPrompt := style.Secondary(appInstallPromptNew)
		domainsWithAppsLabels = append(domainsWithAppsLabels, installPrompt)
	}

	switch options.environment {
	case ShowLocalOnly:
		SelectTeamPrompt = "Choose a local environment"
	case ShowHostedOnly:
		SelectTeamPrompt = "Choose a deployed environment"
	}

	var selectedDomain string
	var selectedTeamID string

	selection, err := clients.IO.SelectPrompt(ctx, SelectTeamPrompt, domainsWithAppsLabels, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("team"),
		Required: true,
	})
	if err != nil {
		return "", SelectedApp{}, err
	} else if selection.Flag {

		if selectedTeam, err := getTeamByFlag(selection.Option, workspaceApps, options.installStatus); err != nil {
			return "", SelectedApp{}, err
		} else {
			selectedDomain = selectedTeam.authOrAppTeamDomain()
			selectedTeamID = selectedTeam.authOrAppTeamID()
		}
	} else if selection.Prompt && selection.Index < len(domainsWithApps) {
		selectedDomain = domainsWithApps[selection.Index]
		selectedTeamID = teamIDsWithApps[selection.Index]
	} else if selection.Prompt {
		selection, err = clients.IO.SelectPrompt(ctx, appInstallPromptNew, unusedDomainsLabels, iostreams.SelectPromptConfig{
			Flag:     clients.Config.Flags.Lookup("team"),
			Required: true,
		})
		if err != nil {
			return "", SelectedApp{}, err
		} else if selection.Flag {
			if selectedTeam, err := getTeamByFlag(selection.Option, workspaceApps, options.installStatus); err != nil {
				return "", SelectedApp{}, err
			} else {
				selectedDomain = selectedTeam.authOrAppTeamDomain()
				selectedTeamID = selectedTeam.authOrAppTeamID()
			}
		} else if selection.Prompt {
			selectedDomain = unusedDomains[selection.Index]
			selectedTeamID = unusedTeamIDs[selection.Index]
		}
		clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(appTransferDisclaimer))
	}

	workspace := workspaceApps[selectedTeamID]

	// Depending on the environment, return the SelectedApp that matches
	var selectedApp SelectedApp
	switch options.environment {
	case ShowLocalOnly:
		selectedApp = workspace.Local
	case ShowHostedOnly:
		selectedApp = workspace.Hosted
	}

	return selectedDomain, selectedApp, nil
}

// flatAppSelectPrompt reveals options for apps that match the install status
func flatAppSelectPrompt(
	ctx context.Context,
	clients *shared.ClientFactory,
	environment AppEnvironmentType,
	status AppInstallStatus,
) (
	selected SelectedApp,
	err error,
) {
	switch {
	case environment.Equals(ShowAllEnvironments) && types.IsAppFlagEnvironment(clients.Config.AppFlag):
		switch {
		case types.IsAppFlagDeploy(clients.Config.AppFlag):
			return flatAppSelectPrompt(ctx, clients, ShowHostedOnly, status)
		case types.IsAppFlagLocal(clients.Config.AppFlag):
			return flatAppSelectPrompt(ctx, clients, ShowLocalOnly, status)
		}
	case environment.Equals(ShowLocalOnly) && types.IsAppFlagDeploy(clients.Config.AppFlag):
		return SelectedApp{}, slackerror.New(slackerror.ErrDeployedAppNotSupported)
	case environment.Equals(ShowHostedOnly) && types.IsAppFlagLocal(clients.Config.AppFlag):
		return SelectedApp{}, slackerror.New(slackerror.ErrLocalAppNotSupported)
	}
	defer func() {
		if err != nil {
			return
		}
		err = validateAuth(ctx, clients, &selected.Auth)
		if err != nil {
			return
		}
		clients.Auth().SetSelectedAuth(ctx, selected.Auth, clients.Config, clients.Os)
		if selected.App.IsNew() {
			clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(appTransferDisclaimer))
		}
	}()
	if clients.Config.SkipLocalFs() {
		selection, err := getTokenApp(ctx, clients, clients.Config.TokenFlag, clients.Config.AppFlag)
		if err != nil {
			return SelectedApp{}, err
		}
		if status == ShowInstalledAppsOnly && selection.App.InstallStatus != types.AppStatusInstalled {
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}
		clients.Auth().SetSelectedAuth(ctx, selection.Auth, clients.Config, clients.Os)
		// The development status of an app cannot be determined when local app files
		// do not exist. This defaults to "false" for these cases.
		//
		// Commands such as "platform run" might allow unknown development statuses so
		// we return both the selection and an error here.
		if selection.App.IsDev && environment == ShowHostedOnly {
			return selection, slackerror.New(slackerror.ErrLocalAppNotSupported)
		} else if !selection.App.IsDev && environment == ShowLocalOnly {
			return selection, slackerror.New(slackerror.ErrDeployedAppNotSupported)
		}
		return selection, nil
	}
	appIDs, err := getApps(ctx, clients)
	if err != nil {
		return SelectedApp{}, err
	}
	// teamFlag is set to a team ID if either the --team or --token flag is set
	teamFlag := ""
	if clients.Config.TokenFlag != "" {
		token, err := clients.Auth().AuthWithToken(ctx, clients.Config.TokenFlag)
		if err != nil {
			return SelectedApp{}, err
		}
		switch clients.Config.TeamFlag {
		case "", token.TeamID, token.TeamDomain:
			teamFlag = token.TeamID
		default:
			return SelectedApp{}, slackerror.New(slackerror.ErrTeamNotFound)
		}
	} else if clients.Config.TeamFlag != "" {
		teamFlag = clients.Config.TeamFlag
	}
	filtered := map[string]SelectedApp{}
	for id, app := range appIDs {
		switch environment {
		case ShowAllEnvironments:
		case ShowHostedOnly:
			if app.App.IsDev {
				continue
			}
		case ShowLocalOnly:
			if !app.App.IsDev {
				continue
			}
		}
		switch teamFlag {
		case "":
		case app.App.TeamDomain, app.App.TeamID:
		default:
			continue
		}
		if includeInAppSelect(app.App, status) || showOptionsForNewAppCreation(app.App, status) {
			filtered[id] = app
		}
	}
	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return SelectedApp{}, err
	}
	type Selection struct {
		app   SelectedApp
		label string
	}
	options := []Selection{}
	for id, app := range filtered {
		appID := ""
		if app.App.InstallStatus != types.AppStatusInstalled {
			appID = style.Secondary(id)
		} else {
			appID = style.Selector(id)
		}
		label := strings.TrimSpace(fmt.Sprintf(
			"%s %s %s",
			appID,
			style.Secondary(app.App.TeamDomain),
			style.Faint(app.App.TeamID),
		))
		options = append(options, Selection{app, label})
	}
	slices.SortFunc(options, func(a, b Selection) int {
		if a.app.App.TeamDomain < b.app.App.TeamDomain {
			return -1
		} else if a.app.App.TeamDomain > b.app.App.TeamDomain {
			return 1
		}
		if a.app.App.TeamID < b.app.App.TeamID {
			return -1
		} else if a.app.App.TeamID > b.app.App.TeamID {
			return 1
		}
		if a.app.App.AppID < b.app.App.AppID {
			return -1
		} else if a.app.App.AppID > b.app.App.AppID {
			return 1
		}
		return 0
	})
	switch status {
	case ShowInstalledAppsOnly, ShowInstalledAndUninstalledApps:
		if len(filtered) <= 0 {
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}
	case ShowAllApps, ShowInstalledAndNewApps:
		if manifestSource.Equals(config.ManifestSourceLocal) {
			option := Selection{
				label: style.Secondary("Create a new app"),
			}
			options = append(options, option)
			switch {
			case types.IsAppID(clients.Config.AppFlag):
				// Skip to match the app ID later
			case teamFlag != "":
				// Check for an existing app ID
				selections := []SelectedApp{}
				for _, app := range filtered {
					switch teamFlag {
					case app.App.TeamID, app.App.TeamDomain:
						switch {
						case environment.Equals(ShowAllEnvironments):
							selections = append(selections, app)
						case environment.Equals(ShowHostedOnly) && !app.App.IsDev:
							selections = append(selections, app)
						case environment.Equals(ShowLocalOnly) && app.App.IsDev:
							selections = append(selections, app)
						}
					}
				}
				switch len(selections) {
				case 0:
					// Skip to create a new app
				case 1:
					return selections[0], nil
				default:
					return SelectedApp{}, slackerror.New(slackerror.ErrAppFound).
						WithMessage("Multiple apps exist for the provided team")
				}
				// Create a new app if none exists
				auths, err := getAuths(ctx, clients)
				if err != nil {
					return SelectedApp{}, err
				}
				for _, auth := range auths {
					switch teamFlag {
					case auth.TeamID, auth.TeamDomain:
						app := SelectedApp{
							App:  types.NewApp(),
							Auth: auth,
						}
						return app, nil
					}
				}
				return SelectedApp{}, slackerror.New(slackerror.ErrTeamNotFound)
			}
		}
	}
	labels := []string{}
	for _, label := range options {
		labels = append(labels, label.label)
	}
	switch {
	case types.IsAppID(clients.Config.AppFlag):
		for _, app := range filtered {
			switch clients.Config.AppFlag {
			case app.App.AppID:
				// Confirm the team matches if provided via flag
				switch teamFlag {
				case "", app.App.TeamID, app.App.TeamDomain:
					return app, nil
				}
			}
		}
		return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound)
	case types.IsAppFlagEnvironment(clients.Config.AppFlag) && teamFlag != "":
		for _, app := range filtered {
			switch teamFlag {
			case app.App.TeamID, app.App.TeamDomain:
				return app, nil
			}
		}
		return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound)
	}
	selection, err := clients.IO.SelectPrompt(
		ctx,
		"Select an app",
		labels,
		iostreams.SelectPromptConfig{
			Required: true,

			// Flag is checked before since the value might be an app "environment" while
			// an app ID is required in the return.
			//
			// Flag:  clients.Config.Flags.Lookup("app"),
		})
	if err != nil {
		return SelectedApp{}, err
	}
	creation := style.Secondary("Create a new app")
	switch {
	case selection.Prompt && options[selection.Index].label != creation:
		return options[selection.Index].app, nil
	case selection.Prompt && options[selection.Index].label == creation:
		team, err := flatTeamSelectPrompt(ctx, clients)
		if err != nil {
			return SelectedApp{}, err
		}

		// Guard against overwriting saved apps
		for _, app := range filtered {
			switch team.TeamID {
			case app.App.TeamID:
				return SelectedApp{}, slackerror.New(slackerror.ErrAppExists).
					WithDetails(slackerror.ErrorDetails{{
						Message: fmt.Sprintf(
							`The app "%s" already exists for team "%s" (%s)`,
							app.App.AppID,
							app.App.TeamDomain,
							app.App.TeamID,
						),
					}}).
					WithRemediation("To learn more run: %s", style.Commandf("app list", false))
			}
		}
		return SelectedApp{App: types.NewApp(), Auth: team}, nil
	}
	return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound)
}

// AppSelectPrompt prompts the user to select a workspace then environment for the current command,
// returning the selected app. This app might require installation before use if `status == ShowAllApps`.
func AppSelectPrompt(ctx context.Context, clients *shared.ClientFactory, status AppInstallStatus) (SelectedApp, error) {
	var selectedApp SelectedApp
	var selectedTeam TeamApps
	var tokenAuth types.SlackAuth

	var appFlag = clients.Config.AppFlag     // e.g. 'local', 'deploy', 'deployed', A12345
	var tokenFlag = clients.Config.TokenFlag // e.g. xoxe.xoxp.xxxx
	var teamFlag = clients.Config.TeamFlag   // e.g. T12345678, 'acme-org', 'acme-workspace'

	if clients.Config.SkipLocalFs() {
		clients.IO.PrintDebug(ctx, "selecting app based on token value and app id value '%s'", appFlag)
		selection, err := getTokenApp(ctx, clients, tokenFlag, appFlag)
		if err != nil {
			return SelectedApp{}, err
		}
		if status == ShowInstalledAppsOnly && selection.App.InstallStatus != types.AppStatusInstalled {
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}
		clients.Auth().SetSelectedAuth(ctx, selection.Auth, clients.Config, clients.Os)
		return selection, nil
	}

	if clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		return flatAppSelectPrompt(ctx, clients, ShowAllEnvironments, status)
	}

	// Get all apps
	teamApps, err := getTeamApps(ctx, clients)
	if err != nil {
		return SelectedApp{}, err
	}

	if tokenFlag != "" {
		tokenAuth, err = filterAuthsByToken(ctx, clients, teamApps)
		if err != nil {
			return SelectedApp{}, err
		}

		selectedTeam = teamApps[tokenAuth.TeamID]
		if types.IsAppID(appFlag) &&
			selectedTeam.Hosted.App.AppID != appFlag &&
			selectedTeam.Local.App.AppID != appFlag {
			return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound)
		}
		if status == ShowInstalledAppsOnly &&
			!appExists(selectedTeam.Hosted.App) &&
			!appExists(selectedTeam.Local.App) {
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}
		if !types.IsAppID(appFlag) {
			clients.IO.PrintInfo(ctx, false, style.Secondary("Selecting team '%s' with token belonging to '%s'..."), selectedTeam.authOrAppTeamDomain(), selectedTeam.Auth.TeamDomain)
		}
	} else if teamFlag != "" {
		selectedTeam, err = filterByTeamFlag(teamApps, teamFlag)
		if err != nil {
			return SelectedApp{}, err
		}

		if types.IsTeamID(teamFlag) {
			clients.IO.PrintInfo(ctx, false, style.Secondary("Selecting team '%s' with token belonging to '%s'..."), selectedTeam.authOrAppTeamDomain(), selectedTeam.Auth.TeamDomain)
		}
	} else if !types.IsAppID(appFlag) {
		selectedTeam, err = selectAppWorkspace(ctx, clients, teamApps, status)
		if err != nil {
			return SelectedApp{}, err
		}
	}

	if appFlag != "" {
		if types.IsAppID(appFlag) {
			// If an App ID (e.g. A12345) is supplied we filter the workspaceApps using the App ID
			selectedApp, err = filterByAppID(teamApps, appFlag)
			if err != nil {
				return SelectedApp{}, err
			}

			// User team selection via --token and or --team flags doesn't match the user app selection
			// via --app flag
			if selectedTeam.Auth.TeamID != "" && (selectedApp.App.TeamID != selectedTeam.Auth.TeamID &&
				selectedApp.App.EnterpriseID != selectedTeam.Auth.TeamID) {
				return SelectedApp{}, slackerror.New(slackerror.ErrAppAuthTeamMismatch)
			}
		} else if types.IsAppFlagDeploy(appFlag) {
			selectedApp = selectedTeam.Hosted
		} else if types.IsAppFlagLocal(appFlag) {
			selectedApp = selectedTeam.Local
		} else {
			return SelectedApp{}, slackerror.New(slackerror.ErrInvalidAppFlag)
		}

		if status == ShowInstalledAppsOnly && !appExists(selectedApp.App) {
			if types.IsAppFlagEnvironment(appFlag) {
				return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
			}
			if !appExists(selectedApp.App) || selectedApp.App.IsNew() {
				return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound).
					WithMessage("App \"%s\" not found", appFlag)
			}
		}
	} else {
		selectedApp, err = selectAppEnvironment(ctx, clients, selectedTeam, status)
		if err != nil {
			return SelectedApp{}, err
		}
	}

	// Use provided authentication information if available
	if clients.Config.TokenFlag != "" {
		selectedApp.Auth = tokenAuth
	}

	// Validate auth for the selected app
	err = validateAuth(ctx, clients, &selectedApp.Auth)
	if err != nil {
		return selectedApp, err
	}

	clients.Auth().SetSelectedAuth(ctx, selectedApp.Auth, clients.Config, clients.Os)
	return selectedApp, nil
}

// flatTeamSelectPrompt shows choices for authenticated teams
func flatTeamSelectPrompt(
	ctx context.Context,
	clients *shared.ClientFactory,
) (
	authentication types.SlackAuth,
	err error,
) {
	defer func() {
		if err != nil {
			return
		}
		err = validateAuth(ctx, clients, &authentication)
		if err != nil {
			return
		}
		clients.Auth().SetSelectedAuth(ctx, authentication, clients.Config, clients.Os)
	}()
	allAuths, err := getAuths(ctx, clients)
	if err != nil {
		return types.SlackAuth{}, err
	}
	type Selection struct {
		auth  types.SlackAuth
		label string
	}
	options := []Selection{}
	for _, auth := range allAuths {
		label := fmt.Sprintf(
			"%s %s",
			auth.TeamDomain,
			style.Secondary(auth.TeamID),
		)
		options = append(options, Selection{auth, label})
	}
	slices.SortFunc(options, func(a Selection, b Selection) int {
		if a.auth.TeamDomain < b.auth.TeamDomain {
			return -1
		} else if a.auth.TeamDomain > b.auth.TeamDomain {
			return 1
		}
		if a.auth.TeamID < b.auth.TeamID {
			return -1
		} else if a.auth.TeamID > b.auth.TeamID {
			return 1
		}
		return 0
	})
	labels := []string{}
	for _, option := range options {
		labels = append(labels, option.label)
	}
	selection, err := clients.IO.SelectPrompt(
		ctx,
		"Choose a team",
		labels,
		iostreams.SelectPromptConfig{
			Required: true,
			Flag:     clients.Config.Flags.Lookup("team"),
		})
	if err != nil {
		return types.SlackAuth{}, err
	}
	switch {
	case selection.Flag:
		for _, team := range options {
			if selection.Option == team.auth.TeamID {
				return team.auth, nil
			}
		}
		for _, team := range options {
			if selection.Option == team.auth.TeamDomain {
				return team.auth, nil
			}
		}
		return types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound).
			WithMessage("No credentials found for team \"%s\"", selection.Option)
	case selection.Prompt:
		return options[selection.Index].auth, nil
	}
	return types.SlackAuth{}, slackerror.New(slackerror.ErrTeamNotFound)
}

// TeamAppSelectPrompt prompts the user to select an app from a specified team environment,
// returning the selected app. This app might require installation before use if `status == ShowAllApps`.
func TeamAppSelectPrompt(ctx context.Context, clients *shared.ClientFactory, env AppEnvironmentType, status AppInstallStatus) (SelectedApp, error) {
	var teamFlag = clients.Config.TeamFlag
	var appFlag = clients.Config.AppFlag
	var tokenFlag = clients.Config.TokenFlag

	// Error if an invalid or mismatched --app flag is provided
	if appFlag != "" && !types.IsAppFlagValid(appFlag) {
		return SelectedApp{}, slackerror.New(slackerror.ErrInvalidAppFlag).
			WithRemediation("Choose a specific app with %s", style.Highlight("--app <app_id>"))
	}
	if env == ShowHostedOnly && types.IsAppFlagLocal(appFlag) {
		return SelectedApp{}, slackerror.New(slackerror.ErrLocalAppNotSupported)
	}
	if env == ShowLocalOnly && types.IsAppFlagDeploy(appFlag) {
		return SelectedApp{}, slackerror.New(slackerror.ErrDeployedAppNotSupported)
	}

	if clients.Config.SkipLocalFs() {
		clients.IO.PrintDebug(ctx, "selecting app based on token value and app id value '%s'", appFlag)
		selection, err := getTokenApp(ctx, clients, tokenFlag, appFlag)
		if err != nil {
			return SelectedApp{}, err
		}
		if status == ShowInstalledAppsOnly && selection.App.InstallStatus != types.AppStatusInstalled {
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}
		clients.Auth().SetSelectedAuth(ctx, selection.Auth, clients.Config, clients.Os)
		// The development status of an app cannot be determined when local app files
		// do not exist. This defaults to "false" for these cases.
		//
		// Commands such as "platform run" might allow unknown development statuses so
		// we return both the selection and an error here.
		if selection.App.IsDev && env == ShowHostedOnly {
			return selection, slackerror.New(slackerror.ErrLocalAppNotSupported)
		} else if !selection.App.IsDev && env == ShowLocalOnly {
			return selection, slackerror.New(slackerror.ErrDeployedAppNotSupported)
		}
		return selection, nil
	}

	if clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		return flatAppSelectPrompt(ctx, clients, env, status)
	}

	// Get all apps
	teamApps, err := getTeamApps(ctx, clients)
	if err != nil {
		return SelectedApp{}, err
	}

	// Shortcut selections if a --token flag is provided
	if tokenFlag != "" {
		var selection SelectedApp
		var err error
		selection.Auth, err = filterAuthsByToken(ctx, clients, teamApps)
		if err != nil {
			return SelectedApp{}, err
		}
		clients.IO.PrintDebug(ctx, "selecting team '%s' based on the token value", selection.Auth.TeamDomain)

		teamApps := teamApps[selection.Auth.TeamID]
		switch env {
		case ShowHostedOnly:
			selection.App = teamApps.Hosted.App
		case ShowLocalOnly:
			selection.App = teamApps.Local.App
		}
		if status == ShowInstalledAppsOnly &&
			selection.App.InstallStatus != types.AppStatusInstalled {
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}
		clients.Auth().SetSelectedAuth(ctx, selection.Auth, clients.Config, clients.Os)
		return selection, nil
	}

	// When the --team flag is given, lookup the auth associated with the workspace
	if teamFlag != "" {
		// Find installed apps with a matching workspace flag
		filteredByTeamFlag, err := filterByTeamFlag(teamApps, teamFlag)
		if err != nil {
			return SelectedApp{}, err
		}

		// Choose the app environment specific to this prompt
		var selection SelectedApp
		var appEnvironment string

		switch env {
		case ShowHostedOnly:
			selection = filteredByTeamFlag.Hosted
			appEnvironment = "Deployed"
			if appFlag != "" && appFlag == filteredByTeamFlag.Local.App.AppID {
				return SelectedApp{}, slackerror.New(slackerror.ErrLocalAppNotSupported)
			}
		case ShowLocalOnly:
			selection = filteredByTeamFlag.Local
			appEnvironment = "Local"
			if appFlag != "" && appFlag == filteredByTeamFlag.Hosted.App.AppID {
				return SelectedApp{}, slackerror.New(slackerror.ErrDeployedAppNotSupported)
			}
		}

		// Error if the app flag is an app ID and does not match an installed app
		if types.IsAppID(appFlag) && appFlag != selection.App.AppID {
			return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound).
				WithMessage("%s app \"%s\" not found in team \"%s\"",
					appEnvironment, appFlag, clients.Config.TeamFlag)
		}

		// If we only want to show installed apps but there are none, error out
		if status == ShowInstalledAppsOnly && !appExists(selection.App) {
			if env == ShowLocalOnly {
				return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired).
					WithRemediation("Install the app to a workspace with %s", style.Commandf("run", false))
			}
			return SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired)
		}

		if selection.Auth.Token == "" {
			return SelectedApp{}, slackerror.New(slackerror.ErrCredentialsNotFound).
				WithMessage("No credentials found for team \"%s\"", filteredByTeamFlag.authOrAppTeamDomain())
		}

		clients.Auth().SetSelectedAuth(ctx, selection.Auth, clients.Config, clients.Os)
		return selection, err
	}

	// When the --app flag is given, lookup the auth associated with the app
	if appFlag != "" {

		// Require that the flag is an app ID
		if !types.IsAppID(appFlag) {
			return SelectedApp{}, slackerror.New(slackerror.ErrTeamFlagRequired).
				WithRemediation("Choose a workspace with %s or %s\nOr specify the app with %s",
					style.Highlight("--team <team_domain>"), style.Highlight("--team <team_id>"), style.Highlight("--app <app_id>"))
		}

		// Find the app with a matching app ID
		selection, err := filterByAppID(teamApps, appFlag)
		if err != nil {
			return SelectedApp{}, err
		}

		var appEnvironment string
		switch env {
		case ShowHostedOnly:
			if appExists(selection.App) && selection.App.IsDev {
				return SelectedApp{}, slackerror.New(slackerror.ErrLocalAppNotSupported)
			}
			appEnvironment = "Deployed"
		case ShowLocalOnly:
			if appExists(selection.App) && !selection.App.IsDev {
				return SelectedApp{}, slackerror.New(slackerror.ErrDeployedAppNotSupported)
			}
			appEnvironment = "Local"
		}

		// Error if the app flag does not match the workspace app OR installation is required but the app is not installed
		if appFlag != selection.App.AppID || (status == ShowInstalledAppsOnly && (!appExists(selection.App) || selection.App.IsNew())) {
			return SelectedApp{}, slackerror.New(slackerror.ErrAppNotFound).
				WithMessage("%s app \"%s\" not found", appEnvironment, appFlag)
		}

		// Set the auth context and return the app
		clients.Auth().SetSelectedAuth(ctx, selection.Auth, clients.Config, clients.Os)
		return selection, nil
	}

	selectedDomain, selection, err := selectTeamApp(ctx, clients, teamApps, workspaceOptions{
		environment:   env,
		installStatus: status,
	})
	if err != nil {
		return SelectedApp{}, err
	}

	// Validate auth for the selected app
	err = validateAuth(ctx, clients, &selection.Auth)
	if err != nil {
		return selection, err
	}

	if auth, err := clients.Auth().AuthWithTeamID(ctx, selection.Auth.TeamID); err == nil {
		clients.Auth().SetSelectedAuth(ctx, auth, clients.Config, clients.Os)
		return selection, err
	}

	return SelectedApp{}, slackerror.New(slackerror.ErrCredentialsNotFound).
		WithMessage("No credentials found for team \"%s\"", selectedDomain)
}

// OrgSelectWorkspacePrompt prompts the user to select a single workspace to grant app access to, or grant all workspaces within the org.
func OrgSelectWorkspacePrompt(ctx context.Context, clients *shared.ClientFactory, orgDomain, token string, topOptionAllWorkspaces bool) (string, error) {
	teams, paginationCursor, err := clients.API().AuthTeamsList(ctx, token, api.DefaultAuthTeamsListPageSize)
	if err != nil {
		return "", err
	}

	teamDomains := []string{}
	for _, t := range teams {
		teamDomains = append(teamDomains, fmt.Sprintf("%s %s", t.Name, style.Secondary(t.ID)))
	}
	allWorkspacesOption := "All of them"
	allWorkspacesOptionIndex := 0
	if topOptionAllWorkspaces {
		teamDomains = append([]string{allWorkspacesOption}, teamDomains...)
		teams = append([]types.TeamInfo{{Name: "Placeholder for 'all workspaces' option"}}, teams...)
	} else {
		teamDomains = append(teamDomains, allWorkspacesOption)
		allWorkspacesOptionIndex = len(teamDomains) - 1
	}

	msg := style.Sectionf(style.TextSection{
		Emoji:     "bulb",
		Text:      fmt.Sprintf("Your app will be installed to the \"%s\" organization", orgDomain),
		Secondary: []string{"If you'd like, you can restrict access to only users in a particular workspace"},
	})

	if paginationCursor != "" {
		msg = fmt.Sprintf("%s   %s\n", msg, style.Secondary("Workspace not listed? Use the `--org-workspace-grant=<team_id>` flag"))
	}

	clients.IO.PrintInfo(ctx, false, msg)
	selection, err := clients.IO.SelectPrompt(ctx, "Choose a workspace to grant access:", teamDomains, iostreams.SelectPromptConfig{
		PageSize: 4,
		Required: true,
	})
	if err != nil {
		return "", err
	}

	var workspace string
	if selection.Option == allWorkspacesOption && selection.Index == allWorkspacesOptionIndex {
		workspace = types.GrantAllOrgWorkspaces
	} else {
		workspace = teams[selection.Index].ID
	}

	return workspace, nil
}

// ValidateGetOrgWorkspaceGrant checks that the org-workspace-grant flag is being used appropriately.
// If org-workspace-grant should not be used, it will be reset to an empty string.
func ValidateGetOrgWorkspaceGrant(ctx context.Context, clients *shared.ClientFactory, selection *SelectedApp, orgGrantWorkspaceID string, promptOption bool) (string, error) {
	newAppOrgAuth := selection.App.IsNew() && selection.Auth.IsEnterpriseInstall
	uninstalledOrgApp := selection.App.IsUninstalled() && types.IsEnterpriseTeamID(selection.App.TeamID)

	// Not an org app; should not be setting the org workspace flag
	if !(newAppOrgAuth || types.IsEnterpriseTeamID(selection.App.TeamID)) && orgGrantWorkspaceID != "" {
		orgGrantWorkspaceID = ""
		clients.IO.PrintDebug(ctx, fmt.Sprintf("--%s flag ignored for app that wasn't created on an org", cmdutil.OrgGrantWorkspaceFlag))
	}

	// Prevent user from adding grants for multiple org workspaces
	orgWorkspaceGrantsMatch := len(selection.App.EnterpriseGrants) == 1 && selection.App.EnterpriseGrants[0].WorkspaceID == orgGrantWorkspaceID
	if orgGrantWorkspaceID != "" &&
		orgGrantWorkspaceID != types.GrantAllOrgWorkspaces &&
		selection.App.IsInstalled() &&
		!orgWorkspaceGrantsMatch {
		return "", slackerror.New(slackerror.ErrOrgGrantExists).
			WithMessage(
				"A different org workspace grant already exists for installed app '%s'\n   Workspace Grant: %s",
				selection.App.AppID,
				selection.App.EnterpriseGrants[0].WorkspaceID,
			)
	}

	// If an org app is selected, let the user grant to a specific workspace
	if orgGrantWorkspaceID == "" && (newAppOrgAuth || uninstalledOrgApp) {
		// Use app domain if set
		domain := selection.Auth.TeamDomain
		if selection.App.TeamDomain != "" {
			domain = selection.App.TeamDomain
		}
		// Prompt user
		var err error
		orgGrantWorkspaceID, err = OrgSelectWorkspacePrompt(ctx, clients, domain, selection.Auth.Token, promptOption)
		if err != nil {
			return "", err
		}
	}
	return orgGrantWorkspaceID, nil
}

// SortAlphaNumeric safely orders prompt domains, labels and team ids
func SortAlphaNumeric(teamDomains []string, labels []string, teamIDs []string) error {
	type LabelDomainID struct {
		Label      string
		TeamDomain string
		TeamID     string
	}

	if len(teamDomains) != len(labels) || len(teamIDs) != len(labels) {
		return slackerror.New(slackerror.ErrTeamList)
	}

	// Zip items and labels together
	var itemLabels []LabelDomainID
	for i := range teamDomains {
		itemLabels = append(itemLabels, LabelDomainID{
			TeamDomain: teamDomains[i],
			Label:      labels[i],
			TeamID:     teamIDs[i],
		})
	}

	// Perform the sort by alphanumeric ordering of domain value
	// since this is what is user-facing
	sort.Slice(itemLabels, func(i, j int) bool {
		return itemLabels[i].TeamDomain < itemLabels[j].TeamDomain
	})

	// Replace original team domains, labels, team ids
	for i := range itemLabels {
		teamDomains[i] = itemLabels[i].TeamDomain
		labels[i] = itemLabels[i].Label
		teamIDs[i] = itemLabels[i].TeamID
	}

	return nil
}

// appExists checks if the app exists based on the presence of an app ID
func appExists(app types.App) bool {
	return app.AppID != ""
}

// validateAuth checks if the auth for the selected app is valid and if not,
// prompts the user to re-authenticate
func validateAuth(ctx context.Context, clients *shared.ClientFactory, auth *types.SlackAuth) error {
	apiClient := clients.API()
	if auth == nil {
		auth = &types.SlackAuth{}
	}
	if auth.APIHost != nil {
		apiClient.SetHost(*auth.APIHost)
	}
	_, err := apiClient.ValidateSession(ctx, auth.Token)
	if err == nil {
		return nil
	}
	_, unfilteredError := clients.Auth().FilterKnownAuthErrors(ctx, err)
	if unfilteredError != nil || !clients.IO.IsTTY() {
		return err
	}
	clients.IO.PrintInfo(ctx, false, fmt.Sprintf("\n%sWhoops! Looks like your authentication may be expired or invalid", style.Emoji("lock")))
	reauth, _, err := authpkg.Login(ctx, apiClient, clients.Auth(), clients.IO, "", false)
	if err != nil {
		return err
	}
	*auth = reauth
	return nil
}
