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
	"encoding/json"
	"fmt"
	"net/url"

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
	DeleteSandbox(ctx context.Context, token string, sandboxTeamID, sandboxType string) error
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

	req := types.CreateSandboxRequest{
		Token:       token,
		OrgName:     name,
		Domain:      domain,
		Password:    password,
		Locale:      locale,
		OwningOrgID: owningOrgID,
		TemplateID:  templateID,
		EventCode:   eventCode,
		ArchiveDate: archiveDate,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return types.CreateSandboxResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, sandboxCreateMethod, token, "", body)
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
func (c *Client) ListSandboxes(ctx context.Context, token string, filter string) ([]types.Sandbox, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ListSandboxes")
	defer span.Finish()

	if filter != "" {
		valid := false
		for _, v := range listSandboxFilterEnum {
			if filter == v {
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
	if filter != "" {
		values.Add("filter", filter)
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
func (c *Client) DeleteSandbox(ctx context.Context, token string, sandboxTeamID, sandboxType string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.DeleteSandbox")
	defer span.Finish()

	values := url.Values{}
	values.Add("token", token)
	values.Add("sandbox_team_id", sandboxTeamID)
	if sandboxType != "" {
		values.Add("sandbox_type", sandboxType)
	}

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
