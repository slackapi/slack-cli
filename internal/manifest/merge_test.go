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
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Merge(t *testing.T) {
	tests := map[string]struct {
		local       types.AppManifest
		remote      types.AppManifest
		resolutions []FieldResolution
		expected    types.AppManifest
	}{
		"resolve all local keeps local values": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Local App", Description: "Local desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Remote App", Description: "Remote desc"},
			},
			resolutions: []FieldResolution{
				{Path: "display_information.name", Resolution: ResolveLocal},
				{Path: "display_information.description", Resolution: ResolveLocal},
			},
			expected: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Local App", Description: "Local desc"},
			},
		},
		"resolve all remote keeps remote values": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Local App", Description: "Local desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Remote App", Description: "Remote desc"},
			},
			resolutions: []FieldResolution{
				{Path: "display_information.name", Resolution: ResolveRemote},
				{Path: "display_information.description", Resolution: ResolveRemote},
			},
			expected: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Remote App", Description: "Remote desc"},
			},
		},
		"mixed resolution picks per field": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Local App", Description: "Local desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Remote App", Description: "Remote desc"},
			},
			resolutions: []FieldResolution{
				{Path: "display_information.name", Resolution: ResolveLocal},
				{Path: "display_information.description", Resolution: ResolveRemote},
			},
			expected: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "Local App", Description: "Remote desc"},
			},
		},
		"local-only field resolved as local is included": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Has desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			resolutions: []FieldResolution{
				{Path: "display_information.description", Resolution: ResolveLocal},
			},
			expected: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Has desc"},
			},
		},
		"local-only field resolved as remote is excluded": {
			local: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App", Description: "Has desc"},
			},
			remote: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
			resolutions: []FieldResolution{
				{Path: "display_information.description", Resolution: ResolveRemote},
			},
			expected: types.AppManifest{
				DisplayInformation: types.DisplayInformation{Name: "App"},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := Merge(tc.local, tc.remote, tc.resolutions)
			require.NoError(t, err)
			assert.Equal(t, tc.expected.DisplayInformation.Name, result.DisplayInformation.Name)
			assert.Equal(t, tc.expected.DisplayInformation.Description, result.DisplayInformation.Description)
		})
	}
}

func Test_unflatten_PathCollision(t *testing.T) {
	tests := map[string]struct {
		flat map[string]any
	}{
		"leaf at parent path collides with deeper path": {
			flat: map[string]any{
				"a.b":   "scalar",
				"a.b.c": "deep",
			},
		},
		"deeper path collides with leaf at parent": {
			flat: map[string]any{
				"a.b.c": "deep",
				"a.b":   "scalar",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := unflatten(tc.flat)
			require.Error(t, err, "expected path-collision error, got nil")
			slackErr := slackerror.ToSlackError(err)
			assert.Equal(t, slackerror.ErrInvalidManifest, slackErr.Code)
		})
	}
}

func Test_MergeAllFrom(t *testing.T) {
	local := types.AppManifest{
		DisplayInformation: types.DisplayInformation{Name: "Local", Description: "Local desc"},
	}
	remote := types.AppManifest{
		DisplayInformation: types.DisplayInformation{Name: "Remote", Description: "Remote desc"},
	}
	diffs, err := Diff(local, remote)
	require.NoError(t, err)

	t.Run("merge all local", func(t *testing.T) {
		result, err := MergeAllFrom(local, remote, diffs, MergeAllLocal)
		require.NoError(t, err)
		assert.Equal(t, "Local", result.DisplayInformation.Name)
		assert.Equal(t, "Local desc", result.DisplayInformation.Description)
	})

	t.Run("merge all remote", func(t *testing.T) {
		result, err := MergeAllFrom(local, remote, diffs, MergeAllRemote)
		require.NoError(t, err)
		assert.Equal(t, "Remote", result.DisplayInformation.Name)
		assert.Equal(t, "Remote desc", result.DisplayInformation.Description)
	})
}
