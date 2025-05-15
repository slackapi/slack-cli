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
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	generateAuthTicketMethod = "apps.hosted.generateAuthTicket"
	exchangeAuthTicketMethod = "apps.hosted.exchangeAuthTicket"
	rotateTokenMethod        = "tooling.tokens.rotate"
)

type AuthClient interface {
	ExchangeAuthTicket(ctx context.Context, ticket string, challenge string, cliVersion string) (ExchangeAuthTicketResult, error)
	GenerateAuthTicket(ctx context.Context, cliVersion string, serviceTokenFlag bool) (GenerateAuthTicketResult, error)
	RotateToken(ctx context.Context, auth types.SlackAuth) (RotateTokenResult, error)
}

// ExchangeAuthTicketResult details to be returned
type ExchangeAuthTicketResult struct {
	EnterpriseID        string `json:"enterprise_id,omitempty"`
	ExpiresAt           int    `json:"exp,omitempty"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install,omitempty"`
	IsReady             bool   `json:"is_ready,omitempty"`
	RefreshToken        string `json:"refresh_token,omitempty"`
	TeamDomain          string `json:"team_domain"`
	TeamID              string `json:"team_id"`
	TeamName            string `json:"team_name"`
	Token               string `json:"token,omitempty"`
	UserID              string `json:"user_id"`
}

// GenerateAuthTicketResult details to be returned
type GenerateAuthTicketResult struct {
	Ticket string `json:"ticket,omitempty"`
}

type generateAuthTicketResponse struct {
	extendedBaseResponse
	GenerateAuthTicketResult
}

type exchangeAuthTicketMethodResponse struct {
	extendedBaseResponse
	ExchangeAuthTicketResult
}

type RotateTokenResult struct {
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TeamID       string `json:"team_id"`
	UserID       string `json:"user_id"`
	IssuedAt     int    `json:"iat"`
	ExpiresAt    int    `json:"exp"`
}

type rotateTokenMethodResponse struct {
	extendedBaseResponse
	RotateTokenResult
}

// ExchangeAuthTicket
func (c *Client) ExchangeAuthTicket(ctx context.Context, ticket string, challenge string, cliVersion string) (ExchangeAuthTicketResult, error) {
	var values = url.Values{}
	values.Add("ticket", ticket)
	values.Add("challenge", challenge)
	values.Add("slack_cli_version", cliVersion)

	b, err := c.postForm(ctx, exchangeAuthTicketMethod, values)
	if err != nil {
		return ExchangeAuthTicketResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return ExchangeAuthTicketResult{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := exchangeAuthTicketMethodResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return ExchangeAuthTicketResult{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(exchangeAuthTicketMethod)
	}

	if !resp.Ok {
		return ExchangeAuthTicketResult{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, exchangeAuthTicketMethod)
	}

	// ExchangeAuthTicketResult must have a token to be valid
	if resp.Token == "" {
		return ExchangeAuthTicketResult{}, slackerror.New(fmt.Sprintf("No token returned from the following Slack API method %s. Login can not be completed.", exchangeAuthTicketMethod))
	}

	return resp.ExchangeAuthTicketResult, nil
}

// GenerateAuthTicket
func (c *Client) GenerateAuthTicket(ctx context.Context, cliVersion string, serviceTokenFlag bool) (GenerateAuthTicketResult, error) {
	var values = url.Values{}
	values.Add("slack_cli_version", cliVersion)
	if serviceTokenFlag {
		values.Add("no_rotation", strconv.FormatBool(serviceTokenFlag))
	}

	b, err := c.postForm(ctx, generateAuthTicketMethod, values)
	if err != nil {
		return GenerateAuthTicketResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return GenerateAuthTicketResult{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := generateAuthTicketResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return GenerateAuthTicketResult{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(generateAuthTicketMethod)
	}

	if !resp.Ok {
		return GenerateAuthTicketResult{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, generateAuthTicketMethod)
	}

	return resp.GenerateAuthTicketResult, nil
}

// RotateToken calls tooling.tokens.rotate
func (c *Client) RotateToken(ctx context.Context, auth types.SlackAuth) (RotateTokenResult, error) {
	if c.host == "" {
		return RotateTokenResult{}, slackerror.New("api host not found")
	}

	// Avoid waiting on slow starts from ephemeral authentication hosts
	if strings.Contains(c.host, "dev") {
		timeout := c.httpClient.Timeout
		c.httpClient.Timeout = 2 * time.Second
		defer func() {
			c.httpClient.Timeout = timeout
		}()
	}

	if auth.RefreshToken == "" {
		return RotateTokenResult{}, slackerror.New("refresh token is empty")
	}

	var values = url.Values{}
	values.Add("refresh_token", auth.RefreshToken)

	b, err := c.postForm(ctx, rotateTokenMethod, values)
	if err != nil {
		return RotateTokenResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return RotateTokenResult{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := rotateTokenMethodResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return RotateTokenResult{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(rotateTokenMethod)
	}

	if !resp.Ok {
		return RotateTokenResult{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, rotateTokenMethod)
	}

	return resp.RotateTokenResult, nil
}
