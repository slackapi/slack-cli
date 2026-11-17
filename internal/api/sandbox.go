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
	"context"
	"net/url"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	sandboxCreateMethod         = "enterprise.signup.createDevOrg"
	sandboxDeleteMethod         = "developer.sandbox.delete"
	sandboxExchangeTokensMethod = "developer.sandbox.exchangeTokens"
	sandboxListMethod           = "developer.sandbox.list"
	sandboxSignupVerifiedMethod = "developer.account.createVerified"
)

// ExchangeTokensResult contains the tokens and team details for a sandbox
type ExchangeTokensResult struct {
	SandboxTeamID string `json:"sandbox_team_id"`
	SandboxName   string `json:"sandbox_name"`
	SandboxDomain string `json:"sandbox_domain"`
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresAt     int    `json:"exp"`
	Type          string `json:"type"`
}

// SignupRequest contains the fields for Developer Program signup
type SignupRequest struct {
	FirstName              string
	LastName               string
	PostalCode             string
	CountryOption          string
	NewsletterSubscription bool
	AcceptTermsOfService   bool
}

// SignupResponse contains the response from Developer Program signup
type SignupResponse struct {
	DeveloperID string `json:"developer_id"`
}

// SandboxClient is the interface for sandbox-related API calls
type SandboxClient interface {
	CreateSandbox(ctx context.Context, token, name, domain, password, locale, owningOrgID string, templateID int, eventCode string, archiveDate int64, isPartner, useEmailLogin bool) (teamID, sandboxURL string, err error)
	DeleteSandbox(ctx context.Context, token, sandboxID string) error
	ExchangeTokens(ctx context.Context, token, sandboxTeamID string) (*ExchangeTokensResult, error)
	ListSandboxes(ctx context.Context, token string, status string) ([]types.Sandbox, error)
	Signup(ctx context.Context, token string, req SignupRequest) (*SignupResponse, error)
}

type exchangeTokensResponse struct {
	extendedBaseResponse
	ExchangeTokensResult
}

type createSandboxResponse struct {
	extendedBaseResponse
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
	URL    string `json:"url"`
}

type listSandboxesResponse struct {
	extendedBaseResponse
	Sandboxes []types.Sandbox `json:"sandboxes"`
}

// CreateSandbox creates a new developer sandbox
func (c *Client) CreateSandbox(ctx context.Context, token, name, domain, password, locale, owningOrgID string, templateID int, eventCode string, archiveDate int64, isPartner, useEmailLogin bool) (teamID, sandboxURL string, err error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.CreateSandbox")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("org_name", name)
	if domain != "" {
		values.Add("domain", domain)
	}
	if useEmailLogin {
		values.Add("is_password_free", "true")
	} else {
		values.Add("password", password)
	}
	if locale != "" {
		values.Add("locale", locale)
	}
	if owningOrgID != "" {
		values.Add("owning_org_id", owningOrgID)
	}
	if templateID != 0 {
		values.Add("template_id", strconv.Itoa(templateID))
	}
	if eventCode != "" {
		values.Add("event_code", eventCode)
	}
	if archiveDate > 0 {
		values.Add("archive_date", strconv.FormatInt(archiveDate, 10))
	}
	if isPartner {
		values.Add("is_partner", "true")
	}

	b, err := c.postForm(ctx, sandboxCreateMethod, values)
	if err != nil {
		return "", "", errHTTPRequestFailed.WithRootCause(err)
	}

	resp := createSandboxResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return "", "", errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(sandboxCreateMethod)
	}

	if !resp.Ok {
		return "", "", slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, sandboxCreateMethod)
	}

	return resp.TeamID, resp.URL, nil
}

// DeleteSandbox permanently deletes a developer sandbox
func (c *Client) DeleteSandbox(ctx context.Context, token, sandboxID string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.DeleteSandbox")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("sandbox_team_id", sandboxID)

	b, err := c.postForm(ctx, sandboxDeleteMethod, values)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	resp := extendedBaseResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(sandboxDeleteMethod)
	}

	if !resp.Ok {
		return slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, sandboxDeleteMethod)
	}

	return nil
}

// ExchangeTokens exchanges the caller's token for access and refresh tokens for a sandbox
func (c *Client) ExchangeTokens(ctx context.Context, token, sandboxTeamID string) (*ExchangeTokensResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ExchangeTokens")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("sandbox_team_id", sandboxTeamID)

	b, err := c.postForm(ctx, sandboxExchangeTokensMethod, values)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := exchangeTokensResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(sandboxExchangeTokensMethod)
	}

	if !resp.Ok {
		return nil, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, sandboxExchangeTokensMethod)
	}

	return &resp.ExchangeTokensResult, nil
}

// ListSandboxes returns all sandboxes owned by the Developer Account with an email address that matches the authenticated user
func (c *Client) ListSandboxes(ctx context.Context, token string, status string) ([]types.Sandbox, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ListSandboxes")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	if status != "" {
		values.Add("status", status)
	}

	b, err := c.postForm(ctx, sandboxListMethod, values)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := listSandboxesResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(sandboxListMethod)
	}

	if !resp.Ok {
		return nil, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, sandboxListMethod)
	}

	if resp.Sandboxes == nil {
		return []types.Sandbox{}, nil
	}

	return resp.Sandboxes, nil
}

type signupAPIResponse struct {
	extendedBaseResponse
	SignupResponse
}

// Signup signs the authenticated user up for the Slack Developer Program.
// The email address of the developer account is derived from the token, so no email verification is needed.
func (c *Client) Signup(ctx context.Context, token string, req SignupRequest) (*SignupResponse, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.Signup")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("first_name", req.FirstName)
	values.Add("last_name", req.LastName)
	if req.PostalCode != "" {
		values.Add("postal_code", req.PostalCode)
	}
	if req.CountryOption != "" {
		values.Add("country_option", req.CountryOption)
	}
	if req.NewsletterSubscription {
		values.Add("newsletter_subscription", "true")
	}

	// Setting this parameter to true indicates the user has agreed to the terms and acknowledges the privacy policies:
	// - Developer Program Agreement: https://slack.com/terms-of-service/slack-developer-program
	// - Slack Privacy Policy: https://slack.com/trust/privacy/privacy-policy
	// - Salesforce Privacy Statement: https://www.salesforce.com/company/privacy/
	values.Add("accept_tos", strconv.FormatBool(req.AcceptTermsOfService))

	b, err := c.postForm(ctx, sandboxSignupVerifiedMethod, values)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := signupAPIResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(sandboxSignupVerifiedMethod)
	}

	if !resp.Ok {
		return nil, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, sandboxSignupVerifiedMethod)
	}

	return &resp.SignupResponse, nil
}
