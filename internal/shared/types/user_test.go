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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SlackUser_String(t *testing.T) {
	tests := map[string]struct {
		slackUser      *SlackUser
		expectedString string
	}{
		"UserName exists only": {
			slackUser:      &SlackUser{UserName: "charlie"},
			expectedString: "charlie",
		},
		"UserName and ID exists": {
			slackUser:      &SlackUser{UserName: "charlie", ID: "U1234"},
			expectedString: "charlie (U1234)",
		},
		"UserName,ID, and Email exists": {
			slackUser:      &SlackUser{UserName: "charlie", ID: "U1234", Email: "user@domain.com"},
			expectedString: "charlie (U1234, user@domain.com)",
		},
		"UserName, ID, Email, and PermissionType exists": {
			slackUser:      &SlackUser{UserName: "charlie", ID: "U1234", Email: "user@domain.com", PermissionType: "owner"},
			expectedString: "charlie (U1234, user@domain.com, owner)",
		},
		"UserName does not exist": {
			slackUser:      &SlackUser{},
			expectedString: "",
		},
		"UserName does not exist but other properties exist": {
			slackUser:      &SlackUser{ID: "U1234", PermissionType: "owner"},
			expectedString: " (U1234, owner)", // TODO - confirm that this is the result we want
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedString := tc.slackUser.String()
			require.Equal(t, tc.expectedString, returnedString)
		})
	}
}

func Test_SlackUser_ShorthandF(t *testing.T) {
	tests := map[string]struct {
		slackUser      *SlackUser
		expectedString string
	}{
		"ID exists": {
			slackUser:      &SlackUser{ID: "U1234"},
			expectedString: "U1234",
		},
		"Email exists": {
			slackUser:      &SlackUser{Email: "user@domain.com"},
			expectedString: "user@domain.com",
		},
		"ID and Email exist": {
			slackUser:      &SlackUser{ID: "U1234", Email: "user@domain.com"},
			expectedString: "user@domain.com",
		},
		"ID and Email do not exist": {
			slackUser:      &SlackUser{},
			expectedString: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedString := tc.slackUser.ShorthandF()
			require.Equal(t, tc.expectedString, returnedString)
		})
	}
}
