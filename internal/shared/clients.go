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
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/auth"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/runtime"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/tracking"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// ClientFactory are shared clients and configurations for use across the CLI commands (cmd) and handlers (pkg).
type ClientFactory struct {
	API           func() api.APIInterface
	AppClient     func() *app.Client
	AuthInterface func() auth.AuthInterface
	CLIVersion    string
	Config        *config.Config
	EventTracker  tracking.TrackingManager
	HookExecutor  hooks.HookExecutor
	IO            iostreams.IOStreamer
	Runtime       runtime.Runtime
	SDKConfig     hooks.SDKCLIConfig

	// Browser can display or open URLs as webpages on the internet
	Browser func() slackdeps.Browser

	// Fs is the file system module that's shared by all packages and enables testing & mock of the file system
	Fs afero.Fs

	// Os are a group of operating system functions following the `os` interface that are shared by all packages and enables testing & mocking
	Os types.Os

	// CleanupWaitGroup is a group of wait groups shared by all packages and allow functions to cleanup before the process terminates
	CleanupWaitGroup sync.WaitGroup

	// Cobra are a group of Cobra functions shared by all packages and enables tests & mocking
	Cobra struct {
		// GenMarkdownTree defaults to `doc.GenMarkdownTree(...)` and can be mocked to test specific use-cases
		// TODO - This can be moved to cmd/docs/docs.go when `NewCommand` returns an instance of that can store `GenMarkdownTree` as
		//        a private member. The current thinking is that `NewCommand` would return a `SlackCommand` instead of `CobraCommand`
		GenMarkdownTree func(cmd *cobra.Command, dir string) error
	}
}

const sdkSlackDevDomainFlag = "sdk-slack-dev-domain"
const sdkUnsafelyIgnoreCertErrorsFlag = "sdk-unsafely-ignore-certificate-errors"

// NewClientFactory creates a new ClientFactory type with the default function handlers
func NewClientFactory(options ...func(*ClientFactory)) *ClientFactory {
	clients := &ClientFactory{}

	// TODO: is there a better place to put this?
	clients.CleanupWaitGroup = sync.WaitGroup{}

	// External dependencies that belong to clients for testing and mocking
	clients.Os = slackdeps.NewOs()
	clients.Fs = slackdeps.NewFs()

	// Default values
	clients.Config = config.NewConfig(clients.Fs, clients.Os)
	clients.IO = iostreams.NewIOStreams(clients.Config, clients.Fs, clients.Os)
	clients.HookExecutor = &hooks.HookExecutorDefaultProtocol{
		IO: clients.IO,
	}
	clients.EventTracker = tracking.NewEventTracker()
	clients.API = clients.defaultAPIFunc
	clients.AppClient = clients.defaultAppClientFunc
	clients.AuthInterface = clients.defaultAuthInterfaceFunc
	clients.Browser = clients.defaultBrowserFunc

	// Command-specific dependencies
	// TODO - These are methods that belong to specific commands and should be moved under each command
	//        when we replace NewCommand with NewSlackCommand that can store member variables.
	clients.Cobra.GenMarkdownTree = doc.GenMarkdownTree

	// TODO: Temporary hack to get around circular dependency in internal/api/client.go since that imports version
	// Follows pattern demonstrated by the GitHub CLI here https://github.com/cli/cli/blob/5a46c1cab601a3394caa8de85adb14f909b811e9/pkg/cmd/factory/default.go#L29
	// Used by the APIClient for its userAgent
	// Currently needed because trying to get the version of the CLI from pkg/version/version.go would cause a circular dependency
	// We can get rid of this once we refactor the code relationship between pkg/ and internal/
	// userAgent can get Slack CLI version from context which is defined in main.go, this approach bypass circular dependency. The clients.CLIVersion is retained for future code refactor purpose and serve SetVersion function
	clients.CLIVersion = ""

	// Custom values set by functional options
	// Learn more: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
	for _, option := range options {
		option(clients)
	}

	return clients
}

// defaultAPIClientFunc return a new API Client using the ConfigAPIHost
func (c *ClientFactory) defaultAPIClientFunc() *api.Client {
	return api.NewClient(nil, c.Config.APIHostResolved, c.IO)
}

// defaultAPIFunc return a new API Client using the ConfigAPIHost
func (c *ClientFactory) defaultAPIFunc() api.APIInterface {
	return c.defaultAPIClientFunc()
}

// defaultAppClientFunc return a new App Client
func (c *ClientFactory) defaultAppClientFunc() *app.Client {
	return app.NewClient(c.API(), c.Config, c.Fs, c.Os)
}

