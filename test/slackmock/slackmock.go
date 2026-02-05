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

package slackmock

import (
	"context"
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// createProject will mock a project's required directory and files.
// If there is an error, it will fail the test.
func CreateProject(t require.TestingT, ctx context.Context, fs afero.Fs, os types.Os, projectDirPath string) {
	var err error

	// Create the directory: `path/to/project`
	err = fs.Mkdir(projectDirPath, 0755)
	require.NoError(t, err)

	// Create the project-level config directory: `path/to/project/.slack/`
	err = fs.Mkdir(filepath.Join(projectDirPath, config.ProjectConfigDirName), 0755)
	require.NoError(t, err)

	// Create the project-level config file: `path/to/project/.slack/config.json`
	err = afero.WriteFile(fs, config.GetProjectConfigJSONFilePath(projectDirPath), []byte("{}\n"), 0644)
	require.NoError(t, err)

	// Create the project-level hooks file: `path/to/project/.slack/hooks.json`
	err = afero.WriteFile(fs, config.GetProjectHooksJSONFilePath(projectDirPath), []byte("{}\n"), 0644)
	require.NoError(t, err)
}
