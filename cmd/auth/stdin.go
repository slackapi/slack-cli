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

package auth

import (
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

const stdinTokenValue = "-"

// ResolveTokenInput returns the token read from standard input when the value
// is "-" and otherwise returns the value unchanged. Reading from standard input
// keeps the token out of shell history and process listings.
func ResolveTokenInput(clients *shared.ClientFactory, value string) (string, error) {
	if value != stdinTokenValue {
		return value, nil
	}
	if clients.IO.IsStdinTTY() {
		return "", slackerror.New(slackerror.ErrMissingInput).
			WithMessage("No token was provided on standard input").
			WithRemediation("Pipe a token into the command, e.g. %s", style.Highlight("--token - < token.txt"))
	}
	token, err := clients.IO.ReadInAll()
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrMissingInput)
	}
	if token == "" {
		return "", slackerror.New(slackerror.ErrMissingInput).
			WithMessage("No token was provided on standard input")
	}
	return token, nil
}
