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

func Test_Client_AddCollaborator(t *testing.T) {
	tests := map[string]struct {
		response  string
		user      types.SlackUser
		expectErr string
	}{
		"adds a collaborator by email": {
			response: `{"ok":true}`,
			user:     types.SlackUser{Email: "user@example.com", PermissionType: "owner"},
		},
		"adds a collaborator by user ID": {
			response: `{"ok":true}`,
			user:     types.SlackUser{ID: "U123", PermissionType: "owner"},
		},
		"returns error when user not found": {
			response:  `{"ok":false,"error":"user_not_found"}`,
			user:      types.SlackUser{Email: "bad@example.com"},
			expectErr: "user_not_found",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: collaboratorsAddMethod,
				Response:       tc.response,
			})
			defer teardown()
			err := c.AddCollaborator(ctx, "token", "A123", tc.user)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_Client_ListCollaborators(t *testing.T) {
	tests := map[string]struct {
		response      string
		expectedCount int
		expectedID    string
		expectErr     string
	}{
		"returns a list of collaborators": {
			response:      `{"ok":true,"owners":[{"user_id":"U123","username":"Test User"}]}`,
			expectedCount: 1,
			expectedID:    "U123",
		},
		"returns error when app not found": {
			response:  `{"ok":false,"error":"app_not_found"}`,
			expectErr: "app_not_found",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: collaboratorsListMethod,
				Response:       tc.response,
			})
			defer teardown()
			users, err := c.ListCollaborators(ctx, "token", "A123")
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
			} else {
				require.NoError(t, err)
				require.Len(t, users, tc.expectedCount)
				require.Equal(t, tc.expectedID, users[0].ID)
			}
		})
	}
}

func Test_Client_RemoveCollaborator(t *testing.T) {
	tests := map[string]struct {
		response  string
		user      types.SlackUser
		expectErr string
	}{
		"removes a collaborator": {
			response: `{"ok":true}`,
			user:     types.SlackUser{ID: "U123"},
		},
		"returns error when removing owner": {
			response:  `{"ok":false,"error":"cannot_remove_owner"}`,
			user:      types.SlackUser{Email: "owner@example.com"},
			expectErr: "cannot_remove_owner",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: collaboratorsRemoveMethod,
				Response:       tc.response,
			})
			defer teardown()
			warnings, err := c.RemoveCollaborator(ctx, "token", "A123", tc.user)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
			} else {
				require.NoError(t, err)
				require.Empty(t, warnings)
			}
		})
	}
}

func Test_Client_UpdateCollaborator(t *testing.T) {
	tests := map[string]struct {
		response  string
		user      types.SlackUser
		expectErr string
	}{
		"updates a collaborator": {
			response: `{"ok":true}`,
			user:     types.SlackUser{ID: "U123", PermissionType: "collaborator"},
		},
		"returns error for invalid permission": {
			response:  `{"ok":false,"error":"invalid_permission"}`,
			user:      types.SlackUser{ID: "U123", PermissionType: "invalid"},
			expectErr: "invalid_permission",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: collaboratorsUpdateMethod,
				Response:       tc.response,
			})
			defer teardown()
			err := c.UpdateCollaborator(ctx, "token", "A123", tc.user)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
