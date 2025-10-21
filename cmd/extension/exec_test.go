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

package extension

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/spf13/cobra"
)

func Test_Extension_ExecCommand(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"errors without the experiment flag": {
			CmdArgs:       []string{"glitch"},
			ExpectedError: slackerror.New(slackerror.ErrMissingExperiment),
		},
		"no extension to execute errors": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{"extension"}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedError: slackerror.New(slackerror.ErrMissingExtension),
		},
		"the glitch extension exists": {
			CmdArgs: []string{"glitch"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{"extension"}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedStdoutOutputs: []string{"\x1b[?25l"},
		},
		"attempts to execute a missing extension": {
			CmdArgs: []string{"404"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.AddDefaultMocks()
				cm.Config.ExperimentsFlag = []string{"extension"}
				cm.Config.LoadExperiments(ctx, cm.IO.PrintDebug)
			},
			ExpectedErrorStrings: []string{
				slackerror.ErrSDKHookInvocationFailed,
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		cmd := NewExtensionExecCommand(cf)
		return cmd
	})
}
