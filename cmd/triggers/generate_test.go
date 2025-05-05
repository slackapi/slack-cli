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

package triggers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type unitTest struct {
	name                   string
	triggersListResponse   []types.DeployedTrigger
	triggersCreateResponse types.DeployedTrigger
	globResponse           []string
	check                  func(*testing.T, *types.DeployedTrigger, error)
}

func Test_TriggerGenerate_accept_prompt(t *testing.T) {
	projDir, _ := os.Getwd()
	tt := unitTest{
		name:                   "Accept prompt to create a trigger",
		triggersListResponse:   []types.DeployedTrigger{}, // no existing triggers
		globResponse:           []string{fmt.Sprintf("%v/triggers/trigger.ts", projDir)},
		triggersCreateResponse: types.DeployedTrigger{Name: "trigger name", ID: "Ft123", Type: "shortcut"},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			assert.Equal(t, trigger.ID, "Ft123")
			assert.Nil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, nil /*trigger create error*/)

		clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("trigger-def"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "trigger/file.ts",
			Index:  0,
		}, nil)
		clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
		clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
			Return(types.PermissionEveryone, []string{}, nil).Once()

		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
	})
}

func Test_TriggerGenerate_decline_prompt(t *testing.T) {
	projDir, _ := os.Getwd()
	tt := unitTest{
		name:                   "Decline prompt to create a trigger",
		triggersListResponse:   []types.DeployedTrigger{}, // no existing triggers
		triggersCreateResponse: types.DeployedTrigger{},
		globResponse:           []string{fmt.Sprintf("%v/triggers/trigger.ts", projDir)},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			assert.Nil(t, trigger)
			assert.Nil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, nil /*trigger create error*/)
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("trigger-def"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "Do not create a trigger",
			Index:  1,
		}, nil)

		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
	})
}

func Test_TriggerGenerate_skip_prompt(t *testing.T) {
	tt := unitTest{
		name:                 "Skip prompt if app has at least one trigger",
		triggersListResponse: []types.DeployedTrigger{{Name: "existing trigger", ID: "Ft456", Type: "scheduled"}},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			assert.Nil(t, trigger)
			assert.Nil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, nil /*trigger create error*/)
		clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
		clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
			Return(types.PermissionEveryone, []string{}, nil).Once()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
	})
}

func Test_TriggerGenerate_handle_error(t *testing.T) {
	projDir, _ := os.Getwd()
	tt := unitTest{
		name:                   "Handle error on trigger creation",
		triggersListResponse:   []types.DeployedTrigger{}, // no existing triggers
		globResponse:           []string{fmt.Sprintf("%v/triggers/trigger.ts", projDir)},
		triggersCreateResponse: types.DeployedTrigger{Name: "trigger name", ID: "Ft123", Type: "shortcut"},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			assert.Nil(t, trigger)
			assert.NotNil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, fmt.Errorf("something went wrong") /*trigger create error*/)
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("trigger-def"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "triggers/trigger.ts",
			Index:  0,
		}, nil)
		clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
		clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
			Return(types.PermissionEveryone, []string{}, nil).Once()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
	})
}

func Test_TriggerGenerate_handle_invalid_paths(t *testing.T) {
	projDir, _ := os.Getwd()
	tt := unitTest{
		name:                   "Omit invalid paths in triggers project directory",
		triggersListResponse:   []types.DeployedTrigger{},                                 // no existing triggers
		globResponse:           []string{fmt.Sprintf("%v/triggers/trigger.ts~", projDir)}, // invalid trigger
		triggersCreateResponse: types.DeployedTrigger{Name: "trigger name", ID: "Ft123", Type: "shortcut"},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			// should not create a trigger or error but rather treat globResponse as having 0 valid triggers
			assert.Nil(t, trigger)
			assert.Nil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, nil)

		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
	})
}

