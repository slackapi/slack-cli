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
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/require"
)

var fakeResult = `{"ok":true,
"activities": [{"trace_id":"12345"}]
}`

func Test_APIClient_ActivityErrorsIfAppIDIsEmpty(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod: appActivityMethod,
	})
	defer teardown()
	_, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID: "",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "app is not deployed")
}

func Test_APIClient_ActivityBasicSuccessfulGET(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID: "A123",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityEventType(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&log_event_type=silly",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:     "A123",
		EventType: "silly",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityLogLevel(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&min_log_level=silly",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:           "A123",
		MinimumLogLevel: "silly",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityMinDateCreated(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&min_date_created=1337",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:              "A123",
		MinimumDateCreated: 1337,
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityComponentType(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&component_type=defirbulator",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:         "A123",
		ComponentType: "defirbulator",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityComponentID(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&component_id=raspberry",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:       "A123",
		ComponentID: "raspberry",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivitySource(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&source=beer",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:  "A123",
		Source: "beer",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityTraceID(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123&limit=0&trace_id=stealth",
		Response:            fakeResult,
	})
	defer teardown()
	result, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID:   "A123",
		TraceID: "stealth",
	})
	require.NoError(t, err)
	require.Equal(t, result.Activities[0].TraceID, "12345")
}

func Test_APIClient_ActivityResponseNotOK(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123",
		Response:            `{"ok":false, "error": "internal_error"}`,
	})
	defer teardown()
	_, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID: "A123",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal_error")
}

func Test_APIClient_ActivityInvalidResponse(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123",
		Response:            `badjson`,
	})
	defer teardown()
	_, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID: "A123",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), slackerror.ErrHTTPResponseInvalid)
}

func Test_APIClient_ActivityInvalidJSON(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:      appActivityMethod,
		ExpectedQuerystring: "app_id=A123",
		Response:            `badtime`,
	})
	defer teardown()
	_, err := c.Activity(ctx, "token", types.ActivityRequest{
		AppID: "A123",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), slackerror.ErrUnableToParseJSON)
}
