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

package datastore

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCountCommandPreRun(t *testing.T) {
	tests := map[string]struct {
		mockFlagForce        bool
		mockManifestResponse types.SlackYaml
		mockManifestError    error
		mockManifestSource   config.ManifestSource
		mockWorkingDirectory string
		expectedError        error
	}{
		"continues if the application is hosted on slack": {
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.SlackHosted,
					},
				},
			},
			mockManifestError:    nil,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        nil,
		},
		"errors if the application is not hosted on slack": {
			mockManifestResponse: types.SlackYaml{
				AppManifest: types.AppManifest{
					Settings: &types.AppSettings{
						FunctionRuntime: types.Remote,
					},
				},
			},
			mockManifestError:    nil,
			mockManifestSource:   config.ManifestSourceLocal,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrAppNotHosted),
		},
		"continues if the force flag is used in a project": {
			mockFlagForce:        true,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        nil,
		},
		"errors if the project manifest cannot be retrieved": {
			mockManifestResponse: types.SlackYaml{},
			mockManifestError:    slackerror.New(slackerror.ErrSDKHookInvocationFailed),
			mockManifestSource:   config.ManifestSourceLocal,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrSDKHookInvocationFailed),
		},
		"errors if the command is not run in a project": {
			mockManifestResponse: types.SlackYaml{},
			mockManifestError:    slackerror.New(slackerror.ErrSDKHookNotFound),
			mockWorkingDirectory: "",
			expectedError:        slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
		"errors if the manifest source is set to remote": {
			mockManifestSource:   config.ManifestSourceRemote,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrAppNotHosted),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On(
				"GetManifestLocal",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tt.mockManifestResponse,
				tt.mockManifestError,
			)
			clientsMock.AppClient.Manifest = manifestMock
			projectConfigMock := config.NewProjectConfigMock()
			projectConfigMock.On(
				"GetManifestSource",
				mock.Anything,
			).Return(
				tt.mockManifestSource,
				nil,
			)
			clientsMock.Config.ProjectConfig = projectConfigMock
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.Config.ForceFlag = tt.mockFlagForce
				cf.SDKConfig.WorkingDirectory = tt.mockWorkingDirectory
			})
			cmd := NewCountCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tt.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tt.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCountCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"default to the empty expression when no expression is passed": {
			CmdArgs: []string{"--datastore", "tasks"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("AppsDatastoreCount", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreCountResult{Datastore: "tasks", Count: 12}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountTotal, []string{"12"})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountDatastore, []string{"tasks"})
				cm.API.AssertCalled(
					t,
					"AppsDatastoreCount",
					mock.Anything,
					mock.Anything,
					types.AppDatastoreCount{Datastore: "tasks", App: "A001"},
				)
			},
		},
		"pass an empty expression through arguments": {
			CmdArgs: []string{`{"datastore":"tasks"}`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("AppsDatastoreCount", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreCountResult{Datastore: "tasks", Count: 12}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountTotal, []string{"12"})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountDatastore, []string{"tasks"})
				cm.API.AssertCalled(
					t,
					"AppsDatastoreCount",
					mock.Anything,
					mock.Anything,
					types.AppDatastoreCount{Datastore: "tasks", App: "A001"},
				)
			},
		},
		"pass an expression through arguments": {
			CmdArgs: []string{
				`{"datastore":"tasks","expression":"#task_id < :num","expression_attributes":{"#task_id":"task_id"},"expression_values":{":num":"3"}}`,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("AppsDatastoreCount", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreCountResult{Datastore: "tasks", Count: 12}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountTotal, []string{"12"})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountDatastore, []string{"tasks"})
				cm.API.AssertCalled(
					t,
					"AppsDatastoreCount",
					mock.Anything,
					mock.Anything,
					types.AppDatastoreCount{
						Datastore:  "tasks",
						App:        "A001",
						Expression: "#task_id < :num",
						ExpressionAttributes: map[string]interface{}{
							"#task_id": "task_id",
						},
						ExpressionValues: map[string]interface{}{
							":num": "3",
						},
					},
				)
			},
		},
		"pass an extended expression through arguments": {
			CmdArgs: []string{
				`{"datastore":"Todos","app":"A001","expression":"#task_id < :num AND #status = :progress","expression_attributes":{"#task_id":"task_id","#status":"status"},"expression_values":{":num":"3",":progress":"wip"}}`,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("AppsDatastoreCount", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreCountResult{Datastore: "tasks", Count: 12}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountTotal, []string{"12"})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountDatastore, []string{"tasks"})
				cm.API.AssertCalled(
					t,
					"AppsDatastoreCount",
					mock.Anything,
					mock.Anything,
					types.AppDatastoreCount{
						Datastore:  "Todos",
						App:        "A001",
						Expression: "#task_id < :num AND #status = :progress",
						ExpressionAttributes: map[string]interface{}{
							"#task_id": "task_id",
							"#status":  "status",
						},
						ExpressionValues: map[string]interface{}{
							":num":      "3",
							":progress": "wip",
						},
					},
				)
			},
		},
		"pass an empty expression through prompts": {
			CmdArgs: []string{"--unstable"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						Datastores: map[string]types.ManifestDatastore{"numbers": {PrimaryKey: "n"}},
					},
				}, nil)
				cm.AppClient.Manifest = manifestMock
				cm.IO.On("SelectPrompt", mock.Anything, "Select a datastore", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: cm.Config.Flags.Lookup("name"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "numbers",
					Index:  0,
				}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter an expression", iostreams.InputPromptConfig{
					Required: false,
				}).Return("")
				cm.API.On("AppsDatastoreCount", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreCountResult{Datastore: "numbers", Count: 12}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountTotal, []string{"12"})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountDatastore, []string{"numbers"})
				cm.API.AssertCalled(
					t,
					"AppsDatastoreCount",
					mock.Anything,
					mock.Anything,
					types.AppDatastoreCount{Datastore: "numbers", App: "A001"},
				)
			},
			Teardown: func() {
				unstableFlag = false
			},
		},
		"pass an extended expression through prompts": {
			CmdArgs: []string{"--unstable"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				manifestMock := &app.ManifestMockObject{}
				manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
					AppManifest: types.AppManifest{
						Datastores: map[string]types.ManifestDatastore{
							"numbers": {
								PrimaryKey: "n",
								Attributes: map[string]types.ManifestAttribute{
									"n":     {Type: "number"},
									"prime": {Type: "boolean"},
								},
							},
						},
					},
				}, nil)
				cm.AppClient.Manifest = manifestMock
				cm.IO.On("SelectPrompt", mock.Anything, "Select a datastore", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: cm.Config.Flags.Lookup("name"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "numbers",
					Index:  0,
				}, nil)
				cm.IO.On("InputPrompt", mock.Anything, "Enter an expression", iostreams.InputPromptConfig{
					Required: false,
				}).Return("#n < :num AND #n <> :zero AND #prime = :bool")
				cm.IO.On("InputPrompt", mock.Anything, "Enter a value for ':num'", iostreams.InputPromptConfig{
					Required: true,
				}).Return("12")
				cm.IO.On("InputPrompt", mock.Anything, "Enter a value for ':zero'", iostreams.InputPromptConfig{
					Required: true,
				}).Return("0")
				cm.IO.On("InputPrompt", mock.Anything, "Enter a value for ':bool'", iostreams.InputPromptConfig{
					Required: true,
				}).Return("true")
				cm.API.On("AppsDatastoreCount", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreCountResult{Datastore: "numbers", Count: 6}, nil)
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountSuccess, mock.Anything)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountTotal, []string{"6"})
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.DatastoreCountDatastore, []string{"numbers"})
				cm.API.AssertCalled(
					t,
					"AppsDatastoreCount",
					mock.Anything,
					mock.Anything,
					types.AppDatastoreCount{
						Datastore:  "numbers",
						App:        "A001",
						Expression: "#n < :num AND #n <> :zero AND #prime = :bool",
						ExpressionAttributes: map[string]interface{}{
							"#n":     "n",
							"#prime": "prime",
						},
						ExpressionValues: map[string]interface{}{
							":num":  "12",
							":zero": "0",
							":bool": "true",
						},
					},
				)
			},
			Teardown: func() {
				unstableFlag = false
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCountCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			clients.Config.SetFlags(cmd)
			return nil
		}
		appSelectMock := prompts.NewAppSelectMock()
		appSelectPromptFunc = appSelectMock.AppSelectPrompt
		appSelectMock.On("AppSelectPrompt").
			Return(prompts.SelectedApp{App: types.App{AppID: "A001"}}, nil)
		return cmd
	})
}
