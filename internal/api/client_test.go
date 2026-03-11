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
