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
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_TriggerPermissionsSet(t *testing.T) {
	var fakeTriggerID = "Ft123"
	var fakeToken = "xoxp-123"

	tests := []struct {
		name            string
		resultJSON      string
		expectedRequest string
		permissionType  types.Permission
		users           string
		channels        string
		workspaces      string
		organizations   string
		wantErr         bool
		errMessage      string
	}{
		{
			name:            "Set to everyone",
			permissionType:  types.EVERYONE,
			users:           "",
			expectedRequest: `permission_type=everyone&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true, "permission_type": "everyone"}`,
		},
		{
			name:            "Set to collaborators",
			permissionType:  types.APP_COLLABORATORS,
			users:           "U0001",
			expectedRequest: `permission_type=app_collaborators&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true, "permission_type": "app_collaborators", "user_ids": [ "U0001" ]}`,
		},
		{
			name:            "Set to named_entities (users)",
			permissionType:  types.NAMED_ENTITIES,
			users:           "U0001,U0002",
			expectedRequest: `permission_type=named_entities&token=xoxp-123&trigger_id=Ft123&user_ids=U0001%2CU0002`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0001", "U0002" ]}`,
		},
		{
			name:            "Set to named_entities (channels)",
			permissionType:  types.NAMED_ENTITIES,
			channels:        "C0001,C0002",
			expectedRequest: `channel_ids=C0001%2CC0002&permission_type=named_entities&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0001", "C0002" ]}`,
		},
		{
			name:            "Set to named_entities (workspaces)",
			permissionType:  types.NAMED_ENTITIES,
			workspaces:      "T0001,T0002",
			expectedRequest: `permission_type=named_entities&team_ids=T0001%2CT0002&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "teams_ids": [ "T0001", "T0002" ]}`,
		},
		{
			name:            "Set to named_entities (organizations)",
			permissionType:  types.NAMED_ENTITIES,
			organizations:   "E0001,E0002",
			expectedRequest: `org_ids=E0001%2CE0002&permission_type=named_entities&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0001", "E0002" ]}`,
		},
		{
			name:       "Propagates errors",
			resultJSON: `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:    true,
			errMessage: "invalid_scopes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsSetMethod,
				ExpectedRequest: tt.expectedRequest,
				Response:        tt.resultJSON,
			})
			defer teardown()

			// execute
			if tt.users != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tt.users, tt.permissionType, "users")

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
			} else if tt.channels != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tt.channels, tt.permissionType, "channels")

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
			} else if tt.workspaces != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tt.workspaces, tt.permissionType, "workspaces")

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
			} else if tt.organizations != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tt.organizations, tt.permissionType, "organizations")

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
			}
		})
	}

	verifyCommonErrorCases(t, workflowsTriggersPermissionsSetMethod, func(c *Client) error {
		ctx := slackcontext.MockContext(t.Context())
		_, err := c.TriggerPermissionsSet(ctx, "xoxp-123", "Ft123", "user1", types.APP_COLLABORATORS, "users")
		return err
	})
}

func TestClient_TriggerPermissionsAddEntities(t *testing.T) {
	var fakeTriggerID = "Ft123"
	var fakeToken = "xoxp-123"

	tests := []struct {
		name            string
		resultJSON      string
		expectedRequest string
		users           string
		channels        string
		workspaces      string
		organizations   string
		wantErr         bool
		errMessage      string
	}{
		{
			name:            "Add user successfully",
			users:           "U0001",
			expectedRequest: `token=xoxp-123&trigger_id=Ft123&user_ids=U0001`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0001", "U0002" ]}`,
		},
		{
			name:            "Add channel successfully",
			channels:        "C0001",
			expectedRequest: `channel_ids=C0001&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0001", "C0002" ]}`,
		},
		{
			name:            "Add workspace successfully",
			workspaces:      "T0001",
			expectedRequest: `team_ids=T0001&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "team_ids": [ "T0001", "T0002" ]}`,
		},
		{
			name:            "Add organization successfully",
			organizations:   "E0001",
			expectedRequest: `org_ids=E0001&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0001", "E0002" ]}`,
		},
		{
			name:       "Propagates errors",
			resultJSON: `{"ok": false, "error":"user_not_found"}`,
			wantErr:    true,
			errMessage: "user_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsAddMethod,
				ExpectedRequest: tt.expectedRequest,
				Response:        tt.resultJSON,
			})
			defer teardown()

			// execute
			if tt.users != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tt.users, "users")
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
			} else if tt.channels != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tt.channels, "channels")
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
			} else if tt.workspaces != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tt.workspaces, "workspaces")
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
			} else if tt.organizations != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tt.organizations, "organizations")
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
			}

		})
	}

	verifyCommonErrorCases(t, workflowsTriggersPermissionsAddMethod, func(c *Client) error {
		ctx := slackcontext.MockContext(t.Context())
		return c.TriggerPermissionsAddEntities(ctx, "xoxp-123", "Ft123", "user1", "users")
	})
}

