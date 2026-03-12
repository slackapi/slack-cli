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

func Test_ConfirmPromptConfig(t *testing.T) {
	tests := map[string]struct {
		cfg      ConfirmPromptConfig
		required bool
	}{
		"required true": {
			cfg:      ConfirmPromptConfig{Required: true},
			required: true,
		},
		"required false": {
			cfg:      ConfirmPromptConfig{Required: false},
			required: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.required, tc.cfg.IsRequired())
			assert.Empty(t, tc.cfg.GetFlags())
		})
	}
}

func Test_DefaultSelectPromptConfig(t *testing.T) {
	cfg := DefaultSelectPromptConfig()
	assert.True(t, cfg.IsRequired())
	assert.Empty(t, cfg.GetFlags())
}

func Test_InputPromptConfig(t *testing.T) {
	tests := map[string]struct {
		cfg      InputPromptConfig
		required bool
	}{
		"required true": {
			cfg:      InputPromptConfig{Required: true, Placeholder: "hint"},
			required: true,
		},
		"required false": {
			cfg:      InputPromptConfig{Required: false},
			required: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.required, tc.cfg.IsRequired())
			assert.Empty(t, tc.cfg.GetFlags())
		})
	}
}

func Test_IOStreams_retrieveFlagValue(t *testing.T) {
	fsMock := slackdeps.NewFsMock()
	osMock := slackdeps.NewOsMock()
	cfg := config.NewConfig(fsMock, osMock)
	io := NewIOStreams(cfg, fsMock, osMock)

	tests := map[string]struct {
		flagset       []*pflag.Flag
		expectedFlag  bool
		expectedError string
	}{
		"nil flagset returns nil": {
			flagset:      nil,
			expectedFlag: false,
		},
		"no changed flags returns nil": {
			flagset: func() []*pflag.Flag {
				var v string
				fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
				fs.StringVar(&v, "x", "", "")
				return []*pflag.Flag{fs.Lookup("x")}
			}(),
			expectedFlag: false,
		},
		"one changed flag returns it": {
			flagset: func() []*pflag.Flag {
				var v string
				fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
				fs.StringVar(&v, "y", "", "")
				f := fs.Lookup("y")
				f.Changed = true
				return []*pflag.Flag{f}
			}(),
			expectedFlag: true,
		},
		"two changed flags returns error": {
			flagset: func() []*pflag.Flag {
				var v1, v2 string
				fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
				fs.StringVar(&v1, "a", "", "")
				fs.StringVar(&v2, "b", "", "")
				fa := fs.Lookup("a")
				fa.Changed = true
				fb := fs.Lookup("b")
				fb.Changed = true
				return []*pflag.Flag{fa, fb}
			}(),
			expectedError: slackerror.ErrMismatchedFlags,
		},
		"nil flag in set is skipped": {
			flagset:      []*pflag.Flag{nil},
			expectedFlag: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			flag, err := io.retrieveFlagValue(tc.flagset)
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				if tc.expectedFlag {
					assert.NotNil(t, flag)
				} else {
					assert.Nil(t, flag)
				}
			}
		})
	}
}

func Test_MultiSelectPromptConfig(t *testing.T) {
	tests := map[string]struct {
		cfg      MultiSelectPromptConfig
		required bool
	}{
		"required true": {
			cfg:      MultiSelectPromptConfig{Required: true},
			required: true,
		},
		"required false": {
			cfg:      MultiSelectPromptConfig{Required: false},
			required: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.required, tc.cfg.IsRequired())
			assert.Empty(t, tc.cfg.GetFlags())
		})
	}
}

func Test_PasswordPromptConfig(t *testing.T) {
	t.Run("without flag", func(t *testing.T) {
		cfg := PasswordPromptConfig{Required: true}
		assert.True(t, cfg.IsRequired())
		assert.Empty(t, cfg.GetFlags())
	})
	t.Run("with flag", func(t *testing.T) {
		var val string
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringVar(&val, "token", "", "token flag")
		flag := fs.Lookup("token")
		cfg := PasswordPromptConfig{Required: false, Flag: flag}
		assert.False(t, cfg.IsRequired())
		assert.Len(t, cfg.GetFlags(), 1)
		assert.Equal(t, "token", cfg.GetFlags()[0].Name)
	})
}

func Test_SelectPromptConfig(t *testing.T) {
	t.Run("no flags", func(t *testing.T) {
		cfg := SelectPromptConfig{Required: true}
		assert.True(t, cfg.IsRequired())
		assert.Empty(t, cfg.GetFlags())
	})
	t.Run("single flag", func(t *testing.T) {
		var val string
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringVar(&val, "app", "", "app flag")
		flag := fs.Lookup("app")
		cfg := SelectPromptConfig{Flag: flag}
		assert.Len(t, cfg.GetFlags(), 1)
	})
	t.Run("multiple flags", func(t *testing.T) {
		var v1, v2 string
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringVar(&v1, "a", "", "")
		fs.StringVar(&v2, "b", "", "")
		cfg := SelectPromptConfig{Flags: []*pflag.Flag{fs.Lookup("a"), fs.Lookup("b")}}
		assert.Len(t, cfg.GetFlags(), 2)
	})
}

func Test_SurveyOptions(t *testing.T) {
	tests := map[string]struct {
		cfg         PromptConfig
		expectedLen int
	}{
		"required config returns 5 options": {
			cfg:         ConfirmPromptConfig{Required: true},
			expectedLen: 5,
		},
		"optional config returns 5 options": {
			cfg:         ConfirmPromptConfig{Required: false},
			expectedLen: 5,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opts := SurveyOptions(tc.cfg)
			assert.Len(t, opts, tc.expectedLen)
		})
	}
}

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

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			if tc.IsInteractive {
				osMock.On("Stdout").Return(interactiveStdout)
			} else {
				osMock.On("Stdout").Return(nonInteractiveStdout)
			}
			config := config.NewConfig(fsMock, osMock)
			ioStreams := NewIOStreams(config, fsMock, osMock)

			if tc.FlagChanged {
				mockFlag.Changed = true
				_ = mockFlag.Value.Set(tc.FlagValue)
			} else {
				mockFlag.Changed = false
				_ = mockFlag.Value.Set("")
			}

			selection, err := ioStreams.PasswordPrompt(ctx, "Enter a password", PasswordPromptConfig{
				Flag:     mockFlag,
				Required: tc.Required,
			})

			if tc.ExpectedError != nil {
				assert.Equal(t, tc.ExpectedError.Code, slackerror.ToSlackError(err).Code)
				if tc.ExpectedError.Code == slackerror.ErrPrompt {
					assert.Contains(t, err.Error(), fmt.Sprintf("--%s", mockFlag.Name))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, selection.Value, tc.ExpectedValue)
				if tc.FlagChanged {
					assert.Equal(t, selection.Flag, true)
				} else {
					assert.Equal(t, selection.Prompt, true)
				}
			}
		})
	}
}
