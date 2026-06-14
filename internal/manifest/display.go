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

// formatValue renders a leaf value for display. Strings are quoted, other
// values are JSON-encoded, and any value longer than 80 characters is
// truncated with an ellipsis.
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
