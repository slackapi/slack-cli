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

package platform

import (
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAPIHostVariable(t *testing.T) {
	tests := map[string]struct {
		apiHostFlag     string
		apiHostResolved string
		isSlackDev      bool
		addVariableErr  error
		expectUpdate    bool
	}{
		"explicit custom host is added": {
			apiHostFlag:     "api.your.test.endpoint",
			apiHostResolved: "https://api.your.test.endpoint",
			expectUpdate:    true,
		},
		"explicit custom host update returns an error": {
			apiHostFlag:     "api.your.test.endpoint",
			apiHostResolved: "https://api.your.test.endpoint",
			addVariableErr:  errors.New("variable update failed"),
			expectUpdate:    true,
		},
		"development host without an explicit flag is added": {
			apiHostResolved: "https://dev.slack.com",
			isSlackDev:      true,
			expectUpdate:    true,
		},
		"production host without an explicit flag is not added": {
			apiHostResolved: "https://slack.com",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := config.SetContextToken(context.Background(), "token")
			clientsMock := shared.NewClientsMock()
			clientsMock.Config.APIHostFlag = tc.apiHostFlag
			clientsMock.Config.APIHostResolved = tc.apiHostResolved
			if tc.apiHostFlag == "" {
				clientsMock.Auth.On("IsAPIHostSlackDev", tc.apiHostResolved).Return(tc.isSlackDev).Once()
			}
			if tc.expectUpdate {
				clientsMock.API.On(
					"AddVariable",
					ctx,
					"token",
					"A123",
					"SLACK_API_URL",
					tc.apiHostResolved+"/api/",
				).Return(tc.addVariableErr).Once()
			}
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			err := setAPIHostVariable(ctx, clients, "A123")

			require.ErrorIs(t, err, tc.addVariableErr)
			clientsMock.API.AssertNumberOfCalls(t, "AddVariable", map[bool]int{true: 1, false: 0}[tc.expectUpdate])
			clientsMock.Auth.AssertExpectations(t)
			clientsMock.API.AssertExpectations(t)
		})
	}
}

func TestDeploySuccessText(t *testing.T) {
	tests := map[string]struct {
		app         types.App
		manifest    types.SlackYaml
		authSession api.AuthSession
		deployTime  string
		expected    []string
	}{
		"information from a workspace deploy is printed": {
			app: types.App{AppID: "A123"},
			manifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{Name: "DeployerApp"},
				},
			},
			authSession: api.AuthSession{
				UserName: new("slackbot"),
				UserID:   new("USLACKBOT"),
				TeamName: new("speck"),
				TeamID:   new("T001"),
			},
			deployTime: "12.34",
			expected: []string{
				"DeployerApp deployed in 12.34",
				"Dashboard:  https://slacker.com/apps/A123",
				"App Owner:  slackbot (USLACKBOT)",
				"Workspace:  speck (T001)",
			},
		},
		"information from an enterprise deploy is printed": {
			app: types.App{AppID: "A999"},
			manifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{Name: "Spackulen"},
				},
			},
			authSession: api.AuthSession{
				UserName:            new("stub"),
				UserID:              new("U111"),
				TeamName:            new("spack"),
				TeamID:              new("E002"),
				EnterpriseID:        new("E002"),
				IsEnterpriseInstall: new(bool(true)),
			},
			deployTime: "8.05",
			expected: []string{
				"Spackulen deployed in 8.05",
				"Dashboard   :  https://slacker.com/apps/A999",
				"App Owner   :  stub (U111)",
				"Organization:  spack (E002)",
			},
		},
		"a message is still displayed with missing info": {
			app:         types.App{},
			manifest:    types.SlackYaml{},
			authSession: api.AuthSession{},
			expected: []string{
				"Successfully deployed the app!",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.API.On("Host").Return("https://slacker.com")
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			output := deploySuccessText(clients, tc.app, tc.manifest, tc.authSession, tc.deployTime)
			for _, line := range tc.expected {
				assert.Contains(t, output, line)
			}
		})
	}
}
