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

package hooks

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SDKCLIConfig_Exists(t *testing.T) {
	tests := map[string]struct {
		sdkCLIConfig  SDKCLIConfig
		expectedError error
		exists        bool
	}{
		"an initialized sdk configuration exists": {
			sdkCLIConfig: SDKCLIConfig{WorkingDirectory: "/path/to/project"},
			exists:       true,
		},
		"an uninitialized sdk configuration does not exist": {
			sdkCLIConfig:  SDKCLIConfig{WorkingDirectory: ""},
			expectedError: slackerror.New(slackerror.ErrInvalidSlackProjectDirectory),
			exists:        false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			exists, err := tc.sdkCLIConfig.Exists()
			require.Equal(t, tc.expectedError, err)
			require.Equal(t, tc.exists, exists)
		})
	}
}

func Test_SDKCLIConfig_JSONSchema(t *testing.T) {
	// lookup returns the schema value at the path of nested object keys
	lookup := func(t *testing.T, raw json.RawMessage, path ...string) json.RawMessage {
		for _, key := range path {
			var obj map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(raw, &obj))
			require.Contains(t, obj, key)
			raw = obj[key]
		}
		return raw
	}
	properties := func(t *testing.T, raw json.RawMessage) []string {
		var props map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(lookup(t, raw, "properties"), &props))
		return slices.Collect(maps.Keys(props))
	}
	jsonTags := func(v any) []string {
		tags := []string{}
		typ := reflect.TypeOf(v)
		for i := range typ.NumField() {
			if tag := typ.Field(i).Tag.Get("json"); tag != "" {
				tags = append(tags, strings.Split(tag, ",")[0])
			}
		}
		return tags
	}

	schema, err := os.ReadFile(filepath.Join("..", "..", "schemas", "hooks.schema.json"))
	require.NoError(t, err)
	var config SDKCLIConfig

	assert.ElementsMatch(t, append(jsonTags(config), "$schema"), properties(t, schema))
	assert.ElementsMatch(t, jsonTags(config.Hooks), properties(t, lookup(t, schema, "properties", "hooks")))
	assert.ElementsMatch(t, jsonTags(config.Config), properties(t, lookup(t, schema, "properties", "config")))
	assert.ElementsMatch(t, jsonTags(WatchOpts{}), properties(t, lookup(t, schema, "properties", "config", "properties", "watch")))
	assert.ElementsMatch(t, jsonTags(ManifestWatchOpts{}), properties(t, lookup(t, schema, "definitions", "watcher")))
	assert.ElementsMatch(t, jsonTags(AppWatchOpts{}), properties(t, lookup(t, schema, "definitions", "watcher")))
}

func Test_ProtocolResolution(t *testing.T) {
	tests := map[string]struct {
		config SDKCLIConfig
		check  func(t *testing.T, p Protocol)
	}{
		"Returns the first valid protocol version": {
			config: SDKCLIConfig{Config: struct {
				Watch                WatchOpts        `json:"watch,omitempty"`
				SDKManagedConnection bool             `json:"sdk-managed-connection-enabled,omitempty"`
				TriggerPaths         []string         `json:"trigger-paths,omitempty"`
				SupportedProtocols   ProtocolVersions `json:"protocol-version,omitempty"`
			}{
				SupportedProtocols: ProtocolVersions{
					"fake-news",
					HookProtocolV2,
					"news-fake",
					HookProtocolDefault,
				},
			}},
			check: func(t *testing.T, p Protocol) {
				assert.Equal(t, p, HookProtocolV2)
			},
		},
		"Returns default config if no valid protocols are present": {
			config: SDKCLIConfig{Config: struct {
				Watch                WatchOpts        `json:"watch,omitempty"`
				SDKManagedConnection bool             `json:"sdk-managed-connection-enabled,omitempty"`
				TriggerPaths         []string         `json:"trigger-paths,omitempty"`
				SupportedProtocols   ProtocolVersions `json:"protocol-version,omitempty"`
			}{
				SupportedProtocols: ProtocolVersions{
					"fake-news",
				},
			}},
			check: func(t *testing.T, p Protocol) {
				assert.Equal(t, p, HookProtocolDefault)
			},
		},
		"Returns default config if no protocols are present": {
			config: SDKCLIConfig{},
			check: func(t *testing.T, p Protocol) {
				assert.Equal(t, p, HookProtocolDefault)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			protocol := tc.config.Config.SupportedProtocols.Preferred()
			tc.check(t, protocol)
		})
	}
}

func Test_WatchOpts_IsAvailable(t *testing.T) {
	tests := map[string]struct {
		watchOpts           *WatchOpts
		expectedIsAvailable bool
	}{
		"WatchOpts exists": {
			watchOpts:           &WatchOpts{},
			expectedIsAvailable: true,
		},
		"WatchOpts not exists": {
			watchOpts:           nil,
			expectedIsAvailable: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			returnedIsAvailable := tc.watchOpts.IsAvailable()
			require.Equal(t, tc.expectedIsAvailable, returnedIsAvailable)
		})
	}
}

