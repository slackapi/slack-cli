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
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setup is a generic test setup method doing repetitive, regular instantiation of objects needed to run tests
// following the template provided in https://medium.com/nerd-for-tech/setup-and-teardown-unit-test-in-go-bd6fa1b785cd
// see above link for more goodies, such as teardown methods and suite vs. test setup/teardown
func setup(t *testing.T) (appClient *AppClient, fsMock *slackdeps.FsMock, osMock *slackdeps.OsMock, pathToAppsJSON string, pathToDevAppsJSON string, cleanup func(*testing.T)) {
	fs := slackdeps.NewFsMock()
	os := slackdeps.NewOsMock()
	os.AddDefaultMocks()
	cfg := config.NewConfig(fs, os)
	ac := NewAppClient(cfg, fs, os)
	dir, _ := os.Getwd()
	pathToAppsJSON = filepath.Join(dir, deployedAppsFilename)
	pathToDevAppsJSON = filepath.Join(dir, devAppsFilename)
	return ac, fs, os, pathToAppsJSON, pathToDevAppsJSON, func(t *testing.T) {
		_ = fs.Remove(pathToAppsJSON)
		_ = fs.Remove(pathToDevAppsJSON)
	}
}

func Test_AppClient_EnsureDir(t *testing.T) {
	ac, fs, os, _, _, teardown := setup(t)
	defer teardown(t)
	// Test to make sure ensureDir created a directory
	dirName, _ := os.Getwd()
	fileName := filepath.Join(dirName, "file")
	err := ac.ensureDir(fileName)
	require.NoError(t, err)
	fi, _ := ac.fs.Stat(dirName)
	require.True(t, fi.IsDir())
	// Test failure mode returns error
	fs.On("MkdirAll", mock.Anything, mock.Anything).Return(errors.New("no write permission"))
	err = ac.ensureDir(fileName)
	require.Error(t, err)
}

func Test_AppClient_Set_ErrorWhenMissingTeamID(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	err := ac.apps.Set(types.App{
		TeamID:     "", // set this deliberately as empty string with no value
		TeamDomain: "reversestring",
		AppID:      "A123",
	})
	require.Error(t, err)
}

// Test that deployed app details get written to apps.json
func Test_AppClient_SaveDeployedApps(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	err := ac.apps.Set(types.App{
		TeamID:     "T1234",
		TeamDomain: "reversestring",
		AppID:      "A123",
	})
	require.NoError(t, err)

	err = ac.saveDeployedApps()
	require.NoError(t, err)
	f, _ := afero.ReadFile(ac.fs, pathToAppsJSON)
	acFromFile := AppClient{}
	_ = json.Unmarshal(f, &acFromFile.apps)
	myApp := acFromFile.apps.GetDeployedByTeamDomain("reversestring")
	require.NotNil(t, myApp)
	assert.Equal(t, "A123", myApp.AppID)
}

// Test that dev app details get written to apps.json
func Test_AppClient_SaveLocalApps(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	err := ac.apps.SetLocal(types.App{
		TeamDomain: "PDE",
		AppID:      "A123",
		TeamID:     "T123",
	})
	require.NoError(t, err)

	err = ac.saveLocalApps()
	require.NoError(t, err)

	f, _ := afero.ReadFile(ac.fs, pathToDevAppsJSON)
	acFromFile := types.Apps{}
	_ = json.Unmarshal(f, &acFromFile.LocalApps)

	myApp := acFromFile.GetLocalByTeamID("T123")
	require.NotNil(t, myApp)
	assert.Equal(t, "A123", myApp.AppID)
}

// Test that pre-existing deployed app details get read from apps.json
func Test_AppClient_ReadDeployedApps_ExistingAppsJSON(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	jsonContents := []byte(`{
		"apps":{
			"T456":{
				"team_domain":"twistandshout",
				"app_id":"A123",
				"team_id":"T456"
			}
		}
	}`)
	_ = afero.WriteFile(ac.fs, pathToAppsJSON, jsonContents, 0600)
	err := ac.readDeployedApps()
	require.NoError(t, err)
	myApp := ac.apps.GetDeployedByTeamDomain("twistandshout")
	require.NotNil(t, myApp)
	assert.Equal(t, "A123", myApp.AppID)
	assert.Equal(t, "T456", myApp.TeamID)
	assert.Equal(t, "twistandshout", myApp.TeamDomain)
}

