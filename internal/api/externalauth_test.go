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

func Test_API_AppsAuthExternalStart(t *testing.T) {
	tests := map[string]struct {
		argsToken                string
		argsAppID                string
		argsProviderKey          string
		httpResponseJSON         string
		expectedAuthorizationURL string
		expectedErrorContains    string
	}{
		"Successful request": {
			argsToken:                "xoxp-123",
			argsAppID:                "A0123",
			argsProviderKey:          "provider-key",
			httpResponseJSON:         `{"ok": true, "authorization_url": "http://slack.com/authorization/url"}`,
			expectedAuthorizationURL: "http://slack.com/authorization/url",
			expectedErrorContains:    "",
		},
		"Response contains an error": {
			argsToken:                "xoxp-123",
			argsAppID:                "A0123",
			argsProviderKey:          "provider-key",
			httpResponseJSON:         `{"ok":false,"error":"invalid scopes"}`,
			expectedAuthorizationURL: "",
			expectedErrorContains:    "invalid scopes",
		},
		"Response contains invalid JSON": {
			argsToken:                "xoxp-123",
			argsAppID:                "A0123",
			argsProviderKey:          "provider-key",
			httpResponseJSON:         `this is not valid json {"ok": true, "authorization_url": "http://slack.com/authorization/url"}`,
			expectedAuthorizationURL: "",
			expectedErrorContains:    errHTTPResponseInvalid.Code,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appsAuthExternalStartMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			authorizationURL, err := c.AppsAuthExternalStart(ctx, tc.argsToken, tc.argsAppID, tc.argsProviderKey)

			// Assertions
			require.Equal(t, tc.expectedAuthorizationURL, authorizationURL)
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}

func Test_API_AppsAuthExternalRemove(t *testing.T) {
	tests := map[string]struct {
		argsToken             string
		argsAppID             string
		argsProviderKey       string
		httpResponseJSON      string
		expectedErrorContains string
	}{
		"Successful request": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `{"ok": true}`,
			expectedErrorContains: "",
		},
		"Response contains an error": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `{"ok":false,"error":"invalid scopes"}`,
			expectedErrorContains: "invalid scopes",
		},
		"Response contains invalid JSON": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHTTPResponseInvalid.Code,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appsAuthExternalDeleteMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			err := c.AppsAuthExternalDelete(ctx, tc.argsToken, tc.argsAppID, tc.argsProviderKey, "")

			// Assertions
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}

func Test_API_AppsAuthExternalClientSecretAdd(t *testing.T) {
	tests := map[string]struct {
		argsToken             string
		argsAppID             string
		argsProviderKey       string
		argsClientSecret      string
		httpResponseJSON      string
		expectedErrorContains string
	}{
		"Successful request": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsClientSecret:      "xxx-secret-xxx",
			httpResponseJSON:      `{"ok": true}`,
			expectedErrorContains: "",
		},
		"Response contains an error": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsClientSecret:      "",
			httpResponseJSON:      `{"ok":false,"error":"client secret cannot be empty"}`,
			expectedErrorContains: "client secret cannot be empty",
		},
		"Response contains invalid JSON": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHTTPResponseInvalid.Code,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appsAuthExternalClientSecretAddMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			err := c.AppsAuthExternalClientSecretAdd(ctx, tc.argsToken, tc.argsAppID, tc.argsProviderKey, tc.argsClientSecret)

			// Assertions
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}

