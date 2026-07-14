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

package useragent

import (
	"fmt"
	"io"
	"os"
)

// claudeCodePluginHint is the one-line marker that recommends the official Slack
// plugin to Claude Code users. Claude Code scans command output for this tag,
// strips it before the output reaches the model (so it never counts toward token
// usage), and shows a one-time install prompt. The tag must occupy its own line
// and the plugin must live in Anthropic's official marketplace to have any
// effect. See https://code.claude.com/docs/en/plugin-hints.
const claudeCodePluginHint = `<claude-code-hint v="1" type="plugin" value="slack@claude-plugins-official" />`

// EmitClaudeCodePluginHint writes the Claude Code plugin-recommendation marker to
// w on its own line when the CLI is running inside Claude Code, prompting a
// one-time install of the official Slack plugin. It is a no-op in every other
// environment. Claude Code deduplicates the hint by plugin and per session, so
// callers may emit it on every invocation without spamming the user.
func EmitClaudeCodePluginHint(w io.Writer) {
	if os.Getenv("CLAUDECODE") != "1" {
		return
	}
	fmt.Fprintln(w, claudeCodePluginHint)
}
