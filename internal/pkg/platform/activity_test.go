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

package platform

import (
	"context"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_prettifyActivity(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"nil payload should result in valid log without nulls": {
			activity: api.Activity{
				TraceID:       "a123",
				Level:         "info",
				EventType:     "unknown",
				Source:        "slack",
				ComponentType: "new_thing",
				ComponentID:   "a789",
				Created:       1686939542,
			},
			expectedResults: []string{
				"info",
				"a789",
				"Trace=a123",
			},
		},
		"empty activity should not contain nulls": {
			activity: api.Activity{},
		},
		"unknown EventType should result in valid log without nulls": {
			activity: api.Activity{
				TraceID:       "a123",
				Level:         "info",
				EventType:     "unknown",
				Source:        "slack",
				ComponentType: "new_thing",
				ComponentID:   "a789",
				Payload: map[string]any{
					"some": "data",
				},
				Created: 1686939542,
			},
			expectedResults: []string{
				"info",
				"a789",
				"Trace=a123",
				`{"some":"data"}`,
			},
		},
		"warn level should be styled": {
			activity: api.Activity{
				Level:     types.WARN,
				EventType: "unknown",
			},
			expectedResults: []string{},
		},
		"error level should be styled": {
			activity: api.Activity{
				Level:     types.ERROR,
				EventType: "unknown",
			},
			expectedResults: []string{},
		},
		"fatal level should be styled": {
			activity: api.Activity{
				Level:     types.FATAL,
				EventType: "unknown",
			},
			expectedResults: []string{},
		},
		"datastore_request_result event": {
			activity: api.Activity{
				EventType:   types.DatastoreRequestResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"request_type":   "get",
					"datastore_name": "MyDS",
					"details":        "id: 123",
				},
			},
			expectedResults: []string{"get", "MyDS", "succeeded"},
		},
		"external_auth_missing_function event": {
			activity: api.Activity{
				EventType:   types.ExternalAuthMissingFunction,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"function_id": "fn1",
				},
			},
			expectedResults: []string{"fn1", "missing"},
		},
		"external_auth_result event": {
			activity: api.Activity{
				EventType:   types.ExternalAuthResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"user_id":      "U1",
					"team_id":      "T1",
					"app_id":       "A1",
					"provider_key": "google",
				},
			},
			expectedResults: []string{"U1", "T1"},
		},
		"external_auth_started event": {
			activity: api.Activity{
				EventType:   types.ExternalAuthStarted,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"user_id":      "U1",
					"team_id":      "T1",
					"app_id":       "A1",
					"provider_key": "google",
				},
			},
			expectedResults: []string{"U1", "T1"},
		},
		"external_auth_token_fetch_result event": {
			activity: api.Activity{
				EventType:   types.ExternalAuthTokenFetchResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"user_id":      "U1",
					"team_id":      "T1",
					"app_id":       "A1",
					"provider_key": "google",
				},
			},
			expectedResults: []string{"U1", "T1"},
		},
		"function_deployment event": {
			activity: api.Activity{
				EventType:   types.FunctionDeployment,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"user_id": "U1",
					"team_id": "T1",
					"action":  "deploy",
				},
			},
			expectedResults: []string{"U1", "T1"},
		},
		"function_execution_output event": {
			activity: api.Activity{
				EventType:   types.FunctionExecutionOutput,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"log": "some output",
				},
			},
			expectedResults: []string{"some output"},
		},
		"function_execution_result event": {
			activity: api.Activity{
				EventType:   types.FunctionExecutionResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"function_name": "MyFunc",
					"function_type": "custom",
				},
			},
			expectedResults: []string{"MyFunc", "completed"},
		},
		"function_execution_started event": {
			activity: api.Activity{
				EventType:   types.FunctionExecutionStarted,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"function_name": "MyFunc",
					"function_type": "custom",
				},
			},
			expectedResults: []string{"MyFunc", "started"},
		},
		"trigger_executed event": {
			activity: api.Activity{
				EventType:   types.TriggerExecuted,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"function_name": "MyFunc",
				},
			},
			expectedResults: []string{"MyFunc"},
		},
		"trigger_payload_received event": {
			activity: api.Activity{
				EventType:   types.TriggerPayloadReceived,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"log": "payload data",
				},
			},
			expectedResults: []string{"payload data"},
		},
		"workflow_billing_result event": {
			activity: api.Activity{
				EventType:   types.WorkflowBillingResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"workflow_name":     "WF1",
					"is_billing_result": true,
					"billing_reason":    "function_execution",
				},
			},
			expectedResults: []string{"WF1"},
		},
		"workflow_bot_invited event": {
			activity: api.Activity{
				EventType:   types.WorkflowBotInvited,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"channel_id":  "C1",
					"bot_user_id": "B1",
				},
			},
			expectedResults: []string{"C1", "B1"},
		},
		"workflow_created_from_template event": {
			activity: api.Activity{
				EventType:   types.WorkflowCreatedFromTemplate,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"workflow_name": "WF1",
					"template_id":   "tmpl1",
				},
			},
			expectedResults: []string{"WF1", "tmpl1"},
		},
		"workflow_execution_result event": {
			activity: api.Activity{
				EventType:   types.WorkflowExecutionResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"workflow_name": "WF1",
				},
			},
			expectedResults: []string{"WF1", "completed"},
		},
		"workflow_execution_started event": {
			activity: api.Activity{
				EventType:   types.WorkflowExecutionStarted,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"workflow_name": "WF1",
				},
			},
			expectedResults: []string{"WF1", "started"},
		},
		"workflow_published event": {
			activity: api.Activity{
				EventType:   types.WorkflowPublished,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"workflow_name": "WF1",
				},
			},
			expectedResults: []string{"WF1", "published"},
		},
		"workflow_step_execution_result event": {
			activity: api.Activity{
				EventType:   types.WorkflowStepExecutionResult,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"function_name": "MyFunc",
					"function_type": "custom",
				},
			},
			expectedResults: []string{"MyFunc", "completed"},
		},
		"workflow_step_started event": {
			activity: api.Activity{
				EventType:   types.WorkflowStepStarted,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"current_step": float64(2),
					"total_steps":  float64(5),
				},
			},
			expectedResults: []string{"2", "5", "started"},
		},
		"workflow_unpublished event": {
			activity: api.Activity{
				EventType:   types.WorkflowUnpublished,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"workflow_name": "WF1",
				},
			},
			expectedResults: []string{"WF1", "unpublished"},
		},
		"external_auth_missing_selected_auth event": {
			activity: api.Activity{
				EventType:   types.ExternalAuthMissingSelectedAuth,
				Level:       types.INFO,
				ComponentID: "c1",
				TraceID:     "t1",
				Payload: map[string]any{
					"code": "auth_error",
				},
			},
			expectedResults: []string{"auth_error"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualResult := prettifyActivity(tc.activity)
			for _, expectedResult := range tc.expectedResults {
				require.Contains(t, actualResult, expectedResult)
			}
			// Confirm no nil pointers leak to output
			require.NotContains(t, actualResult, "nil")
			require.NotContains(t, actualResult, "NIL")
			require.NotContains(t, actualResult, "null")
		})
	}
}

