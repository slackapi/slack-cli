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
	"crypto/tls"
	"net/http"
	"time"
)

const (
	// defaultTotalTimeout is the default timeout for the life of the request.
	defaultTotalTimeout = 30 * time.Second
)

// HTTPClient interface for http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Head(url string) (*http.Response, error)
}

// HTTPClientOptions allows for the customization of a http.Client.
type HTTPClientOptions struct {
	AttemptTimeout   *time.Duration // AttemptTimeout is how long each request waits before cancelled.
	Backoff          *time.Duration // Backoff is a constant duration to wait between requests.
	Retries          int            // Retries for a request, if 0 then no retry logic is added.
	RetryErrorCodes  []string       // RetryErrorCode is list of error codes to retry on.
	RetryStatusCodes []int          // RetryStatusCode is list of http status codes to retry on.
	SkipTLSVerify    bool           // SkipTLSVerify will skip verifying the host's certificate.
	TotalTimeOut     time.Duration  // TotalTimeOut for the life of the request (default: defaultTotalTimeout).
}

// NewHTTPClient returns an http.Client configured with HTTPClientOptions.
func NewHTTPClient(opts HTTPClientOptions) *http.Client {
	client := &http.Client{}

	if opts.TotalTimeOut == 0 {
		client.Timeout = defaultTotalTimeout
	} else {
		client.Timeout = opts.TotalTimeOut
	}

	var transport = http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.SkipTLSVerify,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	// If retries aren't specified return a non-retrying transport.
	if opts.Retries == 0 {
		client.Transport = &transport
		return client
	}

	// TODO: Implement retry logic
	client.Transport = &transport
	return client
}
