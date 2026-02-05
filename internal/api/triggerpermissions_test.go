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

func TestClient_TriggerPermissionsSet(t *testing.T) {
	var fakeTriggerID = "Ft123"
	var fakeToken = "xoxp-123"

	tests := map[string]struct {
		httpResponseJSON string
		expectedRequest  string
		permissionType   types.Permission
		users            string
		channels         string
		workspaces       string
		organizations    string
		wantErr          bool
		errMessage       string
	}{
		"Set to everyone": {
			permissionType:   types.PermissionEveryone,
			users:            "",
			expectedRequest:  `permission_type=everyone&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true, "permission_type": "everyone"}`,
		},
		"Set to collaborators": {
			permissionType:   types.PermissionAppCollaborators,
			users:            "U0001",
			expectedRequest:  `permission_type=app_collaborators&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true, "permission_type": "app_collaborators", "user_ids": [ "U0001" ]}`,
		},
		"Set to named_entities (users)": {
			permissionType:   types.PermissionNamedEntities,
			users:            "U0001,U0002",
			expectedRequest:  `permission_type=named_entities&token=xoxp-123&trigger_id=Ft123&user_ids=U0001%2CU0002`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0001", "U0002" ]}`,
		},
		"Set to named_entities (channels)": {
			permissionType:   types.PermissionNamedEntities,
			channels:         "C0001,C0002",
			expectedRequest:  `channel_ids=C0001%2CC0002&permission_type=named_entities&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0001", "C0002" ]}`,
		},
		"Set to named_entities (workspaces)": {
			permissionType:   types.PermissionNamedEntities,
			workspaces:       "T0001,T0002",
			expectedRequest:  `permission_type=named_entities&team_ids=T0001%2CT0002&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "teams_ids": [ "T0001", "T0002" ]}`,
		},
		"Set to named_entities (organizations)": {
			permissionType:   types.PermissionNamedEntities,
			organizations:    "E0001,E0002",
			expectedRequest:  `org_ids=E0001%2CE0002&permission_type=named_entities&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0001", "E0002" ]}`,
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
				ExpectedMethod:  workflowsTriggersPermissionsSetMethod,
				ExpectedRequest: tc.expectedRequest,
				Response:        tc.httpResponseJSON,
			})
			defer teardown()

			// execute
			if tc.users != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tc.users, tc.permissionType, "users")

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
			} else if tc.channels != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tc.channels, tc.permissionType, "channels")

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
			} else if tc.workspaces != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tc.workspaces, tc.permissionType, "workspaces")

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
			} else if tc.organizations != "" {
				_, err := c.TriggerPermissionsSet(ctx, fakeToken, fakeTriggerID, tc.organizations, tc.permissionType, "organizations")

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
			}
		})
	}

	verifyCommonErrorCases(t, workflowsTriggersPermissionsSetMethod, func(c *Client) error {
		ctx := slackcontext.MockContext(t.Context())
		_, err := c.TriggerPermissionsSet(ctx, "xoxp-123", "Ft123", "user1", types.PermissionAppCollaborators, "users")
		return err
	})
}

