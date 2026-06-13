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

package manifest

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Flatten(t *testing.T) {
	tests := map[string]struct {
		manifest types.AppManifest
		expected map[string]any
	}{
		"flattens display_information fields": {
			manifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name:        "My App",
					Description: "A test app",
				},
			},
			expected: map[string]any{
				"display_information.name":        "My App",
				"display_information.description": "A test app",
			},
		},
		"flattens nested settings": {
			manifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "App",
				},
				Settings: &types.AppSettings{
					FunctionRuntime: types.LocallyRun,
				},
			},
			expected: map[string]any{
				"display_information.name":  "App",
				"settings.function_runtime": "local",
			},
		},
		"flattens functions map": {
			manifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "App",
				},
				Functions: map[string]types.ManifestFunction{
					"greet": {
						Title:       "Greet",
						Description: "Greets a user",
					},
				},
			},
			expected: map[string]any{
				"display_information.name":    "App",
				"functions.greet.title":       "Greet",
				"functions.greet.description": "Greets a user",
			},
		},
		"treats arrays as leaf values": {
			manifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "App",
				},
				OAuthConfig: &types.OAuthConfig{
					Scopes: &types.ManifestScopes{
						Bot: []string{"chat:write", "channels:read"},
					},
				},
			},
			expected: map[string]any{
				"display_information.name": "App",
				"oauth_config.scopes.bot":  []any{"chat:write", "channels:read"},
			},
		},
		"empty manifest has only display_information.name": {
			manifest: types.AppManifest{
				DisplayInformation: types.DisplayInformation{
					Name: "App",
				},
			},
			expected: map[string]any{
				"display_information.name": "App",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := Flatten(tc.manifest)
			require.NoError(t, err)
			for key, expectedVal := range tc.expected {
				assert.Contains(t, result, key)
				assert.Equal(t, expectedVal, result[key], "mismatch at key %s", key)
			}
		})
	}
}

func Test_Flatten_EscapesDotsInKeys(t *testing.T) {
	// Manifest function IDs may contain literal dots (e.g. "slack.users.lookup").
	// Flatten must backslash-escape those dots so the path remains parseable.
	manifest := types.AppManifest{
		DisplayInformation: types.DisplayInformation{Name: "App"},
		Functions: map[string]types.ManifestFunction{
			"slack.users.lookup": {Title: "Lookup", Description: "Lookup a user"},
		},
	}

	flat, err := Flatten(manifest)
	require.NoError(t, err)

	assert.Contains(t, flat, `functions.slack\.users\.lookup.title`)
	assert.Equal(t, "Lookup", flat[`functions.slack\.users\.lookup.title`])
	assert.Equal(t, "Lookup a user", flat[`functions.slack\.users\.lookup.description`])
}

func Test_SortedKeys(t *testing.T) {
	m := map[string]any{
		"z.field": "val",
		"a.field": "val",
		"m.field": "val",
	}
	keys := SortedKeys(m)
	assert.Equal(t, []string{"a.field", "m.field", "z.field"}, keys)
}
