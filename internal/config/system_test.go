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
	"encoding/json"
	"errors"
	_os "os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_SystemConfig_SetCustomConfigDirPath(t *testing.T) {
	fs := slackdeps.NewFsMock()
	os := slackdeps.NewOsMock()
	systemConfig := NewSystemConfig(fs, os)

	// Should set a custom path
	systemConfig.SetCustomConfigDirPath("/tmp/abc123")
	require.Equal(t, "/tmp/abc123", systemConfig.customConfigDirPath)

	// Should remove whitespace from a custom path
	systemConfig.SetCustomConfigDirPath("  /tmp/abc123  ")
	require.Equal(t, "/tmp/abc123", systemConfig.customConfigDirPath)

	// Should set an empty path (no custom path)
	systemConfig.SetCustomConfigDirPath("")
	require.Equal(t, "", systemConfig.customConfigDirPath)
}

func Test_SystemConfig_UserConfig(t *testing.T) {
	t.Run("Error reading configuration directory", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Home directory and working directory both return errors
		os.On("UserHomeDir").Return("", errors.New("cannot get home directory"))
		os.On("Getwd").Return("", errors.New("cannot get working directory"))
		os.AddDefaultMocks()

		systemConfig := NewSystemConfig(fs, os)
		_, err := systemConfig.UserConfig(ctx)

		require.Error(t, err)
	})

	t.Run("Error when configuration file does not exist", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		os.AddDefaultMocks()

		fs.On("Stat", mock.Anything).Return(nil, errors.New("Mockerror"))
		os.On("IsNotExist", mock.Anything).Return(true)

		systemConfig := NewSystemConfig(fs, os)
		_, err := systemConfig.UserConfig(ctx)

		require.Error(t, err)
	})

	t.Run("Error when configuration file has bad formatting", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		os.AddDefaultMocks()

		expectedConfigFileData := "{\"hello\""
		configFilePath := filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName)
		err := afero.WriteFile(fs, configFilePath, []byte(expectedConfigFileData), 0600)
		require.NoError(t, err)

		fs.On("Stat", mock.Anything).Return(nil, errors.New("Mockerror"))
		os.On("IsNotExist", mock.Anything).Return(true)

		systemConfig := NewSystemConfig(fs, os)
		_, err = systemConfig.UserConfig(ctx)

		require.Error(t, err)
		assert.Equal(t, slackerror.ToSlackError(err).Code, slackerror.ErrUnableToParseJson)
		assert.Equal(t, slackerror.ToSlackError(err).Message, "Failed to parse contents of system-level config file")
	})
}

func Test_Config_SlackConfigDir(t *testing.T) {
	t.Run("Return home directory by default", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		dir, err := config.SystemConfig.SlackConfigDir(ctx)

		require.Contains(t, dir, slackdeps.MockHomeDirectory)
		require.NoError(t, err)
	})

	t.Run("Return custom directory when it exists", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		fs.On("Stat", slackdeps.MockCustomConfigDirectory).Return(nil, nil)
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		config.SystemConfig.SetCustomConfigDirPath(slackdeps.MockCustomConfigDirectory)
		dir, err := config.SystemConfig.SlackConfigDir(ctx)

		require.Contains(t, dir, slackdeps.MockCustomConfigDirectory)
		require.NoError(t, err)
	})

	t.Run("Return error when custom directory is missing", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		fs.On("Stat", slackdeps.MockCustomConfigDirectory).Return(nil, _os.ErrNotExist)
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		config.SystemConfig.SetCustomConfigDirPath(slackdeps.MockCustomConfigDirectory)
		dir, err := config.SystemConfig.SlackConfigDir(ctx)

		require.Empty(t, dir)
		require.Error(t, err)
	})

	t.Run("Return error when home directory and working directory are unavailable", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Home directory and working directory both return errors
		os.On("UserHomeDir").Return("", errors.New("cannot get home directory"))
		os.On("Getwd").Return("", errors.New("cannot get working directory"))
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		dir, err := config.SystemConfig.SlackConfigDir(ctx)

		require.Empty(t, dir)
		require.Error(t, err)
	})

	t.Run("Return error when creating default config directory fails", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Mock fail to create hidden home directory
		fs.On("Stat", mock.Anything).Return(nil, _os.ErrNotExist)
		fs.On("Mkdir", mock.Anything, mock.Anything).Return(errors.New("no write permission"))
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		dir, err := config.SystemConfig.SlackConfigDir(ctx)

		require.Empty(t, dir)
		require.Error(t, err)
	})

	t.Run("Keep existing configuration files", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Setup: create configuration directory and files
		if err := fs.MkdirAll(filepath.Join(slackdeps.MockHomeDirectory, configFolderName), 0755); err != nil {
			require.Fail(t, "Creating fixture should not fail")
		}
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		dir, err := config.SystemConfig.SlackConfigDir(ctx)

		require.Contains(t, dir, slackdeps.MockHomeDirectory)
		require.Nil(t, err)
	})
}

