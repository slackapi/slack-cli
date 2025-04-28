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

// Protocol versions understood by the CLI
type Protocol string

const (
	HookProtocolDefault Protocol = "default"
	HookProtocolV2      Protocol = "message-boundaries"
)

func (p Protocol) String() string {
	return string(p)
}

// Valid returns true if this protocol is understood by the CLI.
func (p Protocol) Valid() bool {
	return p == HookProtocolDefault || p == HookProtocolV2
}
