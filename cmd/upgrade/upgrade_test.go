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

package upgrade

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UpdatePkgMock struct {
	mock.Mock
}

func (m *UpdatePkgMock) CheckForUpdates(clients *shared.ClientFactory, cmd *cobra.Command, autoApprove bool) error {
	args := m.Called(clients, cmd, autoApprove)
	return args.Error(0)
}

func TestUpgradeCommand(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Create the command
	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	updatePkgMock := new(UpdatePkgMock)
	checkForUpdatesFunc = updatePkgMock.CheckForUpdates

	// Test default behavior (no auto-approve)
	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything, false).Return(nil)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade had unexpected error")
	}
	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything, false)

	// Test with auto-approve flag
	cmd = NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--auto-approve"})
	
	updatePkgMock = new(UpdatePkgMock)
	checkForUpdatesFunc = updatePkgMock.CheckForUpdates
	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything, true).Return(nil)
	
	err = cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade with auto-approve had unexpected error")
	}
	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything, true)
}