func Test_SystemConfig_LogsDir(t *testing.T) {
	t.Run("Create logs folder in .slack directory", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		dotSlackFolder, _ := config.SystemConfig.SlackConfigDir(ctx)
		logsDir, err := config.SystemConfig.LogsDir(ctx)

		require.Contains(t, logsDir, dotSlackFolder)
		require.NoError(t, err)
	})
}

func Test_Config_GetLastUpdateCheckedAt(t *testing.T) {
	t.Run("Error reading User Configuration file returns current time", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Home directory and working directory both return errors
		os.On("UserHomeDir").Return("", errors.New("cannot get home directory"))
		os.On("Getwd").Return("", errors.New("cannot get working directory"))
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		lastUpdateCheckedAt, err := config.SystemConfig.GetLastUpdateCheckedAt(ctx)
		currentTime := time.Now()

		// Return timestamp should be "now" so test that the two timestamps are within 100ms of each other
		require.InDelta(t, currentTime.Local().UnixMilli(), lastUpdateCheckedAt.UnixMilli(), 100.0)
		require.Error(t, err)
	})

	t.Run("Returns LastUpdateCheckedAt from User Configuration File", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		os.AddDefaultMocks()

		config := NewConfig(fs, os)

		expectedTime := time.Now().Add(-1 * time.Hour)
		_, err := config.SystemConfig.SetLastUpdateCheckedAt(ctx, expectedTime)
		require.NoError(t, err)

		actualTime, err := config.SystemConfig.GetLastUpdateCheckedAt(ctx)
		require.Equal(t, expectedTime.Local().String(), actualTime.Local().String())
		require.NoError(t, err)
	})
}

func Test_Config_SetLastUpdateCheckedAt(t *testing.T) {
	t.Run("Error reading User Configuration file", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Home directory and working directory both return errors
		os.On("UserHomeDir").Return("", errors.New("cannot get home directory"))
		os.On("Getwd").Return("", errors.New("cannot get working directory"))
		os.AddDefaultMocks()

		ts := time.Now()
		config := NewConfig(fs, os)
		configFilePath, err := config.SystemConfig.SetLastUpdateCheckedAt(ctx, ts)

		require.Empty(t, configFilePath)
		require.Error(t, err)
	})

	t.Run("Writes LastUpdateCheckedAt to User Configuration file", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		os.AddDefaultMocks()

		ts := time.Now()
		config := NewConfig(fs, os)

		configFilePath, err := config.SystemConfig.SetLastUpdateCheckedAt(ctx, ts)
		require.NotEmpty(t, configFilePath)
		require.NoError(t, err)

		userConfig, err := config.SystemConfig.UserConfig(ctx)
		require.Equal(t, userConfig.LastUpdateCheckedAt.String(), ts.Local().String())
		require.NoError(t, err)
	})
}

func Test_SystemConfig_GetSystemID(t *testing.T) {
	t.Run("When no system_id is set, should return empty string", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		systemID, err := config.SystemConfig.GetSystemID(ctx)

		require.NoError(t, err)
		require.Empty(t, systemID)
	})

	t.Run("When system_id is set, should return system_id", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Set a system_id
		expectedSystemID := uuid.New().String()
		systemConfig := &SystemConfig{SystemID: expectedSystemID}
		systemConfigBytes, err := json.Marshal(systemConfig)
		require.NoError(t, err)
		err = afero.WriteFile(fs, filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName), systemConfigBytes, 0600)
		require.NoError(t, err)

		config := NewConfig(fs, os)
		systemID, err := config.SystemConfig.GetSystemID(ctx)

		require.NoError(t, err)
		require.Equal(t, expectedSystemID, systemID)
	})
}

