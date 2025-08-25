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

package extension

import (
	"context"
	"strings"

	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewExtensionExecCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "exec [flags] <cmd> [args]",
		Hidden: true,
		Short:  "Run an extension",
		Long: strings.Join([]string{
			"Execute the extension with provided arguments",
		}, "\n"),
		Args: cobra.MinimumNArgs(0),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Run the \"glitch\" extension",
				Command: "extension exec glitch",
			},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return preRunExtensionExecCommandFunc(ctx, clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtensionExecCommandFunc(clients, cmd, args)
		},
	}

	return cmd
}

// preRunExtensionExecCommandFunc determines if the command is available.
func preRunExtensionExecCommandFunc(ctx context.Context, clients *shared.ClientFactory) error {
	if !clients.Config.WithExperimentOn(experiment.Extension) {
		return slackerror.New(slackerror.ErrMissingExperiment).
			WithRemediation("Run the command again with the \"--experiment extension\" flag.")
	}
	return nil
}

// runExtensionExecCommandFunc runs the provided extension.
func runExtensionExecCommandFunc(
	clients *shared.ClientFactory,
	cmd *cobra.Command,
	args []string,
) error {
	ctx := cmd.Context()
	switch {
	case len(args) == 0:
		return slackerror.New(slackerror.ErrMissingExtension)
	case args[0] == "glitch":
		clients.IO.PrintDebug(ctx, ":space_invader:")
		clients.IO.PrintInfo(ctx, false, "\x1b[?25l")
		return nil
	}
	opts := hooks.HookExecOpts{
		Hook: hooks.HookScript{
			Name:    "extension",
			Command: "slack-" + strings.Join(args, " "),
		},
		Stdin:  clients.IO.ReadIn(),
		Stdout: clients.IO.WriteOut(),
		Stderr: clients.IO.WriteErr(),
		Env: map[string]string{
			"SLACK_CLI_ALIAS": cmdutil.GetProcessName(),
		},
	}
	// The "default" protocol is used because a certain response is not expected of
	// scripts provided to the "extension" hook.
	//
	// The "message boundaries" protocol appends information to scripts using flags
	// which might cause some commands to error.
	//
	// The hook executor attached to the provided clients might use either protocol
	// so we instantiate the default here.
	shell := hooks.HookExecutorDefaultProtocol{
		IO: clients.IO,
	}
	_, err := shell.Execute(ctx, opts)
	if err != nil {
		return err
	}
	return nil
}
