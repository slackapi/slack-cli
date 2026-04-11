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

var docsBaseURL = "https://docs.slack.dev"

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

func buildDocsSearchURL(baseURL, query string, limit int) (string, error) {
	endpoint := fmt.Sprintf("%s?query=%s&limit=%d", docsSearchMethod, url.QueryEscape(query), limit)
	sURL, err := url.Parse(baseURL + "/" + endpoint)
	if err != nil {
		return "", err
	}
	return sURL.String(), nil
}

func buildDocsSearchRequest(ctx context.Context, urlStr, cliVersion string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS))
	return req, nil
}

// DocsSearch searches the Slack developer docs API
func (c *Client) DocsSearch(ctx context.Context, query string, limit int) (*DocsSearchResponse, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "apiclient.DocsSearch")
	defer span.Finish()

	urlStr, err := buildDocsSearchURL(docsBaseURL, query, limit)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	sURL, _ := url.Parse(urlStr)
	span.SetTag("request_url", sURL)

	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return nil, err
	}

	req, err := buildDocsSearchRequest(ctx, urlStr, cliVersion)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

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
