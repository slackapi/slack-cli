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

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func Test_API_TeamInfoResponse(t *testing.T) {
	tests := []struct {
		name                  string
		argsToken             string
		argsTeamID            string
		httpResponseJSON      string
		expectedTeamsInfo     *types.TeamInfo
		expectedErrorContains string
	}{
		{
			name:             "Successful request",
			argsToken:        "xoxp-123",
			argsTeamID:       "T0123",
			httpResponseJSON: `{"ok": true,  "team": { "id": "T12345", "name": "My Team", "domain": "example", "email_domain": "example.com", "enterprise_id": "E1234A12AB", "enterprise_name": "Umbrella Corporation" }}`,
			expectedTeamsInfo: &types.TeamInfo{
				ID:   "T12345",
				Name: "My Team",
			},
			expectedErrorContains: "",
		},
		{
			name:                  "Response contains an error",
			argsToken:             "xoxp-123",
			argsTeamID:            "T0123",
			httpResponseJSON:      `{"ok":false,"error":"team_not_found"}`,
			expectedErrorContains: "team_not_found",
		},
		{
			name:                  "Response contains invalid JSON",
			argsToken:             "xoxp-123",
			argsTeamID:            "T0123",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHttpResponseInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// Setup HTTP test server
			httpHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
				json := tt.httpResponseJSON
				_, err := fmt.Fprintln(w, json)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(httpHandlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)

			// Execute test
			actual, err := apiClient.TeamsInfo(ctx, tt.argsToken, tt.argsTeamID)

			// Assertions
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedTeamsInfo, actual)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
