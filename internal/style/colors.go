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

import (
	"image/color"
	"os"

	lipgloss "charm.land/lipgloss/v2"
)

// hasDarkBG caches the terminal background detection at package init.
var hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

// lightDark selects a color based on the terminal background.
var lightDark = lipgloss.LightDark(hasDarkBG)

// Brand colors — primary on light backgrounds, secondary on dark backgrounds
var (
	slackAubergine color.Color = lightDark(lipgloss.Color("#4a154b"), lipgloss.Color("#b26ec3")) // Core/Aubergine → Sec/Mauve
	slackBlue      color.Color = lightDark(lipgloss.Color("#36c5f0"), lipgloss.Color("#78d7dd")) // Core/Slack blue → Sec/Pool
	slackGreen     color.Color = lipgloss.Color("#2eb67d")                                       // Core/Slack green (reads well on both)
	slackYellow    color.Color = lightDark(lipgloss.Color("#ecb22e"), lipgloss.Color("#ffd57e")) // Core/Slack yellow → Sec/Sandbar
	slackRed       color.Color = lightDark(lipgloss.Color("#e01e5a"), lipgloss.Color("#f2606a")) // Core/Slack red → Sec/Salmon
	slackRedDark   color.Color = lightDark(lipgloss.Color("#5e1237"), lipgloss.Color("#f2606a")) // Sec/Berry → Sec/Salmon
)

// Supplementary colors
var (
	slackPool      color.Color = lipgloss.Color("#78d7dd")                                       // Sec/Pool
	slackLegalGray color.Color = lightDark(lipgloss.Color("#5e5d60"), lipgloss.Color("#eaeaea")) // Sec/Legal → Sec/Inactive gray
)

// Adaptive text colors
var (
	slackOptionText      color.Color = lightDark(lipgloss.Color("#1d1c1d"), lipgloss.Color("#f4ede4")) // Core/Black → Core/Horchatta
	slackDescriptionText color.Color = lightDark(lipgloss.Color("#454447"), lipgloss.Color("#5e5d60")) // Sec/Small text → Sec/Inactive gray
	slackPlaceholderText color.Color = lightDark(lipgloss.Color("#5e5d60"), lipgloss.Color("#5e5d60")) // Sec/Legal → Sec/Inactive gray
)
