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

package runtime

import (
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/runtime/deno"
	"github.com/slackapi/slack-cli/internal/runtime/node"
	"github.com/slackapi/slack-cli/internal/runtime/python"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Runtime_New(t *testing.T) {
	tests := map[string]struct {
		runtime             string
		expectedRuntimeType Runtime
	}{
		"Deno SDK": {
			runtime:             "deno",
			expectedRuntimeType: deno.New(),
		},
		"Bolt for JavaScript": {
			runtime:             "node",
			expectedRuntimeType: node.New(),
		},
		"Bolt for Python": {
			runtime:             "python",
			expectedRuntimeType: python.New(),
		},
		"Unsupported Runtime": {
			runtime:             "biggly-boo",
			expectedRuntimeType: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Run the test
			rt, _ := New(tc.runtime)
			require.IsType(t, tc.expectedRuntimeType, rt)
		})
	}
}

func Test_ActivatePythonVenvIfPresent(t *testing.T) {
	tests := map[string]struct {
		createVenv        bool
		expectedActivated bool
	}{
		"activates venv when it exists": {
			createVenv:        true,
			expectedActivated: true,
		},
		"no-op when venv does not exist": {
			createVenv:        false,
			expectedActivated: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			projectDir := "/path/to/project"
			venvPath := filepath.Join(projectDir, ".venv")

			osMock.On("Getenv", "PATH").Return("/usr/bin:/bin")
			osMock.AddDefaultMocks()

			if tc.createVenv {
				// Create the pip executable so venvExists returns true
				pipPath := filepath.Join(venvPath, "bin", "pip")
				err := fs.MkdirAll(filepath.Dir(pipPath), 0755)
				require.NoError(t, err)
				err = afero.WriteFile(fs, pipPath, []byte(""), 0755)
				require.NoError(t, err)
			}

			activated, err := ActivatePythonVenvIfPresent(fs, osMock, projectDir)
			require.NoError(t, err)
			require.Equal(t, tc.expectedActivated, activated)

			if tc.expectedActivated {
				osMock.AssertCalled(t, "Setenv", "VIRTUAL_ENV", venvPath)
				osMock.AssertCalled(t, "Setenv", "PATH", mock.Anything)
				osMock.AssertCalled(t, "Unsetenv", "PYTHONHOME")
			} else {
				osMock.AssertNotCalled(t, "Setenv", mock.Anything, mock.Anything)
				osMock.AssertNotCalled(t, "Unsetenv", mock.Anything)
			}
		})
	}
}

func Test_Runtime_NewDetectProject(t *testing.T) {
	tests := map[string]struct {
		sdkConfig           hooks.SDKCLIConfig
		expectedRuntimeType Runtime
	}{
		"Deno SDK": {
			sdkConfig:           hooks.SDKCLIConfig{Runtime: "deno"},
			expectedRuntimeType: deno.New(),
		},
		"Bolt for JavaScript": {
			sdkConfig:           hooks.SDKCLIConfig{Runtime: "node"},
			expectedRuntimeType: node.New(),
		},
		"Bolt for Python": {
			sdkConfig:           hooks.SDKCLIConfig{Runtime: "python"},
			expectedRuntimeType: python.New(),
		},
		"Unsupported Runtime": {
			sdkConfig:           hooks.SDKCLIConfig{Runtime: ""},
			expectedRuntimeType: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			ctx := slackcontext.MockContext(t.Context())
			fs := afero.NewMemMapFs()
			projectDirPath := "/path/to/project-name"

			// Run the test
			rt, _ := NewDetectProject(ctx, fs, projectDirPath, tc.sdkConfig)
			require.IsType(t, tc.expectedRuntimeType, rt)
		})
	}
}
