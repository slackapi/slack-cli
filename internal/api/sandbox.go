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

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const sandboxListMethod = "developer.sandbox.list"

// SandboxClient is the interface for sandbox-related API calls
type SandboxClient interface {
	ListSandboxes(ctx context.Context, token string, filter string) ([]types.Sandbox, error)
}

type listSandboxesResponse struct {
	extendedBaseResponse
	Sandboxes []types.Sandbox `json:"sandboxes"`
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
