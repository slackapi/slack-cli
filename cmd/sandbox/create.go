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
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

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
	templateID  int
	eventCode   string
	archiveTTL  string // TTL duration, e.g. 1d, 2h
	archiveDate string // explicit date yyyy-mm-dd
}

var createCmdFlags createFlags

func NewCreateCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [flags]",
		Short: "Create a developer sandbox",
		Long:  `Create a new Slack developer sandbox`,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "sandbox create --name test-box --password mypass", Meaning: "Create a sandbox named test-box"},
			{Command: "sandbox create --name test-box --password mypass --domain test-box --archive-ttl 1d", Meaning: "Create a temporary sandbox that will be archived in 1 day"},
			{Command: "sandbox create --name test-box --password mypass --domain test-box --archive-date 2025-12-31", Meaning: "Create a sandbox that will be archived on a specific date"},
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
	cmd.Flags().StringVar(&createCmdFlags.locale, "locale", "", "Locale (eg. en-us, languageCode-countryCode)")
	cmd.Flags().IntVar(&createCmdFlags.templateID, "template", 0, "Template ID for pre-defined data to preload")
	cmd.Flags().StringVar(&createCmdFlags.eventCode, "event-code", "", "Event code for the sandbox")
	cmd.Flags().StringVar(&createCmdFlags.archiveTTL, "archive-ttl", "", "Time-to-live duration; sandbox will be archived at end of day after this period (e.g., 2h, 1d, 7d)")
	cmd.Flags().StringVar(&createCmdFlags.archiveDate, "archive-date", "", "Explicit archive date in yyyy-mm-dd format. Cannot be used with --archive")

	// If one's developer account is managed by multiple Production Slack teams, one of those team IDs must be provided in the command
	cmd.Flags().StringVar(&createCmdFlags.owningOrgID, "owning-org-id", "", "Enterprise team ID that manages your developer account, if applicable")

	if err := cmd.MarkFlagRequired("name"); err != nil {
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
		var err error
		domain, err = domainFromName(createCmdFlags.name)
		if err != nil {
			return err
		}
	}

	if createCmdFlags.archiveTTL != "" && createCmdFlags.archiveDate != "" {
		return slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Cannot use both --archive-ttl and --archive-date").
			WithRemediation("Use only one: --archive-ttl for TTL (e.g., 3d) or --archive-date for a specific date (yyyy-mm-dd)")
	}

	archiveEpochDatetime := int64(0)
	if createCmdFlags.archiveTTL != "" {
		archiveEpochDatetime, err = getEpochFromTTL(createCmdFlags.archiveTTL)
		if err != nil {
			return err
		}
	} else if createCmdFlags.archiveDate != "" {
		archiveEpochDatetime, err = getEpochFromDate(createCmdFlags.archiveDate)
		if err != nil {
			return err
		}
	}

	teamID, sandboxURL, err := clients.API().CreateSandbox(ctx, auth.Token,
		createCmdFlags.name,
		domain,
		createCmdFlags.password,
		createCmdFlags.locale,
		createCmdFlags.owningOrgID,
		createCmdFlags.templateID,
		createCmdFlags.eventCode,
		archiveEpochDatetime,
	)
	if err != nil {
		return err
	}

	printCreateSuccess(cmd, clients, teamID, sandboxURL)

	return nil
}

// getEpochFromTTL parses a time-to-live string (e.g., "24h", "1d", "7d") and returns the Unix epoch
// when the sandbox will be archived. Supports Go duration format (h, m, s) and "Nd" for days.
// The value cannot exceed 6 months.
func getEpochFromTTL(ttl string) (int64, error) {
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
	return time.Now().Add(d).Unix(), nil
}

// getEpochFromDate parses a date in yyyy-mm-dd format and returns the Unix epoch at start of that day (UTC).
func getEpochFromDate(dateStr string) (int64, error) {
	dateFormat := "2006-01-02"
	t, err := time.ParseInLocation(dateFormat, dateStr, time.UTC)
	if err != nil {
		return 0, slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Invalid archive date: %q", dateStr).
			WithRemediation("Use yyyy-mm-dd format (e.g., 2025-12-31)")
	}
	return t.Unix(), nil
}

// domainFromName derives domain-safe text from the name of the sandbox (lowercase, alphanumeric + hyphens).
func domainFromName(name string) (string, error) {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	var domain []byte
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			domain = append(domain, byte(r))
		}
	}
	domain = []byte(strings.Trim(string(domain), "-"))
	if len(domain) == 0 {
		return "", slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Provide a valid domain name with the --domain flag")
	}
	return string(domain), nil
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
	clients.IO.PrintInfo(ctx, false, "Manage this sandbox from the CLI or visit\n%s", style.Secondary("https://api.slack.com/developer-program/sandboxes"))
}
