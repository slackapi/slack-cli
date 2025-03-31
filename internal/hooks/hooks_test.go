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

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/require"
)

func Test_Hooks_GetHookExecutor(t *testing.T) {
	tests := map[string]struct {
		protocolVersions ProtocolVersions
		expectedType     interface{}
	}{
		"Type HOOK_PROTOCOL_V2": {
			protocolVersions: ProtocolVersions{HOOK_PROTOCOL_V2},
			expectedType:     &HookExecutorMessageBoundaryProtocol{},
		},
		"Type HOOK_PROTOCOL_DEFAULT": {
			protocolVersions: ProtocolVersions{HOOK_PROTOCOL_DEFAULT},
			expectedType:     &HookExecutorDefaultProtocol{},
		},
		"Both HOOK_PROTOCOL_V2 and HOOK_PROTOCOL_DEFAULT": {
			protocolVersions: ProtocolVersions{HOOK_PROTOCOL_V2, HOOK_PROTOCOL_DEFAULT},
			expectedType:     &HookExecutorMessageBoundaryProtocol{},
		},
		"Both HOOK_PROTOCOL_DEFAULT and HOOK_PROTOCOL_V2": {
			protocolVersions: ProtocolVersions{HOOK_PROTOCOL_DEFAULT, HOOK_PROTOCOL_V2},
			expectedType:     &HookExecutorDefaultProtocol{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os := slackdeps.NewOsMock()
			os.AddDefaultMocks()
			fs := slackdeps.NewFsMock()
			config := config.NewConfig(fs, os)
			io := iostreams.NewIOStreamsMock(config, fs, os)
			sdkConfig := NewSDKConfigMock()
			sdkConfig.Config.SupportedProtocols = tt.protocolVersions
			hookExecutor := GetHookExecutor(io, sdkConfig)
			require.IsType(t, tt.expectedType, hookExecutor)
		})
	}
}
