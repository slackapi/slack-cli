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
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewAddCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Initiate the OAuth2 flow for a provider",
		Long: strings.Join([]string{
			"Initiate the OAuth2 flow for an external auth provider of a workflow app.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Select a provider to initiate the OAuth2 flow for",
				Command: "external-auth add",
			},
			{
				Meaning: "Initiate the OAuth2 flow for the provided provider",
				Command: "external-auth add -p github",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunAddCommand(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddCommand(clients, cmd)
		},
	}

	// Flags
	cmd.PersistentFlags().StringVarP(&providerFlag, "provider", "p", "", "the external auth Provider Key to add a secret to")

	return cmd
}

// preRunAddCommand determines if the command is supported for a project and
// configures flags
func preRunAddCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

// runAddCommand prompts the user to select an external auth provider to connect their account to
func runAddCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Get the app selection and accompanying auth from the prompt
	// Note: Only installed apps will be shown in the prompt as installation is required for this command.
	//       This is consistent with the experience for other commands that require installation.
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Get the oauth2 provider keys
	providerAuths, err := clients.ApiInterface().AppsAuthExternalList(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		false, /*include_workflows flag to return workflow auth info*/
	)
	if err != nil {
		return err
	}
	authorizationList := providerAuths.Authorizations
	if len(authorizationList) > 0 && authorizationList[0].ExternalTokens != nil {
		clients.IO.PrintInfo(ctx, false,
			"\nThis command is used to only add an account.\n\nTo use an added account in a workflow steps, please use the %s command after this command.\n",
			style.Commandf("external-auth select-auth", false),
		)
	}

	// Show provider prompt and fetch the selection
	if providerFlag == "" {
		// Get oauth2 providerAuths
		providerAuth, err := providerSelectionFunc(ctx, clients, providerAuths)
		if err != nil {
			return err
		}
		providerFlag = providerAuth.ProviderKey
		if providerFlag == "" {
			return slackerror.New("Unable to get a provider selection")
		}
	}

	clientSecretExists := false
	for _, provider := range providerAuths.Authorizations {
		if provider.ProviderKey == providerFlag {
			clientSecretExists = provider.ClientSecretExists
			break
		}
	}

	if !clientSecretExists {
		command := style.Commandf("external-auth add-secret", false)
		return slackerror.New(fmt.Sprintf("Error: No client secret exists. Add one with %s", command))
	}

	authorizationUrl, err := clients.ApiInterface().AppsAuthExternalStart(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		providerFlag,
	)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "Redirecting to browser...")
	clients.Browser().OpenURL(authorizationUrl)

	return err
}
