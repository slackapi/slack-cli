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
	"bufio"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type QueryDatastorePkgMock struct {
	mock.Mock
}

func (m *QueryDatastorePkgMock) Query(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, query types.AppDatastoreQuery) (*logger.LogEvent, error) {
	m.Called(ctx, clients, log, query)
	log.Data["queryResult"] = types.AppDatastoreQueryResult{}
	log.Log("info", "on_query_result")
	return log.SuccessEvent(), nil
}

func TestQueryCommandPreRun(t *testing.T) {
	tests := map[string]struct {
		mockFlagForce        bool
		mockFlagOutput       string
		mockFlagToFile       string
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
			mockManifestSource:   config.ManifestSourceLocal,
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
			mockManifestSource:   config.ManifestSourceLocal,
			expectedError:        slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
		"errors if both the output file and type flag are specified": {
			mockFlagToFile: "output.json",
			mockFlagOutput: "text",
			expectedError: slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage("Output type for --to-file cannot be specified"),
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
			cmd := NewQueryCommand(clients)
			clients.Config.SetFlags(cmd)
			if tt.mockFlagOutput != "" {
				clients.Config.Flags.Lookup("output").Changed = true
			}
			if tt.mockFlagToFile != "" {
				clients.Config.Flags.Lookup("to-file").Changed = true
			}
			err := cmd.PreRunE(cmd, nil)
			if tt.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tt.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQueryCommand(t *testing.T) {
	tests := map[string]struct {
		Setup    func(*shared.ClientsMock)
		Prompts  func(*shared.ClientsMock)
		Teardown func()
		Query    types.AppDatastoreQuery
	}{
		"pass an empty expression through arguments": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, fmt.Sprintf(`{"datastore":"Todos","app":"%s"}`, mockAppID))
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreQuery{
				Datastore: "Todos",
				App:       mockAppID,
			},
		},
		"pass an expression through arguments": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, fmt.Sprintf(`{"datastore":"Todos","app":"%s","expression":"#task_id < :num","expression_attributes":{"#task_id":"task_id"},"expression_values":{":num":"3"}}`, mockAppID))
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreQuery{
				Datastore:  "Todos",
				App:        mockAppID,
				Expression: "#task_id < :num",
				ExpressionAttributes: map[string]interface{}{
					"#task_id": "task_id",
				},
				ExpressionValues: map[string]interface{}{
					":num": "3",
				},
			},
		},
		"pass an extended expression through arguments": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, fmt.Sprintf(`{"datastore":"Todos","app":"%s","expression":"#task_id < :num AND #status = :progress","expression_attributes":{"#task_id":"task_id","#status":"status"},"expression_values":{":num":"3",":progress":"wip"}}`, mockAppID))
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreQuery{
				Datastore:  "Todos",
				App:        mockAppID,
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
		},
		"pass an expression through arguments with limit and cursor": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, fmt.Sprintf(`{"datastore":"Todos","app":"%s","expression":"#task_id < :num","expression_attributes":{"#task_id":"task_id"},"expression_values":{":num":"8"},"limit":4,"cursor":"alwaysmoretodo"}`, mockAppID))
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreQuery{
				Datastore:  "Todos",
				App:        mockAppID,
				Expression: "#task_id < :num",
				ExpressionAttributes: map[string]interface{}{
					"#task_id": "task_id",
				},
				ExpressionValues: map[string]interface{}{
					":num": "8",
				},
				Limit:  4,
				Cursor: "alwaysmoretodo",
			},
		},
		"pass an empty expression through prompts": {
			Setup: func(clientsMock *shared.ClientsMock) {
				os.Args = append(os.Args, `--unstable`)
			},
			Prompts: func(clientsMock *shared.ClientsMock) {
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a datastore", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("datastore"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "Todos",
					Index:  0,
				}, nil)
				clientsMock.IO.On("InputPrompt", mock.Anything, "Enter an expression", iostreams.InputPromptConfig{
					Required: false,
				}).Return("")
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreQuery{
				Datastore: "Todos",
				App:       mockAppID,
			},
		},
		"pass an extended expression through prompts": {
			Setup: func(clientsMock *shared.ClientsMock) {
				os.Args = append(os.Args, `--unstable`)
			},
			Prompts: func(clientsMock *shared.ClientsMock) {
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a datastore", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("datastore"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "Todos",
					Index:  0,
				}, nil)
				clientsMock.IO.On("InputPrompt", mock.Anything, "Enter an expression", iostreams.InputPromptConfig{
					Required: false,
				}).Return("#task_id < :num AND #task_id <> :zero AND #notes = :progress")
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an attribute for '#notes'", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("attributes"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "status",
					Index:  0,
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an attribute for '#task_id'", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("attributes"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "task_id",
					Index:  1,
				}, nil)
				clientsMock.IO.On("InputPrompt", mock.Anything, "Enter a value for ':num'", iostreams.InputPromptConfig{
					Required: true,
				}).Return("3")
				clientsMock.IO.On("InputPrompt", mock.Anything, "Enter a value for ':zero'", iostreams.InputPromptConfig{
					Required: true,
				}).Return("0")
				clientsMock.IO.On("InputPrompt", mock.Anything, "Enter a value for ':progress'", iostreams.InputPromptConfig{
					Required: true,
				}).Return("wip")
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreQuery{
				Datastore:  "Todos",
				App:        mockAppID,
				Expression: "#task_id < :num AND #task_id <> :zero AND #notes = :progress",
				ExpressionAttributes: map[string]interface{}{
					"#task_id": "task_id",
					"#notes":   "status",
				},
				ExpressionValues: map[string]interface{}{
					":num":      "3",
					":zero":     "0",
					":progress": "wip",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := setupDatastoreMocks()
			if tt.Setup != nil {
				tt.Setup(clientsMock)
			}
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Prepare mocked command
			queryMock := new(QueryDatastorePkgMock)
			Query = queryMock.Query
			queryMock.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			cmd := NewQueryCommand(clients)
			// TODO: could maybe refactor this to the os/fs mocks level to more clearly communicate "fake being in an app directory"
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				clientsMock.Config.SetFlags(cmd)
				if tt.Prompts != nil {
					tt.Prompts(clientsMock)
				}
				return nil
			}

			clients.IO.SetCmdIO(cmd)

			// Perform test
			err := cmd.ExecuteContext(ctx)
			if assert.NoError(t, err) {
				queryMock.AssertCalled(t, "Query", mock.Anything, mock.Anything, mock.Anything, tt.Query)
			}

			// Cleanup when done
			if tt.Teardown != nil {
				tt.Teardown()
			}
		})
	}
}

