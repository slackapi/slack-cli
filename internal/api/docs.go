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
	"runtime"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/slackcontext"
)

var docsBaseURL = "https://docs-slack-d-search-api-duu9zr.herokuapp.com"

const docsSearchMethod = "api/v1/search"

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

	endpoint := fmt.Sprintf("%s?query=%s&limit=%d", docsSearchMethod, url.QueryEscape(query), limit)
	sURL, err := url.Parse(docsBaseURL + "/" + endpoint)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	span.SetTag("request_url", sURL)

	req, err := http.NewRequestWithContext(ctx, "GET", sURL.String(), nil)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS))

	c.printRequest(ctx, req, false)

	respBytes, err := c.DoWithRetry(ctx, req, span, false, sURL)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	var searchResponse DocsSearchResponse
	if err := json.Unmarshal(respBytes, &searchResponse); err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err)
	}

	return &searchResponse, nil
}
