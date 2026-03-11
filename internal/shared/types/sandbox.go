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
	DateArchived  int64  `json:"date_archived"`   // When the developer sandbox is or will be archived, as epoch seconds
	DateCreated   int64  `json:"date_created"`    // When the developer sandbox was created, as epoch seconds
	SandboxDomain string `json:"sandbox_domain"`  // Domain of the developer sandbox
	SandboxName   string `json:"sandbox_name"`    // Name of the developer sandbox
	SandboxTeamID string `json:"sandbox_team_id"` // Encoded team ID of the developer sandbox
	Status        string `json:"status"`          // Status of the developer sandbox: Active or Archived
}