// Test that a missing apps.json writes an empty apps.json
func Test_AppClient_ReadDeployedApps_NoAppsJSON(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	err := ac.readDeployedApps()
	require.NoError(t, err)
	f, _ := afero.ReadFile(ac.fs, pathToAppsJSON)
	assert.Equal(t, "{}", string(f))
}

func Test_AppClient_ReadDeployedApps_BrokenAppsJSON(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	jsonContents := []byte(`{`)
	err := afero.WriteFile(ac.fs, pathToAppsJSON, jsonContents, 0600)
	require.NoError(t, err)
	err = ac.readDeployedApps()
	require.Error(t, err)
	assert.Equal(t, err.(*slackerror.Error).Code, slackerror.ErrUnableToParseJSON)
}

// Test that pre-existing dev app details get read from apps.dev.json
func Test_AppClient_ReadDevApps_ExistingAppsJSON(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	jsonContents := []byte(`{
		"T456":{
			"name":"dev",
			"app_id":"A123",
			"team_id":"T456",
			"user_id":"U123"
		}
	}`)
	_ = afero.WriteFile(ac.fs, pathToDevAppsJSON, jsonContents, 0600)
	err := ac.readLocalApps()
	require.NoError(t, err)
	myApp := ac.apps.GetLocalByTeamID("T456")
	require.NotNil(t, myApp)
	assert.Equal(t, "A123", myApp.AppID)
	assert.Equal(t, "T456", myApp.TeamID)
}

// Test that a missing apps.dev.json writes an empty apps.dev.json
func Test_AppClient_ReadDevApps_NoAppsJSON(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	err := ac.readLocalApps()
	require.NoError(t, err)
	f, _ := afero.ReadFile(ac.fs, pathToDevAppsJSON)
	assert.Equal(t, "{}", string(f))
}

func Test_AppClient_ReadDevApps_BrokenAppsJSON(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	jsonContents := []byte(`{`)
	err := afero.WriteFile(ac.fs, pathToDevAppsJSON, jsonContents, 0600)
	require.NoError(t, err)
	err = ac.readLocalApps()
	require.Error(t, err)
	assert.Equal(t, err.(*slackerror.Error).Code, slackerror.ErrUnableToParseJSON)
}

// Test that a team flag config defines the default app name in an empty AppClient
func Test_AppClient_getDeployedAppTeamDomain_ViaCLIFlag(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	ac.config.TeamFlag = "flagname"
	name := ac.getDeployedAppTeamDomain(ctx)
	assert.Equal(t, "flagname", name)
}

// Test that the default app name is equal to the only app in an AppClient
func Test_AppClient_getDeployedAppTeamDomain_ViaDefaultDef(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	err := ac.apps.Set(types.App{
		TeamID:     "T123",
		TeamDomain: "shouty-rooster",
		AppID:      "A123",
	})
	require.NoError(t, err)
	name := ac.getDeployedAppTeamDomain(ctx)
	assert.Equal(t, "shouty-rooster", name)
}

// Test that the default app name is the global default app name when AppClient is empty
func Test_AppClient_getDeployedAppTeamDomain_ViaDefaultAppName(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	name := ac.getDeployedAppTeamDomain(ctx)
	assert.Equal(t, "prod", name)
}

// Test that GetDeployed returns a New App Instance when AppClient is empty
func Test_AppClient_GetDeployed_Empty(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())

	// FIXME: This action is unsafely fetching an app. Please explicitly supply a TeamID
	// This test case is specifically trying to test existing behavior when "" supplied
	app, err := ac.GetDeployed(ctx, "")
	require.NoError(t, err)
	assert.True(t, app.IsNew())
}

