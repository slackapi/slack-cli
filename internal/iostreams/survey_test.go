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

package iostreams

import (
	"fmt"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestPasswordPrompt(t *testing.T) {
	tests := map[string]struct {
		FlagChanged   bool
		FlagValue     string
		Required      bool
		IsInteractive bool
		ExpectedError *slackerror.Error
		ExpectedValue string
	}{
		"error if no flag is set": {
			IsInteractive: false,
			ExpectedError: slackerror.New(slackerror.ErrPrompt),
		},
		"error if the flag is missing a required value": {
			FlagChanged:   true,
			FlagValue:     "",
			Required:      true,
			ExpectedError: slackerror.New(slackerror.ErrMissingFlag),
		},
		"allow an empty flag value if not required": {
			FlagChanged:   true,
			FlagValue:     "",
			Required:      false,
			ExpectedValue: "",
		},
		"use the provided flag value": {
			FlagChanged:   true,
			FlagValue:     "something secret",
			ExpectedValue: "something secret",
		},
		// "values can be entered interactively": {
		// IsInteractive: true,
		// },
	}

	var mockFlagValue string
	pflag.StringVar(&mockFlagValue, "mockedflag", "", "mock usage")
	mockFlag := pflag.Lookup("mockedflag")

	interactiveStdout := &slackdeps.FileMock{
		FileInfo: &slackdeps.FileInfoCharDevice{},
	}
	nonInteractiveStdout := &slackdeps.FileMock{
		FileInfo: &slackdeps.FileInfoNamedPipe{},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			if tt.IsInteractive {
				osMock.On("Stdout").Return(interactiveStdout)
			} else {
				osMock.On("Stdout").Return(nonInteractiveStdout)
			}
			config := config.NewConfig(fsMock, osMock)
			ioStreams := NewIOStreams(config, fsMock, osMock)

			if tt.FlagChanged {
				mockFlag.Changed = true
				_ = mockFlag.Value.Set(tt.FlagValue)
			} else {
				mockFlag.Changed = false
				_ = mockFlag.Value.Set("")
			}

			selection, err := ioStreams.PasswordPrompt(ctx, "Enter a password", PasswordPromptConfig{
				Flag:     mockFlag,
				Required: tt.Required,
			})

			if tt.ExpectedError != nil {
				assert.Equal(t, tt.ExpectedError.Code, slackerror.ToSlackError(err).Code)
				if tt.ExpectedError.Code == slackerror.ErrPrompt {
					assert.Contains(t, err.Error(), fmt.Sprintf("--%s", mockFlag.Name))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, selection.Value, tt.ExpectedValue)
				if tt.FlagChanged {
					assert.Equal(t, selection.Flag, true)
				} else {
					assert.Equal(t, selection.Prompt, true)
				}
			}
		})
	}
}
