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

package platform

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_bundleArchive(t *testing.T) {
	tests := map[string]struct {
		files       map[string]string
		missingDir  bool
		expectedErr bool
	}{
		"zips each file with a path relative to the directory": {
			files: map[string]string{
				"manifest.json":    `{"name":"app"}`,
				"functions/run.js": "console.log('hi')",
			},
		},
		"an empty directory creates an empty archive": {
			files: map[string]string{},
		},
		"returns an error when the directory does not exist": {
			missingDir:  true,
			expectedErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := slackcontext.MockContext(t.Context())
			clientsMock := shared.NewClientsMock()
			clientsMock.AddDefaultMocks()
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())
			dir := t.TempDir()
			for path, content := range tc.files {
				require.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(path)), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(dir, path), []byte(content), 0o644))
			}
			if tc.missingDir {
				dir = filepath.Join(dir, "missing")
			}

			result, err := bundleArchive(ctx, clients, dir, "A123")
			if tc.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, packageResult{}, result)
				return
			}
			require.NoError(t, err)
			t.Cleanup(func() { _ = os.Remove(result.Filename) })
			assert.Contains(t, filepath.Base(result.Filename), "slack-cli-package.A123.")
			info, err := os.Stat(result.Filename)
			require.NoError(t, err)
			assert.Equal(t, info.Size(), result.Size)

			reader, err := zip.OpenReader(result.Filename)
			require.NoError(t, err)
			defer reader.Close()
			actual := map[string]string{}
			for _, file := range reader.File {
				rc, err := file.Open()
				require.NoError(t, err)
				content, err := io.ReadAll(rc)
				require.NoError(t, err)
				require.NoError(t, rc.Close())
				actual[filepath.ToSlash(file.Name)] = string(content)
			}
			assert.Equal(t, tc.files, actual)
		})
	}
}

func TestDeploySuccessText(t *testing.T) {
	tests := map[string]struct {
		app         types.App
		manifest    types.SlackYaml
		authSession api.AuthSession
		deployTime  string
		expected    []string
	}{
		"information from a workspace deploy is printed": {
			app: types.App{AppID: "A123"},
			manifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{Name: "DeployerApp"},
				},
			},
			authSession: api.AuthSession{
				UserName: new("slackbot"),
				UserID:   new("USLACKBOT"),
				TeamName: new("speck"),
				TeamID:   new("T001"),
			},
			deployTime: "12.34",
			expected: []string{
				"DeployerApp deployed in 12.34",
				"Dashboard:  https://slacker.com/apps/A123",
				"App Owner:  slackbot (USLACKBOT)",
				"Workspace:  speck (T001)",
			},
		},
		"information from an enterprise deploy is printed": {
			app: types.App{AppID: "A999"},
			manifest: types.SlackYaml{
				AppManifest: types.AppManifest{
					DisplayInformation: types.DisplayInformation{Name: "Spackulen"},
				},
			},
			authSession: api.AuthSession{
				UserName:            new("stub"),
				UserID:              new("U111"),
				TeamName:            new("spack"),
				TeamID:              new("E002"),
				EnterpriseID:        new("E002"),
				IsEnterpriseInstall: new(bool(true)),
			},
			deployTime: "8.05",
			expected: []string{
				"Spackulen deployed in 8.05",
				"Dashboard   :  https://slacker.com/apps/A999",
				"App Owner   :  stub (U111)",
				"Organization:  spack (E002)",
			},
		},
		"a message is still displayed with missing info": {
			app:         types.App{},
			manifest:    types.SlackYaml{},
			authSession: api.AuthSession{},
			expected: []string{
				"Successfully deployed the app!",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			clientsMock := shared.NewClientsMock()
			clientsMock.API.On("Host").Return("https://slacker.com")
			clients := shared.NewClientFactory(clientsMock.MockClientFactory())

			output := deploySuccessText(clients, tc.app, tc.manifest, tc.authSession, tc.deployTime)
			for _, line := range tc.expected {
				assert.Contains(t, output, line)
			}
		})
	}
}
