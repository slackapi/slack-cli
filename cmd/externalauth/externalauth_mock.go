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

package externalauth

import (
	"context"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/mock"
)

func NewExternalAuthMock() *ExternalAuthMock {
	return &ExternalAuthMock{}
}

// ExternalAuthMock implementation mocks for the externalauth package
type ExternalAuthMock struct {
	mock.Mock
}

// ProviderSelectPrompt mocks the provider selection prompt
func (m *ExternalAuthMock) ProviderSelectPrompt(ctx context.Context, clients *shared.ClientFactory, providers types.ExternalAuthorizationInfoLists) (types.ExternalAuthorizationInfo, error) {
	args := m.Called()
	return args.Get(0).(types.ExternalAuthorizationInfo), args.Error(1)
}
func (m *ExternalAuthMock) TokenSelectPrompt(ctx context.Context, clients *shared.ClientFactory, providers types.ExternalAuthorizationInfo) (types.ExternalTokenInfo, error) {
	args := m.Called()
	return args.Get(0).(types.ExternalTokenInfo), args.Error(1)
}
func (m *ExternalAuthMock) WorkflowSelectPrompt(ctx context.Context, clients *shared.ClientFactory, workflows types.ExternalAuthorizationInfoLists) (types.WorkflowsInfo, error) {
	args := m.Called()
	return args.Get(0).(types.WorkflowsInfo), args.Error(1)
}
func (m *ExternalAuthMock) ProviderAuthSelectPrompt(ctx context.Context, clients *shared.ClientFactory, providers types.WorkflowsInfo) (types.ProvidersInfo, error) {
	args := m.Called()
	return args.Get(0).(types.ProvidersInfo), args.Error(1)
}
