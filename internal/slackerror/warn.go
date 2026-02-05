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

package slackerror

import (
	"fmt"
	"slices"
	"strings"

	"github.com/slackapi/slack-cli/internal/style"
)

// Warning contains structured warning information
type Warning struct {
	Code        string `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	Remediation string `json:"remediation,omitempty"`
	Pointer     string `json:"pointer,omitempty"`
}

// Warnings holds information from multiple warning
type Warnings []Warning

// Warning formats the custom Warnings array with a message
func (warn Warnings) Warning(verbose bool, message string) string {
	if warn == nil {
		return ""
	}
	if len(warn) == 0 {
		return message
	}
	type warningKey struct {
		Code        string
		Message     string
		Remediation string
	}
	warningMap := make(map[warningKey][]string)
	for _, d := range warn {
		key := warningKey{Code: d.Code, Message: d.Message, Remediation: d.Remediation}
		warningMap[key] = append(warningMap[key], d.Pointer)
	}
	type source struct {
		Code        string
		Message     string
		Remediation string
		Pointers    []string
	}
	warningSources := []source{}
	for warning, pointers := range warningMap {
		slices.Sort(pointers)
		warningSources = append(warningSources, source{
			Code:        warning.Code,
			Message:     warning.Message,
			Remediation: warning.Remediation,
			Pointers:    pointers,
		})
	}
	slices.SortFunc(warningSources, func(a, b source) int {
		switch {
		case a.Code > b.Code:
			return 1
		case a.Code < b.Code:
			return -1
		case len(a.Pointers) <= 0 && len(b.Pointers) > 0:
			return 1
		case len(b.Pointers) <= 0 && len(a.Pointers) > 0:
			return -1
		case a.Pointers[0] > b.Pointers[0]:
			return 1
		case a.Pointers[0] < b.Pointers[0]:
			return -1
		default:
			return 0
		}
	})
	var warningsF []string
	for _, warning := range warningSources {
		var code, remediation string
		var pointers []string
		if warning.Code != "" {
			code = fmt.Sprintf(" (%s)", warning.Code)
		}
		if verbose || len(warning.Pointers) < 4 {
			for _, src := range warning.Pointers {
				if strings.TrimSpace(src) == "" {
					continue
				}
				pointers = append(pointers, fmt.Sprintf("Source: %s", src))
			}
		} else {
			pointers = append(pointers, fmt.Sprintf("Source: %s", warning.Pointers[0]))
			pointers = append(pointers, fmt.Sprintf("Source: %s", warning.Pointers[1]))
			pointers = append(pointers, fmt.Sprintf("Similar warnings from '%d' other sources can be revealed with %s", len(warning.Pointers)-2, style.Highlight("--verbose")))
		}
		if strings.TrimSpace(warning.Remediation) != "" {
			remediation = fmt.Sprintf("Suggestion: %s", warning.Remediation)
		}
		warningsF = append(warningsF, style.Sectionf(style.TextSection{
			Emoji: "memo",
			Text:  fmt.Sprintf("%s%s", warning.Message, style.Warning(code)),
			Secondary: []string{
				strings.Join(pointers, "\n"),
				remediation,
			},
		}))
	}
	return fmt.Sprintf("%s\n\n%s\n", message, strings.TrimSpace(strings.Join(warningsF, "\n")))
}
