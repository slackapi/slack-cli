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

package api

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_printRequest(t *testing.T) {
	tests := map[string]struct {
		userAgent string
		expected  string
	}{
		"includes User-Agent header in output": {
			userAgent: "slack-cli/v1.2.3 (os: darwin)",
			expected:  "HTTP Request User-Agent: slack-cli/v1.2.3 (os: darwin)",
		},
		"includes empty User-Agent when header is not set": {
			userAgent: "",
			expected:  "HTTP Request User-Agent: ",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fs := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			cfg := config.NewConfig(fs, osMock)
			cfg.DebugEnabled = true
			ioMock := iostreams.NewIOStreamsMock(cfg, fs, osMock)
			ioMock.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))

			c := &Client{io: ioMock}
			req, _ := http.NewRequest("GET", "https://slack.com/api/test", nil)
			if tc.userAgent != "" {
				req.Header.Set("User-Agent", tc.userAgent)
			}

			c.printRequest(ctx, req, false)

			var output string
			for _, call := range ioMock.Calls {
				if call.Method == "PrintDebug" {
					output = fmt.Sprintf(call.Arguments[1].(string), call.Arguments[2].([]any)...)
				}
			}
			require.Contains(t, output, tc.expected)
		})
	}
}

func Test_RedactPII(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := map[string]struct {
		text     string
		expected string
	}{
		"Simple case": {
			text:     "hello world",
			expected: "hello world",
		},
		"Preserve the word XOXP": {
			text:     "This is an XOXP token",
			expected: "This is an XOXP token",
		},
		"Redact actual XOXP token": {
			text:     `{"ok":true,"token":"xoxe.xoxp-123","refresh_token":"xoxe-1-123","team_id":"T0123","user_id":"U0123", "xxtoken":"123"}`,
			expected: `{"ok":true,"token":"...","refresh_token":"...","team_id":"T0123","user_id":"U0123", "xxtoken":"..."}`,
		},
		"Redact home directory": {
			text:     "found authorizations at " + home + "/.slack/credentials.json reading",
			expected: `found authorizations at .../.slack/credentials.json reading`,
		},
		"Redact username with single quotes": {
			text:     `'user':'username'`,
			expected: `'user':"..."`,
		},
		"Redact username with double quotes": {
			text:     `"user":"username"`,
			expected: `"user":"..."`,
		},
		"Redact username with no quotes": {
			text:     `user:username`,
			expected: `user:username`,
		},
		"Redact username in http response": {
			text:     `{"ok":true,"token":"xoxe.xoxp-123","refresh_token":"xoxe-1-123","team_id":"T0123","user_id":"U0123", "xxtoken":"123", "user":"username"}`,
			expected: `{"ok":true,"token":"...","refresh_token":"...","team_id":"T0123","user_id":"U0123", "xxtoken":"...", "user":"..."}`,
		},
		"Preserve the word XOXE": {
			text:     "This is an XOXE token",
			expected: "This is an XOXE token",
		},
		"Redact actual token in HTTP request": {
			text:     "HTTP Request Body:refresh_token=xoxe-1",
			expected: `HTTP Request Body:refresh_token=...`,
		},
		"Display Trace ID in log": {
			text:     "TraceID: 123",
			expected: `TraceID: 123`,
		},
		"Display Team ID in log": {
			text:     "TeamID: T123",
			expected: `TeamID: T123`,
		},
		"Display User ID in log": {
			text:     "UserID: U123",
			expected: `UserID: U123`,
		},
		"Display Slack-CLI version in log": {
			text:     "Slack-CLI Version: v1.10.0",
			expected: `Slack-CLI Version: v1.10.0`,
		},
		"Display user's OS in log": {
			text:     "Operating System (OS): darwin",
			expected: `Operating System (OS): darwin`,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			redacted := goutils.RedactPII(tc.text)
			require.Equal(t, redacted, tc.expected)
		})
	}
}
