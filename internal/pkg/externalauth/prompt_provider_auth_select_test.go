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
	"testing"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrompt_ProviderAuthSelectPrompt_empty_list(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	workflowsInfo := types.WorkflowsInfo{}
	clientsMock.AddDefaultMocks()
	selectedProvider, err := ProviderAuthSelectPrompt(ctx, clients, workflowsInfo)
	require.Empty(t, selectedProvider)
	require.Error(t, err, slackerror.New("No oauth2 providers found"))
	clientsMock.IO.AssertNotCalled(t, "SelectPrompt")
}

func TestPrompt_ProviderAuthSelectPrompt_no_selected_auth(t *testing.T) {
	workflowsInfo := types.WorkflowsInfo{
		WorkflowID: "Wf0548LABCD1",
		CallbackID: "my_callback_id1",
		Providers: []types.ProvidersInfo{
			{
				ProviderKey:  "provider_a",
				ProviderName: "Provider_A",
			},
			{
				ProviderKey:  "provider_b",
				ProviderName: "Provider_B",
				SelectedAuth: types.ExternalTokenInfo{
					ExternalTokenID: "Et0548LABCDE",
					ExternalUserID:  "user_a@example.com",
					DateUpdated:     1682021142,
				},
			},
		},
	}

	tests := []struct {
		ProviderFlag     string
		Selection        iostreams.SelectPromptResponse
		ExpectedProvider types.ProvidersInfo
	}{
		{
			Selection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "provider_a",
				Index:  0,
			},
			ExpectedProvider: workflowsInfo.Providers[0],
		},
		{
			ProviderFlag: "provider_a",
			Selection: iostreams.SelectPromptResponse{
				Flag:   true,
				Option: "provider_a",
			},
			ExpectedProvider: workflowsInfo.Providers[0],
		},
	}

	for _, tc := range tests {
		var mockProviderFlag string
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		clientsMock.Config.Flags.StringVar(&mockProviderFlag, "provider", "", "mock provider flag")
		if tc.ProviderFlag != "" {
			_ = clientsMock.Config.Flags.Set("provider", tc.ProviderFlag)
		}
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clients.Config.Flags.Lookup("provider"),
		})).Return(tc.Selection, nil)

		clientsMock.AddDefaultMocks()

		selectedProvider, err := ProviderAuthSelectPrompt(ctx, clients, workflowsInfo)
		require.Equal(t, selectedProvider.ProviderKey, "provider_a")
		require.Equal(t, selectedProvider, workflowsInfo.Providers[0])
		require.NoError(t, err)
		clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a provider", mock.Anything, mock.Anything)
	}
}

func TestPrompt_ProviderAuthSelectPrompt_with_selected_auth(t *testing.T) {
	workflowsInfo := types.WorkflowsInfo{
		WorkflowID: "Wf0548LABCD1",
		CallbackID: "my_callback_id1",
		Providers: []types.ProvidersInfo{
			{
				ProviderKey:  "provider_a",
				ProviderName: "Provider_A",
			},
			{
				ProviderKey:  "provider_b",
				ProviderName: "Provider_B",
				SelectedAuth: types.ExternalTokenInfo{
					ExternalTokenID: "Et0548LABCDE",
					ExternalUserID:  "user_a@example.com",
					DateUpdated:     1682021142,
				},
			},
		},
	}

	tests := []struct {
		ProviderFlag     string
		Selection        iostreams.SelectPromptResponse
		ExpectedProvider types.ProvidersInfo
	}{
		{
			Selection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "provider_b",
				Index:  1,
			},
			ExpectedProvider: workflowsInfo.Providers[1],
		},
		{
			ProviderFlag: "provider_b",
			Selection: iostreams.SelectPromptResponse{
				Flag:   true,
				Option: "provider_b",
			},
			ExpectedProvider: workflowsInfo.Providers[1],
		},
	}

	for _, tc := range tests {
		var mockProviderFlag string
		ctx := slackcontext.MockContext(t.Context())
		clientsMock := shared.NewClientsMock()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		clientsMock.Config.Flags.StringVar(&mockProviderFlag, "provider", "", "mock provider flag")
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a provider", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("provider"),
		})).Return(tc.Selection, nil)
		clientsMock.AddDefaultMocks()

		selectedProvider, err := ProviderAuthSelectPrompt(ctx, clients, workflowsInfo)
		require.Equal(t, selectedProvider.ProviderKey, "provider_b")
		require.Equal(t, selectedProvider, workflowsInfo.Providers[1])
		require.NoError(t, err)
		clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a provider", mock.Anything, mock.Anything)
	}
}
