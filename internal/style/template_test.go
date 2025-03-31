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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTemplateFuncs(t *testing.T) {
	isStyleEnabled = false
	aliases := map[string]string{
		"run":    "platform",
		"deploy": "platform",
		"delete": "app",
	}
	funcs := getTemplateFuncs()

	t.Run("Title", func(t *testing.T) {
		require.NotNil(t, funcs["Title"])
		title := funcs["Title"].(func(string) string)("section")
		assert.Equal(t, "SECTION", title)
	})
	t.Run("IsAlias", func(t *testing.T) {
		require.NotNil(t, funcs["IsAlias"])
		isAlias := funcs["IsAlias"].(func(string, map[string]string) bool)("delete", aliases)
		assert.True(t, isAlias)
		isAlias = funcs["IsAlias"].(func(string, map[string]string) bool)("doctor", aliases)
		assert.False(t, isAlias)
	})
	t.Run("AliasParent", func(t *testing.T) {
		require.NotNil(t, funcs["AliasParent"])
		aliasParent := funcs["AliasParent"].(func(string, map[string]string) string)("run", aliases)
		assert.Equal(t, "platform", aliasParent)
		aliasParent = funcs["AliasParent"].(func(string, map[string]string) string)("help", aliases)
		assert.Equal(t, "", aliasParent)
	})
	t.Run("Experiments", func(t *testing.T) {
		require.NotNil(t, funcs["Experiments"])
		experiments := funcs["Experiments"].(func([]string) string)([]string{})
		assert.Contains(t, experiments, "None")
		experiments = funcs["Experiments"].(func([]string) string)([]string{"EXP_001", "EXP_002"})
		assert.Contains(t, experiments, "EXP_001", "EXP_002")
	})
	t.Run("HasAliasSubcommands", func(t *testing.T) {
		require.NotNil(t, funcs["HasAliasSubcommands"])
		hasAliasSubcommands := funcs["HasAliasSubcommands"].(func(string, map[string]string) bool)("platform", aliases)
		assert.True(t, hasAliasSubcommands)
		hasAliasSubcommands = funcs["HasAliasSubcommands"].(func(string, map[string]string) bool)("feedback", aliases)
		assert.False(t, hasAliasSubcommands)
	})
	t.Run("Header", func(t *testing.T) {
		require.NotNil(t, funcs["Header"])
		require.Equal(t, funcs["Header"].(func(string) string)("Usage"), Header("Usage"))
	})
	t.Run("LinkText", func(t *testing.T) {
		require.NotNil(t, funcs["LinkText"])
		require.Equal(t, funcs["LinkText"].(func(string) string)("https://example.com"), LinkText("https://example.com"))
	})
	t.Run("AliasPadding", func(t *testing.T) {
		require.NotNil(t, funcs["AliasPadding"])
		aliasPadding := funcs["AliasPadding"].(func() int)()
		assert.GreaterOrEqual(t, aliasPadding, 0)
	})
	t.Run("Error", func(t *testing.T) {
		require.NotNil(t, funcs["Error"])
		actualError := funcs["Error"].(func(string, string) string)("something happened", "cli_error_code")
		assert.Equal(t, "Error: something happened (cli_error_code)", actualError)
	})
	t.Run("Suggestion", func(t *testing.T) {
		require.NotNil(t, funcs["Suggestion"])
		actualSuggestion := funcs["Suggestion"].(func(string) string)("Fix this!")
		assert.Equal(t, "Suggestion: Fix this!", actualSuggestion)
	})
	t.Run("Context", func(t *testing.T) {
		require.NotNil(t, funcs["Context"])
		actualContext := funcs["Context"].(func(string) string)("and more information here")
		assert.Equal(t, "(and more information here)", actualContext)
	})
	t.Run("Green", func(t *testing.T) {
		require.NotNil(t, funcs["Green"])
		green := funcs["Green"].(func(string) string)("success")
		assert.Equal(t, "success", green)
	})
	t.Run("Red", func(t *testing.T) {
		require.NotNil(t, funcs["Red"])
		red := funcs["Red"].(func(string) string)("fail")
		assert.Equal(t, "fail", red)
	})
	t.Run("rpad", func(t *testing.T) {
		require.NotNil(t, funcs["rpad"])
		rpad := funcs["rpad"].(func(string, int) string)("big", 8)
		assert.Equal(t, "big     ", rpad)
		rpad = funcs["rpad"].(func(string, int) string)("small", 2)
		assert.Equal(t, "small", rpad)
	})
	t.Run("trimTrailingWhitespaces", func(t *testing.T) {
		require.NotNil(t, funcs["trimTrailingWhitespaces"])
		trimmed := funcs["trimTrailingWhitespaces"].(func(string) string)("padded   ")
		assert.Equal(t, "padded", trimmed)
	})
}

func TestPrintTemplate(t *testing.T) {
	tests := map[string]struct {
		tmpl     string
		data     interface{}
		expected string
		err      string // only partial
	}{
		"a plain template is output": {
			tmpl:     "hello world!",
			expected: "hello world!",
		},
		"template functions are called": {
			tmpl:     `{{Title "howdy"}}`,
			expected: "HOWDY",
		},
		"data can be used in templates": {
			tmpl:     `{{.Twelve}} is a number`,
			expected: "12 is a number",
			data:     struct{ Twelve int }{Twelve: 12},
		},
		"erroring templates will error": {
			tmpl: `{{RuhRoh .NotaFunction}}`,
			err:  `function "RuhRoh" not defined`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Error("A panic wasn't stopped: ", r)
				}
			}()
			buff := &bytes.Buffer{}
			err := PrintTemplate(buff, tt.tmpl, tt.data)
			if err != nil || tt.err != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.err)
			} else {
				assert.Equal(t, tt.expected, buff.String())
			}
		})
	}
}
