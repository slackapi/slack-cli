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

package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_AuthRevokeToken(t *testing.T) {
	tests := map[string]struct {
		token    string
		response string
		warning  string
		expected error
	}{
		"succeeds when the token is revoked without errors": {
			token:    "xoxb-example-0001",
			response: `{"ok":true}`,
		},
		"continues if the revoke error is already logged out": {
			token:    "xoxb-example-0001",
			response: `{"ok":false,"error":"already_logged_out"}`,
			warning:  slackerror.New(slackerror.ErrAlreadyLoggedOut).Message,
		},
		"continues if the revoke error is invalid auth": {
			token:    "xoxb-example-0001",
			response: `{"ok":false,"error":"invalid_auth"}`,
			warning:  slackerror.New(slackerror.ErrInvalidAuth).Message,
		},
		"continues if the revoke error is token expired": {
			token:    "xoxb-example-0001",
			response: `{"ok":false,"error":"token_expired"}`,
			warning:  slackerror.New(slackerror.ErrTokenExpired).Message,
		},
		"continues if the revoke error is token revoked": {
			token:    "xoxb-example-0001",
			response: `{"ok":false,"error":"token_revoked"}`,
			warning:  slackerror.New(slackerror.ErrTokenRevoked).Message,
		},
		"errors if the revoke error is an unexpected error": {
			token:    "xoxb-example-0001",
			response: `{"ok":false,"error":"not_found"}`,
			expected: slackerror.New(slackerror.ErrNotFound).AddApiMethod("auth.revoke"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			fs := slackdeps.NewFsMock()
			config := config.NewConfig(fs, os)
			io := iostreams.NewIOStreamsMock(config, fs, os)
			io.AddDefaultMocks()
			apic, teardown := api.NewFakeClient(t, api.FakeClientParams{
				ExpectedMethod:  "auth.revoke",
				ExpectedRequest: fmt.Sprintf("token=%s", tt.token),
				Response:        tt.response,
			})
			defer teardown()
			appc := app.NewClient(apic, config, fs, os)
			auth := NewClient(apic, appc, config, io, fs)
			err := auth.RevokeToken(ctx, tt.token)
			assert.Equal(t, tt.expected, err)
			if tt.warning != "" {
				io.AssertCalled(t, "PrintDebug", mock.Anything, "%s.", []any{tt.warning})
			} else {
				io.AssertNotCalled(t, "PrintDebug")
			}
		})
	}
}
