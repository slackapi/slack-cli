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

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_Hash_NewHash(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected Hash
	}{
		"empty inputs hash to a constant": {
			input:    "",
			expected: Hash("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"),
		},
		"json inputs can be encoded": {
			input:    `{"a":3,"b":true,"c":[]}`,
			expected: Hash("5feaedf58240eee75f431c31518fae862011fb6204e6fe3d91f25c4556934657"),
		},
		"json inputs sort keys for the same encoding": {
			input:    `{"b":true,"c":[],"a":3}`,
			expected: Hash("5feaedf58240eee75f431c31518fae862011fb6204e6fe3d91f25c4556934657"),
		},
		"text inputs can be encoded": {
			input:    `{"text`,
			expected: Hash("c8fe2e89bbafe8ca60e751ec27cd5ebf57562d452ebb9470fa6ccddd3c9de904"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hash := NewHash([]byte(tt.input))
			assert.Equal(t, tt.expected, hash)
		})
	}
}

func TestCache_Hash_Equals(t *testing.T) {
	tests := map[string]struct {
		a      Hash
		b      Hash
		equals bool
	}{
		"empty hashes are equal": {
			equals: true,
		},
		"matching hashes are equal": {
			a:      Hash("matching"),
			b:      Hash("matching"),
			equals: true,
		},
		"different hashes are not equal": {
			a:      Hash("different"),
			b:      Hash("not equal"),
			equals: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			left := tt.a.Equals(tt.b)
			right := tt.b.Equals(tt.a)
			assert.Equal(t, tt.equals, left)
			assert.Equal(t, tt.equals, right)
		})
	}
}
