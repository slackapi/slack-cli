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
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type listMockObject struct {
	mock.Mock
}

func (m *listMockObject) MockListFunction(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger) ([]types.SlackAuth, error) {
	args := m.Called()
	return args.Get(0).([]types.SlackAuth), args.Error(1)
}

func TestAuthCommand(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	mock := new(listMockObject)
	listFunc = mock.MockListFunction
	mock.On("MockListFunction").Return([]types.SlackAuth{}, nil)

	// Execute test
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	// Check result
	mock.AssertCalled(t, "MockListFunction")
}
