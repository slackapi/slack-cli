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

package platform

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Setup a mock for the package
type ActivityPkgMock struct {
	mock.Mock
}

func (m *ActivityPkgMock) Activity(
	ctx context.Context,
	clients *shared.ClientFactory,
	log *logger.Logger,
	args types.ActivityArgs,
) error {
	m.Called()
	return nil
}

func TestActivity_Command(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewActivityCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	activityPkgMock := new(ActivityPkgMock)
	activityFunc = activityPkgMock.Activity
	activityPkgMock.On("Activity").Return(nil)

	appSelectMock := prompts.NewAppSelectMock()
	appSelectPromptFunc = appSelectMock.AppSelectPrompt
	appSelectMock.On("AppSelectPrompt", mock.Anything, mock.Anything, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly).Return(prompts.SelectedApp{}, nil)

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	activityPkgMock.AssertCalled(t, "Activity")
}

func TestActivity_Aliases(t *testing.T) {
	cmd := NewActivityCommand(&shared.ClientFactory{})
	require.Contains(t, cmd.Aliases, "log", "should be aliased as log")
	require.Contains(t, cmd.Aliases, "logs", "should be aliased as logs")
}
