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
