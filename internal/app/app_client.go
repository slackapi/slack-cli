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

package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

// Constants
const deployedAppsFilename = ".slack/apps.json"
const devAppsFilename = ".slack/apps.dev.json"
const defaultProdAppTeamDomain = "prod"

type AppClientInterface interface {
	NewDeployed(ctx context.Context, teamID string) (types.App, error)
	GetDeployed(ctx context.Context, teamID string) (types.App, error)
	GetDeployedAll(ctx context.Context) ([]types.App, string, error)
	GetLocal(ctx context.Context, teamID string) (types.App, error)
	GetLocalAll(ctx context.Context) ([]types.App, error)
	RemoveDeployed(ctx context.Context, teamID string) (types.App, error)
	RemoveLocal(ctx context.Context, teamID string) (types.App, error)
	SaveDeployed(ctx context.Context, app types.App) error
	SaveLocal(ctx context.Context, app types.App) error

	Remove(ctx context.Context, app types.App) (types.App, error)
	CleanUp()
}

// AppClient reflects the states of the project's apps files and keeps track of all deployed and dev apps
type AppClient struct {
	// Internal dependencies
	config *config.Config
	apps   types.Apps
	// External dependencies
	// fs is the file system module that's shared by all packages and enables testing & mocking of the file system
	fs afero.Fs
	// os is the `os` package that's shared by all packages and enables testing & mocking
	os types.Os
}

// NewAppClient returns a new, empty instance of the AppClient
// TODO(@mbrooks) - Should this constructor read the file (looks like all public funcs always read first)?
func NewAppClient(config *config.Config, fs afero.Fs, os types.Os) *AppClient {
	var client = AppClient{
		config: config,
		apps: types.Apps{
			DeployedApps: map[string]types.App{},
			LocalApps:    map[string]types.App{},
		},
		fs: fs,
		os: os,
	}

	return &client
}

// NewDeployed returns a new App named for the provided teamID
func (ac *AppClient) NewDeployed(ctx context.Context, teamID string) (types.App, error) {
	var err = ac.readDeployedApps()
	if err != nil {
		return types.App{}, err
	}

	app := ac.apps.GetDeployedByTeamID(teamID)
	if !app.IsNew() {
		return types.App{}, slackerror.New(slackerror.ErrAppFound)
	}

	err = ac.SaveDeployed(ctx, app)
	if err != nil {
		return types.App{}, err
	}

	return app, nil
}

// GetDeployed returns a deployed App by its team ID
//
// IMPORTANT: Pass an explicit teamID in all cases!
// Legacy behavior will rely on a default team domain
// used to mark a current app, but it is not a safe action.
func (ac *AppClient) GetDeployed(ctx context.Context, teamID string) (types.App, error) {
	var err = ac.readDeployedApps()
	if err != nil {
		return types.App{}, err
	}

	if teamID != "" {
		return ac.apps.GetDeployedByTeamID(teamID), nil
	}

	// Legacy behavior is to attempt to fetch a default team domain and
	// then to return a deployed app by that team domain. This
	// is no longer a safe action; team domain is not guaranteed
	// to be unique between org team and workspaces in orgs
	return ac.apps.GetDeployedByTeamDomain(ac.getDeployedAppTeamDomain(ctx)), nil
}

// GetDeployedAll returns all deployed apps (does not include dev apps)
func (ac *AppClient) GetDeployedAll(ctx context.Context) ([]types.App, string, error) {
	var err = ac.readDeployedApps()
	if err != nil {
		return []types.App{}, "", err
	}

	deployedAppsList, defaultAppTeamDomain := ac.apps.GetAllDeployedApps()
	return deployedAppsList, defaultAppTeamDomain, nil
}

// SaveDeployed saves the provided app to the deployed apps file
func (ac *AppClient) SaveDeployed(ctx context.Context, app types.App) error {
	var err = ac.readDeployedApps()
	if err != nil {
		return err
	}

	err = ac.apps.Set(app)
	if err != nil {
		return err
	}

	err = ac.saveDeployedApps()
	if err != nil {
		return err
	}
	return nil
}

// RemoveDeployed removes the app with teamID from the apps.json file
func (ac *AppClient) RemoveDeployed(ctx context.Context, teamID string) (types.App, error) {
	var err = ac.readDeployedApps()
	if err != nil {
		return types.App{}, err
	}
	app := ac.apps.GetDeployedByTeamID(teamID)
	ac.apps.RemoveDeployedByTeamID(teamID)

	err = ac.saveDeployedApps()
	if err != nil {
		return types.App{}, err
	}
	return app, nil
}

// GetLocal returns the local app for the provided teamID
func (ac *AppClient) GetLocal(ctx context.Context, teamID string) (types.App, error) {
	var err = ac.readLocalApps()
	if err != nil {
		return types.App{}, err
	}

	return ac.apps.GetLocalByTeamID(teamID), nil
}

