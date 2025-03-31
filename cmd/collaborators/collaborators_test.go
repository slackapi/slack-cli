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

package collaborators

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCollaboratorsCommand(t *testing.T) {
	// Create mocks
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	clientsMock.ApiInterface.On("ListCollaborators", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]types.SlackUser{}, nil)

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	mockAuths := []types.SlackAuth{}
	clientsMock.AuthInterface.On("Auths", mock.Anything).Return(mockAuths, nil)

	// Create the command
	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	// Execute test
	err := cmd.Execute()
	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	// Check result
	clientsMock.ApiInterface.AssertCalled(t, "ListCollaborators", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCollaboratorsCommand_PrintSuccess(t *testing.T) {

	// Setup

	ctx := context.Background()
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Execute tests

	t.Run("Username will be used if present", func(t *testing.T) {
		user := types.SlackUser{Email: "joe.smith@company.com", ID: "U1234", PermissionType: types.OWNER}
		printSuccess(ctx, clients.IO, user, "added")
		assert.Contains(t, clientsMock.GetStdoutOutput(), "joe.smith@company.com successfully added as an owner collaborator on this app")
	})

	t.Run("User has no email set; fall back on user ID", func(t *testing.T) {
		user := types.SlackUser{ID: "U1234", PermissionType: types.OWNER}
		printSuccess(ctx, clients.IO, user, "removed")
		assert.Contains(t, clientsMock.GetStdoutOutput(), "\nU1234 successfully removed as an owner collaborator on this app\n\n")
	})

	t.Run("Reader-type collaborator", func(t *testing.T) {
		user := types.SlackUser{Email: "joe.smith@company.com", ID: "U1234", PermissionType: types.READER}
		printSuccess(ctx, clients.IO, user, "updated")
		assert.Contains(t, clientsMock.GetStdoutOutput(), "\njoe.smith@company.com successfully updated as a reader collaborator on this app\n\n")
	})

}
