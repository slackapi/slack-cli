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
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrompt_TokenSelectPrompt_empty_list(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	authorizationInfo := types.ExternalAuthorizationInfo{}
	clientsMock.AddDefaultMocks()

	selectedToken, err := TokenSelectPrompt(ctx, clients, authorizationInfo)
	require.Empty(t, selectedToken)
	require.Error(t, err, slackerror.New("No connected accounts found"))
	clientsMock.IO.AssertNotCalled(t, "SelectPrompt")
}

func TestPrompt_TokenSelectPrompt_with_token(t *testing.T) {
	authorizationInfo := types.ExternalAuthorizationInfo{
		ProviderName:       "Google",
		ProviderKey:        "provider_a",
		ClientID:           "xxxxx",
		ClientSecretExists: true,
		ValidTokenExists:   true,
		ExternalTokenIDs:   []string{"Et0548LYDWCT"},
		ExternalTokens: []types.ExternalTokenInfo{
			{
				ExternalTokenID: "Et0548LABCD1",
				ExternalUserID:  "xyz@salesforce.com",
				DateUpdated:     1682021142,
			},
			{
				ExternalTokenID: "Et0548LABCDE2",
				ExternalUserID:  "xyz2@salesforce.com",
				DateUpdated:     1682021192,
			},
		},
	}

	tests := map[string]struct {
		ExternalAccountFlag string
		Selection           iostreams.SelectPromptResponse
	}{
		"Prompt selection": {
			Selection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "xyz2@salesforce.com",
				Index:  1,
			},
		},
		"Flag selection": {
			ExternalAccountFlag: "xyz2@salesforce.com",
			Selection: iostreams.SelectPromptResponse{
				Flag:   true,
				Option: "xyz2@salesforce.com",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var externalAccountFlag string
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			clientsMock.Config.Flags.StringVar(&externalAccountFlag, "external-account", "", "mock external-account flag")
			if tc.ExternalAccountFlag != "" {
				_ = clientsMock.Config.Flags.Set("external-account", tc.ExternalAccountFlag)
			}
			clientsMock.IO.On("SelectPrompt", mock.Anything, "Select an external account", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
				Flag: clientsMock.Config.Flags.Lookup("external-account"),
			})).Return(tc.Selection, nil)
			clientsMock.AddDefaultMocks()

			selectedToken, err := TokenSelectPrompt(ctx, clients, authorizationInfo)
			require.NoError(t, err)
			require.Equal(t, selectedToken, types.ExternalTokenInfo{
				ExternalTokenID: "Et0548LABCDE2",
				ExternalUserID:  "xyz2@salesforce.com",
				DateUpdated:     1682021192,
			})
			clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select an external account", mock.Anything, mock.Anything)
		})
	}
}
