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

package types

import (
	"slices"
	"strings"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Apps tracks all deployed and development instances of the app associated with project
type Apps struct {
	DeployedApps map[string]App `json:"apps,omitempty"`

	// IMPORTANT: Legacy property! Please do not use DefaultAppTeamDomain.
	DefaultAppTeamDomain string         `json:"default,omitempty"`
	LocalApps            map[string]App `json:"dev,omitempty"`
}

// IsEmpty returns whether there are any deployed or local apps
func (a *Apps) IsEmpty() bool {
	return len(a.DeployedApps) == 0 && len(a.LocalApps) == 0 && a.DefaultAppTeamDomain == ""
}

// GetDeployedByTeamDomain returns an existing deployed App by teamDomain, otherwise, returns a new instance of App
//
// IMPORTANT: This is a legacy function that is unsafe.
// Please use GetDeployedByTeamID instead
func (a *Apps) GetDeployedByTeamDomain(teamDomain string) App {
	for _, app := range a.DeployedApps {
		if app.TeamDomain == teamDomain {
			return app
		}
	}

	// (@Sarah) The lines below result in some confusing behavior downstream and we
	// should remove it. This method should return an error and an empty types.App{}
	// If an app doesn't exist within it.
	return App{
		new:        true,
		TeamDomain: teamDomain,
	}
}

// GetDeployedByTeamID returns an existing deployed App by teamID, otherwise, returns a new instance of App
func (a *Apps) GetDeployedByTeamID(teamID string) App {
	for _, app := range a.DeployedApps {
		if app.TeamID == teamID {
			return app
		}
	}

	return App{
		new:    true,
		TeamID: teamID,
	}
}

// GetAllDeployedApps returns all deployed Apps
func (a *Apps) GetAllDeployedApps() ([]App, string) {
	list := []App{}
	for _, app := range a.DeployedApps {
		list = append(list, app)
	}
	return list, a.DefaultAppTeamDomain
}

// Set assigns the provided App to either the dev or deployed App
func (a *Apps) Set(app App) error {
	// Handle when app is dev
	if app.IsDev {
		if err := a.SetLocal(app); err != nil {
			return err
		}
		return nil
	}
	// In order to set an app to apps.json, a team_id for the workspace
	// or org this app belongs to must be provided or we return error
	if app.TeamID == "" {
		return slackerror.New(slackerror.ErrMissingAppTeamID)
	}
	a.DeployedApps[app.TeamID] = app

	// if the provided App is the only deployed App, make it the default
	if len(a.DeployedApps) == 1 {
		a.DefaultAppTeamDomain = app.TeamDomain
	}
	return nil
}

// RemoveDeployedByTeamID removes the deployed App that corresponds to the provided teamID
func (a *Apps) RemoveDeployedByTeamID(teamID string) {
	// IMPORTANT: Legacy backwards compatibility code
	// Remove when no longer fetching any Deployed Apps by TeamDomain
	var deletedTeamDomain string

	for _, app := range a.DeployedApps {
		if app.TeamID == teamID {
			deletedTeamDomain = app.TeamDomain
			delete(a.DeployedApps, app.TeamID)
			break
		}
	}

	// IMPORTANT: Legacy backwards compatibility code
	// Remove when no longer fetching any Deployed Apps by TeamDomain
	// If there are no more apps set the default to nothing
	if len(a.DeployedApps) == 0 {
		a.DefaultAppTeamDomain = ""
		return
	}

	// IMPORTANT: Legacy backwards compatibility code
	// Remove when no longer fetching any Deployed Apps by TeamDomain
	// otherwise if deleted app was the default, reset default first remaining deployed
	if a.DefaultAppTeamDomain == deletedTeamDomain {
		for _, app := range a.DeployedApps {
			a.DefaultAppTeamDomain = app.TeamDomain
			return
		}
	}
}

// GetLocalByTeamID returns a dev App based on the provided teamID; if it doesn't exist, returns a new App
func (a *Apps) GetLocalByTeamID(teamID string) App {
	if app, exists := a.LocalApps[teamID]; exists {
		app.TeamID = teamID
		return app
	}

	// (@Sarah) The lines below result in some confusing behavior downstream and we
	// should remove it. GetLocal should return an error and an empty types.App{}
	// If an app doesn't exist within it.
	return App{
		new:    true,
		TeamID: teamID,
		IsDev:  true,
	}
}

// GetAllLocalApps returns all local Apps in apps.dev.json
func (a *Apps) GetAllLocalApps() []App {
	localApps := []App{}
	for _, app := range a.LocalApps {
		localApps = append(localApps, app)
	}
	return localApps
}

// SetLocal assigns the provided local App to the provided teamID
func (a *Apps) SetLocal(app App) error {
	if app.TeamID == "" {
		return slackerror.New(slackerror.ErrMissingAppTeamID)
	}
	a.LocalApps[app.TeamID] = app
	return nil
}

// RemoveLocalByTeamID removes the local App associated with the provided teamID
func (a *Apps) RemoveLocalByTeamID(teamID string) {
	for _, app := range a.LocalApps {
		if app.TeamID == teamID {
			delete(a.LocalApps, teamID)
			break
		}
	}
}

// MapByTeamID takes a map of apps and returns
// a map guaranteed to be keyed by team_ids
// Historically we have keyed by team_domain, which is not guaranteed to be unique
func (a *Apps) MapByTeamID(apps map[string]App) (appsByTeamID map[string]App, err error) {
	appsByTeamID = map[string]App{}
	for _, app := range apps {
		if app.TeamID == "" {
			return map[string]App{}, slackerror.New(slackerror.ErrMissingAppTeamID)
		}

		// In the past we have not written team_domain property to apps json.
		// If it doesn't exist on the record, use the legacy "name" property
		/// to set the team_domain
		if app.TeamDomain == "" && app.LegacyName != "" {
			app.TeamDomain = app.LegacyName
			app.LegacyName = ""
		}

		// Assign the updated app
		appsByTeamID[app.TeamID] = app

	}
	return appsByTeamID, nil
}

type AppInstallationStatus int

const (
	AppInstallationStatusUnknown AppInstallationStatus = iota
	AppStatusInstalled
	AppStatusUninstalled
)

func (s AppInstallationStatus) String() string {
	switch s {
	case AppStatusInstalled:
		return "Installed"
	case AppStatusUninstalled:
		return "Uninstalled"
	}
	return "Unknown"
}

const GrantAllOrgWorkspaces = "all"

type EnterpriseGrant struct {
	WorkspaceID     string `json:"workspace_id"`
	WorkspaceDomain string `json:"workspace_domain"`
}

// App models app metadata such as team domain, AppID, TeamID and UserID
type App struct {
	AppID            string            `json:"app_id,omitempty"`
	EnterpriseID     string            `json:"enterprise_id,omitempty"`
	EnterpriseGrants []EnterpriseGrant `json:"-"`
	LegacyName       string            `json:"name,omitempty"` // Legacy "name". Do not use this field.
	new              bool
	InstallStatus    AppInstallationStatus `json:"-"` // "-" will always omit when un-marshalled
	IsDev            bool                  `json:",omitempty"`
	TeamDomain       string                `json:"team_domain,omitempty"` // e.g. "arachnoid"
	TeamID           string                `json:"team_id,omitempty"`
	UserID           string                `json:"user_id,omitempty"`
}

// Equals returns true if the fields of the two App objects match
func (app *App) Equals(otherApp App) bool {
	if app.AppID != otherApp.AppID {
		return false
	}
	if app.EnterpriseID != otherApp.EnterpriseID {
		return false
	}
	if app.new != otherApp.new {
		return false
	}
	if app.InstallStatus != otherApp.InstallStatus {
		return false
	}
	if app.IsDev != otherApp.IsDev {
		return false
	}
	if app.TeamDomain != otherApp.TeamDomain {
		return false
	}
	if app.TeamID != otherApp.TeamID {
		return false
	}
	if app.UserID != otherApp.UserID {
		return false
	}
	if len(app.EnterpriseGrants) != len(otherApp.EnterpriseGrants) {
		return false
	}
	for _, g := range app.EnterpriseGrants {
		if !slices.Contains(otherApp.EnterpriseGrants, g) {
			return false
		}
	}
	return true
}

// Equals returns true if the object is empty (all fields are their zero value)
func (app *App) IsEmpty() bool {
	return app.Equals(App{})
}

// IsNew returns whether the app hasn't been written to the file system yet (e.g. apps.json / apps.dev.json)
func (app *App) IsNew() bool {
	return app.new
}

// NewApp returns an app with the internal 'new' property set to true
func NewApp() App {
	return App{new: true}
}

// IsUninstalled returns true if the app's installation status indicates that it is uninstalled
func (app *App) IsUninstalled() bool {
	return app.InstallStatus == AppStatusUninstalled
}

// IsInstalled returns true if the app's installation status indicates that it is installed
func (app *App) IsInstalled() bool {
	return app.InstallStatus == AppStatusInstalled
}

// IsAppFlagValid returns if the flag matches an expected pattern for the app flag
//
// Note: This pattern might be an app ID or the app environment
func IsAppFlagValid(str string) bool {
	return IsAppID(str) || IsAppFlagEnvironment(str)
}

// IsAppID returns true if the flag matches the pattern of an app ID
//
// Note: Validation criteria is an estimate, and not directly related to server-side app_id scheme
// If needed this can be more strict
func IsAppID(str string) bool {
	return strings.HasPrefix(str, "A") && strings.ToUpper(str) == str
}

// IsAppFlagEnvironment returns if the flag denotes the local or deployed app
func IsAppFlagEnvironment(str string) bool {
	return IsAppFlagLocal(str) || IsAppFlagDeploy(str)
}

// IsAppFlagLocal returns if the flag represents the local app environment
func IsAppFlagLocal(str string) bool {
	return str == "local"
}

// IsAppFlagDeploy returns if the flag represents the deployed app environment
func IsAppFlagDeploy(str string) bool {
	return str == "deploy" || str == "deployed"
}

// IsEnterpriseWorkspaceApp returns true if an app was created on a workspace
// which belongs to an org/enterprise
//
// Note: Validation criteria is an estimate, and not directly related to server-side criteria
func (app *App) IsEnterpriseWorkspaceApp() bool {
	return IsAppID(app.AppID) && IsWorkspaceTeamID(app.TeamID) && IsEnterpriseTeamID(app.EnterpriseID) && app.TeamID != app.EnterpriseID && !app.IsEnterpriseApp()
}

// IsEnterpriseApp returns true if an app was created on an enterprise/org
//
// Note: Validation criteria is an estimate, and not directly related to server-side criteria
func (app *App) IsEnterpriseApp() bool {
	return IsAppID(app.AppID) && IsEnterpriseTeamID(app.TeamID) && IsEnterpriseTeamID(app.EnterpriseID) && app.TeamID == app.EnterpriseID && !app.IsEnterpriseWorkspaceApp()
}
