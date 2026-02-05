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
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var allProvidersFlag bool

func NewRemoveCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove the saved tokens for a provider",
		Long: strings.Join([]string{
			"Remove tokens saved to external authentication providers of a workflow app.",
			"",
			"Existing tokens are only removed from your app, but are not revoked or deleted!",
			"Tokens must be invalidated using the provider's developer console or via APIs.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Remove a token from the selected provider",
				Command: "external-auth remove",
			},
			{
				Meaning: "Remove a token from the specified provider",
				Command: "external-auth remove -p github",
			},
			{
				Meaning: "Remove all tokens from the specified provider",
				Command: "external-auth remove --all -p github",
			},
			{
				Meaning: "Remove all tokens from all providers",
				Command: "external-auth remove --all",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunRemoveCommand(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemoveCommand(clients, cmd)
		},
	}

	// Flags
	cmd.PersistentFlags().BoolVarP(&allProvidersFlag, "all", "A", false, "remove tokens for all providers or the specified provider")
	cmd.PersistentFlags().StringVarP(&providerFlag, "provider", "p", "", "the external auth Provider Key to remove a token for")

	return cmd
}

// preRunRemoveCommand determines if the command is supported for a project and
// configures flags
func preRunRemoveCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

// runRemoveCommand prompts the user to select an external auth provider to remove their account from
func runRemoveCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Get the app selection and accompanying auth from the prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowAllEnvironments, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, style.Highlight("\nNote that this command will not revoke existing tokens, only remove them from Slack systems. You might be able to revoke them from a provider's dev console or APIs\n"))

	confirmMessage := "Are you sure you want to remove all tokens for this app relevant to the specified provider from your current team/org?"
	if allProvidersFlag {
		removeAllTokens := false

		if providerFlag == "" {
			confirmMessage := "Are you sure you want to remove all tokens for this app from your current team/org?"
			removeAllTokens, err = clients.IO.ConfirmPrompt(ctx, confirmMessage, false)
		} else {
			removeAllTokens, err = clients.IO.ConfirmPrompt(ctx, confirmMessage, false)
		}
		if err != nil {
			return err
		}

		if !removeAllTokens {
			clients.IO.PrintInfo(ctx, false, "Alright, we will not remove any tokens.")
			return nil
		}

		err = clients.API().AppsAuthExternalDelete(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
			providerFlag,
			"",
		)
		if err != nil {
			return err
		}

	} else if providerFlag != "" {
		removeAllTokens, err := clients.IO.ConfirmPrompt(ctx, confirmMessage, false)
		if err != nil {
			return err
		}
		if !removeAllTokens {
			clients.IO.PrintInfo(ctx, false, "Alright, we will not remove any tokens.")
			return nil
		}

		err = clients.API().AppsAuthExternalDelete(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
			providerFlag,
			"",
		)
		if err != nil {
			return err
		}
	} else {
		// Get oauth2 providerAuths
		providerAuths, err := clients.API().AppsAuthExternalList(
			ctx,
			selection.Auth.Token,
			selection.App.AppID,
			true,
		)
		if err != nil {
			return err
		}
		authorizationList := providerAuths.Authorizations
		if len(authorizationList) > 0 && authorizationList[0].ExternalTokens != nil {
			clients.IO.PrintInfo(ctx, false,
				"\nThis command will affect every workflow step that uses the removed account.\n\nPlease use the %s command to associate another account to each workflow after removing the account.\n",
				style.Commandf("external-auth select-auth", false),
			)
		}

		providerAuth, err := providerSelectionFunc(ctx, clients, providerAuths)
		if err != nil {
			return err
		}

		providerFlag = providerAuth.ProviderKey
		if providerFlag == "" {
			return slackerror.New("Unable to get a provider selection")
		}

		if len(providerAuth.ExternalTokens) > 0 {
			// Token selection prompt
			externalTokenInfo, err := tokenSelectionFunc(ctx, clients, providerAuth)
			if err != nil {
				return err
			}
			externalTokenArg := externalTokenInfo.ExternalTokenID
			if externalTokenArg == "" {
				return slackerror.New("Unable to get a provider selection")
			}
			err = clients.API().AppsAuthExternalDelete(
				ctx,
				selection.Auth.Token,
				selection.App.AppID,
				providerFlag,
				externalTokenArg,
			)
			if err != nil {
				return err
			}
		} else {
			err = clients.API().AppsAuthExternalDelete(
				ctx,
				selection.Auth.Token,
				selection.App.AppID,
				providerFlag,
				"",
			)
			if err != nil {
				return err
			}
		}
	}

	var text string
	if providerFlag == "" {
		text = fmt.Sprintf(
			"%s Token(s) removed for %s",
			style.Emoji("wastebasket"),
			style.Highlight(selection.App.AppID),
		)
	} else {
		text = fmt.Sprintf(
			"%s Token(s) removed for %s and %s",
			style.Emoji("wastebasket"),
			style.Highlight(selection.App.AppID),
			style.Highlight(providerFlag),
		)
	}

	clients.IO.PrintInfo(ctx, false, text)
	return nil
}
