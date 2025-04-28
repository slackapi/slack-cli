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
	"context"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	sessionValidateMethod = "auth.test"
	revokeTokenMethod     = "auth.revoke"
)

type SessionsClient interface {
	ValidateSession(ctx context.Context, token string) (AuthSession, error)
	RevokeToken(ctx context.Context, token string) error
}

type authCheckResponse struct {
	extendedBaseResponse
	AuthSession
}

// ValidateSession ensures that a given token has a valid session with the Slack API.
// cookie only required if token is a xoxc token otherwise "" is okay
func (c *Client) ValidateSession(ctx context.Context, token string) (AuthSession, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ValidateSesson")
	defer span.Finish()

	b, err := c.postForm(ctx, sessionValidateMethod, url.Values{"token": []string{token}})
	if err != nil {
		return AuthSession{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return AuthSession{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	var authResp authCheckResponse
	err = goutils.JsonUnmarshal(b, &authResp)
	if err != nil {
		return AuthSession{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(sessionValidateMethod)
	}

	if !authResp.Ok {
		return AuthSession{}, slackerror.NewApiError(authResp.Error, authResp.Description, authResp.Errors, sessionValidateMethod)
	}

	if authResp.UserID == nil {
		return AuthSession{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("invalid user_id"))
	}

	return authResp.AuthSession, nil
}

// No guaranteed properties in API response so omitempty is used
type AuthSession struct {
	UserName            *string `json:"user,omitempty"`
	UserID              *string `json:"user_id,omitempty"`
	TeamID              *string `json:"team_id,omitempty"`
	TeamName            *string `json:"team,omitempty"`
	EnterpriseID        *string `json:"enterprise_id,omitempty"`
	IsEnterpriseInstall *bool   `json:"is_enterprise_install,omitempty"`
	URL                 *string `json:"url,omitempty"`
}

type authRevokeResponse struct {
	extendedBaseResponse
	IsRevoked bool `json:"revoked,omitempty"`
}

// RevokeToken sends a given token to the Slack API for revocation
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.RevokeToken")
	defer span.Finish()

	b, err := c.postForm(ctx, revokeTokenMethod, url.Values{"token": []string{token}})
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	var revokeResp authRevokeResponse
	err = goutils.JsonUnmarshal(b, &revokeResp)
	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(revokeTokenMethod)
	}

	if !revokeResp.Ok {
		return slackerror.NewApiError(revokeResp.Error, revokeResp.Description, revokeResp.Errors, revokeTokenMethod)
	}

	return nil
}