// defaultAuthClientFunc return a new Auth Client
func (c *ClientFactory) defaultAuthClientFunc() *auth.Client {
	return auth.NewClient(c.API(), c.AppClient(), c.Config, c.IO, c.Fs)
}

// defaultAuthInterfaceFunc return a new Auth Interface
func (c *ClientFactory) defaultAuthInterfaceFunc() auth.AuthInterface {
	return c.defaultAuthClientFunc()
}

// defaultBrowserFunc returns a new Browser
func (c *ClientFactory) defaultBrowserFunc() slackdeps.Browser {
	return slackdeps.NewBrowser(c.IO.WriteOut())
}

// InitRuntime initializes a new Runtime instance from the runtime flag or the
// SDK config or the directory structure
func (c *ClientFactory) InitRuntime(ctx context.Context, dirPath string) error {
	var err error
	var method string
	switch {
	case len(strings.TrimSpace(c.Config.RuntimeFlag)) > 0:
		method = "flag"
		c.Runtime, err = runtime.New(c.Config.RuntimeFlag)
	case len(strings.TrimSpace(c.SDKConfig.Runtime)) > 0:
		method = "hooks.json"
		c.Runtime, err = runtime.New(c.SDKConfig.Runtime)
	default:
		method = "auto-detect"
		c.Runtime, err = runtime.NewDetectProject(ctx, c.Fs, dirPath, c.SDKConfig)
	}
	if err != nil {
		c.IO.PrintDebug(ctx, "failed to initialize the project runtime: %s", err)
		return err
	}
	c.IO.PrintDebug(ctx, "initialize runtime from %s: %s (%s)",
		method, c.Runtime.Name(), c.Runtime.Version())
	return nil
}

// InitSDKConfig finds and loads hook configurations for a project
func (c *ClientFactory) InitSDKConfig(ctx context.Context, dirPath string) error {
	// Read the project Slack hooks file (.slack/hooks.json)
	// A command can be run any the project's root directory or subdirectory, so
	// we begin in the current directory and loop moving up the directory tree
	hooksJSONFilePath := "hooks.json"
	homeDir, err := c.Os.UserHomeDir()
	if err != nil {
		return err
	}
	for {
		// First, check if the hooks file exists in the current working directory
		hooksJSONFilePath = filepath.Join(dirPath, ".slack", "hooks.json")
		info, err := c.Fs.Stat(hooksJSONFilePath)
		if err == nil && !info.IsDir() {
			break
		}
		// Then, fallback to hooks in the deprecated project slack.json file
		// DEPRECATED(semver:major) - Drop support on the next major
		hooksJSONFilePath = filepath.Join(dirPath, "slack.json")
		info, err = c.Fs.Stat(hooksJSONFilePath)
		if err == nil && !info.IsDir() {
			c.IO.PrintDebug(ctx, "%s", slackerror.New(slackerror.ErrSlackJSONLocation))
			break
		}
		// Next, search for the hooks files in the outdated path
		// .slack/slack.json and display an error that this path
		// is deprecated
		// DEPRECATED(semver:major) - Drop support on the next major
		hooksJSONFilePath = filepath.Join(dirPath, ".slack", "slack.json")
		info, err = c.Fs.Stat(hooksJSONFilePath)
		if err == nil && !info.IsDir() {
			c.IO.PrintWarning(ctx, "%s", slackerror.New(slackerror.ErrSlackSlackJSONLocation))
			break
		}
		// Next, search for the hooks files in the outdated path
		// .slack/cli.json and display an error that this path
		// is deprecated
		// DEPRECATED(semver:major) - Drop support on the next major
		hooksJSONFilePath = filepath.Join(dirPath, ".slack", "cli.json")
		info, err = c.Fs.Stat(hooksJSONFilePath)
		if err == nil && !info.IsDir() {
			return slackerror.New(slackerror.ErrCLIConfigLocationError)
		}
		// Return an error if the current path is the project root, identified by the
		// .slack directory, because no hooks file was found
		slackConfigDirPath := filepath.Join(dirPath, ".slack")
		info, err = c.Fs.Stat(slackConfigDirPath)
		if err == nil && info.IsDir() {
			return slackerror.New(slackerror.ErrHooksJSONLocation)
		}
		// Return an error if we have reached the user home directory, root directory,
		// system root volume, or "." (returned by Dir when all path elements are removed)
		switch dirPath {
		case homeDir, "/", filepath.VolumeName(os.Getenv("SYSTEMROOT")) + "\\", ".":
			return slackerror.New(slackerror.ErrHooksJSONLocation)
		}
		// Move upward one directory level
		dirPath = filepath.Dir(dirPath)
	}
	configFileBytes, err := afero.ReadFile(c.Fs, hooksJSONFilePath)
	if err != nil {
		return err // Fixes regression: do not wrap this error, so that the caller can use `os.IsNotExists`
	}

	err = c.InitSDKConfigFromJSON(ctx, configFileBytes)
	// TODO: this is a side-effect-y way of signaling to the rest of the codebase "we are in an app project directory now"
	c.SDKConfig.WorkingDirectory = dirPath

	c.HookExecutor = hooks.GetHookExecutor(c.IO, c.SDKConfig)

	return err
}

