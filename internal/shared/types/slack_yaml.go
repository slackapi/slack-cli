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

package types

import (
	"os"
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/slackerror"
)

type SlackYaml struct {
	AppManifest `yaml:",inline"`
	Icon        string `yaml:"icon"`
	Hash        string
}

var supportedIconExtensions = []string{".png", ".jpg", ".jpeg", ".gif"}

// hasValidIconPath returns false if icon path is provided but is not valid and true otherwise
func (sy *SlackYaml) hasValidIconPath() bool {
	var wd, err = os.Getwd()
	if err == nil {
		if sy.Icon == "" {
			for _, ext := range supportedIconExtensions {
				candidate := filepath.Join(wd, "assets", "icon"+ext)
				if _, err := os.Stat(candidate); !os.IsNotExist(err) {
					sy.Icon = filepath.Join("assets", "icon"+ext)
					break
				}
			}
		} else {
			if _, err := os.Stat(filepath.Join(wd, sy.Icon)); os.IsNotExist(err) {
				return false
			}
		}
	}

	return true
}

// Verify checks that the app manifest meets some basic requirements
func (sy *SlackYaml) Verify() error {
	if !sy.hasValidIconPath() {
		return slackerror.New("Please specify a valid icon path in app manifest")
	}
	return nil
}
