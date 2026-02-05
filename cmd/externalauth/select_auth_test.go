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

package externalauth

import (
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExternalAuthSelectAuthCommandPreRun(t *testing.T) {
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
			mockWorkingDirectory: "/slack/path/to/project",
			mockManifestSource:   config.ManifestSourceLocal,
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
			mockManifestSource:   config.ManifestSourceLocal,
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
			cmd := NewSelectAuthCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExternalAuthSelectAuthCommand(t *testing.T) {

	appSelectTeardown := setupMockAppSelection(installedProdApp)
	defer appSelectTeardown()

	sampleListReturnWithoutTokens := types.ExternalAuthorizationInfoLists{
		Authorizations: []types.ExternalAuthorizationInfo{
			{
				ProviderName:       "Google",
				ProviderKey:        "provider_a",
				ClientID:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   false,
			},
		}}
	sampleListReturnWithoutWorkflows := types.ExternalAuthorizationInfoLists{
		Authorizations: []types.ExternalAuthorizationInfo{
			{
				ProviderName:       "Provider_A",
				ProviderKey:        "provider_a",
				ClientID:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   true,
				ExternalTokenIDs:   []string{"Et0548LYDWCT"},
				ExternalTokens: []types.ExternalTokenInfo{
					{
						ExternalTokenID: "Et0548LABCD1",
						ExternalUserID:  "xyz@salesforce.com",
						DateUpdated:     1682021142,
					},
					{
						ExternalTokenID: "Et0548LABCDE2",
						ExternalUserID:  "xyz2@salesforce.com",
						DateUpdated:     1682021192,
					},
				},
			},
		}}
	sampleListReturnWithWorkflows := types.ExternalAuthorizationInfoLists{
		Authorizations: []types.ExternalAuthorizationInfo{
			{
				ProviderName:       "Provider_A",
				ProviderKey:        "provider_a",
				ClientID:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   true,
				ExternalTokenIDs:   []string{"Et0548LYDWCT"},
				ExternalTokens: []types.ExternalTokenInfo{
					{
						ExternalTokenID: "Et0548LABCD1",
						ExternalUserID:  "xyz@salesforce.com",
						DateUpdated:     1682021142,
					},
					{
						ExternalTokenID: "Et0548LABCDE2",
						ExternalUserID:  "xyz2@salesforce.com",
						DateUpdated:     1682021192,
					},
				},
			},
			{
				ProviderName:       "Provider_B",
				ProviderKey:        "provider_b",
				ClientID:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   false,
				ExternalTokenIDs:   []string{"Et0548LYDWCT"},
				ExternalTokens:     []types.ExternalTokenInfo{},
			},
		},
		Workflows: []types.WorkflowsInfo{
			{
				WorkflowID: "Wf0548LABCD1",
				CallbackID: "my_callback_id1",
				Providers: []types.ProvidersInfo{
					{
						ProviderKey:  "provider_c",
						ProviderName: "Provider_C",
					},
					{
						ProviderKey:  "provider_b",
						ProviderName: "Provider_B",
					},
					{
						ProviderKey:  "provider_a",
						ProviderName: "Provider_A",
						SelectedAuth: types.ExternalTokenInfo{
							ExternalTokenID: "Et0548LABCD1",
							ExternalUserID:  "xyz@salesforce.com",
							DateUpdated:     1682021142,
						},
					},
				},
			},
			{
				WorkflowID: "Wf0548LABCD2",
				CallbackID: "my_callback_id2",
				Providers: []types.ProvidersInfo{
					{
						ProviderKey:  "provider_a",
						ProviderName: "Provider_A",
					},
					{
						ProviderKey:  "provider_b",
						ProviderName: "Provider_B",
					},
				},
			},
		},
	}

	testutil.TableTestCommand(t, testutil.CommandTests{
		"list api returns error": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(types.ExternalAuthorizationInfoLists{}, errors.New("test error"))
			},
			ExpectedErrorStrings: []string{"test error"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with no tokens and no params": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithoutTokens, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{}, slackerror.New(slackerror.ErrMissingOptions))
			},
			ExpectedErrorStrings: []string{"No workflows found that require developer authorization"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything)
			},
		},
		"list with no tokens and workflow flag": {
			CmdArgs: []string{"--workflow", "#/workflows/workflow_callback"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithoutTokens, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/workflow_callback",
				}, nil)
			},
			ExpectedErrorStrings: []string{"Workflow not found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with no workflows and no params": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithoutWorkflows, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{}, slackerror.New(slackerror.ErrMissingOptions))
			},
			ExpectedErrorStrings: []string{"No workflows found that require developer authorization"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with no workflows and workflow flag": {
			CmdArgs: []string{"--workflow", "#/workflows/workflow_callback"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithoutWorkflows, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/workflow_callback",
				}, nil)
			},
			ExpectedErrorStrings: []string{"Workflow not found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and no param": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "workflow2",
					Index:  1,
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "provider_b",
					Index:  1,
				}, nil)
			},
			ExpectedErrorStrings: []string{"No connected accounts found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and invalid workflow param": {
			CmdArgs: []string{"--workflow", "#/workflows/workflow_callback"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "workflow2",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "provider_b",
					Index:  1,
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an external account", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("external-account"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "xyz@salesforce.com",
					Index:  0,
				}, nil)
			},
			ExpectedErrorStrings: []string{"Workflow not found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and valid workflow param": {
			CmdArgs: []string{"--workflow", "#/workflows/my_callback_id2"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.API.On("AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/my_callback_id2",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "provider_b",
				}, nil)
			},
			ExpectedErrorStrings: []string{"No connected accounts found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and valid workflow invalid provider": {
			CmdArgs: []string{"--workflow", "#/workflows/my_callback_id2", "--provider", "test_provider"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/my_callback_id2",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "test_provider",
				}, nil)
			},
			ExpectedErrorStrings: []string{"Provider is not used in the selected workflow"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and valid workflow valid provider": {
			CmdArgs: []string{"--workflow", "#/workflows/my_callback_id2", "--provider", "provider_b"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "#/workflows/my_callback_id2",
					Index:  1,
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "provider_b",
					Index:  1,
				}, nil)
			},
			ExpectedErrorStrings: []string{"No connected accounts found"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and valid workflow valid provider with tokens": {
			CmdArgs: []string{"--workflow", "#/workflows/my_callback_id2", "--provider", "provider_a"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.API.On("AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/my_callback_id2",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "provider_a",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an external account", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("external-account"),
				})).Return(iostreams.SelectPromptResponse{
					Prompt: true,
					Option: "xyz2@salesforce.com",
					Index:  1,
				}, nil)
			},
			ExpectedOutputs: []string{"Workflow #/workflows/my_callback_id2 will use developer account xyz2@salesforce.com when making calls to provider_a APIs"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and valid workflow valid provider invalid account": {
			CmdArgs: []string{"--workflow", "#/workflows/my_callback_id2", "--provider", "provider_a", "--external-account", "test_account"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.API.On("AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/my_callback_id2",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "provider_a",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an external account", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("external-account"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "test_account",
				}, nil)
			},
			ExpectedErrorStrings: []string{"Account is not used in the selected workflow"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		"list with workflows and valid workflow valid provider valid account": {
			CmdArgs: []string{"--workflow", "#/workflows/my_callback_id2", "--provider", "provider_a", "--external-account", "xyz2@salesforce.com"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sampleListReturnWithWorkflows, nil)
				clientsMock.API.On("AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("workflow"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "#/workflows/my_callback_id2",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("provider"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "provider_a",
				}, nil)
				clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an external account", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("external-account"),
				})).Return(iostreams.SelectPromptResponse{
					Flag:   true,
					Option: "xyz2@salesforce.com",
				}, nil)
			},
			ExpectedOutputs: []string{"Workflow #/workflows/my_callback_id2 will use developer account xyz2@salesforce.com when making calls to provider_a APIs"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "AppsAuthExternalSelectAuth", mock.Anything, mock.Anything, fakeAppID, mock.Anything, mock.Anything, mock.Anything)
			},
		}}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewSelectAuthCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}
