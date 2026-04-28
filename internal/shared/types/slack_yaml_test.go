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

package types

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SlackYaml_hasValidIconPath(t *testing.T) {
	tests := map[string]struct {
		icon     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		"valid custom icon path returns true": {
			icon: "custom/icon.png",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "custom"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "custom", "icon.png"), []byte("img"), 0o644))
			},
			expected: true,
		},
		"invalid custom icon path returns false": {
			icon:     "missing/icon.png",
			setup:    func(t *testing.T, dir string) {},
			expected: false,
		},
		"no icon with default assets/icon.png present returns true": {
			icon: "",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "assets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "icon.png"), []byte("img"), 0o644))
			},
			expected: true,
		},
		"no icon with default assets/icon.jpg present returns true": {
			icon: "",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "assets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "icon.jpg"), []byte("img"), 0o644))
			},
			expected: true,
		},
		"no icon with default assets/icon.gif present returns true": {
			icon: "",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "assets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "icon.gif"), []byte("img"), 0o644))
			},
			expected: true,
		},
		"png takes priority over jpg in assets": {
			icon: "",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "assets"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "icon.png"), []byte("img"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "icon.jpg"), []byte("img"), 0o644))
			},
			expected: true,
		},
		"no icon and no default returns true": {
			icon:     "",
			setup:    func(t *testing.T, dir string) {},
			expected: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)

			origDir, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(dir))
			defer func() { require.NoError(t, os.Chdir(origDir)) }()

			sy := &SlackYaml{Icon: tc.icon}
			assert.Equal(t, tc.expected, sy.hasValidIconPath())
		})
	}
}

func Test_SlackYaml_Verify(t *testing.T) {
	tests := map[string]struct {
		icon      string
		setup     func(t *testing.T, dir string)
		expectErr bool
	}{
		"valid icon returns nil error": {
			icon: "icon.png",
			setup: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "icon.png"), []byte("img"), 0o644))
			},
			expectErr: false,
		},
		"invalid icon returns error": {
			icon:      "missing.png",
			setup:     func(t *testing.T, dir string) {},
			expectErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)

			origDir, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(dir))
			defer func() { require.NoError(t, os.Chdir(origDir)) }()

			sy := &SlackYaml{Icon: tc.icon}
			if tc.expectErr {
				assert.Error(t, sy.Verify())
			} else {
				assert.NoError(t, sy.Verify())
			}
		})
	}
}
