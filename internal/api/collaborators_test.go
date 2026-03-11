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

func TestClient_AddCollaborator_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsAddMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	err := c.AddCollaborator(ctx, "token", "A123", types.SlackUser{Email: "user@example.com", PermissionType: "owner"})
	require.NoError(t, err)
}

func TestClient_AddCollaborator_WithUserID(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsAddMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	err := c.AddCollaborator(ctx, "token", "A123", types.SlackUser{ID: "U123", PermissionType: "owner"})
	require.NoError(t, err)
}

func TestClient_AddCollaborator_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsAddMethod,
		Response:       `{"ok":false,"error":"user_not_found"}`,
	})
	defer teardown()
	err := c.AddCollaborator(ctx, "token", "A123", types.SlackUser{Email: "bad@example.com"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "user_not_found")
}

func TestClient_ListCollaborators_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsListMethod,
		Response:       `{"ok":true,"owners":[{"user_id":"U123","username":"Test User"}]}`,
	})
	defer teardown()
	users, err := c.ListCollaborators(ctx, "token", "A123")
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, "U123", users[0].ID)
}

func TestClient_ListCollaborators_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsListMethod,
		Response:       `{"ok":false,"error":"app_not_found"}`,
	})
	defer teardown()
	_, err := c.ListCollaborators(ctx, "token", "A123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "app_not_found")
}

func TestClient_RemoveCollaborator_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsRemoveMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	warnings, err := c.RemoveCollaborator(ctx, "token", "A123", types.SlackUser{ID: "U123"})
	require.NoError(t, err)
	require.Empty(t, warnings)
}

func TestClient_RemoveCollaborator_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsRemoveMethod,
		Response:       `{"ok":false,"error":"cannot_remove_owner"}`,
	})
	defer teardown()
	_, err := c.RemoveCollaborator(ctx, "token", "A123", types.SlackUser{Email: "owner@example.com"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot_remove_owner")
}

func TestClient_UpdateCollaborator_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsUpdateMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	err := c.UpdateCollaborator(ctx, "token", "A123", types.SlackUser{ID: "U123", PermissionType: "collaborator"})
	require.NoError(t, err)
}

func TestClient_UpdateCollaborator_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: collaboratorsUpdateMethod,
		Response:       `{"ok":false,"error":"invalid_permission"}`,
	})
	defer teardown()
	err := c.UpdateCollaborator(ctx, "token", "A123", types.SlackUser{ID: "U123", PermissionType: "invalid"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid_permission")
}
