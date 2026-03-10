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

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSandboxAuth_withTokenFlag(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	expectedAuth := types.SlackAuth{
		Token:      "xoxb-test-token",
		TeamID:     "T123",
		TeamDomain: "my-team",
		UserID:     "U456",
	}
	clientsMock.Config.TokenFlag = "xoxb-test-token"
	clientsMock.Auth.On("AuthWithToken", mock.Anything, "xoxb-test-token").Return(expectedAuth, nil)
	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, "", mock.Anything).Return("https://api.slack.com")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, "https://api.slack.com", "").Return("https://slackb.com/events/cli")

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.CLIVersion = ""

	token, auth, err := getSandboxAuth(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, "xoxb-test-token", token)
	assert.Equal(t, &expectedAuth, auth)
	assert.Equal(t, "https://api.slack.com", clients.Config.APIHostResolved)
	assert.Equal(t, "https://slackb.com/events/cli", clients.Config.LogstashHostResolved)

	clientsMock.Auth.AssertCalled(t, "AuthWithToken", mock.Anything, "xoxb-test-token")
	clientsMock.Auth.AssertCalled(t, "ResolveAPIHost", mock.Anything, "", mock.Anything)
	clientsMock.Auth.AssertCalled(t, "ResolveLogstashHost", mock.Anything, "https://api.slack.com", "")
}

func TestGetSandboxAuth_withTokenFlagAuthError(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authErr := errors.New("invalid token")
	clientsMock.Config.TokenFlag = "xoxb-bad-token"
	clientsMock.Auth.On("AuthWithToken", mock.Anything, "xoxb-bad-token").Return(types.SlackAuth{}, authErr)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, _, err := getSandboxAuth(ctx, clients)
	require.Error(t, err)
	assert.ErrorIs(t, err, authErr)

	clientsMock.Auth.AssertNotCalled(t, "ResolveAPIHost")
	clientsMock.Auth.AssertNotCalled(t, "ResolveLogstashHost")
}

func TestGetSandboxAuth_withoutTokenCallsResolveAuth(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	expectedAuth := types.SlackAuth{
		Token:      "xoxb-resolved",
		TeamID:     "T789",
		TeamDomain: "resolved-team",
		UserID:     "U999",
	}
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{expectedAuth}, nil)
	clientsMock.Auth.On("ResolveAPIHost", mock.Anything, "", mock.Anything).Return("https://api.slack.com")
	clientsMock.Auth.On("ResolveLogstashHost", mock.Anything, "https://api.slack.com", "").Return("https://slackb.com/events/cli")

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.CLIVersion = ""

	token, auth, err := getSandboxAuth(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, "xoxb-resolved", token)
	assert.Equal(t, &expectedAuth, auth)

	clientsMock.Auth.AssertCalled(t, "Auths", mock.Anything)
	clientsMock.Auth.AssertNotCalled(t, "AuthWithToken")
}

func TestResolveAuthForSandbox_tokenFlag(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	expectedAuth := types.SlackAuth{
		Token:      "xoxb-from-flag",
		TeamID:     "T111",
		TeamDomain: "flag-team",
		UserID:     "U111",
	}
	clientsMock.Config.TokenFlag = "xoxb-from-flag"
	clientsMock.Auth.On("AuthWithToken", mock.Anything, "xoxb-from-flag").Return(expectedAuth, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	auth, err := resolveAuthForSandbox(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, &expectedAuth, auth)

	clientsMock.Auth.AssertNotCalled(t, "Auths")
}

func TestResolveAuthForSandbox_tokenFlagError(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authErr := slackerror.New(slackerror.ErrInvalidAuth)
	clientsMock.Config.TokenFlag = "xoxb-bad"
	clientsMock.Auth.On("AuthWithToken", mock.Anything, "xoxb-bad").Return(types.SlackAuth{}, authErr)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	assert.True(t, slackerror.ToSlackError(err).Code == slackerror.ErrInvalidAuth)
}

func TestResolveAuthForSandbox_noAuths(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	clientsMock.Config.TokenFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	se := slackerror.ToSlackError(err)
	assert.Equal(t, slackerror.ErrCredentialsNotFound, se.Code)
	assert.Contains(t, err.Error(), "logged in")
	assert.Contains(t, err.Error(), "slack login")
}

func TestResolveAuthForSandbox_authsError(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authsErr := errors.New("credentials read failed")
	clientsMock.Config.TokenFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth(nil), authsErr)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	// slackerror wraps the root cause
	assert.Contains(t, err.Error(), "credentials read failed")
}

