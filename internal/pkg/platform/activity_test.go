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

package platform

import (
	"context"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_prettifyActivity(t *testing.T) {
	tests := []struct {
		name            string
		activity        api.Activity
		expectedResults []string
	}{
		{
			name: "nil payload should result in valid log without nulls",
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
		{
			name:     "empty activity should not contain nulls",
			activity: api.Activity{},
		},
		{
			name: "unknown EventType should result in valid log without nulls",
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
	for name, tt := range map[string]struct {
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
			if tt.Setup != nil {
				ctxMock = tt.Setup(t, ctxMock, clientsMock)
			}
			// Setup generic test suite mocks
			clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(api.AuthSession{}, nil)
			// Setup default mock actions
			clientsMock.AddDefaultMocks()

			err := Activity(ctxMock, clients, &logger.Logger{}, tt.Args)
			if tt.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.ExpectedError.Error(), err)
			} else {
				require.NoError(t, err)
			}
			// Assert mocks or other custom assertions
			if tt.ExpectedAsserts != nil {
				tt.ExpectedAsserts(t, ctxMock, clientsMock)
			}
		})
	}
}

func TestPlatformActivity_TriggerExecutedToString(t *testing.T) {
	for name, tt := range map[string]struct {
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
			actualResult := triggerExecutedToString(tt.Activity)
			for _, expectedResult := range tt.ExpectedResults {
				require.Contains(t, actualResult, expectedResult)
			}
			// Confirm no nil pointers leak to output
			require.NotContains(t, actualResult, "nil")
			require.NotContains(t, actualResult, "NIL")
		})
	}
}

func Test_datastoreRequestResultToString(t *testing.T) {
	for name, tt := range map[string]struct {
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
			actualResult := datastoreRequestResultToString(tt.activity)
			for _, expectedResult := range tt.expectedResults {
				require.Contains(t, actualResult, expectedResult)
			}
			// Confirm no nil pointers leak to output
			require.NotContains(t, actualResult, "nil")
			require.NotContains(t, actualResult, "NIL")
		})
	}
}
