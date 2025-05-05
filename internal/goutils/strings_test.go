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

package goutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_HashString(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		expected bool
	}{
		{
			name:     "happy path - same string",
			text1:    "one",
			text2:    "one",
			expected: true,
		},
		{
			name:     "different strings",
			text1:    "one",
			text2:    "two",
			expected: false,
		},
		{
			name:     "almost identical",
			text1:    "one ", // add a space
			text2:    "one",
			expected: false,
		},
		{
			name:     "empty strings should be same",
			text1:    "",
			text2:    "",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err1 := HashString(tt.text1)
			hash2, err2 := HashString(tt.text2)
			require.Equal(t, hash1 == hash2, tt.expected)
			require.NoError(t, err1)
			require.NoError(t, err2)
		})
	}
}

func Test_ExtractFirstJSONFromString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "one json - 1",
			text:     "{}",
			expected: "{}",
		},
		{
			name:     "one json - 2",
			text:     "blah blah blah {a: 1}",
			expected: "{a: 1}",
		},
		{
			name:     "one json - 3",
			text:     "blah {a: 1} blah blah",
			expected: "{a: 1}",
		},
		{
			name:     "multiple json",
			text:     "{a: 1} {b: 2}",
			expected: "{a: 1}",
		},
		{
			name:     "no json present",
			text:     "foo bar",
			expected: "",
		},
		{
			name:     "nested json",
			text:     "{a: b: {c: {d: 1}}} {1}",
			expected: "{a: b: {c: {d: 1}}}",
		},
		{
			name:     "nested json - 2",
			text:     "{{{}}}",
			expected: "{{{}}}",
		},
		{
			name:     "malformed json",
			text:     "{{",
			expected: "",
		},
		{
			name:     "malformed json - 2",
			text:     "{{}",
			expected: "",
		},
		{
			name:     "malformed json - 3",
			text:     "}}",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualRes := ExtractFirstJSONFromString(tt.text)
			require.Equal(t, tt.expected, actualRes)
		})
	}
}

