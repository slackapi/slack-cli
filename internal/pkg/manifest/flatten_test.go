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

func Test_SortedKeys(t *testing.T) {
	m := map[string]any{
		"z.field": "val",
		"a.field": "val",
		"m.field": "val",
	}
	keys := SortedKeys(m)
	assert.Equal(t, []string{"a.field", "m.field", "z.field"}, keys)
}
