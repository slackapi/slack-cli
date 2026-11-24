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

package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Disabled(t *testing.T) {
	tests := map[string]struct {
		env      map[string]string
		expected bool
	}{
		"enabled by default":                  {env: map[string]string{}, expected: false},
		"disabled by SLACK_DISABLE_TELEMETRY": {env: map[string]string{"SLACK_DISABLE_TELEMETRY": "true"}, expected: true},
		"enabled when set to false":           {env: map[string]string{"SLACK_DISABLE_TELEMETRY": "false"}, expected: false},
		"enabled when set to 0":               {env: map[string]string{"SLACK_DISABLE_TELEMETRY": " 0 "}, expected: false},
		"disabled by any SLACK_TEST_VERSION":  {env: map[string]string{"SLACK_TEST_VERSION": "v1.0.0"}, expected: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Disabled(func(key string) string { return tc.env[key] }))
		})
	}
}
