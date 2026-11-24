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
	"context"
	"errors"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/version"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_NewAutoUpdateCommand(t *testing.T) {
	tests := map[string]struct {
		args          []string
		autoUpdateErr error
		expectedCalls int
		expectedErr   string
	}{
		"installs the target version": {
			args:          []string{"4.12.0"},
			expectedCalls: 1,
		},
		"returns the error from the update": {
			args:          []string{"4.12.0"},
			autoUpdateErr: errors.New("permission denied"),
			expectedCalls: 1,
			expectedErr:   "permission denied",
		},
		"rejects a target that is not a release version": {
			args:          []string{"4.12"},
			expectedCalls: 0,
			expectedErr:   "Value 4.12 is not a semantic version",
		},
		"requires a target version": {
			args:          []string{},
			expectedCalls: 0,
			expectedErr:   "accepts 1 arg(s), received 0",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			calls := 0
			original := autoUpdateFunc
			t.Cleanup(func() { autoUpdateFunc = original })
			autoUpdateFunc = func(ctx context.Context, c *shared.ClientFactory, currentVersion string, targetVersion string) error {
				calls++
				assert.Equal(t, version.Raw(), currentVersion)
				assert.Equal(t, "4.12.0", targetVersion)
				assert.True(t, c.Config.DisableTelemetryFlag, "telemetry is disabled before the update")
				return tc.autoUpdateErr
			}
			cmd := NewAutoUpdateCommand(clients)
			cmd.SetArgs(tc.args)
			testutil.MockCmdIO(clients.IO, cmd)
			err := cmd.ExecuteContext(ctx)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedCalls, calls)
			assert.True(t, cmd.Hidden)
		})
	}
}
