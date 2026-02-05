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
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const mockAppID = "A0123456"

// setupDatastoreMocks creates a client mock with a datastore in the manifest
// and automatically selects a mock app when prompted
func setupDatastoreMocks() *shared.ClientsMock {
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clientsMock.Config.Flags.String("datastore", "", "mock datastore name flag")

	manifestMock := &app.ManifestMockObject{}
	manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			DisplayInformation: types.DisplayInformation{Name: "Datastorage"},
			Datastores: map[string]types.ManifestDatastore{
				"Todos": {
					PrimaryKey: "task_id",
					Attributes: map[string]types.ManifestAttribute{
						"task_id": {Type: "string"},
						"task":    {Type: "string"},
						"status":  {Type: "string"},
					},
				},
			},
		},
	}, nil)
	clientsMock.AppClient.Manifest = manifestMock

	// Select the same example app each time
	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{
		App: types.App{AppID: mockAppID},
	}, nil)

	return clientsMock
}

func TestDatastoreCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"shows the help page without commands or arguments or flags": {
			ExpectedStdoutOutputs: []string{
				"Interact with the items stored in an app's datastore.",
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCommand(clients)
		return cmd
	})
}

func TestSetQueryExpression(t *testing.T) {
	tests := map[string]struct {
		query           types.Datastorer
		expression      string
		method          string
		appFlag         string
		datastoreFlag   string
		expectedAppFlag string
		expectedQuery   types.Datastorer
		expectedError   error
	}{
		"valid queries form expressions": {
			query:           &types.AppDatastoreGet{},
			expression:      `{"datastore":"Todos","app":"A0001","id":"0002"}`,
			method:          "get",
			expectedAppFlag: "A0001",
			expectedQuery: &types.AppDatastoreGet{
				Datastore: "Todos",
				App:       "A0001",
				ID:        "0002",
			},
		},
		"different queries are supported": {
			query:      &types.AppDatastoreDelete{},
			expression: `{"datastore":"Reminders","id":"0004"}`,
			method:     "delete",
			expectedQuery: &types.AppDatastoreDelete{
				Datastore: "Reminders",
				ID:        "0004",
			},
		},
		"datastore name provided by flag": {
			query:         &types.AppDatastoreDelete{},
			expression:    `{"id":"0004"}`,
			method:        "delete",
			datastoreFlag: "Reminders",
			expectedQuery: &types.AppDatastoreDelete{
				Datastore: "Reminders",
				ID:        "0004",
			},
		},
		"invalid queries form errors": {
			query:         &types.AppDatastoreGet{},
			expression:    `{"datastore":"Todos"`,
			method:        "get",
			expectedError: slackerror.New(slackerror.ErrInvalidDatastoreExpression),
		},
		"mismatched app flag errors": {
			query:         &types.AppDatastoreGet{},
			expression:    `{"datastore":"Todos","app":"A0001"}`,
			method:        "get",
			appFlag:       "A123",
			expectedError: slackerror.New(slackerror.ErrInvalidAppFlag),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := setupDatastoreMocks()
			clientsMock.Config.AppFlag = tc.appFlag
			if tc.datastoreFlag != "" {
				err := clientsMock.Config.Flags.Lookup("datastore").Value.Set(tc.datastoreFlag)
				require.NoError(t, err)
				clientsMock.Config.Flags.Lookup("datastore").Changed = true
			}
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			err := setQueryExpression(clients, tc.query, tc.expression, tc.method)
			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError.(*slackerror.Error).Code, err.(*slackerror.Error).Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedQuery, tc.query)

			}
		})
	}
}
