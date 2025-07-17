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

package manifest

import (
	"context"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type ManifestValidatePkgMock struct {
	mock.Mock
}

func (m *ManifestValidatePkgMock) ManifestValidate(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, app types.App, token string) (*logger.LogEvent, slackerror.Warnings, error) {
	m.Called(ctx, clients, log, app, token)
	return log.SuccessEvent(), nil, nil
}

func TestManifestValidateCommand(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewValidateCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, nil)

	manifestValidatePkgMock := new(ManifestValidatePkgMock)
	manifestValidateFunc = manifestValidatePkgMock.ManifestValidate

	manifestValidatePkgMock.On("ManifestValidate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	manifestValidatePkgMock.AssertCalled(t, "ManifestValidate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestManifestValidateCommand_HandleMissingAppInstallError_ZeroUserAuth(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewValidateCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	// Mock a failed AppSelectPrompt
	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired))

	// Mock zero user auths
	mockAuths := []types.SlackAuth{}
	clientsMock.Auth.On("Auths", mock.Anything).Return(mockAuths, nil)
	clientsMock.AddDefaultMocks()

	// A failed selection/prompt should raise an error
	err := cmd.ExecuteContext(ctx)
	require.ErrorContains(t, err, slackerror.ErrNotAuthed)
}

func TestManifestValidateCommand_HandleMissingAppInstallError_OneUserAuth(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Mock one user auths
	mockAuths := []types.SlackAuth{
		{
			Token:        "mocktokenval",
			TeamDomain:   "mock",
			TeamID:       "mockteamID",
			UserID:       "mockuser",
			LastUpdated:  time.Time{},
			RefreshToken: "refresh",
			ExpiresAt:    0,
		},
	}
	clientsMock.Auth.On("Auths", mock.Anything).Return(mockAuths, nil)
	clientsMock.AddDefaultMocks()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewValidateCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	// Mock a failed AppSelectPrompt
	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired))

	// Mock the manifest validate package
	manifestValidatePkgMock := new(ManifestValidatePkgMock)
	manifestValidateFunc = manifestValidatePkgMock.ManifestValidate
	manifestValidatePkgMock.On("ManifestValidate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Should execute without error
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)
	clientsMock.Auth.AssertCalled(t, "SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestManifestValidateCommand_HandleMissingAppInstallError_MoreThanOneUserAuth(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewValidateCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	// Mock a failed AppSelectPrompt
	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, slackerror.New(slackerror.ErrInstallationRequired))
	clientsMock.IO.On("SelectPrompt", mock.Anything, prompts.SelectTeamPrompt, mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clients.Config.Flags.Lookup("team"),
	})).Return(iostreams.SelectPromptResponse{
		Prompt: true,
		Option: "mock2",
		Index:  1,
	}, nil)

	// Mock multiple user auths
	mockAuths := []types.SlackAuth{
		{
			Token:        "mocktokenval",
			TeamDomain:   "mock",
			TeamID:       "mockteamID",
			UserID:       "mockuser",
			LastUpdated:  time.Time{},
			RefreshToken: "refresh",
			ExpiresAt:    0,
		},
		{
			Token:        "mocktokenval",
			TeamDomain:   "mock2",
			TeamID:       "mockteamID",
			UserID:       "mockuser",
			LastUpdated:  time.Time{},
			RefreshToken: "refresh",
			ExpiresAt:    0,
		},
	}
	clientsMock.Auth.On("Auths", mock.Anything).Return(mockAuths, nil)
	clientsMock.Auth.On("AuthWithTeamDomain", mock.Anything, "mock2").Return(mockAuths[1], nil)
	clientsMock.AddDefaultMocks()

	// Mock the manifest validate package
	manifestValidatePkgMock := new(ManifestValidatePkgMock)
	manifestValidateFunc = manifestValidatePkgMock.ManifestValidate
	manifestValidatePkgMock.On("ManifestValidate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Should execute without error
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)
	clientsMock.Auth.AssertCalled(t, "SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestManifestValidateCommand_HandleOtherErrors(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewValidateCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	// Mock a failed AppSelectPrompt with a different error
	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	errMsg := "Unrelated error"
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, slackerror.New(errMsg))

	err := cmd.ExecuteContext(ctx)
	require.ErrorContains(t, err, errMsg)
}
