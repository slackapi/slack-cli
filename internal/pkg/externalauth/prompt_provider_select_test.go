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

package externalauth

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrompt_ProviderSelectPrompt_empty_list(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	authorizationInfoLists := types.ExternalAuthorizationInfoLists{}
	clientsMock.AddDefaultMocks()
	ctx := context.Background()
	selectedProvider, err := ProviderSelectPrompt(ctx, clients, authorizationInfoLists)
	require.Empty(t, selectedProvider)
	require.Error(t, err, slackerror.New("No oauth2 providers found"))
}

func TestPrompt_ProviderSelectPrompt_no_token(t *testing.T) {
	authorizationInfoLists := types.ExternalAuthorizationInfoLists{
		Authorizations: []types.ExternalAuthorizationInfo{
			{
				ProviderName:       "Google",
				ProviderKey:        "provider_a",
				ClientId:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   false,
			},
		}}

	tests := []struct {
		ProviderFlag          string
		Selection             iostreams.SelectPromptResponse
		ExpectedAuthorization types.ExternalAuthorizationInfo
	}{
		{
			ProviderFlag: "provider_a",
			Selection: iostreams.SelectPromptResponse{
				Flag:   true,
				Option: "provider_a",
			},
			ExpectedAuthorization: authorizationInfoLists.Authorizations[0],
		},
		{
			Selection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "provider_a",
				Index:  0,
			},
			ExpectedAuthorization: authorizationInfoLists.Authorizations[0],
		},
	}

	for _, tt := range tests {
		var mockProviderFlag string
		clientsMock := shared.NewClientsMock()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		clientsMock.Config.Flags.StringVar(&mockProviderFlag, "provider", "", "mock provider flag")
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("provider"),
		})).Return(tt.Selection, nil)
		clientsMock.AddDefaultMocks()
		ctx := context.Background()

		selectedProvider, err := ProviderSelectPrompt(ctx, clients, authorizationInfoLists)
		require.Equal(t, selectedProvider.ProviderKey, "provider_a")
		require.Equal(t, selectedProvider, authorizationInfoLists.Authorizations[0])
		require.NoError(t, err)
		clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a provider", mock.Anything, mock.Anything)
	}
}

func TestPrompt_ProviderSelectPrompt_with_token(t *testing.T) {
	authorizationInfoLists := types.ExternalAuthorizationInfoLists{
		Authorizations: []types.ExternalAuthorizationInfo{
			{
				ProviderName:       "Google",
				ProviderKey:        "provider_a",
				ClientId:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   true,
				ExternalTokenIds:   []string{"Et0548LYDWCT"},
				ExternalTokens: []types.ExternalTokenInfo{
					{
						ExternalTokenId: "Et0548LABCDE",
						ExternalUserId:  "xyz@salesforce.com",
						DateUpdated:     1682021142,
					},
				},
			},
		}}

	tests := []struct {
		ProviderFlag          string
		Selection             iostreams.SelectPromptResponse
		ExpectedAuthorization types.ExternalAuthorizationInfo
	}{
		{
			ProviderFlag: "provider_a",
			Selection: iostreams.SelectPromptResponse{
				Flag:   true,
				Option: "provider_a",
			},
			ExpectedAuthorization: authorizationInfoLists.Authorizations[0],
		},
		{
			Selection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "provider_a",
				Index:  0,
			},
			ExpectedAuthorization: authorizationInfoLists.Authorizations[0],
		},
	}

	for _, tt := range tests {
		var mockProviderFlag string
		clientsMock := shared.NewClientsMock()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		clientsMock.Config.Flags.StringVar(&mockProviderFlag, "provider", "", "mock provider flag")
		if tt.ProviderFlag != "" {
			_ = clientsMock.Config.Flags.Set("provider", tt.ProviderFlag)
		}
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("provider"),
		})).Return(tt.Selection, nil)
		clientsMock.AddDefaultMocks()
		ctx := context.Background()

		selectedProvider, err := ProviderSelectPrompt(ctx, clients, authorizationInfoLists)
		require.Equal(t, selectedProvider.ProviderKey, "provider_a")
		require.Equal(t, selectedProvider, authorizationInfoLists.Authorizations[0])
		require.NoError(t, err)
		clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a provider", mock.Anything, mock.Anything)
	}
}
