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

package datastore

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type BulkPutPkgMock struct {
	mock.Mock
}

func (m *BulkPutPkgMock) BulkPut(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, query types.AppDatastoreBulkPut) (*logger.LogEvent, error) {
	m.Called(ctx, clients, log, query)
	log.Data["bulkPutResult"] = types.AppDatastoreBulkPutResult{}
	log.Log("info", "on_bulk_put_result")
	return log.SuccessEvent(), nil
}

func TestBulkPutCommandPreRun(t *testing.T) {
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
			expectedError:        slackerror.New(slackerror.ErrInvalidAppDirectory),
		},
		"errors if the manifest source is set to remote": {
			mockManifestSource:   config.ManifestSourceRemote,
			mockWorkingDirectory: "/slack/path/to/project",
			expectedError:        slackerror.New(slackerror.ErrAppNotHosted),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			manifestMock := &app.ManifestMockObject{}
			manifestMock.On(
				"GetManifestLocal",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(
				tc.mockManifestResponse,
				tc.mockManifestError,
			)
			clientsMock.AppClient.Manifest = manifestMock
			projectConfigMock := config.NewProjectConfigMock()
			projectConfigMock.On(
				"GetManifestSource",
				mock.Anything,
			).Return(
				tc.mockManifestSource,
				nil,
			)
			clientsMock.Config.ProjectConfig = projectConfigMock
			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(cf *shared.ClientFactory) {
				cf.Config.ForceFlag = tc.mockFlagForce
				cf.SDKConfig.WorkingDirectory = tc.mockWorkingDirectory
			})
			cmd := NewBulkPutCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBulkPutCommand(t *testing.T) {
	tests := map[string]struct {
		Setup    func(clientsMock *shared.ClientsMock)
		Teardown func()
		Query    types.AppDatastoreBulkPut
	}{
		"pass an expression through arguments": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, fmt.Sprintf(`{"datastore":"Todos","app":"%s","items":[{"task_id":"0002","task":"counting","status":"ongoing"}]}`, mockAppID))
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreBulkPut{
				Datastore: "Todos",
				App:       mockAppID,
				Items: []map[string]interface{}{
					{
						"task_id": "0002",
						"task":    "counting",
						"status":  "ongoing",
					},
				},
			},
		},
		"pass an expression through arguments and select the app": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, `{"datastore":"Todos","items":[{"task_id":"0101","task":"write code","status":"wip"}]}`)
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreBulkPut{
				Datastore: "Todos",
				App:       mockAppID,
				Items: []map[string]interface{}{
					{
						"task_id": "0101",
						"task":    "write code",
						"status":  "wip",
					},
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := setupDatastoreMocks()
			if tc.Setup != nil {
				tc.Setup(clientsMock)
			}
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			// Prepare mocked command
			bulkPutMock := new(BulkPutPkgMock)
			BulkPut = bulkPutMock.BulkPut
			bulkPutMock.On("BulkPut", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			cmd := NewBulkPutCommand(clients)
			// TODO: could maybe refactor this to the os/fs mocks level to more clearly communicate "fake being in an app directory"
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				clients.Config.SetFlags(cmd)
				return nil
			}
			clients.IO.SetCmdIO(cmd)

			// Perform test
			err := cmd.ExecuteContext(ctx)
			if assert.NoError(t, err) {
				bulkPutMock.AssertCalled(t, "BulkPut", mock.Anything, mock.Anything, mock.Anything, tc.Query)
			}

			// Cleanup when done
			if tc.Teardown != nil {
				tc.Teardown()
			}
		})
	}
}

