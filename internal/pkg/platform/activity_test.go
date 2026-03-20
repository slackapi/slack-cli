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
				Payload: map[string]interface{}{
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
		"warn level activity should contain the message": {
			activity: api.Activity{
				TraceID:     "w123",
				Level:       types.WARN,
				EventType:   "unknown",
				ComponentID: "w789",
				Payload: map[string]interface{}{
					"some": "warning",
				},
				Created: 1686939542,
			},
			expectedResults: []string{
				`{"some":"warning"}`,
			},
		},
		"error level activity should contain the message": {
			activity: api.Activity{
				TraceID:     "e123",
				Level:       types.ERROR,
				EventType:   "unknown",
				ComponentID: "e789",
				Payload: map[string]interface{}{
					"some": "error",
				},
				Created: 1686939542,
			},
			expectedResults: []string{
				`{"some":"error"}`,
			},
		},
		"fatal level activity should contain the message": {
			activity: api.Activity{
				TraceID:     "f123",
				Level:       types.FATAL,
				EventType:   "unknown",
				ComponentID: "f789",
				Payload: map[string]interface{}{
					"some": "fatal",
				},
				Created: 1686939542,
			},
			expectedResults: []string{
				`{"some":"fatal"}`,
			},
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
				Payload: map[string]interface{}{
					"function_name": "Send a greeting",
					"trigger": map[string]interface{}{
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
				Payload: map[string]interface{}{
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
				Payload: map[string]interface{}{
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
				Payload: map[string]interface{}{
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
				Payload: map[string]interface{}{
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

func Test_prettifyActivity_allEventTypes(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"DatastoreRequestResult routes correctly": {
			activity: api.Activity{
				EventType: types.DatastoreRequestResult,
				Payload: map[string]interface{}{
					"request_type":   "get",
					"datastore_name": "DS1",
					"details":        "id: 123",
				},
			},
			expectedResults: []string{"Datastore get succeeded"},
		},
		"ExternalAuthMissingFunction routes correctly": {
			activity: api.Activity{
				EventType: types.ExternalAuthMissingFunction,
				Payload: map[string]interface{}{
					"function_id": "fn1",
				},
			},
			expectedResults: []string{"Step function 'fn1' is missing"},
		},
		"ExternalAuthMissingSelectedAuth routes correctly": {
			activity: api.Activity{
				EventType: types.ExternalAuthMissingSelectedAuth,
				Payload: map[string]interface{}{
					"code": "wf1",
				},
			},
			expectedResults: []string{"Missing mapped token for workflow 'wf1'"},
		},
		"ExternalAuthResult routes correctly": {
			activity: api.Activity{
				EventType: types.ExternalAuthResult,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"user_id":      "U1",
					"team_id":      "T1",
					"app_id":       "A1",
					"provider_key": "google",
				},
			},
			expectedResults: []string{"Auth completed"},
		},
		"ExternalAuthStarted routes correctly": {
			activity: api.Activity{
				EventType: types.ExternalAuthStarted,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"user_id":      "U1",
					"team_id":      "T1",
					"app_id":       "A1",
					"provider_key": "google",
				},
			},
			expectedResults: []string{"Auth start succeeded"},
		},
		"ExternalAuthTokenFetchResult routes correctly": {
			activity: api.Activity{
				EventType: types.ExternalAuthTokenFetchResult,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"user_id":      "U1",
					"team_id":      "T1",
					"app_id":       "A1",
					"provider_key": "google",
				},
			},
			expectedResults: []string{"Token fetch succeeded"},
		},
		"FunctionDeployment routes correctly": {
			activity: api.Activity{
				EventType: types.FunctionDeployment,
				Payload: map[string]interface{}{
					"action":  "deploye",
					"user_id": "U1",
					"team_id": "T1",
				},
				Created: 1686939542000000,
			},
			expectedResults: []string{"Application deployed"},
		},
		"FunctionExecutionOutput routes correctly": {
			activity: api.Activity{
				EventType: types.FunctionExecutionOutput,
				Payload: map[string]interface{}{
					"log": "output here",
				},
			},
			expectedResults: []string{"Function output:"},
		},
		"TriggerPayloadReceived routes correctly": {
			activity: api.Activity{
				EventType: types.TriggerPayloadReceived,
				Payload: map[string]interface{}{
					"log": "payload here",
				},
			},
			expectedResults: []string{"Trigger payload:"},
		},
		"FunctionExecutionResult routes correctly": {
			activity: api.Activity{
				EventType: types.FunctionExecutionResult,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"function_name": "fn1",
					"function_type": "custom",
				},
			},
			expectedResults: []string{"Function 'fn1' (custom function) completed"},
		},
		"FunctionExecutionStarted routes correctly": {
			activity: api.Activity{
				EventType: types.FunctionExecutionStarted,
				Payload: map[string]interface{}{
					"function_name": "fn1",
					"function_type": "custom",
				},
			},
			expectedResults: []string{"Function 'fn1' (custom function) started"},
		},
		"TriggerExecuted routes correctly": {
			activity: api.Activity{
				EventType: types.TriggerExecuted,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"function_name": "fn1",
				},
			},
			expectedResults: []string{"Trigger successfully started execution"},
		},
		"WorkflowBillingResult routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowBillingResult,
				Payload: map[string]interface{}{
					"workflow_name":     "wf1",
					"is_billing_result": false,
				},
			},
			expectedResults: []string{"excluded from billing"},
		},
		"WorkflowBotInvited routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowBotInvited,
				Payload: map[string]interface{}{
					"channel_id":  "C1",
					"bot_user_id": "B1",
				},
			},
			expectedResults: []string{"Channel C1 detected"},
		},
		"WorkflowCreatedFromTemplate routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowCreatedFromTemplate,
				Payload: map[string]interface{}{
					"workflow_name": "wf1",
					"template_id":   "tmpl1",
				},
			},
			expectedResults: []string{"Workflow 'wf1' created from template 'tmpl1'"},
		},
		"WorkflowExecutionResult routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowExecutionResult,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"workflow_name": "wf1",
				},
			},
			expectedResults: []string{"Workflow 'wf1' completed"},
		},
		"WorkflowExecutionStarted routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowExecutionStarted,
				Payload: map[string]interface{}{
					"workflow_name": "wf1",
				},
			},
			expectedResults: []string{"Workflow 'wf1' started"},
		},
		"WorkflowPublished routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowPublished,
				Payload: map[string]interface{}{
					"workflow_name": "wf1",
				},
			},
			expectedResults: []string{"Workflow 'wf1' published"},
		},
		"WorkflowStepExecutionResult routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowStepExecutionResult,
				Level:     types.INFO,
				Payload: map[string]interface{}{
					"function_name": "fn1",
				},
			},
			expectedResults: []string{"Workflow step 'fn1' completed"},
		},
		"WorkflowStepStarted routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowStepStarted,
				Payload: map[string]interface{}{
					"current_step": float64(1),
					"total_steps":  float64(3),
				},
			},
			expectedResults: []string{"Workflow step 1 of 3 started"},
		},
		"WorkflowUnpublished routes correctly": {
			activity: api.Activity{
				EventType: types.WorkflowUnpublished,
				Payload: map[string]interface{}{
					"workflow_name": "wf1",
				},
			},
			expectedResults: []string{"Workflow 'wf1' unpublished"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := prettifyActivity(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_externalAuthMissingFunctionToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.ERROR,
		ComponentID: "comp1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"function_id": "my_func",
		},
		Created: 1686939542000000,
	}
	result := externalAuthMissingFunctionToString(activity)
	assert.Contains(t, result, "Step function 'my_func' is missing")
	assert.Contains(t, result, "comp1")
	assert.Contains(t, result, "Trace=trace1")
}

func Test_externalAuthMissingSelectedAuthToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.ERROR,
		ComponentID: "comp1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"code": "wf_abc",
		},
		Created: 1686939542000000,
	}
	result := externalAuthMissingSelectedAuthToString(activity)
	assert.Contains(t, result, "Missing mapped token for workflow 'wf_abc'")
	assert.Contains(t, result, "Trace=trace1")
}

func Test_externalAuthResultToString(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"completed auth result": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"user_id":      "U123",
					"team_id":      "T123",
					"app_id":       "A123",
					"provider_key": "google",
				},
			},
			expectedResults: []string{
				"Auth completed",
				"U123",
				"T123",
				"A123",
				"google",
			},
		},
		"failed auth result with error details": {
			activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]interface{}{
					"user_id":       "U123",
					"team_id":       "T123",
					"app_id":        "A123",
					"provider_key":  "google",
					"code":          "invalid_grant",
					"extra_message": "token expired",
				},
			},
			expectedResults: []string{
				"Auth failed",
				"U123",
				"invalid_grant",
				"token expired",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := externalAuthResultToString(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_externalAuthStartedToString(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"succeeded auth start": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"user_id":      "U123",
					"team_id":      "T123",
					"app_id":       "A123",
					"provider_key": "google",
				},
			},
			expectedResults: []string{
				"Auth start succeeded",
				"U123",
				"T123",
				"A123",
				"google",
			},
		},
		"failed auth start with error code": {
			activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]interface{}{
					"user_id":      "U123",
					"team_id":      "T123",
					"app_id":       "A123",
					"provider_key": "google",
					"code":         "auth_failed",
				},
			},
			expectedResults: []string{
				"Auth start failed",
				"U123",
				"auth_failed",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := externalAuthStartedToString(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_externalAuthTokenFetchResult(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"succeeded token fetch": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"user_id":      "U123",
					"team_id":      "T123",
					"app_id":       "A123",
					"provider_key": "google",
				},
			},
			expectedResults: []string{
				"Token fetch succeeded",
				"U123",
				"T123",
				"A123",
				"google",
			},
		},
		"failed token fetch with error code": {
			activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]interface{}{
					"user_id":      "U123",
					"team_id":      "T123",
					"app_id":       "A123",
					"provider_key": "google",
					"code":         "token_revoked",
				},
			},
			expectedResults: []string{
				"Token fetch failed",
				"U123",
				"token_revoked",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := externalAuthTokenFetchResult(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_functionDeploymentToString(t *testing.T) {
	activity := api.Activity{
		Level: types.INFO,
		Payload: map[string]interface{}{
			"action":  "deploye",
			"user_id": "U123",
			"team_id": "T123",
		},
		Created: 1686939542000000,
	}
	result := functionDeploymentToString(activity)
	assert.Contains(t, result, "Application deployed")
	assert.Contains(t, result, "U123")
	assert.Contains(t, result, "T123")
}

func Test_functionExecutionOutputToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.INFO,
		ComponentID: "fn1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"log": "hello world",
		},
		Created: 1686939542000000,
	}
	result := functionExecutionOutputToString(activity)
	assert.Contains(t, result, "Function output:")
	assert.Contains(t, result, "hello world")
	assert.Contains(t, result, "Trace=trace1")
}