func TestPlatformActivity_StreamingLogs(t *testing.T) {
	for name, tc := range map[string]struct {
		Setup           func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context
		Args            types.ActivityArgs
		ExpectedAsserts func(*testing.T, context.Context, *shared.ClientsMock) // Optional
		ExpectedError   error                                                  // Optional
	}{
		"should return error if context contains no token": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				ctx = context.WithValue(ctx, config.ContextToken, nil)
				return ctx
			},
			ExpectedError: slackerror.New(slackerror.ErrAuthToken),
		},
		"should return error if session validation fails": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				cm.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, slackerror.New("mock_broken_validation"))
				return ctx
			},
			ExpectedError: slackerror.New("mock_broken_validation"),
		},
		"should return error from Activity API request if TailArg is not set": {
			Args: types.ActivityArgs{
				TailArg: false,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				cm.API.On("Activity", mock.Anything, mock.Anything, mock.Anything).Return(api.ActivityResult{}, slackerror.New("mock_broken_logs"))
				return ctx
			},
			ExpectedError: slackerror.New("mock_broken_logs"),
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.API.AssertNumberOfCalls(t, "Activity", 1)
			},
		},
		"should return nil and invoke Activity API only once if TailArg is not set": {
			Args: types.ActivityArgs{
				TailArg:           false,
				IdleTimeoutM:      1,
				PollingIntervalMS: 20, // poll activity every 20 ms
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				cm.API.On("Activity", mock.Anything, mock.Anything, mock.Anything).
					Return(api.ActivityResult{
						Activities: []api.Activity{
							{
								Level:       types.WARN,
								ComponentID: "a123",
								TraceID:     "tr123",
							},
							{
								Level:       types.ERROR,
								ComponentID: "a456",
								TraceID:     "tr456",
							},
						},
					}, nil)
				ctx, cancel := context.WithCancel(ctx)
				go func() {
					time.Sleep(time.Millisecond * 50) // cancel activity in 50 ms
					cancel()
				}()
				return ctx
			},
			ExpectedError: nil,
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// with the above tail argument not included we call the method once.
				cm.API.AssertNumberOfCalls(t, "Activity", 1)
				assert.Contains(t, cm.GetStdoutOutput(), "[warn] [a123] (Trace=tr123)")
				assert.Contains(t, cm.GetStdoutOutput(), "[error] [a456] (Trace=tr456)")
			},
		},
		"should return nil and invoke Activity API twice if TailArg is set while polling": {
			Args: types.ActivityArgs{
				TailArg:           true,
				IdleTimeoutM:      0,
				PollingIntervalMS: 20, // poll activity every 20 ms
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				cm.API.On("Activity", mock.Anything, mock.Anything, mock.Anything).Return(api.ActivityResult{}, nil)
				ctx, cancel := context.WithCancel(ctx)
				go func() {
					time.Sleep(time.Millisecond * 50) // cancel activity in 50 ms
					cancel()
				}()
				return ctx
			},
			ExpectedError: nil,
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// with the above polling/canceling setup, expectation is activity called two times.
				cm.API.AssertNumberOfCalls(t, "Activity", 2)
			},
		},
		"should return nil if TailArg is set and context is canceled": {
			Args: types.ActivityArgs{
				TailArg:           true,
				PollingIntervalMS: 20, // poll activity every 20 ms
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				cm.API.On("Activity", mock.Anything, mock.Anything, mock.Anything).Return(api.ActivityResult{}, nil)
				ctx, cancel := context.WithCancel(ctx)
				go func() {
					time.Sleep(time.Millisecond * 10) // cancel activity in 10 ms
					cancel()
				}()
				return ctx
			},
			ExpectedError: nil,
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// with the above polling/canceling setup, expectation is activity called only once.
				cm.API.AssertNumberOfCalls(t, "Activity", 1)
			},
		},
		"should return nil if TailArg is set and activity request fails while polling": {
			Args: types.ActivityArgs{
				TailArg:           true,
				IdleTimeoutM:      1,
				PollingIntervalMS: 20, // poll activity every 20 ms
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) context.Context {
				cm.API.On("Activity", mock.Anything, mock.Anything, mock.Anything).Return(api.ActivityResult{}, slackerror.New("mock_broken_logs"))
				ctx, cancel := context.WithCancel(ctx)
				go func() {
					time.Sleep(time.Millisecond * 50) // cancel activity in 50 ms
					cancel()
				}()
				return ctx
			},
			ExpectedError: nil,
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// with the above polling/canceling setup, expectation is activity called three times.
				cm.API.AssertNumberOfCalls(t, "Activity", 3)
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Create mocks
			ctxMock := slackcontext.MockContext(t.Context())
			ctxMock = context.WithValue(ctxMock, config.ContextToken, "sometoken")
			clientsMock := shared.NewClientsMock()
			// Create clients that is mocked for testing
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			// Setup custom per-test mocks (higher priority than default mocks)
			if tc.Setup != nil {
				ctxMock = tc.Setup(t, ctxMock, clientsMock)
			}
			// Setup generic test suite mocks
			clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
			// Setup default mock actions
			clientsMock.AddDefaultMocks()

			err := Activity(ctxMock, clients, tc.Args)
			if tc.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedError.Error(), err)
			} else {
				require.NoError(t, err)
			}
			// Assert mocks or other custom assertions
			if tc.ExpectedAsserts != nil {
				tc.ExpectedAsserts(t, ctxMock, clientsMock)
			}
		})
	}
}

