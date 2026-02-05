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
	"context"
	"fmt"
	"time"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Prompt to select an oauth2 provider from the selected workflow
func ProviderAuthSelectPrompt(ctx context.Context, clients *shared.ClientFactory, workflowInfo types.WorkflowsInfo) (types.ProvidersInfo, error) {
	providers := workflowInfo.Providers
	var timeFormat = "2006-01-02 15:04:05 Z07:00" // Based loosely on time.RFC3339
	var providerMap = make(map[string]types.ProvidersInfo)
	var providerOptions = []string{}
	var selectedProvider types.ProvidersInfo
	for _, provider := range providers {
		providerMap[provider.ProviderKey] = provider
		selectedAuth := provider.SelectedAuth
		if selectedAuth.ExternalTokenID == "" {
			providerOptions = append(providerOptions,
				fmt.Sprintf("Key: %s, Name: %s, Selected Account: None", provider.ProviderKey, provider.ProviderName))
		} else {
			lastUpdated := time.Unix(int64(selectedAuth.DateUpdated), 0).Format(timeFormat)
			providerOptions = append(providerOptions, fmt.Sprintf("Key: %s, Name: %s, Selected Account: %s, Last Updated: %s", provider.ProviderKey, provider.ProviderName, selectedAuth.ExternalUserID, lastUpdated))
		}
	}
	if len(providerOptions) == 0 {
		return types.ProvidersInfo{}, slackerror.New("No provider found which requires developer authentication")
	}
	selection, err := clients.IO.SelectPrompt(ctx, "Select a provider", providerOptions, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("provider"),
		Required: true,
	})
	if err != nil {
		return types.ProvidersInfo{}, err
	} else if selection.Flag {
		selectedProvider = providerMap[selection.Option]
	} else if selection.Prompt {
		selectedProvider = providers[selection.Index]
	}
	return selectedProvider, nil
}
