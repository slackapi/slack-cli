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

	"github.com/slackapi/slack-cli/internal/icon"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

type SlackYaml struct {
	AppManifest `yaml:",inline"`
	Icon        string `yaml:"icon"`
	Hash        string
}

// hasValidIconPath returns false if icon path is provided but is not valid and true otherwise
	wd, err := os.Getwd()
	if err != nil {
		return true
	}
	if sy.Icon == "" {
		sy.Icon = icon.ResolveIconPath(afero.NewBasePathFs(fs, wd), "")
	} else {
		if _, err := fs.Stat(filepath.Join(wd, sy.Icon)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Verify checks that the app manifest meets some basic requirements
	if !sy.hasValidIconPath(fs) {
		return slackerror.New("Please specify a valid icon path in app manifest")
	}
	return nil
}
