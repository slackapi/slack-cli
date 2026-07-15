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

package manifest

import (
	"context"
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type syncTestFixture struct {
	ctx           context.Context
	clients       *shared.ClientFactory
	clientsMock   *shared.ClientsMock
	manifestMock  *app.ManifestMockObject
	projectConfig *config.ProjectConfigMock
	cacheMock     *cache.CacheMock
	fs            afero.Fs
}

func newSyncTestFixture(t *testing.T) *syncTestFixture {
	t.Helper()
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	manifestMock := &app.ManifestMockObject{}
	clientsMock.AppClient.Manifest = manifestMock

	projectConfig := config.NewProjectConfigMock()
	cacheMock := cache.NewCacheMock()
	projectConfig.On("Cache").Return(cacheMock)
	clientsMock.Config.ProjectConfig = projectConfig

	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clients.SDKConfig = hooks.SDKCLIConfig{WorkingDirectory: "/project"}

	fs := afero.NewMemMapFs()
	clients.Fs = fs

	ctx := slackcontext.MockContext(t.Context())

	return &syncTestFixture{
		ctx:           ctx,
		clients:       clients,
		clientsMock:   clientsMock,
		manifestMock:  manifestMock,
		projectConfig: projectConfig,
		cacheMock:     cacheMock,
		fs:            fs,
	}
}

func Test_Sync(t *testing.T) {
	testApp := types.App{AppID: "A123", TeamDomain: "test-team"}
	testAuth := types.SlackAuth{Token: "xoxb-test"}

	localManifest := types.SlackYaml{
		AppManifest: types.AppManifest{DisplayInformation: types.DisplayInformation{Name: "App", Description: "Local"}},
	}
	remoteManifest := types.SlackYaml{
		AppManifest: types.AppManifest{DisplayInformation: types.DisplayInformation{Name: "App", Description: "Remote"}},
	}
	identicalManifest := types.SlackYaml{
		AppManifest: types.AppManifest{DisplayInformation: types.DisplayInformation{Name: "App", Description: "Same"}},
	}

	t.Run("returns error when manifest source is remote", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceRemote, nil)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		slackErr := slackerror.ToSlackError(err)
		assert.Equal(t, slackerror.ErrAppManifestUpdate, slackErr.Code)
	})

	t.Run("returns error when GetManifestSource fails", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, fmt.Errorf("config read failed"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when GetManifestLocal fails", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(types.SlackYaml{}, fmt.Errorf("hook failed"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "local manifest")
	})

	t.Run("returns error when GetManifestRemote fails", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(types.SlackYaml{}, fmt.Errorf("api timeout"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "remote manifest")
	})

	t.Run("no differences returns in-sync result", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(identicalManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(identicalManifest, nil)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.HasDifferences)
	})

	t.Run("non-TTY without force returns error with remediation", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = false

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		slackErr := slackerror.ToSlackError(err)
		assert.Equal(t, slackerror.ErrAppManifestUpdate, slackErr.Code)
	})

	t.Run("force flag merges all local and pushes to API", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = true
		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).Return(cache.Hash("newhash"), nil)
		f.cacheMock.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		_ = afero.WriteFile(f.fs, "/project/manifest.json", []byte(`{"display_information":{"name":"App"}}`), 0644)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.HasDifferences)
		assert.True(t, result.WriteBack.Written)
		f.clientsMock.API.AssertCalled(t, "UpdateApp", mock.Anything, "xoxb-test", "A123", mock.Anything, true, true)
	})

	t.Run("API UpdateApp failure is propagated", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = true
		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, fmt.Errorf("rate limited"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "update app settings")
	})

	t.Run("cache NewManifestHash failure is propagated", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = true
		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).
			Return(cache.Hash(""), fmt.Errorf("hash failure"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "hash")
	})

	t.Run("cache SetManifestHash failure is propagated", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = true
		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).Return(cache.Hash("h"), nil)
		f.cacheMock.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).
			Return(fmt.Errorf("write failed"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cache")
	})

	t.Run("missing manifest.json still succeeds with warning", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = true
		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).Return(cache.Hash("h"), nil)
		f.cacheMock.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.HasDifferences)
		assert.False(t, result.WriteBack.Written)
	})

	t.Run("TTY interactive resolution with all-local strategy", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = false

		// Override IsTTY to return true
		f.clientsMock.IO.ExpectedCalls = removeCallsByMethod(f.clientsMock.IO.ExpectedCalls, "IsTTY")
		f.clientsMock.IO.On("IsTTY").Return(true)

		// User picks "Use all project values" (index 0 = MergeAllLocal)
		f.clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Index: 0}, nil)

		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).Return(cache.Hash("h"), nil)
		f.cacheMock.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		_ = afero.WriteFile(f.fs, "/project/manifest.json", []byte(`{"display_information":{"name":"App"}}`), 0644)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.HasDifferences)
		assert.True(t, result.WriteBack.Written)
	})

	t.Run("TTY interactive resolution with all-remote strategy", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = false

		f.clientsMock.IO.ExpectedCalls = removeCallsByMethod(f.clientsMock.IO.ExpectedCalls, "IsTTY")
		f.clientsMock.IO.On("IsTTY").Return(true)

		// User picks "Use all app settings values" (index 1 = MergeAllRemote)
		f.clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Index: 1}, nil)

		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).Return(cache.Hash("h"), nil)
		f.cacheMock.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		_ = afero.WriteFile(f.fs, "/project/manifest.json", []byte(`{"display_information":{"name":"App"}}`), 0644)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.HasDifferences)
	})

	t.Run("TTY interactive per-field resolution", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = false

		f.clientsMock.IO.ExpectedCalls = removeCallsByMethod(f.clientsMock.IO.ExpectedCalls, "IsTTY")
		f.clientsMock.IO.On("IsTTY").Return(true)

		// First prompt: user picks "Choose for each difference" (index 2 = MergePerField)
		// Second prompt: user picks local for the single diff (index 0)
		f.clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Index: 2}, nil).Once()
		f.clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{Index: 0}, nil).Once()

		f.clientsMock.API.On("UpdateApp", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(api.UpdateAppResult{}, nil)
		f.cacheMock.On("NewManifestHash", mock.Anything, mock.Anything).Return(cache.Hash("h"), nil)
		f.cacheMock.On("SetManifestHash", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		_ = afero.WriteFile(f.fs, "/project/manifest.json", []byte(`{"display_information":{"name":"App"}}`), 0644)

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.HasDifferences)
		assert.True(t, result.WriteBack.Written)
	})

	t.Run("TTY prompt error is propagated", func(t *testing.T) {
		f := newSyncTestFixture(t)
		f.projectConfig.On("GetManifestSource", mock.Anything).Return(config.ManifestSourceLocal, nil)
		f.manifestMock.On("GetManifestLocal", mock.Anything, mock.Anything, mock.Anything).
			Return(localManifest, nil)
		f.manifestMock.On("GetManifestRemote", mock.Anything, mock.Anything, mock.Anything).
			Return(remoteManifest, nil)
		f.clients.Config.ForceFlag = false

		f.clientsMock.IO.ExpectedCalls = removeCallsByMethod(f.clientsMock.IO.ExpectedCalls, "IsTTY")
		f.clientsMock.IO.On("IsTTY").Return(true)

		f.clientsMock.IO.On("SelectPrompt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(iostreams.SelectPromptResponse{}, fmt.Errorf("interrupt"))

		result, err := Sync(f.ctx, f.clients, testApp, testAuth)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func removeCallsByMethod(calls []*mock.Call, method string) []*mock.Call {
	filtered := make([]*mock.Call, 0, len(calls))
	for _, c := range calls {
		if c.Method != method {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
