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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

// Constants
const credentialsFileName = "credentials.json"
const configFolderName = ".slack"
const configFileName = "config.json"
const logsFolderName = "logs"

// SystemConfigManager is the interface for interacting with the system config
type SystemConfigManager interface {
	SetCustomConfigDirPath(customConfigDirPath string)
	UserConfig(ctx context.Context) (*SystemConfig, error)
	SlackConfigDir(ctx context.Context) (string, error)
	LogsDir(ctx context.Context) (string, error)
	GetTrustUnknownSources(ctx context.Context) (bool, error)
	SetTrustUnknownSources(ctx context.Context, value bool) error
	GetLastUpdateCheckedAt(ctx context.Context) (time.Time, error)
	SetLastUpdateCheckedAt(ctx context.Context, lastUpdateCheckedAt time.Time) (path string, err error)
	InitSystemID(ctx context.Context) (string, error)
	GetSystemID(ctx context.Context) (string, error)
	SetSystemID(ctx context.Context, systemID string) (string, error)
	GetSurveyConfig(ctx context.Context, id string) (SurveyConfig, error)
	SetSurveyConfig(ctx context.Context, id string, surveyConfig SurveyConfig) error
	initializeConfigFiles(ctx context.Context, dir string) error
	readConfigFile(configFilePath string) ([]byte, error)
	writeConfigFile(configFilePath string, configFileBytes []byte) error
	lock()
	unlock()
}

// SystemConfig contains the system-level config file
type SystemConfig struct {
	Experiments         []experiment.Experiment `json:"experiments,omitempty"`
	LastUpdateCheckedAt time.Time               `json:"last_update_checked_at,omitempty"`
	Surveys             map[string]SurveyConfig `json:"surveys,omitempty"`
	SystemID            string                  `json:"system_id,omitempty"`
	TrustUnknownSources bool                    `json:"trust_unknown_sources,omitempty"`

	// fs is the file system module that's shared by all packages and enables testing & mock of the file system
	fs afero.Fs

	// os is the `os` package that's shared by all packages and enables testing & mocking
	os types.Os

	// customConfigDirPath is an optional path for the SystemConfig directory
	customConfigDirPath string

	// configFileLock is a file locking mutex that works across goroutines for configFileName
	configFileLock sync.Mutex
}

// NewSystemConfig read and writes to the system-level configuration directory
func NewSystemConfig(fs afero.Fs, os types.Os) *SystemConfig {
	systemConfig := &SystemConfig{
		fs: fs,
		os: os,
	}

	return systemConfig
}

func (c *SystemConfig) lock() {
	c.configFileLock.Lock()
}

func (c *SystemConfig) unlock() {
	c.configFileLock.Unlock()
}

// SetCustomConfigDirPath sanitizes and sets a custom system config directory path
func (c *SystemConfig) SetCustomConfigDirPath(customConfigDirPath string) {
	c.customConfigDirPath = strings.TrimSpace(customConfigDirPath)
}

// UserConfig returns the system-level config.json file contents
func (c *SystemConfig) UserConfig(ctx context.Context) (*SystemConfig, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "ReadUserConfig")
	defer span.Finish()

	var config SystemConfig

	dir, err := c.SlackConfigDir(ctx)
	if err != nil {
		return &SystemConfig{}, err
	}
	var path string = filepath.Join(dir, configFileName)

	if _, err := c.fs.Stat(path); os.IsNotExist(err) {
		return &config, err
	}

	configFileBytes, err := c.readConfigFile(path)
	if err != nil {
		return &config, err
	}

	err = json.Unmarshal(configFileBytes, &config)
	if err != nil {
		return &config, slackerror.New(slackerror.ErrUnableToParseJSON).
			WithMessage("Failed to parse contents of system-level config file").
			WithRootCause(err).
			WithRemediation("Check that %s is valid JSON", style.HomePath(path))
	}

	if config.Surveys == nil {
		config.Surveys = map[string]SurveyConfig{}
	}

	return &config, nil
}

