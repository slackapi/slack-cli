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

package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// ManifestClient can manage the state of the project's app manifest file
type ManifestClient struct {
	apiClient        api.ApiInterface
	domainAuthTokens string
	Env              map[string]string
}

type ManifestClientInterface interface {
	GetManifestLocal(sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor) (types.SlackYaml, error)
	GetManifestRemote(ctx context.Context, token string, appID string) (types.SlackYaml, error)
}

// SetManifestEnvTeamVars sets environment variables that may affect app manifest values
func SetManifestEnvTeamVars(manifestEnv map[string]string, appTeamDomain string, isDev bool) map[string]string {
	localOrDeployed := "deployed"
	if isDev {
		localOrDeployed = "local"
	}

	if manifestEnv == nil {
		manifestEnv = map[string]string{}
	}

	manifestEnv["SLACK_WORKSPACE"] = appTeamDomain
	manifestEnv["SLACK_ENV"] = localOrDeployed

	return manifestEnv
}

// NewManifestClient returns a new, empty instance of the ManifestClient
func NewManifestClient(
	apiClient api.ApiInterface,
	config *config.Config,
) *ManifestClient {
	client := &ManifestClient{
		apiClient:        apiClient,
		domainAuthTokens: config.DomainAuthTokens,
		Env:              config.ManifestEnv,
	}
	return client
}

// GetManifestLocal gathers manifest content from the "get-manifest" hook
func (c *ManifestClient) GetManifestLocal(sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor) (types.SlackYaml, error) {
	var sl types.SlackYaml

	if !sdkConfig.Hooks.GetManifest.IsAvailable() {
		return sl, slackerror.New(slackerror.ErrSDKHookNotFound).
			WithMessage("The `get-manifest` script was not found")
	}

	var manifestHookOpts = hooks.HookExecOpts{
		Args: map[string]string{
			"source": sdkConfig.WorkingDirectory,
		},
		Env: map[string]string{
			"DENO_AUTH_TOKENS": c.domainAuthTokens,
		},
		Hook: sdkConfig.Hooks.GetManifest,
	}

	for name, val := range c.Env {
		manifestHookOpts.Env[name] = val
	}

	slackManifestInfo, err := hookExecutor.Execute(manifestHookOpts)
	if err != nil {
		return sl, slackerror.New("Failed to get app manifest details. Please check your manifest file.").
			WithRootCause(err).
			WithCode(slackerror.ErrInvalidManifest)
	}
	// Because we listen to file changes on localrun, we sometimes get a list
	// of files that changed printed on the console so we want to skip straight to where the
	// actual definition starts
	start := strings.Index(slackManifestInfo, "{")
	if start != -1 {
		slackManifestInfo = slackManifestInfo[start:]
	} else {
		// the app manifest has to be a json so needs to have the character `{`
		return sl, slackerror.New("Invalid app manifest format, must be valid JSON").
			WithRootCause(err).
			WithCode(slackerror.ErrInvalidManifest)
	}

	err = json.Unmarshal([]byte(slackManifestInfo), &sl)
	return sl, err
}

// GetManifestRemote retrieves the current app manifest from app settings
func (c *ManifestClient) GetManifestRemote(ctx context.Context, token string, appID string) (types.SlackYaml, error) {
	response, err := c.apiClient.ExportAppManifest(ctx, token, appID)
	if err != nil {
		return types.SlackYaml{}, err
	}
	return response.Manifest, nil
}