func Test_functionExecutionResultToString(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"completed function execution": {
			activity: api.Activity{
				Level:       types.INFO,
				ComponentID: "fn1",
				TraceID:     "trace1",
				Payload: map[string]interface{}{
					"function_name": "my_function",
					"function_type": "custom",
				},
				Created: 1686939542000000,
			},
			expectedResults: []string{
				"Function 'my_function' (custom function) completed",
				"Trace=trace1",
			},
		},
		"failed function execution with error": {
			activity: api.Activity{
				Level:       types.ERROR,
				ComponentID: "fn1",
				TraceID:     "trace1",
				Payload: map[string]interface{}{
					"function_name": "my_function",
					"function_type": "custom",
					"error":         "something went wrong",
				},
				Created: 1686939542000000,
			},
			expectedResults: []string{
				"Function 'my_function' (custom function) failed",
				"something went wrong",
			},
		},
		"fatal function execution": {
			activity: api.Activity{
				Level: types.FATAL,
				Payload: map[string]interface{}{
					"function_name": "my_function",
					"function_type": "builtin",
				},
			},
			expectedResults: []string{"Function 'my_function' (builtin function) failed"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := functionExecutionResultToString(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_functionExecutionStartedToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.INFO,
		ComponentID: "fn1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"function_name": "my_function",
			"function_type": "custom",
		},
		Created: 1686939542000000,
	}
	result := functionExecutionStartedToString(activity)
	assert.Contains(t, result, "Function 'my_function' (custom function) started")
	assert.Contains(t, result, "Trace=trace1")
}

func Test_triggerPayloadReceivedOutputToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.INFO,
		ComponentID: "trigger1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"log": "payload data here",
		},
		Created: 1686939542000000,
	}
	result := triggerPayloadReceivedOutputToString(activity)
	assert.Contains(t, result, "Trigger payload:")
	assert.Contains(t, result, "payload data here")
	assert.Contains(t, result, "Trace=trace1")
}

