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
	"net/http"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_FunctionDistributionAddUsers(t *testing.T) {
	tests := map[string]struct {
		httpResponseJSON string
		wantErr          bool
		errMessage       string
	}{
		"Add user success": {
			httpResponseJSON: `{"ok": true, "distribution_type": "named_entities", "user_ids": ["user1", "user2"]}`,
		},
		"Add user: validation error": {
			httpResponseJSON: `{"ok": false, "error":"user_not_found"}`,
			wantErr:          true,
			errMessage:       "user_not_found",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: functionDistributionsPermissionsAddMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			err := c.FunctionDistributionAddUsers(ctx, "valid_function", "app", "user1,user2")

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

func TestClient_FunctionDistributionRemoveUsers(t *testing.T) {
	tests := map[string]struct {
		httpResponseJSON string
		wantErr          bool
		errMessage       string
	}{
		"Remove user success": {
			httpResponseJSON: `{"ok": true, "distribution_type": "named_entities", "user_ids": []}`,
		},
		"Remove user: distribution type not named_entities": {
			httpResponseJSON: `{"ok":false,"error":"invalid_distribution_type"}`,
			wantErr:          true,
			errMessage:       "invalid_distribution_type",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: functionDistributionsPermissionsRemoveMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			err := c.FunctionDistributionRemoveUsers(ctx, "valid_function", "app", "user1,user2")

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

func TestClient_FunctionDistributionSet(t *testing.T) {
	tests := map[string]struct {
		httpResponseJSON string
		wantErr          bool
		errMessage       string
	}{
		"Set access type success": {
			httpResponseJSON: `{"ok": true, "distribution_type": "everyone", "user_ids": []}`,
		},
		"Set access type: access type not recognized by backend": {
			httpResponseJSON: `{"ok":false,"error":"invalid_arguments"}`,
			wantErr:          true,
			errMessage:       "invalid_arguments",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: functionDistributionsPermissionsSetMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			_, err := c.FunctionDistributionSet(ctx, "valid_function", "app", types.PermissionEveryone, "")

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
