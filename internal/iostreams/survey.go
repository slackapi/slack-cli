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

// Survey-based prompt implementations using the survey library.
// Templates, color helpers, and survey-specific configuration live here.
// Shared prompt types and IOStreams methods live in prompts.go.

import (
	"context"
	"fmt"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// blue returns a color code for blue text that can be used in survey templates
//
// The code is compatible with the ANSI range supported by the OS and terminal
func blue() string {
	if runtime.GOOS == "windows" {
		return "cyan"
	}
	return "39"
}

// gray returns a color code for gray text that can be used in survey templates
//
// The code is compatible with the ANSI range supported by the OS and terminal
func gray() string {
	if runtime.GOOS == "windows" {
		return "default"
	}
	return "246"
}

// MimicInputPrompt formats a message and value to appear as a prompted input
func MimicInputPrompt(message string, value string) string {
	return fmt.Sprintf(
		"%s %s %s",
		style.Darken("?"),
		style.Highlight(message),
		style.Input(value),
	)
}

// SurveyOptions returns the current options applied to survey prompts
func SurveyOptions(cfg PromptConfig) []survey.AskOpt {
	var filter survey.AskOpt
	var icons survey.AskOpt
	var validator survey.AskOpt

	// surveyFilterOff removes the filtering feature when typing on a prompt to
	// guard against confusion from accidental keystrokes. The template must be
	// changed to remove the text hint in the prompt.
	surveyFilterOff := survey.WithFilter(func(filterValue string, optValue string, optIndex int) bool {
		return true
	})

	// filter contains the filter applied to select-like prompts.
	// off by default but should be extended to use a cfg filter.
	filter = surveyFilterOff
	icons = style.SurveyIcons()
	if cfg.IsRequired() {
		validator = survey.WithValidator(survey.Required)
	}

	return []survey.AskOpt{
		filter,
		icons,
		validator,
		survey.WithRemoveSelectAll(),
		survey.WithRemoveSelectNone(),
	}
}

// ConfirmQuestionTemplate is the formatted template for the confirm prompt
//
// Reference: https://github.com/go-survey/survey/blob/fa37277e6394c29db7bcc94062cb30cd7785a126/confirm.go#L25
var ConfirmQuestionTemplate = fmt.Sprintf(`
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }} {{color "reset"}}
{{- if .Answer}}
  {{- color "%s"}}{{.Answer}}{{color "reset"}}{{"\n"}}
{{- else }}
  {{- if and .Help (not .ShowHelp)}}{{color "cyan"}}[{{ .Config.HelpInput }} for help]{{color "reset"}} {{end}}
  {{- color "%s"}}{{if .Default}}(Y/n) {{else}}(y/N) {{end}}{{color "reset"}}
{{- end}}`, blue(), gray())

// surveyConfirmPrompt prompts for a yes/no confirmation using a survey form
func surveyConfirmPrompt(io *IOStreams, _ context.Context, message string, defaultValue bool) (bool, error) {
	// Temporarily swap default template for custom one
	defaultConfirmTemplate := survey.ConfirmQuestionTemplate
	survey.ConfirmQuestionTemplate = ConfirmQuestionTemplate
	defer func() {
		survey.ConfirmQuestionTemplate = defaultConfirmTemplate
	}()

	// TODO: move this config to the function parameter!
	// NOTE: currently here as a placeholder for survey options
	cfg := ConfirmPromptConfig{Required: true}

	var choice bool
	err := survey.AskOne(&survey.Confirm{
		Message: message,
		Default: defaultValue,
	}, &choice, SurveyOptions(cfg)...)

	if err != nil {
		if err == terminal.InterruptErr {
			io.SetExitCode(ExitCancel)
			return false, slackerror.New(slackerror.ErrProcessInterrupted)
		}
		return false, err
	}
	return choice, nil
}

// InputQuestionTemplate is a formatted template for the a text based prompt
//
// Reference: https://github.com/go-survey/survey/blob/fa37277e6394c29db7bcc94062cb30cd7785a126/input.go#L43
var InputQuestionTemplate = fmt.Sprintf(`
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }} {{color "reset"}}
{{- if .ShowAnswer}}
  {{- color "%s"}}{{.Answer}}{{color "reset"}}{{"\n"}}
{{- else if .PageEntries -}}
  {{- .Answer}} [Use arrows to move, enter to select, type to continue]
  {{- "\n"}}
  {{- range $ix, $choice := .PageEntries}}
    {{- if eq $ix $.SelectedIndex }}{{color $.Config.Icons.SelectFocus.Format }}{{ $.Config.Icons.SelectFocus.Text }} {{else}}{{color "default"}}  {{end}}
    {{- $choice.Value}}
    {{- color "reset"}}{{"\n"}}
  {{- end}}
{{- else }}
  {{- if or (and .Help (not .ShowHelp)) .Suggest }}{{color "cyan"}}[
    {{- if and .Help (not .ShowHelp)}}{{ print .Config.HelpInput }} for help {{- if and .Suggest}}, {{end}}{{end -}}
    {{- if and .Suggest }}{{color "cyan"}}{{ print .Config.SuggestInput }} for suggestions{{end -}}
  ]{{color "reset"}} {{end}}
  {{- if .Default}}{{color "%s"}}({{.Default}}) {{color "reset"}}{{end}}
{{- end}}`, blue(), gray())

// surveyInputPrompt prompts for text input using a survey form
func surveyInputPrompt(io *IOStreams, _ context.Context, message string, cfg InputPromptConfig) (string, error) {
	defaultInputTemplate := survey.InputQuestionTemplate
	survey.InputQuestionTemplate = InputQuestionTemplate
	defer func() {
		survey.InputQuestionTemplate = defaultInputTemplate
	}()

	var input string
	err := survey.AskOne(&survey.Input{
		Message: message,
		Default: cfg.Placeholder,
	}, &input, SurveyOptions(cfg)...)

	if err != nil {
		if err == terminal.InterruptErr {
			io.SetExitCode(ExitCancel)
			return "", slackerror.New(slackerror.ErrProcessInterrupted)
		}
		return "", err
	}
	return input, nil
}

// MultiSelectQuestionTemplate represents a formatted template with all hints
// and filters removed
//
// Reference: https://github.com/go-survey/survey/blob/fa37277e6394c29db7bcc94062cb30cd7785a126/multiselect.go#L71
var MultiSelectQuestionTemplate = fmt.Sprintf(`
{{- define "option"}}
    {{- if eq .SelectedIndex .CurrentIndex }}{{color .Config.Icons.SelectFocus.Format }}{{ .Config.Icons.SelectFocus.Text }}{{color "reset"}}{{else}} {{end}}
    {{- if index .Checked .CurrentOpt.Index }}{{color .Config.Icons.MarkedOption.Format }} {{ .Config.Icons.MarkedOption.Text }} {{else}}{{color .Config.Icons.UnmarkedOption.Format }} {{ .Config.Icons.UnmarkedOption.Text }} {{end}}
    {{- color "reset"}}
    {{- " "}}{{- .CurrentOpt.Value}}{{ if ne ($.GetDescription .CurrentOpt) "" }} - {{color "cyan"}}{{ $.GetDescription .CurrentOpt }}{{color "reset"}}{{end}}
{{end}}
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }}{{color "reset"}}
{{- if .ShowAnswer}}{{color "%s"}} {{.Answer}}{{color "reset"}}{{"\n"}}
{{- else }}
	{{- "  "}}{{- color "%s"}}[Space to select{{- if not .Config.RemoveSelectAll }}, <right> to all{{end}}{{- if not .Config.RemoveSelectNone }}, <left> to none{{end}}{{- if and .Help (not .ShowHelp)}}, {{ .Config.HelpInput }} for more help{{end}}]{{color "reset"}}
  {{- "\n"}}
  {{- range $ix, $option := .PageEntries}}
    {{- template "option" $.IterateOption $ix $option}}
  {{- end}}
{{- end}}`, blue(), blue())

// surveyMultiSelectPrompt prompts for multiple selections using a survey form
func surveyMultiSelectPrompt(io *IOStreams, _ context.Context, message string, options []string) ([]string, error) {
	defaultMultiSelectTemplate := survey.MultiSelectQuestionTemplate
	survey.MultiSelectQuestionTemplate = MultiSelectQuestionTemplate
	defer func() {
		survey.MultiSelectQuestionTemplate = defaultMultiSelectTemplate
	}()

	// TODO: move this config to the function parameter!
	// NOTE: currently here as a placeholder for survey options
	cfg := MultiSelectPromptConfig{Required: true}

	// Collect the selected values
	var values []string
	err := survey.AskOne(&survey.MultiSelect{
		Message: message,
		Options: options,
	}, &values, SurveyOptions(cfg)...)

	if err != nil {
		if err == terminal.InterruptErr {
			io.SetExitCode(ExitCancel)
			return []string{}, slackerror.New(slackerror.ErrProcessInterrupted)
		}
		return []string{}, err
	}
	return values, nil
}

// passwordQuestionTemplate is a template with custom formatting
//
// Reference: https://github.com/go-survey/survey/blob/fa37277e6394c29db7bcc94062cb30cd7785a126/password.go#L32
var passwordQuestionTemplate = fmt.Sprintf(`
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }} {{color "%s"}}
{{- if and .Help (not .ShowHelp)}}{{color "cyan"}}[{{ .Config.HelpInput }} for help]{{color "reset"}} {{end}}`, gray())

// surveyPasswordPrompt prompts for a password (hidden input) using a survey form
func surveyPasswordPrompt(io *IOStreams, _ context.Context, message string, cfg PasswordPromptConfig) (PasswordPromptResponse, error) {
	defaultPasswordTemplate := survey.PasswordQuestionTemplate
	if cfg.Template != "" {
		survey.PasswordQuestionTemplate = cfg.Template
	} else {
		survey.PasswordQuestionTemplate = passwordQuestionTemplate
	}
	defer func() {
		survey.PasswordQuestionTemplate = defaultPasswordTemplate
		if !io.config.NoColor {
			_, _ = io.WriteOut().Write([]byte("\x1b[0m")) // Reset prompt format
		}
	}()

	var input string
	err := survey.AskOne(&survey.Password{
		Message: message,
	}, &input, SurveyOptions(cfg)...)

	if err != nil {
		if err == terminal.InterruptErr {
			io.SetExitCode(ExitCancel)
			return PasswordPromptResponse{}, slackerror.New(slackerror.ErrProcessInterrupted)
		}
		return PasswordPromptResponse{}, err
	}
	return PasswordPromptResponse{Prompt: true, Value: input}, nil
}

// selectQuestionTemplate represents a formatted template with all hints and
// filters removed
//
// Reference: https://github.com/go-survey/survey/blob/fa37277e6394c29db7bcc94062cb30cd7785a126/select.go#L69
var selectQuestionTemplate = fmt.Sprintf(`
{{- define "option"}}
	{{- if eq .SelectedIndex .CurrentIndex }}{{color .Config.Icons.SelectFocus.Format }}{{ .Config.Icons.SelectFocus.Text }}{{color "default+b"}} {{else}}{{color "default"}}  {{end}}
	{{- .CurrentOpt.Value}}{{color "reset"}}{{ if ne ($.GetDescription .CurrentOpt) "" }}{{"\n  "}}{{color "250"}}{{ $.GetDescription .CurrentOpt }}{{"\n"}}{{end}}
	{{- color "reset"}}
{{end}}
{{- if .ShowHelp }}{{- color .Config.Icons.Help.Format }}{{ .Config.Icons.Help.Text }} {{ .Help }}{{color "reset"}}{{"\n"}}{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }}{{color "reset"}}
{{- if .ShowAnswer}}{{color "%s"}} {{.Answer}}{{color "reset"}}{{"\n"}}
{{- else}}
	{{- "\n"}}
	{{- range $ix, $option := .PageEntries}}
		{{- template "option" $.IterateOption $ix $option}}
	{{- end}}
{{- end}}`, blue())

// surveySelectPrompt prompts for a single selection using a survey form
func surveySelectPrompt(io *IOStreams, _ context.Context, msg string, options []string, cfg SelectPromptConfig) (SelectPromptResponse, error) {
	defaultSelectTemplate := survey.SelectQuestionTemplate
	if cfg.Template != "" {
		survey.SelectQuestionTemplate = cfg.Template
	} else {
		survey.SelectQuestionTemplate = selectQuestionTemplate
	}
	defer func() {
		survey.SelectQuestionTemplate = defaultSelectTemplate
	}()

	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}

	if cfg.Description != nil {
		prompt.Description = cfg.Description
	}
	if cfg.PageSize != 0 {
		prompt.PageSize = cfg.PageSize
	}

	// Collect the selected index and return with the selected value
	var index int
	if err := survey.AskOne(prompt, &index, SurveyOptions(cfg)...); err != nil {
		if err == terminal.InterruptErr {
			io.SetExitCode(ExitCancel)
			return SelectPromptResponse{}, slackerror.New(slackerror.ErrProcessInterrupted)
		}
		return SelectPromptResponse{}, err
	}
	return SelectPromptResponse{Prompt: true, Index: index, Option: options[index]}, nil
}
