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

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Fixture
const app1TeamID = "T123"
const app1TeamDomain = "oranges"

var app1 = App{
	AppID:      "A123",
	TeamID:     app1TeamID,
	TeamDomain: app1TeamDomain,
	UserID:     "U123",
}

const app2TeamID = "T456"
const app2TeamDomain = "apples"

var app2 = App{
	AppID:      "A456",
	TeamID:     app2TeamID,
	TeamDomain: app2TeamDomain,
	UserID:     "U123",
}

func Test_Apps_IsEmpty(t *testing.T) {
	defaultApps := Apps{
		DefaultAppTeamDomain: app1TeamDomain,
		DeployedApps: map[string]App{
			app1TeamDomain: app1,
		},
		LocalApps: map[string]App{
			app2TeamDomain: app2,
		},
	}

	tests := map[string]struct {
		expectedApps    Apps
		expectedIsEmpty bool
	}{
		"should be empty when properties nil": {
			expectedApps:    Apps{},
			expectedIsEmpty: true,
		},
		"should be empty when length 0": {
			expectedApps: Apps{
				DefaultAppTeamDomain: "",
				DeployedApps:         map[string]App{},
				LocalApps:            map[string]App{},
			},
			expectedIsEmpty: true,
		},
		"should not be empty when DefaultAppTeamDomain exists": {
			expectedApps: Apps{
				DefaultAppTeamDomain: defaultApps.DefaultAppTeamDomain,
			},
			expectedIsEmpty: false,
		},
		"should not be empty when DeployedApps exist": {
			expectedApps: Apps{
				DeployedApps: defaultApps.DeployedApps,
			},
			expectedIsEmpty: false,
		},
		"should not be empty when LocalApps exist": {
			expectedApps: Apps{
				LocalApps: defaultApps.LocalApps,
			},
			expectedIsEmpty: false,
		},
		"should not be empty when all properties exist": {
			expectedApps: Apps{
				DefaultAppTeamDomain: defaultApps.DefaultAppTeamDomain,
				DeployedApps:         defaultApps.DeployedApps,
				LocalApps:            defaultApps.LocalApps,
			},
			expectedIsEmpty: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.expectedIsEmpty, tt.expectedApps.IsEmpty())
		})
	}
}
func Test_Apps_MapByTeamID(t *testing.T) {
	// Setup legacy apps format keyed by team_domain
	var legacyApps = map[string]App{
		app1TeamDomain: app1,
		app2TeamDomain: app2,
	}

	var apps = Apps{
		DeployedApps: legacyApps,
	}

	migratedApps, err := apps.MapByTeamID(legacyApps)
	require.NoError(t, err)

	// Expect migrated have the same length
	require.Equal(t, len(migratedApps), len(legacyApps))

	// Expect migrated apps to be keyed by team_id
	if _, exists := migratedApps[app1TeamID]; !exists {
		require.FailNow(t, "missing app entry with team id")
	}
	if _, exists := migratedApps[app2TeamID]; !exists {
		require.FailNow(t, "missing app entry with team id")
	}

	for _, app := range legacyApps {
		migratedApp := migratedApps[app.TeamID]
		require.Equal(t, migratedApp.TeamDomain, app.TeamDomain)
		require.Equal(t, migratedApp.AppID, app.AppID)
	}
}

func Test_Apps_MapByTeamID_LegacyName(t *testing.T) {

	var app3TeamID = "T789"
	var legacyEnvValue = "prod"

	var app3 = App{
		AppID:      "A789",
		TeamID:     app3TeamID,
		LegacyName: legacyEnvValue,
		TeamDomain: "", // deliberately set as empty string
	}
	// Setup legacy apps format keyed by team_domain
	var legacyApps = map[string]App{
		legacyEnvValue: app3,
	}

	var apps = Apps{
		DeployedApps: legacyApps,
	}

	migratedApps, err := apps.MapByTeamID(legacyApps)
	require.NoError(t, err)

	// Expect migrated have the same length
	require.Equal(t, len(migratedApps), len(legacyApps))

	// Expect migrated apps to be keyed by team_id
	if _, exists := migratedApps[app3TeamID]; !exists {
		require.FailNow(t, "missing app entry with team id")
	}

	for _, app := range legacyApps {
		migratedApp := migratedApps[app.TeamID]
		require.Equal(t, migratedApp.TeamDomain, app.LegacyName, "Should set migrated app TeamDomain = LegacyName")
		require.Equal(t, migratedApp.LegacyName, "", "Should set LegacyName to empty string")
	}
}

