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

	"github.com/slackapi/slack-cli/internal/iostreams"
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
	eventCode   string
	archiveTTL  string // TTL duration, e.g. 1d, 2w, 3mo
	archiveDate string // explicit date yyyy-mm-dd
	partner     bool
}

var createCmdFlags createFlags

// templateNameToID maps user-friendly template names to integer IDs
var templateNameToID = map[string]int{
	"default": 1, // The default template
	"empty":   0, // The sandbox will be empty if the template param is not set
}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateCommand(cmd, clients)
		},
	}

	cmd.Flags().StringVar(&createCmdFlags.name, "name", "", "Organization name for the new sandbox")
	cmd.Flags().StringVar(&createCmdFlags.domain, "domain", "", "Team domain. Derived from org name if not provided")
	cmd.Flags().StringVar(&createCmdFlags.password, "password", "", "Password used to log into the sandbox")
	cmd.Flags().StringVar(&createCmdFlags.locale, "locale", "", "Locale (eg. en-us, languageCode-countryCode)")
	cmd.Flags().StringVar(&createCmdFlags.template, "template", "", "Template with sample data to apply to the sandbox (options: default, empty)")
	cmd.Flags().StringVar(&createCmdFlags.eventCode, "event-code", "", "Event code for the sandbox")
	cmd.Flags().StringVar(&createCmdFlags.archiveTTL, "archive-ttl", "", "Time-to-live duration (eg. 1d, 2w, 3mo). Cannot be used with --archive-date")
	cmd.Flags().StringVar(&createCmdFlags.archiveDate, "archive-date", "", "Explicit archive date in yyyy-mm-dd format. Cannot be used with --archive-ttl")
	cmd.Flags().BoolVar(&createCmdFlags.partner, "partner", false, "Developers who are part of the Partner program can create partner sandboxes")

	// If one's developer account is managed by multiple Production Slack teams, one of those team IDs must be provided in the command
	cmd.Flags().StringVar(&createCmdFlags.owningOrgID, "owning-org-id", "", "Enterprise team ID that manages your developer account, if applicable")

	return cmd
}

func runCreateCommand(cmd *cobra.Command, clients *shared.ClientFactory) error {
	ctx := cmd.Context()

	auth, err := getSandboxAuth(ctx, clients)
	if err != nil {
		return err
	}

	name := createCmdFlags.name
	if name == "" {
		name, err = clients.IO.InputPrompt(
			ctx,
			"Enter a name for the sandbox",
			iostreams.InputPromptConfig{
				Required: true,
			},
		)
		if err != nil {
			return err
		}
	}

	password := createCmdFlags.password
	if password == "" {
		password, err = clients.IO.InputPrompt(
			ctx,
			"Enter a password for the sandbox",
			iostreams.InputPromptConfig{
				Required: true,
			},
		)
		if err != nil {
			return err
		}
	}

	domain := createCmdFlags.domain
	if domain == "" {
		var err error
		domain, err = domainFromName(name)
		if err != nil {
			return err
		}
	}

	if createCmdFlags.archiveTTL != "" && createCmdFlags.archiveDate != "" {
		return slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Cannot use both --archive-ttl and --archive-date")
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

	templateID, err := getTemplateID(createCmdFlags.template)
	if err != nil {
		return err
	}

	teamID, sandboxURL, err := clients.API().CreateSandbox(ctx, auth.Token,
		name,
		domain,
		password,
		createCmdFlags.locale,
		createCmdFlags.owningOrgID,
		templateID,
		createCmdFlags.eventCode,
		archiveEpochDatetime,
		createCmdFlags.partner,
	)
	if err != nil {
		return err
	}

	printCreateSuccess(cmd, clients, teamID, sandboxURL)

	return nil
}

// getEpochFromTTL parses a time-to-live string (e.g., "1d", "2w", "3mo") and returns the Unix epoch
// when the sandbox will be archived. Supports days (d), weeks (w), and months (mo).
func getEpochFromTTL(ttl string) (int64, error) {
	lower := strings.TrimSpace(strings.ToLower(ttl))
	if lower == "" {
		return 0, slackerror.New(slackerror.ErrInvalidSandboxArchiveTTL)
	}

	var target time.Time
	now := time.Now()

	switch {
	case strings.HasSuffix(lower, "d"):
		n, err := strconv.Atoi(strings.TrimSuffix(lower, "d"))
		if err != nil || n < 1 {
			return 0, slackerror.New(slackerror.ErrInvalidSandboxArchiveTTL)
		}
		target = now.AddDate(0, 0, n)
	case strings.HasSuffix(lower, "w"):
		n, err := strconv.Atoi(strings.TrimSuffix(lower, "w"))
		if err != nil || n < 1 {
			return 0, slackerror.New(slackerror.ErrInvalidSandboxArchiveTTL)
		}
		target = now.AddDate(0, 0, n*7)
	case strings.HasSuffix(lower, "mo"):
		n, err := strconv.Atoi(strings.TrimSuffix(lower, "mo"))
		if err != nil || n < 1 {
			return 0, slackerror.New(slackerror.ErrInvalidSandboxArchiveTTL)
		}
		target = now.AddDate(0, n, 0)
	default:
		return 0, slackerror.New(slackerror.ErrInvalidSandboxArchiveTTL)
	}

	return target.Unix(), nil
}

// getEpochFromDate parses a date in yyyy-mm-dd format and returns the Unix epoch at the start of that day (UTC)
func getEpochFromDate(dateStr string) (int64, error) {
	dateFormat := "2006-01-02"
	t, err := time.ParseInLocation(dateFormat, dateStr, time.UTC)
	if err != nil {
		return 0, slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Invalid archive date: %q", dateStr).
			WithRemediation("Use yyyy-mm-dd format")
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if t.Before(today) {
		return 0, slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("Archive date must be in the future")
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

// getTemplateID converts a template string to an integer ID
func getTemplateID(template string) (int, error) {
	if template == "" {
		return 0, nil
	}
	key := strings.ToLower(strings.TrimSpace(template))
	// If the provided string is present in the map, return the ID
	if id, ok := templateNameToID[key]; ok {
		return id, nil
	}
	// We also accept an integer passed directly via the flag
	if id, err := strconv.Atoi(key); err == nil {
		return id, nil
	}
	return 0, slackerror.New(slackerror.ErrInvalidSandboxTemplateID).
		WithMessage("Invalid template: %q", template)
}

func printCreateSuccess(cmd *cobra.Command, clients *shared.ClientFactory, teamID, url string) {
	ctx := cmd.Context()
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "beach_with_umbrella",
		Text:  "Sandbox Created",
		Secondary: []string{
			fmt.Sprintf("Team ID: %s", teamID),
			fmt.Sprintf("URL: %s", url),
		},
	}))
	clients.IO.PrintInfo(ctx, false, "Manage this sandbox from the CLI or visit\n%s", style.Secondary("https://api.slack.com/developer-program/sandboxes"))
}
