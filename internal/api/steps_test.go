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

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/require"
)

func TestClient_StepsList_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  workflowsStepsListMethod,
		ExpectedRequest: `{"workflow_app_id":"A123","workflow":"#/workflows/my-workflow","function_id":"Fn010N"}`,
		Response:        `{"ok":true,"steps_versions":[{"title":"cool form","workflow_id":"Wf123","step_id":"0","is_deleted":false,"workflow_version_created":"1234"}]}`,
	})
	defer teardown()
	versions, err := c.StepsList(ctx, "token", "#/workflows/my-workflow", "A123")
	require.NoError(t, err)
	require.ElementsMatch(t, versions, []StepVersion{
		{
			Title:                  "cool form",
			WorkflowID:             "Wf123",
			IsDeleted:              false,
			StepID:                 "0",
			WorkflowVersionCreated: "1234",
		},
	})
}

func TestClient_StepsList_Errors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, workflowsStepsListMethod, func(c *Client) error {
		_, err := c.StepsList(ctx, "token", "#/workflows/my-workflow", "A123")
		return err
	})
}

func TestClient_StepsResponsesExport_Ok(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	c, teardown := NewFakeClient(t, FakeClientParams{
		ExpectedMethod:  workflowsStepsResponsesExportMethod,
		ExpectedRequest: `{"workflow_app_id":"A123","workflow":"#/workflows/my-workflow","step_id":"0"}`,
		Response:        `{"ok":true}`,
	})
	defer teardown()
	err := c.StepsResponsesExport(ctx, "token", "#/workflows/my-workflow", "A123", "0")
	require.NoError(t, err)
}

func TestClient_StepsResponsesExport_Errors(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	verifyCommonErrorCases(t, workflowsStepsResponsesExportMethod, func(c *Client) error {
		return c.StepsResponsesExport(ctx, "token", "#/workflows/my-workflow", "A123", "0")
	})
}
