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
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceCommand(t *testing.T) {
	// Create mocks
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	// Create the command
	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	listPkgMock := new(ListPkgMock)
	listFunc = listPkgMock.List
	listPkgMock.On("List").Return(nil)

	err := cmd.Execute()
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	listPkgMock.AssertCalled(t, "List")
}

// TODO: this test may need a stubbed out parent (root) command to get aliasing working
/*
func TestPostRunWorkspaceDeprecationMessage(t *testing.T) {

	// Create mocks
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})
	clients.IO = clientsMock.IO
	cmd := NewCommand(clients)
	// TODO: could maybe refactor this to the os/fs mocks level to more clearly communicate "fake being in an app directory"
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error { return nil }
	args := []string{"team"}
	cmd.SetArgs(args)

	testutil.MockCmdIO(clientsMock.IO, cmd)
	listPkgMock := new(ListPkgMock)
	listFunc = listPkgMock.List
	listPkgMock.On("List").Return(nil)

	err := cmd.Execute()
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error", err.Error())
	}
	require.Contains(t, clientsMock.GetStdoutOutput(), "You can now use")
}
*/
