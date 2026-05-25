package manifest

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
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
