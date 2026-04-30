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
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

// Init copies a template placeholder file (.env.sample or .env.example) to
// .env. It returns an error if .env already exists, or if no placeholder file
// is found.
func Init(fs afero.Fs) (string, error) {
	sampleFiles := []string{".env.sample", ".env.example"}

	exists, err := afero.Exists(fs, ".env")
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrDotEnvFileRead).
			WithMessage("Failed to read the .env file: %s", err)
	}
	if exists {
		return "", slackerror.New(slackerror.ErrDotEnvFileAlreadyExists)
	}

	for _, name := range sampleFiles {
		data, err := afero.ReadFile(fs, name)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", slackerror.Wrap(err, slackerror.ErrDotEnvFileRead).
				WithMessage("Failed to read the %s file: %s", name, err)
		}
		if _, err := godotenv.UnmarshalBytes(data); err != nil {
			return "", slackerror.Wrap(err, slackerror.ErrDotEnvFileParse).
				WithMessage("Failed to parse the %s file", name)
		}
		if err := writeFile(fs, data); err != nil {
			return "", err
		}
		return name, nil
	}

	return "", slackerror.New(slackerror.ErrDotEnvPlaceholderNotFound)
}

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
	_, found := existing[name]
	if !found {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return writeFile(fs, []byte(content+newEntry+"\n"))
	}

	re := entryPattern(name)
	match := re.FindStringSubmatchIndex(content)
	if match != nil {
		prefix := ""
		if strings.Contains(content[match[0]:match[1]], "export") {
			prefix = "export "
		}
		comment := ""
		if match[4] >= 0 {
			comment = content[match[4]:match[5]]
		}
		content = content[:match[0]] + prefix + newEntry + comment + content[match[1]:]
	} else {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += newEntry + "\n"
	}
	return writeFile(fs, []byte(content))
}

// Unset removes a single environment variable from the .env file, preserving
// comments, blank lines, and other formatting. If the file does not exist or
// the key is not found, no action is taken.
func Unset(fs afero.Fs, name string) error {
	// Check for an existing .env file and parse it to detect existing keys.
	existing, err := Read(fs)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	_, found := existing[name]
	if !found {
		return nil
	}

	// Read the raw file content to find and remove the entry.
	raw, err := afero.ReadFile(fs, ".env")
	if err != nil {
		return slackerror.Wrap(err, slackerror.ErrDotEnvFileRead).
			WithMessage("Failed to read the .env file: %s", err)
	}
	content := string(raw)

	re := entryPattern(name)
	match := re.FindStringIndex(content)
	if match != nil {
		// Remove the matched entry and its trailing newline if present.
		end := match[1]
		if end < len(content) && content[end] == '\n' {
			end++
		}
		content = content[:match[0]] + content[end:]
		return writeFile(fs, []byte(content))
	}

	return nil
}

// entryPattern builds a regex that matches a .env entry for the given variable
// name. It handles optional export prefix, leading whitespace, spaces around
// the equals sign, quoted (double, single, backtick) and unquoted values
// including multiline double-quoted values, and optional inline comments.
func entryPattern(name string) *regexp.Regexp {
	return regexp.MustCompile(
		`(?m)(^[^\S\n]*export[^\S\n]+|^[^\S\n]*)` + regexp.QuoteMeta(name) + `[^\S\n]*=[^\S\n]*` +
			`(?:` +
			`"(?:[^"\\]|\\.)*"` + // double-quoted (with escapes)
			`|'[^']*'` + // single-quoted
			"|`[^`]*`" + // backtick-quoted
			`|(?:[^\s\n#]|\S#)*` + // unquoted: stop before inline comment (space + #)
			`)` +
			`([^\S\n]+#[^\n]*)?`, // optional inline comment
	)
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
