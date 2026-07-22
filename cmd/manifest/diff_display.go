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
	internalmanifest "github.com/slackapi/slack-cli/internal/manifest"
	"github.com/slackapi/slack-cli/internal/style"
)

func displayDiffs(ctx context.Context, io iostreams.IOStreamer, diffs *internalmanifest.DiffResult) {
	if !diffs.HasDifferences() {
		return
	}

	sorted := make([]internalmanifest.FieldDiff, len(diffs.Diffs))
	copy(sorted, diffs.Diffs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	io.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "books",
		Text:  "Manifest Diff",
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
		case internalmanifest.DiffLocalOnly:
			local = formatValue(d.LocalValue)
			remote = absentValue
		case internalmanifest.DiffRemoteOnly:
			local = absentValue
			remote = formatValue(d.RemoteValue)
		default:
			local = formatValue(d.LocalValue)
			remote = formatValue(d.RemoteValue)
		}
		io.PrintInfo(ctx, false, "  %s", d.Path)
		io.PrintInfo(ctx, false, "    Project:      %s", local)
		io.PrintInfo(ctx, false, "    App settings: %s", remote)
	}
	io.PrintInfo(ctx, false, "")
}

const absentValue = "(not set)"

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
		return style.TruncateRunes(string(data), 80)
	}
}
