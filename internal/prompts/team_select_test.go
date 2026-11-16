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

package prompts

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPromptTeamSlackAuth(t *testing.T) {
	tests := map[string]struct {
		auths        []types.SlackAuth
		expectedAuth types.SlackAuth
	}{
		"selects the only authenticated account without a prompt": {
			auths: []types.SlackAuth{
				{Token: team1Token, TeamID: team1TeamID, TeamDomain: team1TeamDomain},
			},
			expectedAuth: types.SlackAuth{Token: team1Token, TeamID: team1TeamID, TeamDomain: team1TeamDomain},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On(Auths, mock.Anything).Return(tc.auths, nil)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			auth, err := PromptTeamSlackAuth(ctx, clients, "Select a team", nil)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedAuth, *auth)
			clientsMock.Auth.AssertCalled(t, "SetSelectedAuth", mock.Anything, tc.expectedAuth, clients.Config, clients.Os)
			clientsMock.IO.AssertNotCalled(t, SelectPrompt)
		})
	}
}
