package manifest

import (
	"encoding/json"
	"fmt"

	"github.com/slackapi/slack-cli/internal/shared/types"
)

// DiffType describes how a field differs between local and remote.
type DiffType int

const (
	DiffModified   DiffType = iota // Both sides have the field but with different values
	DiffLocalOnly                  // Field exists only in local (added locally or deleted remotely)
	DiffRemoteOnly                 // Field exists only in remote (added remotely or deleted locally)
)

// FieldDiff represents a single difference between local and remote manifests.
type FieldDiff struct {
	Path        string
	Type        DiffType
	LocalValue  any
	RemoteValue any
}

// DiffResult holds all differences found between two manifests.
type DiffResult struct {
	Diffs []FieldDiff
}

// HasDifferences returns true if any differences were found.
func (dr *DiffResult) HasDifferences() bool {
	return len(dr.Diffs) > 0
}

// Diff performs a two-way comparison between local and remote manifests,
// returning all fields that differ between them.
func Diff(local, remote types.AppManifest) (*DiffResult, error) {
	localFlat, err := Flatten(local)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten local manifest: %w", err)
	}
	remoteFlat, err := Flatten(remote)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten remote manifest: %w", err)
	}
	return diffFlat(localFlat, remoteFlat)
}

func diffFlat(local, remote map[string]any) (*DiffResult, error) {
	result := &DiffResult{}
	seen := make(map[string]bool)

	for path, localVal := range local {
		seen[path] = true
		remoteVal, exists := remote[path]
		if !exists {
			result.Diffs = append(result.Diffs, FieldDiff{
				Path:       path,
				Type:       DiffLocalOnly,
				LocalValue: localVal,
			})
			continue
		}
		equal, err := valuesEqual(localVal, remoteVal)
		if err != nil {
			return nil, fmt.Errorf("failed to compare manifest values at %q: %w", path, err)
		}
		if !equal {
			result.Diffs = append(result.Diffs, FieldDiff{
				Path:        path,
				Type:        DiffModified,
				LocalValue:  localVal,
				RemoteValue: remoteVal,
			})
		}
	}

	for path, remoteVal := range remote {
		if seen[path] {
			continue
		}
		result.Diffs = append(result.Diffs, FieldDiff{
			Path:        path,
			Type:        DiffRemoteOnly,
			RemoteValue: remoteVal,
		})
	}

	return result, nil
}

func valuesEqual(a, b any) (bool, error) {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	return string(aJSON) == string(bJSON), nil
}
