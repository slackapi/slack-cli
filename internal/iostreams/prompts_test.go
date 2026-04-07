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
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptConfigs(t *testing.T) {
	mockFlag := &pflag.Flag{Name: "token"}
	mockFlag2 := &pflag.Flag{Name: "team"}

	tests := map[string]struct {
		cfg      PromptConfig
		required bool
		flags    int
	}{
		"ConfirmPromptConfig required": {
			cfg: ConfirmPromptConfig{
				Required: true,
			},
			required: true,
			flags:    0,
		},
		"ConfirmPromptConfig optional": {
			cfg: ConfirmPromptConfig{
				Required: false,
			},
			required: false,
			flags:    0,
		},
		"InputPromptConfig required": {
			cfg: InputPromptConfig{
				Required:    true,
				Placeholder: "hint",
			},
			required: true,
			flags:    0,
		},
		"InputPromptConfig optional": {
			cfg: InputPromptConfig{
				Required: false,
			},
			required: false,
			flags:    0,
		},
		"MultiSelectPromptConfig required": {
			cfg: MultiSelectPromptConfig{
				Required: true,
			},
			required: true,
			flags:    0,
		},
		"MultiSelectPromptConfig optional": {
			cfg: MultiSelectPromptConfig{
				Required: false,
			},
			required: false,
			flags:    0,
		},
		"PasswordPromptConfig with flag": {
			cfg: PasswordPromptConfig{
				Flag:     mockFlag,
				Required: true,
			},
			required: true,
			flags:    1,
		},
		"PasswordPromptConfig without flag": {
			cfg: PasswordPromptConfig{
				Required: false,
			},
			required: false,
			flags:    0,
		},
		"SelectPromptConfig with single flag": {
			cfg: SelectPromptConfig{
				Flag:     mockFlag,
				Required: true,
			},
			required: true,
			flags:    1,
		},
		"SelectPromptConfig with multiple flags": {
			cfg: SelectPromptConfig{
				Flags:    []*pflag.Flag{mockFlag, mockFlag2},
				Required: true,
			},
			required: true,
			flags:    2,
		},
		"SelectPromptConfig without flags": {
			cfg: SelectPromptConfig{
				Required: false,
			},
			required: false,
			flags:    0,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.required, tc.cfg.IsRequired())
			assert.Len(t, tc.cfg.GetFlags(), tc.flags)
		})
	}
}

func TestRetrieveFlagValue(t *testing.T) {
	tests := map[string]struct {
		flagset       []*pflag.Flag
		expectedFlag  bool
		expectedError string
	}{
		"nil flagset returns nil": {
			flagset:      nil,
			expectedFlag: false,
		},
		"nil flag in set is skipped": {
			flagset:      []*pflag.Flag{nil},
			expectedFlag: false,
		},
		"unchanged flag returns nil": {
			flagset: []*pflag.Flag{
				{Name: "test", Changed: false},
			},
			expectedFlag: false,
		},
		"single changed flag is returned": {
			flagset: []*pflag.Flag{
				{Name: "test", Changed: true},
			},
			expectedFlag: true,
		},
		"multiple changed flags returns error": {
			flagset: []*pflag.Flag{
				{Name: "a", Changed: true},
				{Name: "b", Changed: true},
			},
			expectedError: slackerror.ErrMismatchedFlags,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			io := &IOStreams{}
			flag, err := io.retrieveFlagValue(tc.flagset)
			if err != nil {
				assert.Nil(t, flag)
				assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFlag, flag != nil)
			}
		})
	}
}

func TestErrInteractivityFlags(t *testing.T) {
	tests := map[string]struct {
		cfg      PromptConfig
		contains []string
	}{
		"no flags shows generic message": {
			cfg:      ConfirmPromptConfig{},
			contains: []string{"not a TTY"},
		},
		"one flag suggests the flag name": {
			cfg: PasswordPromptConfig{
				Flag: &pflag.Flag{Name: "token"},
			},
			contains: []string{"--token"},
		},
		"multiple flags lists all flag names": {
			cfg: SelectPromptConfig{Flags: []*pflag.Flag{
				{Name: "app"},
				{Name: "team"},
			}},
			contains: []string{"--app", "--team"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := errInteractivityFlags(tc.cfg)
			for _, s := range tc.contains {
				assert.Contains(t, err.Error(), s)
			}
		})
	}
}

