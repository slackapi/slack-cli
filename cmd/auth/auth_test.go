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

package auth

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAuthCommand(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clientsMock.AddDefaultMocks()

	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.NoError(t, err)
	// Verify the auth command delegates to the list command; detailed list output testing is in TestListCommand
	assert.Contains(t, clientsMock.GetCombinedOutput(), "logged in")
}
