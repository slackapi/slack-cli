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
	appDatastoreQueryMethod      = "apps.datastore.query"
	appDatastoreCountMethod      = "apps.datastore.count"
	appDatastorePutMethod        = "apps.datastore.put"
	appDatastoreBulkPutMethod    = "apps.datastore.bulkPut"
	appDatastoreUpdateMethod     = "apps.datastore.update"
	appDatastoreDeleteMethod     = "apps.datastore.delete"
	appDatastoreBulkDeleteMethod = "apps.datastore.bulkDelete"
	appDatastoreGetMethod        = "apps.datastore.get"
	appDatastoreBulkGetMethod    = "apps.datastore.bulkGet"
)

type DatastoresClient interface {
	AppsDatastorePut(ctx context.Context, token string, request types.AppDatastorePut) (types.AppDatastorePutResult, error)
	AppsDatastoreBulkPut(ctx context.Context, token string, request types.AppDatastoreBulkPut) (types.AppDatastoreBulkPutResult, error)
	AppsDatastoreUpdate(ctx context.Context, token string, request types.AppDatastoreUpdate) (types.AppDatastoreUpdateResult, error)
	AppsDatastoreQuery(ctx context.Context, token string, query types.AppDatastoreQuery) (types.AppDatastoreQueryResult, error)
	AppsDatastoreCount(ctx context.Context, token string, query types.AppDatastoreCount) (types.AppDatastoreCountResult, error)
	AppsDatastoreDelete(ctx context.Context, token string, request types.AppDatastoreDelete) (types.AppDatastoreDeleteResult, error)
	AppsDatastoreBulkDelete(ctx context.Context, token string, request types.AppDatastoreBulkDelete) (types.AppDatastoreBulkDeleteResult, error)
	AppsDatastoreGet(ctx context.Context, token string, request types.AppDatastoreGet) (types.AppDatastoreGetResult, error)
	AppsDatastoreBulkGet(ctx context.Context, token string, request types.AppDatastoreBulkGet) (types.AppDatastoreBulkGetResult, error)
}

type datastoreBaseResponse struct {
	extendedBaseResponse
	Details string `json:"details,omitempty"`
}

func (c *Client) AppsDatastorePut(ctx context.Context, token string, request types.AppDatastorePut) (types.AppDatastorePutResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastorePut")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string                 `json:"datastore,omitempty"`
		AppID     string                 `json:"app_id,omitempty"`
		Item      map[string]interface{} `json:"item,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.Item,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastorePutResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastorePutMethod, token, "", body)
	if err != nil {
		return types.AppDatastorePutResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastorePutResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastorePutResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastorePutMethod)
	}

	if !resp.Ok {
		return types.AppDatastorePutResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastorePutMethod)
	}

	return resp.AppDatastorePutResult, nil
}

func (c *Client) AppsDatastoreBulkPut(ctx context.Context, token string, request types.AppDatastoreBulkPut) (types.AppDatastoreBulkPutResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreBulkPut")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string                   `json:"datastore,omitempty"`
		AppID     string                   `json:"app_id,omitempty"`
		Items     []map[string]interface{} `json:"items,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.Items,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreBulkPutResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreBulkPutMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreBulkPutResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreBulkPutResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreBulkPutResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreBulkPutMethod)
	}

	if !resp.Ok && len(resp.FailedItems) == 0 {
		return types.AppDatastoreBulkPutResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreBulkPutMethod)
	}

	if resp.Datastore == "" {
		resp.Datastore = request.Datastore
	}

	return resp.AppDatastoreBulkPutResult, nil
}

