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

package types

type ActivityArgs struct {
	TeamID            string
	AppID             string
	TailArg           bool
	Browser           bool
	PollingIntervalMS int
	IdleTimeoutM      int
	Limit             int
	MinDateCreated    int64
	MaxDateCreated    int64
	MinLevel          string
	EventType         string
	ComponentType     string
	ComponentID       string
	Source            string
	TraceID           string
}

type ActivityLevel string

const (
	TRACE ActivityLevel = "trace"
	DEBUG ActivityLevel = "debug"
	INFO  ActivityLevel = "info"
	WARN  ActivityLevel = "warn"
	ERROR ActivityLevel = "error"
	FATAL ActivityLevel = "fatal"
)

type EventType string

const (
	DatastoreRequestResult          EventType = "datastore_request_result"
	ExternalAuthMissingFunction     EventType = "external_auth_missing_function"
	ExternalAuthMissingSelectedAuth EventType = "external_auth_missing_oauth_token_or_selected_auth"
	ExternalAuthResult              EventType = "external_auth_result"
	ExternalAuthStarted             EventType = "external_auth_started"
	ExternalAuthTokenFetchResult    EventType = "external_auth_token_fetch_result"
	FunctionDeployment              EventType = "function_deployment"
	FunctionExecutionOutput         EventType = "function_execution_output"
	FunctionExecutionResult         EventType = "function_execution_result"
	FunctionExecutionStarted        EventType = "function_execution_started"
	TriggerExecuted                 EventType = "trigger_executed"
	TriggerPayloadReceived          EventType = "trigger_payload_received"
	WorkflowBillingResult           EventType = "workflow_billing_result"
	WorkflowBotInvited              EventType = "workflow_bot_invited"
	WorkflowCreatedFromTemplate     EventType = "workflow_created_from_template"
	WorkflowExecutionResult         EventType = "workflow_execution_result"
	WorkflowExecutionStarted        EventType = "workflow_execution_started"
	WorkflowPublished               EventType = "workflow_published"
	WorkflowStepExecutionResult     EventType = "workflow_step_execution_result"
	WorkflowStepStarted             EventType = "workflow_step_started"
	WorkflowUnpublished             EventType = "workflow_unpublished"
)

type ActivityRequest struct {
	AppID              string `json:"app_id"`
	NextCursor         string `json:"cursor,omitempty"`
	Limit              int    `json:"limit,omitempty"`
	MinimumDateCreated int64  `json:"min_date_created,omitempty"`
	MaximumDateCreated int64  `json:"max_date_created,omitempty"`
	MinimumLogLevel    string `json:"min_log_level,omitempty"`
	EventType          string `json:"log_event_type,omitempty"`
	ComponentType      string `json:"component_type,omitempty"`
	ComponentID        string `json:"component_id,omitempty"`
	Source             string `json:"source,omitempty"`
	TraceID            string `json:"trace_id,omitempty"`
}
