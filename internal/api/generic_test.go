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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestIO() *iostreams.IOStreamsMock {
	fs := slackdeps.NewFsMock()
	os := slackdeps.NewOsMock()
	cfg := config.NewConfig(fs, os)
	m := iostreams.NewIOStreamsMock(cfg, fs, os)
	m.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	return m
}

func Test_RawRequest_BasicPOST(t *testing.T) {
	var receivedMethod string
	var receivedPath string
	var receivedAuth string
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		receivedAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("X-Test", "value")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	io := newTestIO()
	client := NewClient(nil, server.URL, io)

	resp, err := client.RawRequest(ctx, "POST", "auth.test", "xoxb-token", strings.NewReader("hello=world"), "application/x-www-form-urlencoded", nil)

	assert.NoError(t, err)
	assert.Equal(t, "POST", receivedMethod)
	assert.Equal(t, "/api/auth.test", receivedPath)
	assert.Equal(t, "Bearer xoxb-token", receivedAuth)
	assert.Equal(t, "hello=world", receivedBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "value", resp.Header.Get("X-Test"))
	assert.Equal(t, `{"ok":true}`, string(resp.Body))
}

func Test_RawRequest_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"ok":false,"error":"invalid_auth"}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	io := newTestIO()
	client := NewClient(nil, server.URL, io)

	resp, err := client.RawRequest(ctx, "POST", "auth.test", "bad-token", nil, "", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, `{"ok":false,"error":"invalid_auth"}`, string(resp.Body))
}

func Test_RawRequest_CustomHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	io := newTestIO()
	client := NewClient(nil, server.URL, io)

	headers := map[string]string{"X-Custom": "my-value", "X-Another": "other"}
	resp, err := client.RawRequest(ctx, "GET", "api.test", "", nil, "", headers)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "my-value", receivedHeaders.Get("X-Custom"))
	assert.Equal(t, "other", receivedHeaders.Get("X-Another"))
}

func Test_RawRequest_RetryOnTooManyRequests(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	io := newTestIO()
	client := NewClient(nil, server.URL, io)

	resp, err := client.RawRequest(ctx, "POST", "auth.test", "token", nil, "", nil)

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"ok":true}`, string(resp.Body))
}