func Test_SystemConfig_SetSystemID(t *testing.T) {
	t.Run("Should update the system_id", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Set current system_id
		currentSystemID := uuid.New().String()
		systemConfig := &SystemConfig{SystemID: currentSystemID}
		systemConfigBytes, err := json.Marshal(systemConfig)
		require.NoError(t, err)
		err = afero.WriteFile(fs, filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName), systemConfigBytes, 0600)
		require.NoError(t, err)

		config := NewConfig(fs, os)

		// Assert current system_id is set
		systemID, err := config.SystemConfig.GetSystemID(ctx)
		require.NoError(t, err)
		require.Equal(t, currentSystemID, systemID)

		// Assert updated system_id
		updatedSystemID := uuid.New().String()
		_, err = config.SystemConfig.SetSystemID(ctx, updatedSystemID)
		require.NoError(t, err)
		systemID, err = config.SystemConfig.GetSystemID(ctx)
		require.NoError(t, err)
		assert.Equal(t, updatedSystemID, systemID)
	})
}

func Test_SystemConfig_InitSystemID(t *testing.T) {
	t.Run("When system_id is empty, should generate a new system_id", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Set empty system_id
		systemConfig := &SystemConfig{SystemID: ""}
		systemConfigBytes, err := json.Marshal(systemConfig)
		require.NoError(t, err)
		err = afero.WriteFile(fs, filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName), systemConfigBytes, 0600)
		require.NoError(t, err)

		config := NewConfig(fs, os)

		// Assert initialize generates a new system_id
		initSystemID, err := config.SystemConfig.InitSystemID(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, initSystemID)

		// Assert system_id written to file
		systemID, err := config.SystemConfig.GetSystemID(ctx)
		require.NoError(t, err)
		assert.Equal(t, initSystemID, systemID)
	})

	t.Run("When system_id is not empty, should not overwrite it", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Set empty system_id
		expectedSystemID := uuid.New().String()
		systemConfig := &SystemConfig{SystemID: expectedSystemID}
		systemConfigBytes, err := json.Marshal(systemConfig)
		require.NoError(t, err)
		err = afero.WriteFile(fs, filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName), systemConfigBytes, 0600)
		require.NoError(t, err)

		config := NewConfig(fs, os)

		// Assert initialize generates a new system_id
		systemID, err := config.SystemConfig.InitSystemID(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedSystemID, systemID)
	})
}

func Test_SystemConfig_readConfigFile(t *testing.T) {
	t.Run("Should read the config file", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Setup config
		config := NewConfig(fs, os)
		expectedConfigFileData := "{\"hello\":\"world\"}\n"

		// Write a config file
		configFilePath := filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName)
		err := afero.WriteFile(fs, configFilePath, []byte(expectedConfigFileData), 0600)
		require.NoError(t, err)

		// Assert reading the file
		configFileBytes, err := config.SystemConfig.readConfigFile(configFilePath)
		require.NoError(t, err)
		assert.Equal(t, expectedConfigFileData, string(configFileBytes))
	})

	t.Run("Should support file locking", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Setup config
		config := NewConfig(fs, os)
		configFilePath := filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName)
		expectedConfigFileData := "{\"hello\":\"world\"}\n"

		// Write a config file to be read later
		err := config.SystemConfig.writeConfigFile(configFilePath, []byte(expectedConfigFileData))
		require.NoError(t, err)

		// Setup the WaitGroup
		var wg sync.WaitGroup
		wg.Add(2)

		// Step 1) Lock the file; it will be unlocked in Step 4)
		config.SystemConfig.lock()

		// Step 2) Try to read the file, which will block on the internal file locking
		go func() {
			configFileBytes, err := config.SystemConfig.readConfigFile(configFilePath)
			require.NoError(t, err)
			require.Equal(t, expectedConfigFileData, string(configFileBytes))
			wg.Done()
		}()

		// Step 3) Wait a few seconds to guarantee Step 2) goroutine executes first
		time.Sleep(50 * time.Millisecond)

		// Step 4) Unlock the file so that Step 2) can stop waiting and read the config file
		go func() {
			config.SystemConfig.unlock()
			wg.Done()
		}()

		wg.Wait()
	})
}

