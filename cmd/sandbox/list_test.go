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

package sandbox

import (
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	// Enable sandboxes experiment
	clientsMock.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
	clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewListCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	// Use global --token (Config.TokenFlag) to bypass stored credentials; mock AuthWithToken
	testToken := "xoxb-test-token"
	clientsMock.Config.TokenFlag = testToken
	clientsMock.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")

	// Mock ListSandboxes to return empty list
	clientsMock.API.On("ListSandboxes", mock.Anything, testToken, "").Return([]types.Sandbox{}, nil)

	cmd.SetArgs([]string{})
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	clientsMock.Auth.AssertCalled(t, "AuthWithToken", mock.Anything, testToken)
	clientsMock.API.AssertCalled(t, "ListSandboxes", mock.Anything, testToken, "")
	assert.Contains(t, clientsMock.GetStdoutOutput(), "No sandboxes found")
}

func TestListCommand_withSandboxes(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	clientsMock.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
	clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewListCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	testToken := "xoxb-test-token"
	clientsMock.Config.TokenFlag = testToken
	clientsMock.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")

	sandboxes := []types.Sandbox{
		{
			SandboxTeamID: "T123",
			SandboxName:   "my-sandbox",
			SandboxDomain: "my-sandbox",
			Status:        "active",
			DateCreated:   1700000000,
			DateArchived:  0,
		},
	}
	clientsMock.API.On("ListSandboxes", mock.Anything, testToken, "").Return(sandboxes, nil)

	cmd.SetArgs([]string{})
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	clientsMock.API.AssertCalled(t, "ListSandboxes", mock.Anything, testToken, "")
	assert.Contains(t, clientsMock.GetStdoutOutput(), "my-sandbox")
	assert.Contains(t, clientsMock.GetStdoutOutput(), "T123")
	assert.Contains(t, clientsMock.GetStdoutOutput(), "https://my-sandbox.slack.com")
}

func TestListCommand_withFilter(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	clientsMock.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
	clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewListCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	testToken := "xoxb-test-token"
	clientsMock.Config.TokenFlag = testToken
	clientsMock.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
	clientsMock.API.On("ListSandboxes", mock.Anything, testToken, "active").Return([]types.Sandbox{}, nil)

	cmd.SetArgs([]string{"--filter", "active"})
	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	clientsMock.API.AssertCalled(t, "ListSandboxes", mock.Anything, testToken, "active")
}

func TestListCommand_listError(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	clientsMock.Config.ExperimentsFlag = []string{string(experiment.Sandboxes)}
	clientsMock.Config.LoadExperiments(ctx, clientsMock.IO.PrintDebug)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewListCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	testToken := "xoxb-test-token"
	clientsMock.Config.TokenFlag = testToken
	clientsMock.Auth.On("AuthWithToken", mock.Anything, testToken).Return(types.SlackAuth{Token: testToken}, nil)
	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, mock.Anything, mock.Anything).Return("https://api.slack.com")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, mock.Anything, mock.Anything).Return("https://slackb.com/events/cli")
	clientsMock.API.On("ListSandboxes", mock.Anything, testToken, "").
		Return([]types.Sandbox(nil), errors.New("api_error"))

	cmd.SetArgs([]string{})
	err := cmd.ExecuteContext(ctx)
	require.Error(t, err)
}

func TestListCommand_experimentRequired(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	// Do NOT enable sandboxes experiment
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewListCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	clientsMock.Config.TokenFlag = "xoxb-test"
	cmd.SetArgs([]string{})
	err := cmd.ExecuteContext(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sandbox")
	clientsMock.API.AssertNotCalled(t, "ListSandboxes", mock.Anything, mock.Anything, mock.Anything)
}
