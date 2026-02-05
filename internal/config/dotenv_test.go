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
	"testing"

	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_DotEnv_GetDotEnvFileVariables(t *testing.T) {
	tests := map[string]struct {
		globalVariableName  string
		globalVariableValue string
		localEnvFile        string
		expectedValues      map[string]string
	}{
		"environment file variables are read": {
			localEnvFile:   "SLACK_VARIABLE=12\n",
			expectedValues: map[string]string{"SLACK_VARIABLE": "12"},
		},
		"variable casing is preserved on load": {
			localEnvFile:   "secret_Token=Key123!\n",
			expectedValues: map[string]string{"secret_Token": "Key123!"},
		},
		"global environment variables are ignored": {
			globalVariableName:  "SLACK_VARIABLE",
			globalVariableValue: "12",
			expectedValues:      map[string]string{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			os.Setenv(tt.globalVariableName, tt.globalVariableValue)
			err := afero.WriteFile(fs, ".env", []byte(tt.localEnvFile), 0600)
			assert.NoError(t, err)
			config := NewConfig(fs, os)
			variables, err := config.GetDotEnvFileVariables()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedValues, variables)
		})
	}
}

func Test_DotEnv_LoadEnvironmentVariables(t *testing.T) {
	tableTests := map[string]struct {
		envName        string
		envValue       string
		assertOnConfig func(t *testing.T, cfg *Config)
	}{
		"SLACK_TEST_TRACE=true should enable SlackTestTraceFlag": {
			envName:  "SLACK_TEST_TRACE",
			envValue: "true",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.SlackTestTraceFlag)
			},
		},
		"SLACK_TEST_TRACE=1 should enable SlackTestTraceFlag": {
			envName:  "SLACK_TEST_TRACE",
			envValue: "1",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.SlackTestTraceFlag)
			},
		},
		"SLACK_TEST_TRACE=any should enable SlackTestTraceFlag": {
			envName:  "SLACK_TEST_TRACE",
			envValue: "any",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.SlackTestTraceFlag)
			},
		},
		"SLACK_TEST_TRACE=false should disable SlackTestTraceFlag": {
			envName:  "SLACK_TEST_TRACE",
			envValue: "false",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.SlackTestTraceFlag)
			},
		},
		"SLACK_TEST_TRACE=0 should disable SlackTestTraceFlag": {
			envName:  "SLACK_TEST_TRACE",
			envValue: "0",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.SlackTestTraceFlag)
			},
		},
		"SLACK_TEST_TRACE unset should disable SlackTestTraceFlag": {
			envName:  "SLACK_TEST_TRACE",
			envValue: "",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.SlackTestTraceFlag)
			},
		},
		`SLACK_TEST_TRACE="  true  " should trim whitespace and enable SlackTestTraceFlag`: {
			envName:  "SLACK_TEST_TRACE",
			envValue: "    true    ",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.SlackTestTraceFlag)
			},
		},
		`SLACK_TEST_TRACE="  false  " should trim whitespace and enable SlackTestTraceFlag`: {
			envName:  "SLACK_TEST_TRACE",
			envValue: "    false    ",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.SlackTestTraceFlag)
			},
		},
		"SLACK_DISABLE_TELEMETRY=true should set DisableTelemetryFlag to true": {
			envName:  "SLACK_DISABLE_TELEMETRY",
			envValue: "true",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.DisableTelemetryFlag)
			},
		},
		"SLACK_DISABLE_TELEMETRY=any should set DisableTelemetryFlag to true": {
			envName:  "SLACK_DISABLE_TELEMETRY",
			envValue: "any",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.DisableTelemetryFlag)
			},
		},
		"empty SLACK_DISABLE_TELEMETRY should set DisableTelemetryFlag to false": {
			envName:  "SLACK_DISABLE_TELEMETRY",
			envValue: "",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.DisableTelemetryFlag)
			},
		},
		"SLACK_DISABLE_TELEMETRY=false should set DisableTelemetryFlag to false": {
			envName:  "SLACK_DISABLE_TELEMETRY",
			envValue: "false",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.DisableTelemetryFlag)
			},
		},
		"SLACK_DISABLE_TELEMETRY=0 should set DisableTelemetryFlag to false": {
			envName:  "SLACK_DISABLE_TELEMETRY",
			envValue: "0",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.DisableTelemetryFlag)
			},
		},
		"SLACK_TEST_VERSION=any should set DisableTelemetryFlag to true": {
			envName:  "SLACK_TEST_VERSION",
			envValue: "any",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.DisableTelemetryFlag)
			},
		},
		"empty SLACK_TEST_VERSION should set DisableTelemetryFlag to false": {
			envName:  "SLACK_TEST_VERSION",
			envValue: "",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.DisableTelemetryFlag)
			},
		},
		"SLACK_AUTO_REQUEST_AAA=true should set AutoRequestAAAFlag to true": {
			envName:  "SLACK_AUTO_REQUEST_AAA",
			envValue: "true",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, true, cfg.AutoRequestAAAFlag)
			},
		},
		"SLACK_AUTO_REQUEST_AAA=false should set AutoRequestAAAFlag to false": {
			envName:  "SLACK_AUTO_REQUEST_AAA",
			envValue: "false",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.AutoRequestAAAFlag)
			},
		},
		"SLACK_AUTO_REQUEST_AAA=0 should set AutoRequestAAAFlag to false": {
			envName:  "SLACK_AUTO_REQUEST_AAA",
			envValue: "0",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, false, cfg.AutoRequestAAAFlag)
			},
		},
		"SLACK_CONFIG_DIR=/path/to/config should set the config dir": {
			envName:  "SLACK_CONFIG_DIR",
			envValue: "/path/to/config",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "/path/to/config", cfg.ConfigDirFlag)
			},
		},
		"SLACK_CONFIG_DIR= should not set config dir": {
			envName:  "SLACK_CONFIG_DIR",
			envValue: "",
			assertOnConfig: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "", cfg.ConfigDirFlag)
			},
		},
	}

	for name, tt := range tableTests {
		t.Run(name, func(t *testing.T) {
			// Setup
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()

			// Mocks
			os.On("Getenv", tt.envName).Return(tt.envValue)
			os.AddDefaultMocks()

			// Load environment variables
			config := NewConfig(fs, os)
			err := config.LoadEnvironmentVariables()

			// Assert
			assert.NoError(t, err)
			tt.assertOnConfig(t, config)
		})
	}
}
