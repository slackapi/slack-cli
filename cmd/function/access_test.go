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

package function

import (
	"context"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFunctionDistributionCommand(t *testing.T) {
	var appSelectTeardown func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"pass flags to set distribution to app collaborators": {
			// TODO: passing --name flag so that we don't need to mock chooseFunctionPrompt, which needs a proper mock
			CmdArgs: []string{"--app-collaborators", "--name", "F1234"},
			ExpectedOutputs: []string{
				"Function 'F1234' can be added to workflows by app collaborators",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				// set distribution
				clientsMock.ApiInterface.On("FunctionDistributionSet", mock.Anything, mock.Anything, mock.Anything, types.APP_COLLABORATORS, mock.Anything).
					Return([]types.FunctionDistributionUser{}, nil).Once()

				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []types.FunctionDistributionUser{}, nil).Once()

				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockAppSelection(installedProdApp)
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to grant distribution to specific users": {
			// TODO: passing --name flag so that we don't need to mock chooseFunctionPrompt, which needs a proper mock
			CmdArgs: []string{"--users", "U00,U01", "--grant", "--name", "F1234"},
			ExpectedOutputs: []string{
				"Function access granted to the provided users",
				"Function 'F1234' can be added to workflows by the following users",
				"U00",
				"U01",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				// check if distribution type is named_entities
				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.APP_COLLABORATORS, []types.FunctionDistributionUser{}, nil).Once()
				// set distribution type to named_entities
				clientsMock.ApiInterface.On("FunctionDistributionSet", mock.Anything, mock.Anything, mock.Anything, types.NAMED_ENTITIES, mock.Anything).
					Return([]types.FunctionDistributionUser{}, nil).Once()
				// add users
				clientsMock.ApiInterface.On("FunctionDistributionAddUsers", mock.Anything, mock.Anything, mock.Anything, "U00,U01").
					Return(nil).Once()
				// print access
				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []types.FunctionDistributionUser{{UserName: "user 0", ID: "U00"}, {UserName: "user 1", ID: "U01"}}, nil).Once()

				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockAppSelection(installedProdApp)
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to revoke distribution from specific users": {
			// TODO: passing --name flag so that we don't need to mock chooseFunctionPrompt, which needs a proper mock
			CmdArgs: []string{"--users", "U00", "--revoke", "--name", "F1234"},
			ExpectedOutputs: []string{
				"Function access revoked for the provided user",
				"Function 'F1234' can be added to workflows by the following users",
				"U01",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				// check if distribution type is named_entities
				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []types.FunctionDistributionUser{{UserName: "user 0", ID: "U00"}, {UserName: "user 1", ID: "U01"}}, nil).Once()
				// remove users
				clientsMock.ApiInterface.On("FunctionDistributionRemoveUsers", mock.Anything, mock.Anything, mock.Anything, "U00").
					Return(nil).Once()
				// print access
				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, mock.Anything, mock.Anything).
					Return(types.NAMED_ENTITIES, []types.FunctionDistributionUser{{UserName: "user 1", ID: "U01"}}, nil).Once()

				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockAppSelection(installedProdApp)
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"pass flags to retrieve only distribution information": {
			// TODO: passing --name flag so that we don't need to mock chooseFunctionPrompt, which needs a proper mock
			CmdArgs: []string{"--info", "--name", "F1234"},
			ExpectedOutputs: []string{
				"Function 'F1234' can be added to workflows by the following users",
				"everyone in the workspace",
			},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, "F1234", fakeApp.AppID).
					Return(types.EVERYONE, []types.FunctionDistributionUser{}, nil).Once()
				clientsMock.AddDefaultMocks()

				appSelectTeardown = setupMockAppSelection(installedProdApp)
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.ApiInterface.AssertNotCalled(t, "FunctionDistributionAddUsers")
				cm.ApiInterface.AssertNotCalled(t, "FunctionDistributionRemoveUsers")
				cm.ApiInterface.AssertNotCalled(t, "FunctionDistributionSet")
			},
			Teardown: func() {
				appSelectTeardown()
			},
		},
		"attempt to read permissions with the file flag": {
			CmdArgs: []string{"--file", "permissions.json"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.AddDefaultMocks()
				appSelectTeardown = setupMockAppSelection(installedProdApp)
				err := clients.AppClient().SaveDeployed(context.Background(), fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedErrorStrings: []string{"file does not exist"},
			Teardown: func() {
				appSelectTeardown()
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewDistributeCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}

func TestFunctionDistributionCommand_PermissionsFile(t *testing.T) {
	tests := map[string]struct {
		filename  string
		data      string
		functions map[string]struct {
			currentType      types.Permission
			currentEntities  []types.FunctionDistributionUser
			expectedType     types.Permission
			expectedEntities []types.FunctionDistributionUser
		}
		expectedError error
		expectedWarn  string
	}{
		"invalid file extension errors": {
			filename:      "permissions.jsn",
			data:          `{}`,
			expectedError: slackerror.New(slackerror.ErrUnknownFileType),
		},
		"invalid permission type errors": {
			filename: "permissions.json",
			data: `{
    "function_distribution": {
        "greeting_function": {
            "type": "every1"
        }
    }
}`,
			expectedError: slackerror.New(slackerror.ErrInvalidPermissionType),
		},
		"set permissions for multiple functions": {
			filename: "permissions.json",
			data: `{
    "function_distribution": {
        "greeting_function": {
            "type": "everyone"
        },
        "goodbye_function": {
            "type": "app_collaborators"
        },
        "momentary_function": {
            "type": "named_entities",
            "user_ids": ["USLACKBOT", "U123"]
        }
    }
}`,
			functions: map[string]struct {
				currentType      types.Permission
				currentEntities  []types.FunctionDistributionUser
				expectedType     types.Permission
				expectedEntities []types.FunctionDistributionUser
			}{
				"greeting_function": {
					currentType:      types.EVERYONE,
					currentEntities:  []types.FunctionDistributionUser{},
					expectedType:     types.EVERYONE,
					expectedEntities: []types.FunctionDistributionUser{},
				},
				"goodbye_function": {
					currentType:      types.EVERYONE,
					currentEntities:  []types.FunctionDistributionUser{},
					expectedType:     types.APP_COLLABORATORS,
					expectedEntities: []types.FunctionDistributionUser{},
				},
				"momentary_function": {
					currentType:     types.EVERYONE,
					currentEntities: []types.FunctionDistributionUser{},
					expectedType:    types.NAMED_ENTITIES,
					expectedEntities: []types.FunctionDistributionUser{
						{ID: "USLACKBOT"},
						{ID: "U123"},
					},
				},
			},
			expectedError: nil,
			expectedWarn:  "",
		},
		"warn when removing permissions from all entities": {
			filename: "permissions.json",
			data: `{
    "function_distribution": {
        "greeting_function": {
            "type": "named_entities",
            "user_ids": []
        }
    }
}`,
			functions: map[string]struct {
				currentType      types.Permission
				currentEntities  []types.FunctionDistributionUser
				expectedType     types.Permission
				expectedEntities []types.FunctionDistributionUser
			}{
				"greeting_function": {
					currentType:      types.EVERYONE,
					currentEntities:  []types.FunctionDistributionUser{},
					expectedType:     types.NAMED_ENTITIES,
					expectedEntities: []types.FunctionDistributionUser{},
				},
			},
			expectedError: nil,
			expectedWarn:  "No users will have access to 'greeting_function'",
		},
		"warn when named entities are ignored": {
			filename: "permissions.json",
			data: `{
    "function_distribution": {
        "goodbye_function": {
            "type": "app_collaborators",
            "user_ids": ["USLACKBOT"]
        }
    }
}`,
			functions: map[string]struct {
				currentType      types.Permission
				currentEntities  []types.FunctionDistributionUser
				expectedType     types.Permission
				expectedEntities []types.FunctionDistributionUser
			}{
				"goodbye_function": {
					currentType: types.NAMED_ENTITIES,
					currentEntities: []types.FunctionDistributionUser{
						{ID: "USLACKBOT"},
					},
					expectedType:     types.APP_COLLABORATORS,
					expectedEntities: []types.FunctionDistributionUser{},
				},
			},
			expectedError: nil,
			expectedWarn:  "The supplied user IDs to 'goodbye_function' are overridden by the 'app_collaborators' permission",
		},
		"correctly unmarshal yaml permissions": {
			filename: "function-perms.yaml",
			data: `function_distribution:
  greeting_function:
    type: everyone
  goodbye_function:
    type: app_collaborators`,
			functions: map[string]struct {
				currentType      types.Permission
				currentEntities  []types.FunctionDistributionUser
				expectedType     types.Permission
				expectedEntities []types.FunctionDistributionUser
			}{
				"greeting_function": {
					currentType:      types.APP_COLLABORATORS,
					currentEntities:  []types.FunctionDistributionUser{},
					expectedType:     types.EVERYONE,
					expectedEntities: []types.FunctionDistributionUser{},
				},
				"goodbye_function": {
					currentType: types.NAMED_ENTITIES,
					currentEntities: []types.FunctionDistributionUser{
						{ID: "USLACKBOT"},
					},
					expectedType:     types.APP_COLLABORATORS,
					expectedEntities: []types.FunctionDistributionUser{},
				},
			},
			expectedError: nil,
			expectedWarn:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.AddDefaultMocks()
			err := afero.WriteFile(clientsMock.Fs, tt.filename, []byte(tt.data), 0644)
			require.NoError(t, err)
			for function, permissions := range tt.functions {
				clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, function, mock.Anything).
					Return(permissions.currentType, permissions.currentEntities, nil)
			}
			clientsMock.ApiInterface.On("FunctionDistributionRemoveUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
			clientsMock.ApiInterface.On("FunctionDistributionSet", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return([]types.FunctionDistributionUser{}, nil)

			ctx := context.Background()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			app := types.App{AppID: "A123"}

			err = distributePermissionFile(ctx, clients, app, tt.filename)
			if err != nil || tt.expectedError != nil {
				assert.Equal(t,
					slackerror.ToSlackError(tt.expectedError).Code,
					slackerror.ToSlackError(err).Code)
			}
			if clientsMock.GetCombinedOutput() != "" || tt.expectedWarn != "" {
				assert.Contains(t, clientsMock.GetCombinedOutput(), tt.expectedWarn)
			}
			for function, permissions := range tt.functions {
				entityIDs := []string{}
				for _, entity := range permissions.expectedEntities {
					entityIDs = append(entityIDs, entity.ID)
				}
				entities := strings.Join(entityIDs, ",")
				clientsMock.ApiInterface.AssertCalled(t, "FunctionDistributionSet", mock.Anything, function, app.AppID, permissions.expectedType, entities)
			}
		})
	}
}

func TestFunctionDistributeCommand_UpdateNamedEntitiesDistribution(t *testing.T) {
	tests := map[string]struct {
		currentEntities []types.FunctionDistributionUser
		updatedEntities []string
		removedEntities string
	}{
		"no change to an empty distribution": {
			[]types.FunctionDistributionUser{},
			[]string{},
			"",
		},
		"entities are added to distribution": {
			[]types.FunctionDistributionUser{},
			[]string{"USLACKBOT", "U123"},
			"",
		},
		"entities are removed from a distribution": {
			[]types.FunctionDistributionUser{
				{ID: "USLACKBOT"},
				{ID: "U123"},
			},
			[]string{},
			"USLACKBOT,U123",
		},
		"only included entities are distributed": {
			[]types.FunctionDistributionUser{
				{ID: "USLACKBOT"},
				{ID: "U123"},
			},
			[]string{"U123", "U456"},
			"USLACKBOT",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.ApiInterface.On("FunctionDistributionSet", mock.Anything, mock.Anything, mock.Anything, types.NAMED_ENTITIES, mock.Anything).
				Return([]types.FunctionDistributionUser{}, nil).
				Run(func(args mock.Arguments) {
					clientsMock.ApiInterface.On("FunctionDistributionList", mock.Anything, mock.Anything, mock.Anything).
						Return(types.NAMED_ENTITIES, tt.currentEntities, nil).
						Run(func(args mock.Arguments) {
							clientsMock.ApiInterface.On("FunctionDistributionRemoveUsers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
								Return(nil)
						})
				})

			app := types.App{AppID: "A123"}
			function := "Ft123"
			ctx := context.Background()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			err := updateNamedEntitiesDistribution(ctx, clients, app, function, tt.updatedEntities)
			assert.NoError(t, err)
			clientsMock.ApiInterface.AssertCalled(t, "FunctionDistributionList", mock.Anything, function, app.AppID)
			entities := strings.Join(tt.updatedEntities, ",")
			clientsMock.ApiInterface.AssertCalled(t, "FunctionDistributionSet", mock.Anything, function, app.AppID, types.NAMED_ENTITIES, entities)
			clientsMock.ApiInterface.AssertCalled(t, "FunctionDistributionRemoveUsers", mock.Anything, function, app.AppID, tt.removedEntities)
		})
	}
}
