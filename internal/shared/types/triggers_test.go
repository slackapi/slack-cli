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

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Triggers_IsKnownType(t *testing.T) {
	tests := map[string]struct {
		trigger      *Trigger
		expectedBool bool
	}{
		"Trigger type is TriggerTypeShortcut": {
			trigger:      &Trigger{Type: TriggerTypeShortcut},
			expectedBool: true,
		},
		"Trigger type is TriggerTypeSlashCommand": {
			trigger:      &Trigger{Type: TriggerTypeSlashCommand},
			expectedBool: true,
		},
		"Trigger type is TriggerTypeMessageShortcut": {
			trigger:      &Trigger{Type: TriggerTypeMessageShortcut},
			expectedBool: true,
		},
		"Trigger type is TriggerTypeEvent": {
			trigger:      &Trigger{Type: TriggerTypeEvent},
			expectedBool: true,
		},
		"Trigger type is TriggerTypeWebhook": {
			trigger:      &Trigger{Type: TriggerTypeWebhook},
			expectedBool: true,
		},
		"Trigger type is TriggerTypeScheduled": {
			trigger:      &Trigger{Type: TriggerTypeScheduled},
			expectedBool: true,
		},
		"Trigger type is invalid": {
			trigger:      &Trigger{Type: "pickle pie"},
			expectedBool: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedBool := tc.trigger.IsKnownType()
			require.Equal(t, tc.expectedBool, returnedBool)
		})
	}
}