func TestPasswordPrompt(t *testing.T) {
	tests := map[string]struct {
		flagChanged   bool
		flagValue     string
		required      bool
		expectedError string
		expectedValue string
		expectedFlag  bool
	}{
		"error if no flag is set": {
			expectedError: slackerror.ErrPrompt,
		},
		"error if the flag is missing a required value": {
			flagChanged:   true,
			flagValue:     "",
			required:      true,
			expectedError: slackerror.ErrMissingFlag,
		},
		"allow an empty flag value if not required": {
			flagChanged:   true,
			flagValue:     "",
			required:      false,
			expectedValue: "",
			expectedFlag:  true,
		},
		"use the provided flag value": {
			flagChanged:   true,
			flagValue:     "something secret",
			expectedValue: "something secret",
			expectedFlag:  true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			// TODO: add interactive tests with FileInfoCharDevice for TTY stdout
			osMock.On("Stdout").Return(&slackdeps.FileMock{FileInfo: &slackdeps.FileInfoNamedPipe{}})
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)

			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			fs.String("token", "", "")
			if tc.flagChanged {
				require.NoError(t, fs.Set("token", tc.flagValue))
			}
			flag := fs.Lookup("token")

			selection, err := io.PasswordPrompt(ctx, "Enter a password", PasswordPromptConfig{
				Flag:     flag,
				Required: tc.required,
			})

			if err != nil {
				assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, selection.Value)
				assert.Equal(t, tc.expectedFlag, selection.Flag)
			}
		})
	}
}

func TestConfirmPrompt(t *testing.T) {
	tests := map[string]struct {
		expectedError string
	}{
		"error if non-TTY": {
			expectedError: slackerror.ErrPrompt,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.On("Stdout").Return(&slackdeps.FileMock{FileInfo: &slackdeps.FileInfoNamedPipe{}})
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)

			_, err := io.ConfirmPrompt(ctx, "Continue?", false)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
		})
	}
}

func TestInputPrompt(t *testing.T) {
	tests := map[string]struct {
		required      bool
		expectedError string
		expectedValue string
	}{
		"error if non-TTY and required": {
			required:      true,
			expectedError: slackerror.ErrPrompt,
		},
		"no error if non-TTY and optional": {
			required:      false,
			expectedValue: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.On("Stdout").Return(&slackdeps.FileMock{FileInfo: &slackdeps.FileInfoNamedPipe{}})
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)

			value, err := io.InputPrompt(ctx, "Enter name", InputPromptConfig{
				Required: tc.required,
			})

			if err != nil {
				assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedValue, value)
			}
		})
	}
}

func TestMultiSelectPrompt(t *testing.T) {
	tests := map[string]struct {
		expectedError string
	}{
		"error if non-TTY": {
			expectedError: slackerror.ErrPrompt,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.On("Stdout").Return(&slackdeps.FileMock{FileInfo: &slackdeps.FileInfoNamedPipe{}})
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)

			_, err := io.MultiSelectPrompt(ctx, "Pick items", []string{"a", "b"})

			assert.Error(t, err)
			assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
		})
	}
}

func TestSelectPrompt(t *testing.T) {
	tests := map[string]struct {
		flagValue     string
		flagChanged   bool
		options       []string
		required      bool
		expectedError string
		expectedResp  SelectPromptResponse
	}{
		"use provided flag value": {
			flagValue:   "A123",
			flagChanged: true,
			options:     []string{"A123", "A456"},
			required:    true,
			expectedResp: SelectPromptResponse{
				Flag:   true,
				Option: "A123",
			},
		},
		"error if required flag is empty": {
			flagValue:     "",
			flagChanged:   true,
			options:       []string{"A123"},
			required:      true,
			expectedError: slackerror.ErrMissingFlag,
		},
		"error if options are empty": {
			options:       []string{},
			required:      true,
			expectedError: slackerror.ErrMissingOptions,
		},
		"error if non-TTY and required": {
			options:       []string{"a", "b"},
			required:      true,
			expectedError: slackerror.ErrPrompt,
		},
		"no error if non-TTY and optional": {
			options:      []string{"a", "b"},
			required:     false,
			expectedResp: SelectPromptResponse{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())

			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			osMock.On("Stdout").Return(&slackdeps.FileMock{FileInfo: &slackdeps.FileInfoNamedPipe{}})
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)

			var flag *pflag.Flag
			if tc.flagChanged {
				fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
				fs.String("app", "", "")
				require.NoError(t, fs.Set("app", tc.flagValue))
				flag = fs.Lookup("app")
			}

			selection, err := io.SelectPrompt(ctx, "Choose", tc.options, SelectPromptConfig{
				Flag:     flag,
				Required: tc.required,
			})

			if err != nil {
				assert.Equal(t, tc.expectedError, slackerror.ToSlackError(err).Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, selection)
			}
		})
	}
}
