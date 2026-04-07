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
	"strings"

	"github.com/slackapi/slack-cli/internal/version"
)

// LoadEnvironmentVariables sets flags based on their environment variable value
func (c *Config) LoadEnvironmentVariables() error {
	// Skip when dependencies are not configured
	if c.os == nil {
		return nil
	}

	// Load accessible mode from environment variables
	var accessible = strings.TrimSpace(c.os.Getenv(slackAccessibleEnv))
	if accessible != "" && accessible != "false" && accessible != "0" {
		c.Accessible = true
	}

	// Load slackTestTraceFlag from environment variables
	var testTrace = strings.TrimSpace(c.os.Getenv(slackTestTraceEnv))
	if testTrace != "" && testTrace != "false" && testTrace != "0" {
		c.SlackTestTraceFlag = true
	}

	// Load AAA automation setting from environment variables
	var autoAAA = strings.TrimSpace(c.os.Getenv(slackAutoRequestAAAEnv))
	if autoAAA != "" && autoAAA != "false" && autoAAA != "0" {
		c.AutoRequestAAAFlag = true
	}

	// Load slackConfigDirEnv from environment variables
	var configDir = strings.TrimSpace(c.os.Getenv(slackConfigDirEnv))
	if configDir != "" {
		c.ConfigDirFlag = configDir
	}

	// Disable telemetry if either disable-telemetry or test-version environment variables
	var disableTelemetry = strings.TrimSpace(c.os.Getenv(slackDisableTelemetryEnv))
	var testVersion = strings.TrimSpace(c.os.Getenv(version.EnvTestVersion))
	if (disableTelemetry != "" && disableTelemetry != "false" && disableTelemetry != "0") || testVersion != "" {
		c.DisableTelemetryFlag = true
	}

	return nil
}
