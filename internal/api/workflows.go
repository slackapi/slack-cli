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
	"fmt"

	"github.com/pkg/errors"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

const (
	workflowsTriggersCreateMethod = "workflows.triggers.create"
	workflowsTriggersUpdateMethod = "workflows.triggers.update"
	workflowsTriggersDeleteMethod = "workflows.triggers.delete"
	workflowsTriggersListMethod   = "workflows.triggers.list"
	workflowsTriggersInfoMethod   = "workflows.triggers.info"
)

type triggerResult struct {
	Trigger types.DeployedTrigger `json:"trigger,omitempty"`
}

type Shortcut struct {
	ButtonText string `json:"button_text,omitempty"`
}

type Input struct {
	Value        string `json:"value,omitempty"`
	Customizable bool   `json:"customizable,omitempty"`
}

type Inputs map[string]*Input

type TriggerRequest struct {
	Type          string         `json:"type"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Shortcut      *Shortcut      `json:"shortcut,omitempty"`
	Workflow      string         `json:"workflow"`
	WorkflowAppId string         `json:"workflow_app_id"`
	Inputs        Inputs         `json:"inputs,omitempty"`
	Event         *types.RawJSON `json:"event,omitempty"`
	Schedule      *types.RawJSON `json:"schedule,omitempty"`
	WebHook       *types.RawJSON `json:"webhook,omitempty"`
	Service       *types.RawJSON `json:"service,omitempty"`
}

type triggerInfoRequest struct {
	TriggerId string `json:"trigger_id"`
}

type triggerInfoResponse struct {
	extendedBaseResponse
	triggerResult
}

type triggerDeleteRequest struct {
	TriggerId string `json:"trigger_id"`
}

type triggerDeleteResponse struct {
	extendedBaseResponse
}

type TriggerUpdateRequest struct {
	TriggerRequest
	TriggerId string `json:"trigger_id"`
}

type triggerCreateOrUpdateErrorDetails []triggerCreateOrUpdateErrorDetail

type MissingParameterDetail struct {
	Name string `json:"missing_parameter_name"`
	Type string `json:"missing_parameter_type"`
}

type triggerCreateOrUpdateErrorDetail struct {
	slackerror.ErrorDetail
	MissingParameterDetail
}

type triggerCreateOrUpdateBaseResponse struct {
	baseResponse
	warningResponse
	Errors triggerCreateOrUpdateErrorDetails `json:"errors,omitempty"`
}

type triggerCreateOrUpdateResponse struct {
	triggerCreateOrUpdateBaseResponse
	triggerResult
}

type TriggerCreateOrUpdateError struct {
	Err                    error
	MissingParameterDetail MissingParameterDetail
}

func (e *TriggerCreateOrUpdateError) Error() string {
	return e.Err.Error()
}

type TriggerListRequest struct {
	AppId  string `json:"app_id,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Type   string `json:"type,omitempty"`
}

// For easy mocking
type WorkflowsClient interface {
	WorkflowsTriggersCreate(context.Context, string, TriggerRequest) (types.DeployedTrigger, error)
	WorkflowsTriggersUpdate(context.Context, string, TriggerUpdateRequest) (types.DeployedTrigger, error)
	WorkflowsTriggersDelete(context.Context, string, string) error
	WorkflowsTriggersInfo(context.Context, string, string) (types.DeployedTrigger, error)
	WorkflowsTriggersList(context.Context, string, TriggerListRequest) ([]types.DeployedTrigger, string, error)
}

// WorkflowsTriggersCreate will create a new trigger using the method workflows.trigger.create
func (c *Client) WorkflowsTriggersCreate(ctx context.Context, token string, trigger TriggerRequest) (types.DeployedTrigger, error) {

	body, err := json.Marshal(trigger)
	if err != nil {
		return types.DeployedTrigger{}, errInvalidArguments.WithRootCause(err)
	}

	return c.workflowsTriggerSave(ctx, token, workflowsTriggersCreateMethod, body)

}

// WorkflowsTriggersUpdate will update an existing trigger using the method workflows.trigger.update
func (c *Client) WorkflowsTriggersUpdate(ctx context.Context, token string, trigger TriggerUpdateRequest) (types.DeployedTrigger, error) {

	body, err := json.Marshal(trigger)
	if err != nil {
		return types.DeployedTrigger{}, errInvalidArguments.WithRootCause(err)
	}

	return c.workflowsTriggerSave(ctx, token, workflowsTriggersUpdateMethod, body)
}

