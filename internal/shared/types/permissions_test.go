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

import (
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func Test_Permissions_StringToAppCollaboratorPermission(t *testing.T) {
	tests := map[string]struct {
		input                             string
		expectedErrorType                 error
		expectedAppCollaboratorPermission AppCollaboratorPermission
	}{
		"owner": {
			input:                             "owner",
			expectedErrorType:                 nil,
			expectedAppCollaboratorPermission: OWNER,
		},
		"reader": {
			input:                             "reader",
			expectedErrorType:                 nil,
			expectedAppCollaboratorPermission: READER,
		},
		"default": {
			input:                             "",
			expectedErrorType:                 fmt.Errorf("invalid"),
			expectedAppCollaboratorPermission: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			appCollaboratorPermission, err := StringToAppCollaboratorPermission(tc.input)

			require.IsType(t, err, tc.expectedErrorType)
			require.Equal(t, tc.expectedAppCollaboratorPermission, appCollaboratorPermission)
		})
	}
}

func Test_Permissions_AppCollaboratorPermissionF(t *testing.T) {
	tests := map[string]struct {
		acp            AppCollaboratorPermission
		expectedString string
	}{
		"owner": {
			acp:            OWNER,
			expectedString: "an owner collaborator",
		},
		"reader": {
			acp:            READER,
			expectedString: "a reader collaborator",
		},
		"default": {
			acp:            "",
			expectedString: "a collaborator",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedString := tc.acp.AppCollaboratorPermissionF()

			require.Equal(t, tc.expectedString, returnedString)
		})
	}
}

func Test_Permission_IsValid(t *testing.T) {
	tests := map[string]struct {
		permission      Permission
		expectedIsValid bool
	}{
		"Permission App Collaborators": {
			permission:      PermissionAppCollaborators,
			expectedIsValid: true,
		},
		"Permission Everyone": {
			permission:      PermissionEveryone,
			expectedIsValid: true,
		},
		"Permission Named Entities": {
			permission:      PermissionNamedEntities,
			expectedIsValid: true,
		},
		"Invalid empty": {
			permission:      "",
			expectedIsValid: false,
		},
		"Invalid whitespace": {
			permission:      "  \n  ",
			expectedIsValid: false,
		},
		"Invalid string": {
			permission:      "cats_and_dogs",
			expectedIsValid: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedIsValid := tc.permission.IsValid()

			require.Equal(t, tc.expectedIsValid, returnedIsValid)
		})
	}
}

func Test_Permission_ToString(t *testing.T) {
	tests := map[string]struct {
		permission     Permission
		expectedString string
	}{
		"Permission App Collaborators": {
			permission:     PermissionAppCollaborators,
			expectedString: "app collaborators",
		},
		"Permission Everyone": {
			permission:     PermissionEveryone,
			expectedString: "everyone",
		},
		"Permission Named Entities": {
			permission:     PermissionNamedEntities,
			expectedString: "specific entities",
		},
		"Invalid string": {
			permission:     "cats_and_dogs",
			expectedString: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedString := tc.permission.ToString()

			require.Equal(t, tc.expectedString, returnedString)
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

	tests := map[string]struct {
		flags                     TestFlags
		expectedIsNamedEntityFlag bool
	}{
		"--channels flag is set": {
			flags: TestFlags{
				channelsValue: "C1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		"--organizations flag is set": {
			flags: TestFlags{
				organizationsValue: "E1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		"--users flag is set": {
			flags: TestFlags{
				usersValue: "U1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		"--workspaces flag is set": {
			flags: TestFlags{
				workspacesValue: "T1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		"Multiple valid entities are set": {
			flags: TestFlags{
				channelsValue:   "C1234",
				workspacesValue: "T1234",
			},
			expectedIsNamedEntityFlag: true,
		},
		"--everyone is set": {
			flags: TestFlags{
				everyoneValue: true,
			},
			expectedIsNamedEntityFlag: false,
		},
		"--everyone and a valid entitle are set": {
			flags: TestFlags{
				everyoneValue: true,
				channelsValue: "C1234",
			},
			expectedIsNamedEntityFlag: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			flags := pflag.NewFlagSet("entities", pflag.ContinueOnError)

			// Define the flags
			flags.StringP("channels", "", "", "channels usage")
			flags.BoolP("everyone", "", false, "everyone usage")
			flags.StringP("organizations", "", "", "organizations usage")
			flags.StringP("users", "", "", "users usage")
			flags.StringP("workspaces", "", "", "workspaces usage")

			// Set the flags based on the test values
			everyoneFlag := ""
			if tc.flags.everyoneValue {
				everyoneFlag = "--everyone"
			}

			args := []string{
				everyoneFlag,
				fmt.Sprintf("--channels=%s", tc.flags.channelsValue),
				fmt.Sprintf("--organizations=%s", tc.flags.organizationsValue),
				fmt.Sprintf("--users=%s", tc.flags.usersValue),
				fmt.Sprintf("--workspaces=%s", tc.flags.workspacesValue),
			}

			// Parse the flagset
			if err := flags.Parse(args); err != nil {
				require.Fail(t, err.Error(), "Flags parse error")
			}

			returnedIsNamedEntityFlag := IsNamedEntityFlag(flags)
			require.Equal(t, tc.expectedIsNamedEntityFlag, returnedIsNamedEntityFlag)
		})
	}
}

func Test_Permissions_GetAccessTypeDescriptionForEveryone(t *testing.T) {
	tests := map[string]struct {
		app            App
		expectedString string
	}{
		"Enterprise App": {
			app: App{ // app.IsEnterpriseApp() == true
				AppID:        "A1234",
				TeamID:       "E1234",
				EnterpriseID: "E1234",
			},
			expectedString: "everyone in all workspaces in this org granted to this app",
		},
		"Workspace App": {
			app: App{ // app.IsEnterpriseApp() == false
				AppID:  "A1234",
				TeamID: "T1234",
			},
			expectedString: "everyone in the workspace",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedString := GetAccessTypeDescriptionForEveryone(tc.app)
			require.Equal(t, tc.expectedString, returnedString)
		})
	}
}
