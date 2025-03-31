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

func Test_SlackUser_String(t *testing.T) {
	tests := []struct {
		name           string
		slackUser      *SlackUser
		expectedString string
	}{
		{
			name:           "UserName exists only",
			slackUser:      &SlackUser{UserName: "charlie"},
			expectedString: "charlie",
		},
		{
			name:           "UserName and ID exists",
			slackUser:      &SlackUser{UserName: "charlie", ID: "U1234"},
			expectedString: "charlie (U1234)",
		},
		{
			name:           "UserName,ID, and Email exists",
			slackUser:      &SlackUser{UserName: "charlie", ID: "U1234", Email: "user@domain.com"},
			expectedString: "charlie (U1234, user@domain.com)",
		},
		{
			name:           "UserName, ID, Email, and PermissionType exists",
			slackUser:      &SlackUser{UserName: "charlie", ID: "U1234", Email: "user@domain.com", PermissionType: "owner"},
			expectedString: "charlie (U1234, user@domain.com, owner)",
		},
		{
			name:           "UserName does not exist",
			slackUser:      &SlackUser{},
			expectedString: "",
		},
		{
			name:           "UserName does not exist but other properties exist",
			slackUser:      &SlackUser{ID: "U1234", PermissionType: "owner"},
			expectedString: " (U1234, owner)", // TODO - confirm that this is the result we want
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedString := tt.slackUser.String()
			require.Equal(t, tt.expectedString, returnedString)
		})
	}
}

func Test_SlackUser_ShorthandF(t *testing.T) {
	tests := []struct {
		name           string
		slackUser      *SlackUser
		expectedString string
	}{
		{
			name:           "ID exists",
			slackUser:      &SlackUser{ID: "U1234"},
			expectedString: "U1234",
		},
		{
			name:           "Email exists",
			slackUser:      &SlackUser{Email: "user@domain.com"},
			expectedString: "user@domain.com",
		},
		{
			name:           "ID and Email exist",
			slackUser:      &SlackUser{ID: "U1234", Email: "user@domain.com"},
			expectedString: "user@domain.com",
		},
		{
			name:           "ID and Email do not exist",
			slackUser:      &SlackUser{},
			expectedString: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedString := tt.slackUser.ShorthandF()
			require.Equal(t, tt.expectedString, returnedString)
		})
	}
}
