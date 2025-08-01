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

// TeamApps contains the apps (local and hosted), auth and information for a team
type TeamApps struct {
	Auth   types.SlackAuth
	Hosted SelectedApp
	Local  SelectedApp
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

// AppSelectPrompt reveals options for apps that match the install status
func AppSelectPrompt(
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
			return AppSelectPrompt(ctx, clients, ShowHostedOnly, status)
		case types.IsAppFlagLocal(clients.Config.AppFlag):
			return AppSelectPrompt(ctx, clients, ShowLocalOnly, status)
		}
	case environment.Equals(ShowLocalOnly) && types.IsAppFlagDeploy(clients.Config.AppFlag):
		return SelectedApp{}, slackerror.New(slackerror.ErrDeployedAppNotSupported)
	case environment.Equals(ShowHostedOnly) && types.IsAppFlagLocal(clients.Config.AppFlag):
		return SelectedApp{}, slackerror.New(slackerror.ErrLocalAppNotSupported)
	case clients.Config.AppFlag != "" && !types.IsAppFlagValid(clients.Config.AppFlag):
		return SelectedApp{}, slackerror.New(slackerror.ErrInvalidAppFlag)
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
		team, err := teamSelectPrompt(ctx, clients)
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

// teamSelectPrompt shows choices for authenticated teams
func teamSelectPrompt(
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