func Test_Apps_MapByTeamID_TeamID_Error(t *testing.T) {
	var apps = Apps{
		DeployedApps: map[string]App{
			app1TeamDomain: {
				AppID:      app1.AppID,
				TeamID:     "", // Set to "" to cause the error
				LegacyName: app1.LegacyName,
				TeamDomain: app1.TeamDomain,
			},
		},
	}

	_, err := apps.MapByTeamID(apps.DeployedApps)
	require.Error(t, err)
}

func Test_Apps_GetDeployedByTeamDomain_ExistingApp(t *testing.T) {

	// Setup deployed apps format keyed by team_id
	var apps = Apps{
		DeployedApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	// Returns an app matching the team_ID of the selected
	a := apps.GetDeployedByTeamDomain(app1TeamDomain)
	require.False(t, a.IsNew())
	require.Equal(t, a.TeamID, app1.TeamID)
}

func Test_Apps_GetDeployedByTeamDomain_NoExistingApp(t *testing.T) {

	// Setup deployed apps format keyed by team_id
	var apps = Apps{
		DeployedApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	// Returns a "New" app matching team_domain supplied
	a := apps.GetDeployedByTeamDomain("DoesNotExist")
	require.True(t, a.IsNew())
	require.Equal(t, a.TeamDomain, "DoesNotExist")
}

func Test_Apps_GetDeployedByTeamID_ExistingApp(t *testing.T) {

	// Setup deployed apps format keyed by team_id
	var apps = Apps{
		DeployedApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	// Returns an app matching the team_ID of the selected
	a := apps.GetDeployedByTeamID(app1TeamID)
	require.False(t, a.IsNew())
	require.Equal(t, a.TeamDomain, app1.TeamDomain)
}

func Test_Apps_GetDeployedByTeamID_NoExistingApp(t *testing.T) {

	// Setup deployed apps format keyed by team_id
	var apps = Apps{
		DeployedApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	// Returns an app matching the team_ID of the selected
	a := apps.GetDeployedByTeamID("T111111")
	require.True(t, a.IsNew())
	require.Equal(t, a.TeamID, "T111111")
}

func Test_Apps_GetAllDeployedApps(t *testing.T) {
	var expectedApps = Apps{
		DefaultAppTeamDomain: app1TeamID,
		DeployedApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	appList, defaultAppTeamDomain := expectedApps.GetAllDeployedApps()
	require.ElementsMatch(t, []App{app1, app2}, appList) // Use ElementsMatch to ignore returned order
	require.Equal(t, expectedApps.DefaultAppTeamDomain, defaultAppTeamDomain)
}

func Test_Apps_Set(t *testing.T) {
	tests := map[string]struct {
		expectedApp                  App
		isError                      bool
		isDeployedApp                bool
		isLocalApp                   bool
		expectedDefaultAppTeamDomain string
	}{
		"should set local app when IsDev": {
			expectedApp: App{
				AppID:      app1.AppID,
				TeamID:     app1.TeamID,
				TeamDomain: app1.TeamDomain,
				UserID:     app1.UserID,
				IsDev:      true,
			},
			isError:    false,
			isLocalApp: true,
		},
		"should error when IsDev and no TeamID": {
			expectedApp: App{
				AppID:      app1.AppID,
				TeamID:     "",
				TeamDomain: app1.TeamDomain,
				UserID:     app1.UserID,
				IsDev:      true,
			},
			isError: true,
		},
		"should error when not IsDev and no TeamID": {
			expectedApp: App{
				AppID:      app1.AppID,
				TeamID:     "",
				TeamDomain: app1.TeamDomain,
				UserID:     app1.UserID,
				IsDev:      false,
			},
			isError: true,
		},
		"should set deployed app when not IsDev": {
			expectedApp: App{
				AppID:      app1.AppID,
				TeamID:     app1.TeamID,
				TeamDomain: app1.TeamDomain,
				UserID:     app1.UserID,
				IsDev:      false,
			},
			isError:       false,
			isDeployedApp: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			apps := Apps{
				DefaultAppTeamDomain: "",
				DeployedApps:         map[string]App{},
				LocalApps:            map[string]App{},
			}
			err := apps.Set(tt.expectedApp)
			require.Equal(t, tt.isError, err != nil)

			if tt.isLocalApp {
				actualLocalApps := apps.GetAllLocalApps()
				require.Contains(t, actualLocalApps, tt.expectedApp)
			}

			if tt.isDeployedApp {
				actualDeployedApps, _ := apps.GetAllDeployedApps()
				require.Contains(t, actualDeployedApps, tt.expectedApp)
			}

			if tt.expectedDefaultAppTeamDomain != "" {
				require.Equal(t, tt.expectedDefaultAppTeamDomain, apps.DefaultAppTeamDomain)
			}
		})
	}
}

func Test_Apps_Set_DefaultAppTeamDomain(t *testing.T) {
	// Create apps
	apps := Apps{
		DefaultAppTeamDomain: "",
		DeployedApps:         map[string]App{},
		LocalApps:            map[string]App{},
	}

	// Set first deployed app
	deployedApp := App{
		AppID:      app1.AppID,
		TeamID:     app1.TeamID,
		TeamDomain: app1.TeamDomain,
		UserID:     app1.UserID,
		IsDev:      false, // false will set a deployed app
	}
	err := apps.Set(deployedApp)

	// Assert the app was set to DeployedApps and DefaultAppTeamDomain set to the app
	require.NoError(t, err)
	require.Len(t, apps.DeployedApps, 1)
	require.Equal(t, deployedApp, apps.DeployedApps[app1.TeamID])
	require.Equal(t, apps.DefaultAppTeamDomain, app1.TeamDomain)

	// Set second deployed app
	deployedApp = App{
		AppID:      app2.AppID,
		TeamID:     app2.TeamID,
		TeamDomain: app2.TeamDomain,
		UserID:     app2.UserID,
		IsDev:      false, // false will set a deployed app
	}
	err = apps.Set(deployedApp)

	// Assert the app was added to DeployedApps but DefaultAppTeamDomain unchanged
	require.NoError(t, err)
	require.Len(t, apps.DeployedApps, 2)
	require.Equal(t, deployedApp, apps.DeployedApps[app2.TeamID])
	require.Equal(t, apps.DefaultAppTeamDomain, app1.TeamDomain) // Unchanged
}

func Test_Apps_GetLocalByTeamID(t *testing.T) {
	var localApps = Apps{
		LocalApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	// Returns app matching team_id
	la := localApps.GetLocalByTeamID(app1TeamID)
	require.False(t, la.IsNew())
	require.Equal(t, la.TeamID, app1TeamID)
}

func Test_Apps_GetLocalByTeamID_NoExistingApp(t *testing.T) {
	var localApps = Apps{
		LocalApps: map[string]App{},
	}
	// Returns new app matching team_id but
	la := localApps.GetLocalByTeamID(app1TeamID)
	require.True(t, la.IsNew())
}

func Test_Apps_SetLocal(t *testing.T) {
	var localApps = Apps{
		LocalApps: map[string]App{},
	}
	err := localApps.SetLocal(app1)
	require.NoError(t, err)
	app := localApps.GetLocalByTeamID(app1TeamID)
	require.Equal(t, app1.AppID, app.AppID)
}

func Test_Apps_RemoveLocalByTeamID(t *testing.T) {
	var localApps = Apps{
		LocalApps: map[string]App{
			app1TeamID: app1,
			app2TeamID: app2,
		},
	}
	// Remove app matching team_id and
	// expect to not get it
	localApps.RemoveLocalByTeamID(app1TeamID)
	app := localApps.GetLocalByTeamID(app1TeamID)
	require.True(t, app.IsNew())
}

func Test_Apps_RemoveDeployedByTeamID(t *testing.T) {
	tests := map[string]struct {
		apps                         Apps
		removedTeamID                string
		expectedDefaultAppTeamDomain string
	}{
		"should remove app2 and not update DefaultTeamDomain": {
			apps: Apps{
				DefaultAppTeamDomain: app1.TeamDomain,
				DeployedApps: map[string]App{
					app1TeamID: app1,
					app2TeamID: app2,
				},
			},
			removedTeamID:                app2.TeamID,
			expectedDefaultAppTeamDomain: app1.TeamDomain,
		},
		"should remove app1 and update DefaultTeamDomain to app2": {
			apps: Apps{
				DefaultAppTeamDomain: app1.TeamDomain,
				DeployedApps: map[string]App{
					app1TeamID: app1,
					app2TeamID: app2,
				},
			},
			removedTeamID:                app1.TeamID,
			expectedDefaultAppTeamDomain: app2.TeamDomain,
		},
		"should remove all apps and update DefaultTeamDomain to empty string": {
			apps: Apps{
				DefaultAppTeamDomain: app1.TeamDomain,
				DeployedApps: map[string]App{
					app1TeamID: app1, // Only 1 app, so slice is empty when app is removed
				},
			},
			removedTeamID:                app1.TeamID,
			expectedDefaultAppTeamDomain: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Assert app belonging to removedTeamID exists in apps
			require.NotEmpty(t, tt.apps.DeployedApps[tt.removedTeamID])

			// Remove the app
			tt.apps.RemoveDeployedByTeamID(tt.removedTeamID)

			// Assert app removed
			require.Empty(t, tt.apps.DeployedApps[tt.removedTeamID])

			// Assert DefaultAppTeamDomain change
			require.Equal(t, tt.expectedDefaultAppTeamDomain, tt.apps.DefaultAppTeamDomain)
		})
	}
}

func Test_AppInstallationStatus_String(t *testing.T) {
	var appInstallStatus AppInstallationStatus

	appInstallStatus = AppInstallationStatusUnknown
	require.Equal(t, "Unknown", appInstallStatus.String())

	appInstallStatus = AppStatusInstalled
	require.Equal(t, "Installed", appInstallStatus.String())

	appInstallStatus = AppStatusUninstalled
	require.Equal(t, "Uninstalled", appInstallStatus.String())
}

func Test_App_IsAppFlagValid(t *testing.T) {
	validAppFlags := []string{"local", "deploy", "deployed", "A0123456789"}
	invalidAppFlag := "bogusness"

	for _, flag := range validAppFlags {
		require.Equal(t, true, IsAppFlagValid(flag), "valid app flag should pass")
	}
	require.Equal(t, false, IsAppFlagValid(invalidAppFlag), "invalid app flag should not pass")
}

func Test_App_IsAppID(t *testing.T) {
	var validAppID = "A014YLZ969Y"
	var invalidAppID = "local"

	require.Equal(t, true, IsAppID(validAppID), "valid App ID should pass")
	require.Equal(t, false, IsAppID(invalidAppID), "invalid App ID should not pass")
}

func Test_App_IsAppFlagEnvironment(t *testing.T) {
	validAppFlags := []string{"local", "deploy", "deployed"}
	invalidAppFlag := "A456789"

	for _, flag := range validAppFlags {
		require.Equal(t, true, IsAppFlagEnvironment(flag), "valid environment app flag should pass")
	}
	require.Equal(t, false, IsAppFlagEnvironment(invalidAppFlag), "invalid environment app flag should not pass")
}

func Test_App_IsAppFlagDeploy(t *testing.T) {
	var validAppFlag = "deployed"
	require.Equal(t, true, IsAppFlagDeploy(validAppFlag), "valid deploy app flag should pass")

	validAppFlag = "deploy"
	require.Equal(t, true, IsAppFlagDeploy(validAppFlag), "valid deploy app flag should pass")

	var invalidAppFlag = "local"
	require.Equal(t, false, IsAppFlagDeploy(invalidAppFlag), "invalid deploy app flag should not pass")
}

func Test_App_IsAppFlagLocal(t *testing.T) {
	var validAppFlag = "local"
	var invalidAppFlag = "deployed"

	require.Equal(t, true, IsAppFlagLocal(validAppFlag), "valid local app flag should pass")
	require.Equal(t, false, IsAppFlagLocal(invalidAppFlag), "invalid local app flag should not pass")
}

func Test_App_IsEnterpriseWorkspaceApp_IsEnterpriseApp(t *testing.T) {
	var enterpriseWorkspaceApp = App{
		TeamDomain:   "acmecorp",
		TeamID:       "T123456789",
		AppID:        "A1234567894",
		EnterpriseID: "E123456789",
		UserID:       "U12345678",
	}

	require.Equal(t, true, enterpriseWorkspaceApp.IsEnterpriseWorkspaceApp(), "A valid enterprise workspace app should pass validation")
	require.Equal(t, false, enterpriseWorkspaceApp.IsEnterpriseApp(), "A valid enterprise workspace app should not validate as an enterprise app")

	var invalidMissingAppID = App{
		TeamDomain:   "acmecorp",
		TeamID:       "T123456789",
		EnterpriseID: "E123456789",
		UserID:       "U12345678",
	}

	require.Equal(t, false, invalidMissingAppID.IsEnterpriseWorkspaceApp(), "Cannot be a valid app without app ID")
	require.Equal(t, false, invalidMissingAppID.IsEnterpriseApp(), "Cannot be a valid app without app ID")

	var enterpriseApp = App{
		TeamDomain:   "acmecorp",
		TeamID:       "E123456789",
		AppID:        "A1234567894",
		EnterpriseID: "E123456789",
		UserID:       "U12345678",
	}

	require.Equal(t, false, enterpriseApp.IsEnterpriseWorkspaceApp(), "An enterprise app should not validate as an enterprise workspace app")
	require.Equal(t, true, enterpriseApp.IsEnterpriseApp(), "An enterprise app should validate as an enterprise app")
}

func Test_App_Equals(t *testing.T) {
	expectedApp := App{
		AppID:            "A",
		EnterpriseID:     "E123",
		new:              false,
		InstallStatus:    AppStatusInstalled,
		IsDev:            true,
		TeamDomain:       "Team Domain",
		TeamID:           "T123",
		UserID:           "U123",
		EnterpriseGrants: []EnterpriseGrant{{"T1", "T One"}, {"T2", "T Two"}},
	}

	tests := map[string]struct {
		app     App
		matches bool
	}{
		"matching app is equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: true,
		},
		"app_id field is not equal": {
			app: App{
				AppID:            "B",
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"enterprise_id field is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     "E456",
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"new property is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              !expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"InstallStatus property is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    AppStatusUninstalled,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"IsDev property is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            !expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"team_domain field is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       "Different Team Domain",
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"team_id field is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           "T456",
				UserID:           expectedApp.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"user_id field is not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           "U456",
				EnterpriseGrants: append([]EnterpriseGrant{}, expectedApp.EnterpriseGrants...),
			},
			matches: false,
		},
		"enterpriseGrants different length and not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: []EnterpriseGrant{{"T3", "T Three"}},
			},
			matches: false,
		},
		"enterpriseGrants same length but not equal": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: []EnterpriseGrant{{"T3", "T Three"}, {"T4", "T Four"}},
			},
			matches: false,
		},
		"enterpriseGrants are equal but ordered differently": {
			app: App{
				AppID:            expectedApp.AppID,
				EnterpriseID:     expectedApp.EnterpriseID,
				new:              expectedApp.new,
				InstallStatus:    expectedApp.InstallStatus,
				IsDev:            expectedApp.IsDev,
				TeamDomain:       expectedApp.TeamDomain,
				TeamID:           expectedApp.TeamID,
				UserID:           expectedApp.UserID,
				EnterpriseGrants: []EnterpriseGrant{{"T2", "T Two"}, {"T1", "T One"}},
			},
			matches: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.matches, expectedApp.Equals(tt.app))
		})
	}
}

