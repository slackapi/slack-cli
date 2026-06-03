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
	"bytes"
	"encoding/json"
	"io"
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
		Use:   "preview",
		Short: "Preview Block Kit blocks in the Block Kit Builder",
		Long: strings.Join([]string{
			"Preview Block Kit blocks in the Block Kit Builder.",
			"",
			"The blocks JSON must be a JSON object containing a top-level \"blocks\"",
			"key whose value is an array of Block Kit block objects.",
			"",
			"Blocks JSON is read from stdin:",
			"  cat blocks.json | slack blocks preview --team T123 --output preview.png",
			"  echo '{\"blocks\":[...]}' | slack blocks preview --team T123 --output preview.png",
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if teamID == "" {
				return slackerror.New(slackerror.ErrBlocksPreview).
					WithMessage("Team ID is required").
					WithRemediation("Provide a team ID with --team <team_id>")
			}
			if outputFlag == "" {
				return slackerror.New(slackerror.ErrBlocksPreview).
					WithMessage("Output file path is required").
					WithRemediation("Provide an output path with --output <file_path>")
			}

			data, err := io.ReadAll(clients.IO.ReadIn())
			if err != nil {
				return slackerror.New(slackerror.ErrBlocksPreview).
					WithMessage("Failed to read blocks JSON from stdin").
					WithRemediation("Pipe blocks JSON via stdin:\n  cat blocks.json | slack blocks preview --team T123 --output preview.png")
			}
			blocksJSON := strings.TrimSpace(string(data))

			if blocksJSON == "" {
				return slackerror.New(slackerror.ErrBlocksPreview).
					WithMessage("No blocks JSON provided").
					WithRemediation("Pipe blocks JSON via stdin:\n  cat blocks.json | slack blocks preview --team T123 --output preview.png")
			}

			blocksJSON, err = compactJSON(blocksJSON)
			if err != nil {
				return err
			}
			if err := validateBlocksPayload(blocksJSON); err != nil {
				return err
			}

			filePath, err := blocks.Preview(ctx, clients, teamID, blocksJSON, outputFlag)
			if err != nil {
				return err
			}
			cmd.Println(filePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&teamID, "team", "", "team ID for Block Kit Builder (required)")
	cmd.Flags().StringVarP(&outputFlag, "output", "o", "", "file path to save the screenshot image (required)")

	return cmd
}

func compactJSON(input string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(input)); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrInvalidBlocksJSON)
	}
	return buf.String(), nil
}

func validateBlocksPayload(blocksJSON string) error {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(blocksJSON), &parsed); err != nil {
		return slackerror.New(slackerror.ErrInvalidBlocksJSON).
			WithMessage("The blocks JSON must be a JSON object containing a \"blocks\" array").
			WithRemediation("Provide a JSON object with a top-level \"blocks\" key, e.g. {\"blocks\": [...]}")
	}

	blocksRaw, ok := parsed["blocks"]
	if !ok {
		return slackerror.New(slackerror.ErrInvalidBlocksJSON).
			WithMessage("The blocks JSON is missing the required \"blocks\" field").
			WithRemediation("Provide a JSON object with a top-level \"blocks\" key, e.g. {\"blocks\": [...]}")
	}

	if len(blocksRaw) == 0 || blocksRaw[0] != '[' {
		return slackerror.New(slackerror.ErrInvalidBlocksJSON).
			WithMessage("The \"blocks\" field must be an array").
			WithRemediation("Provide a JSON object where \"blocks\" is an array, e.g. {\"blocks\": [...]}")
	}

	return nil
}
