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

package style

// Slack brand color palette.
// Single source of truth for all styling: lipgloss, huh themes, and bubbletea components.
//
// Colors from https://a.slack-edge.com/4d5bb/marketing/img/media-kit/slack_brand_guidelines_september2020.pdf

import "github.com/charmbracelet/lipgloss"

// Brand colors
var (
	slackAubergine = lipgloss.Color("#7C2852")
	slackBlue      = lipgloss.Color("#36c5f0")
	slackGreen     = lipgloss.Color("#2eb67d")
	slackYellow    = lipgloss.Color("#ecb22e")
	slackRed       = lipgloss.Color("#e01e5a")
	slackRedDark   = lipgloss.Color("#a01040")
)

// Supplementary colors
var (
	slackPool      = lipgloss.Color("#78d7dd")
	slackLegalGray = lipgloss.Color("#5e5d60")
)

// Adaptive colors that adjust for light/dark terminal backgrounds
var (
	slackOptionText      = lipgloss.AdaptiveColor{Light: "#1d1c1d", Dark: "#f4ede4"}
	slackDescriptionText = lipgloss.AdaptiveColor{Light: "#454447", Dark: "#b9b5b0"}
	slackPlaceholderText = lipgloss.AdaptiveColor{Light: "#5e5d60", Dark: "#868380"}
)
