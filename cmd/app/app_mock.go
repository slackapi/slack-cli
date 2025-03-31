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

	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/mock"
)

// NewAppCommandMock returns a mock of the Workspace prompts.
func NewAppCommandMock() *AppMock {
	return &AppMock{}
}

// AppMock implementation mocks for the app package
type AppMock struct {
	mock.Mock
}

// RunAddCommand mocks the app add (install) command
func (m *AppMock) RunAddCommand(ctx context.Context, clients *shared.ClientFactory, selectedApp *prompts.SelectedApp, orgGrant string) (context.Context, types.InstallState, types.App, error) {
	args := m.Called(ctx, clients, selectedApp, orgGrant)
	return args.Get(0).(context.Context), args.Get(1).(types.InstallState), args.Get(2).(types.App), args.Error(3)
}
