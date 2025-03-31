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

	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	workflowsStepsListMethod            = "functions.workflows.steps.list"
	workflowsStepsResponsesExportMethod = "functions.workflows.steps.responses.export"
)

const openFormFunctionId = "Fn010N"

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
	StepsList(ctx context.Context, token string, workflow string, appId string) ([]StepVersion, error)
	StepsResponsesExport(ctx context.Context, token string, workflow string, appId string, stepId string) error
}

func (c *Client) StepsList(ctx context.Context, token string, workflow string, appId string) ([]StepVersion, error) {
	args := struct {
		WorkflowAppId string `json:"workflow_app_id"`
		Workflow      string `json:"workflow"`
		FunctionId    string `json:"function_id"`
	}{
		appId,
		workflow,
		openFormFunctionId,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return []StepVersion{}, errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, workflowsStepsListMethod, token, "", body)
	if err != nil {
		return []StepVersion{}, errHttpRequestFailed.WithRootCause(err)
	}

	resp := stepsListResponse{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return []StepVersion{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(workflowsStepsListMethod)
	}

	if !resp.Ok {
		return []StepVersion{}, slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, workflowsStepsListMethod)
	}

	return resp.StepVersions, nil
}

func (c *Client) StepsResponsesExport(ctx context.Context, token string, workflow string, appId string, stepId string) error {
	args := struct {
		WorkflowAppId string `json:"workflow_app_id"`
		Workflow      string `json:"workflow"`
		StepId        string `json:"step_id"`
	}{
		appId,
		workflow,
		stepId,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, workflowsStepsResponsesExportMethod, token, "", body)
	if err != nil {
		return errHttpRequestFailed.WithRootCause(err)
	}

	resp := exportResponse{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return errHttpResponseInvalid.WithRootCause(err).AddApiMethod(workflowsStepsResponsesExportMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, workflowsStepsResponsesExportMethod)
	}

	return nil
}
