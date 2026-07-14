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

// claudeCodePluginHint is the marker to recommend the official Slack plugin to
// users of Claude Code.
//
// https://code.claude.com/docs/en/plugin-hints
const claudeCodePluginHint = `<claude-code-hint v="1" type="plugin" value="slack@claude-plugins-official" />`

// EmitClaudeCodePluginHint writes the Claude Code plugin recommendation marker
// to a writer that must be stderr to prompt installation without an appearance
// in actual outputs.
func EmitClaudeCodePluginHint(w io.Writer) {
	if os.Getenv("CLAUDECODE") == "" {
		return
	}
	fmt.Fprintln(w, claudeCodePluginHint)
}
