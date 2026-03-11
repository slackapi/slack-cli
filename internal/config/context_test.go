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

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Context_Token(t *testing.T) {
	tests := map[string]struct {
		token    string
		expected string
	}{
		"set and get a token": {
			token:    "xoxb-test-token",
			expected: "xoxb-test-token",
		},
		"set and get an empty token": {
			token:    "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := SetContextToken(context.Background(), tc.token)
			assert.Equal(t, tc.expected, GetContextToken(ctx))
		})
	}

	t.Run("get from empty context returns empty string", func(t *testing.T) {
		assert.Equal(t, "", GetContextToken(context.Background()))
	})
}

func Test_Context_EnterpriseID(t *testing.T) {
	tests := map[string]struct {
		enterpriseID string
		expected     string
	}{
		"set and get an enterprise ID": {
			enterpriseID: "E12345",
			expected:     "E12345",
		},
		"set and get an empty enterprise ID": {
			enterpriseID: "",
			expected:     "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := SetContextEnterpriseID(context.Background(), tc.enterpriseID)
			assert.Equal(t, tc.expected, GetContextEnterpriseID(ctx))
		})
	}

	t.Run("get from empty context returns empty string", func(t *testing.T) {
		assert.Equal(t, "", GetContextEnterpriseID(context.Background()))
	})
}

func Test_Context_TeamID(t *testing.T) {
	tests := map[string]struct {
		teamID   string
		expected string
	}{
		"set and get a team ID": {
			teamID:   "T12345",
			expected: "T12345",
		},
		"set and get an empty team ID": {
			teamID:   "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := SetContextTeamID(context.Background(), tc.teamID)
			assert.Equal(t, tc.expected, GetContextTeamID(ctx))
		})
	}

	t.Run("get from empty context returns empty string", func(t *testing.T) {
		assert.Equal(t, "", GetContextTeamID(context.Background()))
	})
}

func Test_Context_TeamDomain(t *testing.T) {
	tests := map[string]struct {
		teamDomain string
		expected   string
	}{
		"set and get a team domain": {
			teamDomain: "subarachnoid",
			expected:   "subarachnoid",
		},
		"set and get an empty team domain": {
			teamDomain: "",
			expected:   "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := SetContextTeamDomain(context.Background(), tc.teamDomain)
			assert.Equal(t, tc.expected, GetContextTeamDomain(ctx))
		})
	}

	t.Run("get from empty context returns empty string", func(t *testing.T) {
		assert.Equal(t, "", GetContextTeamDomain(context.Background()))
	})
}

func Test_Context_UserID(t *testing.T) {
	tests := map[string]struct {
		userID   string
		expected string
	}{
		"set and get a user ID": {
			userID:   "U12345",
			expected: "U12345",
		},
		"set and get an empty user ID": {
			userID:   "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := SetContextUserID(context.Background(), tc.userID)
			assert.Equal(t, tc.expected, GetContextUserID(ctx))
		})
	}

	t.Run("get from empty context returns empty string", func(t *testing.T) {
		assert.Equal(t, "", GetContextUserID(context.Background()))
	})
}
