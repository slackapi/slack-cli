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
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_WorkflowsTriggerCreate(t *testing.T) {
	var inputs = make(map[string]*Input)
	inputs["test"] = &Input{Value: "val"}

	tests := map[string]struct {
		httpResponseJSON string
		expectedJSON     string
		inputTrigger     TriggerRequest
		wantErr          bool
		errMessage       string
	}{
		"Valid shortcut": {
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeShortcut,
				Workflow:      "#/workflows/test",
				WorkflowAppID: "A1234",
				Name:          "name",
				Description:   "desc",
				Shortcut:      &Shortcut{},
			},
			expectedJSON:     `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234"}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		"Valid shortcut, with inputs": {
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeShortcut,
				Workflow:      "#/workflows/test",
				WorkflowAppID: "A1234",
				Name:          "name",
				Description:   "desc",
				Shortcut:      &Shortcut{},
				Inputs:        inputs,
			},
			expectedJSON:     `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234","inputs":{"test":{"value":"val"}}}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		"Valid event": {
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeEvent,
				Workflow:      "#/workflows/test",
				WorkflowAppID: "A1234",
				Name:          "name",
				Description:   "desc",
				Event:         types.ToRawJSON(`{"event_type":"reaction_added","channel_ids":["C1234"]}`),
			},
			expectedJSON:     `{"type":"event","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","event":{"event_type":"reaction_added","channel_ids":["C1234"]}}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"event", "name":"name", "desc":"desc"}}`,
		},
		"Valid schedule": {
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeScheduled,
				Workflow:      "#/workflows/test",
				WorkflowAppID: "A1234",
				Name:          "name",
				Description:   "desc",
				Schedule:      types.ToRawJSON(`{"start_time":"2020-03-15","frequency":{"type":"daily"}}`),
			},
			expectedJSON:     `{"type":"scheduled","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","schedule":{"start_time":"2020-03-15","frequency":{"type":"daily"}}}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"scheduled", "name":"name", "desc":"desc"}}`,
		},
		"Valid webhook": {
			inputTrigger: TriggerRequest{
				Type:          types.TriggerTypeWebhook,
				Workflow:      "#/workflows/test",
				WorkflowAppID: "A1234",
				Name:          "name",
				Description:   "desc",
				WebHook:       types.ToRawJSON(`{"filter":{"root":{},"version":1},"channel_ids":["C1234"]}`),
			},
			expectedJSON:     `{"type":"webhook","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","webhook":{"filter":{"root":{},"version":1},"channel_ids":["C1234"]}}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"webhook", "name":"name", "desc":"desc"}}`,
		},
		"Propagates errors": {
			httpResponseJSON: `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:          true,
			errMessage:       "invalid_scopes",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersCreateMethod,
				ExpectedRequest: tc.expectedJSON,
				Response:        tc.httpResponseJSON,
			})
			defer teardown()

			// execute
			_, err := c.WorkflowsTriggersCreate(ctx, "token", tc.inputTrigger)

			// check
			if (err != nil) != tc.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", name, err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_WorkflowsTriggerUpdate(t *testing.T) {
	var inputs = make(map[string]*Input)
	inputs["test"] = &Input{Value: "val"}

	tests := map[string]struct {
		httpResponseJSON string
		expectedJSON     string
		input            TriggerUpdateRequest
		wantErr          bool
		errMessage       string
	}{
		"Valid shortcut": {
			input: TriggerUpdateRequest{
				TriggerID: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Workflow:      "#/workflows/test",
					WorkflowAppID: "A1234",
					Name:          "name",
					Description:   "desc",
					Shortcut:      &Shortcut{},
				},
			},
			expectedJSON:     `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234","trigger_id":"Ft123"}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		"Valid shortcut, with inputs": {
			input: TriggerUpdateRequest{
				TriggerID: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeShortcut,
					Workflow:      "#/workflows/test",
					WorkflowAppID: "A1234",
					Name:          "name",
					Description:   "desc",
					Shortcut:      &Shortcut{},
					Inputs:        inputs,
				},
			},
			expectedJSON:     `{"type":"shortcut","name":"name","description":"desc","shortcut":{},"workflow":"#/workflows/test","workflow_app_id":"A1234","inputs":{"test":{"value":"val"}},"trigger_id":"Ft123"}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"shortcut", "name":"name", "desc":"desc"}}`,
		},
		"Valid event": {
			input: TriggerUpdateRequest{
				TriggerID: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeEvent,
					Workflow:      "#/workflows/test",
					WorkflowAppID: "A1234",
					Name:          "name",
					Description:   "desc",
					Event:         types.ToRawJSON(`{"event_type":"reaction_added","channel_ids":["C1234"]}`),
				},
			},
			expectedJSON:     `{"type":"event","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","event":{"event_type":"reaction_added","channel_ids":["C1234"]},"trigger_id":"Ft123"}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"event", "name":"name", "desc":"desc"}}`,
		},
		"Valid schedule": {
			input: TriggerUpdateRequest{
				TriggerID: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeScheduled,
					Workflow:      "#/workflows/test",
					WorkflowAppID: "A1234",
					Name:          "name",
					Description:   "desc",
					Schedule:      types.ToRawJSON(`{"start_time":"2020-03-15","frequency":{"type":"daily"}}`),
				},
			},
			expectedJSON:     `{"type":"scheduled","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","schedule":{"start_time":"2020-03-15","frequency":{"type":"daily"}},"trigger_id":"Ft123"}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"scheduled", "name":"name", "desc":"desc"}}`,
		},
		"Valid webhook": {
			input: TriggerUpdateRequest{
				TriggerID: "Ft123",
				TriggerRequest: TriggerRequest{
					Type:          types.TriggerTypeWebhook,
					Workflow:      "#/workflows/test",
					WorkflowAppID: "A1234",
					Name:          "name",
					Description:   "desc",
					WebHook:       types.ToRawJSON(`{"filter":{"root":{},"version":1},"channel_ids":["C1234"]}`),
				},
			},
			expectedJSON:     `{"type":"webhook","name":"name","description":"desc","workflow":"#/workflows/test","workflow_app_id":"A1234","webhook":{"filter":{"root":{},"version":1},"channel_ids":["C1234"]},"trigger_id":"Ft123"}`,
			httpResponseJSON: `{"ok": true, "trigger": {"id":"Ft123", "type":"webhook", "name":"name", "desc":"desc"}}`,
		},
		"Propagates errors": {
			httpResponseJSON: `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:          true,
			errMessage:       "invalid_scopes",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersUpdateMethod,
				ExpectedRequest: tc.expectedJSON,
				Response:        tc.httpResponseJSON,
			})
			defer teardown()

			// execute
			_, err := c.WorkflowsTriggersUpdate(ctx, "token", tc.input)

			// check
			if (err != nil) != tc.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", name, err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func TestClient_WorkflowsTriggerDelete(t *testing.T) {
	tests := map[string]struct {
		response   string
		wantErr    bool
		errMessage string
	}{
		"OK result": {
			response: `{"ok":true}`,
		},
		"Error result": {
			response:   `{"ok":false,"error":"invalid_scopes"}`,
			wantErr:    true,
			errMessage: "invalid_scopes",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersDeleteMethod,
				ExpectedRequest: `{"trigger_id":"FtABC"}`,
				Response:        tc.response,
			})
			defer teardown()

			// execute
			err := c.WorkflowsTriggersDelete(ctx, "token", "FtABC")

			// check
			if (err != nil) != tc.wantErr {
				t.Errorf("%s test error = %v, wantErr %v", name, err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"test error contains invalid message",
				)
			}
		})
	}
}

func Test_API_WorkflowTriggersList(t *testing.T) {
	tests := map[string]struct {
		argsToken             string
		argsAppID             string
		argsLimit             int
		argsCursor            string
		argsType              string
		httpResponseJSON      string
		expectedErrorContains string
		expected              []types.DeployedTrigger
	}{
		"Successful request": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "123",  "type": "shortcut", "workflow": { "id": "456"}}]}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "123", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		"Response contains an error": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `{"ok":false,"error":"internal_error"}`,
			expectedErrorContains: "internal_error",
		},
		"Response contains invalid JSON": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHTTPResponseInvalid.Code,
		},
		"Successful request with limit and cursor": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsLimit:             4,
			argsCursor:            "",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "1",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "2",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "3",  "type": "shortcut", "workflow": { "id": "456"}}], "response_metadata": { "next_cursor": "fake_cursor" }}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "1", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "2", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "3", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		"Successful request with type all": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsLimit:             4,
			argsCursor:            "",
			argsType:              "all",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "1",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "2",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "3",  "type": "shortcut", "workflow": { "id": "456"}}], "response_metadata": { "next_cursor": "fake_cursor" }}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "1", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "2", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "3", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		"Successful request with type shortcut": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsType:              "shortcut",
			httpResponseJSON:      `{"ok": true,  "triggers": [ { "id": "1",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "2",  "type": "shortcut", "workflow": { "id": "456"}}, { "id": "3",  "type": "shortcut", "workflow": { "id": "456"}}], "response_metadata": { "next_cursor": "fake_cursor" }}`,
			expectedErrorContains: "",
			expected:              []types.DeployedTrigger{{ID: "1", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "2", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}, {ID: "3", Type: "shortcut", Workflow: types.TriggerWorkflow{ID: "456"}}},
		},
		"Invalid trigger type": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsLimit:             4,
			argsCursor:            "",
			argsType:              "fake_type",
			httpResponseJSON:      `{"ok":false,"error":"invalid_arguments"}`,
			expectedErrorContains: "invalid_arguments",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: workflowsTriggersListMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			args := TriggerListRequest{
				AppID:  tc.argsAppID,
				Limit:  tc.argsLimit,
				Cursor: tc.argsCursor,
			}
			actual, _, err := c.WorkflowsTriggersList(ctx, tc.argsToken, args)

			// Assertions
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			} else {
				require.Contains(t, err.Error(), tc.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
