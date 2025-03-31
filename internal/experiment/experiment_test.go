// Copyright 2022-2025 Salesforce, Inc.
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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Includes(t *testing.T) {
	// Test valid experiment
	require.Equal(t, true, Includes(Experiment(Placeholder)))

	// Test expected experiments
	require.Equal(t, true, Includes(Experiment("bolt")))
	require.Equal(t, true, Includes(Experiment("read-only-collaborators")))

	// Test invalid experiment
	require.Equal(t, false, Includes(Experiment("should-fail")))
}

func Test_AllExperimentsListedAreValid(t *testing.T) {
	for _, exp := range AllExperiments {
		require.Equal(t, true, isValid(string(exp)))
	}
}

func Test_EnabledExperimentsListedAreValid(t *testing.T) {
	for _, exp := range EnabledExperiments {
		require.Equal(t, true, isValid(string(exp)))
	}
}

func Test_IsValid(t *testing.T) {
	tableTests := map[string]struct {
		experiment  string
		expectedRes bool
	}{
		"accepts valid experiment format: no hyphen": {
			experiment:  "match",
			expectedRes: true,
		},
		"accepts valid experiment format: kebab-case": {
			experiment:  "should-match",
			expectedRes: true,
		},
		"accepts valid experiment format: kebab-case-multi": {
			experiment:  "should-also-match",
			expectedRes: true,
		},
		"accepts valid experiment format: kebab-case-numerical": {
			experiment:  "should-also-match3",
			expectedRes: true,
		},
		"rejects invalid experiment format: capitalized": {
			experiment:  "Nomatch",
			expectedRes: false,
		},
		"rejects invalid experiment format: capitalized-with-hyphen": {
			experiment:  "No-match",
			expectedRes: false,
		},
		"rejects invalid experiment format: hyphen-prefix-or-suffix": {
			experiment:  "-no-match-",
			expectedRes: false,
		},
		"rejects invalid experiment format: mixed-case-capital": {
			experiment:  "NoMatch",
			expectedRes: false,
		},
		"rejects invalid experiment format: mixed-case-lower": {
			experiment:  "noMatch",
			expectedRes: false,
		},
		"rejects invalid experiment format: snake_case": {
			experiment:  "no_match",
			expectedRes: false,
		},
	}
	for name, tt := range tableTests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.expectedRes, isValid(tt.experiment))
		})
	}
}
