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

package types

// Sandbox represents a Slack Developer Sandbox from the developer.sandbox.list API.
type Sandbox struct {
	SandboxTeamID string `json:"sandbox_team_id"` // Encoded team ID of the developer sandbox
	SandboxName   string `json:"sandbox_name"`    // Name of the developer sandbox
	SandboxDomain string `json:"sandbox_domain"`  // Domain of the developer sandbox
	DateCreated   int64  `json:"date_created"`    // When the developer sandbox was created, as epoch seconds
	DateArchived  int64  `json:"date_archived"`   // When the developer sandbox is or will be archived, as epoch seconds
	Status        string `json:"status"`          // Status of the developer sandbox: Active or Archived
}

// CreateSandboxRequest is the request payload for creating a sandbox.
// Matches enterprise.signup.createDevOrg API contract.
type CreateSandboxRequest struct {
	Token       string `json:"token"`
	OrgName     string `json:"org_name"`
	Domain      string `json:"domain"`
	Password    string `json:"password,omitempty"`
	Locale      string `json:"locale,omitempty"`
	OwningOrgID string `json:"owning_org_id,omitempty"`
	TemplateID  string `json:"template_id,omitempty"`
	EventCode   string `json:"event_code,omitempty"`
	ArchiveDate int64  `json:"archive_date,omitempty"` // When the sandbox will be archived, as epoch seconds
}

// CreateSandboxResult is the response from creating a sandbox.
// Matches enterprise.signup.createDevOrg API output.
type CreateSandboxResult struct {
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
	URL    string `json:"url"`
}
