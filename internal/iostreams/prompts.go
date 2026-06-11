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

// Prompts handle flag values and interactive forms for gathering user input.

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/pflag"
)

// PromptConfig contains general information about a prompt
type PromptConfig interface {
	GetFlags() []*pflag.Flag // GetFlags returns all flags for the prompt
	IsRequired() bool        // IsRequired returns if a response must be provided
}

// PromptOption pairs an interactive option label with the flag invocation
// that picks the same option non-interactively. When a prompt is reached in
// a non-TTY context, the resulting error renders one of these per option so
// agents and scripts can re-run with the right --flag=value.
type PromptOption struct {
	Label string // The option as rendered in the interactive list
	Flag  string // The pflag name, e.g. "team", "app"
	Value string // The value to pass, e.g. "T0123" or "A0ABCD"
}

// PromptOptionsConfig is optionally implemented by prompt configs that can
// enumerate options as flag invocations. Configs that do not implement it
// (or return an empty slice) keep the simpler "Try --flag" remediation.
type PromptOptionsConfig interface {
	GetPromptOptions() []PromptOption
}

// ConfirmPromptConfig holds additional configs for a Confirm prompt
type ConfirmPromptConfig struct {
	Required bool // If a response is required
}

// GetFlags returns all flags for the Confirm prompt
func (cfg ConfirmPromptConfig) GetFlags() []*pflag.Flag {
	return []*pflag.Flag{}
}

// IsRequired returns if a response is required
func (cfg ConfirmPromptConfig) IsRequired() bool {
	return cfg.Required
}

// InputPromptConfig holds additional config for an Input prompt
type InputPromptConfig struct {
	Required    bool   // Whether the input must be non-empty
	Placeholder string // Placeholder text shown when input is empty
}

// GetFlags returns all flags for the Input prompt
func (cfg InputPromptConfig) GetFlags() []*pflag.Flag {
	return []*pflag.Flag{}
}

// IsRequired returns if a response is required
func (cfg InputPromptConfig) IsRequired() bool {
	return cfg.Required
}

// MultiSelectPromptConfig holds additional configs for a MultiSelect prompt
type MultiSelectPromptConfig struct {
	Options  []PromptOption // Optional flag invocations parallel to the prompt's options
	Required bool           // If a response is required
}

// GetFlags returns all flags for the MultiSelect prompt
func (cfg MultiSelectPromptConfig) GetFlags() []*pflag.Flag {
	return []*pflag.Flag{}
}

// GetPromptOptions returns flag invocations for each option, when set
func (cfg MultiSelectPromptConfig) GetPromptOptions() []PromptOption {
	return cfg.Options
}

// IsRequired returns if a response is required
func (cfg MultiSelectPromptConfig) IsRequired() bool {
	return cfg.Required
}

// PasswordPromptConfig holds additional config for a password prompt
type PasswordPromptConfig struct {
	Flag     *pflag.Flag // The flag substitute for this prompt
	Required bool        // If a response is required
	Template string      // DEPRECATED: Custom formatting of the password prompt
}

// GetFlags returns all flags for the password prompt
func (cfg PasswordPromptConfig) GetFlags() []*pflag.Flag {
	switch {
	case cfg.Flag != nil:
		return []*pflag.Flag{cfg.Flag}
	default:
		return []*pflag.Flag{}
	}
}

// IsRequired returns if a response is required
func (cfg PasswordPromptConfig) IsRequired() bool {
	return cfg.Required
}

// PasswordPromptResponse holds response information from a password prompt
type PasswordPromptResponse struct {
	Value  string // The value of the response
	Flag   bool   // If a flag value was used
	Prompt bool   // If a survey input was provided
}

// SelectPromptConfig holds additional config for a selection prompt
type SelectPromptConfig struct {
	Description func(value string, index int) string // Optional text displayed with each prompt option
	Flag        *pflag.Flag                          // The single flag substitute for this prompt
	Flags       []*pflag.Flag                        // Otherwise multiple flag substitutes for this prompt
	Help        string                               // Optional help text displayed below the select title
	Options     []PromptOption                       // Optional flag invocations parallel to the prompt's options
	PageSize    int                                  // DEPRECATED: The number of options displayed before the user needs to scroll
	Required    bool                                 // If a response is required
	Template    string                               // DEPRECATED: Custom formatting of the selection prompt
}

// GetFlags returns all flags for the prompt
func (cfg SelectPromptConfig) GetFlags() []*pflag.Flag {
	switch {
	case cfg.Flag != nil:
		return []*pflag.Flag{cfg.Flag}
	case cfg.Flags != nil:
		return cfg.Flags
	default:
		return []*pflag.Flag{}
	}
}

// GetPromptOptions returns flag invocations for each option, when set
func (cfg SelectPromptConfig) GetPromptOptions() []PromptOption {
	return cfg.Options
}

// IsRequired returns if a response is required
func (cfg SelectPromptConfig) IsRequired() bool {
	return cfg.Required
}

// SelectPromptResponse holds response information from a selection prompt
type SelectPromptResponse struct {
	Index  int    // The index of any selection
	Option string // The value of the response

	Flag   bool // If a flag value was used
	Prompt bool // If a survey selection was made
}

// retrieveFlagValue returns the only changed flag in the flagset
func (io *IOStreams) retrieveFlagValue(flagset []*pflag.Flag) (*pflag.Flag, error) {
	var flag *pflag.Flag
	if flagset == nil {
		return nil, nil
	}
	for _, opt := range flagset {
		if opt == nil {
			continue
		}
		if !opt.Changed {
			continue
		} else if flag != nil {
			return nil, slackerror.New(slackerror.ErrMismatchedFlags)
		}
		flag = opt
	}
	return flag, nil
}

