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
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateSandbox_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxCreateMethod,
		Response:       `{"ok":true,"team_id":"T123","user_id":"U123","url":"https://my-sandbox.slack.com"}`,
	})
	defer teardown()
	teamID, sandboxURL, err := c.CreateSandbox(ctx, "token", "My Sandbox", "my-sandbox", "secret", "", "", 0, "", 0, false, false)
	require.NoError(t, err)
	require.Equal(t, "T123", teamID)
	require.Equal(t, "https://my-sandbox.slack.com", sandboxURL)
}

func TestClient_CreateSandbox_WithOptionalParams(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxCreateMethod,
		Response:       `{"ok":true,"team_id":"T456","user_id":"U456","url":"https://other-sandbox.slack.com"}`,
	})
	defer teardown()
	teamID, sandboxURL, err := c.CreateSandbox(ctx, "token", "My Sandbox", "my-sandbox", "secret", "en-US", "O123", 1, "EVENT123", 1234567890, false, false)
	require.NoError(t, err)
	require.Equal(t, "T456", teamID)
	require.Equal(t, "https://other-sandbox.slack.com", sandboxURL)
}

func TestClient_CreateSandbox_PasswordFree(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxCreateMethod,
		Response:       `{"ok":true,"team_id":"T789","user_id":"U789","url":"https://pwfree-sandbox.slack.com"}`,
	})
	defer teardown()
	teamID, sandboxURL, err := c.CreateSandbox(ctx, "token", "PW Free Sandbox", "pwfree-sandbox", "", "", "", 0, "", 0, false, true)
	require.NoError(t, err)
	require.Equal(t, "T789", teamID)
	require.Equal(t, "https://pwfree-sandbox.slack.com", sandboxURL)
}

func TestClient_CreateSandbox_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxCreateMethod, func(c *Client) error {
		_, _, err := c.CreateSandbox(ctx, "token", "name", "domain", "password", "", "", 0, "", 0, false, false)
		return err
	})
}

func TestClient_DeleteSandbox_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxDeleteMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	err := c.DeleteSandbox(ctx, "token", "T123")
	require.NoError(t, err)
}

func TestClient_DeleteSandbox_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxDeleteMethod, func(c *Client) error {
		return c.DeleteSandbox(ctx, "token", "T123")
	})
}

func TestClient_ListSandboxes_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true,"sandboxes":[{"sandbox_team_id":"T1","sandbox_name":"Sandbox 1","sandbox_domain":"sb1","status":"active","date_created":123,"date_archived":0},{"sandbox_team_id":"T2","sandbox_name":"Sandbox 2","sandbox_domain":"sb2","status":"active","date_created":456,"date_archived":0}]}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "")
	require.NoError(t, err)
	require.Len(t, sandboxes, 2)
	require.Equal(t, "T1", sandboxes[0].TeamID)
	require.Equal(t, "Sandbox 1", sandboxes[0].Name)
	require.Equal(t, "sb1", sandboxes[0].Domain)
	require.Equal(t, "active", sandboxes[0].Status)
	require.Equal(t, "T2", sandboxes[1].TeamID)
}

func TestClient_ListSandboxes_Empty(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true,"sandboxes":[]}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "")
	require.NoError(t, err)
	require.Empty(t, sandboxes)
}

func TestClient_ListSandboxes_NilSandboxesReturnsEmptySlice(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "")
	require.NoError(t, err)
	require.Equal(t, []types.Sandbox{}, sandboxes)
}

func TestClient_ListSandboxes_WithStatusFilter(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: sandboxListMethod,
		Response:       `{"ok":true,"sandboxes":[{"sandbox_team_id":"T1","sandbox_name":"Archived","sandbox_domain":"arch","status":"archived","date_created":100,"date_archived":200}]}`,
	})
	defer teardown()
	sandboxes, err := c.ListSandboxes(ctx, "token", "archived")
	require.NoError(t, err)
	require.Len(t, sandboxes, 1)
	require.Equal(t, "archived", sandboxes[0].Status)
}

func TestClient_SandboxLogin_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  sandboxLoginMethod,
		ExpectedRequest: "sandbox_team_id=E123&token=token",
		Response:        `{"ok":true,"sandbox_team_id":"E123","sandbox_name":"Sandbox","sandbox_domain":"sb","sandbox_user_id":"U123","access_token":"xoxe.xoxp-access","refresh_token":"xoxe-refresh","exp":1700000000,"type":"user"}`,
	})
	defer teardown()
	result, err := c.SandboxLogin(ctx, "token", "E123")
	require.NoError(t, err)
	require.Equal(t, &SandboxLoginResult{
		SandboxTeamID: "E123",
		SandboxName:   "Sandbox",
		SandboxDomain: "sb",
		SandboxUserID: "U123",
		AccessToken:   "xoxe.xoxp-access",
		RefreshToken:  "xoxe-refresh",
		ExpiresAt:     1700000000,
		Type:          "user",
	}, result)
}

