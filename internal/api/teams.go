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
	"sort"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	teamsInfoMethod     = "team.info"
	authTeamsListMethod = "auth.teams.list"
)

type TeamClient interface {
	TeamsInfo(ctx context.Context, token, teamID string) (*types.TeamInfo, error)
	AuthTeamsList(ctx context.Context, token string, limit int) ([]types.TeamInfo, string, error)
}

type TeamInfoResponse struct {
	extendedBaseResponse
	Team types.TeamInfo `json:"team"`
}

// TeamInfo returns information about the team such as team name
func (c *Client) TeamsInfo(ctx context.Context, token, teamID string) (*types.TeamInfo, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.teamsInfo")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("team", teamID)

	b, err := c.postForm(ctx, teamsInfoMethod, values)
	if err != nil {
		return nil, errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return nil, errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := TeamInfoResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return nil, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsTriggersPermissionsListMethod)
	}

	if !resp.Ok {
		return nil, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsTriggersPermissionsListMethod)
	}

	return &resp.Team, nil
}

type AuthTeamsListResponse struct {
	extendedBaseResponse
	Teams            []types.TeamInfo `json:"teams"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

// Default value for the maximum number of workspaces to return
var DefaultAuthTeamsListPageSize = 100

// AuthTeamsList returns a list of workspaces that the user has access to through the token
// as well as a pagination cursor if the org has more workspaces than returned in this request.
// Used to retrieve enterprise workspaces that belong to an org.
// Specify the maximum number of workspaces to return via the `limit` param (value between 1 and 1000).
func (c *Client) AuthTeamsList(ctx context.Context, token string, limit int) ([]types.TeamInfo, string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.authTeamsList")
	defer span.Finish()

	var values = url.Values{}
	values.Add("token", token)
	values.Add("limit", fmt.Sprint(limit))

	b, err := c.postForm(ctx, authTeamsListMethod, values)
	if err != nil {
		return nil, "", errHTTPRequestFailed.WithRootCause(err)
	}

	if b == nil {
		return nil, "", errHTTPResponseInvalid.WithRootCause(slackerror.New("empty body"))
	}

	resp := AuthTeamsListResponse{}
	err = goutils.JSONUnmarshal(b, &resp)

	if err != nil {
		return nil, "", errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(authTeamsListMethod)
	}

	if !resp.Ok {
		return nil, "", slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, authTeamsListMethod)
	}

	sort.Slice(resp.Teams, func(i, j int) bool {
		return resp.Teams[i].Name < resp.Teams[j].Name
	})

	return resp.Teams, resp.ResponseMetadata.NextCursor, nil
}
