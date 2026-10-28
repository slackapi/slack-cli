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

package issueresolver

// GitHubIssuePageAudit is a concrete example of a page-level accessibility audit
// based on the Lighthouse model.
type GitHubIssuePageAudit struct {
	Page               string          `json:"page"`
	AccessibilityScore int             `json:"accessibilityScore"`
	Rules              []LighthouseRule `json:"rules"`
}

// NewGitHubIssuePageAudit creates a sample audit for the GitHub issues page.
func NewGitHubIssuePageAudit() GitHubIssuePageAudit {
	return GitHubIssuePageAudit{
		Page:               "https://github.com/actions/add-to-project/issues",
		AccessibilityScore: 95,
		Rules: []LighthouseRule{
			ButtonNameRule,
			HeadingOrderRule,
		},
	}
}

// AuditSummary produces a score summary for the GitHub issue page audit.
func (a GitHubIssuePageAudit) AuditSummary() ScoreSummary {
	return ScoreRules(a.Rules)
}

// Ensure the package contains a concrete registry of weighted Lighthouse audits.
var GitHubIssuePageAuditRules = []LighthouseRule{
	ButtonNameRule,
	HeadingOrderRule,
}