func TestClient_SandboxLogin_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxLoginMethod, func(c *Client) error {
		_, err := c.SandboxLogin(ctx, "token", "E123")
		return err
	})
}

func Test_Client_Signup(t *testing.T) {
	tests := map[string]struct {
		request          SignupRequest
		expectedRequest  string
		response         string
		expectedResponse *SignupResponse
		expectedError    string
	}{
		"omits the terms version when not accepted in the request": {
			request:          SignupRequest{FirstName: "Ada", LastName: "Lovelace", CountryOption: "US"},
			expectedRequest:  "country_option=US&first_name=Ada&last_name=Lovelace&token=token",
			response:         `{"ok":true,"developer_id":"D123"}`,
			expectedResponse: &SignupResponse{DeveloperID: "D123"},
		},
		"sends the accepted terms version": {
			request:          SignupRequest{FirstName: "Ada", LastName: "Lovelace", AcceptTermsOfServiceVersion: "may2024"},
			expectedRequest:  "accept_tos_version=may2024&first_name=Ada&last_name=Lovelace&token=token",
			response:         `{"ok":true,"developer_id":"D123"}`,
			expectedResponse: &SignupResponse{DeveloperID: "D123"},
		},
		"returns the current terms with the terms not accepted error": {
			request:         SignupRequest{FirstName: "Ada", LastName: "Lovelace"},
			expectedRequest: "first_name=Ada&last_name=Lovelace&token=token",
			response: `{
				"ok": false,
				"error": "tos_not_accepted",
				"details": "Terms of service for the Slack Developer program have not been accepted. By signing up, you agree to the terms and acknowledge the privacy policies. Retry the request with accept_tos_version set to the latest version.",
				"tos_version": "may2024",
				"terms_and_policies": [
					{"name": "Developer Program Agreement", "url": "https://slack.com/terms-of-service/slack-developer-program"},
					{"name": "Slack API Terms of Service", "url": "https://slack.com/terms-of-service/api"},
					{"name": "Slack Privacy Policy", "url": "https://slack.com/trust/privacy/privacy-policy"},
					{"name": "Salesforce Privacy Statement", "url": "https://www.salesforce.com/company/legal/privacy/"}
				]
			}`,
			expectedResponse: &SignupResponse{
				TermsOfServiceVersion: "may2024",
				TermsAndPolicies: []TermsAndPolicy{
					{Name: "Developer Program Agreement", URL: "https://slack.com/terms-of-service/slack-developer-program"},
					{Name: "Slack API Terms of Service", URL: "https://slack.com/terms-of-service/api"},
					{Name: "Slack Privacy Policy", URL: "https://slack.com/trust/privacy/privacy-policy"},
					{Name: "Salesforce Privacy Statement", URL: "https://www.salesforce.com/company/legal/privacy/"},
				},
			},
			expectedError: slackerror.ErrTosNotAccepted,
		},
		"returns an empty response with the terms not accepted error": {
			request:          SignupRequest{FirstName: "Ada", LastName: "Lovelace"},
			expectedRequest:  "first_name=Ada&last_name=Lovelace&token=token",
			response:         `{"ok":false,"error":"tos_not_accepted"}`,
			expectedResponse: &SignupResponse{},
			expectedError:    slackerror.ErrTosNotAccepted,
		},
		"returns no response with other errors": {
			request:         SignupRequest{FirstName: "Ada", LastName: "Lovelace"},
			expectedRequest: "first_name=Ada&last_name=Lovelace&token=token",
			response:        `{"ok":false,"error":"email_already_in_use"}`,
			expectedError:   slackerror.ErrEmailAlreadyInUse,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			c, teardown := NewFakeClient(t, FakeClientParams{
				ExpectedMethod:  sandboxSignupVerifiedMethod,
				ExpectedRequest: tc.expectedRequest,
				Response:        tc.response,
			})
			defer teardown()
			resp, err := c.Signup(ctx, "token", tc.request)
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResponse, resp)
		})
	}
}

func TestClient_Signup_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxSignupVerifiedMethod, func(c *Client) error {
		_, err := c.Signup(ctx, "token", SignupRequest{})
		return err
	})
}

func TestClient_ListSandboxes_CommonErrors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, sandboxListMethod, func(c *Client) error {
		_, err := c.ListSandboxes(ctx, "token", "")
		return err
	})
}