func Test_addLogWhenValExist(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		val      string
		expected string
	}{
		{
			name:     "Empty string when val is empty",
			title:    "hello world",
			val:      "",
			expected: "",
		},
		{
			name:     "Empty string when val only has space",
			title:    "hello world",
			val:      " ",
			expected: "",
		},
		{
			name:     "Return string when val is not empty",
			title:    "hello world",
			val:      "slack",
			expected: "hello world: [slack]\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := AddLogWhenValExist(tt.title, tt.val)
			require.Equal(t, output, tt.expected)
		})
	}
}
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
			name:     "Don't redact username with no quotes",
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
			name:     "App Token (xapp) as JSON value",
			text:     `{"ok":true,"api_access_tokens":{"app_level":"xapp-1-A000-1111-ABCD"}}`,
			expected: `{"ok":true,"api_access_tokens":{"app_level":"..."}}`,
		},
		{
			name:     "App Token (xapp) as open text",
			text:     `Logging app token xapp-1-A000-1111-ABCD in the output`,
			expected: `Logging app token ... in the output`,
		},
		{
			name:     "Bot Token (xoxb) as JSON value",
			text:     `{"ok":true,"api_access_tokens":{"bot":"xoxb-1111-2222-ABCD"}}`,
			expected: `{"ok":true,"api_access_tokens":{"bot":"..."}}`,
		},
		{
			name:     "Bot Token (xoxb) as text",
			text:     `Logging bot token xoxb-1111-2222-ABCD in the output`,
			expected: `Logging bot token ... in the output`,
		},
		{
			name:     "User Token (xoxp) as JSON value",
			text:     `{"ok":true,"api_access_tokens":{"user":"xoxp-1111-2222-ABCD"}}`,
			expected: `{"ok":true,"api_access_tokens":{"user":"..."}}`,
		},
		{
			name:     "User Token (xoxp) as text",
			text:     `Logging user token xoxp-1111-2222-ABCD in the output`,
			expected: `Logging user token ... in the output`,
		},
		{
			name:     "Refresh Token (xoxe) as JSON value",
			text:     `{"ok":true,"refresh_token":"xoxe-1-A000-1111-ABCD}`,
			expected: `{"ok":true,"refresh_token":"..."}`,
		},
		{
			name:     "Refresh Token (xoxe) as text",
			text:     `Logging user token xoxe-1-A000-1111-ABCD in the output`,
			expected: `Logging user token ... in the output`,
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
		{
			name:     "Escape oauth_authorize_url",
			text:     `"oauth_authorize_url":"www.fake.com"`,
			expected: `"oauth_authorize_url":"..."`,
		},
		{
			name:     "Escape provider_key",
			text:     `"provider_key":"provider_key"`,
			expected: `"provider_key":"..."`,
		},
		{
			name:     "Escape authorizations",
			text:     `"authorizations":"authorizations"`,
			expected: `"authorizations":"..."`,
		},
		{
			name:     "Escape authorization_url",
			text:     `"authorization_url":"authorization_url"`,
			expected: `"authorization_url":"..."`,
		},
		{
			name:     "Escape secret",
			text:     `"secret":"secret"`,
			expected: `"secret":"..."`,
		},
		{
			name:     "Escape secret with prefix",
			text:     `"client_secret":"secret"`,
			expected: `"client_secret":"..."`,
		},
		{
			name:     "Escape client_id",
			text:     `"client_id":"client_id"`,
			expected: `"client_id":"..."`,
		},
		{
			name:     "Escape variables",
			text:     `"variables":[{"foo":"bar", "hello":"world"}]`,
			expected: `"variables":"..."`,
		},
		{
			name:     "Escape sensitive data from mock HTTP response",
			text:     `{"ok":true,"app_id":"A123","credentials":{"client_id":"123","client_secret":"123","verification_token":"123","signing_secret":"123"},"oauth_authorize_url":"123":\/\/slack.com\/oauth\/v2\/authorize?client_id=123&scope=commands,chat:write,chat:write.public"}`,
			expected: `{"ok":true,"app_id":"A123","credentials":{"client_id":"...","client_secret":"...","verification_token":"...","signing_secret":"..."},"oauth_authorize_url":"...":\/\/slack.com\/oauth\/v2\/authorize?client_id=...&scope=commands,chat:write,chat:write.public"}`,
		},
		{
			name:     "Escape from `Command` for external-auth add-secret",
			text:     `slack external-auth add-secret --provider google --secret 123abcd`,
			expected: "slack external-auth add-secret ...",
		},
		{
			name:     "Escape from `Command` for var add",
			text:     `slack var add topsecret 123`,
			expected: `slack var add ...`,
		},
		{
			name:     "Escape from `Command` for external-auth add",
			text:     `slack external-auth add topsecret 123`,
			expected: `slack external-auth add ...`,
		},
		{
			name:     "Escape from `Command` for var remove",
			text:     `slack var remove topsecret 123`,
			expected: `slack var remove ...`,
		},
		{
			name:     "Escape from `Command` for env add",
			text:     `slack env add topsecret 123`,
			expected: `slack env add ...`,
		},
		{
			name:     "Escape from `Command` for env remove",
			text:     `slack env remove topsecret 123`,
			expected: `slack env remove ...`,
		},
		{
			name:     "Escape from `Command` for vars add",
			text:     `slack vars add topsecret 123`,
			expected: `slack vars add ...`,
		},
		{
			name:     "Escape from `Command` for vars remove",
			text:     `slack vars remove topsecret 123`,
			expected: `slack vars remove ...`,
		},
		{
			name:     "Escape from `Command` for variables add",
			text:     `slack variables add topsecret 123`,
			expected: `slack variables add ...`,
		},
		{
			name:     "Escape from `Command` for variables remove",
			text:     `slack variables remove topsecret 123`,
			expected: `slack variables remove ...`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := RedactPII(tt.text)
			require.Equal(t, tt.expected, redacted)
		})
	}
}
func Test_UpperCaseTrimAll(t *testing.T) {
	tests := []struct {
		name          string
		namedEntities string
		expected      string
	}{
		{
			name:          "Empty string when val is empty",
			namedEntities: "",
			expected:      "",
		},
		{
			name:          "Empty string when val only has space",
			namedEntities: " ",
			expected:      "",
		},
		{
			name:          "Return string when val is all lower-case",
			namedEntities: "slack",
			expected:      "SLACK",
		},
		{
			name:          "Return string when val contains spaces",
			namedEntities: "hello,   world",
			expected:      "HELLO,WORLD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := UpperCaseTrimAll(tt.namedEntities)
			require.Equal(t, output, tt.expected)
		})
	}
}

func Test_ToHTTPS(t *testing.T) {
	tests := []struct {
		name     string
		urlAddr  string
		expected string
	}{
		{
			name:     "url with https protocol",
			urlAddr:  "https://www.xyz.com",
			expected: "https://www.xyz.com",
		},
		{
			name:     "url with http protocol",
			urlAddr:  "http://www.xyz.com",
			expected: "https://www.xyz.com",
		},
		{
			name:     "url with no protocol",
			urlAddr:  "www.xyz.com",
			expected: "https://www.xyz.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ToHTTPS(tt.urlAddr)
			require.Equal(t, output, tt.expected)
		})
	}
}