func TestPlatformActivity_TriggerExecutedToString(t *testing.T) {
	for name, tc := range map[string]struct {
		Activity        api.Activity
		ExpectedResults []string
	}{
		"successful trigger trip with trigger information": {
			Activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]any{
					"function_name": "Send a greeting",
					"trigger": map[string]any{
						"type": "shortcut",
					},
				},
			},
			ExpectedResults: []string{
				"Shortcut trigger successfully started execution of function 'Send a greeting'",
			},
		},
		"successful trigger trip without trigger information": {
			Activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]any{
					"function_name": "Send a greeting",
				},
			},
			ExpectedResults: []string{
				"Trigger successfully started execution of function 'Send a greeting'",
			},
		},
		"reason 'parameter_validation_failed' with 1 error": {
			Activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]any{
					"function_name": "Send a greeting",
					"reason":        "parameter_validation_failed",
					"errors":        "[\"Null value for non-nullable parameter `channel`\"]",
				},
			},
			ExpectedResults: []string{
				"Trigger for workflow 'Send a greeting' failed: parameter_validation_failed",
				"- Null value for non-nullable parameter `channel`",
			},
		},
		"reason 'parameter_validation_failed' with 2 errors": {
			Activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]any{
					"function_name": "Send a greeting",
					"reason":        "parameter_validation_failed",
					"errors":        "[\"Null value for non-nullable parameter `channel`\",\"Null value for non-nullable parameter `interactivity`\"]",
				},
			},
			ExpectedResults: []string{
				"Trigger for workflow 'Send a greeting' failed: parameter_validation_failed",
				"- Null value for non-nullable parameter `channel`",
				"- Null value for non-nullable parameter `interactivity`",
			},
		},
		"reason 'parameter_validation_failed' with nil errors": {
			Activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]any{
					"function_name": "Send a greeting",
					"reason":        "parameter_validation_failed",
					"errors":        nil,
				},
			},
			ExpectedResults: []string{
				"Trigger for workflow 'Send a greeting' failed: parameter_validation_failed",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actualResult := triggerExecutedToString(tc.Activity)
			for _, expectedResult := range tc.ExpectedResults {
				require.Contains(t, actualResult, expectedResult)
			}
			// Confirm no nil pointers leak to output
			require.NotContains(t, actualResult, "nil")
			require.NotContains(t, actualResult, "NIL")
		})
	}
}