// SlackConfigDir returns a folder/directory location for storing
// auth credentials and other config info.  It should return a
// new hidden folder in the home directory where possible.
func (c *SystemConfig) SlackConfigDir(ctx context.Context) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SlackConfigDir")
	defer span.Finish()

	var configDirPath = ""

	// Try to use a custom config directory
	if c.customConfigDirPath != "" {
		if _, err := c.fs.Stat(c.customConfigDirPath); os.IsNotExist(err) {
			return "", slackerror.Wrap(err, slackerror.ErrHomeDirectoryAccessFailed)
		}
		configDirPath = c.customConfigDirPath
	}

	// Default to the user home directory
	if configDirPath == "" {
		homeDirPath, err := c.os.UserHomeDir()
		if err != nil {
			return "", slackerror.New(slackerror.ErrHomeDirectoryAccessFailed)
		}

		configDirPath = filepath.Join(homeDirPath, configFolderName)
		if _, err := c.fs.Stat(configDirPath); os.IsNotExist(err) {
			// Hidden slack folder does not exist so let's create it
			if err := c.fs.Mkdir(configDirPath, 0755); err != nil {
				return "", slackerror.New(slackerror.ErrHomeDirectoryAccessFailed)
			}
		}
	}

	// Initialize the config files
	if err := c.initializeConfigFiles(ctx, configDirPath); err != nil {
		return "", err
	}

	return configDirPath, nil
}

// LogsDir returns the logs directory path stored in the system configuration directory
// When the directory doesn't exist, then it will create it
func (c *SystemConfig) LogsDir(ctx context.Context) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "LogsDir")
	defer span.Finish()

	dotSlackFolder, err := c.SlackConfigDir(ctx)
	if err != nil {
		return "", err
	}

	var logsdir = filepath.Join(dotSlackFolder, logsFolderName)
	if _, err := c.fs.Stat(logsdir); os.IsNotExist(err) {
		// the logs folder does not exist so let's create it
		if err := c.fs.Mkdir(logsdir, 0755); err != nil {
			return "", slackerror.Wrap(err, "failed to create logs folder in slack directory")
		}
	}
	return logsdir, nil
}

// GetLastUpdateCheckedAt reads the time of the LastUpdateCheckedAt property in UserConfig file
func (c *SystemConfig) GetLastUpdateCheckedAt(ctx context.Context) (time.Time, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "LastUpdateCheckedAt")
	defer span.Finish()

	var userConfig, err = c.UserConfig(ctx)
	if err != nil {
		return time.Now(), err
	}

	return userConfig.LastUpdateCheckedAt, nil
}

// SetLastUpdateCheckedAt writes the lastUpdateCheckAt time to the UserConfig file.
// When successful, the config file path is returned.
func (c *SystemConfig) SetLastUpdateCheckedAt(ctx context.Context, lastUpdateCheckedAt time.Time) (path string, err error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SetLastUpdateCheckedAt")
	defer span.Finish()

	userConfig, err := c.UserConfig(ctx)
	if err != nil {
		return "", err
	}

	userConfig.LastUpdateCheckedAt = lastUpdateCheckedAt

	b, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return "", err
	}

	dir, err := c.SlackConfigDir(ctx)
	if err != nil {
		return "", err
	}
	path = filepath.Join(dir, configFileName)

	err = c.writeConfigFile(path, b)
	if err != nil {
		return path, err
	}

	return path, nil
}

// InitSystemID sets the system_id in the user-level config to a random SHA256 string when it's currently unset
func (c *SystemConfig) InitSystemID(ctx context.Context) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "InitSystemID")
	defer span.Finish()

	systemID, err := c.GetSystemID(ctx)
	if err != nil {
		return "", err
	}

	if systemID != "" {
		return systemID, nil
	}

	return c.SetSystemID(ctx, uuid.New().String())
}

// GetSystemID reads the system_id from the user-level config file
func (c *SystemConfig) GetSystemID(ctx context.Context) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetSystemID")
	defer span.Finish()

	var userConfig, err = c.UserConfig(ctx)
	if err != nil {
		return "", err
	}

	return userConfig.SystemID, nil
}

// SetSystemID sets the system_id to a random SHA256 string in the user-level config file
func (c *SystemConfig) SetSystemID(ctx context.Context, systemID string) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SetSystemID")
	defer span.Finish()

	userConfig, err := c.UserConfig(ctx)
	if err != nil {
		return "", err
	}

	userConfig.SystemID = systemID

	b, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return "", err
	}

	dir, err := c.SlackConfigDir(ctx)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, configFileName)

	err = c.writeConfigFile(path, b)
	if err != nil {
		return "", err
	}

	return userConfig.SystemID, nil
}

// GetSurveyConfig reads the survey for the given survey ID from the project-level config file
func (c *SystemConfig) GetSurveyConfig(ctx context.Context, name string) (SurveyConfig, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetSurveyConfig")
	defer span.Finish()

	var userConfig, err = c.UserConfig(ctx)
	if err != nil {
		return SurveyConfig{}, err
	}

	survey, ok := userConfig.Surveys[name]
	if !ok {
		return SurveyConfig{}, slackerror.New(slackerror.ErrSurveyConfigNotFound)
	}

	return survey, nil
}