func Test_WatchOpts_GetManifestWatchConfig(t *testing.T) {
	tests := map[string]struct {
		watchOpts       *WatchOpts
		expectedPaths   []string
		expectedRegex   string
		expectedEnabled bool
	}{
		"Nil WatchOpts pointer": {
			watchOpts:       nil,
			expectedPaths:   nil,
			expectedRegex:   "",
			expectedEnabled: false,
		},
		"Nested manifest config": {
			watchOpts: &WatchOpts{
				Manifest: &ManifestWatchOpts{
					Paths:       []string{"manifest.json", "workflows/"},
					FilterRegex: "\\.json$",
				},
			},
			expectedPaths:   []string{"manifest.json", "workflows/"},
			expectedRegex:   "\\.json$",
			expectedEnabled: true,
		},
		"Legacy flat config": {
			watchOpts: &WatchOpts{
				Paths:       []string{"manifest.json", "src/"},
				FilterRegex: "\\.(json|ts)$",
			},
			expectedPaths:   []string{"manifest.json", "src/"},
			expectedRegex:   "\\.(json|ts)$",
			expectedEnabled: true,
		},
		"Nested config takes precedence over legacy": {
			watchOpts: &WatchOpts{
				Paths:       []string{"old-path/"},
				FilterRegex: "old-regex",
				Manifest: &ManifestWatchOpts{
					Paths:       []string{"new-path/"},
					FilterRegex: "new-regex",
				},
			},
			expectedPaths:   []string{"new-path/"},
			expectedRegex:   "new-regex",
			expectedEnabled: true,
		},
		"Empty nested manifest config": {
			watchOpts: &WatchOpts{
				Manifest: &ManifestWatchOpts{
					Paths: []string{},
				},
			},
			expectedPaths:   []string{},
			expectedRegex:   "",
			expectedEnabled: false,
		},
		"Empty legacy config": {
			watchOpts: &WatchOpts{
				Paths: []string{},
			},
			expectedPaths:   []string{},
			expectedRegex:   "",
			expectedEnabled: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			paths, regex, enabled := tt.watchOpts.GetManifestWatchConfig()
			assert.Equal(t, tt.expectedPaths, paths)
			assert.Equal(t, tt.expectedRegex, regex)
			assert.Equal(t, tt.expectedEnabled, enabled)
		})
	}
}

func Test_WatchOpts_GetAppWatchConfig(t *testing.T) {
	tests := map[string]struct {
		watchOpts       *WatchOpts
		expectedPaths   []string
		expectedRegex   string
		expectedEnabled bool
	}{
		"Nil WatchOpts pointer": {
			watchOpts:       nil,
			expectedPaths:   nil,
			expectedRegex:   "",
			expectedEnabled: false,
		},
		"Nested app config": {
			watchOpts: &WatchOpts{
				App: &AppWatchOpts{
					Paths:       []string{"src/", "functions/"},
					FilterRegex: "\\.(ts|js)$",
				},
			},
			expectedPaths:   []string{"src/", "functions/"},
			expectedRegex:   "\\.(ts|js)$",
			expectedEnabled: true,
		},
		"Legacy config does not enable app watching": {
			watchOpts: &WatchOpts{
				Paths:       []string{"manifest.json", "src/"},
				FilterRegex: "\\.(json|ts)$",
			},
			expectedPaths:   nil,
			expectedRegex:   "",
			expectedEnabled: false,
		},
		"Empty nested app config": {
			watchOpts: &WatchOpts{
				App: &AppWatchOpts{
					Paths: []string{},
				},
			},
			expectedPaths:   []string{},
			expectedRegex:   "",
			expectedEnabled: false,
		},
		"Nil app config": {
			watchOpts:       &WatchOpts{},
			expectedPaths:   nil,
			expectedRegex:   "",
			expectedEnabled: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			paths, regex, enabled := tt.watchOpts.GetAppWatchConfig()
			assert.Equal(t, tt.expectedPaths, paths)
			assert.Equal(t, tt.expectedRegex, regex)
			assert.Equal(t, tt.expectedEnabled, enabled)
		})
	}
}

func Test_WatchOpts_String(t *testing.T) {
	tests := map[string]struct {
		watchOpts      WatchOpts
		expectedString string
	}{
		"Nested config with both manifest and app": {
			watchOpts: WatchOpts{
				Manifest: &ManifestWatchOpts{
					Paths:       []string{"manifest.json"},
					FilterRegex: "\\.json$",
				},
				App: &AppWatchOpts{
					Paths:       []string{"src/", "functions/"},
					FilterRegex: "\\.(ts|js)$",
				},
			},
			expectedString: "{Manifest:{Paths:[manifest.json] FilterRegex:\\.json$} App:{Paths:[src/ functions/] FilterRegex:\\.(ts|js)$}}",
		},
		"Nested manifest only": {
			watchOpts: WatchOpts{
				Manifest: &ManifestWatchOpts{
					Paths:       []string{"manifest.json"},
					FilterRegex: "\\.json$",
				},
			},
			expectedString: "{Manifest:{Paths:[manifest.json] FilterRegex:\\.json$}}",
		},
		"Nested app only": {
			watchOpts: WatchOpts{
				App: &AppWatchOpts{
					Paths:       []string{"src/"},
					FilterRegex: "\\.(ts|js)$",
				},
			},
			expectedString: "{App:{Paths:[src/] FilterRegex:\\.(ts|js)$}}",
		},
		"Legacy config": {
			watchOpts: WatchOpts{
				Paths:       []string{"manifest.json", "src/"},
				FilterRegex: "\\.(json|ts)$",
			},
			expectedString: "{Paths:[manifest.json src/] FilterRegex:\\.(json|ts)$}",
		},
		"Empty config": {
			watchOpts:      WatchOpts{},
			expectedString: "{}",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.watchOpts.String()
			assert.Equal(t, tt.expectedString, result)
		})
	}
}