// Test that GetDeployed returns existing App when AppClient has a single app as default
func Test_AppClient_GetDeployed_DefaultApp(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	jsonContents := []byte(`{
		"apps":{
			"twistandshout":{
				"team_domain":"twistandshout",
				"app_id":"A123",
				"team_id":"T456"
			}
		},
		"default":"twistandshout"
	}`)
	_ = afero.WriteFile(ac.fs, pathToAppsJSON, jsonContents, 0600)

	// FIXME: This action is unsafely fetching an app. Please explicitly supply a TeamID
	// TODO: We eventually want to get rid of support for default app, and retire this test
	app, err := ac.GetDeployed(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "twistandshout", app.TeamDomain)
}

// Test that GetDeployed returns an existing app when given a correct team ID
func Test_AppClient_GetDeployed_TeamID_NoDefault(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	jsonContents := []byte(`{
		"apps":{
			"T1":{
				"team_domain":"twistandshout",
				"app_id":"A1",
				"team_id":"T1"
			}
		},
		"default":""
	}`)
	_ = afero.WriteFile(ac.fs, pathToAppsJSON, jsonContents, 0600)
	app, err := ac.GetDeployed(ctx, "T1")
	require.NoError(t, err)
	assert.Equal(t, "A1", app.AppID)
}

// Test that GetLocal returns a New App when AppClient is empty
func Test_AppClient_GetLocal_Empty(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	app, err := ac.GetLocal(ctx, "T123")
	require.NoError(t, err)
	assert.True(t, app.IsNew())
}

// Test that GetLocal returns existing App Instance when AppClient has a single app
func Test_AppClient_GetLocal_SingleApp(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	jsonContents := []byte(`{
		"U123":{
			"name":"dev",
			"app_id":"A123",
			"team_id":"T456",
			"user_id":"U123"
		}
	}`)
	_ = afero.WriteFile(ac.fs, pathToDevAppsJSON, jsonContents, 0600)
	app, err := ac.GetLocal(ctx, "T456")
	require.NoError(t, err)
	assert.Equal(t, "A123", app.AppID)
	assert.Equal(t, app.IsNew(), false)
}

// Test that GetDeployedAll returns an empty Apps list when AppClient is empty
func Test_AppClient_GetDeployedAll_Empty(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	apps, defaultName, err := ac.GetDeployedAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, []types.App{}, apps)
	assert.Equal(t, "", defaultName)
}

// Test that GetDeployedAll returns the correct Apps when AppClient is not empty
func Test_AppClient_GetDeployedAll_SomeApps(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	jsonContents := []byte(`{
		"apps":{
			"twistandshout":{
				"name":"twistandshout",
				"app_id":"A123",
				"team_id":"T456"
			},
			"otherspace":{
				"name":"otherspace",
				"app_id":"A456",
				"team_id":"T123"
			}
		},
		"default":"twistandshout"
	}`)
	_ = afero.WriteFile(ac.fs, pathToAppsJSON, jsonContents, 0600)
	apps, defaultName, err := ac.GetDeployedAll(ctx)
	appIDs := []string{apps[0].AppID, apps[1].AppID}
	require.NoError(t, err)
	assert.Contains(t, appIDs, "A123")
	assert.Contains(t, appIDs, "A456")
	assert.Equal(t, "twistandshout", defaultName)
}

// Test that GetLocalAll returns an empty App Instance list when AppClient is empty
func Test_AppClient_GetLocalAll_Empty(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	apps, err := ac.GetLocalAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, []types.App{}, apps)
}

// Test that GetLocalAll returns the correct App list when AppClient is not empty
func Test_AppClient_GetLocalAll_SomeApps(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	jsonContents := []byte(`{
		"U123":{
			"name":"dev",
			"app_id":"A123",
			"team_id":"T456",
			"user_id":"U123"
		}
	}`)
	_ = afero.WriteFile(ac.fs, pathToDevAppsJSON, jsonContents, 0600)
	apps, err := ac.GetLocalAll(ctx)
	require.NoError(t, err)
	assert.Equal(t, "A123", apps[0].AppID)
}

