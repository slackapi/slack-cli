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
	"testing"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/require"
)

func Test_HookScript_IsAvailable(t *testing.T) {
	tests := map[string]struct {
		hookScript          *HookScript
		expectedIsAvailable bool
	}{
		"Available when command exists": {
			hookScript:          &HookScript{Name: "start", Command: "npm start"},
			expectedIsAvailable: true,
		},
		"Available when command exists with whitespace/newline padding": {
			hookScript:          &HookScript{Name: "start", Command: "  npm start  \n  "},
			expectedIsAvailable: true,
		},
		"Not available when no command": {
			hookScript:          &HookScript{Name: "start", Command: ""},
			expectedIsAvailable: false,
		},
		"Not available when only whitespace/newline padding": {
			hookScript:          &HookScript{Name: "start", Command: "   \n\n \n   "},
			expectedIsAvailable: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.hookScript.IsAvailable(), tt.expectedIsAvailable)
		})
	}
}

func Test_HookScript_UnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		blob                string
		expectedErrorType   error
		expectedHookScripts []HookScript
	}{
		"Unmarshal HookScript Array": {
			blob:                `["hook-one", "hook-two", "hook-three"]`,
			expectedErrorType:   nil,
			expectedHookScripts: []HookScript{{Command: "hook-one"}, {Command: "hook-two"}, {Command: "hook-three"}},
		},
		"Unmarshal Error": {
			blob:                `[1, 2, 3]`,
			expectedErrorType:   &json.UnmarshalTypeError{},
			expectedHookScripts: []HookScript{{Command: ""}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var hookScripts []HookScript
			err := json.Unmarshal([]byte(tt.blob), &hookScripts)
			require.IsType(t, err, tt.expectedErrorType)
			require.Equal(t, tt.expectedHookScripts, hookScripts)
		})
	}
}

func Test_HookScript_Get(t *testing.T) {
	tests := map[string]struct {
		hookScript      *HookScript
		expectedError   error
		expectedCommand string
	}{
		"Command exists": {
			hookScript:      &HookScript{Name: "start", Command: "npm start"},
			expectedError:   nil,
			expectedCommand: "npm start",
		},
		"Command does not exist": {
			hookScript:      &HookScript{Name: "start", Command: ""},
			expectedError:   slackerror.New(slackerror.ErrSDKHookNotFound).WithMessage("The command for 'start' was not found"),
			expectedCommand: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd, err := tt.hookScript.Get()
			require.Equal(t, err, tt.expectedError)
			require.Equal(t, cmd, tt.expectedCommand)
		})
	}
}
