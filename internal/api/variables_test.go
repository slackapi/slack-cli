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

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_AddVariable_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  varAddMethod,
		ExpectedRequest: `{"app_id":"A123","variables":[{"name":"dummy_var","value":"dummy_val"}]}`,
		Response:        `{"ok":true}`,
	})
	defer teardown()
	err := c.AddVariable(ctx, "token", "A123", "dummy_var", "dummy_val")
	require.NoError(t, err)
}

func TestClient_AddVariable_NotOk(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  varAddMethod,
		ExpectedRequest: `{"app_id":"A123","variables":[{"name":"dummy_var","value":"dummy_val"}]}`,
		Response:        `{"ok":false}`,
	})
	defer teardown()
	err := c.AddVariable(ctx, "token", "A123", "dummy_var", "dummy_val")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown_error")
}

func TestClient_AddVariable_HTTPResponseInvalid(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: varAddMethod,
	})
	defer teardown()
	err := c.AddVariable(ctx, "token", "", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
}

func TestClient_AddVariable_HTTPRequestFailed(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: varAddMethod,
		StatusCode:     http.StatusInternalServerError,
	})
	defer teardown()
	err := c.AddVariable(ctx, "token", "", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_request_failed")
}

func TestClient_ListVariables_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  varListMethod,
		ExpectedRequest: `{"app_id":"A123"}`,
		Response:        `{"ok":true,"variable_names":["var1"]}`,
	})
	defer teardown()
	VariableNames, err := c.ListVariables(ctx, "token", "A123")
	require.NoError(t, err)
	require.Equal(t, VariableNames[0], "var1")
}

func TestClient_ListVariables_NotOk(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  varListMethod,
		ExpectedRequest: `{"app_id":"A123"}`,
		Response:        `{"ok":false}`,
	})
	defer teardown()
	_, err := c.ListVariables(ctx, "token", "A123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown_error")
}

func TestClient_ListVariables_HTTPResponseInvalid(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: varListMethod,
	})
	defer teardown()
	_, err := c.ListVariables(ctx, "token", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
}

func TestClient_ListVariables_HTTPRequestFailed(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: varListMethod,
		StatusCode:     http.StatusInternalServerError,
	})
	defer teardown()
	_, err := c.ListVariables(ctx, "token", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_request_failed")
}

func TestClient_RemoveVariable_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  varRemoveMethod,
		ExpectedRequest: `{"app_id":"A123","variable_names":["dummy_var_name"]}`,
		Response:        `{"ok":true}`,
	})
	defer teardown()
	err := c.RemoveVariable(ctx, "token", "A123", "dummy_var_name")
	require.NoError(t, err)
}

func TestClient_RemoveVariable_NotOk(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  varRemoveMethod,
		ExpectedRequest: `{"app_id":"A123","variable_names":["dummy_var_name"]}`,
		Response:        `{"ok":false}`,
	})
	defer teardown()
	err := c.RemoveVariable(ctx, "token", "A123", "dummy_var_name")
	require.Error(t, err)
}

func TestClient_RemoveVariable_HTTPResponseInvalid(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: varRemoveMethod,
	})
	defer teardown()
	err := c.RemoveVariable(ctx, "token", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
}

func TestClient_RemoveVariable_HTTPRequestFailed(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: varRemoveMethod,
		StatusCode:     http.StatusInternalServerError,
	})
	defer teardown()
	err := c.RemoveVariable(ctx, "token", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_request_failed")
}
