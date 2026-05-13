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

package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	internalapi "github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_NewCommand(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewCommand(clients)

	assert.Equal(t, "api <method> [key=value ...] [flags]", cmd.Use)
	assert.Equal(t, "Call any Slack API method", cmd.Short)
}

func Test_runAPICommand_BodyFormats(t *testing.T) {
	tests := map[string]struct {
		flags          cmdFlags
		args           []string
		expectedMethod string
		expectedCT     string
		expectedAuth   string
		bodyContains   []string
		bodyEquals     string
	}{
		"form-encoded key=value params": {
			flags:          cmdFlags{method: "POST"},
			args:           []string{"chat.postMessage", "channel=C123", "text=hello"},
			expectedMethod: "POST",
			expectedCT:     "application/x-www-form-urlencoded",
			bodyContains:   []string{"channel=C123", "text=hello", "token=xoxb-test-token"},
		},
		"JSON auto-detect from arg": {
			flags:        cmdFlags{method: "POST"},
			args:         []string{"chat.postMessage", `{"channel":"C123","text":"hello"}`},
			expectedCT:   "application/json; charset=utf-8",
			expectedAuth: "Bearer xoxb-test-token",
			bodyEquals:   `{"channel":"C123","text":"hello"}`,
		},
		"JSON via --json flag": {
			flags:        cmdFlags{method: "POST", json: `{"channel":"C123"}`},
			args:         []string{"auth.test"},
			expectedCT:   "application/json; charset=utf-8",
			expectedAuth: "Bearer xoxb-test-token",
			bodyEquals:   `{"channel":"C123"}`,
		},
		"form-encoded via --data flag": {
			flags:        cmdFlags{method: "POST", data: "channel=C123&text=hello"},
			args:         []string{"chat.postMessage"},
			expectedCT:   "application/x-www-form-urlencoded",
			bodyContains: []string{"channel=C123", "text=hello", "token=xoxb-test-token"},
		},
		"GET method": {
			flags:          cmdFlags{method: "GET"},
			args:           []string{"auth.test"},
			expectedMethod: "GET",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var receivedMethod string
			var receivedContentType string
			var receivedAuth string
			var receivedBody string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				receivedContentType = r.Header.Get("Content-Type")
				receivedAuth = r.Header.Get("Authorization")
				body, _ := io.ReadAll(r.Body)
				receivedBody = string(body)
				fmt.Fprint(w, `{"ok":true}`)
			}))
			defer server.Close()

			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clientsMock.Config.TokenFlag = "xoxb-test-token"
			clientsMock.Config.APIHostResolved = server.URL
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			cmd := NewCommand(clients)
			testutil.MockCmdIO(clients.IO, cmd)

			flags = tc.flags
			cmd.SetArgs(tc.args)
			err := cmd.ExecuteContext(ctx)

			assert.NoError(t, err)
			if tc.expectedMethod != "" {
				assert.Equal(t, tc.expectedMethod, receivedMethod)
			}
			if tc.expectedCT != "" {
				assert.Equal(t, tc.expectedCT, receivedContentType)
			}
			if tc.expectedAuth != "" {
				assert.Equal(t, tc.expectedAuth, receivedAuth)
			}
			if tc.bodyEquals != "" {
				assert.Equal(t, tc.bodyEquals, receivedBody)
			}
			for _, s := range tc.bodyContains {
				assert.Contains(t, receivedBody, s)
			}
		})
	}
}

