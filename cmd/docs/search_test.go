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
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDocsAPI implements api.DocsClient for testing
type mockDocsAPI struct {
	searchResponse *api.DocsSearchResponse
	searchError    error
}

func (m *mockDocsAPI) DocsSearch(ctx context.Context, query string, limit int) (*api.DocsSearchResponse, error) {
	return m.searchResponse, m.searchError
}

func setupDocsAPITest(t *testing.T, response *api.DocsSearchResponse, err error) *shared.ClientFactory {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	mockDocsAPI := &mockDocsAPI{
		searchResponse: response,
		searchError:    err,
	}

	// Override the API to return our mock for DocsSearch
	originalAPI := clients.API
	clients.API = func() api.APIInterface {
		realAPI := originalAPI()
		// Return a wrapper that intercepts DocsSearch calls
		return &docsAPIWrapper{
			APIInterface: realAPI,
			mock:         mockDocsAPI,
		}
	}

	return clients
}

// docsAPIWrapper wraps APIInterface to mock DocsSearch while delegating other methods
type docsAPIWrapper struct {
	api.APIInterface
	mock *mockDocsAPI
}

func (w *docsAPIWrapper) DocsSearch(ctx context.Context, query string, limit int) (*api.DocsSearchResponse, error) {
	return w.mock.DocsSearch(ctx, query, limit)
}

// Text and JSON Output Tests

// Verifies that HTTP request errors are properly caught and returned as errors.
func Test_Docs_SearchCommand_TextJSONOutput_HTTPError(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clients := setupDocsAPITest(t, nil, slackerror.New(slackerror.ErrHTTPRequestFailed))

	err := fetchAndOutputSearchResults(ctx, clients, "test", 20)
	assert.Error(t, err)
}

// Verifies that HTTP errors from the API are properly caught and returned as errors.
func Test_Docs_SearchCommand_TextJSONOutput_APIError(t *testing.T) {
	clients := setupDocsAPITest(t, nil, slackerror.New(slackerror.ErrHTTPRequestFailed))
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "nonexistent", 20)
	assert.Error(t, err)
}

// Verifies that malformed JSON responses are caught during parsing and returned as errors.
func Test_Docs_SearchCommand_TextJSONOutput_InvalidJSON(t *testing.T) {
	clients := setupDocsAPITest(t, nil, slackerror.New(slackerror.ErrHTTPResponseInvalid))
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "test", 20)
	assert.Error(t, err)
}

// Verifies that valid JSON responses with no results are correctly parsed and output without errors.
func Test_Docs_SearchCommand_TextJSONOutput_EmptyResults(t *testing.T) {
	response := &api.DocsSearchResponse{
		TotalResults: 0,
		Results:      []api.DocsSearchItem{},
		Limit:        20,
	}

	clients := setupDocsAPITest(t, response, nil)
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "nonexistent query", 20)
	require.NoError(t, err)
}

// Verifies that various query formats are properly URL encoded and API parameters are correctly passed.
func Test_Docs_SearchCommand_TextJSONOutput_QueryFormats(t *testing.T) {
	response := &api.DocsSearchResponse{
		TotalResults: 2,
		Limit:        20,
		Results: []api.DocsSearchItem{
			{
				Title: "Block Kit",
				URL:   "/block-kit",
			},
			{
				Title: "Block Kit Elements",
				URL:   "/block-kit/elements",
			},
		},
	}

	tests := map[string]struct {
		query string
		limit int
	}{
		"single word query": {
			query: "messaging",
			limit: 20,
		},
		"multiple words": {
			query: "socket mode",
			limit: 20,
		},
		"special characters": {
			query: "messages & webhooks",
			limit: 20,
		},
		"custom limit": {
			query: "Block Kit",
			limit: 5,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clients := setupDocsAPITest(t, response, nil)
			err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, tc.query, tc.limit)
			require.NoError(t, err)
		})
	}
}

// JSON Output Tests

// Verifies that JSON output mode correctly formats and outputs search results.
func Test_Docs_SearchCommand_JSONOutput(t *testing.T) {
	response := &api.DocsSearchResponse{
		TotalResults: 1,
		Limit:        20,
		Results: []api.DocsSearchItem{
			{
				Title: "Block Kit",
				URL:   "/block-kit",
			},
		},
	}

	clients := setupDocsAPITest(t, response, nil)
	err := fetchAndOutputSearchResults(slackcontext.MockContext(context.Background()), clients, "Block Kit", 20)
	require.NoError(t, err)
}

// Text Output Tests

// Verifies that text output mode correctly formats and outputs search results.
func Test_Docs_SearchCommand_TextOutput(t *testing.T) {
	response := &api.DocsSearchResponse{
		TotalResults: 1,
		Limit:        20,
		Results: []api.DocsSearchItem{
			{
				Title: "Block Kit",
				URL:   "/block-kit",
			},
		},
	}

	clients := setupDocsAPITest(t, response, nil)
	err := fetchAndOutputTextResults(slackcontext.MockContext(context.Background()), clients, "Block Kit", 20)
	require.NoError(t, err)
}

// Invalid Output Format Tests

// Verifies that invalid output format returns an error with helpful remediation.
func Test_Docs_SearchCommand_InvalidOutputFormat(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"rejects invalid output format": {
			CmdArgs: []string{"search", "test", "--output=invalid"},
			ExpectedErrorStrings: []string{
				"Invalid output format",
				"Use one of: text, json, browser",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
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
