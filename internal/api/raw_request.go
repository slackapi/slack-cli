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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/uber/jaeger-client-go"
)

// RawResponse holds the full HTTP response from a generic API call.
type RawResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// RawRequest makes a generic HTTP request to a Slack API endpoint and returns the
// full response without interpreting the body. Unlike the typed methods on Client,
// this does not error on non-200 status codes — it always returns the response body.
// Retries are still performed for 429/503 with Retry-After headers.
func (c *Client) RawRequest(ctx context.Context, httpMethod, endpoint, token string, body io.Reader, contentType string, headers map[string]string) (*RawResponse, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "RawRequest")
	defer span.Finish()

	sURL, err := url.Parse(fmt.Sprintf("%s/api/%s", c.host, endpoint))
	if err != nil {
		return nil, err
	}
	span.SetTag("request_url", sURL.String())

	request, err := http.NewRequest(httpMethod, sURL.String(), body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	cliVersion, err := slackcontext.Version(ctx)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS))

	if jaegerSpanContext, ok := span.Context().(jaeger.SpanContext); ok {
		request.Header.Set("x-b3-sampled", "0")
		request.Header.Set("x-b3-spanid", jaegerSpanContext.SpanID().String())
		request.Header.Set("x-b3-traceid", jaegerSpanContext.TraceID().String())
		request.Header.Set("x-b3-parentspanid", jaegerSpanContext.ParentID().String())
	}

	for k, v := range headers {
		request.Header.Set(k, v)
	}

	c.printRequest(ctx, request, false)

	var data []byte
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

		delay, ok := getRetryAfter(r)
		if !ok {
			break
		}
		r.Body.Close()
		c.io.PrintDebug(ctx, "%s responded with status %d. Retrying request in %s...", sURL.Path, r.StatusCode, delay)
		time.Sleep(delay)
	}
	defer r.Body.Close()

	c.printResponse(ctx, r, false)
	span.SetTag("status_code", r.StatusCode)

	respBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return &RawResponse{
		StatusCode: r.StatusCode,
		Header:     r.Header,
		Body:       respBody,
	}, nil
}