func Test_runAPICommand_IncludeHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "test-value")
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clientsMock.Config.TokenFlag = "xoxb-test-token"
	clientsMock.Config.APIHostResolved = server.URL
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	flags = cmdFlags{method: "POST", include: true}
	cmd.SetArgs([]string{"auth.test"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	output := clientsMock.GetStdoutOutput()
	assert.Contains(t, output, "HTTP 200")
	assert.Contains(t, output, "X-Custom: test-value")
	assert.Contains(t, output, `"ok":true`)
}

func Test_runAPICommand_IncludeHeadersSorted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Zebra", "last")
		w.Header().Set("X-Alpha", "first")
		w.Header().Set("X-Middle", "middle")
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clientsMock.Config.TokenFlag = "xoxb-test-token"
	clientsMock.Config.APIHostResolved = server.URL
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	flags = cmdFlags{method: "POST", include: true}
	cmd.SetArgs([]string{"auth.test"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	output := clientsMock.GetStdoutOutput()
	alphaIdx := strings.Index(output, "X-Alpha:")
	middleIdx := strings.Index(output, "X-Middle:")
	zebraIdx := strings.Index(output, "X-Zebra:")
	assert.Greater(t, alphaIdx, -1)
	assert.Greater(t, middleIdx, alphaIdx)
	assert.Greater(t, zebraIdx, middleIdx)
}

func Test_runAPICommand_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"ok":false,"error":"method_not_found"}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clientsMock.Config.TokenFlag = "xoxb-test-token"
	clientsMock.Config.APIHostResolved = server.URL
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	flags = cmdFlags{method: "POST"}
	cmd.SetArgs([]string{"nonexistent.method"})
	err := cmd.ExecuteContext(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
	output := clientsMock.GetStdoutOutput()
	assert.Contains(t, output, `"ok":false`)
}

func Test_runAPICommand_InvalidParam(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Config.TokenFlag = "xoxb-test-token"
	clientsMock.Config.APIHostResolved = "https://slack.com"
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	flags = cmdFlags{method: "POST"}
	cmd.SetArgs([]string{"auth.test", "not-a-key-value"})
	err := cmd.ExecuteContext(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key=value")
}

func Test_runAPICommand_CustomHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()
	clientsMock.Config.TokenFlag = "xoxb-test-token"
	clientsMock.Config.APIHostResolved = server.URL
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	flags = cmdFlags{method: "POST", headers: []string{"X-Custom: my-value"}}
	cmd.SetArgs([]string{"auth.test"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "my-value", receivedHeaders.Get("X-Custom"))
}

func Test_resolveToken_TokenFlag(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Config.TokenFlag = "xoxb-direct-token"
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-direct-token", token)
}

func Test_resolveToken_EnvBotToken(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-env-bot-token")

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-env-bot-token", token)
}

func Test_resolveToken_EnvUserToken(t *testing.T) {
	t.Setenv("SLACK_USER_TOKEN", "xoxp-env-user-token")

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxp-env-user-token", token)
}

func Test_resolveToken_EnvOverridesAppPrompt(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-env-bot-token")

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig.WorkingDirectory = "/fake/project"

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-env-bot-token", token)
	clientsMock.IO.AssertNotCalled(t, "SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_resolveToken_AppOverridesEnv(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-env-bot-token")

	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Os.AddDefaultMocks()
	clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	clientsMock.Config.AppFlag = "A111"

	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
		{Token: "xoxp-tooling", TeamID: "T111", TeamDomain: "team-a"},
	}, nil)
	clientsMock.Auth.On("SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appClientMock := &app.AppClientMock{}
	appClientMock.On("GetDeployedAll", mock.Anything).Return([]types.App{
		{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"},
	}, "", nil)
	appClientMock.On("GetLocalAll", mock.Anything).Return([]types.App{}, nil)
	appClientMock.On("GetDeployed", mock.Anything, "T111").Return(types.App{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"}, nil)
	appClientMock.On("GetLocal", mock.Anything, "T111").Return(types.App{}, nil)
	clientsMock.AppClient.AppClientInterface = appClientMock

	clientsMock.API.On("GetAppStatus", mock.Anything, "xoxp-tooling", []string{"A111"}, "T111").
		Return(internalapi.GetAppStatusResult{Apps: []internalapi.AppStatusResultAppInfo{{AppID: "A111", Installed: true}}}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(internalapi.AuthSession{}, nil)

	manifestMock := clientsMock.AppClient.Manifest.(*app.ManifestMockObject)
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			OAuthConfig: &types.OAuthConfig{Scopes: &types.ManifestScopes{Bot: []string{"chat:write"}}},
		},
	}, nil)

	clientsMock.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, "xoxp-tooling", mock.Anything, []string{"chat:write"}, []string{}, "", false).
		Return(internalapi.DeveloperAppInstallResult{APIAccessTokens: struct {
			Bot      string `json:"bot,omitempty"`
			AppLevel string `json:"app_level,omitempty"`
			User     string `json:"user,omitempty"`
		}{Bot: "xoxb-app-bot-token"}}, types.InstallSuccess, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig.WorkingDirectory = "/fake/project"

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-app-bot-token", token)
}

func Test_resolveToken_AppFlag_ByID(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Os.AddDefaultMocks()
	clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	clientsMock.Config.AppFlag = "A111"

	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
		{Token: "xoxp-tooling", TeamID: "T111", TeamDomain: "team-a"},
	}, nil)
	clientsMock.Auth.On("SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appClientMock := &app.AppClientMock{}
	appClientMock.On("GetDeployedAll", mock.Anything).Return([]types.App{
		{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"},
	}, "", nil)
	appClientMock.On("GetLocalAll", mock.Anything).Return([]types.App{}, nil)
	appClientMock.On("GetDeployed", mock.Anything, "T111").Return(types.App{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"}, nil)
	appClientMock.On("GetLocal", mock.Anything, "T111").Return(types.App{}, nil)
	clientsMock.AppClient.AppClientInterface = appClientMock

	clientsMock.API.On("GetAppStatus", mock.Anything, "xoxp-tooling", []string{"A111"}, "T111").
		Return(internalapi.GetAppStatusResult{Apps: []internalapi.AppStatusResultAppInfo{{AppID: "A111", Installed: true}}}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, "xoxp-tooling").Return(internalapi.AuthSession{}, nil)

	manifestMock := clientsMock.AppClient.Manifest.(*app.ManifestMockObject)
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			OAuthConfig: &types.OAuthConfig{Scopes: &types.ManifestScopes{Bot: []string{"chat:write"}}},
		},
	}, nil)

	clientsMock.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, "xoxp-tooling", mock.Anything, []string{"chat:write"}, []string{}, "", false).
		Return(internalapi.DeveloperAppInstallResult{APIAccessTokens: struct {
			Bot      string `json:"bot,omitempty"`
			AppLevel string `json:"app_level,omitempty"`
			User     string `json:"user,omitempty"`
		}{Bot: "xoxb-app-bot-token"}}, types.InstallSuccess, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig.WorkingDirectory = "/fake/project"

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-app-bot-token", token)
}

func Test_resolveToken_AppFlag_Local(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Os.AddDefaultMocks()
	clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	clientsMock.Config.AppFlag = "local"

	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
		{Token: "xoxp-tooling", TeamID: "T111", TeamDomain: "team-a"},
	}, nil)
	clientsMock.Auth.On("SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appClientMock := &app.AppClientMock{}
	appClientMock.On("GetDeployedAll", mock.Anything).Return([]types.App{}, "", nil)
	appClientMock.On("GetLocalAll", mock.Anything).Return([]types.App{
		{AppID: "A111", TeamID: "T111", TeamDomain: "team-a", IsDev: true},
	}, nil)
	appClientMock.On("GetDeployed", mock.Anything, "T111").Return(types.App{}, nil)
	appClientMock.On("GetLocal", mock.Anything, "T111").Return(types.App{AppID: "A111", TeamID: "T111", TeamDomain: "team-a", IsDev: true}, nil)
	clientsMock.AppClient.AppClientInterface = appClientMock

	clientsMock.API.On("GetAppStatus", mock.Anything, "xoxp-tooling", []string{"A111"}, "T111").
		Return(internalapi.GetAppStatusResult{Apps: []internalapi.AppStatusResultAppInfo{{AppID: "A111", Installed: true}}}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, mock.Anything).Return(internalapi.AuthSession{}, nil)
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an app", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{Index: 0, Prompt: true}, nil)

	manifestMock := clientsMock.AppClient.Manifest.(*app.ManifestMockObject)
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			OAuthConfig: &types.OAuthConfig{Scopes: &types.ManifestScopes{Bot: []string{"commands"}}},
		},
	}, nil)

	clientsMock.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, "xoxp-tooling", mock.Anything, []string{"commands"}, []string{}, "", false).
		Return(internalapi.DeveloperAppInstallResult{APIAccessTokens: struct {
			Bot      string `json:"bot,omitempty"`
			AppLevel string `json:"app_level,omitempty"`
			User     string `json:"user,omitempty"`
		}{Bot: "xoxb-local-bot-token"}}, types.InstallSuccess, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig.WorkingDirectory = "/fake/project"

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-local-bot-token", token)
}