func TestBulkPutCommandImport(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"import items from a file successfully": {
			CmdArgs: []string{
				`{"datastore":"Todos"}`,
				`--from-file=my-file`,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				*cm = *setupDatastoreMocks()
				cm.API.On("AppsDatastoreBulkPut", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreBulkPutResult{}, nil)

				itemsFile, err := cm.Fs.Create("my-file")
				assert.NoError(t, err)

				_, err = prepareImportMockData(itemsFile, 30, 0)
				assert.NoError(t, err)

				*cf = *shared.NewClientFactory(cm.MockClientFactory())
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				status, _ := importProgressSpinner.Status()
				assert.Contains(t, status, "Successfully imported (30) items! (0) items failed to be imported. Total processed items is (30)")

				output := cm.GetCombinedOutput()

				filePattern := regexp.MustCompile(`(/[^\s]+)`)
				filePaths := filePattern.FindAllString(output, -1)
				assert.Equal(t, len(filePaths), 1)
			},
		},
		"import items from a file with errors": {
			CmdArgs: []string{
				`{"datastore":"Todos"}`,
				`--from-file=my-file`,
			},
			ExpectedOutputs: []string{"Some items failed to be imported"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				*cm = *setupDatastoreMocks()
				cm.API.On("AppsDatastoreBulkPut", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreBulkPutResult{}, nil)

				itemsFile, err := cm.Fs.Create("my-file")
				assert.NoError(t, err)

				_, err = prepareImportMockData(itemsFile, 30, 5)
				assert.NoError(t, err)

				*cf = *shared.NewClientFactory(cm.MockClientFactory())
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				status, _ := importProgressSpinner.Status()
				assert.Contains(t, status, "Successfully imported (30) items! (5) items failed to be imported. Total processed items is (35)")

				output := cm.GetCombinedOutput()

				filePattern := regexp.MustCompile(`(/[^\s]+)`)
				filePaths := filePattern.FindAllString(output, -1)
				assert.Equal(t, len(filePaths), 2) //we log both items file and error log file

				errorsLogFilePath := filePaths[1]

				errorsFile, err := cm.Fs.Open(errorsLogFilePath)
				assert.NoError(t, err)

				errorLogs := []string{}
				scanner := bufio.NewScanner(errorsFile)
				for scanner.Scan() {
					errorLogs = append(errorLogs, scanner.Text())
				}
				assert.NoError(t, scanner.Err())
				assert.Equal(t, len(errorLogs), 5)
			},
		},
		"import the first 5000 items only": {
			CmdArgs: []string{
				`{"datastore":"Todos"}`,
				`--from-file=my-file`,
			},
			ExpectedOutputs: []string{"Import will be limited to the first 5000 items in the file"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				*cm = *setupDatastoreMocks()
				cm.API.On("AppsDatastoreBulkPut", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreBulkPutResult{}, nil)

				itemsFile, err := cm.Fs.Create("my-file")
				assert.NoError(t, err)

				_, err = prepareImportMockData(itemsFile, 5050, 0)
				assert.NoError(t, err)

				*cf = *shared.NewClientFactory(cm.MockClientFactory())
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				status, _ := importProgressSpinner.Status()
				assert.Contains(t, status, "Successfully imported (5000) items! (0) items failed to be imported. Total processed items is (5000)")
			},
		},
		"import retries on failures": {
			CmdArgs: []string{
				`{"datastore":"Todos"}`,
				`--from-file=my-file`,
			},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				*cm = *setupDatastoreMocks()

				itemsFile, err := cm.Fs.Create("my-file")
				assert.NoError(t, err)

				items, err := prepareImportMockData(itemsFile, 2, 0)
				assert.NoError(t, err)

				cm.API.On("AppsDatastoreBulkPut", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreBulkPutResult{FailedItems: items[:1]}, nil).Once()

				cm.API.On("AppsDatastoreBulkPut", mock.Anything, mock.Anything, mock.Anything).
					Return(types.AppDatastoreBulkPutResult{}, nil).Once()

				*cf = *shared.NewClientFactory(cm.MockClientFactory())
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				status, _ := importProgressSpinner.Status()
				assert.Contains(t, status, "Successfully imported (2) items! (0) items failed to be imported. Total processed items is (2)")

				cm.API.AssertNumberOfCalls(t, "AppsDatastoreBulkPut", 2)
				cm.API.AssertCalled(t, "AppsDatastoreBulkPut", mock.Anything, mock.Anything, types.AppDatastoreBulkPut{
					Datastore: "Todos",
					App:       "A0123456",
					Items: []map[string]interface{}{
						{"task_id": "0001", "task": "counting", "status": "ongoing"},
						{"task_id": "0002", "task": "counting", "status": "ongoing"},
					},
				})
				cm.API.AssertCalled(t, "AppsDatastoreBulkPut", mock.Anything, mock.Anything, types.AppDatastoreBulkPut{
					Datastore: "Todos",
					App:       "A0123456",
					Items: []map[string]interface{}{
						{"task_id": "0001", "task": "counting", "status": "ongoing"},
					},
				})
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewBulkPutCommand(cf)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func prepareImportMockData(file afero.File, numberOfValidRows int, numberOfInvalidRows int) ([]map[string]interface{}, error) {
	data := []map[string]interface{}{}
	for i := 1; i <= numberOfValidRows+numberOfInvalidRows; i++ {
		item := map[string]interface{}{
			"task_id": fmt.Sprintf("%04d", i),
			"task":    "counting",
			"status":  "ongoing",
		}
		if i > numberOfValidRows {
			delete(item, "task_id") //items with no primary key are invalid
		} else {
			data = append(data, item)
		}
		stringItem, err := goutils.JSONMarshalUnescaped(item)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		_, err = file.WriteString(stringItem)
		if err != nil {
			return []map[string]interface{}{}, err
		}
	}
	return data, nil
}
