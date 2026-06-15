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
			fmt.Sprintf("Found %d %s between project and app settings", len(sorted), style.Pluralize("difference", "differences", len(sorted))),
		},
	}))

	for i, d := range sorted {
		if i > 0 {
			io.PrintInfo(ctx, false, "")
		}
		var local, remote string
		switch d.Type {
		case DiffLocalOnly:
			local = formatValue(d.LocalValue)
			remote = absentValue
		case DiffRemoteOnly:
			local = absentValue
			remote = formatValue(d.RemoteValue)
		default:
			local = formatValue(d.LocalValue)
			remote = formatValue(d.RemoteValue)
		}
		io.PrintInfo(ctx, false, "  %s", style.Bold(d.Path))
		io.PrintInfo(ctx, false, "    Project:      %s", local)
		io.PrintInfo(ctx, false, "    App settings: %s", remote)
	}
	io.PrintInfo(ctx, false, "")
}

// absentValue is shown opposite a present value when a field exists on only
// one side of a diff. It is intentionally distinct from formatValue(nil)'s
// "(not present)", which represents a JSON null on a side that does have
// the field.
const absentValue = "(not set)"

// formatValue renders a leaf value for display. Strings are quoted, other
// values are JSON-encoded, and any value longer than 80 runes is
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
		return truncateRunes(string(data), 80)
	}
}

// truncateRunes returns s unchanged if it is at most max runes, otherwise it
// returns the first max-3 runes followed by "...". Splitting on runes (rather
// than bytes) avoids cutting through a multi-byte UTF-8 character.
func truncateRunes(s string, max int) string {
	if max <= 3 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}
