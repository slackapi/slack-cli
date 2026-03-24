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

package auth

import (
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCommand(t *testing.T) {
	tests := map[string]struct {
		auths    []types.SlackAuth
		expected []string
	}{
		"no authorized accounts": {
			auths: []types.SlackAuth{},
			expected: []string{
				"You are not logged in to any Slack accounts",
				"login",
			},
		},
		"a single authorized account": {
			auths: []types.SlackAuth{
				{
					TeamDomain:  "test-workspace",
					TeamID:      "T12345",
					UserID:      "U67890",
					LastUpdated: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			expected: []string{
				"test-workspace",
				"T12345",
				"U67890",
			},
		},
		"multiple authorized accounts": {
			auths: []types.SlackAuth{
				{
					TeamDomain:  "alpha-workspace",
					TeamID:      "T11111",
					UserID:      "U11111",
					LastUpdated: time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC),
				},
				{
					TeamDomain:  "beta-workspace",
					TeamID:      "T22222",
					UserID:      "U22222",
					LastUpdated: time.Date(2025, 2, 20, 16, 0, 0, 0, time.UTC),
				},
			},
			expected: []string{
				"alpha-workspace",
				"T11111",
				"U11111",
				"beta-workspace",
				"T22222",
				"U22222",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Auth.On("Auths", mock.Anything).Return(tc.auths, nil)
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			cmd := NewListCommand(clients)
			testutil.MockCmdIO(clients.IO, cmd)

			err := cmd.ExecuteContext(ctx)
			assert.NoError(t, err)

			output := clientsMock.GetCombinedOutput()
			for _, expected := range tc.expected {
				assert.Contains(t, output, expected)
			}
		})
	}
}
