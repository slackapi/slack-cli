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

package update

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/stretchr/testify/assert"
)

func Test_NewCLIDependency(t *testing.T) {
	tests := map[string]struct {
		version string
	}{
		"stores the clients and the current version": {
			version: "v3.4.5",
		},
		"stores an empty version when none is known": {
			version: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			cliDependency := NewCLIDependency(clients, tc.version)
			assert.Same(t, clients, cliDependency.clients)
			assert.Equal(t, tc.version, cliDependency.version)
			assert.Nil(t, cliDependency.releaseInfo)
			assert.Nil(t, cliDependency.releaseError)
		})
	}
}
