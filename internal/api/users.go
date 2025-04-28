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
	usersInfoMethod = "users.info"
)

type UserClient interface {
	UsersInfo(ctx context.Context, token, userID string) (*types.UserInfo, error)
}

type UserInfoResponse struct {
	extendedBaseResponse
	User types.UserInfo `json:"user"`
}

// UsersInfo returns information about the user such as email address
func (c *Client) UsersInfo(ctx context.Context, token, userID string) (*types.UserInfo, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.usersInfo")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("user", userID)

	b, err := c.postForm(ctx, usersInfoMethod, values)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return nil, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := UserInfoResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsTriggersPermissionsListMethod)
	}

	if !resp.Ok {
		return nil, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsTriggersPermissionsListMethod)
	}

	return &resp.User, nil
}
