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

package tracking

type EventType string

const (
	// Error event is raised when an error occurs and the slack CLI will exit with a non-zero exit code.
	Error EventType = "error"
	// Interrupt event occurs when a terminal signal interrupts the CLI process.
	Interrupt EventType = "interrupt"
	// Success event occurs when the CLI intends to exit with a zero exit code.
	Success EventType = "success"
)

type LogstashEvent struct {
	Context   EventContext `json:"context"`
	Data      EventData    `json:"data,omitempty"`
	Event     EventType    `json:"event"`
	Timestamp int64        `json:"timestamp"`
}

// EventData contains logging/metrics data related to this CLI process invocation. DO NOT use this for business logic!
// Try to avoid changing fields in this struct, as this makes long-term data analyses difficult!
// If you add a field here, make sure it has `omitempty` and if it needs PII redaction or similar cleaning operations, implement that in tracking.go's CleanSessionData method
type EventData struct {
	// App holds information about the application being operated on by the CLI
	App AppEventData `json:"app,omitempty"`
	// Auth holds information about the user's authentication
	Auth AuthEventData `json:"auth,omitempty"`
	// ErrorCode holds information about a specific error raised during the CLI session
	ErrorCode string `json:"error_code,omitempty"`
	// ErrorMessage holds information about any errors raised during the CLI session
	ErrorMessage string `json:"error_msg,omitempty"`
}

// EventContext contains information / metadata about the CLI session
type EventContext struct {
	Binary           string   `json:"bin"`
	CLIVersion       string   `json:"cli_version"`
	Command          string   `json:"command,omitempty"`
	CommandCanonical string   `json:"command_canonical,omitempty"`
	Flags            []string `json:"flags,omitempty"`
	Host             string   `json:"host"`
	OS               string   `json:"os"`
	ProjectID        string   `json:"project_id,omitempty"`
	Runtime          string   `json:"runtime,omitempty"`
	RuntimeVersion   string   `json:"runtime_version,omitempty"`
	SessionID        string   `json:"session_id"`
	SystemID         string   `json:"system_id"`
}

// AuthEventData holds information about the user who is using the CLI's authentication
type AuthEventData struct {
	EnterpriseID string `json:"enterprise_id,omitempty"`
	TeamID       string `json:"team_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
}

// AppEventData holds information about the Slack App being interacted with via the CLI
type AppEventData struct {
	EnterpriseID string `json:"enterprise_id,omitempty"`
	TeamID       string `json:"team_id,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	// Template holds information for the sample app template used as part of the `create` command
	Template string `json:"template,omitempty"`
}
