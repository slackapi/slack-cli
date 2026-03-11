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
	"strconv"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/shared"
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
	demoIDs     []string
	eventCode   string
	ttl         string
	output      string
}

var createCmdFlags createFlags

func NewCreateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags]",
		Short: "Create a developer sandbox",
		Long:  `Create a new Slack developer sandbox`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox create --name test-box --password mypass", Meaning: "Create a sandbox named test-box"},
			{Command: "sandbox create --name test-box --password mypass --domain test-box --ttl 1d", Meaning: "Create a temporary sandbox that will be archived in 1 day"},
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
	cmd.Flags().StringVar(&createCmdFlags.template, "template", "", "Template ID for pre-defined data to preload")
	cmd.Flags().StringSliceVar(&createCmdFlags.demoIDs, "demo-ids", nil, "Demo IDs to preload in the sandbox")
	cmd.Flags().StringVar(&createCmdFlags.eventCode, "event-code", "", "Event code for the sandbox")
	cmd.Flags().StringVar(&createCmdFlags.ttl, "ttl", "", "Time-to-live duration; sandbox will be archived after this period (e.g., 2h, 1d, 7d)")
	cmd.Flags().StringVar(&createCmdFlags.output, "output", "text", "Output format: json, text")

	// If one's developer account is managed by multiple Production Slack teams, one of those team IDs must be provided in the command
	cmd.Flags().StringVar(&createCmdFlags.owningOrgID, "owning-org-id", "", "Enterprise team ID that manages your developer account, if applicable")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("domain"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("password"); err != nil {
		panic(err)
	}

	return cmd
}

func runCreateCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	auth, err := getSandboxAuth(ctx, clients)
	if err != nil {
		return err
	}

	domain := createCmdFlags.domain
	if domain == "" {
		domain = slugFromsandboxName(createCmdFlags.name)
	}

	archiveDate, err := ttlToArchiveDate(createCmdFlags.ttl)
	if err != nil {
		return err
	}

	teamID, sandboxURL, err := clients.API().CreateSandbox(ctx, auth.Token,
		createCmdFlags.name,
		domain,
		createCmdFlags.password,
		createCmdFlags.locale,
		createCmdFlags.owningOrgID,
		createCmdFlags.template,
		createCmdFlags.eventCode,
		archiveDate,
	)
	if err != nil {
		return err
	}

	switch createCmdFlags.output {
	case "json":
		encoder := json.NewEncoder(clients.IO.WriteOut())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(map[string]string{"team_id": teamID, "url": sandboxURL}); err != nil {
			return err
		}
	default:
		printCreateSuccess(cmd, clients, teamID, sandboxURL)
	}

	return nil
}

const maxTTL = 180 * 24 * time.Hour // 6 months

// ttlToArchiveDate parses a TTL string (e.g., "24h", "1d", "7d") and returns the Unix epoch
// when the sandbox will be archived. Returns 0 if ttl is empty (no archiving). Supports
// Go duration format (h, m, s) and "Nd" for days. TTL cannot exceed 6 months.
func ttlToArchiveDate(ttl string) (int64, error) {
	if ttl == "" {
		return 0, nil
	}
	var d time.Duration
	if strings.HasSuffix(strings.ToLower(ttl), "d") {
		numStr := strings.TrimSuffix(strings.ToLower(ttl), "d")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, slackerror.New(slackerror.ErrInvalidArguments).
				WithMessage("Invalid TTL: %q", ttl).
				WithRemediation("Use a duration like 2h, 1d, or 7d")
		}
		d = time.Duration(n) * 24 * time.Hour
	} else {
		var err error
		d, err = time.ParseDuration(ttl)
		if err != nil {
			return 0, slackerror.New(slackerror.ErrInvalidArguments).
				WithMessage("Invalid TTL: %q", ttl).
				WithRemediation("Use a duration like 2h, 1d, or 7d")
		}
	}
	if d > maxTTL {
		return 0, slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("TTL cannot exceed 6 months").
			WithRemediation("Use a shorter duration (e.g., 2h, 1d, 7d)")
	}
	return time.Now().Add(d).Unix(), nil
}

// slugFromsandboxName derives a domain-safe slug from org name (lowercase, alphanumeric + hyphens).
func slugFromsandboxName(name string) string {
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

func printCreateSuccess(cmd *cobra.Command, clients *shared.ClientFactory, teamID, url string) {
	ctx := cmd.Context()
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "beach_with_umbrella",
		Text:  " Sandbox Created",
		Secondary: []string{
			fmt.Sprintf("Team ID: %s", teamID),
			fmt.Sprintf("URL: %s", url),
		},
	}))
}
