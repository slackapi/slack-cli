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

package help

import (
	"bytes"
	"fmt"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// HelpFunc prepares the templated output of any help command
func HelpFunc(
	clients *shared.ClientFactory,
	aliases map[string]string,
) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		style.ToggleStyles(clients.IO.IsTTY() && !clients.Config.NoColor)
		if help, _ := clients.Config.Flags.GetBool("help"); help {
			clients.Config.LoadExperiments(ctx, clients.IO.PrintDebug)
		}
		style.ToggleCharm(clients.Config.WithExperimentOn(experiment.Charm))
		experiments := []string{}
		for _, exp := range experiment.AllExperiments {
			if clients.Config.WithExperimentOn(exp) {
				experiments = append(experiments, fmt.Sprintf("%s ENABLED", string(exp)))
			} else {
				experiments = append(experiments, fmt.Sprintf("%s DISABLED", string(exp)))
			}
		}
		// Also show any non-standard experiments that are active but not in AllExperiments
		for _, exp := range clients.Config.GetExperiments() {
			if !experiment.Includes(exp) {
				invalidExperiment := fmt.Sprintf("%s (invalid)", string(exp))
				experiments = append(experiments, style.Secondary(invalidExperiment))
			}
		}
		data := style.TemplateData{
			"Aliases":     aliases,
			"Experiments": experiments,
		}
		PrintHelpTemplate(cmd, data)
	}
}

// PrintHelpTemplate displays the help message for a command with optional data
//
// Note: The cmd.Long text is formatted with templates and data before the rest
func PrintHelpTemplate(cmd *cobra.Command, data style.TemplateData) {
	type templateInfo struct {
		*cobra.Command
		Data map[string]interface{}
	}
	cmdLongF := bytes.Buffer{}
	err := style.PrintTemplate(&cmdLongF, cmd.Long, templateInfo{Data: data})
	if err != nil {
		cmd.PrintErrln(err)
	}
	cmd.Long = cmdLongF.String()
	tmpl := legacyHelpTemplate
	if style.IsCharmEnabled() {
		tmpl = charmHelpTemplate
	}
	err = style.PrintTemplate(cmd.OutOrStdout(), tmpl, templateInfo{cmd, data})
	if err != nil {
		cmd.PrintErrln(err)
	}
}

// ════════════════════════════════════════════════════════════════════════════════
// Charm help template — lipgloss styling
// ════════════════════════════════════════════════════════════════════════════════

const charmHelpTemplate string = `{{.Long | ToDescription}}

{{Header "Usage"}}{{if .Runnable}}
  {{ToPrompt "$ "}}{{ToCommandText .UseLine}}{{end}}{{if gt (len .Aliases) 0}}

{{Header "Aliases"}}
  {{.NameAndAliases | ToCommandText}}{{end}}{{if .HasAvailableSubCommands}}

{{if eq .Name (GetProcessName)}}{{Header "Commands"}}{{range .Commands}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}
  {{.Name | ToGroupName }}{{range .Commands}}{{if (not .Hidden)}}
    {{rpad .Name .NamePadding | ToCommandText}} {{.Short | ToDescription}}{{end}}{{end}}{{end}}{{end}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}{{range .Commands}}{{if and (not .HasAvailableSubCommands) (not .Hidden)}}{{if not (IsAlias .Name $.Data.Aliases)}}
  {{(rpad .Name .NamePadding) | ToGroupName }}{{.Short | ToDescription}}{{end}}{{end}}{{end}}{{end}}{{else}}{{Header "Subcommands"}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}{{range .Commands}}{{if not .HasAvailableSubCommands}}
  {{(rpad .Name .NamePadding) | ToCommandText }} {{.Short | ToDescription}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{Header "Flags"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces | ToFlags}}{{end}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}{{if or (HasAliasSubcommands .Name .Data.Aliases) (eq .Name (GetProcessName))}}

{{Header "Global aliases"}}{{range .Commands}}{{if and (IsAlias .Name $.Data.Aliases) (not .Hidden)}}
  {{(rpad .Name .NamePadding) | ToGroupName }} {{rpad (AliasParent .Name $.Data.Aliases) AliasPadding | ToAliasParent}} {{ToPrompt "❱"}} {{.Name | ToGroupName}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableInheritedFlags}}

{{Header "Global flags"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces | ToFlags}}{{end}}{{if .HasExample}}

{{Header "Example"}}
{{ Examples .Example}}{{end}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}

{{Header "Experiments"}}
{{ Experiments .Data.Experiments }}

{{Header "Additional help"}}
  {{ToSecondary "For more information about a specific command, run:"}}
  {{ToPrompt "$ "}}{{ToCommandText .CommandPath}}{{if eq .Name (GetProcessName)}}{{ToCommandText " <command>"}}{{end}}{{ToCommandText " <subcommand> --help"}}

  {{ToSecondary "For guides and documentation, head over to "}}{{LinkText "https://docs.slack.dev/tools/slack-cli"}}{{end}}

`

// ════════════════════════════════════════════════════════════════════════════════
// DEPRECATED: Legacy help template — aurora styling
//
// Delete this entire block when the charm experiment is permanently enabled.
// ════════════════════════════════════════════════════════════════════════════════

const legacyHelpTemplate string = `{{.Long}}

{{Header "Usage"}}{{if .Runnable}}
  $ {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

{{Header "Aliases"}}
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

{{if eq .Name (GetProcessName)}}{{Header "Commands"}}{{range .Commands}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}
  {{.Name | ToCommandText }}{{range .Commands}}{{if (not .Hidden)}}
    {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}{{range .Commands}}{{if and (not .HasAvailableSubCommands) (not .Hidden)}}{{if not (IsAlias .Name $.Data.Aliases)}}
  {{(rpad .Name .NamePadding) | ToCommandText }}{{.Short}}{{end}}{{end}}{{end}}{{end}}{{else}}{{Header "Subcommands"}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}{{range .Commands}}{{if not .HasAvailableSubCommands}}
  {{(rpad .Name .NamePadding) | ToCommandText }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{Header "Flags"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}{{if or (HasAliasSubcommands .Name .Data.Aliases) (eq .Name (GetProcessName))}}

{{Header "Global aliases"}}{{range .Commands}}{{if and (IsAlias .Name $.Data.Aliases) (not .Hidden)}}
  {{(rpad .Name .NamePadding) | ToBold }} {{rpad (AliasParent .Name $.Data.Aliases) AliasPadding}} > {{.Name}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableInheritedFlags}}

{{Header "Global flags"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

{{Header "Example"}}
{{ Examples .Example}}{{end}}{{if and (.HasAvailableSubCommands) (not .Hidden)}}

{{Header "Experiments"}}
{{ Experiments .Data.Experiments }}

{{Header "Additional help"}}
  For more information about a specific command, run:
  $ {{.CommandPath}}{{if eq .Name (GetProcessName)}} <command>{{end}} <subcommand> --help

  For guides and documentation, head over to {{LinkText "https://docs.slack.dev/tools/slack-cli"}}{{end}}

`
