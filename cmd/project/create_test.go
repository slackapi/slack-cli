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

package project

import (
	"context"
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/create"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type CreateClientMock struct {
	mock.Mock
}

func (m *CreateClientMock) Create(ctx context.Context, clients *shared.ClientFactory, createArgs create.CreateArgs) (string, error) {
	args := m.Called(ctx, clients, createArgs)
	return args.String(0), args.Error(1)
}

func TestCreateCommand(t *testing.T) {
	var createClientMock *CreateClientMock

	testutil.TableTestCommand(t, testutil.CommandTests{
		"creates a bolt application from prompts": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsTTY").Return(true)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("InputPrompt", mock.Anything, "Name your app:", mock.Anything).
					Return("my-app", nil)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					AppPath:  "my-app",
					Template: template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				cm.IO.AssertCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"creates a deno application from flags": {
			CmdArgs: []string{"--template", "slack-samples/deno-starter-template"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsTTY").Return(true)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/deno-starter-template",
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/deno-starter-template",
						},
						nil,
					)
				cm.IO.On("InputPrompt", mock.Anything, "Name your app:", mock.Anything).
					Return("my-deno-app", nil)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/deno-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					AppPath:  "my-deno-app",
					Template: template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				cm.IO.AssertCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"creates an agent app using agent argument shortcut": {
			CmdArgs: []string{"agent"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsTTY").Return(true)
				// Should skip category prompt and go directly to template selection
				cm.IO.On("SelectPrompt", mock.Anything, "Select a template:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Starter Agent
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a Bolt framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Node.js
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select an agent framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Claude Agent SDK
						},
						nil,
					)
				cm.IO.On("InputPrompt", mock.Anything, "Name your app:", mock.Anything).
					Return("my-agent", nil)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-agent")
				require.NoError(t, err)
				template.SetSubdir("claude-agent-sdk")
				expected := create.CreateArgs{
					AppPath:  "my-agent",
					Template: template,
					Subdir:   "claude-agent-sdk",
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that category prompt was NOT called
				cm.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything)
				cm.IO.AssertCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"creates an agent app with app name using agent argument": {
			CmdArgs: []string{"agent", "my-agent-app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Should skip category prompt and go directly to template selection
				cm.IO.On("SelectPrompt", mock.Anything, "Select a template:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Starter Agent
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a Bolt framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  1, // Select Python
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select an agent framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Claude Agent SDK
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-python-starter-agent")
				require.NoError(t, err)
				template.SetSubdir("claude-agent-sdk")
				expected := create.CreateArgs{
					AppPath:  "my-agent-app",
					Template: template,
					Subdir:   "claude-agent-sdk",
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that category prompt was NOT called
				cm.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything)
				// Verify that name prompt was NOT called since name was provided as arg
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"creates a pydantic ai agent app": {
			CmdArgs: []string{"my-pydantic-app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 1}, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a template:", mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 1}, nil) // Select Support Agent
				cm.IO.On("SelectPrompt", mock.Anything, "Select a Bolt framework:", mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 1}, nil) // Select Bolt for Python
				cm.IO.On("SelectPrompt", mock.Anything, "Select an agent framework:", mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Index: 2}, nil) // Select Pydantic AI
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-python-support-agent")
				require.NoError(t, err)
				template.SetSubdir("pydantic-ai")
				expected := create.CreateArgs{
					AppPath:  "my-pydantic-app",
					Template: template,
					Subdir:   "pydantic-ai",
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
			},
		},
		"creates an app named agent when template flag is provided": {
			CmdArgs: []string{"agent", "--template", "slack-samples/bolt-js-starter-template"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/bolt-js-starter-template",
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/bolt-js-starter-template",
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					AppPath:  "agent",
					Template: template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that name prompt was NOT called since name was provided as arg
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"creates an app named agent using name flag without triggering shortcut": {
			CmdArgs: []string{"--name", "agent"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Should prompt for category since agent shortcut is NOT triggered
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select starter app
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Node.js
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					AppPath:     "agent",
					DisplayName: "agent",
					Template:    template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that category prompt WAS called (shortcut was not triggered)
				cm.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything)
				// Verify that name prompt was NOT called since --name flag was provided
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"creates an agent app with name flag overriding positional arg": {
			CmdArgs: []string{"agent", "--name", "my-custom-name"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Should skip category prompt due to agent shortcut
				cm.IO.On("SelectPrompt", mock.Anything, "Select a template:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Starter Agent
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a Bolt framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Node.js
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select an agent framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Claude Agent SDK
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-agent")
				require.NoError(t, err)
				template.SetSubdir("claude-agent-sdk")
				expected := create.CreateArgs{
					AppPath:     "my-custom-name", // --name flag used as path when no positional arg
					DisplayName: "my-custom-name",
					Template:    template,
					Subdir:      "claude-agent-sdk",
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that category prompt was NOT called (shortcut was triggered)
				cm.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything)
			},
		},
		"name flag overrides positional app name argument": {
			CmdArgs: []string{"my-project", "--name", "my-name"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select starter app
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Node.js
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					AppPath:     "my-project", // positional arg preserved as path
					DisplayName: "my-name",    // --name flag sets manifest display name
					Template:    template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that name prompt was NOT called since --name flag was provided
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"name flag overrides positional app name argument with agent shortcut": {
			CmdArgs: []string{"agent", "my-project", "--name", "my-name"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// Should skip category prompt due to agent shortcut
				cm.IO.On("SelectPrompt", mock.Anything, "Select a template:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Starter Agent
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a Bolt framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Node.js
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select an agent framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0, // Select Claude Agent SDK
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-agent")
				require.NoError(t, err)
				template.SetSubdir("claude-agent-sdk")
				expected := create.CreateArgs{
					AppPath:     "my-project", // positional arg preserved as path
					DisplayName: "my-name",    // --name flag sets manifest display name
					Template:    template,
					Subdir:      "claude-agent-sdk",
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that category prompt was NOT called (agent shortcut was triggered)
				cm.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything)
				// Verify that name prompt was NOT called since --name flag was provided
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"name prompt includes placeholder with generated name": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsTTY").Return(true)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("InputPrompt", mock.Anything, "Name your app:", mock.Anything).
					Return("my-app", nil)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Verify that InputPrompt was called with a config that has a non-empty Placeholder
				cm.IO.AssertCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.MatchedBy(func(cfg iostreams.InputPromptConfig) bool {
					return cfg.Placeholder != ""
				}))
			},
		},
		"user accepts default name from prompt": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsTTY").Return(true)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				// Return empty string to simulate pressing Enter (accepting default)
				cm.IO.On("InputPrompt", mock.Anything, "Name your app:", mock.Anything).
					Return("", nil)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
				// When the user accepts the default (empty return), the generated name is used
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.MatchedBy(func(args create.CreateArgs) bool {
					return args.AppPath != ""
				}))
			},
		},
		"non-TTY without name falls back to generated name": {
			CmdArgs: []string{"--template", "slack-samples/bolt-js-starter-template"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				// IsTTY defaults to false via AddDefaultMocks, simulating piped output
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/bolt-js-starter-template",
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/bolt-js-starter-template",
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				// Should NOT prompt for name since not a TTY
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
				// Should still call Create with a non-empty generated name
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.MatchedBy(func(args create.CreateArgs) bool {
					return args.AppPath != ""
				}))
			},
		},
		"positional arg skips name prompt": {
			CmdArgs: []string{"my-project"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Prompt: true,
							Index:  0,
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				expected := create.CreateArgs{
					AppPath:  "my-project",
					Template: template,
				}
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, expected)
				// Verify that name prompt was NOT called since name was provided as positional arg
				cm.IO.AssertNotCalled(t, "InputPrompt", mock.Anything, "Name your app:", mock.Anything)
			},
		},
		"subdir without template flag returns error": {
			CmdArgs: []string{"--subdir", "apps/my-app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --subdir flag requires the --template flag"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"passes subdir flag to create function": {
			CmdArgs: []string{"--template", "slack-samples/bolt-js-starter-template", "--subdir", "apps/my-app"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/bolt-js-starter-template",
						},
						nil,
					)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a framework:", mock.Anything, mock.Anything).
					Return(
						iostreams.SelectPromptResponse{
							Flag:   true,
							Option: "slack-samples/bolt-js-starter-template",
						},
						nil,
					)
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return("", nil)
				CreateFunc = createClientMock.Create
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				template, err := create.ResolveTemplateURL("slack-samples/bolt-js-starter-template")
				require.NoError(t, err)
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.MatchedBy(func(args create.CreateArgs) bool {
					return args.AppPath != "" && args.Template == template && args.Subdir == "apps/my-app"
				}))
			},
		},
		"list flag ignores subdir": {
			CmdArgs: []string{"--list", "--subdir", "foo"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedOutputs: []string{
				"Getting started",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"lists all templates with --list flag": {
			CmdArgs: []string{"--list"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedOutputs: []string{
				"Getting started",
				"slack-samples/bolt-js-starter-template",
				"slack-samples/bolt-python-starter-template",
				"Starter agent",
				"slack-samples/bolt-js-starter-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-js-starter-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-starter-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-python-starter-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-starter-agent --subdir pydantic-ai",
				"Support agent",
				"slack-samples/bolt-js-support-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-js-support-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-support-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-python-support-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-support-agent --subdir pydantic-ai",
				"Automation apps",
				"slack-samples/bolt-js-custom-function-template",
				"slack-samples/bolt-python-custom-function-template",
				"slack-samples/deno-starter-template",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"lists agent templates with agent --list flag": {
			CmdArgs: []string{"agent", "--list"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedOutputs: []string{
				"Starter agent",
				"slack-samples/bolt-js-starter-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-js-starter-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-starter-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-python-starter-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-starter-agent --subdir pydantic-ai",
				"Support agent",
				"slack-samples/bolt-js-support-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-js-support-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-support-agent --subdir claude-agent-sdk",
				"slack-samples/bolt-python-support-agent --subdir openai-agents-sdk",
				"slack-samples/bolt-python-support-agent --subdir pydantic-ai",
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
				output := cm.GetCombinedOutput()
				assert.NotContains(t, output, "Getting started")
				assert.NotContains(t, output, "Automation apps")
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func TestCreateCommand_confirmExternalTemplateSelection(t *testing.T) {
	// test cases
	tests := map[string]struct {
		templateSource string
		setup          func(cm *shared.ClientsMock, scm *config.SystemConfigMock)
		expect         func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock)
	}{
		"trust, untrusted source": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(true, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "selection should be confirmed")
				}
				// Should not have prompted
				output := cm.GetCombinedOutput()
				assert.NotContains(t, output, "You are trying to use code published by an unknown author")
			},
		},
		"don't trust, trusted source": {
			templateSource: "slack-samples/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "should confirm the selection despite not trusting source since template source is trusted")
				}
			},
		},
		"don't trust, untrusted source, confirm": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {

				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 0, Option: "Yes"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "should prompt")
				}

				// Should have prompted
				output := cm.GetCombinedOutput()
				assert.Contains(t, output, "You are trying to use code published by an unknown author")
			},
		},
		"don't trust, untrusted source, do not confirm": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 2, Option: "No"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				assert.False(t, confirmed)
				if err != nil {
					assert.Fail(t, "should have returned with no error")
				}

				// Should have prompted
				output := cm.GetCombinedOutput()
				assert.Contains(t, output, "You are trying to use code published by an unknown author")
			},
		},
		"don't trust, untrusted source, confirm and don't ask again": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				scm.On("SetTrustUnknownSources", mock.Anything, true).Return(nil)

				// Proceed when prompted and select don't ask again
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 1, Option: "Yes, don't ask again"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				// Should have prompted
				output := cm.GetCombinedOutput()
				assert.Contains(t, output, "You are trying to use code published by an unknown author")

				// should have confirmed
				assert.True(t, confirmed)
				if err != nil {
					assert.Fail(t, "should have returned with no error")
				}
				// Should have tried to set the source after confirmation
				scm.AssertCalled(t, "SetTrustUnknownSources", mock.Anything, true)
			},
		},
		"don't trust, untrusted source, confirm and prompt error": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)
				scm.On("SetTrustUnknownSources", mock.Anything, true).Return(nil)

				// Proceed when prompted, don't ask again
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 1, Option: "Yes, don't ask again"}, slackerror.New("error"))
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				// does not return confirmed
				assert.False(t, confirmed)
				if err == nil {
					assert.Fail(t, "should have returned the prompt error")
				}
				// Should not have tried to set the source after confirmation
				scm.AssertNotCalled(t, "SetTrustUnknownSources", mock.Anything, true)
			},
		},
		"don't trust sources, untrusted source, confirm and config write error": {
			templateSource: "untrusted-source/app",
			setup: func(cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				scm.On("GetTrustUnknownSources", mock.Anything).Return(false, nil)

				// Set trust_unknown_sources should fail with an error
				scm.On("SetTrustUnknownSources", mock.Anything, true).Return(slackerror.New("an error!"))

				// Proceed when prompted, select don't ask again
				cm.IO.On("SelectPrompt", mock.Anything, "Proceed?", mock.Anything, mock.Anything).Return(iostreams.SelectPromptResponse{Index: 1, Option: "Yes, don't ask again"}, nil)
			},
			expect: func(confirmed bool, err error, cm *shared.ClientsMock, scm *config.SystemConfigMock) {
				// should confirm with an error
				assert.True(t, confirmed)
				if err == nil {
					assert.Fail(t, "should have returned an error")
				}
				assert.ErrorContains(t, err, "an error!")
			},
		},
	}

	// test!
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup
			cm := shared.NewClientsMock()
			cm.AddDefaultMocks()
			scm := &config.SystemConfigMock{}
			tc.setup(cm, scm)
			cm.Config.SystemConfig = scm
			clients := shared.NewClientFactory(cm.MockClientFactory())
			cmd := NewCreateCommand(clients)
			testutil.MockCmdIO(clients.IO, cmd)

			// test
			template, err := create.ResolveTemplateURL(tc.templateSource)
			require.NoError(t, err)
			confirmed, err := confirmExternalTemplateSelection(cmd, clients, template)
			tc.expect(confirmed, err, cm, scm)
		})
	}
}

