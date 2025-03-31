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

package cmdutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProcessName(t *testing.T) {
	// Restore global `os.Args` after test completes
	// Borrowed from https://golang.org/src/flag/flag_test.go
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"slack", "--help"}
	assert.Equal(t, "slack", GetProcessName(), "should return 'slack'")

	os.Args = []string{"slack-cli", "--help"}
	assert.Equal(t, "slack-cli", GetProcessName(), "should return 'slack-cli'")

	os.Args = []string{"/path/to/slack", "--help"}
	assert.Equal(t, "slack", GetProcessName(), "should remove path from executable name")
}
