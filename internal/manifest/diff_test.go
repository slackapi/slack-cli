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

func Test_Diff(t *testing.T) {
	tests := map[string]struct {
		local    types.AppManifest
		remote   types.AppManifest
		expected []FieldDiff
	}{
		"identical manifests produce no diffs": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			expected: nil,
		},
		"modified field detected": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Local desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Remote desc"},
			},
			expected: []FieldDiff{
				{Path: "display_information.description", Type: DiffModified, LocalValue: "Local desc", RemoteValue: "Remote desc"},
			},
		},
		"local-only field detected": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Has desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			expected: []FieldDiff{
				{Path: "display_information.description", Type: DiffLocalOnly, LocalValue: "Has desc"},
			},
		},
		"remote-only field detected": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Remote only"},
			},
			expected: []FieldDiff{
				{Path: "display_information.description", Type: DiffRemoteOnly, RemoteValue: "Remote only"},
			},
		},
		"function added locally": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
				Functions: map[string]types.ManifestFunction{
					"greet": {Title: "Greet", Description: "Hello"},
				},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			expected: []FieldDiff{
				{Path: "functions.greet.description", Type: DiffLocalOnly, LocalValue: "Hello"},
				{Path: "functions.greet.title", Type: DiffLocalOnly, LocalValue: "Greet"},
			},
		},
		"array values compared as wholes": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
				OAuthConfig: &types.OAuthConfig{
					Scopes: &types.ManifestScopes{
						Bot: []string{"chat:write", "users:read"},
					},
				},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
				OAuthConfig: &types.OAuthConfig{
					Scopes: &types.ManifestScopes{
						Bot: []string{"chat:write", "files:read"},
					},
				},
			},
			expected: []FieldDiff{
				{
					Path:        "oauth_config.scopes.bot",
					Type:        DiffModified,
					LocalValue:  []any{"chat:write", "users:read"},
					RemoteValue: []any{"chat:write", "files:read"},
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := Diff(tc.local, tc.remote)
			require.NoError(t, err)
			if tc.expected == nil {
				assert.False(t, result.HasDifferences())
				return
			}
			assert.True(t, result.HasDifferences())
			for _, expectedDiff := range tc.expected {
				found := false
				for _, actualDiff := range result.Diffs {
					if actualDiff.Path == expectedDiff.Path {
						found = true
						assert.Equal(t, expectedDiff.Type, actualDiff.Type, "diff type mismatch for path %s", expectedDiff.Path)
						if expectedDiff.LocalValue != nil {
							assert.Equal(t, expectedDiff.LocalValue, actualDiff.LocalValue, "local value mismatch for path %s", expectedDiff.Path)
						}
						if expectedDiff.RemoteValue != nil {
							assert.Equal(t, expectedDiff.RemoteValue, actualDiff.RemoteValue, "remote value mismatch for path %s", expectedDiff.Path)
						}
						break
					}
				}
				assert.True(t, found, "expected diff not found for path %s", expectedDiff.Path)
			}
		})
	}
}

func Test_DiffResult_HasDifferences(t *testing.T) {
	t.Run("empty result has no differences", func(t *testing.T) {
		result := &DiffResult{}
		assert.False(t, result.HasDifferences())
	})

	t.Run("result with diffs has differences", func(t *testing.T) {
		result := &DiffResult{
			Diffs: []FieldDiff{{Path: "test", Type: DiffModified}},
		}
		assert.True(t, result.HasDifferences())
	})
}
