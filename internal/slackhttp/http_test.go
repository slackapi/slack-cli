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

package slackhttp

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewHTTPClient(t *testing.T) {
	tests := map[string]struct {
		proxy                 func(*http.Request) (*url.URL, error)
		expectedProxy         func(*http.Request) (*url.URL, error)
		retries               int
		expectedRetries       int
		skipTLSVerify         bool
		expectedSkipTLSVerify bool
		timeout               time.Duration
		expectedTimeout       time.Duration
	}{
		"Returns a httpClient with default configuration": {
			expectedTimeout:       defaultTotalTimeout,
			expectedRetries:       0,
			expectedSkipTLSVerify: false,
			expectedProxy:         nil,
		},
		"Returns a httpClient with custom configuration": {
			proxy:         http.ProxyFromEnvironment,
			expectedProxy: http.ProxyFromEnvironment,

			retries:         3,
			expectedRetries: 3,

			skipTLSVerify:         true,
			expectedSkipTLSVerify: true,

			timeout:         120 * time.Second,
			expectedTimeout: 120 * time.Second,
		},
		"Zero timeout uses default timeout": {
			timeout:               0,
			expectedTimeout:       defaultTotalTimeout,
			expectedSkipTLSVerify: false,
		},
		"Custom timeout is used when non-zero": {
			timeout:               60 * time.Second,
			expectedTimeout:       60 * time.Second,
			expectedSkipTLSVerify: false,
		},
		"SkipTLSVerify false keeps verification enabled": {
			skipTLSVerify:         false,
			expectedSkipTLSVerify: false,
			expectedTimeout:       defaultTotalTimeout,
		},
		"Retries zero returns non-retrying transport": {
			retries:               0,
			expectedRetries:       0,
			expectedTimeout:       defaultTotalTimeout,
			expectedSkipTLSVerify: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			opts := HTTPClientOptions{
				Retries:       tc.expectedRetries,
				SkipTLSVerify: tc.expectedSkipTLSVerify,
				TotalTimeOut:  tc.timeout,
			}

			httpClient := NewHTTPClient(opts)

			// Assertions
			assert.Equal(t, tc.expectedTimeout, httpClient.Timeout)
			assert.Equal(t, tc.expectedSkipTLSVerify, httpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
			assert.ObjectsAreEqual(tc.expectedProxy, httpClient.Transport.(*http.Transport).Proxy)
			// TODO: add assertion for retries
		})
	}
}
