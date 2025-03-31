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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SlackTeam_String(t *testing.T) {
	tests := []struct {
		name           string
		slackTeam      *SlackTeam
		expectedString string
	}{
		{
			name:           "ID and TeamName exists",
			slackTeam:      &SlackTeam{ID: "T1234", TeamName: "Team Cats"},
			expectedString: "Team Cats (T1234)",
		},
		{
			name:           "Only TeamName exists",
			slackTeam:      &SlackTeam{ID: "", TeamName: "Team Cats"},
			expectedString: "(Team Cats)", // TODO - confirm that this is the actual format that's desired
		},
		{
			name:           "Only ID exists",
			slackTeam:      &SlackTeam{ID: "T1234", TeamName: ""},
			expectedString: "", // TODO - confirm that this is the actual format that's desired
		},
		{
			name:           "None exists",
			slackTeam:      &SlackTeam{ID: "", TeamName: ""},
			expectedString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedString := tt.slackTeam.String()
			require.Equal(t, tt.expectedString, returnedString)
		})
	}
}

func Test_SlackTeam_Team_Workspace_Enterprise_Identifiers(t *testing.T) {
	var validEnterpriseID = "E013A7H7BBL"
	var validWorkspaceTeamID = "T014W0LPQNN"

	require.Equal(t, true, IsEnterpriseTeamID(validEnterpriseID), "Enterprise ID should pass as enterprise team ID")
	require.Equal(t, false, IsEnterpriseTeamID(validWorkspaceTeamID), "Workspace team ID should NOT pass as enterprise team ID")

	require.Equal(t, true, IsWorkspaceTeamID(validWorkspaceTeamID), "Workspace team ID should pass as workspace team ID")
	require.Equal(t, false, IsWorkspaceTeamID(validEnterpriseID), "Enterprise ID should NOT pass as workspace team ID")

	require.Equal(t, true, IsTeamID(validEnterpriseID), "Enterprise ID should pass as a team ID")
	require.Equal(t, true, IsTeamID(validWorkspaceTeamID), "Workspace ID should pass as team ID")
}
