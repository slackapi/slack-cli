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

package hooks

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SDKCLIConfig_Exists(t *testing.T) {
	tests := map[string]struct {
		sdkCLIConfig  SDKCLIConfig
		expectedError error
		exists        bool
	}{
		"an initialized sdk configuration exists": {
			sdkCLIConfig: SDKCLIConfig{WorkingDirectory: "/path/to/project"},
			exists:       true,
		},
		"an uninitialized sdk configuration does not exist": {
			sdkCLIConfig:  SDKCLIConfig{WorkingDirectory: ""},
			expectedError: slackerror.New(slackerror.ErrInvalidSlackProjectDirectory),
			exists:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err, exists := tt.sdkCLIConfig.Exists()
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.exists, exists)
		})
	}
}

func Test_ProtocolResolution(t *testing.T) {
	tests := map[string]struct {
		config SDKCLIConfig
		check  func(t *testing.T, p Protocol)
	}{
		"Returns the first valid protocol version": {
			config: SDKCLIConfig{Config: struct {
				Watch                WatchOpts        `json:"watch,omitempty"`
				SDKManagedConnection bool             `json:"sdk-managed-connection-enabled,omitempty"`
				TriggerPaths         []string         `json:"trigger-paths,omitempty"`
				SupportedProtocols   ProtocolVersions `json:"protocol-version,omitempty"`
			}{
				SupportedProtocols: ProtocolVersions{
					"fake-news",
					HOOK_PROTOCOL_V2,
					"news-fake",
					HOOK_PROTOCOL_DEFAULT,
				},
			}},
			check: func(t *testing.T, p Protocol) {
				assert.Equal(t, p, HOOK_PROTOCOL_V2)
			},
		},
		"Returns default config if no valid protocols are present": {
			config: SDKCLIConfig{Config: struct {
				Watch                WatchOpts        `json:"watch,omitempty"`
				SDKManagedConnection bool             `json:"sdk-managed-connection-enabled,omitempty"`
				TriggerPaths         []string         `json:"trigger-paths,omitempty"`
				SupportedProtocols   ProtocolVersions `json:"protocol-version,omitempty"`
			}{
				SupportedProtocols: ProtocolVersions{
					"fake-news",
				},
			}},
			check: func(t *testing.T, p Protocol) {
				assert.Equal(t, p, HOOK_PROTOCOL_DEFAULT)
			},
		},
		"Returns default config if no protocols are present": {
			config: SDKCLIConfig{},
			check: func(t *testing.T, p Protocol) {
				assert.Equal(t, p, HOOK_PROTOCOL_DEFAULT)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			protocol := tt.config.Config.SupportedProtocols.Preferred()
			tt.check(t, protocol)
		})
	}
}

func Test_WatchOpts_IsAvailable(t *testing.T) {
	tests := map[string]struct {
		watchOpts           *WatchOpts
		expectedIsAvailable bool
	}{
		"WatchOpts exists": {
			watchOpts:           &WatchOpts{},
			expectedIsAvailable: true,
		},
		"WatchOpts not exists": {
			watchOpts:           nil,
			expectedIsAvailable: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			returnedIsAvailable := tt.watchOpts.IsAvailable()
			require.Equal(t, tt.expectedIsAvailable, returnedIsAvailable)
		})
	}
}