// InitSDKConfigFromJSON merges configuration values from the `get-hooks` hook and the local configuration file
func (c *ClientFactory) InitSDKConfigFromJSON(ctx context.Context, configFileBytes []byte) error {

	// GetHooksConfig maps to the contents of a typical app's `hooks.json` file.
	// Only the `get-hooks` script is extracted with this object but the entire contents of the app's
	// `hooks.json` will later be merged with the config returned by `get-hooks`.
	type GetHooksConfig struct {
		Hooks struct {
			GetHooks hooks.HookScript `json:"get-hooks,omitempty"`
		} `json:"hooks,omitempty"`
	}

	// Load the config with the contents of the file
	getHooksConfig := GetHooksConfig{}
	err := json.Unmarshal(configFileBytes, &getHooksConfig)
	if err != nil {
		return slackerror.JSONUnmarshalError(err, configFileBytes)
	}

	// When GetHooks is available, load the scripts as the default values
	// The project's config will then be used to override these default value
	var SDKHooksResponse string
	if getHooksConfig.Hooks.GetHooks.IsAvailable() {
		getHooksConfig.Hooks.GetHooks.Name = "GetHooks"
		getHooksArgs := map[string]string{}
		if devInstanceHostname := getDevHostname(c.Config.APIHostResolved); devInstanceHostname != "" {
			getHooksArgs[sdkSlackDevDomainFlag] = devInstanceHostname
			getHooksArgs[sdkUnsafelyIgnoreCertErrorsFlag] = devInstanceHostname
		}
		var hookExecOpts = hooks.HookExecOpts{
			Args: getHooksArgs,
			Hook: getHooksConfig.Hooks.GetHooks,
		}
		defaultExecutor := hooks.HookExecutorDefaultProtocol{
			IO: c.IO,
		}
		if SDKHooksResponse, err = defaultExecutor.Execute(ctx, hookExecOpts); err != nil {
			return err
		}
	}

	// Merge the default hooks (from get-hooks) with the hooks from the project file (hooks.json)
	var config hooks.SDKCLIConfig
	if err := goutils.MergeJSON(SDKHooksResponse, string(configFileBytes), &config); err != nil {
		return slackerror.Wrap(err, slackerror.ErrSDKConfigLoad)
	}

	c.SDKConfig = config

	// Reflect on the hooks struct to set the Name field for each hook
	hooks := reflect.ValueOf(&c.SDKConfig.Hooks).Elem()
	fields := reflect.VisibleFields(reflect.TypeOf(c.SDKConfig.Hooks))
	for _, field := range fields {
		hookPointer := hooks.FieldByName(field.Name)
		if hookPointer.IsValid() {
			hookNamePointer := hookPointer.FieldByName("Name")
			if hookNamePointer.IsValid() && hookNamePointer.CanSet() {
				hookNamePointer.SetString(field.Name)
			}
		}
	}

	c.IO.PrintDebug(ctx, "initialized SDK CLI config: %+v", c.SDKConfig)

	return nil
}

// DebugMode is an example of defining a functional options helper
func DebugMode(c *ClientFactory) {
	c.Config.DebugEnabled = true
}

// SetVersion is a functional option that sets the Cli version that the API Client references
// Learn more: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func SetVersion(version string) func(c *ClientFactory) {
	return func(c *ClientFactory) { c.CLIVersion = version }
}

// getDevHostname returns the hostname of the given URL if it is dev or a numbered dev instance
func getDevHostname(host string) string {
	if host == "" {
		return ""
	}

	u, err := url.Parse(host)
	if err != nil {
		return ""
	}
	match, err := regexp.MatchString("[dev|qa]([0-9]+)?\\.slack\\.com", u.Hostname())
	if err != nil {
		panic("Unable to parse regexp")
	}
	if match {
		return u.Hostname()
	}
	return ""
}
