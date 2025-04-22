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

package app

import (
	"context"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Setup a mock for the package
type ListPkgMock struct {
	mock.Mock
}

func (m *ListPkgMock) List(ctx context.Context, clients *shared.ClientFactory) ([]types.App, string, error) {
	m.Called()
	return []types.App{}, "", nil
}

func TestAppsListCommand(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewListCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	listPkgMock := new(ListPkgMock)
	listFunc = listPkgMock.List

	listPkgMock.On("List").Return(nil)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	listPkgMock.AssertCalled(t, "List")
}

func TestAppsListFormat(t *testing.T) {
	mockTeam1Deploy := types.App{
		AppID:         "A1234",
		TeamID:        "T0001",
		TeamDomain:    "teamone",
		InstallStatus: types.AppInstallationStatus(1), // Installed
	}
	mockTeam1Local := types.App{
		AppID:         "A1235",
		TeamID:        "T0001",
		TeamDomain:    "teamone (local)",
		UserID:        "U9999",
		IsDev:         true,
		InstallStatus: types.AppInstallationStatus(2), // Uninstalled
	}
	mockTeam2Deploy := types.App{
		AppID:         "A0001",
		TeamID:        "T0002",
		TeamDomain:    "teamtwo",
		InstallStatus: types.AppInstallationStatus(0), // Unknown
	}
	mockTeam2Local := types.App{
		AppID:         "A0004",
		TeamID:        "T0002",
		TeamDomain:    "teamtwo", // no auth, so retrieved without the (local) tag
		UserID:        "U9997",
		IsDev:         true,
		InstallStatus: types.AppInstallationStatus(0), // Unknown
	}
	mockEnterpriseTeamDeploy := types.App{
		AppID:            "A0001",
		TeamID:           "E0002",
		TeamDomain:       "domain",
		InstallStatus:    types.AppInstallationStatus(1), // Installed
		EnterpriseID:     "E0002",
		EnterpriseGrants: []types.EnterpriseGrant{{WorkspaceDomain: "Domain", WorkspaceID: "ID"}},
	}
	mockEnterpriseTeamLocal := types.App{
		AppID:         "A0002",
		TeamID:        "E0002",
		TeamDomain:    "domain",
		InstallStatus: types.AppInstallationStatus(1), // Installed
		EnterpriseID:  "E0002",
		EnterpriseGrants: []types.EnterpriseGrant{
			{WorkspaceDomain: "Domain1", WorkspaceID: "ID1"},
			{WorkspaceDomain: "Domain2", WorkspaceID: "ID2"},
			{WorkspaceDomain: "Domain3", WorkspaceID: "ID3"},
			{WorkspaceDomain: "Domain4", WorkspaceID: "ID4"}},
	}
	mockGhostApp := types.App{
		AppID:         "",
		TeamID:        "T611",
		TeamDomain:    "spooky",
		InstallStatus: types.AppInstallationStatus(0), // Unknown
	}

	tests := map[string]struct {
		Apps     []types.App
		Expected []string
		Flags    listCmdFlags
	}{
		"no apps exist for the project": {
			Apps:     []types.App{},
			Expected: []string{"This project has no apps"},
		},
		"apps without an app ID are skipped": {
			Apps:     []types.App{mockGhostApp},
			Expected: []string{"This project has no apps"},
		},
		"a single installed deployed app in a single standalone workspace": {
			Apps: []types.App{mockTeam1Deploy},
			Expected: []string{
				"teamone",
				"App  ID: A1234",
				"Team ID: T0001",
				"Status:  Installed",
			},
		},
		"a single uninstalled local app in a single standalone workspace": {
			Apps: []types.App{mockTeam1Local},
			Expected: []string{
				"teamone (local)",
				"App  ID: A1235",
				"Team ID: T0001",
				"User ID: U9999",
				"Status:  Uninstalled",
			},
		},
		"a single unauthed local app in a single standalone workspace": {
			Apps: []types.App{mockTeam2Local},
			Expected: []string{
				"teamtwo (local)",
				"App  ID: A0004",
				"Team ID: T0002",
				"User ID: U9997",
				"Status:  Unknown",
			},
		},
		"multiple apps with various authentication statuses": {
			Apps: []types.App{mockTeam1Local, mockTeam2Deploy},
			Expected: []string{
				"teamone (local)",
				"App  ID: A0001",
				"Team ID: T0002",
				"User ID: U9999",
				"Status:  Uninstalled",
				"teamtwo",
				"App  ID: A0001",
				"Team ID: T0002",
				"Status:  Unknown",
			},
		},
		"enterprise app with single workspace grant": {
			Apps: []types.App{mockEnterpriseTeamDeploy},
			Expected: []string{
				"domain",
				"App  ID: A0001",
				"Team ID: E0002",
				"Status:  Installed",
				"Workspace Grant: Domain (ID)",
			},
		},
		"enterprise app with many workspace grants": {
			Apps: []types.App{mockEnterpriseTeamLocal},
			Expected: []string{
				"App  ID: A0002",
				"Team ID: E0002",
				"Status:  Installed",
				"Workspace Grants:",
				"Domain1 (ID1)",
				"Domain2 (ID2)",
				"Domain3 (ID3)",
				"... and 1 other workspace",
			},
		},
		"display all grants for enterprise app with many workspace grants": {
			Apps:  []types.App{mockEnterpriseTeamLocal},
			Flags: listCmdFlags{displayAllOrgGrants: true},
			Expected: []string{
				"App  ID: A0002",
				"Team ID: E0002",
				"Status:  Installed",
				"Workspace Grants:",
				"Domain1 (ID1)",
				"Domain2 (ID2)",
				"Domain3 (ID3)",
				"Domain4 (ID4)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			listFlags = tt.Flags
			formattedList := formatListSuccess(tt.Apps)
			for ii, value := range formattedList {
				formattedList[ii] = strings.TrimRight(value, ":")
			}
			for _, value := range tt.Expected {
				assert.Contains(t, strings.Join(formattedList, "\n"), value)
			}
		})
	}
}
