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

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/pkg/externalauth"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var selectWorkflowFunc = externalauth.WorkflowSelectPrompt
var selectProviderAuthFunc = externalauth.ProviderAuthSelectPrompt
var workflowFlag string
var accountIdentifierFlag string

func NewSelectAuthCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select-auth",
		Short: "Select developer authentication of a workflow",
		Long: strings.Join([]string{
			"Select the saved developer authentication to use when calling external APIs from",
			"functions in a workflow app.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Select the saved developer authentication in a workflow",
				Command: "external-auth select-auth --workflow #/workflows/workflow_callback --provider google_provider --external-account user@salesforce.com",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunSelectAuthCommand(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelectAuthCommand(clients, cmd)
		},
	}

	// Flags
	cmd.Flags().StringVarP(&workflowFlag, "workflow", "W", "", "workflow to set developer authentication for")
	cmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "provider of the developer account")
	cmd.Flags().StringVarP(&accountIdentifierFlag, "external-account", "E", "", "external account identifier for the provider")
	return cmd
}

// preRunSelectAuthCommand determines if the command is supported for a project
// and configures flags
func preRunSelectAuthCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
	clients.Config.SetFlags(cmd)
	err := cmdutil.IsValidProjectDirectory(clients)
	if err != nil {
		return err
	}
	if clients.Config.ForceFlag {
		return nil
	}
	return cmdutil.IsSlackHostedProject(ctx, clients)
}

// runSelectAuthCommand prompts for developer authentication within a workflow
func runSelectAuthCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Get the app selection and accompanying auth from the prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Get the oauth2 details for the app
	externalAuths, err := clients.ApiInterface().AppsAuthExternalList(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		true, /*include_workflows flag to return workflow auth info*/
	)
	if err != nil {
		return err
	}

	// Get the workflows for the app
	selectedWorkflowAuth, err := selectWorkflowFunc(ctx, clients, externalAuths)
	if err != nil {
		if slackerror.ToSlackError(err).Code == slackerror.ErrMissingOptions {
			return slackerror.New("No workflows found that require developer authorization")
		}
		return err
	}
	if selectedWorkflowAuth.CallBackId == "" {
		return slackerror.New(slackerror.ErrWorkflowNotFound)
	}
	if selectedWorkflowAuth.WorkflowId == "" {
		return slackerror.New("Unable to get a workflow selection")
	}

	// Get the provider for the selected workflow
	var selectedProviderAuth types.ExternalAuthorizationInfo
	selectedProvider, err := selectProviderAuthFunc(ctx, clients, selectedWorkflowAuth)
	if err != nil {
		return err
	}
	for _, authorization := range externalAuths.Authorizations {
		if authorization.ProviderKey == selectedProvider.ProviderKey {
			selectedProviderAuth = authorization
			break
		}
	}
	if selectedProviderAuth.ProviderKey == "" {
		return slackerror.New("Provider is not used in the selected workflow")
	}

	// Get account identifier for selected provider
	selectedAuth, err := tokenSelectionFunc(ctx, clients, selectedProviderAuth)
	if err != nil {
		return err
	}
	if selectedAuth.ExternalTokenID == "" {
		return slackerror.New("Account is not used in the selected workflow")
	}

	err = clients.ApiInterface().AppsAuthExternalSelectAuth(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		selectedProviderAuth.ProviderKey,
		selectedWorkflowAuth.WorkflowId,
		selectedAuth.ExternalTokenID,
	)
	if err != nil {
		return slackerror.New(err.Error())
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "sparkles",
		Text: fmt.Sprintf(
			"Workflow #/workflows/%s will use developer account %s when making calls to %s APIs",
			selectedWorkflowAuth.CallBackId,
			selectedAuth.ExternalUserId,
			selectedProviderAuth.ProviderKey,
		),
	}))

	return nil
}
