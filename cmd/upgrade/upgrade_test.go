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

func (m *UpdatePkgMock) CheckForUpdates(clients *shared.ClientFactory, cmd *cobra.Command, cli bool, sdk bool) error {
	args := m.Called(clients, cmd, cli, sdk)
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

	// Test default behavior (no flags)
	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything, false, false).Return(nil)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade had unexpected error")
	}
	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything, false, false)

	// Test with CLI flag
	cmd = NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--cli"})

	updatePkgMock = new(UpdatePkgMock)
	checkForUpdatesFunc = updatePkgMock.CheckForUpdates
	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything, true, false).Return(nil)

	err = cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade with cli flag had unexpected error")
	}
	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything, true, false)

	// Test with SDK flag
	cmd = NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--sdk"})

	updatePkgMock = new(UpdatePkgMock)
	checkForUpdatesFunc = updatePkgMock.CheckForUpdates
	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything, false, true).Return(nil)

	err = cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade with sdk flag had unexpected error")
	}
	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything, false, true)

	// Test with both CLI and SDK flags
	cmd = NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--cli", "--sdk"})

	updatePkgMock = new(UpdatePkgMock)
	checkForUpdatesFunc = updatePkgMock.CheckForUpdates
	updatePkgMock.On("CheckForUpdates", mock.Anything, mock.Anything, true, true).Return(nil)

	err = cmd.ExecuteContext(ctx)
	if err != nil {
		assert.Fail(t, "cmd.Upgrade with both cli and sdk flags had unexpected error")
	}
	updatePkgMock.AssertCalled(t, "CheckForUpdates", mock.Anything, mock.Anything, true, true)
}

func TestUpgradeCommandWithFlagError(t *testing.T) {
	// Create a mock of UpdateNotification that returns an error on InstallUpdatesWithComponentFlags
	originalCheckForUpdates := checkForUpdatesFunc
	defer func() {
		checkForUpdatesFunc = originalCheckForUpdates
	}()

	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Mock the checkForUpdates function to simulate an error during flag-based updates
	checkForUpdatesFunc = func(clients *shared.ClientFactory, cmd *cobra.Command, cli bool, sdk bool) error {
		if cli || sdk {
			return assert.AnError // Simulate error when either flag is true
		}
		return nil
	}

	// Test with CLI flag causing error
	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--cli"})

	// Execute the command and verify it returns the error
	err := cmd.ExecuteContext(ctx)

	// Verify the error was properly propagated
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Test with SDK flag causing error
	cmd = NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--sdk"})

	// Execute the command and verify it returns the error
	err = cmd.ExecuteContext(ctx)

	// Verify the error was properly propagated
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Test with both flags causing error
	cmd = NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)
	cmd.SetArgs([]string{"--cli", "--sdk"})

	// Execute the command and verify it returns the error
	err = cmd.ExecuteContext(ctx)

	// Verify the error was properly propagated
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}
