package manifest

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/slackapi/slack-cli/internal/shared/types"
)

// Resolution represents which side the user chose for a given field.
type Resolution int

const (
	ResolveLocal  Resolution = iota // Use the local value
	ResolveRemote                   // Use the remote value
)

// MergeStrategy determines how all differences are resolved.
type MergeStrategy int

const (
	MergeAllLocal  MergeStrategy = iota // Use all local values
	MergeAllRemote                      // Use all remote values
	MergePerField                       // Resolve each field individually
)

// FieldResolution pairs a diff path with the user's choice.
type FieldResolution struct {
	Path       string
	Resolution Resolution
}

// Merge applies resolutions to produce a final manifest. It starts with the
// remote manifest as a base, then applies local values for fields resolved as
// local and keeps remote values for fields resolved as remote.
func Merge(local, remote types.AppManifest, resolutions []FieldResolution) (types.AppManifest, error) {
	localFlat, err := Flatten(local)
	if err != nil {
		return types.AppManifest{}, fmt.Errorf("failed to flatten local manifest: %w", err)
	}
	remoteFlat, err := Flatten(remote)
	if err != nil {
		return types.AppManifest{}, fmt.Errorf("failed to flatten remote manifest: %w", err)
	}

	merged := make(map[string]any)
	maps.Copy(merged, remoteFlat)

	// Apply resolutions
	resolutionMap := make(map[string]Resolution)
	for _, r := range resolutions {
		resolutionMap[r.Path] = r.Resolution
	}

	for path, resolution := range resolutionMap {
		switch resolution {
		case ResolveLocal:
			if val, exists := localFlat[path]; exists {
				merged[path] = val
			} else {
				delete(merged, path)
			}
		case ResolveRemote:
			if val, exists := remoteFlat[path]; exists {
				merged[path] = val
			} else {
				delete(merged, path)
			}
		}
	}

	// Also include local-only paths that were resolved as local
	for path, val := range localFlat {
		if _, inRemote := remoteFlat[path]; !inRemote {
			if res, hasResolution := resolutionMap[path]; hasResolution && res == ResolveLocal {
				merged[path] = val
			}
		}
	}

	return unflatten(merged)
}

// MergeAllFrom resolves all differences from one side.
func MergeAllFrom(local, remote types.AppManifest, diffs *DiffResult, strategy MergeStrategy) (types.AppManifest, error) {
	resolutions := make([]FieldResolution, 0, len(diffs.Diffs))
	for _, d := range diffs.Diffs {
		var res Resolution
		switch strategy {
		case MergeAllLocal:
			res = ResolveLocal
		case MergeAllRemote:
			res = ResolveRemote
		}
		resolutions = append(resolutions, FieldResolution{Path: d.Path, Resolution: res})
	}
	return Merge(local, remote, resolutions)
}

// unflatten converts a flat dot-notation map back into a nested structure,
// then marshals/unmarshals into AppManifest.
func unflatten(flat map[string]any) (types.AppManifest, error) {
	nested := unflattenToMap(flat)
	data, err := json.Marshal(nested)
	if err != nil {
		return types.AppManifest{}, fmt.Errorf("failed to marshal merged manifest: %w", err)
	}
	var result types.AppManifest
	if err := json.Unmarshal(data, &result); err != nil {
		return types.AppManifest{}, fmt.Errorf("failed to unmarshal merged manifest: %w", err)
	}
	return result, nil
}

// unflattenToMap reconstructs a nested map from dot-notation paths.
func unflattenToMap(flat map[string]any) map[string]any {
	result := make(map[string]any)
	for path, value := range flat {
		setNestedValue(result, path, value)
	}
	return result
}

func setNestedValue(m map[string]any, path string, value any) {
	parts := splitPath(path)
	current := m
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		next, exists := current[part]
		if !exists {
			next = make(map[string]any)
			current[part] = next
		}
		current = next.(map[string]any)
	}
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
