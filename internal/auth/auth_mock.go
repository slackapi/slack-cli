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

package auth

import (
	"context"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/mock"
)

type AuthMock struct {
	mock.Mock
}

func (m *AuthMock) AddDefaultMocks() {
	m.On("Auths", mock.Anything).Return([]types.SlackAuth{}, nil)
	m.On("FilterKnownAuthErrors", mock.Anything, mock.Anything).Return(false, nil)
	m.On("SetSelectedAuth", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (m *AuthMock) AuthWithToken(ctx context.Context, token string) (types.SlackAuth, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(types.SlackAuth), args.Error(1)
}

func (m *AuthMock) AuthWithTeamDomain(ctx context.Context, teamDomain string) (types.SlackAuth, error) {
	args := m.Called(ctx, teamDomain)
	return args.Get(0).(types.SlackAuth), args.Error(1)
}

func (m *AuthMock) AuthWithTeamID(ctx context.Context, teamID string) (types.SlackAuth, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(types.SlackAuth), args.Error(1)
}

func (m *AuthMock) Auths(ctx context.Context) ([]types.SlackAuth, error) {
	args := m.Called(ctx)
	return args.Get(0).([]types.SlackAuth), args.Error(1)
}

func (m *AuthMock) SetAuth(ctx context.Context, auth types.SlackAuth) (types.SlackAuth, string, error) {
	args := m.Called(ctx, auth)
	return args.Get(0).(types.SlackAuth), args.String(1), args.Error(2)
}

func (m *AuthMock) SetSelectedAuth(ctx context.Context, auth types.SlackAuth, config *config.Config, os types.Os) {
	m.Called(ctx, auth, config, os)
}

func (m *AuthMock) DeleteAuth(ctx context.Context, auth types.SlackAuth) (types.SlackAuth, error) {
	args := m.Called(ctx, auth)
	return args.Get(0).(types.SlackAuth), args.Error(1)
}

func (m *AuthMock) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *AuthMock) ResolveApiHost(ctx context.Context, apiHostFlag string, customAuth *types.SlackAuth) string {
	args := m.Called(ctx, apiHostFlag, customAuth)
	return args.String(0)
}

func (m *AuthMock) ResolveLogstashHost(ctx context.Context, apiHost string, cliVersion string) string {
	args := m.Called(ctx, apiHost, cliVersion)
	return args.String(0)
}

func (m *AuthMock) MapAuthTokensToDomains(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func (m *AuthMock) IsApiHostSlackDev(host string) bool {
	args := m.Called(host)
	return args.Bool(0)
}

func (m *AuthMock) IsApiHostSlackProd(host string) bool {
	args := m.Called(host)
	return args.Bool(0)
}

func (m *AuthMock) FilterKnownAuthErrors(ctx context.Context, err error) (bool, error) {
	args := m.Called(ctx, err)
	return args.Bool(0), args.Error(1)
}
