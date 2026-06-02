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

package blocks

import (
	"github.com/slackapi/slack-cli/internal/pkg/blocks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
)

func NewPreviewCommand(clients *shared.ClientFactory) *cobra.Command {
	var teamID string

	cmd := &cobra.Command{
		Use:   "preview <blocks-json>",
		Short: "Preview Block Kit blocks in the Block Kit Builder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if teamID == "" {
				return slackerror.New(slackerror.ErrBlocksPreview).
					WithMessage("Team ID is required").
					WithRemediation("Provide a team ID with --team <team_id>")
			}
			filePath, err := blocks.Preview(ctx, clients, teamID, args[0])
			if err != nil {
				return err
			}
			clients.IO.PrintInfo(ctx, false, "%s", filePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&teamID, "team", "", "team ID for Block Kit Builder (required)")

	return cmd
}
