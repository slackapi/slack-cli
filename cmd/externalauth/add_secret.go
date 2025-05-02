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
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var secretFlag string

func NewAddClientSecretCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-secret [flags]",
		Short: "Add the client secret for a provider",
		Long: strings.Join([]string{
			"Add the client secret for an external provider of a workflow app.",
			"",
			"This secret will be used when initiating the OAuth2 flow.",
			"",
			"This command is supported for apps deployed to Slack managed infrastructure but",
			"other apps can attempt to run the command with the --force flag.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Input the client secret for an app and provider",
				Command: "external-auth add-secret",
			},
			{
				Meaning: "Set the client secret for an app and provider",
				Command: "external-auth add-secret -p github -x ghp_token",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunAddClientSecretCommand(ctx, clients, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddClientSecretCommand(clients, cmd)
		},
	}

	// Flags
	cmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "the external auth Provider Key to add a secret to")
	cmd.Flags().StringVarP(&secretFlag, "secret", "x", "", "external auth client secret for the provider")
	return cmd
}

// preRunAddClientSecretCommand determines if the command is supported for a
// project and configures flags
func preRunAddClientSecretCommand(ctx context.Context, clients *shared.ClientFactory, cmd *cobra.Command) error {
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

// runAddClientSecretCommand adds a client secret to an authentication provider
func runAddClientSecretCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()

	// Get the app selection and accompanying auth from the prompt
	selection, err := appSelectPromptFunc(ctx, clients, prompts.ShowInstalledAppsOnly)
	if err != nil {
		return err
	}

	// Get the oauth2 provider keys
	providerAuths, err := clients.APIInterface().AppsAuthExternalList(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		false, /*include_workflows flag to return workflow auth info*/
	)
	if err != nil {
		return err
	}

	providerAuth, err := providerSelectionFunc(ctx, clients, providerAuths)
	if err != nil {
		return err
	}
	providerFlag = providerAuth.ProviderKey
	if providerFlag == "" {
		return slackerror.New("Unable to get a provider selection")
	}

	var clientSecret string
	if response, err := clients.IO.PasswordPrompt(ctx, "Enter the client secret", iostreams.PasswordPromptConfig{
		Flag:     clients.Config.Flags.Lookup("secret"),
		Required: true,
	}); err != nil {
		return err
	} else {
		clientSecret = response.Value
	}

	err = clients.APIInterface().AppsAuthExternalClientSecretAdd(
		ctx,
		selection.Auth.Token,
		selection.App.AppID,
		providerFlag,
		clientSecret,
	)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, style.Sectionf(style.TextSection{
		Emoji: "sparkles",
		Text: fmt.Sprintf(
			"Successfully added external auth client secret for %s",
			providerFlag,
		),
	}))
	return nil
}
