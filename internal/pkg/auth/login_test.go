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
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Login(t *testing.T) {
	tests := map[string]struct {
		userToken    string
		setup        func(cm *shared.ClientsMock)
		expectedAuth types.SlackAuth
		expectedPath string
		expectedErr  string
	}{
		"saves the auth for a valid user token": {
			userToken: "xoxp-example",
			setup: func(cm *shared.ClientsMock) {
				teamID, userID, url := "T123", "U123", "https://example.slack.com/"
				cm.API.On("ValidateSession", mock.Anything, "xoxp-example").
					Return(api.AuthSession{TeamID: &teamID, UserID: &userID, URL: &url}, nil)
				cm.API.On("Host").Return("https://slack.com")
				cm.Auth.On("IsAPIHostSlackProd", "https://slack.com").Return(true)
				cm.Auth.On("AuthWithTeamDomain", mock.Anything, "example").
					Return(types.SlackAuth{Token: "xoxp-existing"}, nil)
				cm.Auth.On("SetAuth", mock.Anything, mock.MatchedBy(func(a types.SlackAuth) bool {
					return a.Token == "xoxp-example" && a.TeamID == "T123" && a.UserID == "U123" && a.TeamDomain == "example" && a.APIHost == nil
				})).
					Return(types.SlackAuth{TeamDomain: "example"}, "/path/credentials.json", nil)
			},
			expectedAuth: types.SlackAuth{TeamDomain: "example"},
			expectedPath: "/path/credentials.json",
		},
		"returns an invalid auth error when the user token fails validation": {
			userToken: "xoxp-invalid",
			setup: func(cm *shared.ClientsMock) {
				cm.API.On("ValidateSession", mock.Anything, "xoxp-invalid").
					Return(api.AuthSession{}, errors.New("invalid_auth"))
			},
			expectedErr: slackerror.ErrInvalidAuth,
		},
		"returns the error from requesting an auth ticket without a user token": {
			setup: func(cm *shared.ClientsMock) {
				cm.API.On("GenerateAuthTicket", mock.Anything, mock.Anything, false).
					Return(api.GenerateAuthTicketResult{}, slackerror.New(slackerror.ErrHTTPRequestFailed))
			},
			expectedErr: slackerror.ErrHTTPRequestFailed,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			tc.setup(cm)

			auth, path, err := Login(ctx, cm.API, cm.Auth, cm.Config, cm.IO, tc.userToken, false)
			if tc.expectedErr != "" {
				assert.Equal(t, tc.expectedErr, slackerror.ToSlackError(err).Code)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAuth, auth)
			assert.Equal(t, tc.expectedPath, path)
		})
	}
}
