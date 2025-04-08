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

package function

import (
	"testing"

	"github.com/slackapi/slack-cli/internal/hooks"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFunction_ChooseFunctionPrompt(t *testing.T) {
	tests := map[string]struct {
		withFlag      bool
		flag          string
		withPrompt    bool
		selectIndex   int
		withError     error
		expectedValue string
		expectedError error
	}{
		"errors without a flag and interactivity": {
			withFlag:      false,
			withError:     slackerror.New(slackerror.ErrPrompt),
			expectedError: slackerror.New(slackerror.ErrPrompt),
		},
		"returns the provided flag value": {
			withFlag:      true,
			flag:          "callback_name",
			expectedValue: "callback_name",
		},
		"selects the first ordered value": {
			withPrompt:    true,
			selectIndex:   0,
			expectedValue: "alphabet_function",
		},
		"selects the last ordered value": {
			withPrompt:    true,
			selectIndex:   3,
			expectedValue: "ordered_function",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.IO.On("SelectPrompt", mock.Anything, "Choose a function", mock.Anything, iostreams.MatchPromptConfig(iostreams.SelectPromptConfig{
				Flag: clientsMock.Config.Flags.Lookup("name"),
			})).Return(iostreams.SelectPromptResponse{
				Flag:   tt.withFlag,
				Prompt: tt.withPrompt,
				Option: tt.flag,
				Index:  tt.selectIndex,
			}, tt.withError)
			sdkConfigMock := hooks.NewSDKConfigMock()
			sdkConfigMock.Hooks.GetManifest.Command = "example"
			clientsMock.AddDefaultMocks()

			clients := shared.NewClientFactory(clientsMock.MockClientFactory(), func(clients *shared.ClientFactory) {
				clients.SDKConfig = sdkConfigMock
			})
			functions := []types.Function{
				{CallbackID: "alphabet_function"},
				{CallbackID: "hello_function"},
				{CallbackID: "ordered_function"},
				{CallbackID: "goodbye_function"},
			}
			value, err := chooseFunctionPrompt(ctx, clients, functions)
			if err != nil || tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}

}