// GetLocalAll returns all local apps
func (ac *AppClient) GetLocalAll(ctx context.Context) ([]types.App, error) {
	var err = ac.readLocalApps()
	if err != nil {
		return []types.App{}, err
	}

	localAppsList := ac.apps.GetAllLocalApps()
	return localAppsList, nil
}

// SaveLocal saves the provided app as the local app for the provided teamID
func (ac *AppClient) SaveLocal(ctx context.Context, app types.App) error {
	if err := ac.readLocalApps(); err != nil {
		return err
	}

	if err := ac.apps.SetLocal(app); err != nil {
		return err
	}

	if err := ac.saveLocalApps(); err != nil {
		return err
	}

	return nil
}

// RemoveLocal removes the app with the provided teamID from apps.dev.json
func (ac *AppClient) RemoveLocal(ctx context.Context, teamID string) (types.App, error) {
	err := ac.readLocalApps()
	if err != nil {
		return types.App{}, err
	}
	app := ac.apps.GetLocalByTeamID(teamID)
	ac.apps.RemoveLocalByTeamID(teamID)

	err = ac.saveLocalApps()
	if err != nil {
		return types.App{}, err
	}
	return app, nil
}

// Remove takes the app out of the saved project apps.*.json file
func (ac *AppClient) Remove(ctx context.Context, app types.App) (types.App, error) {
	switch app.IsDev {
	case true:
		return ac.RemoveLocal(ctx, app.TeamID)
	default:
		return ac.RemoveDeployed(ctx, app.TeamID)
	}
}

// CleanUp will first read the contents of apps*.json files and if empty, it would delete these files.
// It will also go ahead to delete the .slack folder if it is also empty.
func (ac *AppClient) CleanUp() {

	// first read the apps*.json files
	if err := ac.readAllApps(); err != nil {
		// if there is an error, let's fail silently
		// since this is just a cleanup exercise
		return
	}

	var wd, _ = ac.os.Getwd()
	var home, _ = ac.os.UserHomeDir()
	if home == wd {
		// if the working directory happens to be the user's home directory, we don't want to delete
		// the .slack/ folder so we should exit immediately.
		// We do this to make sure we do not mistakenly delete the .slack/ folder with the credentials,
		// config file(s) and the slack binaries
		return
	}

	// if there are no tracked apps anymore and no config file, remove the .slack folder.
	// otherwise remove .slack/apps*.json files that contain no apps.
	if ac.apps.IsEmpty() && !config.ProjectConfigJSONFileExists(ac.fs, ac.os, wd) {
		var deployedAppsJSONFilePath = filepath.Join(wd, deployedAppsFilename)
		var dotSlackFolder = filepath.Dir(deployedAppsJSONFilePath)
		_ = ac.fs.RemoveAll(dotSlackFolder)
	} else {
		if deployedApps, _ := ac.apps.GetAllDeployedApps(); len(deployedApps) == 0 {
			var deployedAppsJSONFilePath = filepath.Join(wd, deployedAppsFilename)
			_ = ac.fs.Remove(deployedAppsJSONFilePath)
		}
		if localApps := ac.apps.GetAllLocalApps(); len(localApps) == 0 {
			var devAppsJSONFilePath = filepath.Join(wd, devAppsFilename)
			_ = ac.fs.Remove(devAppsJSONFilePath)
		}
	}
}

