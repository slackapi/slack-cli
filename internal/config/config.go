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
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

// Environment Variable constants
const slackAutoRequestAAAEnv = "SLACK_AUTO_REQUEST_AAA"
const slackConfigDirEnv = "SLACK_CONFIG_DIR"
const slackDisableTelemetryEnv = "SLACK_DISABLE_TELEMETRY"
const slackTestTraceEnv = "SLACK_TEST_TRACE"

type Config struct {
	Flags *pflag.FlagSet // Flags contains the entire set of flags for a command

	// TODO: maybe these metrics-specific bits move to the tracking package now?
	// Command invoked by user (for metrics)
	Command string
	// Command invoked with any aliases resolved (for metrics)
	CommandCanonical string
	// Raw flags (for metrics)
	RawFlags []string
	// Command flags
	APIHostFlag             string
	APIHostResolved         string
	AppFlag                 string
	AutoRequestAAAFlag      bool
	ConfigDirFlag           string
	DebugEnabled            bool
	DeprecatedDevAppFlag    bool
	DeprecatedDevFlag       bool
	DeprecatedWorkspaceFlag string
	DisableTelemetryFlag    bool
	ForceFlag               bool
	LogstashHostResolved    string
	RuntimeFlag             string
	RuntimeName             string
	RuntimeVersion          string
	SkipUpdateFlag          bool
	SlackDevFlag            bool
	SlackTestTraceFlag      bool
	TeamFlag                string
	TokenFlag               string
	NoColor                 bool
	Version                 string

	// Feature experiments
	ExperimentsFlag []string
	experiments     []experiment.Experiment

	// Eventually this will also load the global and project slack config files
	DomainAuthTokens string
	ManifestEnv      map[string]string

	// ProjectID is uuid for the project
	ProjectID string

	// SystemID is the uuid for the user's system profile
	SystemID string

	// TrustUnknownSources is a user defined preference from the global slack config file
	// Set true to ignore CLI warning to user about unknown code sources
	TrustUnknownSources bool

	// fs is the file system module that's shared by all packages and enables testing & mock of the file system
	fs afero.Fs

	// os is the `os` package that's shared by all packages and enables testing & mocking
	os types.Os

	// ProjectConfig is the project-level configuration
	ProjectConfig ProjectConfigManager

	// SystemConfig is the system-level (user home) configuration
	SystemConfig SystemConfigManager
}

// NewConfig creates a new Config type with the default function handlers
func NewConfig(fs afero.Fs, os types.Os) *Config {
	config := &Config{
		fs:            fs,
		os:            os,
		Flags:         &pflag.FlagSet{},
		ProjectConfig: NewProjectConfig(fs, os),
		SystemConfig:  NewSystemConfig(fs, os),
	}

	return config
}

// SkipLocalFs returns if app and auth information is passed by flag and
// indicates that local files should not be used
func (c *Config) SkipLocalFs() bool {
	return c.TokenFlag != "" && types.IsAppID(c.AppFlag)
}
