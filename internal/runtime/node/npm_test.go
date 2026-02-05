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

package node

import (
	"errors"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NPMClient_InstallAllPackages(t *testing.T) {
	tests := map[string]struct {
		hookExecuteStdout     string
		hookExecuteStderr     string
		hookExecuteResponse   string
		hookExecuteError      error
		expectedVerboseOutput string
		expectedValue         string
		expectedError         error
	}{
		"When npm install is successful": {
			hookExecuteStdout: "npm install stdout",
			hookExecuteError:  nil,
			expectedValue:     "npm install stdout",
			expectedError:     nil,
		},
		"Should trim stdout": {
			hookExecuteStdout: "   npm install stdout   ",
			hookExecuteError:  nil,
			expectedValue:     "npm install stdout",
			expectedError:     nil,
		},
		"When error then PrintDebug": {
			hookExecuteStdout:     "npm install stdout",
			hookExecuteError:      errors.New("super error"),
			expectedVerboseOutput: "Error executing 'npm install --no-package-lock --no-audit --progress=false --loglevel=verbose .': super error",
			expectedValue:         "",
			expectedError:         errors.New("super error"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			projectDirPath := "/path/to/project-name"

			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			cfg := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(cfg, fs, os)
			ios.AddDefaultMocks()

			// Mock hook
			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					opts := args.Get(1).(hooks.HookExecOpts)
					_, err := opts.Stdout.Write([]byte(tt.hookExecuteStdout))
					require.NoError(t, err)
				}).
				Return(tt.hookExecuteResponse, tt.hookExecuteError)

			// Test
			npm := NPMClient{}
			value, err := npm.InstallAllPackages(ctx, projectDirPath, mockHookExecutor, ios)

			// Assertions
			require.Contains(t, value, tt.expectedValue)
			require.Equal(t, tt.expectedError, err)
			if tt.expectedVerboseOutput != "" {
				ios.AssertCalled(t, "PrintDebug", mock.Anything, tt.expectedVerboseOutput, mock.MatchedBy(func(args ...any) bool { return true }))
			}
		})
	}
}

func Test_NPMClient_InstallDevPackage(t *testing.T) {
	tests := map[string]struct {
		hookExecuteStdout     string
		hookExecuteStderr     string
		hookExecuteResponse   string
		hookExecuteError      error
		expectedVerboseOutput string
		expectedValue         string
		expectedError         error
	}{
		"When npm install package is successful": {
			hookExecuteStdout: "npm install stdout",
			hookExecuteError:  nil,
			expectedValue:     "npm install stdout",
			expectedError:     nil,
		},
		"Should trim stdout": {
			hookExecuteStdout: "   npm install stdout   ",
			hookExecuteError:  nil,
			expectedValue:     "npm install stdout",
			expectedError:     nil,
		},
		"When error then PrintDebug": {
			hookExecuteStdout:     "npm install stdout",
			hookExecuteError:      errors.New("super error"),
			expectedVerboseOutput: "Error executing 'npm install --save-dev --no-audit --progress=false --loglevel=verbose @slack/cli-hooks': super error",
			expectedValue:         "",
			expectedError:         errors.New("super error"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			projectDirPath := "/path/to/project-name"

			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			cfg := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(cfg, fs, os)
			ios.AddDefaultMocks()

			// Mock hook
			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					opts := args.Get(1).(hooks.HookExecOpts)
					_, err := opts.Stdout.Write([]byte(tt.hookExecuteStdout))
					require.NoError(t, err)
				}).
				Return(tt.hookExecuteResponse, tt.hookExecuteError)

			// Test
			npm := NPMClient{}
			value, err := npm.InstallDevPackage(ctx, slackCLIHooksPkgName, projectDirPath, mockHookExecutor, ios)

			// Assertions
			require.Contains(t, value, tt.expectedValue)
			require.Equal(t, tt.expectedError, err)
			if tt.expectedVerboseOutput != "" {
				ios.AssertCalled(t, "PrintDebug", mock.Anything, tt.expectedVerboseOutput, mock.MatchedBy(func(args ...any) bool { return true }))
			}
		})
	}
}

func Test_NPMClient_ListPackage(t *testing.T) {
	tests := map[string]struct {
		hookExecuteStdout   string
		hookExecuteStderr   string
		hookExecuteResponse string
		hookExecuteError    error
		expectedPkgVersion  string
		expectedPkgExists   bool
	}{
		"When npm list finds the package@version": {
			hookExecuteStdout: strings.Join([]string{
				"project-name@1.0.0 /path/to/project",
				"└── @slack/cli-hooks@1.1.2",
			}, "\n"),
			hookExecuteError:   nil,
			expectedPkgVersion: "@slack/cli-hooks@1.1.2",
			expectedPkgExists:  true,
		},
		"When npm list does not find package and returns error": {
			hookExecuteStdout: strings.Join([]string{
				"project-name@1.0.0 /path/to/project",
				"└── (empty)",
			}, "\n"),
			hookExecuteError:   errors.New("Exit code 1"),
			expectedPkgVersion: "",
			expectedPkgExists:  false,
		},
		"When npm list does not find the package and returns output not containing the package": {
			hookExecuteStdout: strings.Join([]string{
				"project-name@1.0.0 /path/to/project",
				"└── (empty)",
			}, "\n"),
			hookExecuteError:   nil,
			expectedPkgVersion: "",
			expectedPkgExists:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			projectDirPath := "/path/to/project-name"

			fs := slackdeps.NewFsMock()
			os := slackdeps.NewOsMock()
			cfg := config.NewConfig(fs, os)
			ios := iostreams.NewIOStreamsMock(cfg, fs, os)
			ios.AddDefaultMocks()

			// Mock hook
			mockHookExecutor := &hooks.MockHookExecutor{}
			mockHookExecutor.On("Execute", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					opts := args.Get(1).(hooks.HookExecOpts)
					_, err := opts.Stdout.Write([]byte(tt.hookExecuteStdout))
					require.NoError(t, err)
				}).
				Return(tt.hookExecuteResponse, tt.hookExecuteError)

			// Test
			npm := NPMClient{}
			pkgVersion, pkgExists := npm.ListPackage(ctx, slackCLIHooksPkgName, projectDirPath, mockHookExecutor, ios)

			// Assertions
			require.Equal(t, tt.expectedPkgVersion, pkgVersion)
			require.Equal(t, tt.expectedPkgExists, pkgExists)
		})
	}
}
