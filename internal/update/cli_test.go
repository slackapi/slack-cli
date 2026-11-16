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
	"errors"
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

func Test_CLIDependency_HasUpdate(t *testing.T) {
	releaseError := errors.New("release check failed")
	tests := map[string]struct {
		releaseInfo    *LatestCLIRelease
		releaseError   error
		expectedUpdate bool
		expectedError  error
	}{
		"returns false when no newer release was found": {
			releaseInfo:    nil,
			expectedUpdate: false,
		},
		"returns true when a newer release was found": {
			releaseInfo:    &LatestCLIRelease{Version: "v4.0.0"},
			expectedUpdate: true,
		},
		"returns the release error from the update check": {
			releaseError:   releaseError,
			expectedUpdate: false,
			expectedError:  releaseError,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cliDependency := &CLIDependency{
				releaseInfo:  tc.releaseInfo,
				releaseError: tc.releaseError,
			}
			hasUpdate, err := cliDependency.HasUpdate()
			assert.Equal(t, tc.expectedUpdate, hasUpdate)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
