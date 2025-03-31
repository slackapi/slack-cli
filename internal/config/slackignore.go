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

package config

import (
	"os"
	"path/filepath"

	gogitignore "github.com/sabhiram/go-gitignore"
)

const SlackIgnoreFileName = ".slackignore"

var SlackIgnore *gogitignore.GitIgnore

// InitSlackIgnore: Initialize contents (file names, directories etc) of the .slackignore file.
// If an error occurs, it initializes as an empty .slackignore file
func InitSlackIgnore() {
	var wd, err = os.Getwd()
	if err != nil {
		SlackIgnore = &gogitignore.GitIgnore{}
		return
	}

	var slackignore = filepath.Join(wd, SlackIgnoreFileName)
	object, err := gogitignore.CompileIgnoreFile(slackignore)
	if err != nil {
		SlackIgnore = &gogitignore.GitIgnore{}
		return
	}

	SlackIgnore = object
}
