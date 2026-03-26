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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
)

const docsSearchAPIURL = "https://docs-slack-d-search-api-duu9zr.herokuapp.com"

type DocsClient interface {
	DocsSearch(ctx context.Context, query string, limit int) (*DocsSearchResponse, error)
}

type DocsSearchResponse struct {
	TotalResults int              `json:"total_results"`
	Results      []DocsSearchItem `json:"results"`
	Limit        int              `json:"limit"`
}

type DocsSearchItem struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// DocsSearch searches the Slack developer docs API
func (c *Client) DocsSearch(ctx context.Context, query string, limit int) (*DocsSearchResponse, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "apiclient.DocsSearch")
	defer span.Finish()

	endpoint := fmt.Sprintf("api/search?q=%s&limit=%d", url.QueryEscape(query), limit)
	sURL := docsSearchAPIURL + "/" + endpoint

	span.SetTag("request_url", sURL)

	req, err := http.NewRequestWithContext(ctx, "GET", sURL, nil)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errHTTPRequestFailed.WithMessage(fmt.Sprintf("API returned status %d", resp.StatusCode))
	}

	var searchResponse DocsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err)
	}

	return &searchResponse, nil
}
