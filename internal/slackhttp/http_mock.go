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

package slackhttp

import (
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/mock"
)

// HTTPClientMock implements a mock for the HTTPClient interface.
type HTTPClientMock struct {
	mock.Mock
}

// Do is a mock that tracks the calls to Do and returns the mocked http.Response and error.
func (m *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	// http.Response can be nil when an error is provided.
	var httpResp *http.Response
	if _httpResp, ok := args.Get(0).(*http.Response); ok {
		httpResp = _httpResp
	}

	return httpResp, args.Error(1)
}

// Get is a mock that tracks the calls to Get and returns the mocked http.Response and error.
func (m *HTTPClientMock) Get(url string) (*http.Response, error) {
	args := m.Called(url)

	// http.Response can be nil when an error is provided.
	var httpResp *http.Response
	if _httpResp, ok := args.Get(0).(*http.Response); ok {
		httpResp = _httpResp
	}

	return httpResp, args.Error(1)
}

// Head is a mock that tracks the calls to Head and returns the mocked http.Response and error.
func (m *HTTPClientMock) Head(url string) (*http.Response, error) {
	args := m.Called(url)

	// http.Response can be nil when an error is provided.
	var httpResp *http.Response
	if _httpResp, ok := args.Get(0).(*http.Response); ok {
		httpResp = _httpResp
	}

	return httpResp, args.Error(1)
}

// MockHTTPResponse is a helper that returns a mocked http.Response with the provided httpStatus and body.
func MockHTTPResponse(httpStatus int, body string) *http.Response {
	resWriter := httptest.NewRecorder()

	resWriter.WriteHeader(httpStatus)
	_, _ = io.WriteString(resWriter, body)
	res := resWriter.Result()

	return res
}
