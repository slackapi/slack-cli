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

func TestPrompt_WorkflowSelectPrompt_empty_list(t *testing.T) {
	authorizationInfoLists := types.ExternalAuthorizationInfoLists{}

	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clientsMock.Config.Flags.Lookup("workflow"),
	})).Return(iostreams.SelectPromptResponse{}, slackerror.New(slackerror.ErrMissingOptions))
	clientsMock.AddDefaultMocks()
	ctx := context.Background()

	selectedWorkflow, err := WorkflowSelectPrompt(ctx, clients, authorizationInfoLists)
	require.Empty(t, selectedWorkflow)
	require.Error(t, err, slackerror.New("No workflows found that require developer authorization"))
	clientsMock.IO.AssertNotCalled(t, "SelectPrompt")
}

func TestPrompt_WorkflowSelectPrompt_with_no_workflows(t *testing.T) {
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

	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
		Flag: clients.Config.Flags.Lookup("workflow"),
	})).Return(iostreams.SelectPromptResponse{}, slackerror.New(slackerror.ErrMissingOptions))
	clientsMock.AddDefaultMocks()
	ctx := context.Background()

	selectedWorkflow, err := WorkflowSelectPrompt(ctx, clients, authorizationInfoLists)
	require.Empty(t, selectedWorkflow)
	require.Error(t, err, slackerror.New("No workflows found that require developer authorization"))
	clientsMock.IO.AssertNotCalled(t, "SelectPrompt")
}

func TestPrompt_WorkflowSelectPrompt_with_workflows(t *testing.T) {
	authorizationInfoLists := types.ExternalAuthorizationInfoLists{
		Authorizations: []types.ExternalAuthorizationInfo{
			{
				ProviderName:       "Google",
				ProviderKey:        "provider_a",
				ClientId:           "xxxxx",
				ClientSecretExists: true,
				ValidTokenExists:   false,
			},
		},
		Workflows: []types.WorkflowsInfo{
			{
				WorkflowId: "Wf0548LABCD1",
				CallBackId: "my_callback_id1",
				Providers: []types.ProvidersInfo{
					{
						ProviderKey:  "provider_a",
						ProviderName: "Provider_A",
					},
					{
						ProviderKey:  "provider_b",
						ProviderName: "Provider_B",
						SelectedAuth: types.ExternalTokenInfo{
							ExternalTokenId: "Et0548LABCDE",
							ExternalUserId:  "user_a@gmail.com",
							DateUpdated:     1682021142,
						},
					},
				},
			},
			{
				WorkflowId: "Wf0548LABCD2",
				CallBackId: "my_callback_id2",
				Providers: []types.ProvidersInfo{
					{
						ProviderKey:  "provider_a",
						ProviderName: "Provider_A",
					},
					{
						ProviderKey:  "provider_b",
						ProviderName: "Provider_B",
						SelectedAuth: types.ExternalTokenInfo{
							ExternalTokenId: "Et0548LABCDE",
							ExternalUserId:  "user_a@gmail.com",
							DateUpdated:     1682021142,
						},
					},
				},
			},
		},
	}

	tests := []struct {
		WorkflowFlag string
		Selection    iostreams.SelectPromptResponse
	}{
		{
			Selection: iostreams.SelectPromptResponse{
				Prompt: true,
				Option: "my_callback_id2",
				Index:  1,
			},
		},
		{
			WorkflowFlag: "my_callback_id2",
			Selection: iostreams.SelectPromptResponse{
				Flag:   true,
				Option: "my_callback_id2",
			},
		},
	}

	for _, tt := range tests {
		var mockWorkflowFlag string
		clientsMock := shared.NewClientsMock()
		clients := shared.NewClientFactory(clientsMock.MockClientFactory())
		clientsMock.Config.Flags.StringVar(&mockWorkflowFlag, "workflow", "", "mock workflow flag")
		if tt.WorkflowFlag != "" {
			_ = clientsMock.Config.Flags.Set("workflow", tt.WorkflowFlag)
		}
		clientsMock.IO.On("SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
			Flag: clientsMock.Config.Flags.Lookup("workflow"),
		})).Return(tt.Selection, nil)
		clientsMock.AddDefaultMocks()
		ctx := context.Background()

		selectedWorkflow, err := WorkflowSelectPrompt(ctx, clients, authorizationInfoLists)
		require.NoError(t, err)
		require.Equal(t, selectedWorkflow, types.WorkflowsInfo{
			WorkflowId: "Wf0548LABCD2",
			CallBackId: "my_callback_id2",
			Providers: []types.ProvidersInfo{
				{
					ProviderKey:  "provider_a",
					ProviderName: "Provider_A",
				},
				{
					ProviderKey:  "provider_b",
					ProviderName: "Provider_B",
					SelectedAuth: types.ExternalTokenInfo{
						ExternalTokenId: "Et0548LABCDE",
						ExternalUserId:  "user_a@gmail.com",
						DateUpdated:     1682021142,
					},
				},
			},
		})
		clientsMock.IO.AssertCalled(t, "SelectPrompt", mock.Anything, "Select a workflow", mock.Anything, mock.Anything)
	}
}