func Test_TriggerGenerate_Config_TriggerPaths_Default(t *testing.T) {
	projDir, _ := os.Getwd()
	tt := unitTest{
		name:                   "Should use default when 'trigger-path' is missing",
		triggersListResponse:   []types.DeployedTrigger{}, // no existing triggers
		globResponse:           []string{fmt.Sprintf("%v/triggers/trigger.ts", projDir)},
		triggersCreateResponse: types.DeployedTrigger{Name: "trigger name", ID: "Ft123", Type: "shortcut"},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			assert.Equal(t, trigger.ID, "Ft123")
			assert.Nil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, nil /*trigger create error*/)
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("trigger-def"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "triggers/trigger.ts",
			Index:  0,
		}, nil)
		clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
		clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
			Return(types.PermissionEveryone, []string{}, nil).Once()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		clients.SDKConfig.Config.TriggerPaths = []string{} // No 'trigger-path' is set
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
		clientsMock.Os.AssertNumberOfCalls(t, "Glob", 1) // Assert default 'trigger-path' is used as Glob pattern
	})
}

func Test_TriggerGenerate_Config_TriggerPaths_Custom(t *testing.T) {
	projDir, _ := os.Getwd()
	tt := unitTest{
		name:                   "Should use custom 'trigger-path' in hooks.json",
		triggersListResponse:   []types.DeployedTrigger{}, // no existing triggers
		globResponse:           []string{fmt.Sprintf("%v/triggers/trigger.ts", projDir)},
		triggersCreateResponse: types.DeployedTrigger{Name: "trigger name", ID: "Ft123", Type: "shortcut"},
		check: func(t *testing.T, trigger *types.DeployedTrigger, err error) {
			assert.Equal(t, trigger.ID, "Ft123")
			assert.Nil(t, err)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		ctx, clientsMock := prepareMocks(t, tt.triggersListResponse, tt.globResponse, tt.triggersCreateResponse, nil /*trigger create error*/)
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a trigger definition file:", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("trigger-def"),
		})).Return(iostreams.SelectPromptResponse{
			Prompt: true,
			Option: "triggers/trigger.ts",
			Index:  0,
		}, nil)
		clientsMock.API.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)
		clientsMock.API.On("TriggerPermissionsList", mock.Anything, mock.Anything, mock.Anything).
			Return(types.PermissionEveryone, []string{}, nil).Once()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
			clients.SDKConfig = hooks.NewSDKConfigMock()
			clients.SDKConfig.Hooks.GetTrigger = hooks.HookScript{
				Command: "echo {}",
			}
		})

		// Execute test
		clients.SDKConfig.Config.TriggerPaths = []string{"my-triggers/*.ts", "my-triggers/*.js", "my-triggers/*.json"} // Custom 'trigger-path'
		trigger, err := TriggerGenerate(ctx, clients, types.App{})
		tt.check(t, trigger, err)
		clientsMock.Os.AssertNumberOfCalls(t, "Glob", 3) // Assert custom 'trigger-path' is used as Glob pattern
	})
}

