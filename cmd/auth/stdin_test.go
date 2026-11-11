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

package auth

import (
	"bytes"
	"testing"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ResolveTokenInput(t *testing.T) {
	tests := map[string]struct {
		value         string
		stdin         string
		stdinTTY      bool
		expectedToken string
		expectedError string
	}{
		"returns the value when it is a token": {
			value:         "xoxp-example",
			expectedToken: "xoxp-example",
		},
		"returns an empty value unchanged": {
			value:         "",
			expectedToken: "",
		},
		"reads the token from standard input": {
			value:         "-",
			stdin:         "xoxp-example\n",
			expectedToken: "xoxp-example",
		},
		"errors when standard input is a terminal": {
			value:         "-",
			stdinTTY:      true,
			expectedError: "No token was provided on standard input",
		},
		"errors when standard input is empty": {
			value:         "-",
			stdin:         "  \n",
			expectedError: "No token was provided on standard input",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("IsStdinTTY").Return(tc.stdinTTY)
			clientsMock.IO.Stdin = bytes.NewBufferString(tc.stdin)
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			token, err := ResolveTokenInput(clients, tc.value)
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, slackerror.ErrMissingInput, slackerror.ToSlackError(err).Code)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedToken, token)
		})
	}
}
