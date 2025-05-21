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
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// Prompt user to select an oauth2 provider from manifest
func ProviderSelectPrompt(ctx context.Context, clients *shared.ClientFactory, providerAuths types.ExternalAuthorizationInfoLists) (types.ExternalAuthorizationInfo, error) {
	providers := providerAuths.Authorizations
	var providersMap = make(map[string]types.ExternalAuthorizationInfo)
	var providerOptions = []string{}
	var selectedProvider types.ExternalAuthorizationInfo
	for _, provider := range providers {
		providersMap[provider.ProviderKey] = provider
		var secretExists string
		if provider.ClientSecretExists {
			secretExists = "Yes"
		} else {
			secretExists = "No"
		}
		var externalTokens = provider.ExternalTokens
		if len(externalTokens) > 0 {
			externalAccountsList := []string{}
			for _, externalToken := range externalTokens {
				externalAccountsList = append(externalAccountsList, externalToken.ExternalUserID)
			}
			providerOptions = append(providerOptions, fmt.Sprintf("Provider Key: %s\n  Provider Name: %s\n  Client ID: %s\n  Client Secret Exists? %s\n  Valid Tokens: %s\n", provider.ProviderKey, provider.ProviderName, provider.ClientID, secretExists, strings.Join(externalAccountsList, ", ")))
		} else {
			var externalTokenExists string
			if provider.ValidTokenExists {
				externalTokenExists = "Yes"
			} else {
				externalTokenExists = "No"
			}
			providerOptions = append(providerOptions, fmt.Sprintf("Provider Key: %s\n  Provider Name: %s\n  Client ID: %s\n  Client Secret Exists? %s\n  Valid Token Exists? %s\n", provider.ProviderKey, provider.ProviderName, provider.ClientID, secretExists, externalTokenExists))
		}
	}
	if len(providerOptions) == 0 {
		return types.ExternalAuthorizationInfo{}, slackerror.New("No OAuth2 providers found")
	}
	selection, err := clients.IO.SelectPrompt(ctx, "Select a provider", providerOptions, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("provider"),
		Required: true,
	})
	if err != nil {
		return types.ExternalAuthorizationInfo{}, err
	} else if selection.Flag {
		selectedProvider = providersMap[selection.Option]
	} else if selection.Prompt {
		selectedProvider = providers[selection.Index]
	}
	return selectedProvider, nil
}
