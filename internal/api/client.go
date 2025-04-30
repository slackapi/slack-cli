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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/uber/jaeger-client-go"
)

// Client provides an http connection for communicating with the Slack API.
type Client struct {
	host       string
	httpClient *http.Client
	io         iostreams.IOStreamer
}

type baseResponse struct {
	Ok               bool             `json:"ok"`
	Error            string           `json:"error,omitempty"`
	Description      string           `json:"slack_cli_error_description,omitempty"`
	Warning          string           `json:"warning,omitempty"`
	ResponseMetadata responseMetadata `json:"response_metadata,omitempty"`
}

type warningResponse struct {
	Warnings slackerror.Warnings `json:"warnings,omitempty"`
}

type extendedBaseResponse struct {
	baseResponse
	warningResponse
	Errors slackerror.ErrorDetails `json:"errors,omitempty"`
}

type responseMetadata struct {
	Warnings   []string `json:"warnings,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
}

var (
	errInvalidArguments    = slackerror.New(slackerror.ErrInvalidArguments)
	errHTTPResponseInvalid = slackerror.New(slackerror.ErrHTTPResponseInvalid)
	errHTTPRequestFailed   = slackerror.New(slackerror.ErrHTTPRequestFailed)
)

// NewClient accepts an httpClient to facilitate making http requests to Slack.
// Client does not attempt to evaluate the response body, leaving that to the caller.
func NewClient(client *http.Client, host string, io iostreams.IOStreamer) *Client {
	if client == nil {
		client = NewHTTPClient(HTTPClientOptions{TotalTimeOut: 60 * time.Second})
	}
	if io == nil {
		fs := slackdeps.NewFs()
		os := slackdeps.NewOs()
		config := config.NewConfig(fs, os)
		io = iostreams.NewIOStreams(config, fs, os)
	}
	return &Client{host: host, httpClient: client, io: io}
}

func (c *Client) postForm(ctx context.Context, endpoint string, formValues url.Values) ([]byte, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "postForm")
	defer span.Finish()

	var sURL *url.URL
	var err error

	sURL, err = url.Parse(fmt.Sprintf("%s/api/%s", c.host, endpoint))
	if err != nil {
		return nil, err
	}

	span.SetTag("request_url", sURL)

	// TODO: what happens if we don't have this?
	if formValues == nil {
		formValues = url.Values{}
	}

	var request *http.Request
	request, err = http.NewRequest("POST", sURL.String(), strings.NewReader(formValues.Encode()))
	if err != nil {
		return nil, err
	}

	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return nil, err
	}
	var userAgent = fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS)

	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	request.Header.Add("User-Agent", userAgent)
	if jaegerSpanContext, ok := span.Context().(jaeger.SpanContext); ok {
		request.Header.Add("x-b3-sampled", "0")
		request.Header.Add("x-b3-spanid", jaegerSpanContext.SpanID().String())
		// Use the custom trace_id from context
		request.Header.Add("x-b3-traceid", jaegerSpanContext.TraceID().String())
		request.Header.Add("x-b3-parentspanid", jaegerSpanContext.ParentID().String())
	}

	// Log request to verbose output
	skipDebugLog := shouldSkipDebugLog(endpoint)
	c.printRequest(ctx, request, skipDebugLog)

	bytes, err := c.DoWithRetry(ctx, request, span, skipDebugLog, sURL)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// postJSON sends a JSON payload to the given endpoint, and returns the response bytes from the server. The
// response bytes are never nil.
func (c *Client) postJSON(ctx context.Context, endpoint, token string, cookie string, body []byte) ([]byte, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "postJSON")
	defer span.Finish()

	var sURL, err = url.Parse(c.host + "/api/" + endpoint)
	if err != nil {
		return nil, err
	}
	span.SetTag("request_url", sURL)

	var request *http.Request
	request, err = http.NewRequest("POST", sURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	if token != "" {
		request.Header.Add("Authorization", "Bearer "+token)
	}

	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return nil, err
	}
	var userAgent = fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS)
	request.Header.Add("User-Agent", userAgent)
	if jaegerSpanContext, ok := span.Context().(jaeger.SpanContext); ok {
		request.Header.Add("x-b3-sampled", "0")
		request.Header.Add("x-b3-spanid", jaegerSpanContext.SpanID().String())
		// Use the custom trace_id from context
		request.Header.Add("x-b3-traceid", jaegerSpanContext.TraceID().String())
		request.Header.Add("x-b3-parentspanid", jaegerSpanContext.ParentID().String())
	}

	if cookie != "" {
		// this is a xoxc request
		request.Header.Add("cookie", cookie)
	}

	// Log request to verbose output
	skipDebugLog := shouldSkipDebugLog(endpoint)
	c.printRequest(ctx, request, skipDebugLog)

	bytes, err := c.DoWithRetry(ctx, request, span, skipDebugLog, sURL)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (c *Client) get(ctx context.Context, endpoint, token string, cookie string) ([]byte, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "get")
	defer span.Finish()

	var sURL, err = url.Parse(c.host + "/api/" + endpoint)
	if err != nil {
		return nil, err
	}

	span.SetTag("request_url", sURL)

	var request *http.Request
	request, err = http.NewRequest("GET", sURL.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")
	if token != "" {
		request.Header.Add("Authorization", "Bearer "+token)
	}

	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return nil, err
	}
	var userAgent = fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS)

	request.Header.Add("User-Agent", userAgent)
	if jaegerSpanContext, ok := span.Context().(jaeger.SpanContext); ok {
		request.Header.Add("x-b3-sampled", "0")
		request.Header.Add("x-b3-spanid", jaegerSpanContext.SpanID().String())
		// Use the custom trace_id from context
		request.Header.Add("x-b3-traceid", jaegerSpanContext.TraceID().String())
		request.Header.Add("x-b3-parentspanid", jaegerSpanContext.ParentID().String())
	}

	if cookie != "" {
		// this is a xoxc request
		request.Header.Add("cookie", cookie)
	}

	// Log request to verbose output
	skipDebugLog := shouldSkipDebugLog(endpoint)
	c.printRequest(ctx, request, skipDebugLog)

	bytes, err := c.DoWithRetry(ctx, request, span, skipDebugLog, sURL)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// DoWithRetry will execute the request and retry failed requests if the status
// indicates the request may be retryable and a Retry-After header is present.
func (c *Client) DoWithRetry(ctx context.Context, request *http.Request, span opentracing.Span, skipDebugLog bool, sURL *url.URL) ([]byte, error) {

	var data []byte
	var err error
	if request.Body != nil {
		data, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}

		if err = request.Body.Close(); err != nil {
			return nil, err
		}
	}

	var r *http.Response
	for {
		request.Body = io.NopCloser(bytes.NewReader(data))
		r, err = c.httpClient.Do(request)
		if err != nil {
			return nil, err
		}

		// If a Retry-After HTTP header is present we will pause for that amount of time before re-attempting the request
		delay, ok := getRetryAfter(r)
		if !ok {
			break
		}

		c.io.PrintDebug(ctx, "%s responded with status %d. Retrying request in %s...", sURL.Path, r.StatusCode, delay)

		time.Sleep(delay)
	}

	defer r.Body.Close()

	c.printResponse(ctx, r, skipDebugLog)

	span.SetTag("status_code", r.StatusCode)

	if r.StatusCode != http.StatusOK {
		return nil, errors.WithStack(fmt.Errorf("Slack API unexpected status code %d returned from url %s", r.StatusCode, sURL))
	}

	var bytes []byte
	bytes, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var baseResp baseResponse
	_ = json.Unmarshal(bytes, &baseResp)

	span.SetTag("ok", baseResp.Ok)
	if !baseResp.Ok {
		// Response is not ok, log the error into debug.log if skipDebugLog is true
		if skipDebugLog {
			c.printRequest(ctx, request, false /* don't skipDebugLog */)
			c.printResponse(ctx, r, false /* don't skipDebugLog */)
		}
		span.SetTag("error", baseResp.Error)
	}

	if bytes == nil {
		return nil, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	return bytes, err
}

// getRetryAfter returns the value of the Retry-After header for applicable requests.
// More information can be found here: https://docs.slack.dev/apis/web-api/rate-limits/
func getRetryAfter(r *http.Response) (time.Duration, bool) {
	if !(r.StatusCode == http.StatusTooManyRequests || r.StatusCode == http.StatusServiceUnavailable) {
		return 0, false
	}

	delayStr := r.Header.Get("Retry-After")
	if len(delayStr) == 0 {
		return 0, false
	}

	delay, err := strconv.Atoi(delayStr)
	if err != nil {
		return 0, false
	}

	return time.Duration(delay) * time.Second, true
}

// Host returns the configured host value
func (c *Client) Host() string {
	return c.host
}

// SetHost sets the host value
func (c *Client) SetHost(host string) {
	c.host = host
}

// contains checks if a string is present in a slice
func shouldSkipDebugLog(endpoint string) bool {
	// A list of API endpoints of which the requests/response won't be written into slack-debug-[date].log
	var skipDebuglogEndpoints = []string{
		"apps.activities.list",
		"apps.hosted.exchangeAuthTicket",
		"apps.connections.open",
	}
	// Sanitize endpoint to remove potential parameters
	sanitizedEndpoint := endpoint
	if idx := strings.IndexByte(endpoint, '?'); idx >= 0 {
		sanitizedEndpoint = endpoint[:idx]
	}
	return goutils.Contains(skipDebuglogEndpoints, sanitizedEndpoint, false /*caseSensitive*/)
}
