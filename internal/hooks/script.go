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
	"strings"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

// HookScript maps to specific hook commands listed out in hooks.json
type HookScript struct {
	Command string
	Name    string
}

// IsAvailable returns true when the HookScript.Command exists
func (s HookScript) IsAvailable() bool {
	return strings.TrimSpace(s.Command) != ""
}

// UnmarshalJSON implements the Unmarshaller interface so that HookScript
// objects can decode themselves.
//
// The Name field should be dynamically overwritten at SDK Config initialization
// time.
//
// Read more: https://pkg.go.dev/encoding/json#Unmarshaler
func (s *HookScript) UnmarshalJSON(data []byte) error {
	var cmd string
	if err := json.Unmarshal(data, &cmd); err != nil {
		return err
	}
	s.Command = cmd
	return nil
}

// Get checks if the hook command has been initialised and returns the command
// string
func (s HookScript) Get() (string, error) {
	if !s.IsAvailable() {
		var err = slackerror.New(slackerror.ErrSDKHookNotFound).WithMessage("The command for '%s' was not found", s.Name)
		return "", err
	}
	return s.Command, nil
}
