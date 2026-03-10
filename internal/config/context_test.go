package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextToken(t *testing.T) {
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

func TestContextEnterpriseID(t *testing.T) {
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

func TestContextTeamID(t *testing.T) {
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

func TestContextTeamDomain(t *testing.T) {
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

func TestContextUserID(t *testing.T) {
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
