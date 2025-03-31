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
	"github.com/stretchr/testify/require"
)

func Test_API_AppsAuthExternalStart(t *testing.T) {
	tests := []struct {
		name                     string
		argsToken                string
		argsAppID                string
		argsProviderKey          string
		httpResponseJSON         string
		expectedAuthorizationURL string
		expectedErrorContains    string
	}{
		{
			name:                     "Successful request",
			argsToken:                "xoxp-123",
			argsAppID:                "A0123",
			argsProviderKey:          "provider-key",
			httpResponseJSON:         `{"ok": true, "authorization_url": "http://slack.com/authorization/url"}`,
			expectedAuthorizationURL: "http://slack.com/authorization/url",
			expectedErrorContains:    "",
		},
		{
			name:                     "Response contains an error",
			argsToken:                "xoxp-123",
			argsAppID:                "A0123",
			argsProviderKey:          "provider-key",
			httpResponseJSON:         `{"ok":false,"error":"invalid scopes"}`,
			expectedAuthorizationURL: "",
			expectedErrorContains:    "invalid scopes",
		},
		{
			name:                     "Response contains invalid JSON",
			argsToken:                "xoxp-123",
			argsAppID:                "A0123",
			argsProviderKey:          "provider-key",
			httpResponseJSON:         `this is not valid json {"ok": true, "authorization_url": "http://slack.com/authorization/url"}`,
			expectedAuthorizationURL: "",
			expectedErrorContains:    errHttpResponseInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup HTTP test server
			httpHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
				json := tt.httpResponseJSON
				_, err := fmt.Fprintln(w, json)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(httpHandlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)

			// Execute test
			authorizationURL, err := apiClient.AppsAuthExternalStart(context.Background(), tt.argsToken, tt.argsAppID, tt.argsProviderKey)

			// Assertions
			require.Equal(t, tt.expectedAuthorizationURL, authorizationURL)
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}

func Test_API_AppsAuthExternalRemove(t *testing.T) {
	tests := []struct {
		name                  string
		argsToken             string
		argsAppID             string
		argsProviderKey       string
		httpResponseJSON      string
		expectedErrorContains string
	}{
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `{"ok": true}`,
			expectedErrorContains: "",
		},
		{
			name:                  "Response contains an error",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `{"ok":false,"error":"invalid scopes"}`,
			expectedErrorContains: "invalid scopes",
		},
		{
			name:                  "Response contains invalid JSON",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHttpResponseInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup HTTP test server
			httpHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
				json := tt.httpResponseJSON
				_, err := fmt.Fprintln(w, json)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(httpHandlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)

			// Execute test
			err := apiClient.AppsAuthExternalDelete(context.Background(), tt.argsToken, tt.argsAppID, tt.argsProviderKey, "")

			// Assertions
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}

func Test_API_AppsAuthExternalClientSecretAdd(t *testing.T) {
	tests := []struct {
		name                  string
		argsToken             string
		argsAppID             string
		argsProviderKey       string
		argsClientSecret      string
		httpResponseJSON      string
		expectedErrorContains string
	}{
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsClientSecret:      "xxx-secret-xxx",
			httpResponseJSON:      `{"ok": true}`,
			expectedErrorContains: "",
		},
		{
			name:                  "Response contains an error",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsClientSecret:      "",
			httpResponseJSON:      `{"ok":false,"error":"client secret cannot be empty"}`,
			expectedErrorContains: "client secret cannot be empty",
		},
		{
			name:                  "Response contains invalid JSON",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHttpResponseInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup HTTP test server
			httpHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
				json := tt.httpResponseJSON
				_, err := fmt.Fprintln(w, json)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(httpHandlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)

			// Execute test
			err := apiClient.AppsAuthExternalClientSecretAdd(context.Background(), tt.argsToken, tt.argsAppID, tt.argsProviderKey, tt.argsClientSecret)

			// Assertions
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}

func Test_API_AppsAuthExternalList(t *testing.T) {
	tests := []struct {
		name                           string
		argsToken                      string
		argsAppID                      string
		httpResponseJSON               string
		expectedAuthorizationInfoLists types.ExternalAuthorizationInfoLists
		expectedErrorContains          string
	}{
		{
			name:             "Successful request",
			argsToken:        "xoxp-123",
			argsAppID:        "A0123",
			httpResponseJSON: `{"ok": true,  "authorizations": [ { "provider_name": "Google",  "provider_key": "google",  "client_id": "xxxxx",  "client_secret_exists": true,  "valid_token_exists": true}]}`,
			expectedAuthorizationInfoLists: types.ExternalAuthorizationInfoLists{
				Authorizations: []types.ExternalAuthorizationInfo{
					{
						ProviderName:       "Google",
						ProviderKey:        "google",
						ClientId:           "xxxxx",
						ClientSecretExists: true,
						ValidTokenExists:   true,
					},
				},
			},
			expectedErrorContains: "",
		},
		{
			name:      "Successful request with external_token_ids and external_tokens",
			argsToken: "xoxp-123",
			argsAppID: "A0123",
			httpResponseJSON: `{"ok": true,
			"authorizations": [
				{ "provider_name": "Google",
				"provider_key": "google",
				"client_id": "xxxxx",
				 "client_secret_exists": true,
				 "valid_token_exists": true, "external_token_ids": [
		        "Et0548LYDWCT"
		    ],
		    "external_tokens": [
		        {
		            "external_token_id": "Et0548LABCDE",
		            "external_user_id": "xyz@salesforce.com",
		            "date_updated": 1682021142
		        }
		    ]}]}`,
			expectedAuthorizationInfoLists: types.ExternalAuthorizationInfoLists{
				Authorizations: []types.ExternalAuthorizationInfo{
					{
						ProviderName:       "Google",
						ProviderKey:        "google",
						ClientId:           "xxxxx",
						ClientSecretExists: true,
						ValidTokenExists:   true,
						ExternalTokenIds:   []string{"Et0548LYDWCT"},
						ExternalTokens: []types.ExternalTokenInfo{
							{
								ExternalTokenId: "Et0548LABCDE",
								ExternalUserId:  "xyz@salesforce.com",
								DateUpdated:     1682021142,
							},
						},
					},
				},
			},
			expectedErrorContains: "",
		},
		{
			name:      "Successful request with workflows",
			argsToken: "xoxp-123",
			argsAppID: "A0123",
			httpResponseJSON: `{
				"ok": true,
				"authorizations":
				[
					{
					"provider_name": "Google",
					"provider_key": "google",
					"client_id": "xxxxx",
				 	"client_secret_exists": true,
				 	"valid_token_exists": true,
				 	"external_token_ids": ["Et0548LYDWCT"],
		    		"external_tokens": [
		        		{
		           			"external_token_id": "Et0548LABCDE",
		           			"external_user_id": "xyz@salesforce.com",
		            		"date_updated": 1682021142
		        		}
						]
					}
				],
				"workflows": [
			{
		    	"workflow_id": "Wf04QXGCK3FF",
		    	"callback_id": "external_auth_demo_workflow",
		   		 "providers": [
		        {
		            "provider_name": "Google",
		            "provider_key": "google"
		        }
		    	]
			}]}`,
			expectedAuthorizationInfoLists: types.ExternalAuthorizationInfoLists{
				Authorizations: []types.ExternalAuthorizationInfo{
					{
						ProviderName:       "Google",
						ProviderKey:        "google",
						ClientId:           "xxxxx",
						ClientSecretExists: true,
						ValidTokenExists:   true,
						ExternalTokenIds:   []string{"Et0548LYDWCT"},
						ExternalTokens: []types.ExternalTokenInfo{
							{
								ExternalTokenId: "Et0548LABCDE",
								ExternalUserId:  "xyz@salesforce.com",
								DateUpdated:     1682021142,
							},
						},
					},
				},
				Workflows: []types.WorkflowsInfo{
					{
						WorkflowId: "Wf04QXGCK3FF",
						CallBackId: "external_auth_demo_workflow",
						Providers: []types.ProvidersInfo{
							{
								ProviderName: "Google",
								ProviderKey:  "google",
							},
						},
					},
				},
			},
			expectedErrorContains: "",
		},
		{
			name:                  "Response contains an error",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `{"ok":false,"error":"app_not_found"}`,
			expectedErrorContains: "app_not_found",
		},
		{
			name:                  "Response contains invalid JSON",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHttpResponseInvalid.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup HTTP test server
			httpHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
				json := tt.httpResponseJSON
				_, err := fmt.Fprintln(w, json)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(httpHandlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)

			// Execute test
			actual, err := apiClient.AppsAuthExternalList(context.Background(), tt.argsToken, tt.argsAppID, false /*include_workflows flag to return workflow auth info*/)

			// Assertions
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedAuthorizationInfoLists, actual)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
func Test_API_AppsAuthExternalSelectAuth(t *testing.T) {
	tests := []struct {
		name                  string
		argsToken             string
		argsAppID             string
		argsProviderKey       string
		argsWorkflowId        string
		argsExternalTokenId   string
		argsMappingOwnerType  string
		httpResponseJSON      string
		expectedErrorContains string
	}{
		{
			name:                  "Successful request",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsWorkflowId:        "WABCD12",
			argsExternalTokenId:   "ET1234AB",
			argsMappingOwnerType:  "DEVELOPER",
			httpResponseJSON:      `{"ok": true}`,
			expectedErrorContains: "",
		},
		{
			name:                  "Response contains an error",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsWorkflowId:        "WABCD12",
			argsExternalTokenId:   "",
			argsMappingOwnerType:  "DEVELOPER",
			httpResponseJSON:      `{"ok":false,"error":"token id cannot be empty"}`,
			expectedErrorContains: "token id cannot be empty",
		},
		{
			name:                  "Response contains invalid JSON",
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsWorkflowId:        "WABCD12",
			argsExternalTokenId:   "",
			httpResponseJSON:      `{"ok":false,"error":"this is not valid json"}`,
			expectedErrorContains: "this is not valid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup HTTP test server
			httpHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
				json := tt.httpResponseJSON
				_, err := fmt.Fprintln(w, json)
				require.NoError(t, err)
			}
			ts := httptest.NewServer(http.HandlerFunc(httpHandlerFunc))
			defer ts.Close()
			apiClient := NewClient(&http.Client{}, ts.URL, nil)

			// Execute test
			err := apiClient.AppsAuthExternalSelectAuth(context.Background(), tt.argsToken, tt.argsAppID, tt.argsProviderKey, tt.argsWorkflowId, tt.argsExternalTokenId)

			// Assertions
			if tt.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tt.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
