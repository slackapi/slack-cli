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
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	functionDistributionsPermissionsListMethod   = "functions.distributions.permissions.list"
	functionDistributionsPermissionsSetMethod    = "functions.distributions.permissions.set"
	functionDistributionsPermissionsAddMethod    = "functions.distributions.permissions.add"
	functionDistributionsPermissionsRemoveMethod = "functions.distributions.permissions.remove"
)

type FunctionDistributionClient interface {
	FunctionDistributionList(ctx context.Context, callbackID, appID string) (types.Permission, []types.FunctionDistributionUser, error)
	FunctionDistributionSet(ctx context.Context, callbackID, appID string, distributionType types.Permission, users string) ([]types.FunctionDistributionUser, error)
	FunctionDistributionAddUsers(ctx context.Context, callbackID, appID, users string) error
	FunctionDistributionRemoveUsers(ctx context.Context, callbackID, appID, users string) error
}

type FunctionDistributionListResponse struct {
	extendedBaseResponse
	DistributionType string                           `json:"distribution_type"`
	Users            []types.FunctionDistributionUser `json:"users"`
}

// FunctionDistributionList returns the distribution type. If type is 'named_entities', the IDs of entities that have access are also returned.
func (c *Client) FunctionDistributionList(ctx context.Context, callbackID, appID string) (types.Permission, []types.FunctionDistributionUser, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.FunctionDistributionList")
	defer span.Finish()

	token := config.GetContextToken(ctx)

	var values = url.Values{}
	values.Add("token", token)
	values.Add("function_callback_id", callbackID)
	values.Add("function_app_id", appID)

	b, err := c.postForm(ctx, functionDistributionsPermissionsListMethod, values)
	if err != nil {
		return "", []types.FunctionDistributionUser{}, errHttpRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return "", []types.FunctionDistributionUser{}, errHttpResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := FunctionDistributionListResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return "", []types.FunctionDistributionUser{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(functionDistributionsPermissionsListMethod)
	}

	if !resp.Ok {
		return "", []types.FunctionDistributionUser{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, functionDistributionsPermissionsListMethod)
	}

	dist := types.Permission(strings.ToLower(resp.DistributionType))
	if !dist.IsValid() {
		errStr := fmt.Sprintf("unrecognized access type %s", dist)
		return "", []types.FunctionDistributionUser{}, slackerror.New(errStr).AddApiMethod(functionDistributionsPermissionsListMethod)

	}

	return types.Permission(resp.DistributionType), resp.Users, nil
}

type FunctionDistributionSetResponse struct {
	extendedBaseResponse
	DistributionType string                           `json:"distribution_type"`
	Users            []types.FunctionDistributionUser `json:"users"`
}

// FunctionDistributionSet sets the distribution type for the given function.
// One can also pass in a list of users to assign access to if distribution type is named_entities.
func (c *Client) FunctionDistributionSet(ctx context.Context, callbackID, appID string, distributionType types.Permission, users string) ([]types.FunctionDistributionUser, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.TypeFunctionAccess")
	defer span.Finish()

	token := config.GetContextToken(ctx)

	var values = url.Values{}
	values.Add("token", token)
	values.Add("function_callback_id", callbackID)
	values.Add("function_app_id", appID)
	values.Add("distribution_type", string(distributionType))
	if distributionType == types.NAMED_ENTITIES && len(users) > 0 {
		values.Add("user_ids", users)
	}

	b, err := c.postForm(ctx, functionDistributionsPermissionsSetMethod, values)
	if err != nil {
		return []types.FunctionDistributionUser{}, errHttpRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return []types.FunctionDistributionUser{}, errHttpResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := FunctionDistributionSetResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return []types.FunctionDistributionUser{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(functionDistributionsPermissionsSetMethod)
	}

	if !resp.Ok {
		return []types.FunctionDistributionUser{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, functionDistributionsPermissionsSetMethod)
	}

	return resp.Users, nil
}

type FunctionDistributionAddUsersResponse struct {
	extendedBaseResponse
}

// FunctionDistributionAddUsers adds the given entities to the access list for functions with distribution type 'named_entities'.
func (c *Client) FunctionDistributionAddUsers(ctx context.Context, callbackID, appID, users string) error {

	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.FunctionDistributionAddUsers")
	defer span.Finish()

	token := config.GetContextToken(ctx)

	var values = url.Values{}
	values.Add("token", token)
	values.Add("function_callback_id", callbackID)
	values.Add("function_app_id", appID)
	values.Add("user_ids", users)

	b, err := c.postForm(ctx, functionDistributionsPermissionsAddMethod, values)
	if err != nil {
		return errHttpRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHttpResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := FunctionDistributionAddUsersResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHttpResponseInvalid.WithRootCause(err).AddApiMethod(functionDistributionsPermissionsAddMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, functionDistributionsPermissionsAddMethod)
	}

	return nil
}

type FunctionDistributionRemoveUsersResponse struct {
	extendedBaseResponse
}

// FunctionDistributionRemoveUsers removes the given entities from the access list for functions with distribution type 'named_entities'.
func (c *Client) FunctionDistributionRemoveUsers(ctx context.Context, callbackID, appID, users string) error {

	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.FunctionDistributionRemoveUsers")
	defer span.Finish()

	token := config.GetContextToken(ctx)

	var values = url.Values{}
	values.Add("token", token)
	values.Add("function_callback_id", callbackID)
	values.Add("function_app_id", appID)
	values.Add("user_ids", users)

	b, err := c.postForm(ctx, functionDistributionsPermissionsRemoveMethod, values)
	if err != nil {
		return errHttpRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return errHttpResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := FunctionDistributionRemoveUsersResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHttpResponseInvalid.WithRootCause(err).AddApiMethod(functionDistributionsPermissionsRemoveMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, functionDistributionsPermissionsRemoveMethod)
	}

	return nil
}
