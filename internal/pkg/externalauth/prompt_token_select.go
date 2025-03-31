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
	"time"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// Prompt user to select an oauth2 token from the exiting token list
func TokenSelectPrompt(ctx context.Context, clients *shared.ClientFactory, selectedProviderAuth types.ExternalAuthorizationInfo) (types.ExternalTokenInfo, error) {
	var timeFormat = "2006-01-02 15:04:05 Z07:00" // Based loosely on time.RFC3339
	var externalTokenOptions = []string{}
	var externalTokenMap = make(map[string]types.ExternalTokenInfo)
	var selectedExternalToken types.ExternalTokenInfo
	for _, externalToken := range selectedProviderAuth.ExternalTokens {
		externalTokenMap[externalToken.ExternalUserId] = externalToken
		lastUpdated := time.Unix(int64(externalToken.DateUpdated), 0).Format(timeFormat)
		externalTokenOptions = append(externalTokenOptions, fmt.Sprintf("Account: %s, Last Updated: %s", externalToken.ExternalUserId, lastUpdated))
	}
	if len(externalTokenOptions) == 0 {
		return types.ExternalTokenInfo{}, slackerror.New("No connected accounts found").WithRemediation("A token can be added to this app by running %s", style.Commandf("external-auth add", true))
	}
	selection, err := clients.IO.SelectPrompt(ctx, "Select an external account", externalTokenOptions, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("external-account"),
		Required: true,
	})
	if err != nil {
		return types.ExternalTokenInfo{}, err
	} else if selection.Flag {
		selectedExternalToken = externalTokenMap[selection.Option]
	} else if selection.Prompt {
		selectedExternalToken = selectedProviderAuth.ExternalTokens[selection.Index]
	}
	return selectedExternalToken, nil
}
