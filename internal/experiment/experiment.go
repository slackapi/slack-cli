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

package experiment

import (
	"regexp"
	"slices"
)

type Experiment string

// When developing features that you would like gated behind an experiment
// add yours here in alphabetical order.
//
// To configure an experiment on at the command level, you may use the experiment
// flag --experiment, -e with value(s) in comma-separated format.
//
// e.g. --experiment=first-toggle,second-toggle

const (
	// Lipgloss experiment shows pretty styles.
	Lipgloss Experiment = "lipgloss"

	// Placeholder experiment is a placeholder for testing and does nothing... or does it?
	Placeholder Experiment = "placeholder"

	// SetIcon experiment enables icon upload for non-hosted apps.
	SetIcon Experiment = "set-icon"
)

// AllExperiments is a list of all available experiments that can be enabled
// Please also add here 👇
var AllExperiments = []Experiment{
	Lipgloss,
	Placeholder,
	SetIcon,
}

// EnabledExperiments is a list of experiments that are permanently enabled
// Please also add here 👇
var EnabledExperiments = []Experiment{}

// Includes checks that a supplied experiment is included within AllExperiments
func Includes(expToCheck Experiment) bool {
	return slices.Contains(AllExperiments, expToCheck)
}

// isValid returns true if experiment follows standard
func isValid(experimentToCheck string) bool {
	experimentPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	match := experimentPattern.Match([]byte(experimentToCheck))
	return match
}
