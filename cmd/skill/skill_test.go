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
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewCommand(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, skillFile, clientsMock.GetStdoutOutput())
	assert.Equal(t, []string{"skills"}, cmd.Aliases)
}

func Test_NewCommand_Args(t *testing.T) {
	ctx := slackcontext.MockContext(t.Context())
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())

	cmd := NewCommand(clients)
	cmd.SetArgs([]string{"unexpected"})
	testutil.MockCmdIO(clients.IO, cmd)

	err := cmd.ExecuteContext(ctx)
	assert.Error(t, err)
}

func Test_skillFile(t *testing.T) {
	assert.Contains(t, skillFile, "---\nname: slack-cli\n", "should start with skill frontmatter")
	assert.Contains(t, skillFile, "\ndescription: ", "should include a skill description")
}
