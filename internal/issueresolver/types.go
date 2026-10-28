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

// LighthouseRule models a single Lighthouse accessibility audit.
// Each audit is treated as a pass/fail rule, and the overall score is the
// weighted aggregate of each rule's assigned impact.
type LighthouseRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Weight      int      `json:"weight"`
	Passes      bool     `json:"passes"`
	Evidence    []string `json:"evidence,omitempty"`

	FailingElements []string `json:"failingElements,omitempty"`
	Recommendation  string   `json:"recommendation,omitempty"`
	ValidPatterns   []string `json:"validPatterns,omitempty"`
	InvalidPatterns []string `json:"invalidPatterns,omitempty"`
}

// ScoreSummary represents the weighted Lighthouse combination of all rule results.
type ScoreSummary struct {
	PassedWeight int `json:"passedWeight"`
	TotalWeight  int `json:"totalWeight"`
	Score        int `json:"score"`
}

// ScoreRules returns a weighted score for a list of Lighthouse accessibility rules.
func ScoreRules(rules []LighthouseRule) ScoreSummary {
	totalWeight := 0
	passedWeight := 0

	for _, rule := range rules {
		totalWeight += rule.Weight
		if rule.Passes {
			passedWeight += rule.Weight
		}
	}

	score := 0
	if totalWeight > 0 {
		score = (passedWeight * 100) / totalWeight
	}

	return ScoreSummary{
		PassedWeight: passedWeight,
		TotalWeight:  totalWeight,
		Score:        score,
	}
}
