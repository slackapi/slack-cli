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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/shared/types"
)

// ignoredDiffPaths are top-level manifest fields that the project may declare
// but Slack's apps.manifest.export does not echo back. Diffs at or under these
// paths are dropped to avoid spurious "only in project" entries on every run.
var ignoredDiffPaths = []string{
	"_metadata", // SDK-side schema annotations; not stored in app settings
}

// devLocalSuffixPaths are flattened paths where Slack's apps.manifest.export
// appends " (local)" for dev-installed apps. Diffs at these paths are
// dropped only when removing the suffix would make the values equal; real
// renames still surface.
var devLocalSuffixPaths = []string{
	"display_information.name",
	"features.bot_user.display_name",
}

const devLocalSuffix = " (local)"

// remoteFalseDefaultPaths are flattened paths where Slack's
// apps.manifest.export emits a default `false` for every app, even when the
// project has not declared the field. Remote-only diffs at these paths are
// dropped when the value is false so users do not see a phantom entry on
// every run; a real disagreement (e.g. local sets the field to true) still
// surfaces as a Modified diff.
var remoteFalseDefaultPaths = []string{
	"settings.is_mcp_enabled",
}

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
// returning all fields that differ between them. Paths under ignoredDiffPaths
// are excluded.
func Diff(local, remote types.AppManifest) (*DiffResult, error) {
	localFlat, err := Flatten(local)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten local manifest: %w", err)
	}
	remoteFlat, err := Flatten(remote)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten remote manifest: %w", err)
	}
	result, err := diffFlat(localFlat, remoteFlat)
	if err != nil {
		return nil, err
	}
	filtered := result.Diffs[:0]
	for _, d := range result.Diffs {
		if isIgnoredPath(d.Path) {
			continue
		}
		if isDevLocalSuffixDiff(d) {
			continue
		}
		if isRemoteFalseDefaultDiff(d) {
			continue
		}
		filtered = append(filtered, d)
	}
	result.Diffs = filtered
	return result, nil
}

// isIgnoredPath reports whether a flattened manifest path is at or under any
// entry in ignoredDiffPaths.
func isIgnoredPath(path string) bool {
	for _, prefix := range ignoredDiffPaths {
		if path == prefix || strings.HasPrefix(path, prefix+".") {
			return true
		}
	}
	return false
}

// isRemoteFalseDefaultDiff reports whether a remote-only diff is purely the
// result of Slack's apps.manifest.export emitting a default `false` for a
// field the project did not declare. Real disagreements (e.g. local sets
// the field to true) are not suppressed because they would surface as a
// Modified diff, not a remote-only diff.
func isRemoteFalseDefaultDiff(d FieldDiff) bool {
	if d.Type != DiffRemoteOnly {
		return false
	}
	matched := false
	for _, p := range remoteFalseDefaultPaths {
		if d.Path == p {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	v, ok := d.RemoteValue.(bool)
	return ok && !v
}

// isDevLocalSuffixDiff reports whether a Modified diff is purely the result
// of Slack's apps.manifest.export appending " (local)" to a name field for a
// dev-installed app. Real renames are not suppressed because trimming the
// suffix from RemoteValue would not produce LocalValue.
func isDevLocalSuffixDiff(d FieldDiff) bool {
	if d.Type != DiffModified {
		return false
	}
	matched := false
	for _, p := range devLocalSuffixPaths {
		if d.Path == p {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	local, ok := d.LocalValue.(string)
	if !ok {
		return false
	}
	remote, ok := d.RemoteValue.(string)
	if !ok {
		return false
	}
	return strings.TrimSuffix(remote, devLocalSuffix) == local
}

// diffFlat compares two flattened manifests and returns one FieldDiff per
// path that differs (modified, local-only, or remote-only).
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

// valuesEqual reports whether two leaf values from a flattened manifest are
// equivalent. It compares their JSON encodings so type-equivalent values
// (e.g. matching arrays or nested objects) compare equal.
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