func Test_API_AppsAuthExternalList(t *testing.T) {
	tests := map[string]struct {
		argsToken                      string
		argsAppID                      string
		httpResponseJSON               string
		expectedAuthorizationInfoLists types.ExternalAuthorizationInfoLists
		expectedErrorContains          string
	}{
		"Successful request": {
			argsToken:        "xoxp-123",
			argsAppID:        "A0123",
			httpResponseJSON: `{"ok": true,  "authorizations": [ { "provider_name": "Google",  "provider_key": "google",  "client_id": "xxxxx",  "client_secret_exists": true,  "valid_token_exists": true}]}`,
			expectedAuthorizationInfoLists: types.ExternalAuthorizationInfoLists{
				Authorizations: []types.ExternalAuthorizationInfo{
					{
						ProviderName:       "Google",
						ProviderKey:        "google",
						ClientID:           "xxxxx",
						ClientSecretExists: true,
						ValidTokenExists:   true,
					},
				},
			},
			expectedErrorContains: "",
		},
		"Successful request with external_token_ids and external_tokens": {
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
						ClientID:           "xxxxx",
						ClientSecretExists: true,
						ValidTokenExists:   true,
						ExternalTokenIDs:   []string{"Et0548LYDWCT"},
						ExternalTokens: []types.ExternalTokenInfo{
							{
								ExternalTokenID: "Et0548LABCDE",
								ExternalUserID:  "xyz@salesforce.com",
								DateUpdated:     1682021142,
							},
						},
					},
				},
			},
			expectedErrorContains: "",
		},
		"Successful request with workflows": {
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
						ClientID:           "xxxxx",
						ClientSecretExists: true,
						ValidTokenExists:   true,
						ExternalTokenIDs:   []string{"Et0548LYDWCT"},
						ExternalTokens: []types.ExternalTokenInfo{
							{
								ExternalTokenID: "Et0548LABCDE",
								ExternalUserID:  "xyz@salesforce.com",
								DateUpdated:     1682021142,
							},
						},
					},
				},
				Workflows: []types.WorkflowsInfo{
					{
						WorkflowID: "Wf04QXGCK3FF",
						CallbackID: "external_auth_demo_workflow",
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
		"Response contains an error": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `{"ok":false,"error":"app_not_found"}`,
			expectedErrorContains: "app_not_found",
		},
		"Response contains invalid JSON": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			httpResponseJSON:      `this is not valid json {"ok": true}`,
			expectedErrorContains: errHTTPResponseInvalid.Code,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appsAuthExternalListMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			actual, err := c.AppsAuthExternalList(ctx, tc.argsToken, tc.argsAppID, false /*include_workflows flag to return workflow auth info*/)

			// Assertions
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
				require.Equal(t, tc.expectedAuthorizationInfoLists, actual)
			} else {
				require.Contains(t, err.Error(), tc.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
func Test_API_AppsAuthExternalSelectAuth(t *testing.T) {
	tests := map[string]struct {
		argsToken             string
		argsAppID             string
		argsProviderKey       string
		argsWorkflowID        string
		argsExternalTokenID   string
		argsMappingOwnerType  string
		httpResponseJSON      string
		expectedErrorContains string
	}{
		"Successful request": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsWorkflowID:        "WABCD12",
			argsExternalTokenID:   "ET1234AB",
			argsMappingOwnerType:  "DEVELOPER",
			httpResponseJSON:      `{"ok": true}`,
			expectedErrorContains: "",
		},
		"Response contains an error": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsWorkflowID:        "WABCD12",
			argsExternalTokenID:   "",
			argsMappingOwnerType:  "DEVELOPER",
			httpResponseJSON:      `{"ok":false,"error":"token id cannot be empty"}`,
			expectedErrorContains: "token id cannot be empty",
		},
		"Response contains invalid JSON": {
			argsToken:             "xoxp-123",
			argsAppID:             "A0123",
			argsProviderKey:       "provider-key",
			argsWorkflowID:        "WABCD12",
			argsExternalTokenID:   "",
			httpResponseJSON:      `{"ok":false,"error":"this is not valid json"}`,
			expectedErrorContains: "this is not valid json",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: appsAuthExternalSelectAuthMethod,
				Response:       tc.httpResponseJSON,
			})
			defer teardown()

			// Execute test
			err := c.AppsAuthExternalSelectAuth(ctx, tc.argsToken, tc.argsAppID, tc.argsProviderKey, tc.argsWorkflowID, tc.argsExternalTokenID)

			// Assertions
			if tc.expectedErrorContains == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.expectedErrorContains, "Expect error contains the message")
			}
		})
	}
}
