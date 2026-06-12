package manifest

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_WriteManifestLocal(t *testing.T) {
	t.Run("writes merged manifest to existing file", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		original := `{
  "display_information": {
    "name": "Original App"
  }
}
`
		require.NoError(t, afero.WriteFile(fs, "/project/manifest.json", []byte(original), 0644))

		manifest := types.AppManifest{
			DisplayInformation: types.DisplayInformation{Name: "Merged App", Description: "New desc"},
		}

		result, err := WriteManifestLocal(fs, "/project", manifest)
		require.NoError(t, err)
		assert.True(t, result.Written)
		assert.Equal(t, "/project/manifest.json", result.FilePath)

		content, err := afero.ReadFile(fs, "/project/manifest.json")
		require.NoError(t, err)
		assert.Contains(t, string(content), "Merged App")
		assert.Contains(t, string(content), "New desc")
	})

	t.Run("preserves original key order", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		original := `{
  "settings": {"function_runtime": "local"},
  "display_information": {"name": "App"}
}
`
		require.NoError(t, afero.WriteFile(fs, "/project/manifest.json", []byte(original), 0644))

		manifest := types.AppManifest{
			DisplayInformation: types.DisplayInformation{Name: "App"},
			Settings:           &types.AppSettings{FunctionRuntime: types.LocallyRun},
		}

		result, err := WriteManifestLocal(fs, "/project", manifest)
		require.NoError(t, err)
		assert.True(t, result.Written)

		content, err := afero.ReadFile(fs, "/project/manifest.json")
		require.NoError(t, err)
		contentStr := string(content)
		// settings should come before display_information (matching original order)
		settingsIdx := indexOf(contentStr, "settings")
		displayIdx := indexOf(contentStr, "display_information")
		assert.Less(t, settingsIdx, displayIdx, "key order should be preserved from original file")
	})

	t.Run("does not leave a temporary file after success", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		original := `{"display_information":{"name":"Original"}}`
		require.NoError(t, afero.WriteFile(fs, "/project/manifest.json", []byte(original), 0644))

		manifest := types.AppManifest{
			DisplayInformation: types.DisplayInformation{Name: "Merged"},
		}

		_, err := WriteManifestLocal(fs, "/project", manifest)
		require.NoError(t, err)

		entries, err := afero.ReadDir(fs, "/project")
		require.NoError(t, err)
		for _, e := range entries {
			assert.NotContains(t, e.Name(), ".tmp", "temporary file should be cleaned up after atomic write")
		}
	})

	t.Run("returns warning when manifest.json does not exist", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		manifest := types.AppManifest{
			DisplayInformation: types.DisplayInformation{Name: "App"},
		}

		result, err := WriteManifestLocal(fs, "/project", manifest)
		require.NoError(t, err)
		assert.False(t, result.Written)
		assert.Contains(t, result.Warning, "No manifest.json found")
	})
}

func indexOf(s, substr string) int {
	for i := range s {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
