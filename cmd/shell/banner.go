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

package shell

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// bannerView returns the styled banner string at the given width.
func bannerView(width int, version string) string {
	line := lipgloss.NewStyle().Foreground(colorPool).Render(strings.Repeat("─", width))
	face := renderSlackbot()
	title := lipgloss.NewStyle().Bold(true).Foreground(colorAubergine).Render("Slack CLI Shell")
	ver := lipgloss.NewStyle().Foreground(colorPool).Render("  " + version)
	hint := lipgloss.NewStyle().Foreground(colorGray).Italic(true).Render("Type 'help' for commands, 'exit' to quit")
	info := title + ver + "\n" + hint
	body := lipgloss.JoinHorizontal(lipgloss.Center, face, "   "+info)
	return line + "\n" + body + "\n" + line
}

// Slack brand colors (reused from internal/style/charm_theme.go)
var (
	colorAubergine = lipgloss.Color("#7C2852")
	colorPool      = lipgloss.Color("#78d7dd")
	colorGray      = lipgloss.Color("#5e5d60")
	colorGreen     = lipgloss.Color("#2eb67d")
	colorBlue      = lipgloss.Color("#36c5f0")
)

// renderSlackbot returns a multi-colored ASCII slackbot face.
func renderSlackbot() string {
	box := lipgloss.NewStyle().Foreground(colorPool)
	eye := lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	smile := lipgloss.NewStyle().Foreground(colorGreen)
	lines := []string{
		box.Render("  ╭───────╮"),
		box.Render("  │ ") + eye.Render("●") + box.Render("   ") + eye.Render("●") + box.Render(" │"),
		box.Render("  │   ") + smile.Render("◡") + box.Render("   │"),
		box.Render("  ╰───────╯"),
	}
	return strings.Join(lines, "\n")
}

// renderGoodbye writes the goodbye message to the writer.
func renderGoodbye(w io.Writer) {
	msg := lipgloss.NewStyle().Foreground(colorGreen).Render("Goodbye!")
	fmt.Fprintf(w, "%s\n", msg)
}
