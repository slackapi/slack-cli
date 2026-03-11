package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHost(t *testing.T) {
	tests := map[string]struct {
		host     string
		expected string
	}{
		"returns the configured host": {
			host:     "https://slack.com",
			expected: "https://slack.com",
		},
		"returns empty when unset": {
			host:     "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{host: tc.host}
			assert.Equal(t, tc.expected, c.Host())
		})
	}
}

func TestSetHost(t *testing.T) {
	tests := map[string]struct {
		initial string
		newHost string
	}{
		"sets a new host": {
			initial: "",
			newHost: "https://dev.slack.com",
		},
		"overwrites existing host": {
			initial: "https://slack.com",
			newHost: "https://dev.slack.com",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{host: tc.initial}
			c.SetHost(tc.newHost)
			assert.Equal(t, tc.newHost, c.Host())
		})
	}
}
