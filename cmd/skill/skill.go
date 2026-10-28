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

package skill

import (
	_ "embed"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// skillFile is the "slack-cli" agent skill from the Slack Skills Plugin. The
// copy here ships with each release so the skill matches the commands of the
// installed version.
//
// https://github.com/slackapi/slack-skills-plugin/blob/main/skills/slack-cli/SKILL.md
//
//go:embed SKILL.md
var skillFile string

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skill",
		Aliases: []string{"skills"},
		Short:   "Print the agent skill for the Slack CLI",
		Long: strings.Join([]string{
			"Print the agent skill that teaches coding agents to use the Slack CLI.",
			"",
			"The skill is a markdown file with frontmatter that follows the Agent Skills",
			"format. Agents can read the output or save it to a skills directory.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Print the agent skill using the command",
				Command: "skill",
			},
			{
				Meaning: "Print the agent skill using the flag",
				Command: "--skill",
			},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunSkillCommand(clients, cmd)
		},
	}
	return cmd
}

// RunSkillCommand writes the agent skill to stdout without formatting
func RunSkillCommand(clients *shared.ClientFactory, cmd *cobra.Command) error {
	ctx := cmd.Context()
	span, _ := opentracing.StartSpanFromContext(ctx, "cmd.skill")
	defer span.Finish()

	_, err := clients.IO.WriteOut().Write([]byte(skillFile))
	return err
}