func TestClient_TriggerPermissionsAddEntities(t *testing.T) {
	var fakeTriggerID = "Ft123"
	var fakeToken = "xoxp-123"

	tests := map[string]struct {
		httpResponseJSON string
		expectedRequest  string
		users            string
		channels         string
		workspaces       string
		organizations    string
		wantErr          bool
		errMessage       string
	}{
		"Add user successfully": {
			users:            "U0001",
			expectedRequest:  `token=xoxp-123&trigger_id=Ft123&user_ids=U0001`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0001", "U0002" ]}`,
		},
		"Add channel successfully": {
			channels:         "C0001",
			expectedRequest:  `channel_ids=C0001&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0001", "C0002" ]}`,
		},
		"Add workspace successfully": {
			workspaces:       "T0001",
			expectedRequest:  `team_ids=T0001&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "team_ids": [ "T0001", "T0002" ]}`,
		},
		"Add organization successfully": {
			organizations:    "E0001",
			expectedRequest:  `org_ids=E0001&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0001", "E0002" ]}`,
		},
		"Propagates errors": {
			httpResponseJSON: `{"ok": false, "error":"user_not_found"}`,
			wantErr:          true,
			errMessage:       "user_not_found",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsAddMethod,
				ExpectedRequest: tc.expectedRequest,
				Response:        tc.httpResponseJSON,
			})
			defer teardown()

			// execute
			if tc.users != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tc.users, "users")
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
			} else if tc.channels != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tc.channels, "channels")
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
			} else if tc.workspaces != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tc.workspaces, "workspaces")
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
			} else if tc.organizations != "" {
				err := c.TriggerPermissionsAddEntities(ctx, fakeToken, fakeTriggerID, tc.organizations, "organizations")
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

	tests := map[string]struct {
		httpResponseJSON string
		expectedRequest  string
		users            string
		channels         string
		workspaces       string
		organizations    string
		wantErr          bool
		errMessage       string
	}{
		"Remove user successfully": {
			users:            "U0001",
			expectedRequest:  `token=xoxp-123&trigger_id=Ft123&user_ids=U0001`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0002" ]}`,
		},
		"Remove channel successfully": {
			channels:         "C0001",
			expectedRequest:  `channel_ids=C0001&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0002" ]}`,
		},
		"Remove workspace successfully": {
			workspaces:       "T0001",
			expectedRequest:  `team_ids=T0001&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "team_ids": [ "T0002" ]}`,
		},
		"Remove organization successfully": {
			organizations:    "E0001",
			expectedRequest:  `org_ids=E0001&token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON: `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0002" ]}`,
		},
		"Propagates errors": {
			httpResponseJSON: `{"ok": false, "error":"user_not_found"}`,
			wantErr:          true,
			errMessage:       "user_not_found",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsRemoveMethod,
				ExpectedRequest: tc.expectedRequest,
				Response:        tc.httpResponseJSON,
			})
			defer teardown()

			// execute
			if tc.users != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tc.users, "users")

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
			} else if tc.channels != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tc.channels, "channels")

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
			} else if tc.workspaces != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tc.workspaces, "workspaces")

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
			} else if tc.organizations != "" {
				err := c.TriggerPermissionsRemoveEntities(ctx, fakeToken, fakeTriggerID, tc.organizations, "organizations")

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

	tests := map[string]struct {
		httpResponseJSON       string
		expectedRequest        string
		expectedPermissionType types.Permission
		expectedUsers          []string
		expectedChannels       []string
		expectedWorkspaces     []string
		expectedOrganizations  []string
		wantErr                bool
		errMessage             string
	}{
		"Access is everyone": {
			expectedPermissionType: types.PermissionEveryone,
			expectedUsers:          []string(nil),
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON:       `{"ok": true, "permission_type": "everyone"}`,
		},
		"Access is collaborators": {
			expectedPermissionType: types.PermissionAppCollaborators,
			expectedUsers:          []string{"U0001"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON:       `{"ok": true, "permission_type": "app_collaborators", "user_ids": [ "U0001" ]}`,
		},
		"Set to named_entities (users)": {
			expectedPermissionType: types.PermissionNamedEntities,
			expectedUsers:          []string{"U0001", "U0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON:       `{"ok": true,"permission_type": "named_entities", "user_ids": [ "U0001", "U0002" ]}`,
		},
		"Set to named_entities (channels)": {
			expectedPermissionType: types.PermissionNamedEntities,
			expectedChannels:       []string{"C0001", "C0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON:       `{"ok": true,"permission_type": "named_entities", "channel_ids": [ "C0001", "C0002" ]}`,
		},
		"Set to named_entities (workspaces)": {
			expectedPermissionType: types.PermissionNamedEntities,
			expectedWorkspaces:     []string{"T0001", "T0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON:       `{"ok": true,"permission_type": "named_entities", "team_ids": [ "T0001", "T0002" ]}`,
		},
		"Set to named_entities (organizations)": {
			expectedPermissionType: types.PermissionNamedEntities,
			expectedOrganizations:  []string{"E0001", "E0002"},
			expectedRequest:        `token=xoxp-123&trigger_id=Ft123`,
			httpResponseJSON:       `{"ok": true,"permission_type": "named_entities", "org_ids": [ "E0001", "E0002" ]}`,
		},
		"Propagates errors": {
			httpResponseJSON: `{"ok": false, "error":"invalid_scopes"}`,
			wantErr:          true,
			errMessage:       "invalid_scopes",
			expectedUsers:    []string{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			// prepare
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  workflowsTriggersPermissionsListMethod,
				ExpectedRequest: tc.expectedRequest,
				Response:        tc.httpResponseJSON,
			})
			defer teardown()

			// execute
			if len(tc.expectedUsers) != 0 {
				actualType, actualUsers, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tc.expectedPermissionType, actualType)
				require.Equal(t, tc.expectedUsers, actualUsers)

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
			} else if len(tc.expectedChannels) != 0 {
				actualType, actualChannels, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tc.expectedPermissionType, actualType)
				require.Equal(t, tc.expectedChannels, actualChannels)

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
			} else if len(tc.expectedWorkspaces) != 0 {
				actualType, actualWorkspaces, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tc.expectedPermissionType, actualType)
				require.Equal(t, tc.expectedWorkspaces, actualWorkspaces)

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
			} else if len(tc.expectedOrganizations) != 0 {
				actualType, actualOrganizations, err := c.TriggerPermissionsList(ctx, fakeToken, fakeTriggerID)
				require.Equal(t, tc.expectedPermissionType, actualType)
				require.Equal(t, tc.expectedOrganizations, actualOrganizations)

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
			}
		})
	}

	verifyCommonErrorCases(t, workflowsTriggersPermissionsListMethod, func(c *Client) error {
		ctx := slackcontext.MockContext(t.Context())
		_, _, err := c.TriggerPermissionsList(ctx, "xoxp-123", "Ft123")
		return err
	})
}