func Test_activityToStringFunctions(t *testing.T) {
	baseActivity := api.Activity{
		TraceID:     "tr123",
		Level:       types.INFO,
		ComponentID: "comp1",
		Created:     1686939542000000,
		Payload: map[string]any{
			"function_id":   "fn1",
			"function_name": "MyFunc",
			"function_type": "custom",
			"workflow_name": "MyWorkflow",
			"user_id":       "U123",
			"team_id":       "T456",
			"app_id":        "A789",
			"provider_key":  "google",
			"code":          "auth_error",
			"extra_message": "details here",
			"action":        "deploy",
			"log":           "some output",
			"channel_id":    "C123",
			"bot_user_id":   "B456",
			"template_id":   "tmpl1",
			"current_step":  float64(2),
			"total_steps":   float64(5),
		},
	}

	tests := map[string]struct {
		fn       func(api.Activity) string
		activity api.Activity
		contains []string
	}{
		"externalAuthMissingFunctionToString": {
			fn:       externalAuthMissingFunctionToString,
			activity: baseActivity,
			contains: []string{"fn1", "missing"},
		},
		"externalAuthMissingSelectedAuthToString": {
			fn:       externalAuthMissingSelectedAuthToString,
			activity: baseActivity,
			contains: []string{"auth_error", "Missing mapped token"},
		},
		"externalAuthResultToString info": {
			fn:       externalAuthResultToString,
			activity: baseActivity,
			contains: []string{"completed", "U123", "T456"},
		},
		"externalAuthStartedToString info": {
			fn:       externalAuthStartedToString,
			activity: baseActivity,
			contains: []string{"succeeded", "U123", "T456"},
		},
		"externalAuthTokenFetchResult info": {
			fn:       externalAuthTokenFetchResult,
			activity: baseActivity,
			contains: []string{"succeeded", "U123", "T456"},
		},
		"functionDeploymentToString": {
			fn:       functionDeploymentToString,
			activity: baseActivity,
			contains: []string{"deployd", "U123", "T456"},
		},
		"functionExecutionOutputToString": {
			fn:       functionExecutionOutputToString,
			activity: baseActivity,
			contains: []string{"Function output", "some output"},
		},
		"triggerPayloadReceivedOutputToString": {
			fn:       triggerPayloadReceivedOutputToString,
			activity: baseActivity,
			contains: []string{"Trigger payload", "some output"},
		},
		"functionExecutionResultToString completed": {
			fn:       functionExecutionResultToString,
			activity: baseActivity,
			contains: []string{"MyFunc", "completed"},
		},
		"functionExecutionResultToString failed": {
			fn: functionExecutionResultToString,
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"function_name": "MyFunc",
					"function_type": "custom",
					"error":         "something went wrong",
				},
			},
			contains: []string{"MyFunc", "failed", "something went wrong"},
		},
		"functionExecutionStartedToString": {
			fn:       functionExecutionStartedToString,
			activity: baseActivity,
			contains: []string{"MyFunc", "started"},
		},
		"workflowBillingResultToString with billing": {
			fn: workflowBillingResultToString,
			activity: api.Activity{
				Level:       types.INFO,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"workflow_name":     "MyWorkflow",
					"is_billing_result": true,
					"billing_reason":    "function_execution",
				},
			},
			contains: []string{"MyWorkflow", "billing reason", "function_execution"},
		},
		"workflowBillingResultToString excluded": {
			fn: workflowBillingResultToString,
			activity: api.Activity{
				Level:       types.INFO,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"is_billing_result": false,
				},
			},
			contains: []string{"excluded from billing"},
		},
		"workflowBotInvitedToString": {
			fn:       workflowBotInvitedToString,
			activity: baseActivity,
			contains: []string{"C123", "B456", "invited"},
		},
		"workflowCreatedFromTemplateToString": {
			fn:       workflowCreatedFromTemplateToString,
			activity: baseActivity,
			contains: []string{"MyWorkflow", "tmpl1"},
		},
		"workflowExecutionResultToString completed": {
			fn:       workflowExecutionResultToString,
			activity: baseActivity,
			contains: []string{"MyWorkflow", "completed"},
		},
		"workflowExecutionResultToString failed": {
			fn: workflowExecutionResultToString,
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"workflow_name": "MyWorkflow",
					"error":         "workflow error",
				},
			},
			contains: []string{"MyWorkflow", "failed", "workflow error"},
		},
		"workflowExecutionStartedToString": {
			fn:       workflowExecutionStartedToString,
			activity: baseActivity,
			contains: []string{"MyWorkflow", "started"},
		},
		"workflowPublishedToString": {
			fn:       workflowPublishedToString,
			activity: baseActivity,
			contains: []string{"MyWorkflow", "published"},
		},
		"externalAuthResultToString error": {
			fn: externalAuthResultToString,
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"user_id":       "U123",
					"team_id":       "T456",
					"app_id":        "A789",
					"provider_key":  "google",
					"code":          "auth_error",
					"extra_message": "details here",
				},
			},
			contains: []string{"failed", "U123", "T456", "auth_error", "details here"},
		},
		"externalAuthStartedToString error": {
			fn: externalAuthStartedToString,
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"user_id":      "U123",
					"team_id":      "T456",
					"app_id":       "A789",
					"provider_key": "google",
					"code":         "auth_start_error",
				},
			},
			contains: []string{"failed", "U123", "auth_start_error"},
		},
		"externalAuthTokenFetchResult error": {
			fn: externalAuthTokenFetchResult,
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"user_id":      "U123",
					"team_id":      "T456",
					"app_id":       "A789",
					"provider_key": "google",
					"code":         "fetch_error",
				},
			},
			contains: []string{"failed", "U123", "fetch_error"},
		},
		"workflowStepExecutionResultToString completed": {
			fn:       workflowStepExecutionResultToString,
			activity: baseActivity,
			contains: []string{"MyFunc", "completed"},
		},
		"workflowStepExecutionResultToString failed": {
			fn: workflowStepExecutionResultToString,
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "comp1",
				TraceID:     "tr1",
				Payload: map[string]any{
					"function_name": "MyFunc",
					"function_type": "custom",
					"error":         "step error",
				},
			},
			contains: []string{"MyFunc", "failed"},
		},
		"workflowStepStartedToString": {
			fn:       workflowStepStartedToString,
			activity: baseActivity,
			contains: []string{"2", "5", "started"},
		},
		"workflowUnpublishedToString": {
			fn:       workflowUnpublishedToString,
			activity: baseActivity,
			contains: []string{"MyWorkflow", "unpublished"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.fn(tc.activity)
			for _, s := range tc.contains {
				assert.Contains(t, result, s)
			}
		})
	}
}

