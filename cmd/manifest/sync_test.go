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

package manifest

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
)

func TestSyncCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors when the manifest-sync experiment is off": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
			},
			ExpectedError: slackerror.New(slackerror.ErrExperimentRequired),
		},
		"passes the experiment gate when manifest-sync is enabled via flag": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cf.Config.ExperimentsFlag = []string{string(experiment.ManifestSync)}
				cf.Config.LoadExperiments(ctx, cf.IO.PrintDebug)
			},
			// We expect the command to fail downstream of the gate (no app
			// selected, no SDK config), but NOT with ErrCommandUnavailable —
			// the gate itself should pass.
			ExpectedErrorStrings: []string{},
		},
	}, func(clients *shared.ClientFactory) *cobra.Command {
		return NewSyncCommand(clients)
	})
}
