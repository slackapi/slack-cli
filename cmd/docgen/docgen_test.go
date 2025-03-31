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

package docgen

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewDocsCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"when no path argument": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Fs.On("MkdirAll", mock.Anything, mock.Anything).Return(nil)
			},
			ExpectedOutputs: []string{
				filepath.Join(slackdeps.MockWorkingDirectory, "docs"),
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.Cobra.AssertCalled(
					t,
					"GenMarkdownTree",
					mock.Anything,
					filepath.Join(slackdeps.MockWorkingDirectory, "docs", "commands"),
				)
				cm.Fs.AssertCalled(
					t,
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "docs"),
					mock.Anything,
				)
				cm.Fs.AssertCalled(
					t,
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "docs", "commands"),
					mock.Anything,
				)
				file, err := cm.Fs.Open(
					filepath.Join(slackdeps.MockWorkingDirectory, "docs", "errors.md"),
				)
				require.NoError(t, err)
				markdownErrors := make([]byte, 1001)
				_, err = file.Read(markdownErrors)
				require.NoError(t, err)
				assert.Contains(t, string(markdownErrors), "# Slack CLI errors reference")
				assert.Contains(t, string(markdownErrors), "## Slack CLI errors list")
				assert.Contains(t, string(markdownErrors), "### access_denied {#access_denied}")
				assert.Contains(
					t,
					string(markdownErrors),
					"**Message**: You don't have the permission to access the specified resource",
				)
			},
		},
		"when path argument exists": {
			CmdArgs: []string{"markdown-docs"},
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Fs.On("MkdirAll", mock.Anything, mock.Anything).Return(nil)
			},
			ExpectedOutputs: []string{
				filepath.Join(slackdeps.MockWorkingDirectory, "markdown-docs"),
			},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.Cobra.AssertCalled(
					t,
					"GenMarkdownTree",
					mock.Anything,
					filepath.Join(slackdeps.MockWorkingDirectory, "markdown-docs", "commands"),
				)
				cm.Fs.AssertCalled(
					t,
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "markdown-docs"),
					mock.Anything,
				)
				cm.Fs.AssertCalled(
					t,
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "markdown-docs", "commands"),
					mock.Anything,
				)
				file, err := cm.Fs.Open(
					filepath.Join(slackdeps.MockWorkingDirectory, "markdown-docs", "errors.md"),
				)
				require.NoError(t, err)
				markdownErrors := make([]byte, 28)
				_, err = file.Read(markdownErrors)
				require.NoError(t, err)
				assert.Contains(t, string(markdownErrors), "# Slack CLI errors reference")
			},
		},
		"when path argument is an empty string of spaces": {
			CmdArgs: []string{"  "},
			ExpectedOutputs: []string{
				filepath.Join(slackdeps.MockWorkingDirectory, "docs"),
			},
		},
		"when Getwd returns error": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Fs.On("MkdirAll", mock.Anything, mock.Anything).Return(nil)
				cm.Os.On("Getwd").Return("", errors.New("somehow there is no cwd"))
			},
			CmdArgs:         []string{},
			ExpectedOutputs: []string{"References saved to: docs"},
			ExpectedAsserts: func(t *testing.T, cm *shared.ClientsMock) {
				cm.Cobra.AssertCalled(
					t,
					"GenMarkdownTree",
					mock.Anything,
					filepath.Join("docs", "commands"),
				)
				cm.Fs.AssertCalled(
					t,
					"MkdirAll",
					"docs",
					mock.Anything,
				)
				cm.Fs.AssertCalled(
					t,
					"MkdirAll",
					filepath.Join("docs", "commands"),
					mock.Anything,
				)
				file, err := cm.Fs.Open(filepath.Join("docs", "errors.md"))
				require.NoError(t, err)
				markdownErrors := make([]byte, 28)
				_, err = file.Read(markdownErrors)
				require.NoError(t, err)
				assert.Contains(t, string(markdownErrors), "# Slack CLI errors reference")
			},
		},
		"when creating the default docs directory fails": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Fs.On(
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "docs"),
					mock.Anything,
				).Return(
					errors.New("no write permission"),
				)
			},
			CmdArgs:              []string{},
			ExpectedErrorStrings: []string{"no write permission"},
		},
		"when creating the default commands directory fails": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Fs.On(
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "docs"),
					mock.Anything,
				).Return(
					nil,
				)
				cm.Fs.On(
					"MkdirAll",
					filepath.Join(slackdeps.MockWorkingDirectory, "docs", "commands"),
					mock.Anything,
				).Return(
					errors.New("no write permission"),
				)
			},
			CmdArgs:              []string{},
			ExpectedErrorStrings: []string{"no write permission"},
		},
		"when generating docs fails": {
			Setup: func(t *testing.T, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.Cobra.On(
					"GenMarkdownTree",
					mock.Anything,
					mock.Anything,
				).Return(
					errors.New("failed to generate docs"),
				)
			},
			CmdArgs:              []string{},
			ExpectedErrorStrings: []string{"failed to generate docs"},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		return NewCommand(clients)
	})
}
