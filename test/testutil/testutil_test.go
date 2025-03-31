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

package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Update_ContainsSemVer(t *testing.T) {
	tests := map[string]struct {
		version  string
		expected bool
	}{
		"valid semver is contained": {
			version:  "1.2.3",
			expected: true,
		},
		"invalid semver is not contained": {
			version:  "dev",
			expected: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := ContainsSemVer(tt.version)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
