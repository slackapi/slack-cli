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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_WorkflowsTriggerCreate(t *testing.T) {
	var inputs = make(map[string]*Input)
	inputs["test"] = &Input{Value: "val"}

	tests := []struct {
		name         string
		resultJson   string
		expectedJson string
		inputTrigger TriggerRequest
		wantErr      bool
		errMessage   string
	}{
		{
			name: "Valid shortcut",
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeShortcut,
				Workflow:      "#/workflows/test",
				WorkflowAppId: "A1234",
				Name:          "name",
				Description:   "desc",
				Shortcut:      &Shortcut{},
			},
			expectedJson: `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234"}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid shortcut, with inputs",
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeShortcut,
				Workflow:      "#/workflows/test",
				WorkflowAppId: "A1234",
				Name:          "name",
				Description:   "desc",
				Shortcut:      &Shortcut{},
				Inputs:        inputs,
			},
			expectedJson: `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234","inputs":{"test":{"value":"val"}}}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid event",
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeEvent,
				Workflow:      "#/workflows/test",
				WorkflowAppId: "A1234",
				Name:          "name",
				Description:   "desc",
				Event:         types.ToRawJson(`{"event_type":"reaction_added","channel_ids":["C1234"]}`),
			},
			expectedJson: `{"type":"event","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","event":{"event_type":"reaction_added","channel_ids":["C1234"]}}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"event", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid schedule",
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeScheduled,
				Workflow:      "#/workflows/test",
				WorkflowAppId: "A1234",
				Name:          "name",
				Description:   "desc",
				Schedule:      types.ToRawJson(`{"start_time":"2020-03-15","frequency":{"type":"daily"}}`),
			},
			expectedJson: `{"type":"scheduled","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","schedule":{"start_time":"2020-03-15","frequency":{"type":"daily"}}}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"scheduled", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid webhook",
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeWebhook,
				Workflow:      "#/workflows/test",
				WorkflowAppId: "A1234",
				Name:          "name",
				Description:   "desc",
				WebHook:       types.ToRawJson(`{"filter":{"root":{},"version":1},"channel_ids":["C1234"]}`),
			},
			expectedJson: `{"type":"webhook","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","webhook":{"filter":{"root":{},"version":1},"channel_ids":["C1234"]}}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"webhook", "name":"name", "desc":"desc"}}`,
		},
		{
			name:       "Propagates errors",
			resultJson: `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:    true,
			errMessage: "invalid_scopes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, workflowsTriggersCreateMethod)
				if tt.expectedJson != "" {
					payload, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.Equal(t, tt.expectedJson, string(payload))
				}
				result := tt.resultJson
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			// execute
			_, err := c.WorkflowsTriggersCreate(ctx, "token", tt.inputTrigger)

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_WorkflowsTriggerUpdate(t *testing.T) {
	var inputs = make(map[string]*Input)
	inputs["test"] = &Input{Value: "val"}

	tests := []struct {
		name         string
		resultJson   string
		expectedJson string
		input        TriggerUpdateRequest
		wantErr      bool
		errMessage   string
	}{
		{
			name: "Valid shortcut",
			input: TriggerUpdateRequest{
				TriggerId: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Workflow:      "#/workflows/test",
					WorkflowAppId: "A1234",
					Name:          "name",
					Description:   "desc",
					Shortcut:      &Shortcut{},
				},
			},
			expectedJson: `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234","trigger_id":"Ft123"}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid shortcut, with inputs",
			input: TriggerUpdateRequest{
				TriggerId: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Workflow:      "#/workflows/test",
					WorkflowAppId: "A1234",
					Name:          "name",
					Description:   "desc",
					Shortcut:      &Shortcut{},
					Inputs:        inputs,
				},
			},
			expectedJson: `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234","inputs":{"test":{"value":"val"}},"trigger_id":"Ft123"}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid event",
			input: TriggerUpdateRequest{
				TriggerId: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeEvent,
					Workflow:      "#/workflows/test",
					WorkflowAppId: "A1234",
					Name:          "name",
					Description:   "desc",
					Event:         types.ToRawJson(`{"event_type":"reaction_added","channel_ids":["C1234"]}`),
				},
			},
			expectedJson: `{"type":"event","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","event":{"event_type":"reaction_added","channel_ids":["C1234"]},"trigger_id":"Ft123"}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"event", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid schedule",
			input: TriggerUpdateRequest{
				TriggerId: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeScheduled,
					Workflow:      "#/workflows/test",
					WorkflowAppId: "A1234",
					Name:          "name",
					Description:   "desc",
					Schedule:      types.ToRawJson(`{"start_time":"2020-03-15","frequency":{"type":"daily"}}`),
				},
			},
			expectedJson: `{"type":"scheduled","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","schedule":{"start_time":"2020-03-15","frequency":{"type":"daily"}},"trigger_id":"Ft123"}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"scheduled", "name":"name", "desc":"desc"}}`,
		},
		{
			name: "Valid webhook",
			input: TriggerUpdateRequest{
				TriggerId: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeWebhook,
					Workflow:      "#/workflows/test",
					WorkflowAppId: "A1234",
					Name:          "name",
					Description:   "desc",
					WebHook:       types.ToRawJson(`{"filter":{"root":{},"version":1},"channel_ids":["C1234"]}`),
				},
			},
			expectedJson: `{"type":"webhook","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","webhook":{"filter":{"root":{},"version":1},"channel_ids":["C1234"]},"trigger_id":"Ft123"}`,
			resultJson:   `{"ok": true, "trigger": {"id":"Ft123", "type":"webhook", "name":"name", "desc":"desc"}}`,
		},
		{
			name:       "Propagates errors",
			resultJson: `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:    true,
			errMessage: "invalid_scopes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, workflowsTriggersUpdateMethod)
				if tt.expectedJson != "" {
					payload, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.Equal(t, tt.expectedJson, string(payload))
				}
				result := tt.resultJson
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			// execute
			_, err := c.WorkflowsTriggersUpdate(ctx, "token", tt.input)

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_WorkflowsTriggerDelete(t *testing.T) {
	tests := []struct {
		name       string
		resultJson string
		wantErr    bool
		errMessage string
	}{
		{
			name:       "OK result",
			resultJson: `{"ok":true}`,
		},
		{
			name:       "Error result",
			resultJson: `{"ok":false,"error":"invalid_scopes"}`,
			wantErr:    true,
			errMessage: "invalid_scopes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.URL.Path, workflowsTriggersDeleteMethod)
				expectedJson := `{"trigger_id":"FtABC"}`
				payload, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, expectedJson, string(payload))
				result := tt.resultJson
				_, err = fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			// execute
			err := c.WorkflowsTriggersDelete(ctx, "token", "FtABC")

			// check
			if (err != nil) != tt.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func Test_API_WorkflowTriggersList(t *testing.T) {
	tests := []struct {
		name                  string
		argsToken             string
		argsAppID             string
		argsLimit             int
		argsCursor            string
		argsType              string
		httpResponseJSON      string
		expectedErrorContains string
		expected              []types.DeployedTrigger
	}{
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "123",  "type": "shortcut", "workflow": { "id": "456"}}]}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "123", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		{
			name:                  "Response contains an error",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `{"ok":false,"error":"internal_error"}`,
			expectedErrorContains: "internal_error",
		},
		{
			name:                  "Response contains invalid JSON",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHttpResponseInvalid.Code,
		},
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsLimit:             4,
			argsCursor:            "",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "1",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "2",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "3",  "type": "shortcut", "workflow": { "id": "456"}}], "response_metadata": { "next_cursor": "fake_cursor" }}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "1", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "2", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "3", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsLimit:             4,
			argsCursor:            "",
			argsType:              "all",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "1",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "2",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "3",  "type": "shortcut", "workflow": { "id": "456"}}], "response_metadata": { "next_cursor": "fake_cursor" }}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "1", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "2", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "3", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsType:              "shortcut",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "1",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "2",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "3",  "type": "shortcut", "workflow": { "id": "456"}}], "response_metadata": { "next_cursor": "fake_cursor" }}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "1", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "2", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "3", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsLimit:             4,
			argsCursor:            "",
			argsType:              "fake_type",
			httpResponseJSON:      `{"ok":false,"error":"invalid_arguments"}`,
			expectedErrorContains: "invalid_arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: workflowsTriggersListMethod,
				Response:       tt.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			args := TriggerListRequest{
				AppId:  tt.argsAppID,
				Limit:  tt.argsLimit,
				Cursor: tt.argsCursor,
			}
			actual, _, err := c.WorkflowsTriggersList(ctx, tt.argsToken, args)

			// Assertions
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
