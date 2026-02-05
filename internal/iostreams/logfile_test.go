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

package iostreams

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			redacted := goutils.RedactPII(tt.text)
			require.Equal(t, redacted, tt.expected)
		})
	}
}

func Test_isOlderThanOneWeek(t *testing.T) {
	tests := map[string]struct {
		date     time.Time
		expected bool
	}{
		"False when log file is created within a week": {
			date:     time.Now().AddDate(0, 0, -6),
			expected: false,
		},
		"True when log file is created more than a week": {
			date:     time.Now().AddDate(0, 0, -8),
			expected: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output := isOlderThanOneWeek(tt.date)
			require.Equal(t, output, tt.expected)
		})
	}
}

func Test_FlushToLogFile(t *testing.T) {
	tests := map[string]struct {
		prefix       string
		errStr       string
		expectedFile []string
		suggestsFile bool
	}{
		"errors are flushed to file": {
			prefix:       "error",
			errStr:       "ohno",
			expectedFile: []string{"ERROR", "ohno\n"},
			suggestsFile: true,
		},
		"debugs are flushed to file": {
			prefix:       "debug",
			errStr:       "s p a c e",
			expectedFile: []string{"DEBUG", "s p a c e\n"},
			suggestsFile: false,
		},
		"ansi is removed from strings": {
			prefix:       "warn",
			errStr:       "go\x1b[1mgo\x1b[0mgo",
			expectedFile: []string{"WARN", "gogogo\n"},
			suggestsFile: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var io *IOStreams
			stderrBuff := bytes.Buffer{}
			stderrLogger := log.Logger{}
			stderrLogger.SetOutput(&stderrBuff)
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.AddDefaultMocks()
			var systemConfigMock config.SystemConfigMock
			mockLogsDir := filepath.Join("logs")
			systemConfigMock.On("LogsDir", mock.Anything).Return(mockLogsDir, nil)
			config := config.NewConfig(fsMock, osMock)
			config.SystemConfig = &systemConfigMock
			io = NewIOStreams(config, fsMock, osMock)
			io.Stderr = &stderrLogger

			err := io.FlushToLogFile(ctx, tt.prefix, tt.errStr)
			require.NoError(t, err)

			errorLogFilePath := filepath.Join(mockLogsDir, filename)
			errorLogFile, err := afero.ReadFile(io.fs, errorLogFilePath)
			require.NoError(t, err)
			for _, expected := range tt.expectedFile {
				assert.Contains(t, string(errorLogFile), expected)
			}
			var expectedStderr string
			if tt.suggestsFile {
				expectedStderr = fmt.Sprintf("Check %s for error logs\n", errorLogFilePath)
			}
			assert.Equal(t, expectedStderr, stderrBuff.String())
		})
	}
}
