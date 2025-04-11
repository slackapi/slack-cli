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

import (
	"fmt"
	"strings"
)

type SlackTeam struct {
	ID       string `json:"team_id,omitempty" yaml:"team_id,omitempty"`
	TeamName string `json:"team_name,omitempty" yaml:"team_name,omitempty"`
}

func (c *SlackTeam) String() string {
	if c.ID != "" && c.TeamName != "" {
		return fmt.Sprintf("%s (%s)", c.TeamName, c.ID)
	} else if c.TeamName != "" {
		return fmt.Sprintf("(%s)", c.TeamName)
	}
	return c.TeamName
}

// Team model with fields that match what is returned from the team.info method
// Method documentation: https://docs.slack.dev/reference/methods/team.info
type TeamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IsTeamID returns true if matches the pattern of a workspace (T-prefixed) team ID
// Note: This criteria is an estimate, and not directly related to server-side team_id scheme
func IsWorkspaceTeamID(str string) bool {
	return strings.HasPrefix(str, "T") && strings.ToUpper(str) == str
}

// IsEnterpriseTeamID returns true if a string matches the pattern of an enterprise (E-prefixed) team ID
// Note: Validation criteria is an estimate, and not directly related to server-side team_id scheme
func IsEnterpriseTeamID(str string) bool {
	return strings.HasPrefix(str, "E") && strings.ToUpper(str) == str
}

// IsTeamID returns true if a string matches the pattern of a team id
// Note: Validation criteria is an estimate, and not directly related to server-side team_id scheme
func IsTeamID(str string) bool {
	return IsEnterpriseTeamID(str) || IsWorkspaceTeamID(str)
}
