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

package prompts

import (
	"context"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/stretchr/testify/mock"
)

// AppSelectMock implementation mocks the app environment prompts
type AppSelectMock struct {
	mock.Mock
}

// NewAppSelectMock returns a mock of the app environment prompts
func NewAppSelectMock() *AppSelectMock {
	return &AppSelectMock{}
}

// AppSelectPrompt mocks the app selection prompt
func (m *AppSelectMock) AppSelectPrompt(ctx context.Context, clients *shared.ClientFactory, env AppEnvironmentType, status AppInstallStatus) (SelectedApp, error) {
	args := m.Called(ctx, clients, env, status)
	return args.Get(0).(SelectedApp), args.Error(1)
}
