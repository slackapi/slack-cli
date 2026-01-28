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
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/require"
)

func TestClient_AppsDatastorePut(t *testing.T) {
	type args struct {
		request types.AppDatastorePut
	}
	tests := map[string]struct {
		args             args
		httpResponseJSON string
		statusCode       int
		want             types.AppDatastorePutResult
		wantErr          bool
		errMessage       string
	}{
		"success": {
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			httpResponseJSON: `{"ok": true, "datastore": "my_datastore", "item": {"my_item": "my_value"}}`,
			want: types.AppDatastorePutResult{
				Datastore: "my_datastore",
				Item: map[string]interface{}{
					"my_item": "my_value",
				},
			},
		},
		"http_error": {
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			statusCode: 500,
			wantErr:    true,
			errMessage: slackerror.ErrHTTPRequestFailed,
		},
		"response_unmarshal_error": {
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			httpResponseJSON: `{"ok": true`,
			wantErr:          true,
			errMessage:       slackerror.ErrHTTPResponseInvalid,
		},
		"api_error": {
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			httpResponseJSON: `{"ok":false,"error":"datastore_error","errors":[{"code":"server_error","message":"Datastore error","pointer":"/datastores"}]}`,
			wantErr:          true,
			errMessage:       "datastore_error",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appDatastorePutMethod,
				Response:       tc.httpResponseJSON,
				StatusCode:     tc.statusCode,
			})
			defer teardown()
			got, err := c.AppsDatastorePut(ctx, "shhh", tc.args.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("Client.AppsDatastorePut() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"Client.AppsDatastorePut() error contains invalid message",
				)
			}
			require.Equal(t, tc.want, got, "Client.AppsDatastorePut() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreUpdate(t *testing.T) {
	type args struct {
		request types.AppDatastoreUpdate
	}
	tests := map[string]struct {
		args             args
		httpResponseJSON string
		statusCode       int
		want             types.AppDatastoreUpdateResult
		wantErr          bool
		errMessage       string
	}{
		"success": {
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			httpResponseJSON: `{"ok": true, "datastore": "my_datastore", "item": {"my_item": "my_value"}}`,
			want: types.AppDatastoreUpdateResult{
				Datastore: "my_datastore",
				Item: map[string]interface{}{
					"my_item": "my_value",
				},
			},
		},
		"http_error": {
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			statusCode: 500,
			wantErr:    true,
			errMessage: slackerror.ErrHTTPRequestFailed,
		},
		"response_unmarshal_error": {
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			httpResponseJSON: `{"ok": true`,
			wantErr:          true,
			errMessage:       slackerror.ErrHTTPResponseInvalid,
		},
		"api_error": {
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			httpResponseJSON: `{"ok":false,"error":"datastore_error","errors":[{"code":"server_error","message":"Datastore error","pointer":"/datastores"}]}`,
			wantErr:          true,
			errMessage:       "datastore_error",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appDatastoreUpdateMethod,
				Response:       tc.httpResponseJSON,
				StatusCode:     tc.statusCode,
			})
			defer teardown()
			got, err := c.AppsDatastoreUpdate(ctx, "shhh", tc.args.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("Client.AppsDatastoreUpdate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"Client.AppsDatastoreUpdate() error contains invalid message",
				)
			}
			require.Equal(t, tc.want, got, "Client.AppsDatastoreUpdate() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreQuery(t *testing.T) {
	type args struct {
		query types.AppDatastoreQuery
	}
	tests := map[string]struct {
		args             args
		httpResponseJSON string
		statusCode       int
		want             types.AppDatastoreQueryResult
		wantErr          bool
		errMessage       string
	}{
		"success": {
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			httpResponseJSON: `{"ok": true, "datastore": "my_datastore", "items": [{"my_item": "my_value"},{"my_item2": "my_value2"}]}`,
			want: types.AppDatastoreQueryResult{
				Datastore: "my_datastore",
				Items: []map[string]interface{}{
					{"my_item": "my_value"},
					{"my_item2": "my_value2"},
				},
			},
		},
		"http_error": {
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			statusCode: 500,
			wantErr:    true,
			errMessage: slackerror.ErrHTTPRequestFailed,
		},
		"response_unmarshal_error": {
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			httpResponseJSON: `{"ok": true`,
			wantErr:          true,
			errMessage:       slackerror.ErrHTTPResponseInvalid,
		},
		"api_error": {
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			httpResponseJSON: `{"ok": false}`,
			wantErr:          true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appDatastoreQueryMethod,
				Response:       tc.httpResponseJSON,
				StatusCode:     tc.statusCode,
			})
			defer teardown()
			got, err := c.AppsDatastoreQuery(ctx, "shhh", tc.args.query)
			if (err != nil) != tc.wantErr {
				t.Errorf("Client.AppsDatastoreQuery() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"Client.AppsDatastoreQuery() error contains invalid message",
				)
			}
			require.Equal(t, tc.want, got, "Client.AppsDatastoreQuery() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreDelete(t *testing.T) {
	type args struct {
		request types.AppDatastoreDelete
	}
	tests := map[string]struct {
		args             args
		httpResponseJSON string
		statusCode       int
		want             types.AppDatastoreDeleteResult
		wantErr          bool
		errMessage       string
	}{
		"success": {
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			httpResponseJSON: `{"ok": true, "datastore": "my_datastore"}`,
			want: types.AppDatastoreDeleteResult{
				Datastore: "my_datastore",
				ID:        "my_id",
			},
		},
		"http_error": {
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			statusCode: 500,
			wantErr:    true,
			errMessage: slackerror.ErrHTTPRequestFailed,
		},
		"response_unmarshal_error": {
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			httpResponseJSON: `{"ok": true`,
			wantErr:          true,
			errMessage:       slackerror.ErrHTTPResponseInvalid,
		},
		"api_error": {
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			httpResponseJSON: `{"ok": false}`,
			wantErr:          true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appDatastoreDeleteMethod,
				Response:       tc.httpResponseJSON,
				StatusCode:     tc.statusCode,
			})
			defer teardown()
			got, err := c.AppsDatastoreDelete(ctx, "shhh", tc.args.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("Client.AppsDatastoreDelete() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"Client.AppsDatastorePut() error contains invalid message",
				)
			}
			require.Equal(t, tc.want, got, "Client.AppsDatastoreDelete() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreGet(t *testing.T) {
	type args struct {
		request types.AppDatastoreGet
	}
	tests := map[string]struct {
		args             args
		httpResponseJSON string
		statusCode       int
		want             types.AppDatastoreGetResult
		wantErr          bool
		errMessage       string
	}{
		"success": {
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			httpResponseJSON: `{"ok": true, "datastore": "my_datastore", "item": {"my_item": "my_value"}}`,
			want: types.AppDatastoreGetResult{
				Datastore: "my_datastore",
				Item: map[string]interface{}{
					"my_item": "my_value",
				},
			},
		},
		"http_error": {
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			statusCode: 500,
			wantErr:    true,
			errMessage: slackerror.ErrHTTPRequestFailed,
		},
		"response_unmarshal_error": {
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			httpResponseJSON: `{"ok": true`,
			wantErr:          true,
			errMessage:       slackerror.ErrHTTPResponseInvalid,
		},
		"api_error": {
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			httpResponseJSON: `{"ok": false}`,
			wantErr:          true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appDatastoreGetMethod,
				Response:       tc.httpResponseJSON,
				StatusCode:     tc.statusCode,
			})
			defer teardown()
			got, err := c.AppsDatastoreGet(ctx, "shhh", tc.args.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("Client.AppsDatastoreGet() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if tc.wantErr {
				require.Contains(
					t,
					err.Error(),
					tc.errMessage,
					"Client.AppsDatastorePut() error contains invalid message",
				)
			}
			require.Equal(t, tc.want, got, "Client.AppsDatastoreGet() result was unexpected")
		})
	}
}
