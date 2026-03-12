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
	"bytes"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_SetFlags(t *testing.T) {
	fs := slackdeps.NewFsMock()
	os := slackdeps.NewOsMock()
	config := NewConfig(fs, os)
	cmd := &cobra.Command{}
	cmd.Flags().String("test-flag", "default", "a test flag")

	config.SetFlags(cmd)
	assert.NotNil(t, config.Flags)
	f := config.Flags.Lookup("test-flag")
	assert.NotNil(t, f)
	assert.Equal(t, "default", f.DefValue)
}

func Test_InitializeGlobalFlags(t *testing.T) {
	fs := slackdeps.NewFsMock()
	os := slackdeps.NewOsMock()
	config := NewConfig(fs, os)
	cmd := &cobra.Command{}

	config.InitializeGlobalFlags(cmd)

	tests := map[string]struct {
		longform  string
		shorthand string
		hidden    bool
	}{
		"apihost": {
			longform: "apihost",
			hidden:   true,
		},
		"app": {
			longform:  "app",
			shorthand: "a",
		},
		"config-dir": {
			longform: "config-dir",
		},
		"experiment": {
			longform: "experiment",
		},
		"force": {
			longform:  "force",
			shorthand: "f",
		},
		"no-color": {
			longform: "no-color",
		},
		"runtime": {
			longform:  "runtime",
			shorthand: "r",
			hidden:    true,
		},
		"skip-update": {
			longform:  "skip-update",
			shorthand: "s",
		},
		"slackdev": {
			longform: "slackdev",
			hidden:   true,
		},
		"team": {
			longform:  "team",
			shorthand: "w",
		},
		"token": {
			longform: "token",
		},
		"verbose": {
			longform:  "verbose",
			shorthand: "v",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := cmd.PersistentFlags().Lookup(tc.longform)
			assert.NotNil(t, f, "flag %s should be registered", tc.longform)
			if tc.shorthand != "" {
				assert.Equal(t, tc.shorthand, f.Shorthand, "flag %s shorthand mismatch", tc.longform)
			}
			assert.Equal(t, tc.hidden, f.Hidden, "flag %s hidden mismatch", tc.longform)
		})
	}
}

func TestDeprecatedFlagSubstitutions(t *testing.T) {
	tests := map[string]struct {
		expectedWarnings    []string
		expectedError       *slackerror.Error
		prepareFlags        func(*Config)
		assertSubstitutions func(*Config)
	}{
		"deprecated dev flag continues with slackdev": {
			expectedWarnings: []string{
				"Deprecation of --dev",
				"--dev flag has been removed",
				"--slackdev flag can now be used as shorthand for --apihost=\"https://dev.slack.com\"",
				"Continuing execution with --slackdev",
			},
			prepareFlags: func(c *Config) {
				c.DeprecatedDevFlag = true
			},
			assertSubstitutions: func(c *Config) {
				assert.True(t, c.SlackDevFlag)
			},
		},
		"deprecated local run flag continues with app local": {
			expectedWarnings: []string{
				"Deprecation of --local-run",
				"The --local-run flag has been removed",
				"Specify a local app with --app local",
				"Continuing execution with --app local",
			},
			prepareFlags: func(c *Config) {
				c.DeprecatedDevAppFlag = true
			},
			assertSubstitutions: func(c *Config) {
				assert.Equal(t, c.AppFlag, "local")
			},
		},
		"deprecated local run flag errors with app id": {
			expectedWarnings: []string{
				"Deprecation of --local-run",
				"The --local-run flag has been removed",
				"Specify a local app with --app local",
			},
			expectedError: slackerror.New(slackerror.ErrMismatchedFlags),
			prepareFlags: func(c *Config) {
				c.DeprecatedDevAppFlag = true
				c.AppFlag = "A0123456789"
			},
			assertSubstitutions: func(c *Config) {
				assert.Equal(t, c.AppFlag, "A0123456789")
			},
		},
		"deprecated workspace flag continues with team": {
			expectedWarnings: []string{
				"Deprecation of --workspace",
				"The --workspace flag has been removed",
				"Specify a Slack workspace or organization with --team <domain|id>",
				"Continuing execution with --team T0123456789",
			},
			prepareFlags: func(c *Config) {
				c.DeprecatedWorkspaceFlag = "T0123456789"
			},
			assertSubstitutions: func(c *Config) {
				assert.Equal(t, c.TeamFlag, "T0123456789")
			},
		},
		"deprecated workspace flag errors with mismatched team flag": {
			expectedWarnings: []string{
				"Deprecation of --workspace",
				"The --workspace flag has been removed",
				"Specify a Slack workspace or organization with --team <domain|id>",
			},
			expectedError: slackerror.New(slackerror.ErrMismatchedFlags),
			prepareFlags: func(c *Config) {
				c.DeprecatedWorkspaceFlag = "bigspeck"
				c.TeamFlag = "T0123456789"
			},
			assertSubstitutions: func(c *Config) {
				assert.Equal(t, c.TeamFlag, "T0123456789")
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			config := NewConfig(fs, os)
			cmd := &cobra.Command{}
			stdout := bytes.Buffer{}
			stderr := bytes.Buffer{}
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			tc.prepareFlags(config)
			err := config.DeprecatedFlagSubstitutions(cmd)
			if tc.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError.Code, slackerror.ToSlackError(err).Code)
			}
			assert.Equal(t, stdout.String(), "")
			for _, line := range tc.expectedWarnings {
				assert.Contains(t, stderr.String(), line)
			}
			tc.assertSubstitutions(config)
		})
	}
}
