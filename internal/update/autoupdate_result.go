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

package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/afero"
)

// autoUpdateResultFileName is the file in the Slack config directory that holds the result
// of a background update until the next command reports it
const autoUpdateResultFileName = "autoupdate.json"

// AutoUpdateResult is the outcome of a background update of the Slack CLI
type AutoUpdateResult struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Error string `json:"error,omitempty"`
}

// readAutoUpdateResult returns the saved result of a background update, or nil without one
func readAutoUpdateResult(ctx context.Context, clients *shared.ClientFactory) (*AutoUpdateResult, error) {
	resultPath, err := autoUpdateResultPath(ctx, clients)
	if err != nil {
		return nil, err
	}
	resultBytes, err := afero.ReadFile(clients.Fs, resultPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result AutoUpdateResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, slackerror.Wrap(err, slackerror.ErrUnableToParseJSON)
	}
	return &result, nil
}

// writeAutoUpdateResult saves the result of a background update. It writes a temporary file
// and renames it into place, so a reader sees either the whole old file or the whole new one.
func writeAutoUpdateResult(ctx context.Context, clients *shared.ClientFactory, result AutoUpdateResult) error {
	resultPath, err := autoUpdateResultPath(ctx, clients)
	if err != nil {
		return err
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	tempFile, err := afero.TempFile(clients.Fs, filepath.Dir(resultPath), autoUpdateResultFileName+".*")
	if err != nil {
		return err
	}
	_, err = tempFile.Write(resultBytes)
	if closeErr := tempFile.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = clients.Fs.Rename(tempFile.Name(), resultPath)
	}
	if err != nil {
		_ = clients.Fs.Remove(tempFile.Name())
		return err
	}
	return nil
}

// clearAutoUpdateResult removes the saved result of a background update, if one exists
func clearAutoUpdateResult(ctx context.Context, clients *shared.ClientFactory) error {
	resultPath, err := autoUpdateResultPath(ctx, clients)
	if err != nil {
		return err
	}
	if err := clients.Fs.Remove(resultPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// autoUpdateResultPath returns the path to the saved result in the Slack config directory
func autoUpdateResultPath(ctx context.Context, clients *shared.ClientFactory) (string, error) {
	configDir, err := clients.Config.SystemConfig.SlackConfigDir(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, autoUpdateResultFileName), nil
}
