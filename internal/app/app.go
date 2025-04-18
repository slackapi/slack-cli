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

package app

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
)

// Client to access the app/project
type Client struct {
	Manifest ManifestClientInterface
	AppClientInterface
}

// NewClient returns a client with access to the API
func NewClient(
	apiClient api.ApiInterface,
	config *config.Config,
	fs afero.Fs,
	os types.Os,
) *Client {
	return &Client{
		Manifest:           NewManifestClient(apiClient, config),
		AppClientInterface: NewAppClient(config, fs, os),
	}
}

// UpdateDefaultProjectFiles should update any project specific files if any
func UpdateDefaultProjectFiles(fs afero.Fs, dirPath string, appDirName string) error {
	var filenames = []string{"manifest.json", "manifest.js", "manifest.ts"}

	for _, filename := range filenames {
		filePath := filepath.Join(dirPath, filename)
		fileData, err := afero.ReadFile(fs, filePath)
		if err != nil {
			continue
		}

		fileData = regexReplaceAppNameInManifest(fileData, appDirName)
		if err := afero.WriteFile(fs, filePath, fileData, 0644); err != nil {
			return err
		}
	}

	return nil
}

// regexReplaceAppNameInManifest will replace the app name for the following app manifest schemas:
// 1. Manifest JSON Schema: updates names in "display_information.name" and "bot_user.display_name"
// 2. Deno Slack SDK Manifest Schema: updates the name in `Manifest({ name: 'app name', ... })`
func regexReplaceAppNameInManifest(src []byte, appName string) []byte {
	// srcUpdated will contain the src with the updated app name
	srcUpdated := make([]byte, len(src))
	copy(srcUpdated, src)

	// List of JSON parent/child keys that reference app names to replace
	// Example: `display_information: { name: "app-name" }` - Parent Key: "display_information", Child Key: "name"
	// Example: `bot_user: { display_name: "app-name" }` - Parent Key: "bot_user", Child Key: "display_name"
	jsonObjects := []struct {
		ParentKey string
		ChildKey  string
	}{
		{
			ParentKey: "display_information",
			ChildKey:  "name",
		},
		{
			ParentKey: "bot_user",
			ChildKey:  "display_name",
		},
	}

	// Golang template for a regular expression that find the app name with 3 match groups:
	// - Match $1 - All content before the app name
	// - Match $2 - App name (excluding quotes, which are included in matches #1 and #3)
	// - Match $3 - All content after app name
	//
	// Try it: https://rubular.com/r/aFfWTkMZJnvReV
	// Please update this link when the pattern is modified
	t, err := template.New("manifest-json-app-name").Parse(
		`(?m)^(.*{{ .ParentKey }}[^}]*{{ .ChildKey }}[\s'"]*:\s*['"])([^'"\n]*)(['"].*)$`,
		//   | #1                | #2 | #3                  | #4    | #5      | #6     |
		//   | Match $1                                             | $2      | $3     |
		//    --------------------------------------------------------------------------
		//
		// Expression explained:
		// (?m) - Multi-line match that enables the use of ^ and $
		// #1   - Match $1 captures start of content to the Parent Key (e.g. "bot_user")
		// #2   - Continue match $1 unless there is a closing brace ("}"). This is the end of the Parnet or Child object.
		// #3   - Continue match $1 to Child Key with optional whitespace, single, or double quotes. This matches keys with or without quotes (JavaScript/TypeScript/JSON).
		// #4   - Continue match $1 to a colon with optional whitespace and a required single or double quote. This is the start of the Child Key name string.
		// #5   - Match $2 captures all text until a single, double quote or newline. Known issue is that it will fail to capture escaped quotes.
		// $6   - Match $3 all text to end of content
	)
	if err != nil {
		return src
	}

	// Update the manifest.json app name references
	for _, jsonObject := range jsonObjects {
		var regexBytes bytes.Buffer
		if err := t.Execute(&regexBytes, jsonObject); err != nil {
			return src
		}
		re := regexp.MustCompile(regexBytes.String())

		repl := fmt.Sprintf("${1}%s${3}", appName) // Match $2 is the original app name, which is replaced by appName
		srcUpdated = re.ReplaceAll(srcUpdated, []byte(repl))
	}

	// Update the Deno SDK app name references
	re := regexp.MustCompile(`(?m)^(.*Manifest\s*\(\s*{[^}]*['"]?name['"]?\s*:\s['"])([^'"]*)(['"].*)$`)
	//                            | #1                 | #2 | #3          | #4       | #5    | #6    |
	//                            | Match $1                                         | $2    | $3    |
	//                             -------------------------------------------------------------------
	//
	// Expression explained:
	// (?m) - Multi-line match that enables the use of ^ and $
	// #1   - Match $1 captures start of content to the `Manifest({` function with optional whitespace between the name, bracket `(```, and brace `{`
	// #2   - Continue match $1 unless a closing brace `}` is encountered
	// #3   - Continue match $1 to the `name` key surrounded by optional single or double quotes
	// #4   - Continue match $1 ignoring whitespace followed by a colon `:` and the opening single or double quotes of the string's value
	// #5   - Match $2 the value of the `name` key
	// #6   - Match $3 all text to end of content

	repl := fmt.Sprintf("${1}%s${3}", appName) // Match $2 is the original app name, which is replaced by appName
	srcUpdated = re.ReplaceAll(srcUpdated, []byte(repl))

	return srcUpdated
}
