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
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_ValidateSession_AuthRespOk(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		Response:       `{"ok":true,"url":"https://www.example.com/","team":"Example Workspace", "user":"grace","team_id":"T123","user_id":"W123","enterprise_id":"E123","is_enterprise_install":true}`,
	})
	defer teardown()
	resp, err := c.ValidateSession(context.Background(), "token")
	require.NoError(t, err)
	require.Equal(t, *resp.UserName, "grace")
	require.Equal(t, *resp.UserID, "W123")
	require.Equal(t, *resp.TeamID, "T123")
	require.Equal(t, *resp.URL, "https://www.example.com/")
	require.Equal(t, *resp.TeamName, "Example Workspace")
	require.Equal(t, *resp.EnterpriseID, "E123")
	require.Equal(t, *resp.IsEnterpriseInstall, true)
}

func TestClient_ValidateSession_AuthRespOk_MissingUserID(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		Response:       `{"ok":true,"url":"https://www.example.com/","team":"Example Workspace", "user":"grace","team_id":"T123"}`,
	})
	defer teardown()
	_, err := c.ValidateSession(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
	require.Contains(t, err.Error(), "invalid user_id")
}

func TestClient_ValidateSession_AuthRespNotOk(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		Response:       `{"ok":false}`,
	})
	defer teardown()
	_, err := c.ValidateSession(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), sessionValidateMethod)
	require.Contains(t, err.Error(), "unknown_error")
}

func TestClient_ValidateSession_AuthRespNotOk_NotAuthed(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		Response:       `{"ok":false,"error":"not_authed"}`,
	})
	defer teardown()
	_, err := c.ValidateSession(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not_authed")
	require.Contains(t, err.Error(), "You are either not logged in or your login session has expired")
}

func TestClient_ValidateSession_AuthRespNotOk_InvalidAuth(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		Response:       `{"ok":false,"error":"invalid_auth"}`,
	})
	defer teardown()
	_, err := c.ValidateSession(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid_auth")
	require.Contains(t, err.Error(), "Your user account authorization isn't valid")
}

func TestClient_ValidateSession_HTTPRequestFailed(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		StatusCode:     http.StatusInternalServerError,
	})
	defer teardown()
	_, err := c.ValidateSession(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_request_failed")
}

func TestClient_RevokeToken_ErrNotLoggedIn(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: revokeTokenMethod,
		StatusCode:     http.StatusInternalServerError,
	})
	defer teardown()
	err := c.RevokeToken(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_request_failed")
}

func TestClient_ValidateSession_HTTPResponseInvalid(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sessionValidateMethod,
		Response:       `badresponse`,
	})
	defer teardown()
	_, err := c.ValidateSession(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
}

func TestClient_RevokeToken_Ok(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: revokeTokenMethod,
		Response:       `{"ok":true,"revoked":true}`,
	})
	defer teardown()
	err := c.RevokeToken(context.Background(), "token")
	require.NoError(t, err)
}

func TestClient_RevokeToken_NotOk(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: revokeTokenMethod,
		Response:       `{"ok":false}`,
	})
	defer teardown()
	err := c.RevokeToken(context.Background(), "token")
	require.Error(t, err)
}

func TestClient_RevokeToken_NotOk_MappedErrorMsgs(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: revokeTokenMethod,
		Response:       `{"ok":false,"error":"token_expired"}`,
	})
	defer teardown()
	err := c.RevokeToken(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "token_expired")
}

func TestClient_RevokeToken_JsonUnmarshalFail(t *testing.T) {
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: revokeTokenMethod,
		Response:       `}{`,
	})
	defer teardown()
	err := c.RevokeToken(context.Background(), "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "http_response_invalid")
}
