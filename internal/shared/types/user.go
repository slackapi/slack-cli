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

type SlackUser struct {
	ID             string                    `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Email          string                    `json:"user_email,omitempty" yaml:"user_email,omitempty"`
	UserName       string                    `json:"username,omitempty" yaml:"username,omitempty"`
	PermissionType AppCollaboratorPermission `json:"permission_type,omitempty" yaml:"permission_type,omitempty"`
}

// String returns a string representation of all SlackUser attributes
func (u *SlackUser) String() string {
	var extraInfo []string
	if u.ID != "" {
		extraInfo = append(extraInfo, u.ID)
	}
	if u.Email != "" {
		extraInfo = append(extraInfo, u.Email)
	}
	if u.PermissionType != "" {
		extraInfo = append(extraInfo, string(u.PermissionType))
	}
	if len(extraInfo) > 0 {
		return fmt.Sprintf("%s (%s)", u.UserName, strings.Join(extraInfo, `, `))
	} else {
		return u.UserName
	}
}

// collaboratorShorthandF returns either the email or ID of the collaborator
func (u *SlackUser) ShorthandF() string {
	if u.Email != "" {
		return u.Email
	} else {
		return u.ID
	}
}

// User model with fields that match what is returned from functions.distributions.permissions methods
type FunctionDistributionUser struct {
	ID       string `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty"`
	UserName string `json:"username,omitempty" yaml:"username,omitempty"`
}

// User model with fields that match what is returned from the users.info method
// Method documentation: https://docs.slack.dev/reference/methods/users.info
type UserInfo struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	RealName string      `json:"real_name"`
	Profile  UserProfile `json:"profile"`
}

type UserProfile struct {
	DisplayName string `json:"display_name"`
}
