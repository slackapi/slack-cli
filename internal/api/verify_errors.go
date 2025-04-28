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
	"net/http"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/require"
)

// verifyCommonErrorCases provides several invalid or error JSON representations as responses to the HTTP request
// made by the callApi function, and verifies that they are handled correctly. All Client methods that call
// HTTP API endpoints should pass these tests.
func verifyCommonErrorCases(t *testing.T, method string, callApi func(c *Client) error) {
	type commonErrorTestArgs struct {
		Name          string
		Response      string
		ExpectedError string
	}

	cases := []commonErrorTestArgs{
		{
			Name:          "failure",
			Response:      `{"ok":false,"error":"internal_error"}`,
			ExpectedError: "internal_error",
		},
		{
			Name:          "invalid json response error",
			Response:      `{`,
			ExpectedError: slackerror.ErrHTTPResponseInvalid,
		},
		{
			Name:          "response not ok with error and errors",
			Response:      `{"ok": false, "error": "the error", "slack_cli_error_description": "the error", "errors": [{"message": "the errors"}]}`,
			ExpectedError: slackerror.NewApiError("the error", "the error", slackerror.ErrorDetails{{Message: "the errors"}}, method).Error(),
		},
		{
			Name:          "response not ok with error only",
			Response:      `{"ok": false, "error": "bad_error_happened", "slack_cli_error_description": "the error"}`,
			ExpectedError: "bad_error_happened",
		},
		{
			Name:          "response not ok with errors only",
			Response:      `{"ok": false, "slack_cli_error_description": "the error", "errors": [{"message": "an err"}]}`,
			ExpectedError: "an err",
		},
		{
			Name:          "response not ok with missing error & errors renders reasonably well (does not panic)",
			Response:      `{"ok": false}`,
			ExpectedError: "unknown_error",
		},
	}

	for _, args := range cases {
		t.Run(args.Name, func(t *testing.T) {
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod: method,
				Response:       args.Response,
			})
			defer teardown()
			err := callApi(c)

			require.Error(t, err)
			require.Contains(t, err.Error(), args.ExpectedError)
		})
	}

	t.Run("handles internal server error", func(t *testing.T) {
		c, teardown := NewFakeClient(t, FakeClientParams{
			ExpectedMethod: method,
			StatusCode:     http.StatusInternalServerError,
		})
		defer teardown()
		err := callApi(c)
		require.Error(t, err)
		require.Contains(t, err.Error(), slackerror.ErrHttpRequestFailed)
	})
}
