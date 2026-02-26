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

package sandbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type listFlags struct {
	output string
	filter string
	token  string
}

var listCmdFlags listFlags

func NewListCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List your Developer Sandboxes",
		Long:  `List all of your active or archived sandboxes.`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox list", Meaning: "List all sandboxes"},
			{Command: "sandbox list --filter active --output json", Meaning: "List active sandboxes as JSON"},
		}),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&listCmdFlags.output, "output", "table", "Output format: table, json, text")
	cmd.Flags().StringVar(&listCmdFlags.filter, "filter", "", "Filter by status: active, archived")
	cmd.Flags().StringVar(&listCmdFlags.token, "token", "", "Service account token for CI/CD authentication")

	return cmd
}

func runListCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	token, err := getSandboxToken(ctx, clients, listCmdFlags.token)
	if err != nil {
		return err
	}

	sandboxes, err := clients.API().ListSandboxes(ctx, token, listCmdFlags.filter)
	if err != nil {
		return err
	}

	switch listCmdFlags.output {
	case "json":
		encoder := json.NewEncoder(clients.IO.WriteOut())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(sandboxes); err != nil {
			return err
		}
	case "text":
		printSandboxesText(cmd, clients, sandboxes)
	default:
		printSandboxesTable(cmd, clients, sandboxes)
	}

	return nil
}

func printSandboxesTable(cmd *cobra.Command, clients *shared.ClientFactory, sandboxes []types.Sandbox) {
	ctx := cmd.Context()

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "beach_with_umbrella",
		Text:  "Developer Sandboxes",
	}))

	if len(sandboxes) == 0 {
		clients.IO.PrintInfo(ctx, false, "%s\n", style.Secondary("No sandboxes found. Create one with `slack sandbox create --name <name>`"))
		return
	}

	timeFormat := "2006-01-02 15:04"
	for _, s := range sandboxes {
		cmd.Printf("  %s (%s)\n", style.Bold(s.SandboxName), s.SandboxTeamID)
		if s.SandboxDomain != "" {
			cmd.Printf("    %s\n", style.Secondary(fmt.Sprintf("URL: https://%s.slack.com", s.SandboxDomain)))
		}
		if s.Status != "" {
			cmd.Printf("    %s\n", style.Secondary(fmt.Sprintf("Status: %s", s.Status)))
		}
		if s.DateCreated > 0 {
			cmd.Printf("    %s\n", style.Secondary(fmt.Sprintf("Created: %s", time.Unix(s.DateCreated, 0).Format(timeFormat))))
		}
		if s.DateArchived > 0 {
			cmd.Printf("    %s\n", style.Secondary(fmt.Sprintf("Archived: %s", time.Unix(s.DateArchived, 0).Format(timeFormat))))
		}
		cmd.Println()
	}
}

func printSandboxesText(cmd *cobra.Command, clients *shared.ClientFactory, sandboxes []types.Sandbox) {
	ctx := cmd.Context()

	if len(sandboxes) == 0 {
		clients.IO.PrintInfo(ctx, false, "%s\n", style.Secondary("No sandboxes found."))
		return
	}

	for _, s := range sandboxes {
		url := ""
		if s.SandboxDomain != "" {
			url = fmt.Sprintf("https://%s.slack.com", s.SandboxDomain)
		}
		clients.IO.PrintInfo(ctx, false, "%s %s %s\n", s.SandboxTeamID, s.SandboxName, url)
	}
}
