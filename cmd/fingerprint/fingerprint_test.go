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

package fingerprint

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FingerprintPkgMock struct {
	mock.Mock
}

const fingerprintHashTest = "d41d8cd98f00b204e9800998ecf8427e"

func TestFingerprintCommand(t *testing.T) {
	// Create mocks
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()

	// Create clients that is mocked for testing
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	// Create the command
	cmd := NewCommand(clients)
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)

	if err != nil {
		assert.Fail(t, "cmd.Execute had unexpected error")
	}

	output := clientsMock.GetCombinedOutput()
	assert.Contains(t, output, fingerprintHashTest)
}
