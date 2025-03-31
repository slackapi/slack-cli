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

package style

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"unicode"
)

// TemplateData contains any optional values to include in the formatted message
//
// Note: A custom template function should also be added for custom values
type TemplateData map[string]interface{}

// getTemplateFuncs returns custom functions that format data within templates
//
// Note: Cobra functions and functions added with AddTemplateFuncs are supported
func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"Title": func(s string) string {
			return Styler().Bold(strings.ToUpper(s)).String()
		},
		"IsAlias": func(cmdName string, aliases map[string]string) bool {
			_, exists := aliases[cmdName]
			return exists
		},
		"AliasParent": func(cmdName string, aliases map[string]string) string {
			for aliasCmdName, aliasParentName := range aliases {
				if aliasCmdName == cmdName {
					return aliasParentName
				}
			}
			return ""
		},
		"Emoji": func(emoji string) string {
			return Emoji(emoji)
		},
		"Examples": func(examples string) string {
			return ExampleTemplatef(examples)
		},
		"Experiments": func(experiments []string) string {
			if len(experiments) == 0 {
				return ExampleTemplatef("None")
			}
			return ExampleTemplatef(strings.Join(experiments, "\n"))
		},
		"HasAliasSubcommands": func(parentName string, aliases map[string]string) bool {
			for _, aliasParentName := range aliases {
				if aliasParentName == parentName {
					return true
				}
			}
			return false
		},
		"Header": func(header string) string {
			return Header(header)
		},
		"LinkText": func(link string) string {
			return LinkText(link)
		},
		"AliasPadding":   func() int { return 5 },
		"ToCommandText":  CommandText,
		"ToBold":         Bold,
		"GetProcessName": processName,
		"Error": func(message string, code string) string {
			text := fmt.Sprintf("Error: %s (%s)", message, code)
			return Styler().Index(redDark, text).String()
		},
		"Suggestion": func(remediation string) string {
			text := fmt.Sprintf("Suggestion: %s", remediation)
			return Secondary(text)
		},
		"Context": func(text string) string {
			return Secondary(fmt.Sprintf("(%s)", text))
		},
		"Green": func(text string) string {
			return Selector(text)
		},
		"Red": func(text string) string {
			return Styler().Index(redDark, text).String()
		},
		"rpad": func(s string, padding int) string {
			formattedString := fmt.Sprintf("%%-%ds", padding)
			return fmt.Sprintf(formattedString, s)
		},
		"trimTrailingWhitespaces": func(s string) string {
			return strings.TrimRightFunc(s, unicode.IsSpace)
		},
	}
}

// PrintTemplate outputs the template with data to the writer or errors
//
// Note: Templates use the Go standard library template package
// https://pkg.go.dev/text/template
func PrintTemplate(w io.Writer, tmpl string, data any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Failed to create the template! %+v", r)
		}
	}()
	t := template.New("template")
	t.Funcs(getTemplateFuncs())
	template.Must(t.Parse(tmpl))
	return t.Execute(w, data)
}
