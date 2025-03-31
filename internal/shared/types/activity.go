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
	TeamId            string
	AppId             string
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
	ComponentId       string
	Source            string
	TraceId           string
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
	DATASTORE_REQUEST_RESULT            EventType = "datastore_request_result"
	EXTERNAL_AUTH_MISSING_FUNCTION      EventType = "external_auth_missing_function"
	EXTERNAL_AUTH_MISSING_SELECTED_AUTH EventType = "external_auth_missing_oauth_token_or_selected_auth"
	EXTERNAL_AUTH_RESULT                EventType = "external_auth_result"
	EXTERNAL_AUTH_STARTED               EventType = "external_auth_started"
	EXTERNAL_AUTH_TOKEN_FETCH_RESULT    EventType = "external_auth_token_fetch_result"
	FUNCTION_DEPLOYMENT                 EventType = "function_deployment"
	FUNCTION_EXECUTION_OUTPUT           EventType = "function_execution_output"
	TRIGGER_PAYLOAD_RECEIVED            EventType = "trigger_payload_received"
	FUNCTION_EXECUTION_RESULT           EventType = "function_execution_result"
	FUNCTION_EXECUTION_STARTED          EventType = "function_execution_started"
	TRIGGER_EXECUTED                    EventType = "trigger_executed"
	WORKFLOW_BILLING_RESULT             EventType = "workflow_billing_result"
	WORKFLOW_BOT_INVITED                EventType = "workflow_bot_invited"
	WORKFLOW_CREATED_FROM_TEMPLATE      EventType = "workflow_created_from_template"
	WORKFLOW_EXECUTION_RESULT           EventType = "workflow_execution_result"
	WORKFLOW_EXECUTION_STARTED          EventType = "workflow_execution_started"
	WORKFLOW_PUBLISHED                  EventType = "workflow_published"
	WORKFLOW_STEP_EXECUTION_RESULT      EventType = "workflow_step_execution_result"
	WORKFLOW_STEP_STARTED               EventType = "workflow_step_started"
	WORKFLOW_UNPUBLISHED                EventType = "workflow_unpublished"
)

type ActivityRequest struct {
	AppId              string `json:"app_id"`
	NextCursor         string `json:"cursor,omitempty"`
	Limit              int    `json:"limit,omitempty"`
	MinimumDateCreated int64  `json:"min_date_created,omitempty"`
	MaximumDateCreated int64  `json:"max_date_created,omitempty"`
	MinimumLogLevel    string `json:"min_log_level,omitempty"`
	EventType          string `json:"log_event_type,omitempty"`
	ComponentType      string `json:"component_type,omitempty"`
	ComponentId        string `json:"component_id,omitempty"`
	Source             string `json:"source,omitempty"`
	TraceId            string `json:"trace_id,omitempty"`
}
