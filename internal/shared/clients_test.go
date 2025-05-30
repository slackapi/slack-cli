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

package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ClientFactory_FunctionalOptions(t *testing.T) {
	var clients *ClientFactory

	// Valid default values
	clients = NewClientFactory()
	require.False(t, clients.Config.DebugEnabled, "default should be false")
	require.False(t, clients.Config.SlackDevFlag, "default should be false")

	// Valid functional options and helper functions work
	devMode := func(c *ClientFactory) {
		c.Config.SlackDevFlag = true
	}
	clients = NewClientFactory(DebugMode, devMode)
	require.True(t, clients.Config.DebugEnabled, "default should be true")
	require.True(t, clients.Config.SlackDevFlag, "default should be true")
}

const getHooksScript = `#!/bin/sh
	echo "{\"hooks\": {\"start\": \"echo 'start' $@\"}}"
`

func Test_ClientFactory_InitSDKConfig(t *testing.T) {
	tests := map[string]struct {
		mockHooksJSONContent     string
		mockHooksJSONFilePath    string
		mockWorkingDirectory     string
		expectedError            error
		expectedWorkingDirectory string
		expectedGetHooksScript   string
	}{
		"initializes hooks from the project slack hooks file": {
			mockHooksJSONContent:     `{"hooks":{"get-hooks":"echo {}"}}`,
			mockHooksJSONFilePath:    filepath.Join(slackdeps.MockHomeDirectory, "project", ".slack", "hooks.json"),
			mockWorkingDirectory:     filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedWorkingDirectory: filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedGetHooksScript:   "echo {}",
		},
		"initializes hooks from within a nested directory": {
			mockHooksJSONContent:     fmt.Sprintf(`{"hooks":{"get-hooks":"%s"}}`, `echo \"{\\\"hooks\\\":{\\\"start\\\":\\\"time\\\"}}\"`),
			mockHooksJSONFilePath:    filepath.Join(slackdeps.MockHomeDirectory, "project", ".slack", "hooks.json"),
			mockWorkingDirectory:     filepath.Join(slackdeps.MockHomeDirectory, "project", ".slack"),
			expectedWorkingDirectory: filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedGetHooksScript:   `echo "{\"hooks\":{\"start\":\"time\"}}"`,
		},
		"initializes hooks from the deprecated project slack configuration": {
			mockHooksJSONContent:     fmt.Sprintf(`{"hooks":{"get-hooks":"%s"}}`, `echo \"{\\\"hooks\\\":{\\\"start\\\":\\\"date\\\"}}\"`),
			mockHooksJSONFilePath:    filepath.Join(slackdeps.MockHomeDirectory, "project", "slack.json"),
			mockWorkingDirectory:     filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedWorkingDirectory: filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedGetHooksScript:   `echo "{\"hooks\":{\"start\":\"date\"}}"`,
		},
		"initializes hooks from the deprecated dotslack slack configuration": {
			mockHooksJSONContent:     fmt.Sprintf(`{"hooks":{"get-hooks":"%s"}}`, `echo \"{\\\"hooks\\\":{\\\"start\\\":\\\"time\\\"}}\"`),
			mockHooksJSONFilePath:    filepath.Join(slackdeps.MockHomeDirectory, "project", ".slack", "slack.json"),
			mockWorkingDirectory:     filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedWorkingDirectory: filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedGetHooksScript:   `echo "{\"hooks\":{\"start\":\"time\"}}"`,
		},
		"errors if an outdated cli configuration file is used": {
			mockHooksJSONContent:  "{}",
			mockHooksJSONFilePath: filepath.Join(slackdeps.MockHomeDirectory, "project", ".slack", "cli.json"),
			mockWorkingDirectory:  filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedError:         slackerror.New(slackerror.ErrCLIConfigLocationError),
		},
		"errors if no project configuration file can be found": {
			mockHooksJSONContent:  "{}",
			mockHooksJSONFilePath: filepath.Join(slackdeps.MockHomeDirectory, "project", ".slack", "apps.json"),
			mockWorkingDirectory:  filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedError:         slackerror.New(slackerror.ErrHooksJSONLocation),
		},
		"errors if no project configuration directory exists": {
			mockHooksJSONContent:  "{}",
			mockHooksJSONFilePath: filepath.Join(slackdeps.MockHomeDirectory, "project", "package.json"),
			mockWorkingDirectory:  filepath.Join(slackdeps.MockHomeDirectory, "project"),
			expectedError:         slackerror.New(slackerror.ErrHooksJSONLocation),
		},
		"errors if no project configuration directory exists and searched upward to system root directory": {
			mockHooksJSONContent:  "{}",
			mockHooksJSONFilePath: filepath.Join("path", "outside", "home", "to", "project", "package.json"),
			mockWorkingDirectory:  filepath.Join("path", "outside", "home", "to", "project"),
			expectedError:         slackerror.New(slackerror.ErrHooksJSONLocation),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := NewClientFactory(clientsMock.MockClientFactory())
			err := clients.Fs.MkdirAll(filepath.Dir(tt.mockHooksJSONFilePath), 0o755)
			require.NoError(t, err)
			err = afero.WriteFile(clients.Fs, tt.mockHooksJSONFilePath, []byte(tt.mockHooksJSONContent), 0o600)
			require.NoError(t, err)
			err = clients.InitSDKConfig(ctx, tt.mockWorkingDirectory)
			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedGetHooksScript, clients.SDKConfig.Hooks.GetHooks.Command)
			assert.Equal(t, tt.expectedWorkingDirectory, clients.SDKConfig.WorkingDirectory)
		})
	}
}

