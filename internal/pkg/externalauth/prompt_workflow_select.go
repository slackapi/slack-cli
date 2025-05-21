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
)

// Prompt user to select a workflow from manifest that has at least one step requiring developer authentication
func WorkflowSelectPrompt(ctx context.Context, clients *shared.ClientFactory, providerAuths types.ExternalAuthorizationInfoLists) (types.WorkflowsInfo, error) {
	workflowCallbackPrefix := "#/workflows/"
	workflows := providerAuths.Workflows
	var workflowMap = make(map[string]types.WorkflowsInfo)
	var workflowOptions = []string{}
	var selectedWorkflow types.WorkflowsInfo
	for _, workflow := range workflows {
		workflowMap[workflow.CallbackID] = workflow
		providers := workflow.Providers
		var providerList strings.Builder
		for _, provider := range providers {
			var selectedExternalAccountID = "None"
			selectedAuth := provider.SelectedAuth
			if selectedAuth.ExternalTokenID != "" {
				selectedExternalAccountID = selectedAuth.ExternalUserID
			}
			fmt.Fprintf(&providerList, "\tKey: %s, Name: %s, Selected Account: %s\n", provider.ProviderKey, provider.ProviderName, selectedExternalAccountID)
		}
		optionText := fmt.Sprintf("Workflow: %s\n  Providers:\n %s", workflowCallbackPrefix+workflow.CallbackID, providerList.String())
		workflowOptions = append(workflowOptions, optionText)
	}
	selection, err := clients.IO.SelectPrompt(ctx, "Select a workflow", workflowOptions, iostreams.SelectPromptConfig{
		Flag:     clients.Config.Flags.Lookup("workflow"),
		Required: true,
	})
	if err != nil {
		return types.WorkflowsInfo{}, err
	} else if selection.Flag {
		callbackID := strings.TrimPrefix(selection.Option, workflowCallbackPrefix)
		selectedWorkflow = workflowMap[callbackID]
	} else if selection.Prompt {
		selectedWorkflow = workflows[selection.Index]
	}
	return selectedWorkflow, nil
}
