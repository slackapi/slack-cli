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
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_GenerateAuthTicket_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  generateAuthTicketMethod,
		ExpectedRequest: ``,
		Response:        `{"ok":true,"ticket":"valid-ticket"}`,
	})
	defer teardown()
	result, err := c.GenerateAuthTicket(ctx, "", false)
	require.NoError(t, err)
	require.Equal(t, "valid-ticket", result.Ticket)
}

func TestClient_GenerateAuthTicket_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  generateAuthTicketMethod,
		ExpectedRequest: ``,
		Response:        `{"ok":false,"error":"fatal_error"}`,
	})
	defer teardown()
	_, err := c.GenerateAuthTicket(ctx, "", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fatal_error")
	require.Contains(t, err.Error(), generateAuthTicketMethod)
}

func TestClient_ExchangeAuthTicket_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  exchangeAuthTicketMethod,
		ExpectedRequest: `challenge=valid-challenge&slack_cli_version=&ticket=valid-ticket`,
		Response:        `{"ok":true,"is_ready":true,"token":"valid-token","refresh_token":"valid-refresh-token"}`,
	})
	defer teardown()
	result, err := c.ExchangeAuthTicket(ctx, "valid-ticket", "valid-challenge", "")
	require.NoError(t, err)
	require.True(t, result.IsReady)
	require.Equal(t, "valid-token", result.Token)
	require.Equal(t, "valid-refresh-token", result.RefreshToken)
}

func TestClient_ExchangeAuthTicket_Ok_MissingToken(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  exchangeAuthTicketMethod,
		ExpectedRequest: `challenge=dummychallenge&slack_cli_version=&ticket=valid-ticket`,
		Response:        `{"ok":true,"token":"","refresh_token":""}`,
	})
	defer teardown()
	_, err := c.ExchangeAuthTicket(ctx, "valid-ticket", "dummychallenge", "")
	require.Error(t, err, "No token returned from the following Slack API method")
}

func TestClient_ExchangeAuthTicket_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  exchangeAuthTicketMethod,
		ExpectedRequest: `challenge=valid-challenge&slack_cli_version=&ticket=valid-ticket`,
		Response:        `{"ok":false,"error":"fatal_error"}`,
	})
	defer teardown()
	result, err := c.ExchangeAuthTicket(ctx, "valid-ticket", "valid-challenge", "")
	require.False(t, result.IsReady)
	require.Error(t, err)
	require.Contains(t, err.Error(), "fatal_error")
	require.Contains(t, err.Error(), exchangeAuthTicketMethod)
}

func TestClient_RotateToken_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  rotateTokenMethod,
		ExpectedRequest: `refresh_token=valid-refresh-token`,
		Response:        `{"ok":true,"token":"rotated-token","refresh_token":"rotated-refreshed-token"}`,
	})
	defer teardown()
	c.httpClient.Timeout = 60 * time.Second
	result, err := c.RotateToken(ctx, types.SlackAuth{
		Token:        `valid-token`,
		RefreshToken: `valid-refresh-token`,
	})
	require.NoError(t, err)
	require.Equal(t, c.httpClient.Timeout, 60*time.Second)
	require.Equal(t, "rotated-token", result.Token)
	require.Equal(t, "rotated-refreshed-token", result.RefreshToken)
}

func TestClient_RotateToken_OkDevHostRestoreTimeout(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  rotateTokenMethod,
		ExpectedRequest: `refresh_token=valid-refresh-token`,
		StatusCode:      500,
	})
	defer teardown()
	devHost := "https://dev1234.slack.com"
	c.host = devHost
	c.httpClient.Timeout = 24 * time.Second
	_, err := c.RotateToken(ctx, types.SlackAuth{
		APIHost:      &devHost,
		Token:        `valid-token`,
		RefreshToken: `valid-refresh-token`,
	})
	require.Equal(t, c.httpClient.Timeout, 24*time.Second)
	require.Error(t, err)
}

func TestClient_RotateToken_Error(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  rotateTokenMethod,
		ExpectedRequest: `refresh_token=valid-refresh-token`,
		Response:        `{"ok":false,"error":"fatal_error"}`,
	})
	defer teardown()
	_, err := c.RotateToken(ctx, types.SlackAuth{
		Token:        `valid-token`,
		RefreshToken: `valid-refresh-token`,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "fatal_error")
	require.Contains(t, err.Error(), rotateTokenMethod)
}