func TestGetExpressionPatterns(t *testing.T) {
	tests := map[string]struct {
		Expression         string
		ExpectedAttributes []string
		ExpectedValues     []string
	}{
		"empty expression": {
			Expression: "",
		},
		"no attributes or values": {
			Expression: "example",
		},
		"single attribute and value": {
			Expression:         "#record = :num",
			ExpectedAttributes: []string{"#record"},
			ExpectedValues:     []string{":num"},
		},
		"multiple attributes and values": {
			Expression:         "#attr < :ceil AND #limit BETWEEN :low AND :high",
			ExpectedAttributes: []string{"#attr", "#limit"},
			ExpectedValues:     []string{":ceil", ":low", ":high"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actualAttributes, actualValues := getExpressionPatterns(tt.Expression)
			assert.Equal(t, tt.ExpectedAttributes, actualAttributes)
			assert.Equal(t, tt.ExpectedValues, actualValues)
		})
	}
}

func TestQueryCommandExport(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"export items to a file successfully": {
			CmdArgs: []string{
				`{"datastore":"Todos"}`,
				`--to-file=my-file`,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				*cm = *setupDatastoreMocks()
				_, err := prepareExportMockData(cm, 10, 2)
				assert.NoError(t, err)

				*cf = *shared.NewClientFactory(cm.MockClientFactory())
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				status, _ := exportProgressSpinner.Status()
				assert.Contains(t, status, "Successfully exported (10) items!")

				itemsFile, err := cm.Fs.Open("my-file")
				assert.NoError(t, err)

				items := []string{}
				scanner := bufio.NewScanner(itemsFile)
				for scanner.Scan() {
					items = append(items, scanner.Text())
				}
				assert.NoError(t, scanner.Err())
				assert.Equal(t, len(items), 10)
			},
		},
		"exports a max of 10000 items only": {
			CmdArgs: []string{
				`{"datastore":"Todos"}`,
				`--to-file=my-file`,
			},
			ExpectedOutputs: []string{"Export will be limited to the first 10000 items in the datastore"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				*cm = *setupDatastoreMocks()
				_, err := prepareExportMockData(cm, 10010, 100)
				assert.NoError(t, err)

				*cf = *shared.NewClientFactory(cm.MockClientFactory())
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				status, _ := exportProgressSpinner.Status()
				assert.Contains(t, status, "Successfully exported (10000) items!")
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewQueryCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func prepareExportMockData(cm *shared.ClientsMock, numberOfItems int, maxItemsToReturn int) ([]map[string]interface{}, error) {
	data := []map[string]interface{}{}
	for i := 1; i <= numberOfItems; i++ {
		item := map[string]interface{}{
			"task_id": fmt.Sprintf("%04d", i),
			"task":    "counting",
			"status":  "ongoing",
		}
		data = append(data, item)
		if len(data) == maxItemsToReturn || i == numberOfItems {
			nextCursor := ""
			if i < numberOfItems { //only set cursor if there are more items
				nextCursor = fmt.Sprintf("%d", i)

			}
			cm.API.On("AppsDatastoreQuery", mock.Anything, mock.Anything, mock.Anything).
				Return(types.AppDatastoreQueryResult{
					Items:      append([]map[string]interface{}{}, data...),
					NextCursor: nextCursor,
				}, nil).Once()
			data = []map[string]interface{}{}
		}
	}
	return data, nil
}
