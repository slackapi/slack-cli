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
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	workflowsTriggersPermissionsListMethod   = "workflows.triggers.permissions.list"
	workflowsTriggersPermissionsSetMethod    = "workflows.triggers.permissions.set"
	workflowsTriggersPermissionsAddMethod    = "workflows.triggers.permissions.add"
	workflowsTriggersPermissionsRemoveMethod = "workflows.triggers.permissions.remove"
)

type TriggerAccessClient interface {
	TriggerPermissionsList(ctx context.Context, token, triggerID string) (types.Permission, []string, error)
	TriggerPermissionsSet(ctx context.Context, token, triggerID, entities string, permissionType types.Permission, entityType string) ([]string, error)
	TriggerPermissionsAddEntities(ctx context.Context, token, triggerID, entities string, entityType string) error
	TriggerPermissionsRemoveEntities(ctx context.Context, token, triggerID, entities string, entityType string) error
}

type TriggerPermissionsListResponse struct {
	extendedBaseResponse
	PermissionType string   `json:"permission_type"`
	Users          []string `json:"user_ids"`
	Channels       []string `json:"channel_ids"`
	Teams          []string `json:"team_ids"`
	Organizations  []string `json:"org_ids"`
}

// TriggerPermissionsList returns the access type. If type is 'named_entities', the IDs of entities that have access are also returned.
func (c *Client) TriggerPermissionsList(ctx context.Context, token, triggerID string) (types.Permission, []string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.TriggerPermissionsList")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("trigger_id", triggerID)

	b, err := c.postForm(ctx, workflowsTriggersPermissionsListMethod, values)
	if err != nil {
		return "", []string{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return "", []string{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := TriggerPermissionsListResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return "", []string{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsTriggersPermissionsListMethod)
	}

	if !resp.Ok {
		return "", []string{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsTriggersPermissionsListMethod)
	}

	dist := types.Permission(strings.ToLower(resp.PermissionType))
	if !dist.IsValid() {
		errStr := fmt.Sprintf("unrecognized access type %s", dist)
		return "", []string{}, slackerror.New(errStr).AddAPIMethod(workflowsTriggersPermissionsListMethod)
	}
	entitiesList := []string{}
	namedEntitiesRespList := [][]string{
		resp.Users,
		resp.Channels,
		resp.Teams,
		resp.Organizations,
	}
	for _, l := range namedEntitiesRespList {
		entitiesList = append(entitiesList, l...)
	}

	return types.Permission(resp.PermissionType), entitiesList, nil
}

type TriggerPermissionsSetResponse struct {
	extendedBaseResponse
	PermissionType string   `json:"permission_type"`
	Users          []string `json:"user_ids"`
	Channels       []string `json:"channel_ids"`
	Teams          []string `json:"team_ids"`
	Organizations  []string `json:"org_ids"`
}

// TriggerPermissionsSet sets the access type for the given trigger.
// One can also pass in a list of entities to assign access to if access type is named_entities.
func (c *Client) TriggerPermissionsSet(ctx context.Context, token, triggerID, entities string, permissionType types.Permission, entityType string) ([]string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.TriggerPermissionsSet")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("trigger_id", triggerID)
	values.Add("permission_type", string(permissionType))
	if permissionType == types.NAMED_ENTITIES && len(entities) > 0 {
		if entityType == "users" {
			values.Add("user_ids", entities)
		} else if entityType == "channels" {
			values.Add("channel_ids", entities)
		} else if entityType == "workspaces" {
			values.Add("team_ids", entities)
		} else if entityType == "organizations" {
			values.Add("org_ids", entities)
		}
	}
	b, err := c.postForm(ctx, workflowsTriggersPermissionsSetMethod, values)
	if err != nil {
		return []string{}, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return []string{}, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := TriggerPermissionsSetResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return []string{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsTriggersPermissionsSetMethod)
	}

	if !resp.Ok {
		return []string{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsTriggersPermissionsSetMethod)
	}

	entitiesList := append(resp.Users, resp.Channels...)
	return entitiesList, nil
}

type TriggerPermissionsAddEntitiesResponse struct {
	extendedBaseResponse
}

// TriggerPermissionsAddEntities adds the given entities to the access list for triggers with access type 'named_entities'.
func (c *Client) TriggerPermissionsAddEntities(ctx context.Context, token, triggerID, entities string, entityType string) error {

	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.TriggerPermissionsAddEntities")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("trigger_id", triggerID)
	if entityType == "users" {
		values.Add("user_ids", entities)
	} else if entityType == "channels" {
		values.Add("channel_ids", entities)
	} else if entityType == "workspaces" {
		values.Add("team_ids", entities)
	} else if entityType == "organizations" {
		values.Add("org_ids", entities)
	}

	b, err := c.postForm(ctx, workflowsTriggersPermissionsAddMethod, values)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := TriggerPermissionsAddEntitiesResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsTriggersPermissionsAddMethod)
	}

	if !resp.Ok {
		return slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsTriggersPermissionsAddMethod)
	}

	return nil
}

type TriggerPermissionsRemoveEntitiesResponse struct {
	extendedBaseResponse
}

// TriggerPermissionsRemoveEntities removes the given entities from the access list for triggers with access type 'named_entities'.
func (c *Client) TriggerPermissionsRemoveEntities(ctx context.Context, token, triggerID, entities string, entityType string) error {

	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.TriggerPermissionsRemoveEntities")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("trigger_id", triggerID)
	if entityType == "users" {
		values.Add("user_ids", entities)
	} else if entityType == "channels" {
		values.Add("channel_ids", entities)
	} else if entityType == "workspaces" {
		values.Add("team_ids", entities)
	} else if entityType == "organizations" {
		values.Add("org_ids", entities)
	}

	b, err := c.postForm(ctx, workflowsTriggersPermissionsRemoveMethod, values)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := TriggerPermissionsRemoveEntitiesResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsTriggersPermissionsRemoveMethod)
	}

	if !resp.Ok {
		return slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsTriggersPermissionsRemoveMethod)
	}

	return nil
}
