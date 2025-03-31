// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Config_ManifestSource_Equals(t *testing.T) {
	tests := map[string]struct {
		a        ManifestSource
		b        ManifestSource
		expected bool
	}{
		"matching project sources are equal": {
			a:        MANIFEST_SOURCE_LOCAL,
			b:        MANIFEST_SOURCE_LOCAL,
			expected: true,
		},
		"matching remote sources are equal": {
			a:        MANIFEST_SOURCE_REMOTE,
			b:        MANIFEST_SOURCE_REMOTE,
			expected: true,
		},
		"different manifest sources are not equal": {
			a:        MANIFEST_SOURCE_LOCAL,
			b:        MANIFEST_SOURCE_REMOTE,
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tt.a.Equals(tt.b)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Config_ManifestSource_Exists(t *testing.T) {
	tests := map[string]struct {
		a        ManifestSource
		expected bool
	}{
		"project source exists": {
			a:        MANIFEST_SOURCE_LOCAL,
			expected: true,
		},
		"remote source exists": {
			a:        MANIFEST_SOURCE_REMOTE,
			expected: true,
		},
		"unknown source exists": {
			a:        "unknonwn",
			expected: true,
		},
		"missing source does not exist": {
			a:        "",
			expected: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tt.a.Exists()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Config_ManifestSource_String(t *testing.T) {
	tests := map[string]struct {
		a        ManifestSource
		expected string
	}{
		"project manifest source is local": {
			a:        MANIFEST_SOURCE_LOCAL,
			expected: "local",
		},
		"remote manifest source is remote": {
			a:        MANIFEST_SOURCE_REMOTE,
			expected: "remote",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual := tt.a.String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
