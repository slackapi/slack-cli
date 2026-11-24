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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func Test_Client_DoWithRetry(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	t.Cleanup(func() { otel.SetTracerProvider(previous) })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"ok":false,"error":"invalid_auth"}`)
	}))
	defer server.Close()
	client := NewClient(nil, server.URL, newTestIO())
	_, _ = client.ChannelsInfo(slackcontext.MockContext(t.Context()), "xoxb", "C1")
	var postForm sdktrace.ReadOnlySpan
	for _, span := range recorder.Ended() {
		if span.Name() == "postForm" {
			postForm = span
		}
	}
	require.NotNil(t, postForm)
	assert.Contains(t, postForm.Attributes(), attribute.String("error", "invalid_auth"))
	assert.Contains(t, postForm.Attributes(), attribute.Bool("ok", false))
}

func Test_Client_Host(t *testing.T) {
	tests := map[string]struct {
		host     string
		expected string
	}{
		"returns the configured host": {
			host:     "https://slack.com",
			expected: "https://slack.com",
		},
		"returns empty when unset": {
			host:     "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{host: tc.host}
			assert.Equal(t, tc.expected, c.Host())
		})
	}
}

func Test_Client_SetHost(t *testing.T) {
	tests := map[string]struct {
		initial string
		newHost string
	}{
		"sets a new host": {
			initial: "",
			newHost: "https://dev.slack.com",
		},
		"overwrites existing host": {
			initial: "https://slack.com",
			newHost: "https://dev.slack.com",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{host: tc.initial}
			c.SetHost(tc.newHost)
			assert.Equal(t, tc.newHost, c.Host())
		})
	}
}
