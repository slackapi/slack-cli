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

package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const SearchIndexURL = "https://docs-slack-d-search-api-duu9zr.herokuapp.com/api/search"

// SearchResult represents a single search result
type SearchResult struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Excerpt     string  `json:"excerpt"`
	Breadcrumb  string  `json:"breadcrumb"`
	ContentType string  `json:"content_type"`
	Score       float64 `json:"score"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query        string         `json:"query"`
	Filter       string         `json:"filter"`
	Results      []SearchResult `json:"results"`
	TotalResults int            `json:"total_results"`
	Showing      int            `json:"showing"`
	Pagination   interface{}    `json:"pagination,omitempty"`
}

// SearchDocs performs a search using the hosted search API
func SearchDocs(query, filter string, limit int) (*SearchResponse, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("q", query)
	if filter != "" {
		params.Set("filter", filter)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	// Make HTTP request
	searchURL := fmt.Sprintf("%s?%s", SearchIndexURL, params.Encode())
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query search API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response directly into our response format
	var apiResponse struct {
		TotalResults int            `json:"total_results"`
		Results      []SearchResult `json:"results"`
		Pagination   interface{}    `json:"pagination,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Build response
	response := &SearchResponse{
		Query:        query,
		Filter:       filter,
		TotalResults: apiResponse.TotalResults,
		Results:      apiResponse.Results,
		Showing:      len(apiResponse.Results),
		Pagination:   apiResponse.Pagination,
	}

	return response, nil
}