func TestResolveAuthForSandbox_teamFlagMatchByTeamID(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authA := types.SlackAuth{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"}
	authB := types.SlackAuth{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = "T2"
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{authA, authB}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	auth, err := resolveAuthForSandbox(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, "tok-b", auth.Token)
	assert.Equal(t, "T2", auth.TeamID)
	assert.Equal(t, "beta", auth.TeamDomain)
}

func TestResolveAuthForSandbox_teamFlagMatchByTeamDomain(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authA := types.SlackAuth{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"}
	authB := types.SlackAuth{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = "beta"
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{authA, authB}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	auth, err := resolveAuthForSandbox(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, "tok-b", auth.Token)
	assert.Equal(t, "beta", auth.TeamDomain)
}

func TestResolveAuthForSandbox_teamFlagNoMatch(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	auths := []types.SlackAuth{
		{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"},
		{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"},
	}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = "nonexistent"
	clientsMock.Auth.On("Auths", mock.Anything).Return(auths, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	se := slackerror.ToSlackError(err)
	assert.Equal(t, slackerror.ErrTeamNotFound, se.Code)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestResolveAuthForSandbox_singleAuth(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	singleAuth := types.SlackAuth{Token: "tok-single", TeamID: "T99", TeamDomain: "solo", UserID: "U99"}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{singleAuth}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	auth, err := resolveAuthForSandbox(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, &singleAuth, auth)

	clientsMock.IO.AssertNotCalled(t, "SelectPrompt")
}

func TestResolveAuthForSandbox_multipleAuthsPromptSelection(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authA := types.SlackAuth{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"}
	authB := types.SlackAuth{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{authA, authB}, nil)

	// Options are sorted by TeamDomain then TeamID, so order is alpha(T1), beta(T2)
	// User selects index 1 (beta)
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a team for authentication", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{Prompt: true, Index: 1, Option: "beta T2"}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	auth, err := resolveAuthForSandbox(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, "tok-b", auth.Token)
	assert.Equal(t, "T2", auth.TeamID)
	assert.Equal(t, "beta", auth.TeamDomain)

	clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a team for authentication", mock.Anything, mock.Anything)
}

func TestResolveAuthForSandbox_multipleAuthsFlagSelectionMatch(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authA := types.SlackAuth{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"}
	authB := types.SlackAuth{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{authA, authB}, nil)

	// Simulate --team T2 passed via flag; SelectPrompt returns Flag with Option "T2"
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a team for authentication", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{Flag: true, Option: "T2"}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	auth, err := resolveAuthForSandbox(ctx, clients)
	require.NoError(t, err)
	assert.Equal(t, "tok-b", auth.Token)
	assert.Equal(t, "T2", auth.TeamID)
}

func TestResolveAuthForSandbox_multipleAuthsFlagSelectionNoMatch(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	authA := types.SlackAuth{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"}
	authB := types.SlackAuth{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{authA, authB}, nil)

	// Simulate --team invalid passed via flag
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a team for authentication", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{Flag: true, Option: "invalid"}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	se := slackerror.ToSlackError(err)
	assert.Equal(t, slackerror.ErrTeamNotFound, se.Code)
	assert.Contains(t, err.Error(), "invalid")
}

func TestResolveAuthForSandbox_selectPromptError(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	auths := []types.SlackAuth{
		{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"},
		{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"},
	}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return(auths, nil)

	promptErr := errors.New("prompt interrupted")
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a team for authentication", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{}, promptErr)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompt interrupted")
}

func TestResolveAuthForSandbox_invalidSelection(t *testing.T) {
	ctx := slackcontext.MockContext(context.Background())
	clientsMock := shared.NewClientsMock()
	clientsMock.IO.AddDefaultMocks()

	auths := []types.SlackAuth{
		{Token: "tok-a", TeamID: "T1", TeamDomain: "alpha", UserID: "U1"},
		{Token: "tok-b", TeamID: "T2", TeamDomain: "beta", UserID: "U2"},
	}
	clientsMock.Config.TokenFlag = ""
	clientsMock.Config.TeamFlag = ""
	clientsMock.Auth.On("Auths", mock.Anything).Return(auths, nil)

	// Neither Flag nor Prompt set - invalid state
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a team for authentication", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveAuthForSandbox(ctx, clients)
	require.Error(t, err)
	se := slackerror.ToSlackError(err)
	assert.Equal(t, slackerror.ErrInvalidAuth, se.Code)
}
