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
	"time"
)

// An authorization can be organization/enterprise level or workspace level depending on the
// Slack client instance the user is logging in with
const AuthLevelEnterprise = "organization"
const AuthLevelWorkspace = "workspace"

// SlackAuth represents an authorization for the CLI to make requests of
// particular Slack team (workspace or organization) on behalf of the authorizing
// user.
//
// A SlackAuth is created by logging in with the CLI via the `login` command
type SlackAuth struct {
	Token               string    `json:"token,omitempty"`
	TeamDomain          string    `json:"team_domain"`
	TeamID              string    `json:"team_id"`
	EnterpriseID        string    `json:"enterprise_id,omitempty"`
	UserID              string    `json:"user_id"`
	LastUpdated         time.Time `json:"last_updated,omitempty"`
	APIHost             *string   `json:"api_host,omitempty"`
	RefreshToken        string    `json:"refresh_token,omitempty"`
	ExpiresAt           int       `json:"exp,omitempty"`
	IsEnterpriseInstall bool      `json:"is_enterprise_install,omitempty"`
}

// AuthLevel returns the authorization level of a specific SlackAuth, e.g. organization or workspace level
func (a *SlackAuth) AuthLevel() string {
	if a.IsEnterpriseInstall {
		return AuthLevelEnterprise
	}
	return AuthLevelWorkspace
}

type NonRotatableAuth struct {
	Token string `json:"token,omitempty"`
}

// AuthByTeamDomain describes the underlying representation of authorizations in ~/.slack/credentials.json
// Historically we have used team_domain as a keys. As of cli v2.4.0, we switch to using the unique
// team_id. AuthByTeamDomain in this case should be understood to be a map of auths where string keys are
// Slack non-unique team_domains. Wherever possible, please use AuthByTeamID below instead
type AuthByTeamDomain map[string]SlackAuth

// AuthByTeamID describes the representation of authorizations in ~/.slack/credentials.json as of
// v2.4.0, a map of auths where string keys are Slack unique team_ids, e.g. org E12345678A or
// workspace T123456789B
type AuthByTeamID map[string]SlackAuth

// Admin Approved Apps on by default in a enterprise/organization.
// Admin Approved Apps is sometimes on in standalone workspaces (an opt-in policy which can be set by Workspace Admins).
//
// Under this policy, developers in orgs must request approval to install if they are requesting their app be granted access to all workspaces.
// Approval is required when specifying that the app be granted access to a single workspace within the org if there's a org workspace-level AAA policy
// enabled.
// Developers in standalone workspaces where policy is on must request app request approval as usual.
type InstallState string

const (
	InstallSuccess          InstallState = "SUCCESS"
	InstallRequestPending   InstallState = "REQUEST_PENDING"
	InstallRequestCancelled InstallState = "REQUEST_CANCELLED"
	InstallRequestNotSent   InstallState = "REQUEST_NOT_SENT"
)

// ShouldRotateToken returns true if an auth credential can be rotated and also expires in <= 5min
func (a *SlackAuth) ShouldRotateToken() bool {

	// if ExpiresAt is 0, then the auth token is not one we can rotate
	// if RefreshToken is empty, then we cannot rotate the token either
	if a == nil || a.ExpiresAt == 0 || a.RefreshToken == "" {
		return false
	}

	// refresh the token if it expires in the next 5 mins
	timeToExpiration := a.ExpiresAt - int(time.Now().Unix())
	fiveMinutes := 60 * 5 // in seconds

	return timeToExpiration <= fiveMinutes
}

// TokenIsExpired returns true if an auth credential is expired or not
func (a *SlackAuth) TokenIsExpired() bool {

	// if expiresAt is 0 then this token does not expire
	if a == nil || a.ExpiresAt == 0 {
		return false
	}

	now := int(time.Now().Unix())
	return now > a.ExpiresAt
}