// readDeployedApps loads the latest deployed apps file into .apps.Apps and the default workspace into .apps.DefaultAppTeamDomain
func (ac *AppClient) readDeployedApps() error {
	var directory, _ = ac.os.Getwd()
	var deployedAppsPath = filepath.Join(directory, deployedAppsFilename)

	err := ac.ensureDir(deployedAppsPath)
	if err != nil {
		return err
	}

	// read in .slack/apps.json file from working directory
	f, err := afero.ReadFile(ac.fs, deployedAppsPath)
	if err != nil {
		// If .slack/apps.json does not exist, create the file with an empty list of apps
		if ac.os.IsNotExist(err) {
			ac.apps = types.Apps{
				DeployedApps: map[string]types.App{},
				LocalApps:    map[string]types.App{},
			}

			if err = ac.saveDeployedApps(); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	if err = json.Unmarshal(f, &ac.apps); err != nil {
		return slackerror.New(slackerror.ErrUnableToParseJSON).
			WithMessage("Failed to parse contents of deployed apps file").
			WithRootCause(err).
			WithRemediation("Check that %s is valid JSON", style.HomePath(deployedAppsPath))
	}
	// TODO: on the next major version we can drop this last bit of backwards compatibility code
	// some (hacky) backwards compatibility checking: if only a single apps.json file exists in the project, and it contains a "dev" key,
	// then we can transparently transition the project to two apps*.json files by calling saveLocalApps() first then saveDeployedApps()
	if strings.Contains(string(f), "\"dev\":") {
		if err = ac.saveLocalApps(); err != nil {
			return err
		}
		if err = ac.saveDeployedApps(); err != nil {
			return err
		}
	}

	// Ensure apps.json is written by team_id
	if err = ac.migrateToAppByTeamID(); err != nil {
		return err
	}

	return nil
}

// saveDeployedApps writes the currently deployed apps to the apps.json file
func (ac *AppClient) saveDeployedApps() error {
	var directory, _ = ac.os.Getwd()
	var path = filepath.Join(directory, deployedAppsFilename)

	// temp struct to omit marshalling the Dev property / apps
	type DeployedOnly struct {
		Apps    map[string]types.App `json:"apps,omitempty"`
		Default string               `json:"default,omitempty"`
	}
	deployedApps := DeployedOnly{
		Apps:    ac.apps.DeployedApps,
		Default: ac.apps.DefaultAppTeamDomain,
	}
	data, err := json.MarshalIndent(deployedApps, "", "  ")
	if err != nil {
		return err
	}
	return afero.WriteFile(ac.fs, path, data, 0600)
}

// readLocalApps loads the latest apps dev json file into ac.apps.LocalApps
func (ac *AppClient) readLocalApps() error {
	var directory, _ = ac.os.Getwd()
	var devAppsPath = filepath.Join(directory, devAppsFilename)

	err := ac.ensureDir(devAppsPath)
	if err != nil {
		return err
	}

	// read in .slack/apps.dev.json file from working directory
	f, err := afero.ReadFile(ac.fs, devAppsPath)
	if err != nil {
		// If .slack/apps.dev.json does not exist, create the file with an empty list of apps
		if ac.os.IsNotExist(err) {
			ac.apps.LocalApps = map[string]types.App{}

			if err = ac.saveLocalApps(); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	err = json.Unmarshal(f, &ac.apps.LocalApps)
	if err != nil {
		return slackerror.New(slackerror.ErrUnableToParseJSON).
			WithMessage("Failed to parse contents of local apps file").
			WithRootCause(err).
			WithRemediation("Check that %s is valid JSON", style.HomePath(devAppsPath))
	}

	// The isDev bool is used to set the SLACK_ENV environment variable
	for _, a := range ac.apps.LocalApps {
		a.IsDev = true
		if err = ac.apps.SetLocal(a); err != nil {
			return err
		}

	}

	// Ensure that apps.dev.json is keyed by team_id
	if err = ac.migrateToAppByTeamIDLocal(); err != nil {
		return err
	}

	return nil
}

// saveLocalApps writes the dev apps to the apps.dev.json file
func (ac *AppClient) saveLocalApps() error {
	var directory, _ = ac.os.Getwd()
	var path = filepath.Join(directory, devAppsFilename)

	data, err := json.MarshalIndent(ac.apps.LocalApps, "", "  ")
	if err != nil {
		return err
	}
	return afero.WriteFile(ac.fs, path, data, 0600)
}

// readAllApps loads the latest App info, both deployed and dev.
func (ac *AppClient) readAllApps() error {
	err := ac.readDeployedApps()
	if err != nil {
		return err
	}
	err = ac.readLocalApps()
	if err != nil {
		return err
	}
	return nil
}

// migrateToAppByTeamID migrates historical apps.json
// to ensure that it is mapped by team_id
// See also: migrateToAuthByTeamID
func (ac *AppClient) migrateToAppByTeamID() error {
	// Ensure apps.json is written by team_id
	deployedAppsByTeamID, err := ac.apps.MapByTeamID(ac.apps.DeployedApps)
	if err != nil {
		return err
	}
	ac.apps.DeployedApps = deployedAppsByTeamID
	if err = ac.saveDeployedApps(); err != nil {
		return err
	}
	return nil
}

// migrateToAppByTeamID migrates historical apps.dev.json
// to ensure that it is mapped by team_id
// See also: migrateToAuthByTeamID
func (ac *AppClient) migrateToAppByTeamIDLocal() error {
	// Ensure apps.dev.json is written by team_id
	localAppsByTeamID, err := ac.apps.MapByTeamID(ac.apps.LocalApps)
	if err != nil {
		return err
	}
	ac.apps.LocalApps = localAppsByTeamID
	if err = ac.saveLocalApps(); err != nil {
		return err
	}
	return nil
}

// getDeployedAppTeamDomain resolves the current app's team from the CLI flag or a default team
// TODO: Deprecate this function and underlying DefaultAppTeamDomain
// as team domain is not guaranteed to be unique
func (ac *AppClient) getDeployedAppTeamDomain(ctx context.Context) string {
	// Find the app team using the priority:
	// 1. Team Flag (set by the user)
	if ac.config.TeamFlag != "" {
		return ac.config.TeamFlag
	}

	// 2. Use the apps json file default definition if set
	// Use default when it's defined and matches an existing app
	if ac.apps.DefaultAppTeamDomain != "" {
		app := ac.apps.GetDeployedByTeamDomain(ac.apps.DefaultAppTeamDomain)
		if !app.IsNew() {
			return ac.apps.DefaultAppTeamDomain
		}
	}

	// 3. Default app team 'prod'
	return defaultProdAppTeamDomain
}

// ensureDir ensures the housing directory for the provided path to a file exists
func (ac *AppClient) ensureDir(pathToFile string) error {
	dir := filepath.Dir(pathToFile)
	err := ac.fs.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
