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

func TestClient_AddRemoveSetAccess(t *testing.T) {
	tests := []struct {
		name         string
		expectedPath string
		resultJSON   string
		testFunc     func(t *testing.T, c *Client) error
		want         string
		wantErr      bool
		errMessage   string
	}{
		{
			name:       "Add user success",
			resultJSON: `{"ok": true, "distribution_type": "named_entities", "user_ids": ["user1", "user2"]}`,
			testFunc: func(t *testing.T, c *Client) error {
				ctx := slackcontext.MockContext(t.Context())
				return c.FunctionDistributionAddUsers(ctx, "valid_function", "app", "user1,user2")
			},
		},
		{
			name:       "Add user: validation error",
			resultJSON: `{"ok": false, "error":"user_not_found"}`,
			testFunc: func(t *testing.T, c *Client) error {
				ctx := slackcontext.MockContext(t.Context())
				return c.FunctionDistributionAddUsers(ctx, "valid_function", "app", "user1,user2")
			},
			wantErr:    true,
			errMessage: "user_not_found",
		},
		{
			name:       "Remove user success",
			resultJSON: `{"ok": true, "distribution_type": "named_entities", "user_ids": []}`,
			testFunc: func(t *testing.T, c *Client) error {
				ctx := slackcontext.MockContext(t.Context())
				return c.FunctionDistributionRemoveUsers(ctx, "valid_function", "app", "user1,user2")
			},
		},
		{
			name:       "Remove user: distribution type not named_entitied",
			resultJSON: `{"ok":false,"error":"invalid_distribution_type"}`,
			testFunc: func(t *testing.T, c *Client) error {
				ctx := slackcontext.MockContext(t.Context())
				return c.FunctionDistributionRemoveUsers(ctx, "valid_function", "app", "user1,user2")
			},
			wantErr:    true,
			errMessage: "invalid_distribution_type",
		},
		{
			name:       "Set access type success",
			resultJSON: `{"ok": true, "distribution_type": "everyone", "user_ids": []}`,
			testFunc: func(t *testing.T, c *Client) error {
				ctx := slackcontext.MockContext(t.Context())
				_, err := c.FunctionDistributionSet(ctx, "valid_function", "app", types.EVERYONE, "")
				return err
			},
		},
		{
			name:       "Set access type: access type not recognized by backend",
			resultJSON: `{"ok":false,"error":"invalid_arguments"}`,
			testFunc: func(t *testing.T, c *Client) error {
				ctx := slackcontext.MockContext(t.Context())
				_, err := c.FunctionDistributionSet(ctx, "valid_function", "app", types.EVERYONE, "")
				return err
			},
			wantErr:    true,
			errMessage: "invalid_arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				result := tt.resultJSON
				_, err := fmt.Fprintln(w, result)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer ts.Close()
			c := NewClient(&http.Client{}, ts.URL, nil)

			// execute
			err := tt.testFunc(t, c)

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

func TestClient_FunctionDistributionList_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: functionDistributionsPermissionsListMethod,
		Response:       `{"ok":true,"distribution_type":"everyone","users":[{"user_id":"W123","username":"grace","email":"grace@gmail.com"}]}`,
	})
	defer teardown()
	_, _, err := c.FunctionDistributionList(ctx, "dummy_callback_id", "dummy_app_id")
	require.NoError(t, err)
}

func TestClient_FunctionDistributionList_HTTPRequestFailed(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: functionDistributionsPermissionsListMethod,
		StatusCode:     http.StatusInternalServerError,
	})
	defer teardown()
	_, _, err := c.FunctionDistributionList(ctx, "dummy_callback_id", "dummy_app_id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_request_failed")
}

func TestClient_FunctionDistributionList_HTTPResponseInvalid(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: functionDistributionsPermissionsListMethod,
		Response:       "}{",
	})
	defer teardown()
	_, _, err := c.FunctionDistributionList(ctx, "dummy_callback_id", "dummy_app_id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
}

func TestClient_FunctionDistributionList_NotOk(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: functionDistributionsPermissionsListMethod,
		Response:       `{"ok":false}`,
	})
	defer teardown()
	_, _, err := c.FunctionDistributionList(ctx, "dummy_callback_id", "dummy_app_id")
	require.Error(t, err)
	require.Contains(t, err.Error(), functionDistributionsPermissionsListMethod)
}

func TestClient_FunctionDistributionList_InvalidDistType(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: functionDistributionsPermissionsListMethod,
		Response:       `{"ok":true,"distribution_type":"banana"}`,
	})
	defer teardown()
	_, _, err := c.FunctionDistributionList(ctx, "dummy_callback_id", "dummy_app_id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unrecognized access type banana")
}
