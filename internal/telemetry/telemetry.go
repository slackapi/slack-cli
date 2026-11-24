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

// Package telemetry decides whether the CLI sends telemetry.
package telemetry

import (
	"strings"

	"github.com/slackapi/slack-cli/internal/version"
)

// EnvDisable is the environment variable that turns telemetry off
const EnvDisable = "SLACK_DISABLE_TELEMETRY"

// Disabled reports whether SLACK_DISABLE_TELEMETRY or SLACK_TEST_VERSION turn telemetry off
func Disabled(getenv func(string) string) bool {
	disable := strings.TrimSpace(getenv(EnvDisable))
	testVersion := strings.TrimSpace(getenv(version.EnvTestVersion))
	return (disable != "" && disable != "false" && disable != "0") || testVersion != ""
}
