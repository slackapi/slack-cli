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

// Package slackdotenv reads and parses .env files from a project directory.
//
// It provides a single entry point for loading environment variables defined in
// a .env file so that multiple packages (commands, config, hooks) can share the
// same parsing behavior.
package slackdotenv

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

// Read parses a .env file from the working directory using the provided
// filesystem. It returns nil if the filesystem is nil or the file does not
// exist.
func Read(fs afero.Fs) (map[string]string, error) {
	if fs == nil {
		return nil, nil
	}
	file, err := afero.ReadFile(fs, ".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, slackerror.Wrap(err, slackerror.ErrDotEnvFileRead).
			WithMessage("Failed to read the .env file: %s", err)
	}
	vars, err := godotenv.UnmarshalBytes(file)
	if err != nil {
		return nil, slackerror.Wrap(err, slackerror.ErrDotEnvFileParse).
			WithMessage("Failed to parse the .env file: %s", err)
	}
	return vars, nil
}

// Set sets a single environment variable in the .env file, preserving
// comments, blank lines, and other formatting. If the key already exists its
// value is replaced in-place. Otherwise the entry is appended. The file is
// created if it does not exist.
func Set(fs afero.Fs, name string, value string) error {
	newEntry, err := godotenv.Marshal(map[string]string{name: value})
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrDotEnvVarMarshal).
			WithMessage("Failed to marshal the .env variable: %s", err)
	}

	// Verify the marshaled entry can be parsed back to avoid writing values
	// that would corrupt the .env file for future reads.
	if _, err := godotenv.Unmarshal(newEntry); err != nil {
		return slackerror.Wrap(err, slackerror.ErrDotEnvVarMarshal).
			WithMessage("Failed to marshal the .env variable: %s", err)
	}

	// Check for an existing .env file and parse it to detect existing keys.
	existing, err := Read(fs)
	if err != nil {
		return err
	}

	// If the file does not exist, create it with the new entry.
	if existing == nil {
		return writeFile(fs, []byte(newEntry+"\n"))
	}

	// Read the raw file content once for either the append or replace path.
	raw, err := afero.ReadFile(fs, ".env")
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrDotEnvFileRead).
			WithMessage("Failed to read the .env file: %s", err)
	}
	content := string(raw)

	// If the key is new, append the entry.
	oldValue, found := existing[name]
	if !found {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return writeFile(fs, []byte(content+newEntry+"\n"))
	}

	// Marshal the old value to find its canonical form in the file.
	oldEntry, err := godotenv.Marshal(map[string]string{name: oldValue})
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrDotEnvVarMarshal).
			WithMessage("Failed to marshal the .env variable: %s", err)
	}
	oldMarshaledValue := strings.TrimPrefix(oldEntry, name+"=")

	// Try each possible form of the old entry, longest (most specific) first.
	// For multiline values, also try the double-quoted raw form since the file
	// stores actual newlines while Marshal escapes them.
	entries := []string{
		"export " + name + "=" + oldMarshaledValue,
		"export " + name + "=" + `"` + oldValue + `"`,
		"export " + name + "=" + oldValue,
		"export " + name + "=" + "'" + oldValue + "'",
		name + "=" + oldMarshaledValue,
		name + "=" + `"` + oldValue + `"`,
		name + "=" + oldValue,
		name + "=" + "'" + oldValue + "'",
	}

	replaced := false
	for _, entry := range entries {
		if strings.Contains(content, entry) {
			if strings.HasPrefix(entry, "export ") {
				content = strings.Replace(content, entry, "export "+newEntry, 1)
			} else {
				content = strings.Replace(content, entry, newEntry, 1)
			}
			replaced = true
			break
		}
	}
	if !replaced {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += newEntry + "\n"
	}
	return writeFile(fs, []byte(content))
}

// writeFile writes data to the .env file, wrapping any error with a structured
// error code.
func writeFile(fs afero.Fs, data []byte) error {
	err := afero.WriteFile(fs, ".env", data, 0600)
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrDotEnvFileWrite).
			WithMessage("Failed to write the .env file: %s", err)
	}
	return nil
}