func Test_workflowBillingResultToString(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"billing result with workflow name": {
			activity: api.Activity{
				Level:       types.INFO,
				ComponentID: "wf1",
				TraceID:     "trace1",
				Payload: map[string]interface{}{
					"workflow_name":     "My Workflow",
					"is_billing_result": true,
					"billing_reason":    "execution",
				},
				Created: 1686939542000000,
			},
			expectedResults: []string{
				"Workflow 'My Workflow'",
				"billing reason 'execution'",
			},
		},
		"billing result without workflow name": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"is_billing_result": true,
					"billing_reason":    "execution",
				},
			},
			expectedResults: []string{
				"Workflow",
				"billing reason 'execution'",
			},
		},
		"excluded from billing": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"workflow_name":     "My Workflow",
					"is_billing_result": false,
				},
			},
			expectedResults: []string{
				"Workflow 'My Workflow'",
				"excluded from billing",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := workflowBillingResultToString(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_workflowBotInvitedToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.INFO,
		ComponentID: "wf1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"channel_id":  "C123",
			"bot_user_id": "B123",
		},
		Created: 1686939542000000,
	}
	result := workflowBotInvitedToString(activity)
	assert.Contains(t, result, "Channel C123 detected")
	assert.Contains(t, result, "Bot user B123 automatically invited")
}

func Test_workflowCreatedFromTemplateToString(t *testing.T) {
	activity := api.Activity{
		Level:       types.INFO,
		ComponentID: "wf1",
		TraceID:     "trace1",
		Payload: map[string]interface{}{
			"workflow_name": "My Workflow",
			"template_id":   "tmpl_123",
		},
		Created: 1686939542000000,
	}
	result := workflowCreatedFromTemplateToString(activity)
	assert.Contains(t, result, "Workflow 'My Workflow' created from template 'tmpl_123'")
}

func Test_workflowExecutionResultToString(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"completed workflow execution": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"workflow_name": "My Workflow",
				},
			},
			expectedResults: []string{"Workflow 'My Workflow' completed"},
		},
		"failed workflow execution with error": {
			activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]interface{}{
					"workflow_name": "My Workflow",
					"error":         "step failed",
				},
			},
			expectedResults: []string{
				"Workflow 'My Workflow' failed",
				"step failed",
			},
		},
		"fatal workflow execution": {
			activity: api.Activity{
				Level: types.FATAL,
				Payload: map[string]interface{}{
					"workflow_name": "My Workflow",
				},
			},
			expectedResults: []string{"Workflow 'My Workflow' failed"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := workflowExecutionResultToString(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_workflowExecutionStartedToString(t *testing.T) {
	activity := api.Activity{
		Level: types.INFO,
		Payload: map[string]interface{}{
			"workflow_name": "My Workflow",
		},
	}
	result := workflowExecutionStartedToString(activity)
	assert.Contains(t, result, "Workflow 'My Workflow' started")
}

func Test_workflowPublishedToString(t *testing.T) {
	activity := api.Activity{
		Level: types.INFO,
		Payload: map[string]interface{}{
			"workflow_name": "My Workflow",
		},
	}
	result := workflowPublishedToString(activity)
	assert.Contains(t, result, "Workflow 'My Workflow' published")
}

func Test_workflowStepExecutionResultToString(t *testing.T) {
	tests := map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"completed step": {
			activity: api.Activity{
				Level: types.INFO,
				Payload: map[string]interface{}{
					"function_name": "send_message",
				},
			},
			expectedResults: []string{"Workflow step 'send_message' completed"},
		},
		"failed step": {
			activity: api.Activity{
				Level: types.ERROR,
				Payload: map[string]interface{}{
					"function_name": "send_message",
				},
			},
			expectedResults: []string{"Workflow step 'send_message' failed"},
		},
		"fatal step": {
			activity: api.Activity{
				Level: types.FATAL,
				Payload: map[string]interface{}{
					"function_name": "send_message",
				},
			},
			expectedResults: []string{"Workflow step 'send_message' failed"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := workflowStepExecutionResultToString(tc.activity)
			for _, expected := range tc.expectedResults {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func Test_workflowStepStartedToString(t *testing.T) {
	activity := api.Activity{
		Level: types.INFO,
		Payload: map[string]interface{}{
			"current_step": float64(2),
			"total_steps":  float64(5),
		},
	}
	result := workflowStepStartedToString(activity)
	assert.Contains(t, result, "Workflow step 2 of 5 started")
}

func Test_workflowUnpublishedToString(t *testing.T) {
	activity := api.Activity{
		Level: types.INFO,
		Payload: map[string]interface{}{
			"workflow_name": "My Workflow",
		},
	}
	result := workflowUnpublishedToString(activity)
	assert.Contains(t, result, "Workflow 'My Workflow' unpublished")
}

func Test_datastoreRequestResultToString(t *testing.T) {
	for name, tc := range map[string]struct {
		activity        api.Activity
		expectedResults []string
	}{
		"successful datastore request event log": {
			activity: api.Activity{
				Payload: map[string]interface{}{
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
				Payload: map[string]interface{}{
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
				Payload: map[string]interface{}{
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
