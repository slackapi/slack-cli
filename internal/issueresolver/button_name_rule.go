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

// ButtonNameRule models the Lighthouse button-name accessibility rule.
var ButtonNameRule = LighthouseRule{
	ID:          "button-name",
	Name:        "Buttons have an accessible name",
	Description: "Buttons must have discernible text that clearly describes the destination, purpose, function, or action for screen reader users.",
	Weight:      10,
	Passes:      false,
	FailingElements: []string{
		`<button id="empty"></button>`,
		`<button id="val" value="Button Name"></button>`,
		`<button id="alempty" aria-label=""></button>`,
		`<button id="albmissing" aria-labelledby="nonexistent"></button>`,
		`<button id="albempty" aria-labelledby="emptydiv"></button>`,
		`<div id="emptydiv"></div>`,
		`<button id="buttonvalue" value="foo" tabindex="-1"></button>`,
	},
	Recommendation: "Add visible text, a non-empty aria-label, or an aria-labelledby reference that resolves to text readable by screen readers.",
	ValidPatterns: []string{
		`<button id="text">Name</button>`,
		`<button id="al" aria-label="Name"></button>`,
		`<button id="alb" aria-labelledby="labeldiv"></button><div id="labeldiv">Button label</div>`,
		`<button id="combo" aria-label="Aria Name">Name</button>`,
		`<button id="buttonTitle" title="Title"></button>`,
	},
	InvalidPatterns: []string{
		`<button id="empty"></button>`,
		`<button id="val" value="Button Name"></button>`,
		`<button id="alempty" aria-label=""></button>`,
		`<button id="albmissing" aria-labelledby="nonexistent"></button>`,
		`<button id="albempty" aria-labelledby="emptydiv"></button><div id="emptydiv"></div>`,
		`<button id="buttonvalue" value="foo" tabindex="-1"></button>`,
	},
}

// HeadingOrderRule models the Lighthouse heading-order accessibility rule.
var HeadingOrderRule = LighthouseRule{
	ID:          "heading-order",
	Name:        "Heading elements appear in a sequentially-descending order",
	Description: "Heading levels should follow a logical structure to support assistive technology navigation.",
	Weight:      3,
	Passes:      false,
	FailingElements: []string{
		`h4.mb-2 ("Want to contribute...")`,
	},
	Recommendation: "Use a sequential heading structure such as h1 > h2 > h3 without skipping levels.",
}

// LighthouseAccessibilityRules is the set of rule definitions used by the issue resolver.
var LighthouseAccessibilityRules = []LighthouseRule{
	ButtonNameRule,
	HeadingOrderRule,
}
