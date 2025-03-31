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

package env

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
)

func Test_Env_Command(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"shows the help page without commands or arguments or flags": {
			ExpectedStdoutOutputs: []string{
				"Add an environment variable",
				"List all environment variables",
				"Remove an environment variable",
			},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		cmd := NewCommand(clients)
		return cmd
	})
}
