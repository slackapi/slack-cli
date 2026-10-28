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

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/issueresolver"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/spf13/cobra"
)

type cmdFlags struct {
	format string
}

var flags cmdFlags

// NewCommand returns a new Cobra command for analyzing issue page quality.
func NewCommand(clients *shared.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-resolver [audit-type]",
		Short: "Analyze GitHub issue pages for accessibility and performance problems",
		Long: strings.Join([]string{
			"Assess issue pages using the same weighted rule model used by Lighthouse audits.",
			"",
			"Supported audit-types:",
			"  - github-accessibility",
			"  - tenant-picker-performance",
			"  - button-name",
			"",
			"If no audit type is supplied, the command returns the default GitHub issue page accessibility audit.",
		}, "\n"),
		Example: strings.Join([]string{
			"slack issue-resolver",
			"slack issue-resolver github-accessibility",
			"slack issue-resolver tenant-picker-performance",
			"slack issue-resolver button-name",
		}, "\n"),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResolverCommand(cmd, clients, args)
		},
	}

	cmd.Flags().StringVarP(&flags.format, "format", "f", "json", "output format: json or text")
	return cmd
}

func runResolverCommand(cmd *cobra.Command, clients *shared.ClientFactory, args []string) error {
	kind := "github-accessibility"
	if len(args) > 0 {
		kind = args[0]
	}

	var payload any
	switch kind {
	case "github-accessibility", "github-issue", "issues":
		payload = issueresolver.NewGitHubIssuePageAudit()
	case "tenant-picker-performance", "tenant-picker", "performance":
		payload = issueresolver.NewTenantPickerPerformanceAnalysis()
	case "button-name", "button":
		payload = issueresolver.ButtonNameRule
	default:
		return fmt.Errorf("unsupported issue resolver audit type %q", kind)
	}

	if flags.format == "text" {
		cmd.Println(renderText(kind, payload))
		return nil
	}

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(out))
	return nil
}

func renderText(kind string, payload any) string {
	switch v := payload.(type) {
	case issueresolver.GitHubIssuePageAudit:
		summary := issueresolver.ScoreRules(v.Rules)
		return fmt.Sprintf("%s\n\nAccessibility score: %d\nWeighted summary: passed=%d/%d\n",
			kind, summary.Score, summary.PassedWeight, summary.TotalWeight)
	case issueresolver.AnalysisModel:
		return fmt.Sprintf("%s\n\nContext: %s\nDiagnostics: %s\nFindings: %d\nRecommendations: %d\n",
			kind, v.Context, v.Diagnostics, len(v.Findings), len(v.Recommendations))
	case issueresolver.LighthouseRule:
		return fmt.Sprintf("%s\n\nRule: %s\nWeight: %d\nPasses: %t\nRecommendation: %s\n",
			kind, v.Name, v.Weight, v.Passes, v.Recommendation)
	default:
		return fmt.Sprintf("%s\n", kind)
	}
}