// errInteractivityFlags formats an error for when flag substitutes are needed.
// It re-renders the prompt question and any enumerable options (with their
// equivalent --flag=value invocations) so agents and devops scripts can read
// the error and re-run with the right flags. Configs that don't expose
// per-option flag invocations fall back to a flag-name-only suggestion.
func errInteractivityFlags(cfg PromptConfig, message string, options []string) error {
	flags := cfg.GetFlags()

	details := slackerror.ErrorDetails{
		slackerror.ErrorDetail{Message: "The input device is not a TTY or does not support interactivity"},
	}

	var promptOptions []PromptOption
	if oc, ok := cfg.(PromptOptionsConfig); ok {
		promptOptions = oc.GetPromptOptions()
	}
	// Only honor per-option flag invocations when they line up with the
	// options actually shown to the user; mismatches indicate a stale config.
	if len(options) > 0 && len(promptOptions) != len(options) {
		promptOptions = nil
	}

	var lines []string
	if message != "" {
		lines = append(lines, fmt.Sprintf("? %s", message))
	}

	hasFlagOptions := false
	for _, opt := range promptOptions {
		if opt.Flag != "" && opt.Value != "" {
			hasFlagOptions = true
			break
		}
	}

	switch {
	case hasFlagOptions:
		for _, opt := range promptOptions {
			if opt.Flag == "" || opt.Value == "" {
				lines = append(lines, fmt.Sprintf("  %s", opt.Label))
				continue
			}
			lines = append(lines, fmt.Sprintf("  %s  %s", opt.Label, style.Secondary(fmt.Sprintf("--%s=%s", opt.Flag, opt.Value))))
		}
		lines = append(lines, "Re-run with one of the values above")
	case len(flags) == 1:
		lines = append(lines, fmt.Sprintf("Try running the command with the `--%s` flag included", flags[0].Name))
		lines = append(lines, "Learn more about this flag with `--help`")
	case len(flags) > 1:
		names := make([]string, 0, len(flags))
		for _, flag := range flags {
			names = append(names, flag.Name)
		}
		lines = append(lines, fmt.Sprintf("Consider using the following flags when running this command:\n   `--%s`", strings.Join(names, "`\n   `--")))
		lines = append(lines, "Learn more about these flags with `--help`")
	default:
		lines = append(lines, "Learn more about this command with `--help`")
	}

	return slackerror.New(slackerror.ErrPrompt).
		WithDetails(details).
		WithRemediation("%s", strings.Join(lines, "\n"))
}

// ConfirmPrompt prompts the user for a "yes" or "no" (true or false) value for
// the message
func (io *IOStreams) ConfirmPrompt(ctx context.Context, message string, defaultValue bool) (bool, error) {
	if !io.IsTTY() {
		return false, errInteractivityFlags(ConfirmPromptConfig{}, message, nil)
	}
	return confirmForm(io, ctx, message, defaultValue)
}

// InputPrompt prompts the user for a string value for the message, which can
// optionally be made required
func (io *IOStreams) InputPrompt(ctx context.Context, message string, cfg InputPromptConfig) (string, error) {
	if !io.IsTTY() {
		if cfg.IsRequired() {
			return "", errInteractivityFlags(cfg, message, nil)
		}
		return "", nil
	}
	return inputForm(io, ctx, message, cfg)
}

// MultiSelectPrompt prompts the user to select multiple values in a list and
// returns the selected values
func (io *IOStreams) MultiSelectPrompt(ctx context.Context, message string, options []string) ([]string, error) {
	if !io.IsTTY() {
		return nil, errInteractivityFlags(MultiSelectPromptConfig{}, message, options)
	}
	return multiSelectForm(io, ctx, message, options)
}

// PasswordPrompt prompts the user with a hidden text input for the message
func (io *IOStreams) PasswordPrompt(ctx context.Context, message string, cfg PasswordPromptConfig) (PasswordPromptResponse, error) {
	if cfg.Flag != nil && cfg.Flag.Changed {
		if cfg.Required && cfg.Flag.Value.String() == "" {
			return PasswordPromptResponse{}, slackerror.New(slackerror.ErrMissingFlag)
		}
		return PasswordPromptResponse{Flag: true, Value: cfg.Flag.Value.String()}, nil
	}
	if !io.IsTTY() {
		return PasswordPromptResponse{}, errInteractivityFlags(cfg, message, nil)
	}

	return passwordForm(io, ctx, message, cfg)
}

// SelectPrompt prompts the user to make a selection and returns the choice
func (io *IOStreams) SelectPrompt(ctx context.Context, msg string, options []string, cfg SelectPromptConfig) (SelectPromptResponse, error) {
	if flag, err := io.retrieveFlagValue(cfg.GetFlags()); err != nil {
		return SelectPromptResponse{}, err
	} else if flag != nil {
		if cfg.Required && cfg.Flag.Value.String() == "" {
			return SelectPromptResponse{}, slackerror.New(slackerror.ErrMissingFlag)
		}
		return SelectPromptResponse{Flag: true, Option: flag.Value.String()}, nil
	}

	if len(options) == 0 {
		return SelectPromptResponse{}, slackerror.New(slackerror.ErrMissingOptions)
	}
	if !io.IsTTY() {
		if cfg.IsRequired() {
			return SelectPromptResponse{}, errInteractivityFlags(cfg, msg, options)
		} else {
			return SelectPromptResponse{}, nil
		}
	}

	return selectForm(io, ctx, msg, options, cfg)
}
