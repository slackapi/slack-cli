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
	"github.com/stretchr/testify/require"
)

func TestExternalAuthAddClientSecretCommandPreRun(t *testing.T) {
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
			cmd := NewAddClientSecretCommand(clients)
			err := cmd.PreRunE(cmd, nil)
			if tc.expectedError != nil {
				assert.Equal(t, slackerror.ToSlackError(tc.expectedError).Code, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExternalAuthAddClientSecretCommand(t *testing.T) {

	appSelectTeardown := setupMockAppSelection(installedProdApp)
	providerSelectTeardown := setupMockProviderSelection()

	defer providerSelectTeardown()
	defer appSelectTeardown()

	testutil.TableTestCommand(t, testutil.CommandTests{
		"no params": {
			CmdArgs: []string{},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(types.ExternalAuthorizationInfoLists{
						Authorizations: []types.ExternalAuthorizationInfo{
							{
								ProviderName:       "Google",
								ProviderKey:        "provider_a",
								ClientID:           "xxxxx",
								ClientSecretExists: true, ValidTokenExists: false,
							},
						}}, nil)

				clientsMock.API.On("AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter the client secret", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("secret"),
				})).Return(iostreams.PasswordPromptResponse{
					Prompt: true,
					Value:  "secret_key_1234",
				}, nil)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedOutputs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, fakeAppID, "provider_a", "secret_key_1234")
			},
		},
		"with --provider": {
			CmdArgs: []string{"--provider", "provider_a"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(types.ExternalAuthorizationInfoLists{
						Authorizations: []types.ExternalAuthorizationInfo{
							{
								ProviderName:       "Google",
								ProviderKey:        "provider_google",
								ClientID:           "xxxxx",
								ClientSecretExists: true, ValidTokenExists: false,
							},
						},
					}, nil)

				clientsMock.API.On("AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter the client secret", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("secret"),
				})).Return(iostreams.PasswordPromptResponse{
					Prompt: true,
					Value:  "secret_key_1234",
				}, nil)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedOutputs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, fakeAppID, "provider_a", "secret_key_1234")
			},
		},
		"with --secret": {
			CmdArgs: []string{"--secret", "secret"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(types.ExternalAuthorizationInfoLists{
						Authorizations: []types.ExternalAuthorizationInfo{
							{
								ProviderName:       "Google",
								ProviderKey:        "provider_a",
								ClientID:           "xxxxx",
								ClientSecretExists: true, ValidTokenExists: false,
							},
						}}, nil)
				clientsMock.API.On("AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter the client secret", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("secret"),
				})).Return(iostreams.PasswordPromptResponse{
					Flag:  true,
					Value: "secret",
				}, nil)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedOutputs: []string{},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, fakeAppID, "provider_a", "secret")
			},
		},
		"with --provider and --secret": {
			CmdArgs: []string{"--provider", "provider_a", "--secret", "secret"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything).
					Return(types.ExternalAuthorizationInfoLists{
						Authorizations: []types.ExternalAuthorizationInfo{
							{
								ProviderName:       "Google",
								ProviderKey:        "provider_a",
								ClientID:           "xxxxx",
								ClientSecretExists: true, ValidTokenExists: false,
							},
						}}, nil)
				clientsMock.API.On("AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter the client secret", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("secret"),
				})).Return(iostreams.PasswordPromptResponse{
					Flag:  true,
					Value: "secret",
				}, nil)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertCalled(t, "AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, fakeAppID, "provider_a", "secret")
			},
		},
		"when list api returns error": {
			CmdArgs: []string{"--provider", "provider_a", "--secret", "secret"},
			Setup: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock, clients *shared.ClientFactory) {
				clientsMock.API.On("AppsAuthExternalList",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(types.ExternalAuthorizationInfoLists{
						Authorizations: []types.ExternalAuthorizationInfo{}}, errors.New("test error"))
				clientsMock.API.On("AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("https://authorizationurl.com", nil)
				clientsMock.IO.On("PasswordPrompt", mock.Anything, "Enter the client secret", iostreams.MatchPromptConfig(iostreams.PasswordPromptConfig{
					Flag: clientsMock.Config.Flags.Lookup("secret"),
				})).Return(iostreams.PasswordPromptResponse{
					Flag:  true,
					Value: "secret",
				}, nil)
				// TODO: testing chicken and egg: we need the default mocks in place before we can use any of the `clients` methods
				clientsMock.AddDefaultMocks()
				// TODO this can probably be replaced by a helper that sets up an apps.json file in
				// the right place on the afero memfs instance
				err := clients.AppClient().SaveDeployed(ctx, fakeApp)
				require.NoError(t, err, "Cant write apps.json")
			},
			ExpectedErrorStrings: []string{"test error"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, clientsMock *shared.ClientsMock) {
				clientsMock.API.AssertNotCalled(t, "AppsAuthExternalClientSecretAdd", mock.Anything, mock.Anything, fakeAppID, "provider_a")
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewAddClientSecretCommand(clients)
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
		return cmd
	})
}
