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
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_deleteAppOnTerminate(t *testing.T) {
	tests := map[string]struct {
		app            types.App
		deleteErr      error
		expectedOutput []string
		unexpected     string
	}{
		"prints that the local app was cleaned up when the delete succeeds": {
			app:            types.App{AppID: "A123", TeamDomain: "ws", TeamID: "T123", IsDev: true},
			expectedOutput: []string{`Cleaned up local app install for "Team"`},
			unexpected:     "Local app cleanup failed",
		},
		"prints a warning when the delete request fails": {
			app:       types.App{AppID: "A123", TeamDomain: "ws", TeamID: "T123", IsDev: true},
			deleteErr: slackerror.New(slackerror.ErrAppNotFound),
			expectedOutput: []string{
				`Cleaning up local app install for "Team" failed`,
				"Local app cleanup failed",
			},
			unexpected: "Cleaned up local app install",
		},
		"prints a warning when the app has no app ID": {
			app: types.App{TeamDomain: "ws", TeamID: "T123", IsDev: true},
			expectedOutput: []string{
				`Cleaning up local app install for "Team" failed`,
				"Local app cleanup failed",
			},
			unexpected: "Cleaned up local app install",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			teamDomain := "ws"
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("api host")
			clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything).Return("logstash host")
			clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{TeamName: &teamDomain}, nil)
			clientsMock.API.On("DeleteApp", mock.Anything, mock.Anything, mock.Anything).Return(tc.deleteErr)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			auth := types.SlackAuth{TeamDomain: "ws", TeamID: "T123", Token: "xoxb-example"}

			deleteAppOnTerminate(ctx, clients, auth, tc.app, "Team")

			for _, expected := range tc.expectedOutput {
				assert.Contains(t, clientsMock.GetCombinedOutput(), expected)
			}
			assert.NotContains(t, clientsMock.GetCombinedOutput(), tc.unexpected)
		})
	}
}
