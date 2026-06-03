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
	"strings"

	"github.com/slackapi/slack-cli/internal/pkg/blocks"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/spf13/cobra"
)

func NewPreviewCommand(clients *shared.ClientFactory) *cobra.Command {
	var teamID string
	var outputFlag string

	cmd := &cobra.Command{
		// TODO: accept blocks JSON from stdin pipe (e.g. echo '{"blocks":[...]}' | slack blocks preview --team T123)
		Use:   "preview <blocks-json>",
		Short: "Preview Block Kit blocks in the Block Kit Builder",
		Long: strings.Join([]string{
			"Preview Block Kit blocks in the Block Kit Builder.",
			"",
			"The <blocks-json> argument must be a JSON object containing a top-level \"blocks\"",
			"key whose value is an array of Block Kit block objects.",
			"",
			"Example: '{\"blocks\":[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"Hello\"}}]}'",
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if teamID == "" {
				return slackerror.New(slackerror.ErrBlocksPreview).
					WithMessage("Team ID is required").
					WithRemediation("Provide a team ID with --team <team_id>")
			}

			filePath, err := blocks.Preview(ctx, clients, teamID, args[0], outputFlag)
			if err != nil {
				return err
			}
			cmd.Println(filePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&teamID, "team", "", "team ID for Block Kit Builder (required)")
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "file path to save the screenshot image (omit to print a data URI to stdout)")

	return cmd
}
