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
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/style"
)

// DisplayDiffs prints the differences to the terminal.
func DisplayDiffs(ctx context.Context, io iostreams.IOStreamer, diffs *DiffResult) {
	if !diffs.HasDifferences() {
		return
	}

	sorted := make([]FieldDiff, len(diffs.Diffs))
	copy(sorted, diffs.Diffs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	io.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "App Manifest",
		Secondary: []string{
			fmt.Sprintf("Found %d difference(s) between project and app settings", len(sorted)),
		},
	}))

	for _, d := range sorted {
		io.PrintInfo(ctx, false, "")
		switch d.Type {
		case DiffModified:
			io.PrintInfo(ctx, false, "  %s", style.Bold(d.Path))
			io.PrintInfo(ctx, false, "    Project:      %s", formatValue(d.LocalValue))
			io.PrintInfo(ctx, false, "    App settings: %s", formatValue(d.RemoteValue))
		case DiffLocalOnly:
			io.PrintInfo(ctx, false, "  %s %s", style.Bold(d.Path), "(only in project)")
			io.PrintInfo(ctx, false, "    Value: %s", formatValue(d.LocalValue))
		case DiffRemoteOnly:
			io.PrintInfo(ctx, false, "  %s %s", style.Bold(d.Path), "(only in app settings)")
			io.PrintInfo(ctx, false, "    Value: %s", formatValue(d.RemoteValue))
		}
	}
	io.PrintInfo(ctx, false, "")
}

// PromptResolutionStrategy asks the user how they want to resolve differences.
func PromptResolutionStrategy(ctx context.Context, io iostreams.IOStreamer) (MergeStrategy, error) {
	options := []string{
		"Use all project values",
		"Use all app settings values",
		"Choose for each difference",
	}
	resp, err := io.SelectPrompt(ctx, "How would you like to resolve these differences?", options, iostreams.SelectPromptConfig{
		Required: true,
	})
	if err != nil {
		return 0, err
	}
	switch resp.Index {
	case 0:
		return MergeAllLocal, nil
	case 1:
		return MergeAllRemote, nil
	default:
		return MergePerField, nil
	}
}

// PromptFieldResolutions asks the user to resolve each difference individually.
func PromptFieldResolutions(ctx context.Context, io iostreams.IOStreamer, diffs *DiffResult) ([]FieldResolution, error) {
	sorted := make([]FieldDiff, len(diffs.Diffs))
	copy(sorted, diffs.Diffs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	resolutions := make([]FieldResolution, 0, len(sorted))
	for _, d := range sorted {
		var options []string
		switch d.Type {
		case DiffModified:
			options = []string{
				fmt.Sprintf("Use project: %s", formatValue(d.LocalValue)),
				fmt.Sprintf("Use app settings: %s", formatValue(d.RemoteValue)),
			}
		case DiffLocalOnly:
			options = []string{
				"Keep (include in merged manifest)",
				"Remove (exclude from merged manifest)",
			}
		case DiffRemoteOnly:
			options = []string{
				"Remove (exclude from merged manifest)",
				"Keep (include in merged manifest)",
			}
		}

		resp, err := io.SelectPrompt(ctx, d.Path, options, iostreams.SelectPromptConfig{
			Required: true,
		})
		if err != nil {
			return nil, err
		}

		var resolution Resolution
		switch d.Type {
		case DiffModified:
			if resp.Index == 0 {
				resolution = ResolveLocal
			} else {
				resolution = ResolveRemote
			}
		case DiffLocalOnly:
			if resp.Index == 0 {
				resolution = ResolveLocal
			} else {
				resolution = ResolveRemote
			}
		case DiffRemoteOnly:
			if resp.Index == 0 {
				resolution = ResolveLocal
			} else {
				resolution = ResolveRemote
			}
		}

		resolutions = append(resolutions, FieldResolution{Path: d.Path, Resolution: resolution})
	}
	return resolutions, nil
}

func formatValue(v any) string {
	if v == nil {
		return "(not present)"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	default:
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		s := string(data)
		if len(s) > 80 {
			return s[:77] + "..."
		}
		return s
	}
}
