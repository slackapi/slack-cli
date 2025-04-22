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

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/mock"
)

type ManifestMockObject struct {
	mock.Mock
}

func (m *ManifestMockObject) GetManifestLocal(ctx context.Context, sdkConfig hooks.SDKCLIConfig, hookExecutor hooks.HookExecutor) (types.SlackYaml, error) {
	args := m.Called(ctx, sdkConfig, hookExecutor)
	return args.Get(0).(types.SlackYaml), args.Error(1)
}

func (m *ManifestMockObject) GetManifestRemote(ctx context.Context, token string, appID string) (types.SlackYaml, error) {
	args := m.Called(ctx, token, appID)
	return args.Get(0).(types.SlackYaml), args.Error(1)
}
