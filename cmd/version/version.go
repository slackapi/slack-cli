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

package version

import (
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long: strings.Join([]string{
			"All software has versions. This is ours.",
			"",
			"Version numbers follow the semantic versioning specification (semver)",
			"and are always prefixed with a `v`, such as `v3.0.1`.",
			"",
			"Given a version number MAJOR.MINOR.PATCH:",
			"",
			"1. MAJOR versions have incompatible, breaking changes",
			"2. MINOR versions add functionality that is a backward compatible",
			"3. PATCH versions make bug fixes that are backward compatible",
			"",
			"Experiments are patch version until officially released.",
			"",
			"Development builds use `git describe` and contain helpful info,",
			"such as the prior release and specific commit of the build.",
			"",
			"Given a version number `v3.0.1-7-g822d09a`:",
			"",
			"1. `v3.0.1`   is the version of the prior release",
			"2. `7`        is the number of commits ahead of the `v3.0.1` git tag",
			"3. `g822d09a` is the git commit for this build, prefixed with `g`",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Print version number using the command",
				Command: "version",
			},
			{
				Meaning: "Print version number using the flag",
				Command: "--version",
			},
			{
				Meaning: "Print version and skip update check",
				Command: "--version --skip-update",
			},
		}),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			span, _ := opentracing.StartSpanFromContext(ctx, "cmd.version")
			defer span.Finish()

			cmd.Println(Template())
		},
	}

	return cmd
}

func Template() string {
	processName := cmdutil.GetProcessName()
	version := version.Get()

	return fmt.Sprintf(style.Secondary("Using %s %s"), processName, version)
}
