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

package upgrade

import (
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/update"
	"github.com/slackapi/slack-cli/internal/version"
	"github.com/spf13/cobra"
)

// autoUpdateFunc is a function pointer for tests to mock the update.AutoUpdate function
var autoUpdateFunc = update.AutoUpdate

// NewAutoUpdateCommand returns the hidden command that installs an update of the CLI in a
// background process. Telemetry is disabled since the command isn't run by the developer.
func NewAutoUpdateCommand(clients *shared.ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:    update.AutoUpdateCommandName + " <version>",
		Short:  "Install an update of the CLI in the background",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clients.Config.DisableTelemetryFlag = true
			return autoUpdateFunc(cmd.Context(), clients, version.Raw(), args[0])
		},
	}
}
