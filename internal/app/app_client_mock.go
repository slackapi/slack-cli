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

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/mock"
)

type AppClientMock struct {
	mock.Mock
}

func (m *AppClientMock) NewDeployed(ctx context.Context, teamId string) (types.App, error) {
	args := m.Called(ctx, teamId)
	return args.Get(0).(types.App), args.Error(1)
}

func (m *AppClientMock) GetDeployed(ctx context.Context, teamID string) (types.App, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(types.App), args.Error(1)
}

func (m *AppClientMock) GetDeployedAll(ctx context.Context) ([]types.App, string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.App), args.String(1), args.Error(2)
}

func (m *AppClientMock) GetLocal(ctx context.Context, teamID string) (types.App, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(types.App), args.Error(1)
}

func (m *AppClientMock) GetLocalAll(ctx context.Context) ([]types.App, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.App), args.Error(1)
}

func (m *AppClientMock) SaveDeployed(ctx context.Context, app types.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *AppClientMock) SaveLocal(ctx context.Context, app types.App) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *AppClientMock) RemoveDeployed(ctx context.Context, teamID string) (types.App, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(types.App), args.Error(1)
}

func (m *AppClientMock) RemoveLocal(ctx context.Context, userID string) (types.App, error) {
	args := m.Called(ctx)
	return args.Get(0).(types.App), args.Error(1)
}

func (m *AppClientMock) Remove(ctx context.Context, app types.App) (types.App, error) {
	args := m.Called(ctx, app)
	return args.Get(0).(types.App), args.Error(1)
}

func (m *AppClientMock) CleanUp() {
	m.Called()
}
