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

// Charm-based prompt implementations using the huh library.
// These are used when the "huh" experiment is enabled.

import (
	"context"
	"errors"
	"slices"

	huh "charm.land/huh/v2"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// newForm wraps a field in a huh form, applying the Slack theme when the lipgloss experiment is enabled.
func newForm(io *IOStreams, field huh.Field) *huh.Form {
	form := huh.NewForm(huh.NewGroup(field))
	if io != nil && io.config.WithExperimentOn(experiment.Lipgloss) {
		form = form.WithTheme(style.ThemeSlack())
	}
	return form
}

// buildInputForm constructs a huh form for text input prompts.
func buildInputForm(io *IOStreams, message string, cfg InputPromptConfig, input *string) *huh.Form {
	field := huh.NewInput().
		Title(message).
		Prompt(style.Chevron() + " ").
		Placeholder(cfg.Placeholder).
		Value(input)
	if cfg.Required {
		field.Validate(huh.ValidateMinLength(1))
	}
	return newForm(io, field)
}

// charmInputPrompt prompts for text input using a charm huh form
func charmInputPrompt(io *IOStreams, _ context.Context, message string, cfg InputPromptConfig) (string, error) {
	var input string
	err := buildInputForm(io, message, cfg, &input).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return "", err
	}
	return input, nil
}

// buildConfirmForm constructs a huh form for yes/no confirmation prompts.
func buildConfirmForm(io *IOStreams, message string, choice *bool) *huh.Form {
	field := huh.NewConfirm().
		Title(message).
		Value(choice)
	return newForm(io, field)
}

// charmConfirmPrompt prompts for a yes/no confirmation using a charm huh form
func charmConfirmPrompt(io *IOStreams, _ context.Context, message string, defaultValue bool) (bool, error) {
	var choice = defaultValue
	err := buildConfirmForm(io, message, &choice).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return false, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return false, err
	}
	return choice, nil
}

// buildSelectForm constructs a huh form for single-selection prompts.
func buildSelectForm(io *IOStreams, msg string, options []string, cfg SelectPromptConfig, selected *string) *huh.Form {
	var opts []huh.Option[string]
	for _, opt := range options {
		key := opt
		if cfg.Description != nil {
			if desc := style.RemoveEmoji(cfg.Description(opt, len(opts))); desc != "" {
				key = opt + " - " + desc
			}
		}
		opts = append(opts, huh.NewOption(key, opt))
	}

	field := huh.NewSelect[string]().
		Title(msg).
		Description(cfg.Help).
		Options(opts...).
		Value(selected)

	return newForm(io, field)
}

// charmSelectPrompt prompts the user to select one option using a charm huh form
func charmSelectPrompt(io *IOStreams, _ context.Context, msg string, options []string, cfg SelectPromptConfig) (SelectPromptResponse, error) {
	var selected string
	err := buildSelectForm(io, msg, options, cfg, &selected).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return SelectPromptResponse{}, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return SelectPromptResponse{}, err
	}

	index := slices.Index(options, selected)
	return SelectPromptResponse{Prompt: true, Index: index, Option: selected}, nil
}

// buildPasswordForm constructs a huh form for password (hidden input) prompts.
func buildPasswordForm(io *IOStreams, message string, cfg PasswordPromptConfig, input *string) *huh.Form {
	field := huh.NewInput().
		Title(message).
		Prompt(style.Chevron() + " ").
		EchoMode(huh.EchoModePassword).
		Value(input)
	if cfg.Required {
		field.Validate(huh.ValidateMinLength(1))
	}
	return newForm(io, field)
}

// charmPasswordPrompt prompts for a password (hidden input) using a charm huh form
func charmPasswordPrompt(io *IOStreams, _ context.Context, message string, cfg PasswordPromptConfig) (PasswordPromptResponse, error) {
	var input string
	err := buildPasswordForm(io, message, cfg, &input).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return PasswordPromptResponse{}, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return PasswordPromptResponse{}, err
	}
	return PasswordPromptResponse{Prompt: true, Value: input}, nil
}

// buildMultiSelectForm constructs a huh form for multiple-selection prompts.
func buildMultiSelectForm(io *IOStreams, message string, options []string, selected *[]string) *huh.Form {
	var opts []huh.Option[string]
	for _, opt := range options {
		opts = append(opts, huh.NewOption(opt, opt))
	}

	field := huh.NewMultiSelect[string]().
		Title(message).
		Options(opts...).
		Value(selected)

	return newForm(io, field)
}

// charmMultiSelectPrompt prompts the user to select multiple options using a charm huh form
func charmMultiSelectPrompt(io *IOStreams, _ context.Context, message string, options []string) ([]string, error) {
	var selected []string
	err := buildMultiSelectForm(io, message, options, &selected).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return []string{}, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return []string{}, err
	}
	return selected, nil
}