func (c *Client) AppsDatastoreUpdate(ctx context.Context, token string, request types.AppDatastoreUpdate) (types.AppDatastoreUpdateResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreUpdate")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string                 `json:"datastore,omitempty"`
		AppID     string                 `json:"app_id,omitempty"`
		Item      map[string]interface{} `json:"item,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.Item,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreUpdateResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreUpdateMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreUpdateResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreUpdateResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreUpdateResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreUpdateMethod)
	}

	if !resp.Ok {
		return types.AppDatastoreUpdateResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreUpdateMethod)
	}

	return resp.AppDatastoreUpdateResult, nil
}

func (c *Client) AppsDatastoreQuery(ctx context.Context, token string, query types.AppDatastoreQuery) (types.AppDatastoreQueryResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreQuery")
	span.SetTag("app", query.App)
	defer span.Finish()

	args := struct {
		Datastore            string                 `json:"datastore,omitempty"`
		AppID                string                 `json:"app_id,omitempty"`
		Expression           string                 `json:"expression,omitempty"`
		ExpressionAttributes map[string]interface{} `json:"expression_attributes,omitempty"`
		ExpressionValues     map[string]interface{} `json:"expression_values,omitempty"`
		Limit                int                    `json:"limit,omitempty"`
		Cursor               string                 `json:"cursor,omitempty"`
	}{
		query.Datastore,
		query.App,
		query.Expression,
		query.ExpressionAttributes,
		query.ExpressionValues,
		query.Limit,
		query.Cursor,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreQueryResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreQueryMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreQueryResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreQueryResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreQueryResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreQueryMethod)
	}

	if !resp.Ok {
		return types.AppDatastoreQueryResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreQueryMethod)
	}

	resp.NextCursor = resp.ResponseMetadata.NextCursor

	return resp.AppDatastoreQueryResult, nil
}

func (c *Client) AppsDatastoreCount(ctx context.Context, token string, count types.AppDatastoreCount) (types.AppDatastoreCountResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreCount")
	span.SetTag("count", count.App)
	defer span.Finish()

	args := struct {
		Datastore            string                 `json:"datastore,omitempty"`
		AppID                string                 `json:"app_id,omitempty"`
		Expression           string                 `json:"expression,omitempty"`
		ExpressionAttributes map[string]interface{} `json:"expression_attributes,omitempty"`
		ExpressionValues     map[string]interface{} `json:"expression_values,omitempty"`
	}{
		count.Datastore,
		count.App,
		count.Expression,
		count.ExpressionAttributes,
		count.ExpressionValues,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreCountResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreCountMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreCountResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreCountResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreCountResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreCountMethod)
	}

	if !resp.Ok {
		return types.AppDatastoreCountResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreCountMethod)
	}

	return resp.AppDatastoreCountResult, nil
}

func (c *Client) AppsDatastoreDelete(ctx context.Context, token string, request types.AppDatastoreDelete) (types.AppDatastoreDeleteResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreDelete")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string `json:"datastore,omitempty"`
		AppID     string `json:"app_id,omitempty"`
		ID        string `json:"id,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.ID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreDeleteResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreDeleteMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreDeleteResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreDeleteResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreDeleteResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreDeleteMethod)
	}

	if !resp.Ok {
		return types.AppDatastoreDeleteResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreDeleteMethod)
	}

	// the delete API doesn't return id or datastore (yet) so set it if empty
	if resp.ID == "" {
		resp.ID = request.ID
	}
	if resp.Datastore == "" {
		resp.Datastore = request.Datastore
	}

	return resp.AppDatastoreDeleteResult, nil
}

func (c *Client) AppsDatastoreBulkDelete(ctx context.Context, token string, request types.AppDatastoreBulkDelete) (types.AppDatastoreBulkDeleteResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreBulkDelete")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string   `json:"datastore,omitempty"`
		AppID     string   `json:"app_id,omitempty"`
		IDs       []string `json:"ids,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.IDs,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreBulkDeleteResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreBulkDeleteMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreBulkDeleteResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreBulkDeleteResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreBulkDeleteResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreBulkDeleteMethod)
	}

	if !resp.Ok && len(resp.FailedItems) == 0 {
		return types.AppDatastoreBulkDeleteResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreBulkDeleteMethod)
	}

	if resp.Datastore == "" {
		resp.Datastore = request.Datastore
	}

	return resp.AppDatastoreBulkDeleteResult, nil
}

func (c *Client) AppsDatastoreGet(ctx context.Context, token string, request types.AppDatastoreGet) (types.AppDatastoreGetResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreGet")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string `json:"datastore,omitempty"`
		AppID     string `json:"app_id,omitempty"`
		ID        string `json:"id,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.ID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreGetResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreGetMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreGetResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreGetResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreGetResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreGetMethod)
	}

	if !resp.Ok {
		return types.AppDatastoreGetResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreGetMethod)
	}

	return resp.AppDatastoreGetResult, nil
}

func (c *Client) AppsDatastoreBulkGet(ctx context.Context, token string, request types.AppDatastoreBulkGet) (types.AppDatastoreBulkGetResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.AppsDatastoreBulkGet")
	span.SetTag("app", request.App)
	defer span.Finish()

	args := struct {
		Datastore string   `json:"datastore,omitempty"`
		AppID     string   `json:"app_id,omitempty"`
		IDs       []string `json:"ids,omitempty"`
	}{
		request.Datastore,
		request.App,
		request.IDs,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return types.AppDatastoreBulkGetResult{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, appDatastoreBulkGetMethod, token, "", body)
	if err != nil {
		return types.AppDatastoreBulkGetResult{}, errHttpRequestFailed.WithRootCause(err)
	}

	type responseWrapper struct {
		datastoreBaseResponse
		types.AppDatastoreBulkGetResult
	}
	resp := responseWrapper{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.AppDatastoreBulkGetResult{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(appDatastoreBulkGetMethod)
	}

	if !resp.Ok && len(resp.FailedItems) == 0 {
		return types.AppDatastoreBulkGetResult{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, appDatastoreBulkGetMethod)
	}

	if resp.Datastore == "" {
		resp.Datastore = request.Datastore
	}

	return resp.AppDatastoreBulkGetResult, nil
}
