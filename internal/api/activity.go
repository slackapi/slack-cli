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
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	appActivityMethod = "apps.activities.list"
)

type ActivityResult struct {
	Activities []Activity `json:"activities,omitempty"`
	NextCursor string
}

type Activity struct {
	TraceID       string                 `json:"trace_id,omitempty"`
	Level         types.ActivityLevel    `json:"level,omitempty"`
	EventType     types.EventType        `json:"event_type,omitempty"`
	Source        string                 `json:"source,omitempty"`
	ComponentType string                 `json:"component_type,omitempty"`
	ComponentID   string                 `json:"component_id,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Created       int64                  `json:"created,omitempty"`
}

func (a *Activity) CreatedPretty() string {
	return time.Unix(a.Created/1000000, 0).Format("2006-01-02 15:04:05")
}

type activityResponse struct {
	extendedBaseResponse
	ActivityResult
}

type ActivityClient interface {
	Activity(ctx context.Context, token string, activityRequest types.ActivityRequest) (ActivityResult, error)
}

// Get the recent activity for an app
func (c *Client) Activity(ctx context.Context, token string, activityRequest types.ActivityRequest) (ActivityResult, error) {
	var (
		err  error
		span opentracing.Span
	)

	span, ctx = opentracing.StartSpanFromContext(ctx, "apiclient.Activity")
	defer span.Finish()

	if activityRequest.AppID == "" {
		return ActivityResult{}, slackerror.New("app is not deployed")
	}

	// Add the mandatory app_id field
	url := fmt.Sprintf("%s?app_id=%s", appActivityMethod, activityRequest.AppID)

	// Along with any optional filters requested
	url += fmt.Sprintf("&limit=%d", activityRequest.Limit)

	if activityRequest.MinimumLogLevel != "" {
		url += fmt.Sprintf("&min_log_level=%s", activityRequest.MinimumLogLevel)
	}

	if activityRequest.EventType != "" {
		url += fmt.Sprintf("&log_event_type=%s", activityRequest.EventType)
	}

	if activityRequest.MinimumDateCreated != 0 {
		url += fmt.Sprintf("&min_date_created=%d", activityRequest.MinimumDateCreated)
	}

	if activityRequest.MaximumDateCreated != 0 {
		url += fmt.Sprintf("&max_date_created=%d", activityRequest.MaximumDateCreated)
	}

	if activityRequest.ComponentType != "" {
		url += fmt.Sprintf("&component_type=%s", activityRequest.ComponentType)
	}

	if activityRequest.ComponentID != "" {
		url += fmt.Sprintf("&component_id=%s", activityRequest.ComponentID)
	}

	if activityRequest.Source != "" {
		url += fmt.Sprintf("&source=%s", activityRequest.Source)
	}

	if activityRequest.TraceID != "" {
		url += fmt.Sprintf("&trace_id=%s", activityRequest.TraceID)
	}

	b, err := c.get(ctx, url, token, "")
	if err != nil {
		return ActivityResult{}, slackerror.New(slackerror.ErrHTTPRequestFailed).WithRootCause(err)
	}

	resp := activityResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return ActivityResult{}, slackerror.New(slackerror.ErrHTTPResponseInvalid).WithRootCause(err).AddAPIMethod(appActivityMethod)
	}

	if !resp.Ok {
		return ActivityResult{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, appActivityMethod)
	}
	resp.NextCursor = resp.ResponseMetadata.NextCursor

	return resp.ActivityResult, nil
}
