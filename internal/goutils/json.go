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

package goutils

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

// MergeJSON will merge customJSON into defaultJSON and unmarshal the results to config.
// When JSON arguments are an empty string, it will default the arg to "{}"
func MergeJSON(defaultJSON string, customJSON string, config interface{}) error {
	// Validate JSON
	if strings.TrimSpace(defaultJSON) == "" {
		defaultJSON = "{}"
	}
	if strings.TrimSpace(customJSON) == "" {
		customJSON = "{}"
	}

	// Default values
	if err := json.Unmarshal([]byte(defaultJSON), &config); err != nil {
		return err
	}

	// Overwrite defaults with custom values
	if err := json.NewDecoder(strings.NewReader(customJSON)).Decode(&config); err != nil {
		return err
	}

	return nil
}

// JsonMarshalUnescaped converts a struct into a JSON encoding without escaping
// characters
func JsonMarshalUnescaped(v interface{}) (string, error) {
	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

// JsonMarshalUnescapedIndent converts a struct into an easily readable JSON
// encoding without escaping characters
func JsonMarshalUnescapedIndent(v interface{}) (string, error) {
	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(v)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

// JsonUnmarshal is a wrapper for json.Unmarshal which parses the
// JSON-encoded data and stores the result in the value pointed to by v.
// If v is nil or not a pointer, json.Unmarshal returns an InvalidUnmarshalError
// which gets converted to a slackerror that is more human readable.
func JsonUnmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return slackerror.JsonUnmarshalError(err, data)
	}
	return nil
}
