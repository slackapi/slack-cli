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

package runtime

import (
	"context"
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/runtime/deno"
	"github.com/slackapi/slack-cli/internal/runtime/node"
	"github.com/slackapi/slack-cli/internal/runtime/python"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_Runtime_New(t *testing.T) {
	tests := []struct {
		name                string
		runtime             string
		expectedRuntimeType Runtime
	}{
		{
			name:                "Deno SDK",
			runtime:             "deno",
			expectedRuntimeType: deno.New(),
		},
		{
			name:                "Bolt for JavaScript",
			runtime:             "node",
			expectedRuntimeType: node.New(),
		},
		{
			name:                "Bolt for Python",
			runtime:             "python",
			expectedRuntimeType: python.New(),
		},
		{
			name:                "Unsupported Runtime",
			runtime:             "biggly-boo",
			expectedRuntimeType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the test
			rt, _ := New(tt.runtime)
			require.IsType(t, tt.expectedRuntimeType, rt)
		})
	}
}

func Test_Runtime_NewDetectProject(t *testing.T) {
	tests := []struct {
		name                string
		sdkConfig           hooks.SDKCLIConfig
		expectedRuntimeType Runtime
	}{
		{
			name:                "Deno SDK",
			sdkConfig:           hooks.SDKCLIConfig{Runtime: "deno"},
			expectedRuntimeType: deno.New(),
		},
		{
			name:                "Bolt for JavaScript",
			sdkConfig:           hooks.SDKCLIConfig{Runtime: "node"},
			expectedRuntimeType: node.New(),
		},
		{
			name:                "Bolt for Python",
			sdkConfig:           hooks.SDKCLIConfig{Runtime: "python"},
			expectedRuntimeType: python.New(),
		},
		{
			name:                "Unsupported Runtime",
			sdkConfig:           hooks.SDKCLIConfig{Runtime: ""},
			expectedRuntimeType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			fs := afero.NewMemMapFs()
			projectDirPath := "/path/to/project-name"

			// Run the test
			rt, _ := NewDetectProject(ctx, fs, projectDirPath, tt.sdkConfig)
			require.IsType(t, tt.expectedRuntimeType, rt)
		})
	}
}
