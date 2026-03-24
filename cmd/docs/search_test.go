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

package docs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRoundTripper implements http.RoundTripper to mock HTTP responses for testing.
// It allows tests to control the response status code and body without making real network calls.
// It also captures the request URL for assertion purposes.
type mockRoundTripper struct {
	response    string
	status      int
	capturedURL string
}

// RoundTrip executes a mocked HTTP request and returns a controlled response.
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedURL = req.URL.String()
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(bytes.NewBufferString(m.response)),
		Header:     make(http.Header),
	}, nil
}

func setupJSONOutputTest(t *testing.T, response string, status int) (*http.Client, *shared.ClientFactory) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	mockTransport := &mockRoundTripper{
		response: response,
		status:   status,
	}
	mockClient := &http.Client{
		Transport: mockTransport,
	}

	return mockClient, clients
}

func setupJSONOutputTestWithCapture(t *testing.T, response string, status int) (*http.Client, *shared.ClientFactory, *mockRoundTripper) {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	mockTransport := &mockRoundTripper{
		response: response,
		status:   status,
	}
	mockClient := &http.Client{
		Transport: mockTransport,
	}

	return mockClient, clients, mockTransport
}

// JSON Output Tests

// Verifies that HTTP errors from the API are properly caught and returned as errors.
func Test_Docs_SearchCommand_JSONOutput_APIError(t *testing.T) {
	mockClient, clients := setupJSONOutputTest(t, `{"error": "not found"}`, http.StatusNotFound)
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "nonexistent", 20, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API returned status 404")
}

// Verifies that malformed JSON responses are caught during parsing and returned as errors.
func Test_Docs_SearchCommand_JSONOutput_InvalidJSON(t *testing.T) {
	mockClient, clients := setupJSONOutputTest(t, `{invalid json}`, http.StatusOK)
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "test", 20, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse search results")
}

// Verifies that valid JSON responses with no results are correctly parsed and output without errors.
func Test_Docs_SearchCommand_JSONOutput_EmptyResults(t *testing.T) {
	mockResponse := `{
		"total_results": 0,
		"limit": 20,
		"results": []
	}`

	mockClient, clients := setupJSONOutputTest(t, mockResponse, http.StatusOK)
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "nonexistent query", 20, mockClient)
	require.NoError(t, err)
}

// Verifies that various query formats are properly URL encoded and API parameters are correctly passed.
func Test_Docs_SearchCommand_JSONOutput_QueryFormats(t *testing.T) {
	mockResponse := `{
		"total_results": 2,
		"limit": 20,
		"results": [
			{
				"title": "Block Kit",
				"url": "https://docs.slack.dev/block-kit"
			},
			{
				"title": "Block Kit Elements",
				"url": "https://docs.slack.dev/block-kit/elements"
			}
		]
	}`

	tests := map[string]struct {
		query    string
		limit    int
		expected string
	}{
		"single word query": {
			query:    "messaging",
			limit:    20,
			expected: "messaging",
		},
		"multiple words": {
			query:    "socket mode",
			limit:    20,
			expected: "socket+mode",
		},
		"special characters": {
			query:    "messages & webhooks",
			limit:    20,
			expected: "messages+%26+webhooks",
		},
		"custom limit": {
			query:    "Block Kit",
			limit:    5,
			expected: "Block+Kit",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockClient, clients, mockTransport := setupJSONOutputTestWithCapture(t, mockResponse, http.StatusOK)
			err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, tc.query, tc.limit, mockClient)
			require.NoError(t, err)
			assert.Contains(t, mockTransport.capturedURL, "q="+tc.expected)
			assert.Contains(t, mockTransport.capturedURL, "limit="+fmt.Sprint(tc.limit))
		})
	}
}

// Browser Output Tests

// Verifies that browser output mode correctly handles various query formats and opens the correct URLs.
func Test_Docs_SearchCommand_BrowserOutput(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"opens browser with search query using space syntax": {
			CmdArgs: []string{"search", "messaging", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=messaging"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=messaging",
			},
		},
		"handles search with multiple arguments": {
			CmdArgs: []string{"search", "Block", "Kit", "Element", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=Block+Kit+Element"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=Block+Kit+Element",
			},
		},
		"handles search query with multiple words": {
			CmdArgs: []string{"search", "socket mode", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=socket+mode"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=socket+mode",
			},
		},
		"handles special characters in search query": {
			CmdArgs: []string{"search", "messages & webhooks", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=messages+%26+webhooks"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=messages+%26+webhooks",
			},
		},
		"handles search query with quotes": {
			CmdArgs: []string{"search", "webhook \"send message\"", "--output=browser"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := "https://docs.slack.dev/search/?q=webhook+%22send+message%22"
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			ExpectedOutputs: []string{
				"Docs Search",
				"https://docs.slack.dev/search/?q=webhook+%22send+message%22",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
}
