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

// Prompt type definitions shared between survey and charm implementations.

import (
	"context"
	"fmt"
	"strings"

	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/pflag"
)

// PromptConfig contains general information about a prompt
type PromptConfig interface {
	GetFlags() []*pflag.Flag // GetFlags returns all flags for the prompt
	IsRequired() bool        // IsRequired returns if a response must be provided
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
	Required bool // If a response is required
}

// GetFlags returns all flags for the MultiSelect prompt
func (cfg MultiSelectPromptConfig) GetFlags() []*pflag.Flag {
	return []*pflag.Flag{}
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
	Description func(value string, index int) string // Optional text displayed below each prompt option
	Flag        *pflag.Flag                          // The single flag substitute for this prompt
	Flags       []*pflag.Flag                        // Otherwise multiple flag substitutes for this prompt
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

// errInteractivityFlags formats an error for when flag substitutes are needed
func errInteractivityFlags(cfg PromptConfig) error {
	flags := cfg.GetFlags()
	var remediation string
	var helpMessage = "Learn more about this command with `--help`"

	if len(flags) == 1 {
		remediation = fmt.Sprintf("Try running the command with the `--%s` flag included", flags[0].Name)
		helpMessage = "Learn more about this flag with `--help`"
	} else if len(flags) > 1 {
		var names []string
		for _, flag := range flags {
			names = append(names, flag.Name)
		}
		flags := strings.Join(names, "`\n   `--")
		remediation = fmt.Sprintf("Consider using the following flags when running this command:\n   `--%s`", flags)
		helpMessage = "Learn more about these flags with `--help`"
	}

	return slackerror.New(slackerror.ErrPrompt).
		WithDetails(slackerror.ErrorDetails{
			slackerror.ErrorDetail{Message: "The input device is not a TTY or does not support interactivity"},
		}).
		WithRemediation("%s\n%s", remediation, helpMessage)
}

// ConfirmPrompt prompts the user for a "yes" or "no" (true or false) value for
// the message
func (io *IOStreams) ConfirmPrompt(ctx context.Context, message string, defaultValue bool) (bool, error) {
	if io.config.WithExperimentOn(experiment.Charm) {
		return charmConfirmPrompt(io, ctx, message, defaultValue)
	}
	return surveyConfirmPrompt(io, ctx, message, defaultValue)
}

// InputPrompt prompts the user for a string value for the message, which can
// optionally be made required
func (io *IOStreams) InputPrompt(ctx context.Context, message string, cfg InputPromptConfig) (string, error) {
	if io.config.WithExperimentOn(experiment.Charm) {
		return charmInputPrompt(io, ctx, message, cfg)
	}
	return surveyInputPrompt(io, ctx, message, cfg)
}

// MultiSelectPrompt prompts the user to select multiple values in a list and
// returns the selected values
func (io *IOStreams) MultiSelectPrompt(ctx context.Context, message string, options []string) ([]string, error) {
	if io.config.WithExperimentOn(experiment.Charm) {
		return charmMultiSelectPrompt(io, ctx, message, options)
	}
	return surveyMultiSelectPrompt(io, ctx, message, options)
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
		return PasswordPromptResponse{}, errInteractivityFlags(cfg)
	}

	if io.config.WithExperimentOn(experiment.Charm) {
		return charmPasswordPrompt(io, ctx, message, cfg)
	}
	return surveyPasswordPrompt(io, ctx, message, cfg)
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
			return SelectPromptResponse{}, errInteractivityFlags(cfg)
		} else {
			return SelectPromptResponse{}, nil
		}
	}

	if io.config.WithExperimentOn(experiment.Charm) {
		return charmSelectPrompt(io, ctx, msg, options, cfg)
	}
	return surveySelectPrompt(io, ctx, msg, options, cfg)
}
