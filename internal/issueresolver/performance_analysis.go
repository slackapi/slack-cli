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

// MetricBreakdown captures a high-level performance metric and its contribution.
type MetricBreakdown struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	Contribution string `json:"contribution"`
}

// Finding models a single issue observed during a web performance analysis.
type Finding struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      string   `json:"impact,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// Recommendation models a specific remediation step.
type Recommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Snippet     string `json:"snippet,omitempty"`
}

// AnalysisModel represents an AI-generated analysis payload for a page or issue.
type AnalysisModel struct {
	Context        string          `json:"context"`
	Diagnostics    string          `json:"diagnostics"`
	Metrics        []MetricBreakdown `json:"metrics"`
	Findings       []Finding       `json:"findings"`
	Recommendations []Recommendation `json:"recommendations"`
}

// NewTenantPickerPerformanceAnalysis creates a concrete example of a render-delay analysis.
func NewTenantPickerPerformanceAnalysis() AnalysisModel {
	return AnalysisModel{
		Context: "Analysis of the admin.openai.com tenant-picker page performance, focusing on a Largest Contentful Paint (LCP) of 1,263 ms and the associated render delays.",
		Diagnostics: "The LCP breakdown shows that while the network response (TTFB) was efficient, the browser experienced significant delays before painting the primary content.",
		Metrics: []MetricBreakdown{
			{Name: "LCP Total", Value: "1,263 ms", Contribution: "100%"},
			{Name: "Time to First Byte (TTFB)", Value: "196 ms", Contribution: "15.5%"},
			{Name: "Element Render Delay", Value: "1,068 ms", Contribution: "84.5%"},
		},
		Findings: []Finding{
			{
				Title:       "LCP Element",
				Description: "h1.text-[28px].leading-7.text-primary (Node ID 68).",
				Impact:      "Primary content is not painted promptly.",
			},
			{
				Title:       "Main Thread Contention",
				Description: "Long tasks blocked the UI, specifically a 174 ms task and a 105 ms task.",
				Impact:      "The browser could not interleave painting and script execution efficiently.",
			},
			{
				Title:       "DOM Manipulation",
				Description: "appendChild operations consumed 117 ms of self-time, indicating heavy client-side rendering or hydration.",
				Impact:      "The DOM is being assembled too late in the critical path.",
			},
			{
				Title:       "Resource Impact",
				Description: "Scripts from oaistatic.com consumed 414 ms of main thread time.",
				Impact:      "Third-party script execution is delaying the first paint.",
			},
		},
		Recommendations: []Recommendation{
			{
				Title:       "Yield to Main Thread",
				Description: "Break up long JavaScript execution blocks to allow the browser to paint the LCP element between tasks.",
				Snippet: `async function initializeApp() {
  await scheduler.yield();
  performHeavyDOMOperations();
}`,
			},
			{
				Title:       "Inline Critical CSS",
				Description: "Inline the styles necessary for the h1 LCP element in the <head> to bypass the network request for external CSS.",
			},
			{
				Title:       "Font Rendering",
				Description: "Apply font-display: swap to custom fonts to ensure LCP text is visible immediately using fallback fonts.",
			},
			{
				Title:       "Script Prioritization",
				Description: "Audit scripts like standalone-entry.js and datadog.client.js. Defer any scripts not required for the initial visual state.",
			},
		},
	}
}
