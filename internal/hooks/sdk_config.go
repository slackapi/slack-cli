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

package hooks

import (
	"strings"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

// SDKCLIConfig contains configuration for communication between the CLI and the SDK.
// It is set by merging the app's local hooks.json and the response from the `get-hooks` hook.
type SDKCLIConfig struct {
	Runtime string `json:"runtime,omitempty"` // Optional, runtime version e.g. deno, deno1.x
	Hooks   struct {
		BuildProject  HookScript `json:"build,omitempty"`
		CheckUpdate   HookScript `json:"check-update,omitempty"`
		Deploy        HookScript `json:"deploy,omitempty"`
		Doctor        HookScript `json:"doctor,omitempty"`
		GetHooks      HookScript `json:"get-hooks,omitempty"`
		GetManifest   HookScript `json:"get-manifest,omitempty"`
		GetTrigger    HookScript `json:"get-trigger,omitempty"`
		InstallUpdate HookScript `json:"install-update,omitempty"`
		Start         HookScript `json:"start,omitempty"`
	} `json:"hooks,omitempty"`
	Config struct {
		Watch                WatchOpts        `json:"watch,omitempty"`
		SDKManagedConnection bool             `json:"sdk-managed-connection-enabled,omitempty"`
		TriggerPaths         []string         `json:"trigger-paths,omitempty"`
		SupportedProtocols   ProtocolVersions `json:"protocol-version,omitempty"`
	} `json:"config,omitempty"`

	WorkingDirectory string
}

// Exists returns true when the SDKCLIConfig was successfully loaded, otherwise false with an error
func (s *SDKCLIConfig) Exists() (error, bool) {
	if strings.TrimSpace(s.WorkingDirectory) == "" {
		return slackerror.New(slackerror.ErrInvalidSlackProjectDirectory), false
	}
	return nil, true
}

type ProtocolVersions []Protocol

// Preferred returns the first valid protocol present in the SDK config.
// Lower indices in the array of protocols received from the SDK config are interpreted as more
// recent or preferred protocol versions.
func (pv ProtocolVersions) Preferred() Protocol {
	for _, p := range pv {
		if p.Valid() {
			return p
		}
	}
	return HookProtocolDefault
}

type WatchOpts struct {
	FilterRegex string   `json:"filter-regex,omitempty"`
	Paths       []string `json:"paths,omitempty"`
}

// IsAvailable returns if watch options exist
func (w *WatchOpts) IsAvailable() bool {
	return w != nil
}
