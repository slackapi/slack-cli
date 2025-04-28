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

// For easy mocking
type ExternalAuthClient interface {
	AppsAuthExternalStart(context.Context, string, string, string) (string, error)
	AppsAuthExternalDelete(context.Context, string, string, string, string) error
	AppsAuthExternalClientSecretAdd(context.Context, string, string, string, string) error
	AppsAuthExternalList(context.Context, string, string, bool) (types.ExternalAuthorizationInfoLists, error)
	AppsAuthExternalSelectAuth(context.Context, string, string, string, string, string) error
}

const (
	appsAuthExternalStartMethod           = "apps.auth.external.start"
	appsAuthExternalDeleteMethod          = "apps.auth.external.delete"
	appsAuthExternalClientSecretAddMethod = "apps.auth.external.clientSecret.add"
	appsAuthExternalListMethod            = "apps.auth.external.list"
	appsAuthExternalSelectAuthMethod      = "apps.auth.external.authMapping.update"
)

type appsAuthExternalStartResponse struct {
	extendedBaseResponse
	AuthorizationUrl string `json:"authorization_url"`
}

type appsAuthExternalDeleteResponse struct {
	extendedBaseResponse
}

type appsAuthExternalClientSecretAddResponse struct {
	extendedBaseResponse
}

type appsAuthExternalListResponse struct {
	extendedBaseResponse
	appsAuthExternalListResult
}

type appsAuthExternalListResult struct {
	Authorizations []types.ExternalAuthorizationInfo `json:"authorizations"`
	Workflows      []types.WorkflowsInfo             `json:"workflows,omitempty"`
}

type appsAuthExternalSelectAuthResponse struct {
	extendedBaseResponse
}

func (c *Client) AppsAuthExternalStart(ctx context.Context, token, appID, providerKey string) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsAuthExternalStart")
	defer span.Finish()

	args := struct {
		AppID       string `json:"app_id,omitempty"`
		ProviderKey string `json:"provider_key,omitempty"`
	}{
		appID,
		providerKey,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return "", errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appsAuthExternalStartMethod, token, "", body)
	if err != nil {
		return "", errHTTPRequestFailed.WithRootCause(err)
	}

	var resp appsAuthExternalStartResponse
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return "", errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appsAuthExternalStartMethod)
	}

	if !resp.Ok {
		return "", slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appsAuthExternalStartMethod)
	}

	return resp.AuthorizationUrl, nil
}

func (c *Client) AppsAuthExternalDelete(ctx context.Context, token, appID, providerKey string, externalTokenId string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsAuthExternalDelete")
	defer span.Finish()

	var body []byte
	var err error
	if externalTokenId != "" {
		args := struct {
			ExternalTokenID string `json:"external_token_id,omitempty"`
		}{
			externalTokenId,
		}
		body, err = json.Marshal(args)
	} else if providerKey == "" {
		args := struct {
			AppID string `json:"app_id,omitempty"`
		}{
			appID,
		}
		body, err = json.Marshal(args)
	} else {
		args := struct {
			AppID       string `json:"app_id,omitempty"`
			ProviderKey string `json:"provider_key,omitempty"`
		}{
			appID,
			providerKey,
		}
		body, err = json.Marshal(args)
	}

	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appsAuthExternalDeleteMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	var resp appsAuthExternalDeleteResponse
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appsAuthExternalDeleteMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appsAuthExternalDeleteMethod)
	}

	return nil
}

func (c *Client) AppsAuthExternalClientSecretAdd(ctx context.Context, token, appID, providerKey string, clientSecret string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsAuthExternalClientSecretAdd")
	defer span.Finish()

	var body []byte
	var err error
	args := struct {
		AppID        string `json:"app_id,omitempty"`
		ProviderKey  string `json:"provider_key,omitempty"`
		ClientSecret string `json:"client_secret,omitempty"`
	}{
		appID,
		providerKey,
		clientSecret,
	}
	body, err = json.Marshal(args)

	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appsAuthExternalClientSecretAddMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	var resp appsAuthExternalClientSecretAddResponse
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appsAuthExternalClientSecretAddMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appsAuthExternalClientSecretAddMethod)
	}

	return nil
}

// https://api.slack.com/methods/apps.auth.external.list
// AppsAuthExternalList returns information about external auth providers settings, associated tokens, workflow list that require developer auths
// and what are the selected auths for each provider in each workflow within the app.
// The boolean flag include_workflows can be used to get the workflow related information when needed.

func (c *Client) AppsAuthExternalList(ctx context.Context, token, appID string, includeWorkflows bool) (types.ExternalAuthorizationInfoLists, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsAuthExternalList")
	defer span.Finish()

	args := struct {
		AppID            string `json:"app_id"`
		IncludeWorkflows bool   `json:"include_workflows,omitempty"`
	}{
		appID,
		includeWorkflows,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.ExternalAuthorizationInfoLists{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appsAuthExternalListMethod, token, "", body)
	if err != nil {
		return types.ExternalAuthorizationInfoLists{}, errHTTPRequestFailed.WithRootCause(err)
	}

	var resp appsAuthExternalListResponse
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.ExternalAuthorizationInfoLists{}, errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appsAuthExternalListMethod)
	}

	if !resp.Ok {
		return types.ExternalAuthorizationInfoLists{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appsAuthExternalListMethod)
	}

	return types.ExternalAuthorizationInfoLists(resp.appsAuthExternalListResult), nil
}

func (c *Client) AppsAuthExternalSelectAuth(ctx context.Context, token, appID, providerKey, workflowId, externalTokenId string) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsAuthExternalSelectAuth")
	defer span.Finish()

	var body []byte
	var err error
	args := struct {
		AppID            string `json:"app_id,omitempty"`
		ProviderKey      string `json:"provider_key,omitempty"`
		WorkflowId       string `json:"workflow_id,omitempty"`
		ExternalTokenID  string `json:"external_token_id,omitempty"`
		MappingOwnerType string `json:"mapping_owner_type,omitempty"`
	}{
		appID,
		providerKey,
		workflowId,
		externalTokenId,
		"DEVELOPER",
	}
	body, err = json.Marshal(args)

	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appsAuthExternalSelectAuthMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	var resp appsAuthExternalSelectAuthResponse
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddApiMethod(appsAuthExternalSelectAuthMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appsAuthExternalSelectAuthMethod)
	}

	return nil
}
