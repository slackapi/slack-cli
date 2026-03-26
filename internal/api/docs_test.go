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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	response    *http.Response
	err         error
	capturedURL string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.capturedURL = req.URL.String()
	return m.response, m.err
}

// Test_DocsSearch_Success verifies successful API response parsing
func Test_DocsSearch_Success(t *testing.T) {
	responseBody := DocsSearchResponse{
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
	}

	bodyBytes, err := json.Marshal(responseBody)
	require.NoError(t, err)

	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
			Header:     make(http.Header),
		},
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client := &Client{
		httpClient: httpClient,
	}

	result, err := client.DocsSearch(context.Background(), "Block Kit", 20)
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalResults)
	assert.Equal(t, 20, result.Limit)
	assert.Len(t, result.Results, 2)
	assert.Equal(t, "Block Kit", result.Results[0].Title)
	assert.Equal(t, "/block-kit", result.Results[0].URL)
}

// Test_DocsSearch_EmptyResults verifies handling of empty results
func Test_DocsSearch_EmptyResults(t *testing.T) {
	responseBody := DocsSearchResponse{
		TotalResults: 0,
		Limit:        20,
		Results:      []DocsSearchItem{},
	}

	bodyBytes, err := json.Marshal(responseBody)
	require.NoError(t, err)

	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
			Header:     make(http.Header),
		},
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client := &Client{
		httpClient: httpClient,
	}

	result, err := client.DocsSearch(context.Background(), "nonexistent", 20)
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalResults)
	assert.Len(t, result.Results, 0)
}

// Test_DocsSearch_QueryEncoding verifies query parameters are properly encoded
func Test_DocsSearch_QueryEncoding(t *testing.T) {
	responseBody := DocsSearchResponse{
		TotalResults: 0,
		Limit:        20,
		Results:      []DocsSearchItem{},
	}

	bodyBytes, err := json.Marshal(responseBody)
	require.NoError(t, err)

	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
			Header:     make(http.Header),
		},
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client := &Client{
		httpClient: httpClient,
	}

	_, err = client.DocsSearch(context.Background(), "messages & webhooks", 5)
	require.NoError(t, err)

	// Verify URL encoding
	assert.Contains(t, mockTransport.capturedURL, "q=messages+%26+webhooks")
	assert.Contains(t, mockTransport.capturedURL, "limit=5")
}

// Test_DocsSearch_HTTPError verifies HTTP request errors are handled
func Test_DocsSearch_HTTPError(t *testing.T) {
	mockTransport := &mockRoundTripper{
		err: fmt.Errorf("network error"),
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client := &Client{
		httpClient: httpClient,
	}

	result, err := client.DocsSearch(context.Background(), "test", 20)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// Test_DocsSearch_NonOKStatus verifies non-200 status codes are handled
func Test_DocsSearch_NonOKStatus(t *testing.T) {
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("")),
			Header:     make(http.Header),
		},
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client := &Client{
		httpClient: httpClient,
	}

	result, err := client.DocsSearch(context.Background(), "test", 20)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "404")
}

// Test_DocsSearch_InvalidJSON verifies invalid JSON responses are handled
func Test_DocsSearch_InvalidJSON(t *testing.T) {
	mockTransport := &mockRoundTripper{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("{invalid json}")),
			Header:     make(http.Header),
		},
	}

	httpClient := &http.Client{
		Transport: mockTransport,
	}

	client := &Client{
		httpClient: httpClient,
	}

	result, err := client.DocsSearch(context.Background(), "test", 20)
	assert.Error(t, err)
	assert.Nil(t, result)
}