func Test_ClientFactory_InitSDKConfigFromJSON(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	path := setupGetHooksScript(t)
	clients := NewClientFactory()
	getHooksJSON := fmt.Sprintf(`{"hooks":{"get-hooks": "%s"}}`, path)
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	require.True(t, clients.SDKConfig.Hooks.GetHooks.IsAvailable())
	require.True(t, clients.SDKConfig.Hooks.Start.IsAvailable())
	require.Equal(t, "echo 'start' ", string(clients.SDKConfig.Hooks.Start.Command))
}

func Test_ClientFactory_InitSDKConfigFromJSON_reflectionSetsNameProperty(t *testing.T) {
	// Setup
	ctx := slackcontext.MockContext(t.Context())
	path := setupGetHooksScript(t)
	clients := NewClientFactory()
	getHooksJSON := fmt.Sprintf(`{"hooks":{"get-hooks": "%s", "get-trigger": "echo {}", "": "echo {}"}}`, path)
	// Execute test
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	// Check
	require.Equal(t, "GetHooks", clients.SDKConfig.Hooks.GetHooks.Name)
	require.Equal(t, "GetTrigger", clients.SDKConfig.Hooks.GetTrigger.Name)
	require.Equal(t, "Start", string(clients.SDKConfig.Hooks.Start.Name))
}

func Test_ClientFactory_InitSDKConfigFromJSON_numberedDevInstance(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	path := setupGetHooksScript(t)
	clients := NewClientFactory()
	getHooksJSON := fmt.Sprintf(`{"hooks":{"get-hooks": "%s"}}`, path)
	clients.Config.APIHostResolved = "https://dev1234.slack.com"
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	require.True(t, clients.SDKConfig.Hooks.GetHooks.IsAvailable())
	require.True(t, clients.SDKConfig.Hooks.Start.IsAvailable())
	require.Contains(t, string(clients.SDKConfig.Hooks.Start.Command), "--sdk-slack-dev-domain=dev1234.slack.com", "--sdk-unsafely-ignore-certificate-errors=dev1234.slack.com")
}

func Test_ClientFactory_InitSDKConfigFromJSON_dev(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	path := setupGetHooksScript(t)
	clients := NewClientFactory()
	getHooksJSON := fmt.Sprintf(`{"hooks":{"get-hooks": "%s"}}`, path)
	clients.Config.APIHostResolved = "https://dev.slack.com"
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	require.True(t, clients.SDKConfig.Hooks.GetHooks.IsAvailable())
	require.True(t, clients.SDKConfig.Hooks.Start.IsAvailable())
	require.Contains(t, string(clients.SDKConfig.Hooks.Start.Command), "--sdk-slack-dev-domain=dev.slack.com", "--sdk-unsafely-ignore-certificate-errors=dev.slack.com")
}

func Test_ClientFactory_InitSDKConfigFromJSON_qa(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	path := setupGetHooksScript(t)
	clients := NewClientFactory()
	getHooksJSON := fmt.Sprintf(`{"hooks":{"get-hooks": "%s"}}`, path)
	clients.Config.APIHostResolved = "https://qa.slack.com"
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	require.True(t, clients.SDKConfig.Hooks.GetHooks.IsAvailable())
	require.True(t, clients.SDKConfig.Hooks.Start.IsAvailable())
	require.Contains(t, string(clients.SDKConfig.Hooks.Start.Command), "--sdk-slack-dev-domain=qa.slack.com", "--sdk-unsafely-ignore-certificate-errors=qa.slack.com")
}

func Test_ClientFactory_InitSDKConfigFromJSON_mergesExistingFile(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	path := setupGetHooksScript(t)
	clients := NewClientFactory()
	getHooksJSON := fmt.Sprintf(`{"hooks":{"get-hooks": "%s", "start": "foobar"}}`, path)
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	require.True(t, clients.SDKConfig.Hooks.GetHooks.IsAvailable())
	require.True(t, clients.SDKConfig.Hooks.Start.IsAvailable())
	require.Equal(t, "foobar", string(clients.SDKConfig.Hooks.Start.Command))
}

func Test_ClientFactory_InitSDKConfigFromJSON_noGetHooks(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clients := NewClientFactory()
	getHooksJSON := `{"hooks":{"start": "foobar"}}`
	clients.Config.APIHostResolved = "https://dev1234.slack.com"
	clients.Config.SlackDevFlag = true
	if err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON)); err != nil {
		t.Errorf("error init'ing SDK from JSON %s", err)
	}
	require.False(t, clients.SDKConfig.Hooks.GetHooks.IsAvailable())
	require.True(t, clients.SDKConfig.Hooks.Start.IsAvailable())
	require.Equal(t, "foobar", string(clients.SDKConfig.Hooks.Start.Command))
}

func Test_ClientFactory_InitSDKConfigFromJSON_brokenGetHooks(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clients := NewClientFactory()
	getHooksJSON := `{"hooks":{"get-hooks": "unknown-command"}}`
	err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON))
	require.Error(t, err)
	assert.Equal(t, slackerror.New(slackerror.ErrSDKHookInvocationFailed).Code, slackerror.ToSlackError(err).Code)
	assert.Contains(t, slackerror.ToSlackError(err).Message, "Error running 'GetHooks' command")
}

func Test_ClientFactory_InitSDKConfigFromJSON_brokenJSONFile(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clients := NewClientFactory()
	getHooksJSON := `{"hooks":{"get-hooks":`
	err := clients.InitSDKConfigFromJSON(ctx, []byte(getHooksJSON))
	require.Error(t, err)
	assert.Equal(t, slackerror.New(slackerror.ErrUnableToParseJSON).Code, slackerror.ToSlackError(err).Code)
}

func setupGetHooksScript(t *testing.T) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "get-hooks.sh")
	err := os.WriteFile(path, []byte(getHooksScript), 0700)
	if err != nil {
		t.Errorf("Failed to write file %s", err)
	}
	return path
}
