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
	"os"
	"runtime"
	"strings"
)

// AIAgent represents a detected AI coding agent that invoked the CLI.
type AIAgent struct {
	Name  string
	Entry string
}

// Detect checks environment variables to determine if the CLI is being run by
// an AI coding agent. Returns nil if no agent is detected. Detection priority:
// CLAUDECODE > CODEX_CI > GEMINI_CLI > CLINE_ACTIVE > CURSOR_AGENT > AGENT.
func Detect() *AIAgent {
	switch {
	case os.Getenv("CLAUDECODE") == "1":
		return &AIAgent{
			Name:  "claude-code",
			Entry: os.Getenv("CLAUDE_CODE_ENTRYPOINT"),
		}
	case os.Getenv("CODEX_CI") == "1":
		return &AIAgent{Name: "codex"}
	case os.Getenv("GEMINI_CLI") == "1":
		return &AIAgent{Name: "gemini-cli"}
	case os.Getenv("CLINE_ACTIVE") == "true":
		return &AIAgent{Name: "cline"}
	case os.Getenv("CURSOR_AGENT") == "1":
		return &AIAgent{Name: "cursor"}
	case os.Getenv("AGENT") != "":
		return &AIAgent{Name: os.Getenv("AGENT")}
	default:
		return nil
	}
}

// DetectName returns the normalized name of the detected AI agent, or an empty
// string if no agent is detected.
func DetectName() string {
	if agent := Detect(); agent != nil {
		return agent.Name
	}
	return ""
}

// BuildUserAgent constructs the HTTP User-Agent header value for the CLI. If an
// AI agent is detected, an "AI-Agent (name=..., entry=...)" suffix is appended.
func BuildUserAgent(cliVersion string) string {
	ua := fmt.Sprintf("slack-cli/%s (os: %s)", cliVersion, runtime.GOOS)
	if agent := Detect(); agent != nil {
		var parts []string
		parts = append(parts, "name="+agent.Name)
		if agent.Entry != "" {
			parts = append(parts, "entry="+agent.Entry)
		}
		ua += " AI-Agent (" + strings.Join(parts, ", ") + ")"
	}
	return ua
}
