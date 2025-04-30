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
	"crypto/tls"
	"net/http"
	"time"
)

const (
	defaultTotalTimeout = 30 * time.Second
)

// Options allows a user of this lib to customize their http Client.
type HTTPClientOptions struct {
	TotalTimeOut     time.Duration // total time for the life of the request.
	SkipTLSVerify    bool
	Retries          int            // how many times we should retry, if 0 no retry logic is added.
	Backoff          *time.Duration // constant duration to wait between requests.
	AttemptTimeout   *time.Duration // how long we give each attempt before we cancel the request.
	RetryErrorCodes  []string       // which slack application error codes we want to retry on.
	RetryStatusCodes []int          // which http status codes we want to retry on.
}

// New returns an http.Client based on a given Options input.
func NewHTTPClient(opts HTTPClientOptions) *http.Client {
	var client http.Client

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

	// if retries aren't specified return a non-retrying transport early.
	if opts.Retries == 0 {
		client.Transport = &transport
		return &client
	}

	client.Transport = &transport

	return &client
}
