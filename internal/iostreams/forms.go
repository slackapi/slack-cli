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

// Interactive form-based prompt implementations with Charm's Huh package.
//
// Reference: https://github.com/charmbracelet/huh?tab=readme-ov-file#huh

import (
	"context"
	"errors"
	"slices"

	huh "charm.land/huh/v2"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// newForm wraps a field in an interactive form with optional Slack theming.
func newForm(io *IOStreams, field huh.Field) *huh.Form {
	form := huh.NewForm(huh.NewGroup(field))
	if io != nil && io.config.NoColor {
		form = form.WithTheme(style.ThemePlain())
	} else if io != nil && io.config.WithExperimentOn(experiment.Lipgloss) {
		form = form.WithTheme(style.ThemeSlack())
	} else {
		form = form.WithTheme(style.ThemeSurvey())
	}
	if io != nil && io.config.Accessible {
		form = form.WithAccessible(true)
	}
	return form
}

// buildInputForm constructs an interactive form for text input prompts.
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

// inputForm interactively prompts for text input.
func inputForm(io *IOStreams, _ context.Context, message string, cfg InputPromptConfig) (string, error) {
	var input string
	err := buildInputForm(io, message, cfg, &input).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return "", slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return "", err
	}
	return input, nil
}

// buildConfirmForm constructs an interactive form for yes/no confirmation prompts.
func buildConfirmForm(io *IOStreams, message string, choice *bool) *huh.Form {
	field := huh.NewConfirm().
		Title(message).
		Value(choice)
	return newForm(io, field)
}

// confirmForm interactively prompts for a yes/no confirmation.
func confirmForm(io *IOStreams, _ context.Context, message string, defaultValue bool) (bool, error) {
	var choice = defaultValue
	err := buildConfirmForm(io, message, &choice).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return false, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return false, err
	}
	return choice, nil
}

// buildSelectForm constructs an interactive form for single-selection prompts.
func buildSelectForm(io *IOStreams, msg string, options []string, cfg SelectPromptConfig, selected *string) *huh.Form {
	var opts []huh.Option[string]
	for _, opt := range options {
		key := opt
		if cfg.Description != nil {
			if desc := style.RemoveEmoji(cfg.Description(opt, len(opts))); desc != "" {
				key = style.Bright(opt) + style.Separator() + style.Secondary(desc)
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

// selectForm interactively prompts the user to select one option.
func selectForm(io *IOStreams, _ context.Context, msg string, options []string, cfg SelectPromptConfig) (SelectPromptResponse, error) {
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

// buildPasswordForm constructs an interactive form for password (hidden input) prompts.
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

// passwordForm interactively prompts for a password with hidden input.
func passwordForm(io *IOStreams, _ context.Context, message string, cfg PasswordPromptConfig) (PasswordPromptResponse, error) {
	var input string
	err := buildPasswordForm(io, message, cfg, &input).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return PasswordPromptResponse{}, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return PasswordPromptResponse{}, err
	}
	return PasswordPromptResponse{Prompt: true, Value: input}, nil
}

// buildMultiSelectForm constructs an interactive form for multiple-selection prompts.
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

// multiSelectForm interactively prompts the user to select multiple options.
func multiSelectForm(io *IOStreams, _ context.Context, message string, options []string) ([]string, error) {
	var selected []string
	err := buildMultiSelectForm(io, message, options, &selected).Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return []string{}, slackerror.New(slackerror.ErrProcessInterrupted)
	} else if err != nil {
		return []string{}, err
	}
	return selected, nil
}
