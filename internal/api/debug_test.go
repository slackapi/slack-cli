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
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/stretchr/testify/require"
)

func Test_RedactPII(t *testing.T) {
	home, _ := os.UserHomeDir()
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Simple case",
			text:     "hello world",
			expected: "hello world",
		},
		{
			name:     "Preserve the word XOXP",
			text:     "This is an XOXP token",
			expected: "This is an XOXP token",
		},
		{
			name:     "Redact actual XOXP token",
			text:     `{"ok":true,"token":"xoxe.xoxp-123","refresh_token":"xoxe-1-123","team_id":"T0123","user_id":"U0123", "xxtoken":"123"}`,
			expected: `{"ok":true,"token":"...","refresh_token":"...","team_id":"T0123","user_id":"U0123", "xxtoken":"..."}`,
		},
		{
			name:     "Redact home directory",
			text:     "found authorizations at " + home + "/.slack/credentials.json reading",
			expected: `found authorizations at .../.slack/credentials.json reading`,
		},
		{
			name:     "Redact username with single quotes",
			text:     `'user':'username'`,
			expected: `'user':"..."`,
		},
		{
			name:     "Redact username with double quotes",
			text:     `"user":"username"`,
			expected: `"user":"..."`,
		},
		{
			name:     "Redact username with no quotes",
			text:     `user:username`,
			expected: `user:username`,
		},
		{
			name:     "Redact username in http response",
			text:     `{"ok":true,"token":"xoxe.xoxp-123","refresh_token":"xoxe-1-123","team_id":"T0123","user_id":"U0123", "xxtoken":"123", "user":"username"}`,
			expected: `{"ok":true,"token":"...","refresh_token":"...","team_id":"T0123","user_id":"U0123", "xxtoken":"...", "user":"..."}`,
		},
		{
			name:     "Preserve the word XOXE",
			text:     "This is an XOXE token",
			expected: "This is an XOXE token",
		},
		{
			name:     "Redact actual token in HTTP request",
			text:     "HTTP Request Body:refresh_token=xoxe-1",
			expected: `HTTP Request Body:refresh_token=...`,
		},
		{
			name:     "Display Trace ID in log",
			text:     "TraceID: 123",
			expected: `TraceID: 123`,
		},
		{
			name:     "Display Team ID in log",
			text:     "TeamID: T123",
			expected: `TeamID: T123`,
		},
		{
			name:     "Display User ID in log",
			text:     "UserID: U123",
			expected: `UserID: U123`,
		},
		{
			name:     "Display Slack-CLI version in log",
			text:     "Slack-CLI Version: v1.10.0",
			expected: `Slack-CLI Version: v1.10.0`,
		},
		{
			name:     "Display user's OS in log",
			text:     "Operating System (OS): darwin",
			expected: `Operating System (OS): darwin`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := goutils.RedactPII(tt.text)
			require.Equal(t, redacted, tt.expected)
		})
	}
}
