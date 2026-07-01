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

package manifest

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/spf13/afero"
)

// FetchAndWriteRemoteManifest fetches the app manifest from remote settings and writes it to the project.
func FetchAndWriteRemoteManifest(ctx context.Context, clients *shared.ClientFactory, token, appID, projectPath string) error {
	slackYaml, err := clients.AppClient().Manifest.GetManifestRemote(ctx, token, appID)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(slackYaml.AppManifest); err != nil {
		return err
	}
	manifestPath := filepath.Join(projectPath, "manifest.json")
	if err := afero.WriteFile(clients.Fs, manifestPath, buf.Bytes(), 0644); err != nil {
		return err
	}
	hash, err := clients.Config.ProjectConfig.Cache().NewManifestHash(ctx, slackYaml.AppManifest)
	if err != nil {
		return err
	}
	return clients.Config.ProjectConfig.Cache().SetManifestHash(ctx, appID, hash)
}
