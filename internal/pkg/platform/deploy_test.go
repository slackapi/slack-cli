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
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

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
				UserName: strPtr("slackbot"),
				UserID:   strPtr("USLACKBOT"),
				TeamName: strPtr("speck"),
				TeamID:   strPtr("T001"),
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
				UserName:            strPtr("stub"),
				UserID:              strPtr("U111"),
				TeamName:            strPtr("spack"),
				TeamID:              strPtr("E002"),
				EnterpriseID:        strPtr("E002"),
				IsEnterpriseInstall: boolPtr(true),
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
