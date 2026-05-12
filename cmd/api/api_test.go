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
	"encoding/json"
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

func Test_runAPICommand_FormEncoded(t *testing.T) {
	var receivedContentType string
	var receivedBody string
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
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

	flags = cmdFlags{method: "POST"}
	cmd.SetArgs([]string{"chat.postMessage", "channel=C123", "text=hello"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "POST", receivedMethod)
	assert.Equal(t, "application/x-www-form-urlencoded", receivedContentType)
	assert.Contains(t, receivedBody, "channel=C123")
	assert.Contains(t, receivedBody, "text=hello")
	assert.Contains(t, receivedBody, "token=xoxb-test-token")
}

func Test_runAPICommand_JSONAutoDetect(t *testing.T) {
	var receivedContentType string
	var receivedBody string
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
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

	flags = cmdFlags{method: "POST"}
	cmd.SetArgs([]string{"chat.postMessage", `{"channel":"C123","text":"hello"}`})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "application/json; charset=utf-8", receivedContentType)
	assert.Equal(t, "Bearer xoxb-test-token", receivedAuth)

	var bodyJSON map[string]string
	err = json.Unmarshal([]byte(receivedBody), &bodyJSON)
	assert.NoError(t, err)
	assert.Equal(t, "C123", bodyJSON["channel"])
	assert.Equal(t, "hello", bodyJSON["text"])
}

func Test_runAPICommand_JSONFlag(t *testing.T) {
	var receivedContentType string
	var receivedAuth string
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		receivedAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
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

	flags = cmdFlags{method: "POST", json: `{"channel":"C123"}`}
	cmd.SetArgs([]string{"auth.test"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "application/json; charset=utf-8", receivedContentType)
	assert.Equal(t, "Bearer xoxb-test-token", receivedAuth)
	assert.Equal(t, `{"channel":"C123"}`, receivedBody)
}

func Test_runAPICommand_DataFlag(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	flags = cmdFlags{method: "POST", data: "channel=C123&text=hello"}
	cmd.SetArgs([]string{"chat.postMessage"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	assert.Contains(t, receivedBody, "channel=C123")
	assert.Contains(t, receivedBody, "text=hello")
	assert.Contains(t, receivedBody, "token=xoxb-test-token")
}

func Test_runAPICommand_GETMethod(t *testing.T) {
	var receivedMethod string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
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

	flags = cmdFlags{method: "GET"}
	cmd.SetArgs([]string{"auth.test"})
	err := cmd.ExecuteContext(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "GET", receivedMethod)
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