// Test that Save adds app details to an empty AppClient
func Test_AppClient_Save_Empty(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	app := types.App{
		TeamID:     "T123",
		TeamDomain: "shouty-rooster",
		AppID:      "A123",
		UserID:     "U123",
	}
	err := ac.SaveDeployed(ctx, app)
	require.NoError(t, err)

	// FIXME: This action is unsafely fetching an app. Please explicitly supply a TeamID
	sameApp, err := ac.GetDeployed(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, app.AppID, sameApp.AppID)
	f, err := afero.ReadFile(ac.fs, pathToAppsJSON)
	require.NoError(t, err)
	apps := types.Apps{}
	err = json.Unmarshal(f, &apps)
	require.NoError(t, err)
	savedApp := apps.GetDeployedByTeamDomain("shouty-rooster")
	assert.Equal(t, app.TeamDomain, savedApp.TeamDomain)
	assert.Equal(t, app.AppID, savedApp.AppID)
	assert.Equal(t, app.UserID, savedApp.UserID)
}

// Test that Save replaces existing app details in a non-empty AppClient
func Test_AppClient_Save_NotEmpty(t *testing.T) {
	ac, _, _, pathToAppsJSON, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	app := types.App{
		TeamID:     "T123",
		TeamDomain: "shouty-rooster",
		AppID:      "A123",
	}
	// Save the app
	err := ac.SaveDeployed(ctx, app)
	require.NoError(t, err)

	sameApp, err := ac.GetDeployed(ctx, app.TeamID)
	require.NoError(t, err)

	// ensure that the returned app matches the saved app
	assert.Equal(t, app.AppID, sameApp.AppID)
	assert.Equal(t, app.TeamID, sameApp.TeamID)

	// read the apps.json
	f, err := afero.ReadFile(ac.fs, pathToAppsJSON)
	require.NoError(t, err)
	apps := types.Apps{}
	err = json.Unmarshal(f, &apps)
	require.NoError(t, err)

	// Ensure that you can get the app using the domain
	savedAppByDomain := apps.GetDeployedByTeamDomain(app.TeamDomain)
	savedAppByTeamID := apps.GetDeployedByTeamID(app.TeamID)

	assert.Equal(t, app.TeamID, savedAppByTeamID.TeamID)
	assert.Equal(t, app.TeamDomain, savedAppByDomain.TeamDomain)
	assert.Equal(t, app.AppID, savedAppByDomain.AppID)
}

// Test that SaveLocal adds app details to an empty AppClient
func Test_AppClient_SaveLocal_Empty(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	appTeamID := "T1"
	app := types.App{
		AppID:  "A123",
		UserID: "U123",
		TeamID: appTeamID,
	}

	// Save new app
	err := ac.SaveLocal(ctx, app)
	require.NoError(t, err)

	// Ensure that you can get the same app
	sameApp, err := ac.GetLocal(ctx, appTeamID)
	require.NoError(t, err)
	assert.Equal(t, app.AppID, sameApp.AppID)

	// Read apps.dev.json
	f, err := afero.ReadFile(ac.fs, pathToDevAppsJSON)
	require.NoError(t, err)
	apps := types.Apps{}
	err = json.Unmarshal(f, &apps.LocalApps)
	require.NoError(t, err)

	// Ensure that you can get the same app after read
	savedApp := apps.GetLocalByTeamID(appTeamID)
	assert.Equal(t, app.AppID, savedApp.AppID)
}

// Test that SaveLocal replaces existing app details in a non-empty AppClient
func Test_AppClient_SaveLocal_NotEmpty(t *testing.T) {
	ac, _, _, _, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	appTeamID := "T1"
	app := types.App{
		AppID:  "A123",
		UserID: "U123",
		TeamID: appTeamID,
	}
	// Save a new app locally
	err := ac.SaveLocal(ctx, app)
	require.NoError(t, err)

	// Get it and ensure it's the same app
	sameApp, err := ac.GetLocal(ctx, appTeamID)
	require.NoError(t, err)
	assert.Equal(t, app.AppID, sameApp.AppID)
	assert.Equal(t, app.TeamID, sameApp.TeamID)

	// Write local app to apps dev json
	f, err := afero.ReadFile(ac.fs, pathToDevAppsJSON)
	require.NoError(t, err)
	apps := types.Apps{}
	err = json.Unmarshal(f, &apps.LocalApps)
	require.NoError(t, err)

	// Ensure that the app can be gotten after read
	savedApp := apps.GetLocalByTeamID(appTeamID)
	assert.Equal(t, app.AppID, savedApp.AppID)
}

