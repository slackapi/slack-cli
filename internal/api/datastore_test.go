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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/require"
)

func TestClient_AppsDatastorePut(t *testing.T) {
	type args struct {
		request types.AppDatastorePut
	}
	type testHandler struct {
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name        string
		args        args
		testHandler testHandler
		want        types.AppDatastorePutResult
		wantErr     bool
		errMessage  string
	}{
		{
			name: "success",
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true, "datastore": "my_datastore", "item": {"my_item": "my_value"}}`)
					require.NoError(t, err)
				},
			},
			want: types.AppDatastorePutResult{
				Datastore: "my_datastore",
				Item: map[string]interface{}{
					"my_item": "my_value",
				},
			},
		},
		{
			name: "http_error",
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpRequestFailed,
		},
		{
			name: "response_unmarshal_error",
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpResponseInvalid,
		},
		{
			name: "api_error",
			args: args{
				request: types.AppDatastorePut{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok":false,"error":"datastore_error","errors":[{"code":"server_error","message":"Datastore error","pointer":"/datastores"}]}`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: "datastore_error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.testHandler.handlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)
			got, err := apiClient.AppsDatastorePut(context.Background(), "shhh", tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AppsDatastorePut() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"Client.AppsDatastorePut() error contains invalid message",
				)
			}
			require.Equal(t, tt.want, got, "Client.AppsDatastorePut() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreUpdate(t *testing.T) {
	type args struct {
		request types.AppDatastoreUpdate
	}
	type testHandler struct {
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name        string
		args        args
		testHandler testHandler
		want        types.AppDatastoreUpdateResult
		wantErr     bool
		errMessage  string
	}{
		{
			name: "success",
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true, "datastore": "my_datastore", "item": {"my_item": "my_value"}}`)
					require.NoError(t, err)
				},
			},
			want: types.AppDatastoreUpdateResult{
				Datastore: "my_datastore",
				Item: map[string]interface{}{
					"my_item": "my_value",
				},
			},
		},
		{
			name: "http_error",
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpRequestFailed,
		},
		{
			name: "response_unmarshal_error",
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpResponseInvalid,
		},
		{
			name: "api_error",
			args: args{
				request: types.AppDatastoreUpdate{
					Datastore: "my_datastore",
					App:       "A1",
					Item: map[string]interface{}{
						"my_item": "my_value",
					},
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok":false,"error":"datastore_error","errors":[{"code":"server_error","message":"Datastore error","pointer":"/datastores"}]}`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: "datastore_error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.testHandler.handlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)
			got, err := apiClient.AppsDatastoreUpdate(context.Background(), "shhh", tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AppsDatastoreUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"Client.AppsDatastoreUpdate() error contains invalid message",
				)
			}
			require.Equal(t, tt.want, got, "Client.AppsDatastoreUpdate() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreQuery(t *testing.T) {
	type args struct {
		query types.AppDatastoreQuery
	}
	type testHandler struct {
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name        string
		args        args
		testHandler testHandler
		want        types.AppDatastoreQueryResult
		wantErr     bool
		errMessage  string
	}{
		{
			name: "success",
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true, "datastore": "my_datastore", "items": [{"my_item": "my_value"},{"my_item2": "my_value2"}]}`)
					require.NoError(t, err)
				},
			},
			want: types.AppDatastoreQueryResult{
				Datastore: "my_datastore",
				Items: []map[string]interface{}{
					{"my_item": "my_value"},
					{"my_item2": "my_value2"},
				},
			},
		},
		{
			name: "http_error",
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpRequestFailed,
		},
		{
			name: "response_unmarshal_error",
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpResponseInvalid,
		},
		{
			name: "api_error",
			args: args{
				query: types.AppDatastoreQuery{
					Datastore: "my_datastore",
					App:       "A1",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": false}`)
					require.NoError(t, err)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.testHandler.handlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)
			got, err := apiClient.AppsDatastoreQuery(context.Background(), "shhh", tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AppsDatastoreQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"Client.AppsDatastoreQuery() error contains invalid message",
				)
			}
			require.Equal(t, tt.want, got, "Client.AppsDatastoreQuery() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreDelete(t *testing.T) {
	type args struct {
		request types.AppDatastoreDelete
	}
	type testHandler struct {
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name        string
		args        args
		testHandler testHandler
		want        types.AppDatastoreDeleteResult
		wantErr     bool
		errMessage  string
	}{
		{
			name: "success",
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true, "datastore": "my_datastore"}`)
					require.NoError(t, err)
				},
			},
			want: types.AppDatastoreDeleteResult{
				Datastore: "my_datastore",
				ID:        "my_id",
			},
		},
		{
			name: "http_error",
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpRequestFailed,
		},
		{
			name: "response_unmarshal_error",
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpResponseInvalid,
		},
		{
			name: "api_error",
			args: args{
				request: types.AppDatastoreDelete{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": false}`)
					require.NoError(t, err)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.testHandler.handlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)
			got, err := apiClient.AppsDatastoreDelete(context.Background(), "shhh", tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AppsDatastoreDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"Client.AppsDatastorePut() error contains invalid message",
				)
			}
			require.Equal(t, tt.want, got, "Client.AppsDatastoreDelete() result was unexpected")
		})
	}
}

func TestClient_AppsDatastoreGet(t *testing.T) {
	type args struct {
		request types.AppDatastoreGet
	}
	type testHandler struct {
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name        string
		args        args
		testHandler testHandler
		want        types.AppDatastoreGetResult
		wantErr     bool
		errMessage  string
	}{
		{
			name: "success",
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true, "datastore": "my_datastore", "item": {"my_item": "my_value"}}`)
					require.NoError(t, err)
				},
			},
			want: types.AppDatastoreGetResult{
				Datastore: "my_datastore",
				Item: map[string]interface{}{
					"my_item": "my_value",
				},
			},
		},
		{
			name: "http_error",
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(500)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpRequestFailed,
		},
		{
			name: "response_unmarshal_error",
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": true`)
					require.NoError(t, err)
				},
			},
			wantErr:    true,
			errMessage: slackerror.ErrHttpResponseInvalid,
		},
		{
			name: "api_error",
			args: args{
				request: types.AppDatastoreGet{
					Datastore: "my_datastore",
					App:       "A1",
					ID:        "my_id",
				},
			},
			testHandler: testHandler{
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					_, err := fmt.Fprintln(w, `{"ok": false}`)
					require.NoError(t, err)
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.testHandler.handlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)
			got, err := apiClient.AppsDatastoreGet(context.Background(), "shhh", tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AppsDatastoreGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Contains(
					t,
					err.Error(),
					tt.errMessage,
					"Client.AppsDatastorePut() error contains invalid message",
				)
			}
			require.Equal(t, tt.want, got, "Client.AppsDatastoreGet() result was unexpected")
		})
	}
}