func Test_resolveToken_AppFlag_NotFound(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Os.AddDefaultMocks()
	clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))
	clientsMock.Config.AppFlag = "A999"

	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
		{Token: "xoxp-tooling", TeamID: "T111", TeamDomain: "team-a"},
	}, nil)

	appClientMock := &app.AppClientMock{}
	appClientMock.On("GetDeployedAll", mock.Anything).Return([]types.App{
		{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"},
	}, "", nil)
	appClientMock.On("GetLocalAll", mock.Anything).Return([]types.App{}, nil)
	appClientMock.On("GetDeployed", mock.Anything, "T111").Return(types.App{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"}, nil)
	appClientMock.On("GetLocal", mock.Anything, "T111").Return(types.App{}, nil)
	clientsMock.AppClient.AppClientInterface = appClientMock

	clientsMock.API.On("GetAppStatus", mock.Anything, "xoxp-tooling", []string{"A111"}, "T111").
		Return(internalapi.GetAppStatusResult{Apps: []internalapi.AppStatusResultAppInfo{{AppID: "A111", Installed: true}}}, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig.WorkingDirectory = "/fake/project"

	_, err := resolveToken(ctx, clients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app_not_found")
}

func Test_resolveToken_AppSelection(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.Os.AddDefaultMocks()
	clientsMock.IO.On("PrintDebug", mock.Anything, mock.Anything, mock.MatchedBy(func(args ...any) bool { return true }))

	clientsMock.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
		{Token: "xoxp-tooling", TeamID: "T111", TeamDomain: "team-a"},
	}, nil)
	clientsMock.Auth.On("SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appClientMock := &app.AppClientMock{}
	appClientMock.On("GetDeployedAll", mock.Anything).Return([]types.App{
		{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"},
	}, "", nil)
	appClientMock.On("GetLocalAll", mock.Anything).Return([]types.App{}, nil)
	appClientMock.On("GetDeployed", mock.Anything, "T111").Return(types.App{AppID: "A111", TeamID: "T111", TeamDomain: "team-a"}, nil)
	appClientMock.On("GetLocal", mock.Anything, "T111").Return(types.App{}, nil)
	clientsMock.AppClient.AppClientInterface = appClientMock

	clientsMock.API.On("GetAppStatus", mock.Anything, "xoxp-tooling", []string{"A111"}, "T111").
		Return(internalapi.GetAppStatusResult{Apps: []internalapi.AppStatusResultAppInfo{{AppID: "A111", Installed: true}}}, nil)
	clientsMock.API.On("ValidateSession", mock.Anything, "xoxp-tooling").Return(internalapi.AuthSession{}, nil)

	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an app", mock.Anything, mock.Anything).
		Return(iostreams.SelectPromptResponse{Index: 0, Prompt: true}, nil)

	manifestMock := clientsMock.AppClient.Manifest.(*app.ManifestMockObject)
	manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).Return(types.SlackYaml{
		AppManifest: types.AppManifest{
			OAuthConfig: &types.OAuthConfig{Scopes: &types.ManifestScopes{Bot: []string{"chat:write"}}},
		},
	}, nil)

	clientsMock.API.On("DeveloperAppInstall", mock.Anything, mock.Anything, "xoxp-tooling", mock.Anything, []string{"chat:write"}, []string{}, "", false).
		Return(internalapi.DeveloperAppInstallResult{APIAccessTokens: struct {
			Bot      string `json:"bot,omitempty"`
			AppLevel string `json:"app_level,omitempty"`
			User     string `json:"user,omitempty"`
		}{Bot: "xoxb-app-bot-token"}}, types.InstallSuccess, nil)

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig.WorkingDirectory = "/fake/project"

	token, err := resolveToken(ctx, clients)
	assert.NoError(t, err)
	assert.Equal(t, "xoxb-app-bot-token", token)
}

func Test_resolveToken_NoTokenFound(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	_, err := resolveToken(ctx, clients)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token found")
}
