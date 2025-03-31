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

package fingerprint

import (
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// A randomly generated hash
const fingerprintHash = "d41d8cd98f00b204e9800998ecf8427e"

// _fingerprint is a hidden command that always displays the same unique string.
// The command can be used to uniquely identify the Slack CLI binary from other
// third-party binaries with the same name.
func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "_fingerprint",
		Short: "Print a string to uniquely identify the Slack CLI",
		Long: strings.Join([]string{
			"Print a string to uniquely identify the Slack CLI from other binaries",
			"",
			"This helps prevent overwriting any custom scripts during installation",
		}, "\n"),
		Hidden: true,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "_fingerprint", Meaning: "Print the unique value that identifies the Slack CLI binary"},
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			var span, _ = opentracing.StartSpanFromContext(cmd.Context(), "cmd._fingerprint")
			defer span.Finish()

			clients.IO.PrintInfo(cmd.Context(), false, fingerprintHash)
			return nil
		},
	}

	return cmd
}