func Test_TriggerGenerate_MismatchedFlags(t *testing.T) {
	definitionFile := "triggers/shortcut.ts"
	tests := map[string]struct {
		flags   createCmdFlags
		err     error
		message []string
	}{
		"no errors with no additional property flags": {
			flags: createCmdFlags{
				triggerDef: definitionFile,
			},
			err: nil,
		},
		"no errors with only property flags": {
			flags: createCmdFlags{
				workflow:          "#/workflows/example_workflow",
				title:             "Example trigger",
				description:       "A quick example for testing",
				interactivity:     true,
				interactivityName: "interactor",
			},
			err: nil,
		},
		"error with both a trigger definition and property flag": {
			flags: createCmdFlags{
				triggerDef: definitionFile,
				title:      "Another example trigger",
			},
			err: slackerror.New(slackerror.ErrMismatchedFlags),
		},
		"error with a definition and property flags noting overridden flags": {
			flags: createCmdFlags{
				triggerDef:        definitionFile,
				description:       "A helpful explanation",
				interactivity:     true,
				interactivityName: "interactors",
				title:             "Another example trigger",
				workflow:          "#/workingflow",
			},
			err: slackerror.New(slackerror.ErrMismatchedFlags),
			message: []string{
				"--description",
				"--interactivity ",
				"--interactivity-name",
				"--title",
				"--workflow",
			},
		},
	}
	var createArgs = func(flags createCmdFlags) []string {
		var arguments []string
		if flags.description != "" {
			arguments = append(arguments, "--description")
			arguments = append(arguments, flags.description)
		}
		if flags.interactivity {
			arguments = append(arguments, "--interactivity")
		}
		if flags.interactivityName != "" {
			arguments = append(arguments, "--interactivity-name")
			arguments = append(arguments, flags.interactivityName)
		}
		if flags.title != "" {
			arguments = append(arguments, "--title")
			arguments = append(arguments, flags.title)
		}
		if flags.triggerDef != "" {
			arguments = append(arguments, "--trigger-def")
			arguments = append(arguments, flags.triggerDef)
		}
		if flags.workflow != "" {
			arguments = append(arguments, "--workflow")
			arguments = append(arguments, flags.triggerDef)
		}
		return arguments
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.Os.On("Glob", mock.Anything).Return([]string{definitionFile}, nil)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cmd := NewCreateCommand(clients)
			err := cmd.Flags().Parse(createArgs(tt.flags))
			require.NoError(t, err, "Failed to parse mocked flags")
			clients.Config.SetFlags(cmd)
			err = afero.WriteFile(clients.Fs, definitionFile, []byte(""), 0600)
			require.NoError(t, err, "Cant write apps.json")

			err = validateCreateCmdFlags(ctx, clients, &tt.flags)
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.Equal(t, slackerror.ToSlackError(tt.err).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.Nil(t, err)
			}
			for _, msg := range tt.message {
				assert.Contains(t, err.Error(), msg)
			}
		})
	}
}

func Test_ShowTriggers(t *testing.T) {
	tests := map[string]struct {
		hideTriggersFlag  bool
		isTTY             bool
		hasGetTriggerHook bool
		expectedShown     bool
	}{
		"defaults to shown": {
			hideTriggersFlag:  false,
			isTTY:             true,
			hasGetTriggerHook: true,
			expectedShown:     true,
		},
		"toggled false by flag": {
			hideTriggersFlag:  true,
			isTTY:             true,
			hasGetTriggerHook: true,
			expectedShown:     false,
		},
		"toggled false without interactivity": {
			hideTriggersFlag:  false,
			isTTY:             false,
			hasGetTriggerHook: true,
			expectedShown:     false,
		},
		"toggled false without the hook": {
			hideTriggersFlag:  false,
			isTTY:             true,
			hasGetTriggerHook: false,
			expectedShown:     false,
		},
		"toggled false without interactivity or hook": {
			hideTriggersFlag:  false,
			isTTY:             false,
			hasGetTriggerHook: false,
			expectedShown:     false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("IsTTY").Return(tt.isTTY)
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = hooks.NewSDKConfigMock()
			})
			if !tt.hasGetTriggerHook {
				clients.SDKConfig.Hooks.GetTrigger.Command = ""
			}
			triggersShown := ShowTriggers(clients, tt.hideTriggersFlag)
			assert.Equal(t, tt.expectedShown, triggersShown)
		})
	}
}

func prepareMocks(t *testing.T, triggersListResponse []types.DeployedTrigger, globResponse []string, triggersCreateResponse types.DeployedTrigger, triggersCreateResponseError error) (context.Context, *shared.ClientsMock) {
	ctx := slackcontext.MockContext(t.Context())
	ctx = config.SetContextToken(ctx, "token")

	clientsMock := shared.NewClientsMock()
	clientsMock.API.On("WorkflowsTriggersList", mock.Anything, mock.Anything, mock.Anything).Return(triggersListResponse, "", nil)
	clientsMock.Os.On("Glob", mock.Anything).Return(globResponse, nil)
	clientsMock.API.On("WorkflowsTriggersCreate", mock.Anything, mock.Anything, mock.Anything).Return(triggersCreateResponse, triggersCreateResponseError)
	clientsMock.HookExecutor.On("Execute", mock.Anything, mock.Anything).Return(`{}`, nil)

	clientsMock.AddDefaultMocks()

	return ctx, clientsMock
}
