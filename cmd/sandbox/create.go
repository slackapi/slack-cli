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

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

type createFlags struct {
	name        string
	domain      string
	password    string
	locale      string
	owningOrgID string
	template    string
	ttl         string
	archiveDate int64
	autoLogin   bool
	output      string
	token       string
}

var createCmdFlags createFlags

func NewCreateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags]",
		Short: "Create a new sandbox",
		Long: `Create a new Slack developer sandbox.

Provisions a new sandbox. Domain is derived from org name if --domain is not provided.
Use --auto-login to open the sandbox URL in your browser after creation.`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox create --name test-box", Meaning: "Create a sandbox named test-box"},
			{Command: "sandbox create --name test-box --password mypass --owning-org-id E12345", Meaning: "Create a sandbox with login password and owning org"},
			{Command: "sandbox create --name test-box --domain test-box --ttl 24h --output json", Meaning: "Create an ephemeral sandbox for CI/CD with JSON output"},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return requireSandboxExperiment(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&createCmdFlags.name, "name", "", "Organization name for the new sandbox")
	cmd.Flags().StringVar(&createCmdFlags.domain, "domain", "", "Team domain (e.g., pizzaknifefight). If not provided, derived from org name")
	cmd.Flags().StringVar(&createCmdFlags.password, "password", "", "Password used to log into the sandbox")
	cmd.Flags().StringVar(&createCmdFlags.owningOrgID, "owning-org-id", "", "Enterprise team ID that manages your developer account, if applicable")
	cmd.Flags().StringVar(&createCmdFlags.template, "template", "", "Template ID for pre-defined data to preload")
	cmd.Flags().StringVar(&createCmdFlags.ttl, "ttl", "", "Time-to-live duration for ephemeral sandboxes (e.g., 2h, 1d, 7d)")
	cmd.Flags().Int64Var(&createCmdFlags.archiveDate, "archive-date", 0, "When the sandbox will be archived, as Unix epoch seconds")
	cmd.Flags().BoolVar(&createCmdFlags.autoLogin, "auto-login", false, "Open the sandbox URL in browser after creation")
	cmd.Flags().StringVar(&createCmdFlags.output, "output", "text", "Output format: json, text")
	cmd.Flags().StringVar(&createCmdFlags.token, "token", "", "Service account token for CI/CD authentication")

	return cmd
}

func runCreateCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	orgName := createCmdFlags.name
	if orgName == "" {
		return slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Organization name is required").
			WithRemediation("Use the --name flag")
	}

	token, err := getSandboxToken(ctx, clients, createCmdFlags.token)
	if err != nil {
		return err
	}

	domain := createCmdFlags.domain
	if domain == "" {
		domain = slugFromOrgName(orgName)
	}

	result, err := clients.API().CreateSandbox(ctx, token,
		orgName,
		domain,
		createCmdFlags.password,
		createCmdFlags.locale,
		createCmdFlags.owningOrgID,
		createCmdFlags.template,
		"", // eventCode
		createCmdFlags.archiveDate,
	)
	if err != nil {
		return err
	}

	switch createCmdFlags.output {
	case "json":
		encoder := json.NewEncoder(clients.IO.WriteOut())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return err
		}
	default:
		printCreateSuccess(cmd, clients, result)
	}

	if createCmdFlags.autoLogin && result.URL != "" {
		clients.Browser().OpenURL(result.URL)
	}

	return nil
}

// slugFromOrgName derives a domain-safe slug from org name (lowercase, alphanumeric + hyphens).
func slugFromOrgName(name string) string {
	var b []byte
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b = append(b, byte(r))
		} else if r >= 'A' && r <= 'Z' {
			b = append(b, byte(r+32))
		} else if r == ' ' || r == '-' || r == '_' {
			if len(b) > 0 && b[len(b)-1] != '-' {
				b = append(b, '-')
			}
		}
	}
	// Trim leading/trailing hyphens
	for len(b) > 0 && b[0] == '-' {
		b = b[1:]
	}
	for len(b) > 0 && b[len(b)-1] == '-' {
		b = b[:len(b)-1]
	}
	if len(b) == 0 {
		return "sandbox"
	}
	return string(b)
}

func printCreateSuccess(cmd *cobra.Command, clients *shared.ClientFactory, result types.CreateSandboxResult) {
	ctx := cmd.Context()
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "beach_with_umbrella",
		Text:  "Sandbox Created",
		Secondary: []string{
			fmt.Sprintf("Team ID: %s", result.TeamID),
			fmt.Sprintf("User ID: %s", result.UserID),
			fmt.Sprintf("URL: %s", result.URL),
		},
	}))
}
