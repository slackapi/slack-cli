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

package icon

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ResolveIconPath(t *testing.T) {
	tests := map[string]struct {
		manifestIcon string
		files        []string
		expected     string
	}{
		"manifest icon set returns it directly": {
			manifestIcon: "custom/my-icon.png",
			expected:     "custom/my-icon.png",
		},
		"manifest icon preferred over magic icon files": {
			manifestIcon: "custom/my-icon.png",
			files:        []string{"assets/icon.png", "icon.png"},
			expected:     "custom/my-icon.png",
		},
		"assets/icon.png found": {
			files:    []string{"assets/icon.png"},
			expected: "assets/icon.png",
		},
		"assets/icon.jpg found": {
			files:    []string{"assets/icon.jpg"},
			expected: "assets/icon.jpg",
		},
		"assets/icon.jpeg found": {
			files:    []string{"assets/icon.jpeg"},
			expected: "assets/icon.jpeg",
		},
		"assets/icon.gif found": {
			files:    []string{"assets/icon.gif"},
			expected: "assets/icon.gif",
		},
		"png wins over other extensions": {
			files:    []string{"assets/icon.jpg", "assets/icon.jpeg", "assets/icon.gif", "assets/icon.png"},
			expected: "assets/icon.png",
		},
		"jpg wins over gif in assets": {
			files:    []string{"assets/icon.jpg", "assets/icon.gif"},
			expected: "assets/icon.jpg",
		},
		"root icon.png found when no assets": {
			files:    []string{"icon.png"},
			expected: "icon.png",
		},
		"root icon.jpg found when no assets": {
			files:    []string{"icon.jpg"},
			expected: "icon.jpg",
		},
		"assets takes priority over root": {
			files:    []string{"assets/icon.gif", "icon.png"},
			expected: "assets/icon.gif",
		},
		"no icon files returns empty": {
			files:    []string{},
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for _, f := range tc.files {
				require.NoError(t, afero.WriteFile(fs, f, []byte("img"), 0o644))
			}
			result := ResolveIconPath(fs, tc.manifestIcon)
			assert.Equal(t, tc.expected, result)
		})
	}
}