func TestClient_TriggerPermissionsRemoveEntities(t *testing.T) {
	var fakeTriggerID = "Ft123"
	var fakeToken = "xoxp-123"

	tests := []struct {
		name            string
		resultJSON      string
		expectedRequest string
		users           string
		channels        string
		workspaces      string
		organizations   string
		wantErr         bool
		errMessage      string
	}{
		{
			name:            "Remove user successfully",
			users:           "U0001",
			expectedRequest: `token=xoxp-123&trigger_id=Ft123&user_ids=U0001`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0002" ]}`,
		},
		{
			name:            "Remove channel successfully",
			channels:        "C0001",
			expectedRequest: `channel_ids=C0001&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0002" ]}`,
		},
		{
			name:            "Remove workspace successfully",
			workspaces:      "T0001",
			expectedRequest: `team_ids=T0001&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "team_ids": [ "T0002" ]}`,
		},
		{
			name:            "Remove organization successfully",
			organizations:   "E0001",
			expectedRequest: `org_ids=E0001&token=xoxp-123&trigger_id=Ft123`,
			resultJSON:      `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0002" ]}`,
		},
		{
			name:       "Propagates errors",
			resultJSON: `{"ok": false, "error":"user_not_found"}`,
			wantErr:    true,
			errMessage: "user_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsRemoveMethod,
				ExpectedRequest: tt.expectedRequest,
				Response:        tt.resultJSON,
			})
			defer teardown()

			// execute
			if tt.users != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tt.users, "users")

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
			} else if tt.channels != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tt.channels, "channels")

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
			} else if tt.workspaces != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tt.workspaces, "workspaces")

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
			} else if tt.organizations != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tt.organizations, "organizations")

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
			}
		})
	}

	verifyCommonErrorCases(t, workflowsTriggersPermissionsRemoveMethod, func(c *Client) error {
		ctx := slackcontext.MockContext(t.Context())
		return c.TriggerPermissionsRemoveEntities(ctx, "xoxp-123", "Ft123", "user1", "users")
	})
}

func TestClient_TriggerPermissionsList(t *testing.T) {
	var fakeTriggerID = "Ft123"
	var fakeToken = "xoxp-123"

	tests := []struct {
		name                   string
		resultJSON             string
		expectedRequest        string
		expectedPermissionType types.Permission
		expectedUsers          []string
		expectedChannels       []string
		expectedWorkspaces     []string
		expectedOrganizations  []string
		wantErr                bool
		errMessage             string
	}{
		{
			name:                   "Access is everyone",
			expectedPermissionType: types.EVERYONE,
			expectedUsers:          []string(nil),
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			resultJSON:             `{"ok": true, "permission_type": "everyone"}`,
		},
		{
			name:                   "Access is collaborators",
			expectedPermissionType: types.APP_COLLABORATORS,
			expectedUsers:          []string{"U0001"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			resultJSON:             `{"ok": true, "permission_type": "app_collaborators", "user_ids": [ "U0001" ]}`,
		},
		{
			name:                   "Set to named_entities (users)",
			expectedPermissionType: types.NAMED_ENTITIES,
			expectedUsers:          []string{"U0001", "U0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			resultJSON:             `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0001", "U0002" ]}`,
		},
		{
			name:                   "Set to named_entities (channels)",
			expectedPermissionType: types.NAMED_ENTITIES,
			expectedChannels:       []string{"C0001", "C0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			resultJSON:             `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0001", "C0002" ]}`,
		},
		{
			name:                   "Set to named_entities (workspaces)",
			expectedPermissionType: types.NAMED_ENTITIES,
			expectedWorkspaces:     []string{"T0001", "T0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			resultJSON:             `{"ok": true,"permission_type": "named_entities", "team_ids": [ "T0001", "T0002" ]}`,
		},
		{
			name:                   "Set to named_entities (organizations)",
			expectedPermissionType: types.NAMED_ENTITIES,
			expectedOrganizations:  []string{"E0001", "E0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			resultJSON:             `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0001", "E0002" ]}`,
		},
		{
			name:          "Propagates errors",
			resultJSON:    `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:       true,
			errMessage:    "invalid_scopes",
			expectedUsers: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsListMethod,
				ExpectedRequest: tt.expectedRequest,
				Response:        tt.resultJSON,
			})
			defer teardown()

			// execute
			if len(tt.expectedUsers) != 0 {
				actualType, actualUsers, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tt.expectedPermissionType, actualType)
				require.Equal(t, tt.expectedUsers, actualUsers)

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
			} else if len(tt.expectedChannels) != 0 {
				actualType, actualChannels, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tt.expectedPermissionType, actualType)
				require.Equal(t, tt.expectedChannels, actualChannels)

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
			} else if len(tt.expectedWorkspaces) != 0 {
				actualType, actualWorkspaces, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tt.expectedPermissionType, actualType)
				require.Equal(t, tt.expectedWorkspaces, actualWorkspaces)

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
			} else if len(tt.expectedOrganizations) != 0 {
				actualType, actualOrganizations, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tt.expectedPermissionType, actualType)
				require.Equal(t, tt.expectedOrganizations, actualOrganizations)

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
			}
		})
	}

	verifyCommonErrorCases(t, workflowsTriggersPermissionsListMethod, func(c *Client) error {
		ctx := slackcontext.MockContext(t.Context())
		_, _, err := c.TriggerPermissionsList(ctx, "xoxp-123", "Ft123")
		return err
	})
}