// Test that Remove removes an existing deployed app
func Test_AppClient_RemoveDeployed(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	app := types.App{
		TeamID:     "T123",
		TeamDomain: "shouty-rooster",
		AppID:      "A123",
	}
	err := ac.SaveDeployed(ctx, app)
	require.NoError(t, err)

	savedApp, err := ac.GetDeployed(ctx, app.TeamID)
	require.NoError(t, err)
	assert.Equal(t, app.AppID, savedApp.AppID)
	removedApp, _ := ac.RemoveDeployed(ctx, "T123")
	assert.Equal(t, app.AppID, removedApp.AppID)

	removedApp, _ = ac.GetDeployed(ctx, app.TeamID)
	assert.True(t, removedApp.IsNew())
	assert.Empty(t, removedApp.AppID)
}

// Test that RemoveLocal removes an existing local app
func Test_AppClient_RemoveLocal(t *testing.T) {
	ac, _, _, _, _, teardown := setup(t)
	defer teardown(t)
	ctx := slackcontext.MockContext(t.Context())
	app := types.App{
		AppID:  "A123",
		UserID: "U123",
		TeamID: "T1",
	}
	err := ac.SaveLocal(ctx, app)
	require.NoError(t, err)
	_, err = ac.RemoveLocal(ctx, "T1")
	require.NoError(t, err)
	removedApp, _ := ac.GetLocal(ctx, "T1")
	assert.True(t, removedApp.IsNew())
	assert.Empty(t, removedApp.AppID)
}

func TestAppClient_migrateToAppByTeamID_DeployedAndLocal(t *testing.T) {
	ac, _, _, pathToAppsJSON, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	legacyAppsJSONContents := []byte(`{
		"apps":{
			"twistandshout":{
				"team_domain":"twistandshout",
				"app_id":"A123",
				"team_id":"T123"
			}
		},
		"default":"twistandshout"
	}`)
	legacyAppsDevJSONContents := []byte(`{
		"heynow":{
			"team_domain":"heynow",
			"app_id":"A456",
			"team_id":"T456"
		}
	}`)

	// write apps.json
	_ = afero.WriteFile(ac.fs, pathToDevAppsJSON, legacyAppsDevJSONContents, 0600)
	_ = afero.WriteFile(ac.fs, pathToAppsJSON, legacyAppsJSONContents, 0600)

	// Read deployed
	err := ac.readDeployedApps()
	require.NoError(t, err)
	deployedApp, exists := ac.apps.DeployedApps["T123"]
	assert.True(t, exists)
	assert.Equal(t, "A123", deployedApp.AppID)

	// Read local
	err = ac.readLocalApps()
	require.NoError(t, err)

	// At this point, everything should be migrated
	localApp, exists := ac.apps.LocalApps["T456"]
	assert.True(t, exists)
	assert.Equal(t, "A456", localApp.AppID)
}

