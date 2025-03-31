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

package style

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Spinner(t *testing.T) {
	defer func() {
		isSpinAllowed = false
	}()
	tests := map[string]struct {
		spins bool
	}{
		"starts to spin with updated messages": {
			spins: true,
		},
		"prints messages without spinning": {
			spins: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			buff := &bytes.Buffer{}
			ToggleSpinner(tt.spins)
			spinner := NewSpinner(buff)
			active := spinner.Active()
			assert.False(t, spinner.active)
			assert.False(t, active)
			spinner.Update("waiting", "").Start()
			status, active := spinner.Status()
			assert.True(t, spinner.active)
			assert.True(t, active)
			assert.Equal(t, "waiting", status)
			spinner.Update("changed", "")
			status, active = spinner.Status()
			assert.True(t, spinner.active)
			assert.True(t, active)
			assert.Equal(t, "changed", status)
			spinner.Update("ending!", "").Stop()
			status, active = spinner.Status()
			assert.False(t, spinner.active)
			assert.False(t, active)
			assert.Equal(t, "ending!", status)
			if !tt.spins {
				assert.Equal(t, buff.String(), "waiting\nchanged\nending!\n")
			} else {
				// No output is gathered because the test is not run in a terminal setting.
				// Adding a delay between the above spins did not change this, but this can
				// be updated if outputs can be collected for tests!
				//
				// https://github.com/briandowns/spinner/blob/55430861f77b20de3f456775b00c706519464f36/spinner.go#L502-L506
				assert.Equal(t, buff.String(), "")
			}
		})
	}
}
