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

package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_IsFlagChanged(t *testing.T) {
	tests := map[string]struct {
		flag     string
		setFlag  bool
		expected bool
	}{
		"returns true when flag is set": {
			flag:     "app-id",
			setFlag:  true,
			expected: true,
		},
		"returns false when flag exists but not set": {
			flag:     "app-id",
			setFlag:  false,
			expected: false,
		},
		"returns false when flag does not exist": {
			flag:     "nonexistent",
			setFlag:  false,
			expected: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("app-id", "", "app ID")
			if tc.setFlag && tc.flag == "app-id" {
				_ = cmd.Flags().Set("app-id", "A12345")
			}
			result := IsFlagChanged(cmd, tc.flag)
			assert.Equal(t, tc.expected, result)
		})
	}
}
