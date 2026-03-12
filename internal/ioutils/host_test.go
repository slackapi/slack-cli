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

package ioutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetHostname(t *testing.T) {
	t.Run("returns a non-empty hashed hostname", func(t *testing.T) {
		hostname := GetHostname()
		assert.NotEmpty(t, hostname)
		// The hostname should be hashed, not the raw hostname
		// It should not be "unknown" on a normal system
		assert.NotEqual(t, "unknown", hostname)
	})

	t.Run("returns consistent results", func(t *testing.T) {
		hostname1 := GetHostname()
		hostname2 := GetHostname()
		assert.Equal(t, hostname1, hostname2)
	})
}
