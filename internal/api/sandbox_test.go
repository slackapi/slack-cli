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

func TestClient_CreateSandbox_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxCreateMethod,
		Response:       `{"ok":true,"team_id":"T123","user_id":"U123","url":"https://my-sandbox.slack.com"}`,
	})
	defer teardown()
	teamID, sandboxURL, err := c.CreateSandbox(ctx, "token", "My Sandbox", "my-sandbox", "secret", "", "", 0, "", 0, false)
	require.NoError(t, err)
	require.Equal(t, "T123", teamID)
	require.Equal(t, "https://my-sandbox.slack.com", sandboxURL)
}

func TestClient_CreateSandbox_WithOptionalParams(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxCreateMethod,
		Response:       `{"ok":true,"team_id":"T456","user_id":"U456","url":"https://other-sandbox.slack.com"}`,
	})
	defer teardown()
	teamID, sandboxURL, err := c.CreateSandbox(ctx, "token", "My Sandbox", "my-sandbox", "secret", "en-US", "O123", 1, "EVENT123", 1234567890, false)
	require.NoError(t, err)
	require.Equal(t, "T456", teamID)
	require.Equal(t, "https://other-sandbox.slack.com", sandboxURL)
}

func TestClient_CreateSandbox_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxCreateMethod, func(c *Client) error {
		_, _, err := c.CreateSandbox(ctx, "token", "name", "domain", "password", "", "", 0, "", 0, false)
		return err
	})
}

func TestClient_DeleteSandbox_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxDeleteMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	err := c.DeleteSandbox(ctx, "token", "T123")
	require.NoError(t, err)
}

func TestClient_DeleteSandbox_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxDeleteMethod, func(c *Client) error {
		return c.DeleteSandbox(ctx, "token", "T123")
	})
}

func TestClient_ListSandboxes_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true,"sandboxes":[{"sandbox_team_id":"T1","sandbox_name":"Sandbox 1","sandbox_domain":"sb1","status":"active","date_created":123,"date_archived":0},{"sandbox_team_id":"T2","sandbox_name":"Sandbox 2","sandbox_domain":"sb2","status":"active","date_created":456,"date_archived":0}]}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "")
	require.NoError(t, err)
	require.Len(t, sandboxes, 2)
	require.Equal(t, "T1", sandboxes[0].TeamID)
	require.Equal(t, "Sandbox 1", sandboxes[0].Name)
	require.Equal(t, "sb1", sandboxes[0].Domain)
	require.Equal(t, "active", sandboxes[0].Status)
	require.Equal(t, "T2", sandboxes[1].TeamID)
}

func TestClient_ListSandboxes_Empty(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true,"sandboxes":[]}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "")
	require.NoError(t, err)
	require.Empty(t, sandboxes)
}

func TestClient_ListSandboxes_NilSandboxesReturnsEmptySlice(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "")
	require.NoError(t, err)
	require.Equal(t, []types.Sandbox{}, sandboxes)
}

func TestClient_ListSandboxes_WithStatusFilter(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true,"sandboxes":[{"sandbox_team_id":"T1","sandbox_name":"Archived","sandbox_domain":"arch","status":"archived","date_created":100,"date_archived":200}]}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "archived")
	require.NoError(t, err)
	require.Len(t, sandboxes, 1)
	require.Equal(t, "archived", sandboxes[0].Status)
}

func TestClient_ListSandboxes_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxListMethod, func(c *Client) error {
		_, err := c.ListSandboxes(ctx, "token", "")
		return err
	})
}
