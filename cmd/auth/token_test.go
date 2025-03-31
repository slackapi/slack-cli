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

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTokenCommand(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clientsMock.ApiInterface.On("ValidateSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(api.AuthSession{UserID: &mockOrgAuth.UserID,
		TeamID:   &mockOrgAuth.TeamID,
		TeamName: &mockOrgAuth.TeamDomain,
		URL:      &mockOrgAuthURL}, nil)
	clientsMock.AuthInterface.On("AuthWithTeamDomain", mock.Anything, mock.Anything).Return(types.SlackAuth{}, nil)
	clientsMock.AuthInterface.On("IsApiHostSlackProd", mock.Anything).Return(true)
	clientsMock.AddDefaultMocks()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Required by underlying loginFunc used for auth token command
	deprecatedUserTokenArg = ""
	tokenFlag = "xoxp-example-1234"

	cmd := NewTokenCommand(clients)
	cmd.SetContext(context.Background())
	testutil.MockCmdIO(clients.IO, cmd)

	serviceTokenFlag = true
	_, err := RunLoginCommand(clients, cmd)
	require.NoError(t, err)
	require.Contains(t, clientsMock.GetStdoutOutput(), "You've successfully authenticated")
}