// SetSurveyConfig writes the survey for the given survey ID from the system-level config file
func (c *SystemConfig) SetSurveyConfig(ctx context.Context, name string, surveyConfig SurveyConfig) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SetSurveyConfig")
	defer span.Finish()

	var userConfig, err = c.UserConfig(ctx)
	if err != nil {
		return err
	}

	userConfig.Surveys[name] = SurveyConfig{
		AskedAt:     surveyConfig.AskedAt,
		CompletedAt: surveyConfig.CompletedAt,
	}

	b, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return err
	}

	dir, err := c.SlackConfigDir(ctx)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, configFileName)

	err = c.writeConfigFile(path, b)
	if err != nil {
		return err
	}

	return nil
}

// GetTrustUnknownSources reads the TrustUnknownSources property from the user-level config file
func (c *SystemConfig) GetTrustUnknownSources(ctx context.Context) (bool, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetTrustUnknownSources")
	defer span.Finish()

	var userConfig, err = c.UserConfig(ctx)
	if err != nil {
		return false, err
	}
	return userConfig.TrustUnknownSources, nil
}

// SetTrustUnknownSources sets the trust_unknown_sources property to a the user-level config file
func (c *SystemConfig) SetTrustUnknownSources(ctx context.Context, value bool) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SetTrustUnknownSources")
	defer span.Finish()

	var userConfig, err = c.UserConfig(ctx)
	if err != nil {
		return err
	}

	userConfig.TrustUnknownSources = value

	b, err := json.MarshalIndent(userConfig, "", "  ")
	if err != nil {
		return err
	}

	dir, err := c.SlackConfigDir(ctx)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, configFileName)

	err = c.writeConfigFile(path, b)
	if err != nil {
		return err
	}
	return nil
}

// initializeConfigFolder creates the required files (credentials.json and config.json)
// in the /.slack/ folder if they do not yet exist
func (c *SystemConfig) initializeConfigFiles(ctx context.Context, dir string) error {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "initializeConfigFolder")
	defer span.Finish()

	// ensure the credentials.json file exists
	var credentialsPath string = filepath.Join(dir, credentialsFileName)
	_, err := c.fs.Stat(credentialsPath)
	if os.IsNotExist(err) {
		// create the credentials file because it does not exist
		err = afero.WriteFile(c.fs, credentialsPath, []byte("{}\n"), 0600)
		if err != nil {
			return slackerror.Wrap(err, "failed to create credentials.json in slack directory")
		}
	}

	// ensure the config.json file exists
	var configPath string = filepath.Join(dir, configFileName)
	_, err = c.fs.Stat(configPath)
	if os.IsNotExist(err) {
		err = c.writeConfigFile(configPath, []byte("{}\n"))
		if err != nil {
			return slackerror.Wrap(err, "failed to create config.json in slack directory")
		}
	}

	// ensure the logs folder exists
	var logsPath string = filepath.Join(dir, logsFolderName)
	_, err = c.fs.Stat(logsPath)
	if os.IsNotExist(err) {
		// create the logs folder because it does not exist
		err := c.fs.Mkdir(logsPath, 0755)
		if err != nil {
			return slackerror.Wrap(err, "failed to create logs folder in slack directory")
		}
	}
	return nil
}

// readConfigFile will read the configFilePath with support for file locking across goroutines.
func (c *SystemConfig) readConfigFile(configFilePath string) ([]byte, error) {
	// Lock the file
	c.configFileLock.Lock()
	defer c.configFileLock.Unlock()

	// Read the content
	bytes, err := afero.ReadFile(c.fs, configFilePath)
	return bytes, err
}

// writeConfigFile will write the configFileBytes to configFilePath with support for file locking
// across goroutines.
func (c *SystemConfig) writeConfigFile(configFilePath string, configFileBytes []byte) error {
	// Lock the file
	c.configFileLock.Lock()
	defer c.configFileLock.Unlock()

	// Good practice to end the file with a newline
	if !strings.HasSuffix(string(configFileBytes), "\n") {
		configFileBytes = []byte(fmt.Sprintf("%s\n", string(configFileBytes)))
	}

	// Write the content
	err := afero.WriteFile(c.fs, configFilePath, configFileBytes, 0600)
	return err
}
