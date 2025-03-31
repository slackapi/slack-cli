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
	"time"

	"github.com/stretchr/testify/require"
)

func Test_SlackAuth_AuthLevel(t *testing.T) {
	tests := []struct {
		name              string
		auth              *SlackAuth
		expectedAuthLevel string
	}{
		{
			name:              "Workspace-level Auth",
			auth:              &SlackAuth{IsEnterpriseInstall: false},
			expectedAuthLevel: AuthLevelWorkspace,
		},
		{
			name:              "Enterprise-level Auth",
			auth:              &SlackAuth{IsEnterpriseInstall: true},
			expectedAuthLevel: AuthLevelEnterprise,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.auth.AuthLevel(), tt.expectedAuthLevel)
		})
	}
}

func Test_SlackAuth_ShouldRotateToken(t *testing.T) {
	var token = "fakeToken"
	var refreshToken = "fakeRefreshToken"
	var timeNow = int(time.Now().Unix())

	tests := []struct {
		name     string
		input    *SlackAuth
		expected bool
	}{
		{
			name:     "nil case",
			input:    nil,
			expected: false,
		},
		{
			name:     "token but no refresh token",
			input:    &SlackAuth{Token: token},
			expected: false,
		},
		{
			name:     "token + refresh token present but expiration time is absent",
			input:    &SlackAuth{Token: token, RefreshToken: refreshToken},
			expected: false,
		},
		{
			name:     "token + refresh token + expiration present - but token expires in less than 5min",
			input:    &SlackAuth{Token: token, RefreshToken: refreshToken, ExpiresAt: timeNow + 290},
			expected: true,
		},
		{
			name:     "token + refresh token + expiration present - and token does not expire in less than 5min",
			input:    &SlackAuth{Token: token, RefreshToken: refreshToken, ExpiresAt: timeNow + 310},
			expected: false,
		},
		{
			name:     "token + refresh token + expiration present - and token expires in exactly 5min",
			input:    &SlackAuth{Token: token, RefreshToken: refreshToken, ExpiresAt: timeNow + 300},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.input.ShouldRotateToken())
		})
	}
}

func Test_SlackAuth_TokenIsExpired(t *testing.T) {
	var token = "fakeToken"
	var timeNow = int(time.Now().Unix())

	tests := []struct {
		name     string
		input    *SlackAuth
		expected bool
	}{
		{
			name:     "nil case",
			input:    nil,
			expected: false,
		},
		{
			name:     "token but no expiration",
			input:    &SlackAuth{Token: token},
			expected: false,
		},
		{
			name:     "token + expiration present - but token is expired",
			input:    &SlackAuth{Token: token, ExpiresAt: timeNow - 1},
			expected: true,
		},
		{
			name:     "token + expiration present - and token is not expired",
			input:    &SlackAuth{Token: token, ExpiresAt: timeNow + 1},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.input.TokenIsExpired())
		})
	}
}