func Test_App_IsEmpty(t *testing.T) {
	tests := map[string]struct {
		app     App
		isEmpty bool
	}{
		"should be empty": {
			app:     App{},
			isEmpty: true,
		},
		"should not be empty": {
			app: App{
				AppID:            app1.AppID,
				EnterpriseID:     app1.EnterpriseID,
				new:              app1.new,
				InstallStatus:    app1.InstallStatus,
				IsDev:            app1.IsDev,
				TeamDomain:       app1.TeamDomain,
				TeamID:           app1.TeamID,
				UserID:           app1.UserID,
				EnterpriseGrants: append([]EnterpriseGrant{}, app1.EnterpriseGrants...),
			},
			isEmpty: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.isEmpty, tt.app.IsEmpty())
		})
	}
}

func Test_App_IsNew(t *testing.T) {
	tests := map[string]struct {
		app   App
		isNew bool
	}{
		"should be new": {
			app: App{
				new: true, // Should be true
			},
			isNew: true,
		},
		"should not be new": {
			app: App{
				new: false, // Should be false
			},
			isNew: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.isNew, tt.app.IsNew())
		})
	}
}

func Test_App_NewApp(t *testing.T) {
	var app = NewApp()
	require.True(t, app.new)
	require.True(t, app.IsNew())
}

func Test_App_IsUninstalled(t *testing.T) {
	// Should be false when an app is installed
	app := App{
		InstallStatus: AppStatusInstalled,
	}
	require.False(t, app.IsUninstalled())

	// Should be true when an app is uninstalled
	app = App{
		InstallStatus: AppStatusUninstalled,
	}
	require.True(t, app.IsUninstalled())
}

func Test_App_IsInstalled(t *testing.T) {
	// Should be false when an app is uninstalled
	app := App{
		InstallStatus: AppStatusUninstalled,
	}
	require.False(t, app.IsInstalled())

	// Should be true when an app is installed
	app = App{
		InstallStatus: AppStatusInstalled,
	}
	require.True(t, app.IsInstalled())
}
