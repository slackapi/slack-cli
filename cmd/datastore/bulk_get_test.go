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
	"fmt"
	"os"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type BulkGetPkgMock struct {
	mock.Mock
}

func (m *BulkGetPkgMock) BulkGet(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, query types.AppDatastoreBulkGet) (*logger.LogEvent, error) {
	m.Called(ctx, clients, log, query)
	log.Data["bulkGetResult"] = types.AppDatastoreBulkGetResult{}
	log.Log("info", "on_bulk_get_result")
	return log.SuccessEvent(), nil
}

func TestBulkGetCommandPreRun(t *testing.T) {
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
			cmd := NewBulkGetCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tt.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tt.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBulkGetCommand(t *testing.T) {
	tests := map[string]struct {
		Setup    func(*shared.ClientsMock)
		Teardown func()
		Query    types.AppDatastoreBulkGet
	}{
		"pass an expression through arguments": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, fmt.Sprintf(`{"datastore":"Todos","app":"%s","ids":["0002"]}`, mockAppID))
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreBulkGet{
				Datastore: "Todos",
				App:       mockAppID,
				IDs:       []string{"0002"},
			},
		},
		"pass an expression through arguments and select the app": {
			Setup: func(*shared.ClientsMock) {
				os.Args = append(os.Args, `{"datastore":"Todos","ids":["0101"]}`)
			},
			Teardown: func() {
				os.Args = os.Args[:len(os.Args)-1]
			},
			Query: types.AppDatastoreBulkGet{
				Datastore: "Todos",
				App:       mockAppID,
				IDs:       []string{"0101"},
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
			bulkGetMock := new(BulkGetPkgMock)
			BulkGet = bulkGetMock.BulkGet
			bulkGetMock.On("BulkGet", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			cmd := NewBulkGetCommand(clients)
			// TODO: could maybe refactor this to the os/fs mocks level to more clearly communicate "fake being in an app directory"
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				clients.Config.SetFlags(cmd)
				return nil
			}
			clients.IO.SetCmdIO(cmd)

			// Perform test
			err := cmd.ExecuteContext(ctx)
			if assert.NoError(t, err) {
				bulkGetMock.AssertCalled(t, "BulkGet", mock.Anything, mock.Anything, mock.Anything, tt.Query)
			}

			// Cleanup when done
			if tt.Teardown != nil {
				tt.Teardown()
			}
		})
	}
}