// WorkflowsTriggersDelete will delete an existing trigger using the method workflows.trigger.delete
func (c *Client) WorkflowsTriggersDelete(ctx context.Context, token string, triggerId string) error {
	deleteRequest := triggerDeleteRequest{
		TriggerId: triggerId,
	}

	body, err := json.Marshal(deleteRequest)
	if err != nil {
		return errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, workflowsTriggersDeleteMethod, token, "", body)
	if err != nil {
		return errHttpRequestFailed.WithRootCause(err)
	}

	resp := triggerDeleteResponse{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return errHttpResponseInvalid.WithRootCause(err).AddApiMethod(workflowsTriggersDeleteMethod)
	}

	if !resp.Ok {
		return slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, workflowsTriggersDeleteMethod)
	}

	return nil
}

func (c *Client) workflowsTriggerSave(ctx context.Context, token string, method string, body []byte) (types.DeployedTrigger, error) {
	b, err := c.postJSON(ctx, method, token, "", body)
	if err != nil {
		return types.DeployedTrigger{}, errHttpRequestFailed.WithRootCause(err)
	}

	resp := triggerCreateOrUpdateResponse{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return types.DeployedTrigger{}, errHttpResponseInvalid.WithRootCause(err).AddApiMethod(method)
	}

	if !resp.Ok {
		errorDetails, missingParamError := parseMissingParameterErrors(resp.Errors)
		err = slackerror.NewApiError(resp.Error, resp.Description, errorDetails, method)
		if missingParamError != nil {
			return types.DeployedTrigger{}, &TriggerCreateOrUpdateError{
				Err: err, MissingParameterDetail: *missingParamError}
		} else {
			return types.DeployedTrigger{}, err
		}
	}

	serverTrigger := resp.Trigger

	return serverTrigger, nil
}

func parseMissingParameterErrors(details triggerCreateOrUpdateErrorDetails) ([]slackerror.ErrorDetail, *MissingParameterDetail) {
	slackErrorDetails := []slackerror.ErrorDetail{}
	var missingParameterDetail *MissingParameterDetail
	for _, detail := range details {
		slackErrorDetails = append(slackErrorDetails, detail.ErrorDetail)
		if detail.Name != "" {
			missingParameterDetail = &detail.MissingParameterDetail
		}
	}
	return slackErrorDetails, missingParameterDetail
}

// WorkflowsTriggersList will list the existing triggers
func (c *Client) WorkflowsTriggersList(ctx context.Context, token string, listArgs TriggerListRequest) ([]types.DeployedTrigger, string, error) {
	requestArgs := listArgs

	if requestArgs.Type == "all" {
		requestArgs.Type = ""
	}

	body, err := json.Marshal(requestArgs)
	if err != nil {
		return []types.DeployedTrigger{}, "", errInvalidArguments.WithRootCause(err)
	}

	b, err := c.postJSON(ctx, workflowsTriggersListMethod, token, "", body)
	if err != nil {
		return []types.DeployedTrigger{}, "", errHttpRequestFailed.WithRootCause(err)
	}

	// workflowTriggersListResponse details to be saved
	type workflowTriggersListResponse struct {
		extendedBaseResponse
		Triggers []types.DeployedTrigger
	}

	resp := workflowTriggersListResponse{}
	err = goutils.JsonUnmarshal(b, &resp)
	if err != nil {
		return []types.DeployedTrigger{}, "", errHttpResponseInvalid.WithRootCause(err).AddApiMethod(workflowsTriggersListMethod)
	}

	if !resp.Ok {
		return []types.DeployedTrigger{}, "", slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, workflowsTriggersListMethod)
	}

	return resp.Triggers, resp.ResponseMetadata.NextCursor, nil
}

// WorkflowsTriggersInfo will retrieve information on an existing trigger
func (c *Client) WorkflowsTriggersInfo(ctx context.Context, token, triggerId string) (types.DeployedTrigger, error) {
	infoRequest := triggerInfoRequest{
		TriggerId: triggerId,
	}

	body, err := json.Marshal(infoRequest)
	if err != nil {
		return types.DeployedTrigger{}, errors.WithStack(fmt.Errorf("%s: : %s", errInvalidArguments, err))
	}

	b, err := c.postJSON(ctx, workflowsTriggersInfoMethod, token, "", body)
	if err != nil {
		return types.DeployedTrigger{}, errors.WithStack(fmt.Errorf("%s: %s", errHttpRequestFailed, err))
	}

	resp := triggerInfoResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return types.DeployedTrigger{}, errors.WithStack(fmt.Errorf("%s: %s", errHttpResponseInvalid, err))
	}

	if !resp.Ok {
		return types.DeployedTrigger{}, errors.WithStack(slackerror.NewApiError(resp.Error, resp.Description, resp.Errors, workflowsTriggersInfoMethod))
	}

	return resp.Trigger, nil
}