func TestCreateCommand_AppFlag(t *testing.T) {
	var createClientMock *CreateClientMock

	testutil.TableTestCommand(t, testutil.CommandTests{
		"app flag without template flag returns error": {
			CmdArgs: []string{"my-app", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --app flag requires the --template flag when used with create"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with environment-style value returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --app flag requires an app ID when used with create"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with lowercase id returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "a0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --app flag requires an app ID when used with create"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"environment flag without app flag returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--environment", "deployed"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --environment flag requires the --app flag when used with create"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"invalid environment flag returns error": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789", "--environment", "invalid"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				CreateFunc = createClientMock.Create
			},
			ExpectedErrorStrings: []string{"The --environment flag must be either 'local' or 'deployed'"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"app flag with template creates project and links a deployed app": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789", "--environment", "deployed"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(t.TempDir(), nil)
				CreateFunc = createClientMock.Create

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{mockCreateLinkAuth}, nil)
				cm.AddDefaultMocks()
				setupCreateLinkMocks(t, ctx, cm, cf)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()
				cm.IO.On("SelectPrompt", mock.Anything, "Select the existing app team", mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: mockCreateLinkAuth.TeamDomain}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter the existing app ID", mock.Anything).
					Return("A0123456789", nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Choose the app environment", mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "deployed"}, nil)
				cm.API.On("GetAppStatus", mock.Anything, mockCreateLinkAuth.Token, []string{"A0123456789"}, mockCreateLinkAuth.TeamID).
					Return(api.GetAppStatusResult{}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
				saved, err := cm.AppClient.GetDeployed(ctx, mockCreateLinkAuth.TeamID)
				require.NoError(t, err)
				assert.Equal(t, "A0123456789", saved.AppID)
				assert.Equal(t, mockCreateLinkAuth.TeamID, saved.TeamID)
				assert.False(t, saved.IsDev)
			},
		},
		"app flag without environment links a local app via prompt": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				createClientMock = new(CreateClientMock)
				createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(t.TempDir(), nil)
				CreateFunc = createClientMock.Create

				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{mockCreateLinkAuth}, nil)
				cm.AddDefaultMocks()
				setupCreateLinkMocks(t, ctx, cm, cf)
				cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil).Maybe()
				cm.IO.On("SelectPrompt", mock.Anything, "Select the existing app team", mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Option: mockCreateLinkAuth.TeamDomain}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter the existing app ID", mock.Anything).
					Return("A0123456789", nil)
				cm.IO.On("SelectPrompt", mock.Anything, "Choose the app environment", mock.Anything, mock.Anything, mock.Anything).
					Return(iostreams.SelectPromptResponse{Prompt: true, Option: "local"}, nil)
				cm.API.On("GetAppStatus", mock.Anything, mockCreateLinkAuth.Token, []string{"A0123456789"}, mockCreateLinkAuth.TeamID).
					Return(api.GetAppStatusResult{}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
				saved, err := cm.AppClient.GetLocal(ctx, mockCreateLinkAuth.TeamID)
				require.NoError(t, err)
				assert.Equal(t, "A0123456789", saved.AppID)
				assert.Equal(t, mockCreateLinkAuth.TeamID, saved.TeamID)
				assert.True(t, saved.IsDev)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

func TestCreateCommand_AppFlag_FetchesRemoteManifest(t *testing.T) {
	var createClientMock *CreateClientMock

	mockAuth := types.SlackAuth{
		Token:      "xoxp-test-token",
		TeamDomain: "test-team",
		TeamID:     "T001",
		UserID:     "U001",
	}
	mockManifest := types.SlackYaml{
		AppManifest: types.AppManifest{
			DisplayInformation: types.DisplayInformation{
				Name:        "My Remote App",
				Description: "An app from remote settings",
			},
		},
	}

	setupAppFlagMocks := func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) string {
		projectDir := t.TempDir()
		createClientMock = new(CreateClientMock)
		createClientMock.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(projectDir, nil)
		CreateFunc = createClientMock.Create

		cm.Os.On("Getwd").Return(projectDir, nil)

		err := cm.Fs.MkdirAll(projectDir+"/.slack", 0755)
		require.NoError(t, err)
		err = afero.WriteFile(cm.Fs, projectDir+"/.slack/hooks.json", []byte("{}"), 0644)
		require.NoError(t, err)

		cm.IO.On("SelectPrompt", mock.Anything, "Select a category:", mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Flag: true, Option: "slack-samples/bolt-js-starter-template"}, nil)

		cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{mockAuth}, nil)
		cm.Auth.On("AuthWithTeamID", mock.Anything, mock.Anything).Return(mockAuth, nil)
		cm.IO.On("SelectPrompt", mock.Anything, "Select the existing app team", mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Prompt: true, Index: 0, Option: mockAuth.TeamDomain}, nil)
		cm.IO.On("SelectPrompt", mock.Anything, "Choose the app environment", mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Prompt: true, Option: "local"}, nil)

		cm.API.On("GetAppStatus", mock.Anything, mockAuth.Token, []string{"A0123456789"}, mockAuth.TeamID).
			Return(api.GetAppStatusResult{}, nil)

		return projectDir
	}

	var projectDir string

	testutil.TableTestCommand(t, testutil.CommandTests{
		"fetches remote manifest after linking app": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789", "--environment", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				projectDir = setupAppFlagMocks(t, ctx, cm, cf)

				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mockAuth.Token, "A0123456789").
					Return(mockManifest, nil)
				cf.AppClient().Manifest = manifestMock
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)

				manifestData, err := afero.ReadFile(cm.Fs, projectDir+"/manifest.json")
				require.NoError(t, err)
				assert.Contains(t, string(manifestData), `"name": "My Remote App"`)
				assert.Contains(t, string(manifestData), `"description": "An app from remote settings"`)
			},
		},
		"returns error on manifest fetch failure": {
			CmdArgs: []string{"my-app", "--template", "slack-samples/bolt-js-starter-template", "--app", "A0123456789", "--environment", "local"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				projectDir = setupAppFlagMocks(t, ctx, cm, cf)

				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mockAuth.Token, "A0123456789").
					Return(types.SlackYaml{}, slackerror.New("network error"))
				cf.AppClient().Manifest = manifestMock
			},
			ExpectedErrorStrings: []string{"network error"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				createClientMock.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCreateCommand(cf)
	})
}

var mockCreateLinkAuth = types.SlackAuth{
	Token:        "xoxp-example",
	TeamDomain:   "team1",
	TeamID:       "T001",
	EnterpriseID: "E001",
	UserID:       "U001",
}

func setupCreateLinkMocks(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
	projectDirPath := slackdeps.MockWorkingDirectory
	cm.Os.On("Getwd").Return(projectDirPath, nil)
	cm.Auth.On("AuthWithTeamID", mock.Anything, mock.Anything).Return(mockCreateLinkAuth, nil)

	if _, err := config.CreateProjectConfigDir(ctx, cm.Fs, projectDirPath); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to create the project config directory: %s", err))
	}
	if _, err := config.CreateProjectHooksJSONFile(cm.Fs, projectDirPath, []byte("{}")); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to create the hooks file: %s", err))
	}
	if err := config.SetManifestSource(ctx, cm.Fs, cm.Os, config.ManifestSourceRemote); err != nil {
		require.FailNow(t, fmt.Sprintf("Failed to set the manifest source: %s", err))
	}

	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
		Return(types.SlackYaml{}, nil)
	manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
		Return(types.SlackYaml{}, nil)
	cf.AppClient().Manifest = manifestMock
}
