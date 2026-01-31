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

package tracking

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Tracking_cleanSessionData(t *testing.T) {
	tests := map[string]struct {
		input          EventData
		expectedOutput EventData
	}{
		"should redact error messages": {
			input: EventData{
				ErrorMessage: "you used this weird token=xoxb-1234-5678",
			},
			expectedOutput: EventData{
				ErrorMessage: "you used this weird token=...",
			},
		},
		"should remove ansi colour codes from error messages": {
			input: EventData{
				ErrorMessage: "This is an invalid Slack app directory\u001b[1;38;5;178m (invalid_app_directory)\u001b[0m\n\nSuggestion:\n\nA valid Slack app directory requires a Slack Configuration file (`slack.json`).\n",
			},
			expectedOutput: EventData{
				ErrorMessage: "This is an invalid Slack app directory (invalid_app_directory)\n\nSuggestion:\n\nA valid Slack app directory requires a Slack Configuration file (`slack.json`).\n",
			},
		},
		"should redact app template messages": {
			input: EventData{
				App: AppEventData{
					Template: "my-xapp-example",
				},
			},
			expectedOutput: EventData{
				App: AppEventData{
					Template: "my-...",
				},
			},
		},
	}

	for name, tc := range tests {
		et := NewEventTracker()
		t.Run(name, func(t *testing.T) {
			actual := et.cleanSessionData(tc.input)
			require.Equal(t, tc.expectedOutput, actual)
		})
	}
}

func Test_Tracking_FlushToLogstash(t *testing.T) {
	tests := map[string]struct {
		exitCode             iostreams.ExitCode
		assertOnRequest      func(t *testing.T, req *http.Request)
		shouldNotSendRequest bool
		setup                func(cfg *config.Config)
	}{
		"should always send an array to the event logging endpoint": {
			exitCode: iostreams.ExitOK,
			assertOnRequest: func(t *testing.T, req *http.Request) {
				payload, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				payloadStr := string(payload)
				require.Equal(t, payloadStr[0:1], "[")
				require.Equal(t, payloadStr[len(payloadStr)-1:], "]")
			},
		},
		"should set event name to 'success' if exit code is OK": {
			exitCode: iostreams.ExitOK,
			assertOnRequest: func(t *testing.T, req *http.Request) {
				payload, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				require.Contains(t, string(payload), fmt.Sprintf("\"event\":\"%s\"", "success"))
			},
		},
		"should set event name to 'interrupt' if exit code is Cancel": {
			exitCode: iostreams.ExitCancel,
			assertOnRequest: func(t *testing.T, req *http.Request) {
				payload, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				require.Contains(t, string(payload), fmt.Sprintf("\"event\":\"%s\"", "interrupt"))
			},
		},
		"should set event name to 'error' if exit code is Error": {
			exitCode: iostreams.ExitError,
			assertOnRequest: func(t *testing.T, req *http.Request) {
				payload, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				require.Contains(t, string(payload), fmt.Sprintf("\"event\":\"%s\"", "error"))
			},
		},
		"should not send an event tracking request if Do Not Track configuration is set to true": {
			setup: func(cfg *config.Config) {
				cfg.DisableTelemetryFlag = true
			},
			shouldNotSendRequest: true,
		},
		"should not send an event tracking request if logstash host has not been resolved": {
			setup: func(cfg *config.Config) {
				cfg.LogstashHostResolved = ""
			},
			shouldNotSendRequest: true,
		},
		"should strip 'v' prefix from version string": {
			setup: func(cfg *config.Config) {
				cfg.Version = "v4.2.0"
			},
			assertOnRequest: func(t *testing.T, req *http.Request) {
				payload, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				require.Contains(t, string(payload), fmt.Sprintf("\"cli_version\":\"%s\"", "4.2.0"))
			},
		},
		"should include os and arch build information": {
			assertOnRequest: func(t *testing.T, req *http.Request) {
				payload, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				require.Contains(t, string(payload), fmt.Sprintf("\"arch\":\"%s\"", runtime.GOARCH))
				require.Contains(t, string(payload), fmt.Sprintf("\"os\":\"%s\"", runtime.GOOS))
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			et := NewEventTracker()
			var requestSent = false
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				requestSent = true
				res.WriteHeader(200)
				_, _ = res.Write([]byte("{\"ok\":true}"))
				if tc.assertOnRequest != nil {
					tc.assertOnRequest(t, req)
				}
			}))
			defer testServer.Close()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			fs := slackdeps.NewFsMock()
			cfg := config.NewConfig(fs, os)
			cfg.LogstashHostResolved = testServer.URL
			if tc.setup != nil {
				tc.setup(cfg)
			}
			ioMock := iostreams.NewIOStreamsMock(cfg, fs, os)
			ioMock.AddDefaultMocks()
			err := et.FlushToLogstash(ctx, cfg, ioMock, tc.exitCode)
			require.NoError(t, err)
			if tc.shouldNotSendRequest && requestSent {
				require.Fail(t, "Expected no event tracking request to be sent, but request was sent")
			}
		})
	}
}

func Test_Tracking_SetSessionData(t *testing.T) {
	tests := map[string]struct {
		setterFunc func(e *EventTracker, value string)
		getterFunc func(e *EventTracker) string
		value      string
	}{
		"error code can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetErrorCode(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().ErrorCode
			},
			value: "example_code",
		},
		"error message can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetErrorMessage(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().ErrorMessage
			},
			value: "example error message",
		},
		"auth enterprise id can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAuthEnterpriseID(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().Auth.EnterpriseID
			},
			value: "E1111111111",
		},
		"auth team id can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAuthTeamID(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().Auth.TeamID
			},
			value: "T1111111111",
		},
		"auth user id can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAuthUserID(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().Auth.UserID
			},
			value: "U1111111111",
		},
		"app enterprise id can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAppEnterpriseID(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().App.EnterpriseID
			},
			value: "E1234567890",
		},
		"app team id can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAppTeamID(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().App.TeamID
			},
			value: "T1234567890",
		},
		"app user id can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAppUserID(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().App.UserID
			},
			value: "U1234567890",
		},
		"app template can be retrieved": {
			setterFunc: func(e *EventTracker, value string) {
				e.SetAppTemplate(value)
			},
			getterFunc: func(e *EventTracker) string {
				return e.getSessionData().App.Template
			},
			value: "slack-samples/deno-hello-world",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			et := NewEventTracker()
			tc.setterFunc(et, tc.value)
			actual := tc.getterFunc(et)
			assert.Equal(t, tc.value, actual)
		})
	}
}