func TestAppClient_CleanupSlackFolder(t *testing.T) {
	ac, _, _, pathToAppsJSON, pathToDevAppsJSON, teardown := setup(t)
	defer teardown(t)
	wd, err := ac.os.Getwd()
	require.NoError(t, err)

	blankAppsJSONContents := []byte(`{}`)
	err = afero.WriteFile(ac.fs, pathToAppsJSON, blankAppsJSONContents, 0600)
	require.NoError(t, err)
	err = afero.WriteFile(ac.fs, pathToDevAppsJSON, blankAppsJSONContents, 0600)
	require.NoError(t, err)

	_, err = ac.fs.Stat(pathToAppsJSON)
	require.NoError(t, err, "failed to access the apps.json file")
	_, err = ac.fs.Stat(pathToDevAppsJSON)
	require.NoError(t, err, "failed to access the apps.dev.json file")

	err = ac.readAllApps()
	require.NoError(t, err)
	assert.True(t, ac.apps.IsEmpty(), "an unexpected app was found")

	require.NoError(t, err)
	assert.False(t, config.ProjectConfigJSONFileExists(ac.fs, ac.os, wd),
		"an unexpected config was found")

	dotSlackFolder := filepath.Dir(pathToAppsJSON)
	_, err = ac.fs.Stat(dotSlackFolder)
	require.NoError(t, err, "failed to access the .slack directory")

	ac.CleanUp()

	folder, err := ac.fs.Stat(dotSlackFolder)
	require.Nil(t, folder, "the folder was not deleted")
	require.ErrorIs(t, err, os.ErrNotExist,
		"an unexpected error occurred while stating the .slack directory")
}

func TestAppClient_CleanupAppsJSONFiles(t *testing.T) {
	blankAppsJSONExample := []byte(`{}`)
	appsJSONExample := []byte(`{
  "apps": {
    "T123456": {
      "app_id": "A000001",
      "team_domain": "workspace-123",
      "team_id": "T123456"
    }
  },
  "default": "workspace-123"
}`)
	devAppsJSONExample := []byte(`{
  "T123456": {
    "app_id": "A123456",
    "IsDev": true,
    "team_domain": "workspace-123",
    "team_id": "T123456",
    "user_id": "U123456"
  }
}`)

	tests := map[string]struct {
		appsJSON    []byte
		devAppsJSON []byte
	}{
		"only deployed apps exist": {
			appsJSON:    appsJSONExample,
			devAppsJSON: blankAppsJSONExample,
		},
		"only local apps exist": {
			appsJSON:    blankAppsJSONExample,
			devAppsJSON: devAppsJSONExample,
		},
		"both local and deployed apps exist": {
			appsJSON:    appsJSONExample,
			devAppsJSON: devAppsJSONExample,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ac, _, _, pathToAppsJSON, pathToDevAppsJSON, teardown := setup(t)
			defer teardown(t)
			ctx := slackcontext.MockContext(t.Context())

			err := afero.WriteFile(ac.fs, pathToAppsJSON, tc.appsJSON, 0600)
			require.NoError(t, err)
			err = afero.WriteFile(ac.fs, pathToDevAppsJSON, tc.devAppsJSON, 0600)
			require.NoError(t, err)

			_, err = ac.fs.Stat(pathToAppsJSON)
			require.NoError(t, err, "failed to access the apps.json file")
			deployedApps, _, err := ac.GetDeployedAll(ctx)
			require.NoError(t, err)

			_, err = ac.fs.Stat(pathToDevAppsJSON)
			require.NoError(t, err, "failed to access the apps.dev.json file")
			localApps, err := ac.GetLocalAll(ctx)
			require.NoError(t, err)

			ac.CleanUp()

			dotSlackFolder := filepath.Dir(pathToAppsJSON)
			_, err = ac.fs.Stat(dotSlackFolder)
			require.NoError(t, err, "failed to access the .slack directory")

			appsJSON, err := afero.ReadFile(ac.fs, pathToAppsJSON)
			if len(deployedApps) == 0 {
				require.ErrorIs(t, err, os.ErrNotExist, "apps.json was not deleted")
			} else {
				require.NoError(t, err, "failed to access the apps.json file")
				assert.Equal(t, appsJSONExample, appsJSON)
			}

			devAppsJSON, err := afero.ReadFile(ac.fs, pathToDevAppsJSON)
			if len(localApps) == 0 {
				require.ErrorIs(t, err, os.ErrNotExist, "apps.dev.json was not deleted")
			} else {
				require.NoError(t, err, "failed to access the apps.dev.json file")
				assert.Equal(t, devAppsJSONExample, devAppsJSON)
			}
		})
	}
}
