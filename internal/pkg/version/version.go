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
	"regexp"
	"strings"
)

// EnvTestVersion is the environment variable to override the Version for Q&A testing
const EnvTestVersion = "SLACK_TEST_VERSION"

// Version is set with `git describe` at build time in the Makefile
var Version = "v0.0.0-dev"

// init attempts to update Version from an env var for testing purposes
func init() {
	// Optionally, override Version with the env var EnvTestVersion
	// This is useful for Q&A testing the update notifications by "rolling back" to a previous version.
	if envVersion := getVersionFromEnv(); envVersion != "" {
		Version = envVersion
	}
}

// getVersionFromEnv will return the formatted version from EnvTestVersion otherwise "".
func getVersionFromEnv() string {
	return strings.Trim(os.Getenv(EnvTestVersion), " ")
}

// Get the version and format it (e.g. `v1.0.0`)
func Get() string {
	version := Version
	if match, _ := regexp.MatchString(`^[^v]`, version); match {
		version = "v" + version
	}
	return version
}

// Raw returns the raw, unformatted version
func Raw() string {
	return Version
}
