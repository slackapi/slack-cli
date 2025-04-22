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

package cmdutil

import (
	"context"
	"fmt"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

// IsSlackHostedProject determines if the manifest represents a ROSI application
func IsSlackHostedProject(ctx context.Context, clients *shared.ClientFactory) error {
	manifestSource, err := clients.Config.ProjectConfig.GetManifestSource(ctx)
	if err != nil {
		return err
	}
	switch {
	case manifestSource.Equals(config.MANIFEST_SOURCE_LOCAL):
		manifest, err := clients.AppClient().Manifest.GetManifestLocal(ctx, clients.SDKConfig, clients.HookExecutor)
		if err != nil {
			return err
		}
		if !manifest.IsFunctionRuntimeSlackHosted() {
			return slackerror.New(slackerror.ErrAppNotHosted)
		}
	case manifestSource.Equals(config.MANIFEST_SOURCE_REMOTE):
		return slackerror.New(slackerror.ErrAppNotHosted).
			WithDetails(slackerror.ErrorDetails{
				{
					Code:        slackerror.ErrInvalidManifestSource,
					Message:     fmt.Sprintf("Slack hosted projects use \"%s\" manifest source", config.MANIFEST_SOURCE_LOCAL),
					Remediation: fmt.Sprintf("This value can be changed in configuration: \"%s\"", config.GetProjectConfigJSONFilePath("")),
				},
			})
	}
	return nil
}

// IsValidProjectDirectory verifies that a command is run in a valid project directory and returns nil, otherwise returns an error
func IsValidProjectDirectory(clients *shared.ClientFactory) error {
	if err, _ := clients.SDKConfig.Exists(); err != nil {
		return slackerror.New(slackerror.ErrInvalidAppDirectory)
	}
	return nil
}
