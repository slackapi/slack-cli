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
	"fmt"
	"net/url"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	sandboxCreateMethod = "enterprise.signup.createDevOrg"
	sandboxListMethod   = "developer.sandbox.list"
	sandboxDeleteMethod = "developer.sandbox.delete"
)

// SandboxClient is the interface for sandbox-related API calls
type SandboxClient interface {
	CreateSandbox(ctx context.Context, token string, name, domain, password, locale, owningOrgID, templateID, eventCode string, archiveDate int64) (types.CreateSandboxResult, error)
	ListSandboxes(ctx context.Context, token string, filter string) ([]types.Sandbox, error)
	DeleteSandbox(ctx context.Context, token string, sandboxTeamID string) error
}

type createSandboxResponse struct {
	extendedBaseResponse
	types.CreateSandboxResult
}

type listSandboxesResponse struct {
	extendedBaseResponse
	Sandboxes []types.Sandbox `json:"sandboxes"`
}

var listSandboxFilterEnum = []string{"active", "archived"}

// CreateSandbox provisions a new Developer Sandbox (developer org and primary user).
func (c *Client) CreateSandbox(ctx context.Context, token string, name, domain, password, locale, owningOrgID, templateID, eventCode string, archiveDate int64) (types.CreateSandboxResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.CreateSandbox")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("org_name", name)
	values.Add("domain", domain)
	values.Add("password", password)
	if locale != "" {
		values.Add("locale", locale)
	}
	if owningOrgID != "" {
		values.Add("owning_org_id", owningOrgID)
	}
	if templateID != "" {
		values.Add("template_id", templateID)
	}
	if eventCode != "" {
		values.Add("event_code", eventCode)
	}
	if archiveDate > 0 {
		values.Add("archive_date", strconv.FormatInt(archiveDate, 10))
	}

	b, err := c.postForm(ctx, sandboxCreateMethod, values)
	if err != nil {
		return types.CreateSandboxResult{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := createSandboxResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return types.CreateSandboxResult{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(sandboxCreateMethod)
	}

	if !resp.Ok {
		return types.CreateSandboxResult{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, sandboxCreateMethod)
	}

	return resp.CreateSandboxResult, nil
}

// ListSandboxes returns all sandboxes owned by the Developer Account with an email address that matches the authenticated user
func (c *Client) ListSandboxes(ctx context.Context, token string, status string) ([]types.Sandbox, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ListSandboxes")
	defer span.Finish()

	if status != "" {
		valid := false
		for _, v := range listSandboxFilterEnum {
			if status == v {
				valid = true
				break
			}
		}
		if !valid {
			return nil, errInvalidArguments.WithRootCause(fmt.Errorf("allowed values for sandbox status filter: %v", listSandboxFilterEnum))
		}
	}

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

// DeleteSandbox permanently deletes the specified sandbox.
// Required: token, sandbox_team_id
func (c *Client) DeleteSandbox(ctx context.Context, token string, sandboxTeamID string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.DeleteSandbox")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("sandbox_team_id", sandboxTeamID)

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
