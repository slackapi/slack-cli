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

type ManifestSource string

const (
	ManifestSourceLocal  ManifestSource = "local"
	ManifestSourceRemote ManifestSource = "remote"
)

// Equals returns true if the manifest source is the same
func (ms ManifestSource) Equals(is ManifestSource) bool {
	return ms == is
}

// Exists returns true if the manifest source is set
func (ms ManifestSource) Exists() bool {
	return ms != ""
}

// String returns the string value of a manifest source
func (ms ManifestSource) String() string {
	return string(ms)
}

type ManifestConfig struct {
	// Source of the manifest using either "local" or "remote" values
	Source string `json:"source,omitempty"`
}
