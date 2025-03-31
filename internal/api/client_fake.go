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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type FakeClientParams struct {
	// The API method we expect to call, like "apps.manifest.create"
	ExpectedMethod string
	// The JSON payload we send to the server. If empty, this assertion is skipped.
	ExpectedRequest string
	// Any querystring parameters we expect the client to append to the URL
	ExpectedQuerystring string
	// The response sent back from the server to the client.
	Response string
	// If > 0, the request terminates with this status code and no response body.
	StatusCode int
}

// NewFakeClient returns a *Client that is wired up to expect the specific request to a particular method, and if
// it receives it, returns a predefined response.
func NewFakeClient(t *testing.T, params FakeClientParams) (*Client, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.Path, params.ExpectedMethod)
		if params.ExpectedQuerystring != "" {
			require.Contains(t, r.URL.RawQuery, params.ExpectedQuerystring)
		}
		if params.StatusCode > 0 {
			w.WriteHeader(params.StatusCode)
			return
		}
		if params.ExpectedRequest != "" {
			payload, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, params.ExpectedRequest, string(payload))
		}
		_, err := fmt.Fprintln(w, params.Response)
		require.NoError(t, err)
	}))
	client := NewClient(&http.Client{}, server.URL, nil)
	return client, func() {
		server.Close()
	}
}
