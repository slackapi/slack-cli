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
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	varListMethod   = "apps.hosted.variables.list"
	varAddMethod    = "apps.hosted.variables.add"
	varRemoveMethod = "apps.hosted.variables.remove"
)

type VariablesClient interface {
	AddVariable(ctx context.Context, token, appID, name, value string) error
	ListVariables(ctx context.Context, token, appID string) ([]string, error)
	RemoveVariable(ctx context.Context, token string, appID string, variableName string) error
}

// AddVariables adds/updates one or more environment variables
func (c *Client) AddVariable(ctx context.Context, token, appID, name, value string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AddVariable")
	defer span.Finish()

	args := struct {
		AppID     string              `json:"app_id"`
		Variables []types.EnvVariable `json:"variables"`
	}{
		appID,
		[]types.EnvVariable{{Name: name, Value: value}},
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, varAddMethod, token, "", body)
	if err != nil {
		return errHttpRequestFailed.WithRootCause(err)
	}

	resp := extendedBaseResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHttpResponseInvalid.WithRootCause(err).AddApiMethod(varAddMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, varAddMethod)
	}

	return nil
}

type listVariablesResponse struct {
	extendedBaseResponse
	VariableNames []string `json:"variable_names"`
}

// ListVariables adds/updates an environment variable
func (c *Client) ListVariables(ctx context.Context, token, appID string) ([]string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.ListVariables")
	defer span.Finish()

	args := struct {
		AppID string `json:"app_id"`
	}{
		appID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return []string{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, varListMethod, token, "", body)
	if err != nil {
		return []string{}, errHttpRequestFailed.WithRootCause(err)
	}

	resp := listVariablesResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return []string{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(varListMethod)
	}

	if !resp.Ok {
		return []string{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, varListMethod)
	}

	return resp.VariableNames, nil
}

// RemoveVariable  removes an environment variable
func (c *Client) RemoveVariable(ctx context.Context, token string, appID string, variableName string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.RemoveVariable")
	defer span.Finish()

	args := struct {
		AppID          string   `json:"app_id"`
		VariablesNames []string `json:"variable_names"`
	}{
		appID,
		[]string{variableName},
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, varRemoveMethod, token, "", body)
	if err != nil {
		return errHttpRequestFailed.WithRootCause(err)
	}

	resp := extendedBaseResponse{}
	err = goutils.JsonUnmarshal(b, &resp)

	if err != nil {
		return errHttpResponseInvalid.WithRootCause(err).AddApiMethod(varRemoveMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, varRemoveMethod)
	}

	return nil
}