func Test_SystemConfig_writeConfigFile(t *testing.T) {
	t.Run("Should add newline to end config file", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Setup config
		config := NewConfig(fs, os)

		// Write a config file that doesn't end with a newline
		configFilePath := filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName)
		err := config.SystemConfig.writeConfigFile(configFilePath, []byte("{}"))
		require.NoError(t, err)

		// Assert that a newline was appended to the end of the file
		bytes, err := afero.ReadFile(fs, configFilePath)
		s := string(bytes)
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(s, "\n"))
	})

	t.Run("Should not add newline to end config file when it already exists", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Setup config
		config := NewConfig(fs, os)

		// Write a config file that doesn't end with a newline
		configFilePath := filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName)
		err := config.SystemConfig.writeConfigFile(configFilePath, []byte("{}\n"))
		require.NoError(t, err)

		// Assert that a newline was appended to the end of the file
		bytes, err := afero.ReadFile(fs, configFilePath)
		s := string(bytes)
		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(s, "\n"))
		assert.False(t, strings.HasSuffix(s, "\n\n"))
	})

	t.Run("Should support file locking", func(t *testing.T) {
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Setup config
		config := NewConfig(fs, os)
		configFilePath := filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName)
		expectedConfigFileData := "{\"hello\":\"world\"}\n"

		// Setup the WaitGroup
		var wg sync.WaitGroup
		wg.Add(2)

		// Step 1) Lock the file; it will be unlocked in Step 4)
		config.SystemConfig.lock()

		// Step 2) Try to write the file, which will block on the internal file locking
		go func() {
			err := config.SystemConfig.writeConfigFile(configFilePath, []byte(expectedConfigFileData))
			require.NoError(t, err)
			wg.Done()
		}()

		// Step 3) Wait a few seconds to guarantee Step 2) go routine executes first
		time.Sleep(50 * time.Millisecond)

		// Step 4) Unlock the file so that Step 2) can stop waiting and write the config file
		go func() {
			config.SystemConfig.unlock()
			wg.Done()
		}()

		wg.Wait()

		// Assert that the file was written
		configFileBytes, err := afero.ReadFile(fs, configFilePath)
		require.NoError(t, err)
		require.Equal(t, expectedConfigFileData, string(configFileBytes))
	})
}

func Test_SystemConfig_GetTrustUnknownSources(t *testing.T) {
	t.Run("When no trust_unknown_sources is set, should return false", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		config := NewConfig(fs, os)
		trustSources, err := config.SystemConfig.GetTrustUnknownSources(ctx)

		require.NoError(t, err)
		require.False(t, trustSources)
	})

	t.Run("When trust_unknown_sources is set, should return true", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Set a trust_unknown_sources
		systemConfig := &SystemConfig{TrustUnknownSources: true}
		systemConfigBytes, err := json.Marshal(systemConfig)
		require.NoError(t, err)
		err = afero.WriteFile(fs, filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName), systemConfigBytes, 0600)
		require.NoError(t, err)

		config := NewConfig(fs, os)
		trustSources, err := config.SystemConfig.GetTrustUnknownSources(ctx)

		require.NoError(t, err)
		require.True(t, trustSources)
	})
}

func Test_SystemConfig_SetTrustUnknownSources(t *testing.T) {
	t.Run("Should update the trust_unknown_sources", func(t *testing.T) {
		ctx := slackcontext.MockContext(t.Context())
		fs := slackdeps.NewFsMock()
		os := slackdeps.NewOsMock()

		// Use default mocks to return home directory path
		os.AddDefaultMocks()

		// Set current trust_unknown_sources to true
		systemConfig := &SystemConfig{TrustUnknownSources: false}
		systemConfigBytes, err := json.Marshal(systemConfig)
		require.NoError(t, err)
		err = afero.WriteFile(fs, filepath.Join(slackdeps.MockHomeDirectory, configFolderName, configFileName), systemConfigBytes, 0600)
		require.NoError(t, err)

		config := NewConfig(fs, os)

		// Assert current trust_unknown_sources is set to false
		trustSources, err := config.SystemConfig.GetTrustUnknownSources(ctx)
		require.NoError(t, err)
		require.False(t, trustSources)

		// Assert updated trust_unknown_sources
		err = config.SystemConfig.SetTrustUnknownSources(ctx, true)
		require.NoError(t, err)
		trustSources, err = config.SystemConfig.GetTrustUnknownSources(ctx)
		require.NoError(t, err)
		assert.True(t, trustSources)
	})
}
