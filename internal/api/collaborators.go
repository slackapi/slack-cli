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
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	collaboratorsAddMethod    = "developer.apps.owners.add"
	collaboratorsListMethod   = "developer.apps.owners.list"
	collaboratorsRemoveMethod = "developer.apps.owners.remove"
	collaboratorsUpdateMethod = "developer.apps.owners.update"
)

type CollaboratorsClient interface {
	AddCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) error
	ListCollaborators(ctx context.Context, token, appID string) ([]types.SlackUser, error)
	RemoveCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) (slackerror.Warnings, error)
	UpdateCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) error
}

// AddCollaborator adds an app collaborator
func (c *Client) AddCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AddCollaborator")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("app_id", appID)
	values.Add("permission_type", string(slackUser.PermissionType))
	if slackUser.Email != "" {
		values.Add("user_email", slackUser.Email)
	} else if slackUser.ID != "" {
		values.Add("user_id", slackUser.ID)
	}

	b, err := c.postForm(ctx, collaboratorsAddMethod, values)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := extendedBaseResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(collaboratorsAddMethod)
	}

	if !resp.Ok {
		return slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, collaboratorsAddMethod)
	}

	return nil
}

type listCollaboratorsResponse struct {
	extendedBaseResponse
	Owners []types.SlackUser `json:"owners"`
}

// ListCollaborators lists app collaborators
func (c *Client) ListCollaborators(ctx context.Context, token, appID string) ([]types.SlackUser, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ListCollaborators")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("app_id", appID)

	b, err := c.postForm(ctx, collaboratorsListMethod, values)
	if err != nil {
		return []types.SlackUser{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return []types.SlackUser{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := listCollaboratorsResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return []types.SlackUser{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(collaboratorsListMethod)
	}

	if !resp.Ok {
		return []types.SlackUser{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, collaboratorsListMethod)
	}

	return resp.Owners, nil
}

// RemoveCollaborator removes an app collaborator
func (c *Client) RemoveCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) (slackerror.Warnings, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.RemoveCollaborator")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("app_id", appID)
	if slackUser.Email != "" {
		values.Add("user_email", slackUser.Email)
	} else if slackUser.ID != "" {
		values.Add("user_id", slackUser.ID)
	}

	b, err := c.postForm(ctx, collaboratorsRemoveMethod, values)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return nil, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := extendedBaseResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return resp.Warnings, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(collaboratorsRemoveMethod)
	}

	if !resp.Ok {
		return resp.Warnings, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, collaboratorsRemoveMethod)
	}

	return resp.Warnings, nil
}

// UpdateCollaborator updates an app collaborator
func (c *Client) UpdateCollaborator(ctx context.Context, token, appID string, slackUser types.SlackUser) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.UpdateCollaborator")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("app_id", appID)
	values.Add("permission_type", string(slackUser.PermissionType))
	if slackUser.Email != "" {
		values.Add("user_email", slackUser.Email)
	} else if slackUser.ID != "" {
		values.Add("user_id", slackUser.ID)
	}

	b, err := c.postForm(ctx, collaboratorsUpdateMethod, values)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := extendedBaseResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(collaboratorsUpdateMethod)
	}

	if !resp.Ok {
		return slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, collaboratorsUpdateMethod)
	}

	return nil
}
