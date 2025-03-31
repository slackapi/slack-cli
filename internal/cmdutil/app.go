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
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

const LocalAppNotInstalledMsg = "Local app is not installed"
const DeployedAppNotInstalledMsg = "App is not installed or deployed"

// appExists returns an error if an app has not yet been created (i.e it doesn't have an App ID)
//
// Note: An app can have an App ID and NOT yet be installed to the workspace because an app's ID is created
// before it is installed. However historically we have not highlighted the distinction for end users.
// Therefore the slackerror message text mentions install. In future, if it becomes possible for an enduser
// to separately manage app creation (manifest.create) and installation (apps.devInstall), this ought to be
// revisited.
func AppExists(app types.App, auth types.SlackAuth) error {
	if app.IsNew() || app.AppID == "" {
		if app.IsDev {
			return slackerror.New(slackerror.ErrAppNotInstalled).WithMessage(`Error: %s to "%s". Use %s to install it.`, LocalAppNotInstalledMsg, auth.TeamDomain, style.Commandf("run", false))
		} else {
			return slackerror.New(slackerror.ErrAppNotInstalled).WithMessage(`Error: %s to "%s". Use %s to install it.`, DeployedAppNotInstalledMsg, auth.TeamDomain, style.Commandf("install", false))
		}
	}
	return nil
}
