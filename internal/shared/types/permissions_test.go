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
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func Test_Permissions_StringToAppCollaboratorPermission(t *testing.T) {
	tests := []struct {
		name                              string
		input                             string
		expectedErrorType                 error
		expectedAppCollaboratorPermission AppCollaboratorPermission
	}{
		{
			name:                              "owner",
			input:                             "owner",
			expectedErrorType:                 nil,
			expectedAppCollaboratorPermission: OWNER,
		},
		{
			name:                              "reader",
			input:                             "reader",
			expectedErrorType:                 nil,
			expectedAppCollaboratorPermission: READER,
		},
		{
			name:                              "default",
			input:                             "",
			expectedErrorType:                 fmt.Errorf("invalid"),
			expectedAppCollaboratorPermission: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appCollaboratorPermission, err := StringToAppCollaboratorPermission(tt.input)

			require.IsType(t, err, tt.expectedErrorType)
			require.Equal(t, tt.expectedAppCollaboratorPermission, appCollaboratorPermission)
		})
	}
}

func Test_Permissions_AppCollaboratorPermissionF(t *testing.T) {
	tests := []struct {
		name           string
		acp            AppCollaboratorPermission
		expectedString string
	}{
		{
			name:           "owner",
			acp:            OWNER,
			expectedString: "an owner collaborator",
		},
		{
			name:           "reader",
			acp:            READER,
			expectedString: "a reader collaborator",
		},
		{
			name:           "default",
			acp:            "",
			expectedString: "a collaborator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedString := tt.acp.AppCollaboratorPermissionF()

			require.Equal(t, tt.expectedString, returnedString)
		})
	}
}

func Test_Permission_IsValid(t *testing.T) {
	tests := []struct {
		name            string
		permission      Permission
		expectedIsValid bool
	}{
		{
			name:            "Permission App Collaborators",
			permission:      PermissionAppCollaborators,
			expectedIsValid: true,
		},
		{
			name:            "Permission Everyone",
			permission:      PermissionEveryone,
			expectedIsValid: true,
		},
		{
			name:            "Permission Named Entities",
			permission:      PermissionNamedEntities,
			expectedIsValid: true,
		},
		{
			name:            "Invalid empty",
			permission:      "",
			expectedIsValid: false,
		},
		{
			name:            "Invalid whitespace",
			permission:      "  \n  ",
			expectedIsValid: false,
		},
		{
			name:            "Invalid string",
			permission:      "cats_and_dogs",
			expectedIsValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedIsValid := tt.permission.IsValid()

			require.Equal(t, tt.expectedIsValid, returnedIsValid)
		})
	}
}

func Test_Permission_ToString(t *testing.T) {
	tests := []struct {
		name           string
		permission     Permission
		expectedString string
	}{
		{
			name:           "Permission App Collaborators",
			permission:     PermissionAppCollaborators,
			expectedString: "app collaborators",
		},
		{
			name:           "Permission Everyone",
			permission:     PermissionEveryone,
			expectedString: "everyone",
		},
		{
			name:           "Permission Named Entities",
			permission:     PermissionNamedEntities,
			expectedString: "specific entities",
		},
		{
			name:           "Invalid string",
			permission:     "cats_and_dogs",
			expectedString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedString := tt.permission.ToString()

			require.Equal(t, tt.expectedString, returnedString)
		})
	}
}

func Test_Permissions_IsNamedEntityFlag(t *testing.T) {
	type TestFlags struct {
		channelsValue      string
		everyoneValue      bool
		organizationsValue string
		usersValue         string
		workspacesValue    string
	}

	tests := []struct {
		name                      string
		flags                     TestFlags
		expectedIsNamedEntityFlag bool
	}{
		{
			name: "--channels flag is set",
			flags: TestFlags{
				channelsValue: "C1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		{
			name: "--organizations flag is set",
			flags: TestFlags{
				organizationsValue: "E1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		{
			name: "--users flag is set",
			flags: TestFlags{
				usersValue: "U1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		{
			name: "--workspaces flag is set",
			flags: TestFlags{
				workspacesValue: "T1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		{
			name: "Multiple valid entities are set",
			flags: TestFlags{
				channelsValue:   "C1234",
				workspacesValue: "T1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		{
			name: "--everyone is set",
			flags: TestFlags{
				everyoneValue: true,
			},
			expectedIsNamedEntityFlag: false,
		},
		{
			name: "--everyone and a valid entitle are set",
			flags: TestFlags{
				everyoneValue: true,
				channelsValue: "C1234",
			},
			expectedIsNamedEntityFlag: false,
		},
	}
	for _, tt := range tests {
		flags := pflag.NewFlagSet("entities", pflag.ContinueOnError)

		// Define the flags
		flags.StringP("channels", "", "", "channels usage")
		flags.BoolP("everyone", "", false, "everyone usage")
		flags.StringP("organizations", "", "", "organizations usage")
		flags.StringP("users", "", "", "users usage")
		flags.StringP("workspaces", "", "", "workspaces usage")

		// Set the flags based on the test values
		everyoneFlag := ""
		if tt.flags.everyoneValue {
			everyoneFlag = "--everyone"
		}

		args := []string{
			everyoneFlag,
			fmt.Sprintf("--channels=%s", tt.flags.channelsValue),
			fmt.Sprintf("--organizations=%s", tt.flags.organizationsValue),
			fmt.Sprintf("--users=%s", tt.flags.usersValue),
			fmt.Sprintf("--workspaces=%s", tt.flags.workspacesValue),
		}

		// Parse the flagset
		if err := flags.Parse(args); err != nil {
			require.Fail(t, err.Error(), "Flags parse error")
		}

		t.Run(tt.name, func(t *testing.T) {
			returnedIsNamedEntityFlag := IsNamedEntityFlag(flags)
			require.Equal(t, tt.expectedIsNamedEntityFlag, returnedIsNamedEntityFlag)
		})
	}
}

func Test_Permissions_GetAccessTypeDescriptionForEveryone(t *testing.T) {
	tests := []struct {
		name           string
		app            App
		expectedString string
	}{
		{
			name: "Enterprise App",
			app: App{ // app.IsEnterpriseApp() == true
				AppID:        "A1234",
				TeamID:       "E1234",
				EnterpriseID: "E1234",
			},
			expectedString: "everyone in all workspaces in this org granted to this app",
		},
		{
			name: "Workspace App",
			app: App{ // app.IsEnterpriseApp() == false
				AppID:  "A1234",
				TeamID: "T1234",
			},
			expectedString: "everyone in the workspace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedString := GetAccessTypeDescriptionForEveryone(tt.app)
			require.Equal(t, tt.expectedString, returnedString)
		})
	}
}
