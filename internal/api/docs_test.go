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

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/require"
)

func Test_Client_DocsSearch(t *testing.T) {
	tests := map[string]struct {
		query                 string
		limit                 int
		response              string
		statusCode            int
		expectedQuerystring   string
		expectedResponse      *DocsSearchResponse
		expectedErrorContains string
	}{
		"returns search results": {
			query:               "Block Kit",
			limit:               20,
			response:            `{"total_results":2,"limit":20,"results":[{"title":"Block Kit","url":"/block-kit"},{"title":"Block Kit Elements","url":"/block-kit/elements"}]}`,
			expectedQuerystring: "query=Block+Kit&limit=20",
			expectedResponse: &DocsSearchResponse{
				TotalResults: 2,
				Limit:        20,
				Results: []DocsSearchItem{
					{
						Title: "Block Kit",
						URL:   "/block-kit",
					},
					{
						Title: "Block Kit Elements",
						URL:   "/block-kit/elements",
					},
				},
			},
		},
		"returns empty results": {
			query:    "nonexistent",
			limit:    20,
			response: `{"total_results":0,"limit":20,"results":[]}`,
			expectedResponse: &DocsSearchResponse{
				TotalResults: 0,
				Limit:        20,
				Results:      []DocsSearchItem{},
			},
		},
		"encodes query parameters": {
			query:               "messages & webhooks",
			limit:               5,
			response:            `{"total_results":0,"limit":5,"results":[]}`,
			expectedQuerystring: "query=messages+%26+webhooks&limit=5",
			expectedResponse: &DocsSearchResponse{
				TotalResults: 0,
				Limit:        5,
				Results:      []DocsSearchItem{},
			},
		},
		"returns error for non-OK status": {
			query:                 "test",
			limit:                 20,
			statusCode:            404,
			expectedErrorContains: slackerror.ErrHTTPRequestFailed,
		},
		"returns error for invalid JSON": {
			query:                 "test",
			limit:                 20,
			response:              `{invalid json}`,
			expectedErrorContains: slackerror.ErrHTTPResponseInvalid,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:      docsSearchMethod,
				ExpectedQuerystring: tc.expectedQuerystring,
				Response:            tc.response,
				StatusCode:          tc.statusCode,
			})
			defer teardown()

			originalURL := docsBaseURL
			docsBaseURL = c.Host()
			defer func() { docsBaseURL = originalURL }()

			result, err := c.DocsSearch(ctx, tc.query, tc.limit)

			if tc.expectedErrorContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrorContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}
		})
	}
}