func Test_datastoreRequestResultToString(t *testing.T) {
	for name, tc := range map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"successful datastore request event log": {
			activity: api.Activity{
				Payload: map[string]any{
					"datastore_name": "MyDatastore",
					"request_type":   "get",
					"details":        "id: f7d1253f-4066-4b83-8330-a483ff555c20",
				},
			},
			expectedResults: []string{"MyDatastore", "get", "succeeded", "f7d1253f-4066-4b83-8330-a483ff555c20"},
		},
		"successful datastore request event log with nil payload": {
			activity:        api.Activity{},
			expectedResults: []string{"succeeded"},
		},
		"failed datastore request error log": {
			activity: api.Activity{
				Level: "error",
				Payload: map[string]any{
					"datastore_name": "MyDatastore",
					"request_type":   "query",
					"details":        `{"expression": "id invalid_operator f7d1253f-4066-4b83-8330-a483ff555c20"}`,
					"error":          "ValidationException",
				},
			},
			expectedResults: []string{"MyDatastore", "query", "failed", "ValidationException", "f7d1253f-4066-4b83-8330-a483ff555c20"},
		},
		"failed datastore request without error field": {
			activity: api.Activity{
				Level: "error",
				Payload: map[string]any{
					"datastore_name": "MyDatastore",
					"request_type":   "query",
					"details":        `{"expression": "id invalid_operator f7d1253f-4066-4b83-8330-a483ff555c20"}`,
				},
			},
			expectedResults: []string{"MyDatastore", "query", "failed", "f7d1253f-4066-4b83-8330-a483ff555c20"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			actualResult := datastoreRequestResultToString(tc.activity)
			for _, expectedResult := range tc.expectedResults {
				require.Contains(t, actualResult, expectedResult)
			}
			// Confirm no nil pointers leak to output
			require.NotContains(t, actualResult, "nil")
			require.NotContains(t, actualResult, "NIL")
		})
	}
}
