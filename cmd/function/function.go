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

package function

import (
	"context"
	"strings"

	"github.com/slackapi/slack-cli/internal/prompts"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

var functionFlag string

var appSelectPromptFunc = prompts.AppSelectPrompt

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "function <subcommand> [flags]",
		Short: "Manage the functions of an app",
		Long: strings.Join([]string{
			"Functions are pieces of logic that complete the puzzle of workflows in Workflow",
			"Builder. Whatever that puzzle might be.",
			"",
			"Inspect and configure the custom functions included in an app with this command.",
			"Functions can be added as a step in Workflow Builder and shared among teammates.",
			"",
			`Learn more about functions: {{LinkText "https://docs.slack.dev/tools/deno-slack-sdk/guides/creating-functions"}}`,
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "function distribute", Meaning: "Select a function and choose distribution options"},
			{Command: "function distribute --name callback_id --everyone", Meaning: "Distribute a function to everyone in a workspace"},
			{Command: "function distribute --info", Meaning: "Lookup the distribution information for a function"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewDistributeCommand(clients))
	return cmd
}

// GetAppFunctions lists a user's custom functions from the app manifest
func GetAppFunctions(ctx context.Context, clients *shared.ClientFactory, app types.App, auth types.SlackAuth) ([]types.Function, error) {
	slackManifest, err := clients.AppClient().Manifest.GetManifestRemote(ctx, auth.Token, app.AppID)
	if err != nil {
		return nil, err
	}

	functions := []types.Function{}
	for callbackID, f := range slackManifest.Functions {
		functions = append(functions, types.Function{CallbackID: callbackID, Title: f.Title, Description: f.Description})
	}

	return functions, nil
}
