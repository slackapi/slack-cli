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

package config

import (
	"fmt"
	"os"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// SetFlags saves the provided command flags to the config
func (c *Config) SetFlags(cmd *cobra.Command) {
	c.Flags = cmd.Flags()
}

// InitializeGlobalFlags configures flags and creates links from cmd to config
func (c *Config) InitializeGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&c.APIHostFlag, "apihost", "", "Slack API host")
	cmd.PersistentFlags().StringVarP(&c.AppFlag, "app", "a", "", "use a specific app ID or environment")
	cmd.PersistentFlags().StringVarP(&c.ConfigDirFlag, "config-dir", "", "", "use a custom path for system config directory")
	cmd.PersistentFlags().BoolVarP(&c.DeprecatedDevAppFlag, "local-run", "l", false, "use the local run app created by the `run` command") // deprecated
	cmd.PersistentFlags().BoolVarP(&c.DeprecatedDevFlag, "dev", "d", false, "use dev apis")                                                // Can be removed after v0.25.0
	cmd.PersistentFlags().StringVarP(&c.DeprecatedWorkspaceFlag, "workspace", "", "", "select workspace or organization by domain name or team ID")
	cmd.PersistentFlags().StringSliceVarP(&c.ExperimentsFlag, "experiment", "e", nil, "use the experiment(s) in the command")
	cmd.PersistentFlags().BoolVarP(&c.ForceFlag, "force", "f", false, "ignore warnings and continue executing command")
	cmd.PersistentFlags().BoolVarP(&c.NoColor, "no-color", "", false, "remove styles and formatting from outputs")
	cmd.PersistentFlags().BoolVarP(&c.SkipUpdateFlag, "skip-update", "s", false, "skip checking for latest version of CLI")
	cmd.PersistentFlags().BoolVarP(&c.SlackDevFlag, "slackdev", "", false, "shorthand for --apihost=https://dev.slack.com")
	cmd.PersistentFlags().StringVarP(&c.RuntimeFlag, "runtime", "r", "", "the project's runtime language:\n  deno (default), deno1.1, deno1.x, etc")
	// TODO - next semver MAJOR can consider a new shorthand flag, right now -t and -T are used by other commands
	cmd.PersistentFlags().StringVarP(&c.TeamFlag, "team", "w", "", "select workspace or organization by team name or ID")
	cmd.PersistentFlags().StringVarP(&c.TokenFlag, "token", "", "", "set the access token associated with a team")
	cmd.PersistentFlags().BoolVarP(&c.DebugEnabled, "verbose", "v", false, "print debug logging and additional info")

	cmd.PersistentFlags().Lookup("apihost").Hidden = true
	cmd.PersistentFlags().Lookup("dev").Hidden = true
	cmd.PersistentFlags().Lookup("local-run").Hidden = true
	cmd.PersistentFlags().Lookup("runtime").Hidden = true
	cmd.PersistentFlags().Lookup("slackdev").Hidden = true
	cmd.PersistentFlags().Lookup("workspace").Hidden = true

	for _, arg := range os.Args {
		if arg == "--verbose" || arg == "-v" {
			cmd.PersistentFlags().Lookup("apihost").Hidden = false
			cmd.PersistentFlags().Lookup("runtime").Hidden = false
			cmd.PersistentFlags().Lookup("slackdev").Hidden = false
		}
	}
}

// DeprecatedFlagSubstitutions displays warnings when using deprecated flags and
// provides alternatives when possible
func (c *Config) DeprecatedFlagSubstitutions(cmd *cobra.Command) error {
	if c.DeprecatedDevFlag {
		deprecationWarning := style.TextSection{
			Emoji: "construction",
			Text:  "Deprecation of --dev",
			Secondary: []string{
				`--dev flag has been removed`,
				`--slackdev flag can now be used as shorthand for --apihost="https://dev.slack.com"`,
				fmt.Sprintf("Continuing execution with %s", style.Highlight("--slackdev")),
			},
		}
		cmd.PrintErr(style.Sectionf(deprecationWarning))
		c.SlackDevFlag = true
	}

	if c.DeprecatedDevAppFlag {
		deprecationWarning := style.TextSection{
			Emoji: "construction",
			Text:  "Deprecation of --local-run",
			Secondary: []string{
				`The --local-run flag has been removed`,
				fmt.Sprintf("Specify a local app with %s", style.Highlight("--app local")),
			},
		}
		if c.AppFlag == "" || c.AppFlag == "local" {
			deprecationWarning.Secondary = append(deprecationWarning.Secondary, fmt.Sprintf(
				"Continuing execution with %s",
				style.Highlight("--app local")),
			)
			cmd.PrintErr(style.Sectionf(deprecationWarning))
			c.AppFlag = "local"
		} else {
			cmd.PrintErr(style.Sectionf(deprecationWarning))
			return slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage("Cannot use both the --local-run and --app flags")
		}
	}

	if c.DeprecatedWorkspaceFlag != "" {
		deprecationWarning := style.TextSection{
			Emoji: "construction",
			Text:  "Deprecation of --workspace",
			Secondary: []string{
				"The --workspace flag has been removed",
				fmt.Sprintf("Specify a Slack workspace or organization with %s",
					style.Highlight("--team <domain|id>")),
			},
		}
		if c.TeamFlag == "" || c.TeamFlag == c.DeprecatedWorkspaceFlag {
			deprecationWarning.Secondary = append(deprecationWarning.Secondary, fmt.Sprintf(
				"Continuing execution with %s",
				style.Highlight(fmt.Sprintf("--team %s", c.DeprecatedWorkspaceFlag))),
			)
			cmd.PrintErr(style.Sectionf(deprecationWarning))
			c.TeamFlag = c.DeprecatedWorkspaceFlag
		} else {
			cmd.PrintErr(style.Sectionf(deprecationWarning))
			return slackerror.New(slackerror.ErrMismatchedFlags).
				WithMessage("Cannot use both the --workspace and --team flags")
		}
	}
	return nil
}
