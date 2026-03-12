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

package deputil

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_URLChecker(t *testing.T) {
	tests := map[string]struct {
		url                 string
		expectedURL         string
		setupHTTPClientMock func(*slackhttp.HTTPClientMock)
	}{
		"Returns the URL when the HTTP status code is http.StatusOK": {
			url:         "https://github.com/slack-samples/deno-starter-template",
			expectedURL: "https://github.com/slack-samples/deno-starter-template",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				resOK := slackhttp.MockHTTPResponse(http.StatusOK, "OK")
				httpClientMock.On("Head", mock.Anything).Return(resOK, nil)
			},
		},
		"Returns an empty string when the HTTP status code is not 200": {
			url:         "https://github.com/slack-samples/template-not-found",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				resNotFound := slackhttp.MockHTTPResponse(http.StatusNotFound, "Not Found")
				httpClientMock.On("Head", mock.Anything).Return(resNotFound, nil)
			},
		},
		"Returns an empty string when the HTTPClient has an error": {
			url:         "invalid_url",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				httpClientMock.On("Head", mock.Anything).Return(nil, fmt.Errorf("HTTPClient error"))
			},
		},
		"Returns an empty string for HTTP 500 Internal Server Error": {
			url:         "https://example.com/server-error",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusInternalServerError, "Internal Server Error")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
		"Returns an empty string for HTTP 301 redirect": {
			url:         "https://example.com/redirect",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusMovedPermanently, "Moved")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
		"Returns an empty string for HTTP 403 Forbidden": {
			url:         "https://example.com/forbidden",
			expectedURL: "",
			setupHTTPClientMock: func(httpClientMock *slackhttp.HTTPClientMock) {
				res := slackhttp.MockHTTPResponse(http.StatusForbidden, "Forbidden")
				httpClientMock.On("Head", mock.Anything).Return(res, nil)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mocks
			httpClientMock := &slackhttp.HTTPClientMock{}
			tc.setupHTTPClientMock(httpClientMock)

			// Execute
			url := URLChecker(httpClientMock, tc.url)

			// Assertions
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}
