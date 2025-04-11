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
	"errors"
	"testing"
	"time"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthList(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	authMockA := types.SlackAuth{
		Token:               "xoxe.xoxp-",
		TeamDomain:          "aaa-aaa-aaa",
		TeamID:              "T123456789A",
		UserID:              "U123",
		LastUpdated:         time.Time{},
		RefreshToken:        "",
		ExpiresAt:           12345,
		IsEnterpriseInstall: false,
	}

	mockAuths := []types.SlackAuth{
		authMockA,
	}
	clientsMock.AuthInterface.On("Auths", mock.Anything).Return(mockAuths, errors.New("There was an error"))

	t.Run("Handles error getting auth slice", func(t *testing.T) {
		_, err := List(ctx, clients, &logger.Logger{})
		require.Error(t, err)
	})
}

func TestAuthList_SortedAuths(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
		clients.SDKConfig = hooks.NewSDKConfigMock()
	})

	authZ := types.SlackAuth{
		Token:       "xoxp-abc",
		TeamDomain:  "zee-zee-zee",
		TeamID:      "T03---",
		UserID:      "U03---",
		LastUpdated: time.Now(),
	}
	authC := types.SlackAuth{
		Token:       "xoxp-abc",
		TeamDomain:  "cee-cee-cee",
		TeamID:      "T02---",
		UserID:      "U02---",
		LastUpdated: time.Now(),
	}
	authB := types.SlackAuth{
		Token:       "xoxp-abc",
		TeamDomain:  "bee-bee-bee",
		TeamID:      "T01---",
		UserID:      "U01---",
		LastUpdated: time.Now(),
	}

	mockAuths := []types.SlackAuth{
		authZ,
		authC,
		authB,
	}

	clientsMock.AuthInterface.On("Auths", mock.Anything).Return(mockAuths, nil)

	_, err := List(ctx, clients, &logger.Logger{})
	require.NoError(t, err)

	// require.Equal(t, expectSortedAuths[0], authB, "should sort alphabetically as first elem")
	// require.Equal(t, "bee-bee-bee", expectSortedAuths[0].TeamDomain, "should add TeamDomain")

	// require.Equal(t, expectSortedAuths[1], authC, "should sort alphabetically as middle elem")
	// require.Equal(t, "cee-cee-cee", expectSortedAuths[1].TeamDomain, "should add TeamDomain")

	// require.Equal(t, expectSortedAuths[2], authZ, "should sort alphabetically z as last elem")
	// require.Equal(t, "zee-zee-zee", expectSortedAuths[2].TeamDomain, "should add TeamDomain")
}
