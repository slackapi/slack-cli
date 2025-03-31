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

package version

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	v := Version
	assert.True(t, len(v) > 0, "some default value exists")
}

// Test overriding the Version with an environment variable
func Test_EnvTestVersion(t *testing.T) {
	// Setup
	var _EnvTestVersion = os.Getenv(EnvTestVersion)

	// Teardown
	restore := func() {
		os.Setenv(EnvTestVersion, _EnvTestVersion)
	}
	defer restore()

	// Test mocking a version from the env var
	restore()
	os.Setenv(EnvTestVersion, "v0.1.2")
	require.Equal(t, "v0.1.2", getVersionFromEnv(), "should override Version with env var")

	// Test trimming whitespace
	restore()
	os.Setenv(EnvTestVersion, " v0.1.2  ")
	require.Equal(t, "v0.1.2", getVersionFromEnv(), "should trim whitespace")

	// Test ignoring empty env var
	restore()
	os.Setenv(EnvTestVersion, " ")
	require.Equal(t, "", getVersionFromEnv(), "should ignore empty env var")
}
