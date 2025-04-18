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
	"fmt"

	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/cobra"
)

// Flag values
const (
	// OrgGrantWorkspaceFlag is used in the `run`, `deploy` and `install` commands
	// to specify an org workspace to add a grant for when installing
	OrgGrantWorkspaceFlag = "org-workspace-grant"
)

// OrgGrantWorkspaceDescription is the description for for --org-workspace-grant flag in the run, deploy and install commands
// This value is a function so that formatting is applied for the help page (when style is enabled).
var OrgGrantWorkspaceDescription = func() string {
	return fmt.Sprintf("grant access to a specific org workspace ID\n  %s",
		style.Secondary("(or 'all' for all workspaces in the org)"))
}

// IsFlagChanged checks if a certain flag has been set in the command
func IsFlagChanged(cmd *cobra.Command, flag string) bool {
	IsFlagSet := cmd.Flags().Lookup(flag)
	if IsFlagSet != nil {
		return cmd.Flag(flag).Changed
	}
	return false
}
