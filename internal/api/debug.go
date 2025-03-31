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
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/slackapi/slack-cli/internal/goutils"
)

// printRequest will print the request to the verbose log output.
func (c *Client) printRequest(ctx context.Context, req *http.Request, skipDebugLog bool) {
	if req == nil {
		return
	}

	// Read and safely restore the body buffer
	reqBody, err := readRequestBody(req)
	if err != nil {
		return
	}

	// Build the response to print
	var outputLines []string
	var output string
	// HTTP Request
	outputLines = append(outputLines, fmt.Sprintf("HTTP Request: %v %v %v", req.Method, req.URL, req.Proto))
	// HTTP Body
	outputLines = append(outputLines, fmt.Sprintf("HTTP Request Body:\n%s", reqBody))
	output = strings.Join(outputLines, "\n")
	output = goutils.RedactPII(output)
	if !skipDebugLog {
		c.io.PrintDebug(ctx, output)
	}
}

// printResponse will print the response to the verbose log output.
func (c *Client) printResponse(ctx context.Context, resp *http.Response, skipDebugLog bool) {
	if resp == nil {
		return
	}

	// Read and safely restore the body buffer
	respBody, err := readResponseBody(resp)
	if err != nil {
		return
	}

	// Build the response to print
	var outputLines []string
	var output string
	// HTTP status
	outputLines = append(outputLines, fmt.Sprintf("HTTP Response Status: %v %v", resp.Proto, resp.Status))
	// HTTP body
	outputLines = append(outputLines, fmt.Sprintf("HTTP Response Body:\n%s", respBody))
	output = strings.Join(outputLines, "\n")
	output = goutils.RedactPII(output)

	if !skipDebugLog {
		c.io.PrintDebug(ctx, output)
	}
}

// readRequestBody will read and restore the request's body buffer.
func readRequestBody(req *http.Request) (string, error) {
	if req == nil {
		return "", nil
	}

	var err error
	var b bytes.Buffer
	savedBody := req.Body

	if req.Body != nil {
		savedBody, req.Body, err = drainBody(req.Body)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(&b, req.Body)
		if err != nil {
			return "", err
		}
	}

	if b.Len() <= 0 {
		b.WriteString("<no body>")
	}

	// Restore the original body buffer
	req.Body = savedBody

	return b.String(), nil
}

// readResponseBody will read response's body buffer and restore the body / content length.
func readResponseBody(resp *http.Response) (string, error) {
	if resp == nil {
		return "", nil
	}

	var err error
	var b bytes.Buffer
	savedBody := resp.Body
	savedContentLength := resp.ContentLength

	if resp.Body != nil {
		savedBody, resp.Body, err = drainBody(resp.Body)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(&b, resp.Body)
		if err != nil {
			return "", err
		}
	}

	if b.Len() <= 0 {
		b.WriteString("<no body>")
	}

	// Restore the original body buffer and content length
	resp.Body = savedBody
	resp.ContentLength = savedContentLength

	return b.String(), nil
}

// drainBody will return the body buffer (b) and a duplicate buffer that
// can be used to restore the original body buffer.
//
// Copied directly from: https://go.dev/src/net/http/httputil/dump.go
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
