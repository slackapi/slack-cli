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
	"encoding/json"

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	workflowsStepsListMethod            = "functions.workflows.steps.list"
	workflowsStepsResponsesExportMethod = "functions.workflows.steps.responses.export"
)

const openFormFunctionID = "Fn010N"

type StepVersion struct {
	Title                  string `json:"title"`
	WorkflowID             string `json:"workflow_id"`
	StepID                 string `json:"step_id"`
	IsDeleted              bool   `json:"is_deleted"`
	WorkflowVersionCreated string `json:"workflow_version_created"`
}

type stepsListResponse struct {
	extendedBaseResponse
	StepVersions []StepVersion `json:"steps_versions"`
}

type exportResponse struct {
	extendedBaseResponse
}

type StepsClient interface {
	StepsList(ctx context.Context, token string, workflow string, appID string) ([]StepVersion, error)
	StepsResponsesExport(ctx context.Context, token string, workflow string, appID string, stepID string) error
}

func (c *Client) StepsList(ctx context.Context, token string, workflow string, appID string) ([]StepVersion, error) {
	args := struct {
		WorkflowAppID string `json:"workflow_app_id"`
		Workflow      string `json:"workflow"`
		FunctionID    string `json:"function_id"`
	}{
		appID,
		workflow,
		openFormFunctionID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return []StepVersion{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, workflowsStepsListMethod, token, "", body)
	if err != nil {
		return []StepVersion{}, errHTTPRequestFailed.WithRootCause(err)
	}

	resp := stepsListResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return []StepVersion{}, errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsStepsListMethod)
	}

	if !resp.Ok {
		return []StepVersion{}, slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsStepsListMethod)
	}

	return resp.StepVersions, nil
}

func (c *Client) StepsResponsesExport(ctx context.Context, token string, workflow string, appID string, stepID string) error {
	args := struct {
		WorkflowAppID string `json:"workflow_app_id"`
		Workflow      string `json:"workflow"`
		StepID        string `json:"step_id"`
	}{
		appID,
		workflow,
		stepID,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, workflowsStepsResponsesExportMethod, token, "", body)
	if err != nil {
		return errHTTPRequestFailed.WithRootCause(err)
	}

	resp := exportResponse{}
	err = goutils.JSONUnmarshal(b, &resp)
	if err != nil {
		return errHTTPResponseInvalid.WithRootCause(err).AddAPIMethod(workflowsStepsResponsesExportMethod)
	}

	if !resp.Ok {
		return slackerror.NewAPIError(resp.Error, resp.Description, resp.Errors, workflowsStepsResponsesExportMethod)
	}

	return nil
}
