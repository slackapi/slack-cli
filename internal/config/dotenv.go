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

package config

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/spf13/afero"
)

// GetDotEnvFileVariables collects only the variables in the .env file
func (c *Config) GetDotEnvFileVariables() (map[string]string, error) {
	variables := map[string]string{}
	file, err := afero.ReadFile(c.fs, ".env")
	if err != nil && !c.os.IsNotExist(err) {
		return variables, err
	}
	return godotenv.UnmarshalBytes(file)
}

// LoadEnvironmentVariables sets flags based on their environment variable value
//
// Note: Values are not loaded from the .env file. Use: `GetDotEnvFileVariables`
func (c *Config) LoadEnvironmentVariables() error {
	// Skip when dependencies are not configured
	if c.os == nil {
		return nil
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
